package server

import (
	"fmt"
	"strings"

	"github.com/dolthub/vitess/go/mysql"
	"github.com/turtacn/guocedb/compute/sql"
	"gopkg.in/src-d/go-errors.v1"
)

// MySQL error code constants
// Reference: https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
const (
	// ERAccessDeniedError - Access denied for user
	ERAccessDeniedError = 1045
	// ERNoDB - No database selected
	ERNoDB = 1046
	// ERBadDB - Unknown database
	ERBadDB = 1049
	// ERBadTable - Unknown table
	ERBadTable = 1051
	// ERBadField - Unknown column
	ERBadField = 1054
	// ERDupEntry - Duplicate entry for key
	ERDupEntry = 1062
	// ERParseError - SQL syntax error
	ERParseError = 1064
	// ERNoSuchTable - Table doesn't exist
	ERNoSuchTable = 1146
	// ERWrongValueCountOnRow - Column count doesn't match value count
	ERWrongValueCountOnRow = 1136
	// ERLockDeadlock - Deadlock found when trying to get lock
	ERLockDeadlock = 1213
	// ERUnknownError - Unknown error
	ERUnknownError = 1105
	// ERUnknownComError - Unknown command
	ERUnknownComError = 1047
	// ERAlreadyExists - Can't create database; database exists
	ERAlreadyExists = 1007
	// ERTableExistsError - Table already exists
	ERTableExistsError = 1050
)

// SQL State constants
const (
	// SSUnknownSQLState - Generic unknown state
	SSUnknownSQLState = "HY000"
	// SSClientError - Client error
	SSClientError = "42000"
	// SSNoDatabase - No database selected
	SSNoDatabase = "3D000"
	// SSNoSuchTable - Table doesn't exist
	SSNoSuchTable = "42S02"
	// SSBadField - Unknown column
	SSBadField = "42S22"
	// SSDupEntry - Duplicate entry
	SSDupEntry = "23000"
	// SSDeadlock - Deadlock
	SSDeadlock = "40001"
	// SSAccessDenied - Access denied
	SSAccessDenied = "28000"
)

// ConvertToMySQLError converts internal errors to MySQL protocol errors
func ConvertToMySQLError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific error types
	switch {
	case sql.ErrDatabaseNotFound.Is(err):
		msg := extractErrorMessage(err, "Unknown database")
		return mysql.NewSQLError(ERBadDB, SSClientError, "%s", msg)
	
	case sql.ErrTableNotFound.Is(err):
		msg := extractErrorMessage(err, "Table doesn't exist")
		return mysql.NewSQLError(ERNoSuchTable, SSNoSuchTable, "%s", msg)
	
	case sql.ErrTableAlreadyExists.Is(err):
		msg := extractErrorMessage(err, "Table already exists")
		return mysql.NewSQLError(ERTableExistsError, SSClientError, "%s", msg)
	
	case sql.ErrDatabaseExists.Is(err):
		msg := extractErrorMessage(err, "Database already exists")
		return mysql.NewSQLError(ERAlreadyExists, SSClientError, "%s", msg)
	
	case sql.ErrInvalidType.Is(err):
		msg := extractErrorMessage(err, "Invalid type")
		return mysql.NewSQLError(ERBadField, SSBadField, "%s", msg)
	
	case sql.ErrUnexpectedRowLength.Is(err):
		msg := extractErrorMessage(err, "Column count doesn't match")
		return mysql.NewSQLError(ERWrongValueCountOnRow, SSClientError, "%s", msg)
	
	case isParseError(err):
		msg := extractErrorMessage(err, "SQL syntax error")
		return mysql.NewSQLError(ERParseError, SSClientError, "%s", msg)
	
	case isDuplicateKeyError(err):
		msg := extractErrorMessage(err, "Duplicate entry")
		return mysql.NewSQLError(ERDupEntry, SSDupEntry, "%s", msg)
	
	case isDeadlockError(err):
		msg := extractErrorMessage(err, "Deadlock found")
		return mysql.NewSQLError(ERLockDeadlock, SSDeadlock, "%s", msg)
	
	case isAccessDeniedError(err):
		msg := extractErrorMessage(err, "Access denied")
		return mysql.NewSQLError(ERAccessDeniedError, SSAccessDenied, "%s", msg)
	}

	// Check if it's already a MySQL error
	if _, ok := err.(*mysql.SQLError); ok {
		return err
	}

	// Default to generic error
	return mysql.NewSQLError(ERUnknownError, SSUnknownSQLState, "%s", err.Error())
}

// extractErrorMessage extracts the message from an error or returns a default
func extractErrorMessage(err error, defaultMsg string) string {
	if err == nil {
		return defaultMsg
	}
	msg := err.Error()
	if msg == "" {
		return defaultMsg
	}
	return msg
}

// isParseError checks if the error is a parse/syntax error
func isParseError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "syntax") ||
		strings.Contains(msg, "parse") ||
		strings.Contains(msg, "unexpected")
}

// isDuplicateKeyError checks if the error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique constraint")
}

// isDeadlockError checks if the error is a deadlock error
func isDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock")
}

// isAccessDeniedError checks if the error is an access denied error
func isAccessDeniedError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "unauthorized")
}

// NewSQLError creates a new MySQL error with the given code, state, and message
func NewSQLError(code int, state string, format string, args ...interface{}) error {
	return mysql.NewSQLError(code, state, format, args...)
}

// NewParseError creates a parse error with proper MySQL error code
func NewParseError(format string, args ...interface{}) error {
	return mysql.NewSQLError(ERParseError, SSClientError, format, args...)
}

// NewDatabaseNotFoundError creates a database not found error
func NewDatabaseNotFoundError(dbName string) error {
	return mysql.NewSQLError(ERBadDB, SSClientError, "Unknown database '%s'", dbName)
}

// NewTableNotFoundError creates a table not found error
func NewTableNotFoundError(tableName string) error {
	return mysql.NewSQLError(ERNoSuchTable, SSNoSuchTable, "Table '%s' doesn't exist", tableName)
}

// NewColumnNotFoundError creates a column not found error
func NewColumnNotFoundError(columnName string) error {
	return mysql.NewSQLError(ERBadField, SSBadField, "Unknown column '%s'", columnName)
}

// NewDuplicateKeyError creates a duplicate key error
func NewDuplicateKeyError(key string) error {
	return mysql.NewSQLError(ERDupEntry, SSDupEntry, "Duplicate entry '%s'", key)
}

// NewAccessDeniedError creates an access denied error
func NewAccessDeniedError(user string) error {
	return mysql.NewSQLError(ERAccessDeniedError, SSAccessDenied, "Access denied for user '%s'", user)
}

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	
	// If it's already a MySQL error, preserve the code but add context
	if sqlErr, ok := err.(*mysql.SQLError); ok {
		return mysql.NewSQLError(sqlErr.Num, sqlErr.State, "%s: %s", context, sqlErr.Message)
	}
	
	// Otherwise convert and add context
	mysqlErr := ConvertToMySQLError(err)
	if sqlErr, ok := mysqlErr.(*mysql.SQLError); ok {
		return mysql.NewSQLError(sqlErr.Num, sqlErr.State, "%s: %s", context, sqlErr.Message)
	}
	
	return fmt.Errorf("%s: %w", context, err)
}

// IsErrorCode checks if an error matches a specific MySQL error code
func IsErrorCode(err error, code int) bool {
	if err == nil {
		return false
	}
	sqlErr, ok := err.(*mysql.SQLError)
	if !ok {
		return false
	}
	return sqlErr.Num == code
}

// ConvertGoErrorKind converts go-errors.v1 Kind errors to MySQL errors
func ConvertGoErrorKind(err error) error {
	if err == nil {
		return nil
	}
	
	// Check if it's a go-errors Kind
	if kindErr, ok := err.(*errors.Error); ok {
		msg := kindErr.Error()
		// Try to determine the appropriate MySQL error based on message
		switch {
		case strings.Contains(msg, "database not found"):
			return mysql.NewSQLError(ERBadDB, SSClientError, "%s", msg)
		case strings.Contains(msg, "table not found"):
			return mysql.NewSQLError(ERNoSuchTable, SSNoSuchTable, "%s", msg)
		case strings.Contains(msg, "already exists"):
			return mysql.NewSQLError(ERTableExistsError, SSClientError, "%s", msg)
		default:
			return mysql.NewSQLError(ERUnknownError, SSUnknownSQLState, "%s", msg)
		}
	}
	
	return ConvertToMySQLError(err)
}
