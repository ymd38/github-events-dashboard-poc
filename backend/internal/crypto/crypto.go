package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// TokenEncryptor encrypts and decrypts tokens using AES-256-GCM.
type TokenEncryptor struct {
	aead cipher.AEAD
}

// NewTokenEncryptor creates a new TokenEncryptor with the given hex-encoded 32-byte key.
func NewTokenEncryptor(hexKey string) (*TokenEncryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: must be hex-encoded: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key: must be 32 bytes (64 hex characters), got %d bytes", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return &TokenEncryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext and returns a hex-encoded ciphertext (nonce + sealed data).
func (e *TokenEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	sealed := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt decrypts a hex-encoded ciphertext and returns the plaintext.
func (e *TokenEncryptor) Decrypt(hexCiphertext string) (string, error) {
	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}
	nonceSize := e.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
