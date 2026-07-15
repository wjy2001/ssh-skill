// Package types defines the shared data structures used across all internal packages.
// It is the bottom layer of the architecture (Types → Config → Repo → Service → Runtime → UI).
package types

import "time"

// AuthMethod represents the SSH authentication method.
type AuthMethod string

const (
	AuthPassword AuthMethod = "password"
	AuthKey      AuthMethod = "key"
	AuthAgent    AuthMethod = "agent"
)

// AuthConfig holds the authentication configuration for a server.
// Only one method is active at a time, determined by the Method field.
type AuthConfig struct {
	Method              AuthMethod `json:"method"`
	EncryptedPassword   string     `json:"encrypted_password,omitempty"`
	PrivateKeyPath      string     `json:"private_key_path,omitempty"`
	EncryptedPassphrase string     `json:"encrypted_passphrase,omitempty"`
}

// BastionConfig holds the jump host (bastion) configuration.
type BastionConfig struct {
	Host string     `json:"host"`
	Port int        `json:"port"`
	User string     `json:"user"`
	Auth AuthConfig `json:"auth"`
}

// ServerConfig represents a configured remote server.
// This is the persistent entity stored in the encrypted vault.
type ServerConfig struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Host    string         `json:"host"`
	Port    int            `json:"port"`
	User    string         `json:"user"`
	Auth    AuthConfig     `json:"auth"`
	Bastion *BastionConfig `json:"bastion,omitempty"`
	Tags    []string       `json:"tags,omitempty"`
}

// Vault is the encrypted storage container holding all server configurations.
type Vault struct {
	Version int            `json:"version"`
	Servers []ServerConfig `json:"servers"`
}

// ExecResult is the result of executing a command on a remote server.
type ExecResult struct {
	ServerID   string `json:"server_id"`
	Command    string `json:"command"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

// FileTransferResult is the result of uploading or downloading a file.
type FileTransferResult struct {
	ServerID   string `json:"server_id"`
	Path       string `json:"path"`
	SizeBytes  int64  `json:"size_bytes"`
	DurationMs int64  `json:"duration_ms"`
}

// ProgressCallback is called during file transfers to report progress.
// bytesTransferred: bytes copied so far
// totalBytes: total file size (may be 0 if unknown)
// elapsed: time since transfer started
type ProgressCallback func(bytesTransferred, totalBytes int64, elapsed time.Duration)

// AuditEntry represents a single command execution audit record.
// Stored as JSONL in ~/.ssh-mcp/audit.log.
type AuditEntry struct {
	Timestamp  string `json:"timestamp"`
	ServerID   string `json:"server_id"`
	ServerHost string `json:"server_host"`
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	StdoutLen  int    `json:"stdout_len"`
	StderrLen  int    `json:"stderr_len"`
	DurationMs int64  `json:"duration_ms"`
}
