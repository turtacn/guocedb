// Package crypto provides cryptographic utilities for guocedb.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/turtacn/guocedb/common/errors"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomBytes creates a cryptographically secure random byte slice.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Simple a placeholder for a more advanced key management system.
type KeyManager struct {
	// In a real system, this would connect to a KMS like Vault or AWS KMS.
}

// NewKeyManager creates a new key manager.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// Encrypt encrypts data using a master key. Placeholder.
func (km *KeyManager) Encrypt(plaintext []byte) ([]byte, error) {
	// This would involve fetching a key, using AES-GCM or similar, etc.
	return nil, errors.ErrNotImplemented
}

// Decrypt decrypts data using a master key. Placeholder.
func (km *KeyManager) Decrypt(ciphertext []byte) ([]byte, error) {
	return nil, errors.ErrNotImplemented
}

// RotateKey rotates the master encryption key. Placeholder.
func (km *KeyManager) RotateKey() error {
	return errors.ErrNotImplemented
}

// SHA256 computes the SHA256 hash of a byte slice.
func SHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
