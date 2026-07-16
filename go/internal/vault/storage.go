package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ssh-skill/internal/types"
)

// Save serializes the vault to JSON, encrypts it, and writes to disk.
func Save(path string, key []byte, vault *types.Vault) error {
	plaintext, err := json.Marshal(vault)
	if err != nil {
		return fmt.Errorf("vault: marshal: %w", err)
	}

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("vault: encrypt: %w", err)
	}

	// Ensure parent directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("vault: create dir %s: %w", dir, err)
	}

	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("vault: write file %s: %w", path, err)
	}

	return nil
}

// Load reads, decrypts, and unmarshals the vault from disk.
// Returns nil and no error if the file does not exist (first run).
func Load(path string, key []byte) (*types.Vault, error) {
	encrypted, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.Vault{Version: 1, Servers: []types.ServerConfig{}}, nil
		}
		return nil, fmt.Errorf("vault: read file %s: %w", path, err)
	}

	plaintext, err := Decrypt(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("vault: decrypt: %w", err)
	}

	var vault types.Vault
	if err := json.Unmarshal(plaintext, &vault); err != nil {
		return nil, fmt.Errorf("vault: unmarshal: %w", err)
	}

	// Ensure Servers is never nil for safe range iteration.
	if vault.Servers == nil {
		vault.Servers = []types.ServerConfig{}
	}

	return &vault, nil
}
