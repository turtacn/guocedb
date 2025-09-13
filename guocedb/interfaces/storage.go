package interfaces

import (
	"context"
	"github.com/dolthub/go-mysql-server/sql"
)

// Iterator is an interface for iterating over key-value pairs within a transaction.
type Iterator interface {
	// Next moves the iterator to the next key/value pair. It returns false when the
	// iteration is over.
	Next() bool
	// Key returns the key of the current key/value pair.
	Key() []byte
	// Value returns the value of the current key/value pair.
	Value() []byte
	// Error returns any accumulated error during iteration.
	Error() error
	// Close closes the iterator and releases associated resources.
	Close() error
}

// Transaction is an interface for database transactions. All data access within
// a storage engine should happen within a transaction.
type Transaction interface {
	// Get retrieves the value for a given key.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Set sets the value for a given key.
	Set(ctx context.Context, key, value []byte) error
	// Delete removes a key from the store.
	Delete(ctx context.Context, key []byte) error
	// Commit commits all changes made in the transaction.
	Commit(ctx context.Context) error
	// Rollback discards all changes made in the transaction.
	Rollback(ctx context.Context) error
	// NewIterator returns a new iterator for scanning key-value pairs with a given prefix.
	NewIterator(ctx context.Context, prefix []byte) (Iterator, error)
}

// Storage is the main interface for a storage engine. It abstracts the underlying
// key-value store and provides methods for managing databases, tables, and transactions.
type Storage interface {
	// NewTransaction starts a new transaction, which can be read-only or read-write.
	NewTransaction(ctx *sql.Context, readOnly bool) (Transaction, error)

	// Database management
	CreateDatabase(ctx *sql.Context, name string) error
	DropDatabase(ctx *sql.Context, name string) error

	// Table management
	CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error
	DropTable(ctx *sql.Context, dbName, tableName string) error
	GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error)

	// Close shuts down the storage engine and releases all resources.
	Close() error
}
