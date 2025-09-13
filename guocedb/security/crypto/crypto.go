package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encryptor is an interface for encrypting data.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
}

// Decryptor is an interface for decrypting data.
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// AESGCM represents an AES-GCM implementation for encryption and decryption.
type AESGCM struct {
	key []byte
	gcm cipher.AEAD
}

// NewAESGCM creates a new AES-GCM encryptor/decryptor with the given key.
// The key must be 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewAESGCM(key []byte) (*AESGCM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESGCM{key: key, gcm: gcm}, nil
}

// Encrypt encrypts data using AES-GCM.
func (a *AESGCM) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal will append the output to the first argument; since we want the nonce to
	// be prepended, we pass it as the first argument.
	return a.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using AES-GCM.
func (a *AESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return a.gcm.Open(nil, nonce, ciphertext, nil)
}

// TODO: Implement a robust key management system (e.g., using a KMS).
// TODO: Implement key rotation mechanisms.
