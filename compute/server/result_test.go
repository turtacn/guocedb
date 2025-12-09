package server

import (
	"io"
	"testing"
	"time"

	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/sql"
)

func TestBuildResult_EmptyIterator(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
		{Name: "name", Type: sql.Text},
	}
	
	iter := &mockRowIter{rows: []sql.Row{}}
	
	result, err := BuildResult(schema, iter)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Fields, 2)
	assert.Len(t, result.Rows, 0)
	assert.Equal(t, uint64(0), result.RowsAffected)
}

func TestBuildResult_WithRows(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
		{Name: "name", Type: sql.Text},
	}
	
	iter := &mockRowIter{
		rows: []sql.Row{
			{int32(1), "Alice"},
			{int32(2), "Bob"},
		},
	}
	
	result, err := BuildResult(schema, iter)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Fields, 2)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, uint64(2), result.RowsAffected)
}

func TestBuildResult_NilIterator(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
	}
	
	result, err := BuildResult(schema, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Rows, 0)
}

func TestBuildOKResult(t *testing.T) {
	result := BuildOKResult(5, 123)
	
	require.NotNil(t, result)
	assert.Equal(t, uint64(5), result.RowsAffected)
	assert.Equal(t, uint64(123), result.InsertID)
}

func TestBuildEmptyResult(t *testing.T) {
	result := BuildEmptyResult()
	
	require.NotNil(t, result)
	assert.Equal(t, uint64(0), result.RowsAffected)
	assert.Len(t, result.Rows, 0)
}

func TestSchemaToFields(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32, Nullable: false, Source: "users"},
		{Name: "email", Type: sql.Text, Nullable: true, Source: "users"},
	}
	
	fields := SchemaToFields(schema)
	
	require.Len(t, fields, 2)
	
	// Check first field
	assert.Equal(t, "id", fields[0].Name)
	assert.Equal(t, "users", fields[0].Table)
	assert.Equal(t, "id", fields[0].OrgName)
	
	// Check second field
	assert.Equal(t, "email", fields[1].Name)
	assert.Equal(t, "users", fields[1].Table)
}

func TestRowToSQL(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
		{Name: "name", Type: sql.Text},
		{Name: "active", Type: sql.Boolean},
	}
	
	row := sql.Row{int32(1), "Alice", true}
	sqlRow := RowToSQL(schema, row)
	
	require.Len(t, sqlRow, 3)
	assert.NotEqual(t, sqltypes.NULL, sqlRow[0])
	assert.NotEqual(t, sqltypes.NULL, sqlRow[1])
	assert.NotEqual(t, sqltypes.NULL, sqlRow[2])
}

func TestValueToSQL_Int(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"int", int(42)},
		{"int8", int8(42)},
		{"int16", int16(42)},
		{"int32", int32(42)},
		{"int64", int64(42)},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValueToSQL(tt.value)
			assert.NotEqual(t, sqltypes.NULL, result)
		})
	}
}

func TestValueToSQL_Uint(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"uint", uint(42)},
		{"uint8", uint8(42)},
		{"uint16", uint16(42)},
		{"uint32", uint32(42)},
		{"uint64", uint64(42)},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValueToSQL(tt.value)
			assert.NotEqual(t, sqltypes.NULL, result)
		})
	}
}

func TestValueToSQL_Float(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"float32", float32(3.14)},
		{"float64", float64(3.14)},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValueToSQL(tt.value)
			assert.NotEqual(t, sqltypes.NULL, result)
		})
	}
}

func TestValueToSQL_String(t *testing.T) {
	result := ValueToSQL("hello")
	assert.NotEqual(t, sqltypes.NULL, result)
}

func TestValueToSQL_Bytes(t *testing.T) {
	result := ValueToSQL([]byte("hello"))
	assert.NotEqual(t, sqltypes.NULL, result)
}

func TestValueToSQL_Time(t *testing.T) {
	now := time.Now()
	result := ValueToSQL(now)
	assert.NotEqual(t, sqltypes.NULL, result)
}

func TestValueToSQL_Bool(t *testing.T) {
	resultTrue := ValueToSQL(true)
	resultFalse := ValueToSQL(false)
	
	assert.NotEqual(t, sqltypes.NULL, resultTrue)
	assert.NotEqual(t, sqltypes.NULL, resultFalse)
}

func TestValueToSQL_Nil(t *testing.T) {
	result := ValueToSQL(nil)
	assert.Equal(t, sqltypes.NULL, result)
}

func TestResultBuilder_Simple(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
	}
	
	builder := NewResultBuilder()
	result, err := builder.
		WithFields(schema).
		WithAffectedRows(10).
		WithInsertID(5).
		Build()
	
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Fields, 1)
	assert.Equal(t, uint64(10), result.RowsAffected)
	assert.Equal(t, uint64(5), result.InsertID)
}

func TestResultBuilder_WithRows(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
		{Name: "name", Type: sql.Text},
	}
	
	iter := &mockRowIter{
		rows: []sql.Row{
			{int32(1), "Alice"},
			{int32(2), "Bob"},
		},
	}
	
	builder := NewResultBuilder()
	result, err := builder.
		WithFields(schema).
		WithRows(schema, iter).
		Build()
	
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, uint64(2), result.RowsAffected)
}

func TestResultBuilder_WithInfo(t *testing.T) {
	builder := NewResultBuilder()
	result, err := builder.
		WithInfo("Query executed successfully").
		Build()
	
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Query executed successfully", result.Info)
}

func TestResultBuilder_Error(t *testing.T) {
	schema := sql.Schema{
		{Name: "id", Type: sql.Int32},
	}
	
	iter := &mockRowIterWithError{}
	
	builder := NewResultBuilder()
	result, err := builder.
		WithFields(schema).
		WithRows(schema, iter).
		Build()
	
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestMergeResults_Empty(t *testing.T) {
	result := MergeResults()
	require.NotNil(t, result)
	assert.Equal(t, uint64(0), result.RowsAffected)
}

func TestMergeResults_Single(t *testing.T) {
	r1 := &sqltypes.Result{RowsAffected: 5}
	result := MergeResults(r1)
	
	require.NotNil(t, result)
	assert.Equal(t, uint64(5), result.RowsAffected)
}

func TestMergeResults_Multiple(t *testing.T) {
	r1 := &sqltypes.Result{RowsAffected: 5}
	r2 := &sqltypes.Result{RowsAffected: 3}
	r3 := &sqltypes.Result{RowsAffected: 2}
	
	result := MergeResults(r1, r2, r3)
	
	require.NotNil(t, result)
	assert.Equal(t, uint64(10), result.RowsAffected)
}

func TestExecutionInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		info     *ExecutionInfo
		expected string
	}{
		{
			name: "with message",
			info: &ExecutionInfo{Message: "custom message"},
			expected: "custom message",
		},
		{
			name: "with affected rows and insert id",
			info: &ExecutionInfo{AffectedRows: 5, LastInsertID: 123},
			expected: "Rows affected: 5, Last insert ID: 123",
		},
		{
			name: "with affected rows only",
			info: &ExecutionInfo{AffectedRows: 3},
			expected: "Rows affected: 3",
		},
		{
			name: "empty",
			info: &ExecutionInfo{},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColumnFlags_NotNullable(t *testing.T) {
	col := &sql.Column{
		Name:     "id",
		Type:     sql.Int32,
		Nullable: false,
	}
	
	flags := columnFlags(col)
	// Should have NOT_NULL flag set
	assert.NotEqual(t, uint32(0), flags)
}

func TestColumnFlags_Nullable(t *testing.T) {
	col := &sql.Column{
		Name:     "email",
		Type:     sql.Text,
		Nullable: true,
	}
	
	flags := columnFlags(col)
	// Might have other flags, but checking it doesn't panic
	assert.NotNil(t, flags)
}

func TestIsUnsigned(t *testing.T) {
	// Test with a type that should be unsigned
	// Note: This depends on the sql.Type implementation
	// Just verify the function doesn't panic
	result := isUnsigned(sql.Int32)
	assert.False(t, result) // Int32 is signed
}

func TestIsBinary(t *testing.T) {
	// Test with a type
	// Just verify the function doesn't panic
	result := isBinary(sql.Text)
	assert.False(t, result) // Text is not binary
}

// mockRowIterWithError is a mock that returns an error
type mockRowIterWithError struct{}

func (m *mockRowIterWithError) Next() (sql.Row, error) {
	return nil, io.ErrUnexpectedEOF
}

func (m *mockRowIterWithError) Close() error {
	return nil
}
