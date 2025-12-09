// Package auth provides authentication services for GuoceDB.
package auth

import (
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor
	DefaultCost = 12
)

// HashPassword generates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword verifies a password against its hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashMySQLNativePassword creates a MySQL native password hash.
// Format: SHA1(SHA1(password))
func HashMySQLNativePassword(password string) string {
	if password == "" {
		return ""
	}
	first := sha1.Sum([]byte(password))
	second := sha1.Sum(first[:])
	return fmt.Sprintf("*%X", second)
}
