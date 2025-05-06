// Package errors defines unified error types and handling functions.
// errors 包定义了统一的错误类型和处理函数。
package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrConfigLoadFailed indicates failure to load configuration.
	// ErrConfigLoadFailed 表示加载配置失败。
	ErrConfigLoadFailed = NewErrorType("config load failed")

	// ErrStorageEngineNotFound indicates that the specified storage engine was not found.
	// ErrStorageEngineNotFound 表示找不到指定的存储引擎。
	ErrStorageEngineNotFound = NewErrorType("storage engine not found")

	// ErrDatabaseNotFound indicates that the specified database does not exist.
	// ErrDatabaseNotFound 表示指定的数据库不存在。
	ErrDatabaseNotFound = NewErrorType("database not found")

	// ErrTableNotFound indicates that the specified table does not exist.
	// ErrTableNotFound 表示指定的表不存在。
	ErrTableNotFound = NewErrorType("table not found")

	// ErrColumnNotFound indicates that the specified column does not exist.
	// ErrColumnNotFound 表示指定的列不存在。
	ErrColumnNotFound = NewErrorType("column not found")

	// ErrIndexNotFound indicates that the specified index does not exist.
	// ErrIndexNotFound 表示指定的索引不存在。
	ErrIndexNotFound = NewErrorType("index not found")

	// ErrTableAlreadyExists indicates that a table with the same name already exists.
	// ErrTableAlreadyExists 表示已存在同名表。
	ErrTableAlreadyExists = NewErrorType("table already exists")

	// ErrDatabaseAlreadyExists indicates that a database with the same name already exists.
	// ErrDatabaseAlreadyExists 表示已存在同名数据库。
	ErrDatabaseAlreadyExists = NewErrorType("database already exists")

	// ErrInvalidSQL indicates that the SQL statement is invalid.
	// ErrInvalidSQL 表示 SQL 语句无效。
	ErrInvalidSQL = NewErrorType("invalid SQL statement")

	// ErrPermissionDenied indicates that the user does not have permission to perform the action.
	// ErrPermissionDenied 表示用户无权执行此操作。
	ErrPermissionDenied = NewErrorType("permission denied")

	// ErrNotImplemented indicates that a feature is not yet implemented.
	// ErrNotImplemented 表示某个功能尚未实现。
	ErrNotImplemented = NewErrorType("not implemented")

	// ErrTransactionCommitFailed indicates failure to commit a transaction.
	// ErrTransactionCommitFailed 表示提交事务失败。
	ErrTransactionCommitFailed = NewErrorType("transaction commit failed")

	// ErrTransactionRollbackFailed indicates failure to rollback a transaction.
	// ErrTransactionRollbackFailed 表示回滚事务失败。
	ErrTransactionRollbackFailed = NewErrorType("transaction rollback failed")

	// ErrPrimaryKeyRequired indicates that a primary key is required but missing.
	// ErrPrimaryKeyRequired 表示需要主键但缺失。
	ErrPrimaryKeyRequired = NewErrorType("primary key required")

	// ErrCorruptedData indicates that stored data is corrupted.
	// ErrCorruptedData 表示存储的数据已损坏。
	ErrCorruptedData = NewErrorType("corrupted data")

	// ErrEncodingFailed indicates failure during data encoding.
	// ErrEncodingFailed 表示数据编码失败。
	ErrEncodingFailed = NewErrorType("data encoding failed")

	// ErrDecodingFailed indicates failure during data decoding.
	// ErrDecodingFailed 表示数据解码失败。
	ErrDecodingFailed = NewErrorType("data decoding failed")

	// ErrBadgerOperationFailed indicates a failure during a Badger operation.
	// ErrBadgerOperationFailed 表示 Badger 操作失败。
	ErrBadgerOperationFailed = NewErrorType("badger operation failed")

	// ErrCatalogOperationFailed indicates a failure during a catalog operation.
	// ErrCatalogOperationFailed 表示目录操作失败。
	ErrCatalogOperationFailed = NewErrorType("catalog operation failed")

	// ErrInternal indicates an unexpected internal error.
	// ErrInternal 表示发生意外的内部错误。
	ErrInternal = NewErrorType("internal error")

	// ErrNetwork indicates a network-related error.
	// ErrNetwork 表示网络相关错误。
	ErrNetwork = NewErrorType("network error")

	// ErrAuthenticationFailed indicates authentication failure.
	// ErrAuthenticationFailed 表示认证失败。
	ErrAuthenticationFailed = NewErrorType("authentication failed")

	// ErrAuthorizationFailed indicates authorization failure.
	// ErrAuthorizationFailed 表示授权失败。
	ErrAuthorizationFailed = NewErrorType("authorization failed")

	// ErrMaintenanceOperationFailed indicates failure during a maintenance operation.
	// ErrMaintenanceOperationFailed 表示维护操作失败。
	ErrMaintenanceOperationFailed = NewErrorType("maintenance operation failed")
)

// ErrorType represents a distinct category of errors.
// ErrorType 代表一种独特的错误类别。
type ErrorType string

// New creates a new error with a message derived from the ErrorType and optional arguments.
// New 创建一个新错误，其消息源自 ErrorType 和可选参数。
func (e ErrorType) New(args ...interface{}) error {
	msg := string(e)
	if len(args) > 0 {
		msg = fmt.Sprintf(msg+": %v", args...)
	}
	return fmt.Errorf(msg)
}

// NewErrorType creates a new ErrorType.
// NewErrorType 创建一个新的 ErrorType。
func NewErrorType(msg string) ErrorType {
	return ErrorType(msg)
}

// Is checks if an error is of a specific ErrorType.
// Is 检查一个错误是否属于特定的 ErrorType。
func Is(err error, errType ErrorType) bool {
	if err == nil {
		return false
	}
	// Unwrap until the root cause is found or no more wrapping is possible
	for err != nil {
		if fmt.Sprintf("%v", err) == string(errType) {
			return true
		}
		// Check if the error message starts with the error type string
		if len(fmt.Sprintf("%v", err)) >= len(string(errType)) && fmt.Sprintf("%v", err)[:len(string(errType))] == string(errType) {
			return true
		}
		// Check if the underlying error is of this type (more robust approach needed for custom error types)
		// For now, relying on string comparison or wrapped errors.Is
		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			break
		}
		err = unwrappedErr
	}
	return false
}

// IsAny checks if an error matches any of the provided ErrorTypes.
// IsAny 检查一个错误是否与提供的任何 ErrorType 匹配。
func IsAny(err error, errTypes ...ErrorType) bool {
	for _, errType := range errTypes {
		if Is(err, errType) {
			return true
		}
	}
	return false
}