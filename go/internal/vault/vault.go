package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	keyLen   = 32 // AES-256
	saltLen  = 16
	nonceLen = 12

	// Argon2id parameters.
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
)

var (
	ErrInvalidCiphertext = errors.New("vault: ciphertext too short")
	ErrDecryptionFailed  = errors.New("vault: decryption failed — wrong key or corrupted data")
)

// DeriveKey derives an AES-256 key from a master key and salt using Argon2id.
// Intended for future passphrase-based key derivation; currently used with the
// raw vault key as the "password" input. We keep the derivation step so the
// on-disk format is forward-compatible with passphrase support.
func DeriveKey(masterKey, salt []byte) []byte {
	return argon2.IDKey(masterKey, salt, argonTime, argonMemory, argonThreads, keyLen)
}

// Encrypt encrypts plaintext using AES-256-GCM.
// The output format is: [16B salt][12B nonce][ciphertext].
// A random salt and nonce are generated for each encryption.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("vault: generate salt: %w", err)
	}

	derivedKey := DeriveKey(key, salt)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("vault: create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: create gcm: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("vault: generate nonce: %w", err)
	}

	// GCM seal appends the ciphertext to the nonce prefix.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Build final payload: salt + nonce + ciphertext
	out := make([]byte, 0, saltLen+nonceLen+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt.
// Expects the input format: [16B salt][12B nonce][ciphertext].
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < saltLen+nonceLen+1 {
		return nil, ErrInvalidCiphertext
	}

	salt := ciphertext[:saltLen]
	nonce := ciphertext[saltLen : saltLen+nonceLen]
	encrypted := ciphertext[saltLen+nonceLen:]

	derivedKey := DeriveKey(key, salt)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("vault: create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: create gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// GenerateRandomKey generates a cryptographically random 32-byte key.
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("vault: generate key: %w", err)
	}
	return key, nil
}

// HashKey returns the SHA-256 hash of a key, used for integrity checks.
func HashKey(key []byte) []byte {
	h := sha256.Sum256(key)
	return h[:]
}

// ValidateKeyHash checks whether the stored hash matches the key's hash.
func ValidateKeyHash(key, storedHash []byte) bool {
	computed := HashKey(key)
	if len(storedHash) != len(computed) {
		return false
	}
	// Constant-time comparison.
	var diff byte
	for i := range computed {
		diff |= computed[i] ^ storedHash[i]
	}
	return diff == 0
}
