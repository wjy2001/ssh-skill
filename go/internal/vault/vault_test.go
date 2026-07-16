package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ssh-skill/internal/types"
)

var isWindows = runtime.GOOS == "windows"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	plaintext := []byte("hello, this is a test vault payload")

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Encrypted output should be longer than plaintext.
	if len(encrypted) <= len(plaintext) {
		t.Errorf("encrypted length %d <= plaintext length %d", len(encrypted), len(plaintext))
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip failed: got %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestEncryptNonceUniqueness(t *testing.T) {
	key, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	plaintext := []byte("same plaintext")

	enc1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("first Encrypt: %v", err)
	}

	enc2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("second Encrypt: %v", err)
	}

	if string(enc1) == string(enc2) {
		t.Error("two encryptions of same plaintext produced identical ciphertexts (nonce reuse?)")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateRandomKey()
	key2, _ := GenerateRandomKey()

	encrypted, err := Encrypt([]byte("secret"), key1)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(encrypted, key2)
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecryptTooShort(t *testing.T) {
	_, err := Decrypt([]byte("short"), make([]byte, keyLen))
	if err != ErrInvalidCiphertext {
		t.Fatalf("expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestGenerateRandomKeyLength(t *testing.T) {
	key, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}
	if len(key) != keyLen {
		t.Fatalf("key length %d, want %d", len(key), keyLen)
	}
}

func TestKeyHashValidation(t *testing.T) {
	key, _ := GenerateRandomKey()
	hash := HashKey(key)

	if !ValidateKeyHash(key, hash) {
		t.Error("ValidateKeyHash failed for correct key")
	}

	wrongKey, _ := GenerateRandomKey()
	if ValidateKeyHash(wrongKey, hash) {
		t.Error("ValidateKeyHash passed for wrong key")
	}
}

func TestSaveAndLoadVault(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "servers.json.age")

	key, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	// Save.
	vault := &types.Vault{
		Version: 1,
		Servers: []types.ServerConfig{
			{ID: "test-1", Name: "Test Server", Host: "10.0.0.1", Port: 22, User: "root", Auth: types.AuthConfig{Method: types.AuthPassword, EncryptedPassword: "enc-pass"}},
		},
	}
	if err := Save(vaultPath, key, vault); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists.
	if _, err := os.Stat(vaultPath); err != nil {
		t.Fatalf("stat vault file: %v", err)
	}

	// Permission check is skipped on Windows (POSIX permissions not enforced).
	if !isWindows {
		info, err := os.Stat(vaultPath)
		if err != nil {
			t.Fatalf("stat vault file: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("vault file permissions %o, want 0600", info.Mode().Perm())
		}
	}

	// Verify file content is not plaintext JSON.
	raw, _ := os.ReadFile(vaultPath)
	if json.Valid(raw) {
		t.Error("vault file contains valid JSON (should be encrypted binary)")
	}

	// Load and verify round-trip.
	loaded, err := Load(vaultPath, key)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Servers) != 1 || loaded.Servers[0].ID != "test-1" {
		t.Fatalf("loaded vault mismatch: %+v", loaded)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "nonexistent.json.age")

	key, _ := GenerateRandomKey()
	vault, err := Load(vaultPath, key)
	if err != nil {
		t.Fatalf("Load nonexistent file: %v", err)
	}
	if vault.Version != 1 || len(vault.Servers) != 0 {
		t.Fatalf("expected empty vault, got %+v", vault)
	}
}

func TestEnsureKeyNewFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, ".vault-key")

	key, err := EnsureKey(keyPath)
	if err != nil {
		t.Fatalf("EnsureKey (new): %v", err)
	}
	if len(key) != keyLen {
		t.Fatalf("key length %d, want %d", len(key), keyLen)
	}

	// Check permissions on key file (skip on Windows).
	if !isWindows {
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("stat key file: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("key file permissions %o, want 0600", info.Mode().Perm())
		}
	}
}

func TestEnsureKeyExistingFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, ".vault-key")

	// Write initial key.
	key1, _ := GenerateRandomKey()
	os.WriteFile(keyPath, key1, 0600)

	// Read back.
	key2, err := EnsureKey(keyPath)
	if err != nil {
		t.Fatalf("EnsureKey (existing): %v", err)
	}

	if string(key1) != string(key2) {
		t.Fatal("EnsureKey returned different key for existing file")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	masterKey := []byte("test-master-key")
	salt := []byte("test-salt-123456")

	dk1 := DeriveKey(masterKey, salt)
	dk2 := DeriveKey(masterKey, salt)

	if string(dk1) != string(dk2) {
		t.Fatal("DeriveKey not deterministic")
	}

	if len(dk1) != keyLen {
		t.Fatalf("derived key length %d, want %d", len(dk1), keyLen)
	}
}
