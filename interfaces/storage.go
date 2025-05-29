// Package interfaces defines the core interfaces for the Guocedb Storage Abstraction Layer (SAL).
// This file is crucial for decoupling the compute layer from the storage layer, allowing the
// compute layer to access underlying storage through a unified interface without concerning
// itself with the specific storage engine implementation. storage/sal/adapter.go will implement
// these interfaces, and concrete storage engines like storage/engines/badger/badger.go
// will also implement or adapt to these interfaces. compute/executor and compute/transaction
// will be the primary consumers of these interfaces.
package interfaces

import (
	"time"

	"github.com/turtacn/guocedb/common/errors"      // For unified error handling.
	"github.com/turtacn/guocedb/common/types/enum"  // For SQL types, component types, etc.
	"github.com/turtacn/guocedb/common/types/value" // For Guocedb's internal Value types.
)

// ID represents a generic identifier type, e.g., for databases, tables, or rows.
type ID uint64

// RowID represents a unique identifier for a row within a table.
type RowID ID

// ColumnID represents a unique identifier for a column within a table.
type ColumnID ID

// StorageEngine represents the top-level interface for any storage engine used by Guocedb.
// It provides methods for managing databases and initiating transactions.
type StorageEngine interface {
	// Initialize prepares the storage engine for use. This might include opening files,
	// recovering from crashes, or setting up internal structures.
	Initialize() error
	// Shutdown gracefully closes the storage engine, flushing data and releasing resources.
	Shutdown() error
	// CreateDatabase creates a new logical database with the given name.
	CreateDatabase(dbName string) error
	// DropDatabase removes an existing logical database.
	DropDatabase(dbName string) error
	// GetDatabase returns a Database interface for the given database name.
	GetDatabase(dbName string) (Database, error)
	// ListDatabases returns a list of all database names managed by the engine.
	ListDatabases() ([]string, error)
	// BeginTransaction starts a new transaction.
	BeginTransaction(isolationLevel enum.IsolationLevel) (Transaction, error)
	// GetEngineType returns the type of this storage engine (e.g., Badger, In-Memory).
	GetEngineType() enum.StorageEngineType
	// GetEngineStats returns statistics about the storage engine.
	GetEngineStats() (interface{}, error) // Generic interface for stats, specific implementations will cast.
}

// Database represents a logical database within the storage engine.
// It provides methods for managing tables.
type Database interface {
	// Name returns the name of the database.
	Name() string
	// CreateTable creates a new table within this database.
	CreateTable(tableName string, schema *TableSchema) error
	// DropTable removes an existing table from this database.
	DropTable(tableName string) error
	// GetTable returns a Table interface for the given table name.
	GetTable(tableName string) (Table, error)
	// ListTables returns a list of all table names within this database.
	ListTables() ([]string, error)
}

// TableSchema defines the schema of a table, including its columns and their types.
type TableSchema struct {
	TableName  string
	Columns    []ColumnDefinition
	PrimaryKey []string // Column names forming the primary key
	// Add indexes, constraints, etc. as needed
}

// ColumnDefinition defines a single column within a table schema.
type ColumnDefinition struct {
	Name         string
	SQLType      enum.SQLType // The SQL data type of the column.
	IsNullable   bool
	IsPrimaryKey bool
	// DefaultValue *value.Value // Optional default value
}

// Table represents a logical table within a database.
// It provides methods for data manipulation (CRUD operations).
type Table interface {
	// Name returns the name of the table.
	Name() string
	// Schema returns the schema of the table.
	Schema() *TableSchema
	// InsertRow inserts a new row into the table.
	// The order of values must match the column order in the schema.
	InsertRow(txn Transaction, values []value.Value) (RowID, error)
	// ReadRow reads a row from the table given its RowID.
	// Returns the values in the order of the table schema's columns.
	ReadRow(txn Transaction, rowID RowID) ([]value.Value, error)
	// UpdateRow updates an existing row.
	// 'updates' is a map of ColumnID to new Value.
	UpdateRow(txn Transaction, rowID RowID, updates map[ColumnID]value.Value) error
	// DeleteRow deletes a row from the table.
	DeleteRow(txn Transaction, rowID RowID) error
	// GetRowIterator returns an iterator for scanning rows.
	GetRowIterator(txn Transaction, opts *ScanOptions) (RowIterator, error)
	// GetApproxRowCount returns an approximate count of rows in the table.
	GetApproxRowCount() (int64, error)
	// GetApproxTableSize returns an approximate size of the table on disk in bytes.
	GetApproxTableSize() (int64, error)
}

// ScanOptions defines options for scanning rows (e.g., filters, projections).
type ScanOptions struct {
	// Predicate represents a filter condition (e.g., for WHERE clause).
	// This would typically be an expression tree or a callback function.
	// For now, it's a placeholder. Concrete implementations might use custom filter types.
	// Predicate interface { Eval(row []value.Value) (bool, error) }
	// For simplicity in this interface definition, we will omit a complex predicate type for now.

	// Projections specifies which columns to return. If empty, all columns.
	ProjectedColumns []ColumnID // ColumnIDs of columns to return.

	// Limit specifies the maximum number of rows to return. 0 for no limit.
	Limit int64

	// Offset specifies the number of rows to skip before starting to return.
	Offset int64

	// OrderBy specifies column IDs and sort order for sorting.
	// This would typically be handled by the compute layer after fetching.
	// For now, we assume the storage layer primarily provides raw scanning.
	// OrderBy []struct { ColumnID ColumnID; Asc bool }
}

// RowIterator allows iterating over rows in a table.
type RowIterator interface {
	// Next advances the iterator to the next row. Returns false if no more rows.
	Next() bool
	// Current returns the current RowID and the row's values.
	// Returns nil for values if Next() returned false or an error occurred.
	Current() (RowID, []value.Value, error)
	// Close releases any resources held by the iterator.
	Close() error
}

// Transaction represents a database transaction.
// It provides methods for commit, rollback, and setting read/write status.
type Transaction interface {
	// Commit attempts to commit the transaction, making all changes permanent.
	Commit() error
	// Rollback aborts the transaction, reverting all changes made within it.
	Rollback() error
	// IsReadOnly returns true if the transaction is read-only.
	IsReadOnly() bool
	// ID returns the unique identifier for this transaction.
	ID() ID
	// SetTimeout sets a timeout for the transaction. If the timeout is reached, the transaction is aborted.
	SetTimeout(d time.Duration)
}

//Personal.AI order the ending
