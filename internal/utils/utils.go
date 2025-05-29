// Package utils defines common utility functions and helper methods used internally
// within the Guocedb project. These functions are not intended for external exposure.
// This file centralizes various helpers such as string manipulations, slice operations,
// time conversions, and file path handling, reducing code duplication and improving
// development efficiency across internal modules.
package utils

import (
	"crypto/rand"   // For secure random number generation.
	"encoding/hex"  // For encoding random bytes to string.
	"fmt"           // For error formatting.
	"io"            // For io.CopyN.
	"os"            // For file operations.
	"path/filepath" // For path manipulation.
	"runtime"       // For getting caller information (e.g., for stack traces).
	"strings"       // For string manipulation.
	"time"          // For time-related utilities.

	"github.com/turtacn/guocedb/common/errors"     // For unified error handling.
	"github.com/turtacn/guocedb/common/types/enum" // For component types in errors/logging.
)

// GenerateRandomBytes generates a cryptographically secure random byte slice of the given length.
// This is useful for salt generation, session tokens, etc.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrSecurity, errors.CodeEncryptionFailed,
			fmt.Sprintf("failed to generate random bytes of length %d", n), err)
	}
	return b, nil
}

// GenerateRandomHex generates a cryptographically secure random hexadecimal string of the given length.
// The resulting string will be twice the length of the byte count (e.g., 16 bytes -> 32 hex chars).
func GenerateRandomHex(n int) (string, error) {
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ClampInt64 clamps an int64 value between a minimum and maximum boundary.
func ClampInt64(val, min, max int64) int64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ContainsString checks if a string exists in a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString removes the first occurrence of a string from a slice of strings.
func RemoveString(slice []string, s string) []string {
	for i, item := range slice {
		if item == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// ConvertDuration parses a duration string (e.g., "5s", "1m", "1h") into time.Duration.
// It returns an error if the string format is invalid.
func ConvertDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("invalid duration string '%s'", s), err)
	}
	return d, nil
}

// PathExists checks if a file or directory exists at the given path.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError,
		fmt.Sprintf("failed to check path existence for %s", path), err)
}

// IsDir checks if the given path is a directory.
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path does not exist, so it's not a directory.
		}
		return false, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError,
			fmt.Sprintf("failed to get file info for %s", path), err)
	}
	return info.IsDir(), nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError,
			fmt.Sprintf("failed to open source file %s", src), err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError,
			fmt.Sprintf("failed to create destination file %s", dst), err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError,
			fmt.Sprintf("failed to copy file from %s to %s", src, dst), err)
	}
	return nil
}

// GetCallerInfo returns the file name and line number of the caller.
// Useful for debugging or augmenting log messages with source location.
func GetCallerInfo(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip + 1) // skip + 1 to get the actual caller of GetCallerInfo
	if !ok {
		return "unknown", 0
	}
	return filepath.Base(file), line // Return only the base file name.
}

// CamelToSnakeCase converts a camelCase string to snake_case.
func CamelToSnakeCase(s string) string {
	var builder strings.Builder
	for i, r := range s {
		if 'A' <= r && r <= 'Z' {
			if i > 0 {
				builder.WriteRune('_')
			}
			builder.WriteRune(r + 32) // Convert to lowercase
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// SnakeToCamelCase converts a snake_case string to camelCase.
func SnakeToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	var builder strings.Builder
	for i, part := range parts {
		if i == 0 {
			builder.WriteString(part)
		} else {
			if len(part) > 0 {
				builder.WriteString(strings.ToUpper(string(part[0])))
				builder.WriteString(part[1:])
			}
		}
	}
	return builder.String()
}

// CalculateCRC32 calculates the CRC-32 checksum of a byte slice.
// This can be used for data integrity checks.
/*
import "hash/crc32" // Uncomment this import for CalculateCRC32

func CalculateCRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
*/

// BytesToHex converts a byte slice to its hexadecimal string representation.
func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToBytes converts a hexadecimal string to a byte slice.
func HexToBytes(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("invalid hex string '%s'", s), err)
	}
	return b, nil
}

//Personal.AI order the ending
