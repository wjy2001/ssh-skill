package cli

import (
	"encoding/hex"
	"fmt"

	"ssh-mcp/internal/vault"
)

func cmdVault(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: ssh-mcp vault init")
		return nil
	}

	switch args[0] {
	case "init":
		app, err := Load()
		if err != nil {
			return err
		}
		fmt.Printf("Vault initialized.\n")
		fmt.Printf("  Config directory: %s\n", app.ConfigDir)
		fmt.Printf("  Vault file:       %s\n", app.VaultPath)
		fmt.Printf("  Key length:       %d bytes\n", len(app.VaultKey))
		fmt.Printf("  Servers:          %d\n", len(app.Vault.Servers))
		// Save the empty vault to ensure files exist.
		if err := app.Save(); err != nil {
			return err
		}
		fmt.Println("Vault key and empty configuration file created.")
		return nil
	case "--help", "-h":
		fmt.Println("Usage: ssh-mcp vault init")
		return nil
	default:
		return fmt.Errorf("unknown vault subcommand: %s", args[0])
	}
}

// encryptField encrypts a plaintext value using the vault key.
func encryptField(vaultKey []byte, plaintext string) (string, error) {
	encrypted, err := vault.Encrypt([]byte(plaintext), vaultKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

// decryptField decrypts an encrypted hex-encoded value using the vault key.
func decryptField(vaultKey []byte, hexCiphertext string) (string, error) {
	encrypted, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("decrypt: invalid hex: %w", err)
	}
	plaintext, err := vault.Decrypt(encrypted, vaultKey)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
