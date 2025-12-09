package server

import (
	"fmt"
	"io"
	"time"

	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/turtacn/guocedb/compute/sql"
)

// BuildResult converts sql.Schema and sql.RowIter to sqltypes.Result
func BuildResult(schema sql.Schema, iter sql.RowIter) (*sqltypes.Result, error) {
	if iter == nil {
		return &sqltypes.Result{}, nil
	}

	// Build field definitions from schema
	fields := SchemaToFields(schema)

	// Collect all rows
	rows := make([][]sqltypes.Value, 0)
	for {
		row, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Convert row to SQL values
		sqlRow := RowToSQL(schema, row)
		rows = append(rows, sqlRow)
	}

	return &sqltypes.Result{
		Fields:       fields,
		Rows:         rows,
		RowsAffected: uint64(len(rows)),
	}, nil
}

// BuildOKResult creates an OK result for DML operations (INSERT/UPDATE/DELETE)
func BuildOKResult(affectedRows, lastInsertID uint64) *sqltypes.Result {
	return &sqltypes.Result{
		RowsAffected: affectedRows,
		InsertID:     lastInsertID,
	}
}

// BuildEmptyResult creates an empty result set
func BuildEmptyResult() *sqltypes.Result {
	return &sqltypes.Result{}
}

// BuildErrorResult wraps an error into a result (not typically used, errors are returned separately)
func BuildErrorResult(err error) error {
	return ConvertToMySQLError(err)
}

// SchemaToFields converts sql.Schema to query.Field array
func SchemaToFields(schema sql.Schema) []*query.Field {
	fields := make([]*query.Field, len(schema))
	for i, col := range schema {
		fields[i] = &query.Field{
			Name:         col.Name,
			Type:         col.Type.Type(),
			Table:        col.Source,
			OrgTable:     col.Source,
			Database:     "", // Not available in basic Column struct
			OrgName:      col.Name,
			ColumnLength: 255, // Default column length
			Charset:      mysql.CharacterSetUtf8,
			Flags:        columnFlags(col),
		}
	}
	return fields
}

// columnFlags determines MySQL column flags
func columnFlags(col *sql.Column) uint32 {
	var flags uint32
	
	if !col.Nullable {
		flags |= uint32(query.MySqlFlag_NOT_NULL_FLAG)
	}
	
	// Note: PrimaryKey and AutoIncrement are not available in basic Column struct
	// These would need to be extended in the sql.Column definition
	
	// Check if column type is unsigned
	if isUnsigned(col.Type) {
		flags |= uint32(query.MySqlFlag_UNSIGNED_FLAG)
	}
	
	// Check if column type is binary
	if isBinary(col.Type) {
		flags |= uint32(query.MySqlFlag_BINARY_FLAG)
	}
	
	return flags
}

// isUnsigned checks if a SQL type is unsigned
func isUnsigned(t sql.Type) bool {
	switch t.Type() {
	case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, 
		 query.Type_UINT32, query.Type_UINT64:
		return true
	}
	return false
}

// isBinary checks if a SQL type is binary
func isBinary(t sql.Type) bool {
	switch t.Type() {
	case query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB:
		return true
	}
	return false
}

// RowToSQL converts sql.Row to []sqltypes.Value
func RowToSQL(schema sql.Schema, row sql.Row) []sqltypes.Value {
	values := make([]sqltypes.Value, len(row))
	for i, val := range row {
		if i < len(schema) {
			values[i] = schema[i].Type.SQL(val)
		} else {
			// Fallback if schema doesn't match row length
			values[i] = ValueToSQL(val)
		}
	}
	return values
}

// ValueToSQL converts a Go value to sqltypes.Value
func ValueToSQL(val interface{}) sqltypes.Value {
	if val == nil {
		return sqltypes.NULL
	}

	switch v := val.(type) {
	case bool:
		if v {
			return sqltypes.NewInt64(1)
		}
		return sqltypes.NewInt64(0)
	
	case int:
		return sqltypes.NewInt64(int64(v))
	case int8:
		return sqltypes.NewInt64(int64(v))
	case int16:
		return sqltypes.NewInt64(int64(v))
	case int32:
		return sqltypes.NewInt32(v)
	case int64:
		return sqltypes.NewInt64(v)
	
	case uint:
		return sqltypes.NewUint64(uint64(v))
	case uint8:
		return sqltypes.NewUint64(uint64(v))
	case uint16:
		return sqltypes.NewUint64(uint64(v))
	case uint32:
		return sqltypes.NewUint64(uint64(v))
	case uint64:
		return sqltypes.NewUint64(v)
	
	case float32:
		return sqltypes.NewFloat64(float64(v))
	case float64:
		return sqltypes.NewFloat64(v)
	
	case string:
		return sqltypes.NewVarChar(v)
	
	case []byte:
		return sqltypes.MakeTrusted(sqltypes.Blob, v)
	
	case time.Time:
		// Format as MySQL datetime
		return sqltypes.NewVarChar(v.Format("2006-01-02 15:04:05"))
	
	default:
		// Fallback to string representation
		return sqltypes.NewVarChar(fmt.Sprintf("%v", v))
	}
}

// SQLTypeToVitessType converts sql.Type to query.Type
func SQLTypeToVitessType(t sql.Type) query.Type {
	return t.Type()
}

// BuildResultWithInfo creates a result with additional execution info
func BuildResultWithInfo(schema sql.Schema, iter sql.RowIter, info *ExecutionInfo) (*sqltypes.Result, error) {
	result, err := BuildResult(schema, iter)
	if err != nil {
		return nil, err
	}
	
	if info != nil {
		result.Info = info.String()
	}
	
	return result, nil
}

// ExecutionInfo contains additional information about query execution
type ExecutionInfo struct {
	AffectedRows uint64
	LastInsertID uint64
	Warnings     uint16
	Message      string
}

// String returns a formatted execution info message
func (e *ExecutionInfo) String() string {
	if e.Message != "" {
		return e.Message
	}
	
	if e.AffectedRows > 0 {
		if e.LastInsertID > 0 {
			return fmt.Sprintf("Rows affected: %d, Last insert ID: %d", e.AffectedRows, e.LastInsertID)
		}
		return fmt.Sprintf("Rows affected: %d", e.AffectedRows)
	}
	
	return ""
}

// ResultBuilder provides a fluent interface for building results
type ResultBuilder struct {
	result *sqltypes.Result
	err    error
}

// NewResultBuilder creates a new result builder
func NewResultBuilder() *ResultBuilder {
	return &ResultBuilder{
		result: &sqltypes.Result{},
	}
}

// WithFields sets the field definitions
func (b *ResultBuilder) WithFields(schema sql.Schema) *ResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.Fields = SchemaToFields(schema)
	return b
}

// WithRows sets the rows from an iterator
func (b *ResultBuilder) WithRows(schema sql.Schema, iter sql.RowIter) *ResultBuilder {
	if b.err != nil {
		return b
	}
	
	rows := make([][]sqltypes.Value, 0)
	for {
		row, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			b.err = err
			return b
		}
		rows = append(rows, RowToSQL(schema, row))
	}
	
	b.result.Rows = rows
	b.result.RowsAffected = uint64(len(rows))
	return b
}

// WithAffectedRows sets the affected rows count
func (b *ResultBuilder) WithAffectedRows(count uint64) *ResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.RowsAffected = count
	return b
}

// WithInsertID sets the last insert ID
func (b *ResultBuilder) WithInsertID(id uint64) *ResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.InsertID = id
	return b
}

// WithInfo sets the info message
func (b *ResultBuilder) WithInfo(info string) *ResultBuilder {
	if b.err != nil {
		return b
	}
	b.result.Info = info
	return b
}

// Build returns the final result or error
func (b *ResultBuilder) Build() (*sqltypes.Result, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.result, nil
}

// MergeResults merges multiple results into one (for multi-query support)
func MergeResults(results ...*sqltypes.Result) *sqltypes.Result {
	if len(results) == 0 {
		return BuildEmptyResult()
	}
	
	if len(results) == 1 {
		return results[0]
	}
	
	// For multiple results, return the last one (typical behavior)
	// In a real implementation, you might want to aggregate affected rows
	merged := results[len(results)-1]
	
	// Sum up affected rows from all results
	var totalAffected uint64
	for _, r := range results {
		totalAffected += r.RowsAffected
	}
	merged.RowsAffected = totalAffected
	
	return merged
}
