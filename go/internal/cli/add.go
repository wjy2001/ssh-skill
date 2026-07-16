package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"ssh-skill/internal/types"
)

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	id := fs.String("id", "", "Server identifier (e.g., prod-web)")
	name := fs.String("name", "", "Human-readable name")
	host := fs.String("host", "", "Hostname or IP address")
	port := fs.Int("port", 22, "SSH port (default 22)")
	user := fs.String("user", "root", "SSH user")
	authType := fs.String("auth-type", "", "Authentication type: password, key, or agent")
	password := fs.String("password", "", "SSH password (will be encrypted in vault)")
	keyPath := fs.String("key-path", "", "Path to SSH private key file")
	fs.Parse(args)

	if *id == "" || *host == "" || *authType == "" {
		fmt.Fprintln(os.Stderr, "Required flags: --id, --host, --auth-type")
		fs.Usage()
		return fmt.Errorf("missing required flags")
	}

	app, err := Load()
	if err != nil {
		return err
	}

	// Check for duplicate ID.
	for _, s := range app.Vault.Servers {
		if s.ID == *id {
			return fmt.Errorf("server with id '%s' already exists", *id)
		}
	}

	authConfig := types.AuthConfig{
		Method: types.AuthMethod(*authType),
	}

	switch types.AuthMethod(*authType) {
	case types.AuthPassword:
		if *password == "" {
			return fmt.Errorf("--password is required for auth-type=password")
		}
		// Encrypt password before storing.
		encrypted, err := encryptField(app.VaultKey, *password)
		if err != nil {
			return fmt.Errorf("encrypt password: %w", err)
		}
		authConfig.EncryptedPassword = encrypted

	case types.AuthKey:
		if *keyPath == "" {
			return fmt.Errorf("--key-path is required for auth-type=key")
		}
		authConfig.PrivateKeyPath = *keyPath

	case types.AuthAgent:
		// No extra fields needed.

	default:
		return fmt.Errorf("unknown auth type: %s (use password, key, or agent)", *authType)
	}

	server := types.ServerConfig{
		ID:   *id,
		Name: *name,
		Host: *host,
		Port: *port,
		User: *user,
		Auth: authConfig,
	}

	app.Vault.Servers = append(app.Vault.Servers, server)
	if err := app.Save(); err != nil {
		return err
	}

	fmt.Printf("Server '%s' (%s@%s:%s) added.\n", server.ID, server.User, server.Host, strconv.Itoa(server.Port))
	return nil
}
