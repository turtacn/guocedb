// Package interfaces defines the core abstraction interfaces for guocedb.
package interfaces

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// Storage is the main interface for a storage engine.
// It provides methods for accessing and managing data and transactions.
type Storage interface {
	// Get retrieves a value for a given key from a specific table.
	Get(ctx *sql.Context, db, table string, key []byte) ([]byte, error)

	// Set stores a key-value pair in a specific table.
	Set(ctx *sql.Context, db, table string, key, value []byte) error

	// Delete removes a key from a specific table.
	Delete(ctx *sql.Context, db, table string, key []byte) error

	// Iterator returns an iterator for a given key prefix in a table.
	Iterator(ctx *sql.Context, db, table string, prefix []byte) (Iterator, error)

	// NewTransaction creates a new transaction.
	NewTransaction(ctx *sql.Context, readOnly bool) (Transaction, error)

	// Database management
	CreateDatabase(ctx *sql.Context, name string) error
	DropDatabase(ctx *sql.Context, name string) error
	ListDatabases(ctx *sql.Context) ([]string, error)

	// Table management
	CreateTable(ctx *sql.Context, dbName string, table sql.Table) error
	DropTable(ctx *sql.Context, dbName, tableName string) error
	GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error)
	ListTables(ctx *sql.Context, dbName string) ([]string, error)

	// Close shuts down the storage engine gracefully.
	Close() error
}

// Transaction represents a transaction in the storage engine.
// All read and write operations within a transaction should be atomic.
type Transaction interface {
	// Get retrieves a value for a given key.
	Get(key []byte) ([]byte, error)

	// Set stores a key-value pair.
	Set(key, value []byte) error

	// Delete removes a key.
	Delete(key []byte) error

	// Iterator returns an iterator for a given key prefix.
	Iterator(prefix []byte) (Iterator, error)

	// Commit finalizes the transaction.
	Commit() error

	// Rollback aborts the transaction.
	Rollback() error

	// IsReadOnly returns true if the transaction is read-only.
	IsReadOnly() bool
}

// Iterator is used to iterate over a range of key-value pairs.
type Iterator interface {
	// Next moves the iterator to the next key-value pair.
	// It returns false when the iteration is finished.
	Next() bool

	// Key returns the current key.
	Key() []byte

	// Value returns the current value.
	Value() []byte

	// Error returns any error that occurred during iteration.
	Error() error

	// Close releases the resources associated with the iterator.
	Close() error
}
