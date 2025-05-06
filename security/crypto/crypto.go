// Package crypto contains cryptographic utilities for the database.
// It includes functions for password hashing, data encryption, etc.
//
// crypto 包包含数据库的密码学工具。
// 它包括密码哈希、数据加密等函数。
package crypto

import (
	"golang.org/x/crypto/bcrypt" // Standard library for password hashing
	"github.com/turtacn/guocedb/common/log"
	"fmt"
)

// PasswordHasher is an interface for hashing and verifying passwords.
// PasswordHasher 是用于哈希和验证密码的接口。
type PasswordHasher interface {
	// HashPassword hashes a plaintext password.
	// HashPassword 哈希明文密码。
	HashPassword(password string) (string, error)

	// VerifyPassword compares a plaintext password with a hashed password.
	// VerifyPassword 将明文密码与哈希密码进行比较。
	VerifyPassword(plaintext, hashed string) (bool, error)
}

// BcryptHasher is a PasswordHasher implementation using bcrypt.
// BcryptHasher 是使用 bcrypt 的 PasswordHasher 实现。
type BcryptHasher struct{}

// NewBcryptHasher creates a new BcryptHasher instance.
// NewBcryptHasher 创建一个新的 BcryptHasher 实例。
func NewBcryptHasher() PasswordHasher {
	log.Info("Initializing bcrypt password hasher.") // 初始化 bcrypt 密码哈希器。
	return &BcryptHasher{}
}

// HashPassword hashes the input password using bcrypt.
// HashPassword 使用 bcrypt 哈希输入密码。
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	// Use bcrypt.GenerateFromPassword with a suitable cost.
	// 使用 bcrypt.GenerateFromPassword 和适当的 cost。
	// bcrypt.DefaultCost is a good starting point.
	// bcrypt.DefaultCost 是一个好的起点。
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash password using bcrypt: %v", err) // 使用 bcrypt 哈希密码失败。
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plaintext password with a bcrypt hashed password.
// VerifyPassword 将明文密码与 bcrypt 哈希密码进行比较。
func (h *BcryptHasher) VerifyPassword(plaintext, hashed string) (bool, error) {
	// Use bcrypt.CompareHashAndPassword.
	// 使用 bcrypt.CompareHashAndPassword。
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext))
	if err == nil {
		// Passwords match
		// 密码匹配
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		// Passwords do not match
		// 密码不匹配
		return false, nil
	}
	// Other errors (e.g., invalid hash)
	// 其他错误（例如，无效哈希）
	log.Error("Error verifying password using bcrypt: %v", err) // 使用 bcrypt 验证密码出错。
	return false, fmt.Errorf("failed to verify password: %w", err)
}


// TODO: Add interfaces and implementations for data encryption/decryption (e.g., AES-GCM).
// TODO: 添加用于数据加密/解密（例如 AES-GCM）的接口和实现。

// TODO: Add interfaces and implementations for TLS certificate management if handling TLS connections directly.
// TODO: 如果直接处理 TLS 连接，添加用于 TLS 证书管理的接口和实现。

// TODO: Add interfaces and implementations for key management (simple file-based, or integrate with KMS).
// TODO: 添加用于密钥管理（简单文件方式，或与 KMS 集成）的接口和实现。