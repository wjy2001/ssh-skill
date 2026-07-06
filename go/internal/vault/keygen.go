package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureKey returns the vault encryption key.
// If the key file exists, it reads and returns the key.
// If not, it generates a new random key, writes it to disk, and returns it.
func EnsureKey(keyPath string) ([]byte, error) {
	// Try reading existing key.
	key, err := os.ReadFile(keyPath)
	if err == nil {
		if len(key) != keyLen {
			return nil, fmt.Errorf("vault: key file %s has wrong length (%d, expected %d)", keyPath, len(key), keyLen)
		}
		return key, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("vault: read key file %s: %w", keyPath, err)
	}

	// Generate new key.
	key, err = GenerateRandomKey()
	if err != nil {
		return nil, err
	}

	// Ensure parent directory exists.
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("vault: create config dir %s: %w", dir, err)
	}

	// Write key with restricted permissions.
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("vault: write key file %s: %w", keyPath, err)
	}

	return key, nil
}
