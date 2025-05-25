// Package utils provides a collection of general utility functions that are shared
// across different parts of the Guocedb project. These functions aim to encapsulate
// common, repetitive, or complex logic, promoting code reuse and reducing redundancy.
//
// 此包提供了一系列在 Guocedb 项目不同部分共享的通用工具函数。
// 这些函数旨在封装常见、重复或复杂的逻辑，促进代码重用并减少冗余。
package utils

import (
	"crypto/sha256" // For SHA256 hashing
	"encoding/hex"  // For encoding/decoding to/from hex string
	"fmt"
	"os"            // For file system operations
	"path/filepath" // For path manipulation
	"strings"       // For string operations
)

// GenerateSHA256Hash computes the SHA256 hash of a given string and returns its hexadecimal representation.
// This is commonly used for password hashing, data integrity checks, or unique ID generation.
//
// GenerateSHA256Hash 计算给定字符串的 SHA256 哈希值，并返回其十六进制表示。
// 这通常用于密码哈希、数据完整性检查或唯一 ID 生成。
func GenerateSHA256Hash(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

// EnsureDirExists checks if a directory exists at the given path.
// If it does not exist, it attempts to create it with the specified permissions.
// Returns an error if the directory cannot be created.
//
// EnsureDirExists 检查给定路径的目录是否存在。
// 如果不存在，它会尝试以指定权限创建该目录。
// 如果目录无法创建，则返回错误。
func EnsureDirExists(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Directory does not exist, create it
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	} else if err != nil {
		// Other error than directory not existing
		return fmt.Errorf("failed to stat directory %s: %w", dirPath, err)
	}
	return nil
}

// IsFileExists checks if a file exists at the given path.
// Returns true if the file exists and is a regular file, false otherwise.
// Returns an error if an unexpected issue occurs while checking.
//
// IsFileExists 检查给定路径的文件是否存在。
// 如果文件存在且是常规文件，则返回 true，否则返回 false。
// 如果检查时发生意外问题，则返回错误。
func IsFileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	return !info.IsDir(), nil // Return true if it's not a directory
}

// SanitizeIdentifier sanitizes a given string to be used as a database identifier (e.g., table name, column name).
// It converts the string to lowercase and replaces non-alphanumeric characters (except underscore) with underscores.
// This helps prevent SQL injection and ensures identifiers are valid for most database systems.
//
// SanitizeIdentifier 清理给定字符串，使其可以用作数据库标识符（例如，表名、列名）。
// 它将字符串转换为小写，并将非字母数字字符（下划线除外）替换为下划线。
// 这有助于防止 SQL 注入并确保标识符对大多数数据库系统有效。
func SanitizeIdentifier(identifier string) string {
	// Convert to lowercase
	s := strings.ToLower(identifier)
	// Replace non-alphanumeric (and non-underscore) characters with underscore
	var builder strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

// GetExecutableDir returns the absolute path to the directory containing the currently running executable.
// This is useful for locating configuration files, data directories, or other resources relative to the application.
//
// GetExecutableDir 返回当前运行可执行文件所在目录的绝对路径。
// 这对于查找配置文件、数据目录或相对于应用程序的其他资源非常有用。
func GetExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	return filepath.Dir(execPath), nil
}

// ContainsString checks if a string exists in a slice of strings (case-sensitive).
//
// ContainsString 检查字符串切片中是否存在某个字符串（区分大小写）。
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString removes the first occurrence of a string from a slice of strings.
// If the string is not found, the original slice is returned.
//
// RemoveString 从字符串切片中移除第一个出现的字符串。
// 如果未找到该字符串，则返回原始切片。
func RemoveString(slice []string, s string) []string {
	for i, item := range slice {
		if item == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// MinInt returns the minimum of two integers.
//
// MinInt 返回两个整数中的最小值。
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt returns the maximum of two integers.
//
// MaxInt 返回两个整数中的最大值。
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
