package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"ssh-mcp/internal/types"
)

const (
	// DefaultTimeout is the default timeout for SSH connections and commands.
	DefaultTimeout = 30 * time.Second
)

var (
	ErrServerNotFound    = errors.New("ssh: server not found in configuration")
	ErrAuthNotConfigured = errors.New("ssh: no valid authentication method configured")
	ErrConnectFailed     = errors.New("ssh: connection failed")
)

// Client wraps an ssh.Client and owns the lifecycle of any bastion (jump host)
// connection underneath. Callers MUST call Close() when done so that both the
// terminal SSH client and — when present — the bastion client are released.
type Client struct {
	*ssh.Client
	bastion *ssh.Client // nil for direct connections
}

// Close closes the underlying ssh.Client and, if present, the bastion client.
// It returns the first error encountered; bastion close errors are ignored
// when the primary close already failed.
func (c *Client) Close() error {
	primaryErr := c.Client.Close()
	if c.bastion != nil {
		_ = c.bastion.Close()
	}
	return primaryErr
}

// FindServer looks up a server configuration by ID in the vault.
// Returns ErrServerNotFound if the server is not found.
func FindServer(vault *types.Vault, serverID string) (*types.ServerConfig, error) {
	for i := range vault.Servers {
		if vault.Servers[i].ID == serverID {
			return &vault.Servers[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrServerNotFound, serverID)
}

// Connect establishes an SSH connection to the given server using the appropriate
// authentication method (password, key, or agent). The returned Client owns the
// connection lifecycle — including any bastion connection — and must be closed.
//
// Security note: HostKeyCallback is currently ssh.InsecureIgnoreHostKey(),
// which means MITM attacks are NOT defended against. See docs/security.md
// "Threat model" for the explicit out-of-scope list.
func Connect(ctx context.Context, cfg *types.ServerConfig) (*Client, error) {
	authMethods, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.Port == 0 {
		addr = fmt.Sprintf("%s:22", cfg.Host)
	}

	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // v1: accept all host keys (MITM not defended)
		Timeout:         DefaultTimeout,
	}

	// Handle bastion (jump host).
	if cfg.Bastion != nil {
		return connectViaBastion(ctx, addr, clientConfig, cfg.Bastion)
	}

	sshClient, err := connect(ctx, addr, clientConfig)
	if err != nil {
		return nil, err
	}
	return &Client{Client: sshClient}, nil
}

func connect(ctx context.Context, addr string, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	d := &net.Dialer{Timeout: cfg.Timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%w: dial %s: %w", ErrConnectFailed, addr, err)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: handshake %s: %w", ErrConnectFailed, addr, err)
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func connectViaBastion(ctx context.Context, targetAddr string, targetCfg *ssh.ClientConfig, bastion *types.BastionConfig) (*Client, error) {
	// Connect to bastion first.
	bastionAuth, err := buildBastionAuth(bastion)
	if err != nil {
		return nil, fmt.Errorf("bastion auth: %w", err)
	}

	bastionAddr := fmt.Sprintf("%s:%d", bastion.Host, bastion.Port)
	if bastion.Port == 0 {
		bastionAddr = fmt.Sprintf("%s:22", bastion.Host)
	}

	bastionCfg := &ssh.ClientConfig{
		User:            bastion.User,
		Auth:            bastionAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // v1: accept all host keys (MITM not defended)
		Timeout:         DefaultTimeout,
	}

	bastionClient, err := connect(ctx, bastionAddr, bastionCfg)
	if err != nil {
		return nil, fmt.Errorf("bastion: %w", err)
	}

	// Dial through bastion.
	conn, err := bastionClient.Dial("tcp", targetAddr)
	if err != nil {
		bastionClient.Close()
		return nil, fmt.Errorf("%w: dial via bastion %s: %w", ErrConnectFailed, targetAddr, err)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, targetCfg)
	if err != nil {
		conn.Close()
		bastionClient.Close()
		return nil, fmt.Errorf("%w: handshake via bastion %s: %w", ErrConnectFailed, targetAddr, err)
	}

	// Wrap so the returned Client owns the bastion lifecycle via Close().
	return &Client{
		Client:  ssh.NewClient(c, chans, reqs),
		bastion: bastionClient,
	}, nil
}

// buildAuthMethods constructs the SSH authentication method list from the server config.
// The cfg.Auth.EncryptedPassword field MUST contain the decrypted plaintext password
// when Method == AuthPassword — decryption is the caller's responsibility (cli layer).
func buildAuthMethods(cfg *types.ServerConfig) ([]ssh.AuthMethod, error) {
	switch cfg.Auth.Method {
	case types.AuthPassword:
		if cfg.Auth.EncryptedPassword == "" {
			return nil, fmt.Errorf("%w: password is empty", ErrAuthNotConfigured)
		}
		return []ssh.AuthMethod{ssh.Password(cfg.Auth.EncryptedPassword)}, nil

	case types.AuthKey:
		if cfg.Auth.PrivateKeyPath == "" {
			return nil, fmt.Errorf("%w: private key path is empty", ErrAuthNotConfigured)
		}
		return buildKeyAuth(cfg.Auth.PrivateKeyPath, cfg.Auth.EncryptedPassphrase)

	case types.AuthAgent:
		return buildAgentAuth()

	default:
		return nil, fmt.Errorf("%w: unknown method %s", ErrAuthNotConfigured, cfg.Auth.Method)
	}
}

func buildBastionAuth(b *types.BastionConfig) ([]ssh.AuthMethod, error) {
	switch b.Auth.Method {
	case types.AuthPassword:
		return []ssh.AuthMethod{ssh.Password(b.Auth.EncryptedPassword)}, nil
	case types.AuthKey:
		return buildKeyAuth(b.Auth.PrivateKeyPath, b.Auth.EncryptedPassphrase)
	case types.AuthAgent:
		return buildAgentAuth()
	default:
		return nil, fmt.Errorf("%w: unknown bastion auth method %s", ErrAuthNotConfigured, b.Auth.Method)
	}
}

func buildKeyAuth(keyPath, encryptedPassphrase string) ([]ssh.AuthMethod, error) {
	expandedPath := os.ExpandEnv(keyPath)
	if expandedPath == "" {
		return nil, fmt.Errorf("%w: private key path is empty after expansion", ErrAuthNotConfigured)
	}
	// Expand ~/ and ~\ prefixes to the user's home directory.
	if strings.HasPrefix(expandedPath, "~/") || strings.HasPrefix(expandedPath, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir: %w", err)
		}
		expandedPath = home + expandedPath[1:]
	}

	keyBytes, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("read key file %s: %w", expandedPath, err)
	}

	var signer ssh.Signer
	if encryptedPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(encryptedPassphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("parse key %s: %w", expandedPath, err)
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

func buildAgentAuth() ([]ssh.AuthMethod, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("%w: SSH_AUTH_SOCK not set", ErrAuthNotConfigured)
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("connect to ssh-agent %s: %w", socket, err)
	}

	agentClient := agent.NewClient(conn)
	return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil
}
