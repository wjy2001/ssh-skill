package cli

import (
	"fmt"

	"ssh-skill/internal/ssh"
	"ssh-skill/internal/types"
)

// resolveServer finds a server by ID and decrypts its password if needed.
// Returns a server config ready for SSH connection.
func resolveServer(app *App, serverID string) (*types.ServerConfig, error) {
	cfg, err := ssh.FindServer(app.Vault, serverID)
	if err != nil {
		return nil, err
	}

	if cfg.Auth.Method == types.AuthPassword && cfg.Auth.EncryptedPassword != "" {
		decrypted, err := decryptField(app.VaultKey, cfg.Auth.EncryptedPassword)
		if err != nil {
			return nil, fmt.Errorf("decrypt password: %w", err)
		}
		cfg.Auth.EncryptedPassword = decrypted
	}

	return cfg, nil
}
