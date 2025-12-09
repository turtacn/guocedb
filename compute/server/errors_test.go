package server

import (
	"errors"
	"testing"

	"github.com/dolthub/vitess/go/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/sql"
)

func TestConvertToMySQLError_DatabaseNotFound(t *testing.T) {
	err := sql.ErrDatabaseNotFound.New("testdb")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERBadDB, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "testdb")
}

func TestConvertToMySQLError_TableNotFound(t *testing.T) {
	err := sql.ErrTableNotFound.New("users")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERNoSuchTable, sqlErr.Num)
	assert.Equal(t, SSNoSuchTable, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "users")
}

func TestConvertToMySQLError_TableAlreadyExists(t *testing.T) {
	err := sql.ErrTableAlreadyExists.New("users")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERTableExistsError, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
}

func TestConvertToMySQLError_DatabaseExists(t *testing.T) {
	err := sql.ErrDatabaseExists.New("testdb")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERAlreadyExists, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
}

func TestConvertToMySQLError_ParseError(t *testing.T) {
	err := errors.New("syntax error near 'SELEC'")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERParseError, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
}

func TestConvertToMySQLError_DuplicateKey(t *testing.T) {
	err := errors.New("duplicate entry for key PRIMARY")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERDupEntry, sqlErr.Num)
	assert.Equal(t, SSDupEntry, sqlErr.State)
}

func TestConvertToMySQLError_Deadlock(t *testing.T) {
	err := errors.New("deadlock detected")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERLockDeadlock, sqlErr.Num)
	assert.Equal(t, SSDeadlock, sqlErr.State)
}

func TestConvertToMySQLError_AccessDenied(t *testing.T) {
	err := errors.New("access denied for user")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERAccessDeniedError, sqlErr.Num)
	assert.Equal(t, SSAccessDenied, sqlErr.State)
}

func TestConvertToMySQLError_GenericError(t *testing.T) {
	err := errors.New("some random error")
	mysqlErr := ConvertToMySQLError(err)
	
	require.NotNil(t, mysqlErr)
	sqlErr, ok := mysqlErr.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERUnknownError, sqlErr.Num)
	assert.Equal(t, SSUnknownSQLState, sqlErr.State)
}

func TestConvertToMySQLError_Nil(t *testing.T) {
	mysqlErr := ConvertToMySQLError(nil)
	assert.Nil(t, mysqlErr)
}

func TestConvertToMySQLError_AlreadyMySQLError(t *testing.T) {
	original := mysql.NewSQLError(1234, "ABC12", "test error")
	result := ConvertToMySQLError(original)
	
	assert.Equal(t, original, result)
}

func TestNewDatabaseNotFoundError(t *testing.T) {
	err := NewDatabaseNotFoundError("testdb")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERBadDB, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "testdb")
}

func TestNewTableNotFoundError(t *testing.T) {
	err := NewTableNotFoundError("users")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERNoSuchTable, sqlErr.Num)
	assert.Equal(t, SSNoSuchTable, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "users")
}

func TestNewColumnNotFoundError(t *testing.T) {
	err := NewColumnNotFoundError("email")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERBadField, sqlErr.Num)
	assert.Equal(t, SSBadField, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "email")
}

func TestNewDuplicateKeyError(t *testing.T) {
	err := NewDuplicateKeyError("PRIMARY")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERDupEntry, sqlErr.Num)
	assert.Equal(t, SSDupEntry, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "PRIMARY")
}

func TestNewAccessDeniedError(t *testing.T) {
	err := NewAccessDeniedError("root")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERAccessDeniedError, sqlErr.Num)
	assert.Equal(t, SSAccessDenied, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "root")
}

func TestNewParseError(t *testing.T) {
	err := NewParseError("syntax error near '%s'", "SELECT")
	
	require.NotNil(t, err)
	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERParseError, sqlErr.Num)
	assert.Equal(t, SSClientError, sqlErr.State)
	assert.Contains(t, sqlErr.Message, "SELECT")
}

func TestIsErrorCode(t *testing.T) {
	err := mysql.NewSQLError(ERBadDB, SSClientError, "test")
	
	assert.True(t, IsErrorCode(err, ERBadDB))
	assert.False(t, IsErrorCode(err, ERNoSuchTable))
	assert.False(t, IsErrorCode(nil, ERBadDB))
	assert.False(t, IsErrorCode(errors.New("test"), ERBadDB))
}

func TestWrapError(t *testing.T) {
	original := sql.ErrDatabaseNotFound.New("testdb")
	wrapped := WrapError(original, "failed to connect")
	
	require.NotNil(t, wrapped)
	sqlErr, ok := wrapped.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERBadDB, sqlErr.Num)
	assert.Contains(t, sqlErr.Message, "failed to connect")
	assert.Contains(t, sqlErr.Message, "testdb")
}

func TestWrapError_Nil(t *testing.T) {
	wrapped := WrapError(nil, "context")
	assert.Nil(t, wrapped)
}

func TestWrapError_MySQLError(t *testing.T) {
	original := mysql.NewSQLError(ERBadDB, SSClientError, "database not found")
	wrapped := WrapError(original, "connection failed")
	
	require.NotNil(t, wrapped)
	sqlErr, ok := wrapped.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, ERBadDB, sqlErr.Num)
	assert.Contains(t, sqlErr.Message, "connection failed")
}

func TestIsParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"syntax error", errors.New("syntax error near SELECT"), true},
		{"parse error", errors.New("parse error: unexpected token"), true},
		{"unexpected token", errors.New("unexpected EOF"), true},
		{"other error", errors.New("connection refused"), false},
		{"nil", nil, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isParseError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"duplicate", errors.New("duplicate entry"), true},
		{"unique constraint", errors.New("unique constraint violation"), true},
		{"other error", errors.New("connection refused"), false},
		{"nil", nil, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDuplicateKeyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDeadlockError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"deadlock", errors.New("deadlock detected"), true},
		{"other error", errors.New("timeout"), false},
		{"nil", nil, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDeadlockError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAccessDeniedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"access denied", errors.New("access denied for user"), true},
		{"permission denied", errors.New("permission denied"), true},
		{"unauthorized", errors.New("unauthorized access"), true},
		{"other error", errors.New("connection refused"), false},
		{"nil", nil, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAccessDeniedError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
