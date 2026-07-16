// Package config handles configuration resolution and defaults.
package config

import (
	"os"
	"path/filepath"
)

const (
	// DefaultConfigDir is the default directory for ssh-skill configuration.
	DefaultConfigDir = ".ssh-skill"

	// EnvConfigDir overrides the default configuration directory.
	EnvConfigDir = "SSH_SKILL_CONFIG_DIR"
)

// Dir returns the configuration directory path.
// Priority: SSH_SKILL_CONFIG_DIR env var > ~/.ssh-skill/
func Dir() (string, error) {
	if envDir := os.Getenv(EnvConfigDir); envDir != "" {
		return envDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DefaultConfigDir), nil
}

// VaultKeyPath returns the path to the vault encryption key file.
func VaultKeyPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".vault-key"), nil
}

// VaultFilePath returns the path to the encrypted server configuration file.
func VaultFilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "servers.json.age"), nil
}

// AuditLogPath returns the path to the audit log file.
func AuditLogPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "audit.log"), nil
}
