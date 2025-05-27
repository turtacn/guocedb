package interfaces

import (
	"context"
	"io"
	"time"

	"github.com/dolthub/go-mysql-server/sql" // For type compatibility (sql.Row, sql.Schema, sql.Type, etc.)

	"github.com/turtacn/guocedb/common/errors"
)

// Note: We aim for conceptual alignment and easy adaptation with go-mysql-server/sql interfaces,
// but define our own interfaces here to decouple the storage layer completely.
// An adapter layer will bridge these interfaces with go-mysql-server's expectations.

// --- Core Storage Engine Interface ---

// StorageEngine represents the top-level entry point to a storage system.
// It manages databases and global engine operations.
type StorageEngine interface {
	// EngineType returns a string identifier for the storage engine type (e.g., "badger", "memory").
	EngineType() string

	// Start initializes the storage engine (e.g., opens files, connects to services).
	Start(ctx context.Context) error
	// Close cleanly shuts down the storage engine.
	Close(ctx context.Context) error

	// CreateDatabase creates a new database within the engine.
	// Returns ErrCodeAlreadyExists if the database name is taken.
	CreateDatabase(ctx context.Context, tx Transaction, name string) error
	// DropDatabase removes an existing database and all its contents.
	// Returns ErrCodeNotFound if the database does not exist.
	DropDatabase(ctx context.Context, tx Transaction, name string) error
	// GetDatabase retrieves a Database handle by name.
	// Returns ErrCodeNotFound if the database does not exist.
	GetDatabase(ctx context.Context, tx Transaction, name string) (Database, error)
	// GetAllDatabaseNames returns a list of all database names managed by the engine.
	GetAllDatabaseNames(ctx context.Context, tx Transaction) ([]string, error)

	// BeginTransaction starts a new transaction.
	BeginTransaction(ctx context.Context, opts TransactionOptions) (Transaction, error)
}

// --- Transaction Interface ---

// Transaction represents an ACID transaction. Operations on Database, Table, etc.,
// are typically performed within the context of a transaction.
type Transaction interface {
	// ID returns a unique identifier for the transaction (can be opaque).
	ID() string // Or perhaps return an interface{} or specific type
	// IsReadOnly indicates if the transaction is read-only.
	IsReadOnly() bool
	// Commit attempts to commit the transaction, making its changes permanent.
	Commit(ctx context.Context) error
	// Rollback discards all changes made within the transaction.
	Rollback(ctx context.Context) error

	// Note: Compatibility with sql.Transaction is primarily through Commit/Rollback.
	// The transaction object itself is passed to other storage methods.
}

// TransactionOptions defines options for starting a new transaction.
type TransactionOptions struct {
	ReadOnly bool
	// Add other options like isolation level, timeout, etc. if needed.
	// IsolationLevel sql.TransactionIsolation // Example
}

// --- Database Interface ---

// Database represents a logical collection of tables and other schema objects.
// Compatible conceptually with sql.Database.
type Database interface {
	// Name returns the name of the database.
	Name() string
	// Engine returns the storage engine instance this database belongs to.
	Engine() StorageEngine

	// CreateTable creates a new table within the database.
	// Returns ErrCodeAlreadyExists if the table name is taken.
	CreateTable(ctx context.Context, tx Transaction, name string, schema Schema) error
	// DropTable removes an existing table and all its data.
	// Returns ErrCodeNotFound if the table does not exist.
	DropTable(ctx context.Context, tx Transaction, name string) error
	// GetTable retrieves a Table handle by name.
	// Returns ErrCodeNotFound if the table does not exist.
	GetTable(ctx context.Context, tx Transaction, name string) (Table, error)
	// GetAllTableNames returns a list of all table names within the database.
	GetAllTableNames(ctx context.Context, tx Transaction) ([]string, error)

	// Note: Methods similar to sql.Database like GetTableInsensitive might be added
	// in the adapter layer or potentially here if deemed essential for the storage layer.
}

// --- Schema Definition ---

// Schema defines the structure of a table.
// Conceptually similar to sql.Schema but potentially with storage-specific annotations.
type Schema interface {
	// Name returns the name of the schema (usually the table name).
	Name() string
	// Columns returns the definitions of the columns in the schema.
	Columns() []ColumnDefinition
	// GetColumn retrieves a column definition by name (case-insensitive). Returns nil if not found.
	GetColumn(name string) ColumnDefinition
	// PrimaryKeyColumnIndexes returns the 0-based indexes of the columns that form the primary key.
	PrimaryKeyColumnIndexes() []int
	// ToSQLSchema converts the storage schema to a sql.Schema.
	ToSQLSchema() (sql.Schema, error)

	// Maybe add methods for constraints, character sets, collations etc.
}

// ColumnDefinition defines a single column within a Schema.
// Similar to sql.Column.
type ColumnDefinition interface {
	Name() string
	Type() sql.Type // Use sql.Type for compatibility
	IsNullable() bool
	DefaultValue() *sql.ColumnDefaultValue // Use sql representation
	IsPrimaryKey() bool
	Comment() string
	// Extra() string // For things like AUTO_INCREMENT
	// Source() string // Usually the table name
	// Maybe storage-specific attributes?
}

// --- Table Interface ---

// Table represents a collection of rows conforming to a specific schema.
// Conceptually similar to sql.Table.
type Table interface {
	// Name returns the name of the table.
	Name() string
	// Database returns the database this table belongs to.
	Database() Database
	// GetSchema returns the schema definition for this table.
	GetSchema(ctx context.Context, tx Transaction) (Schema, error)

	// --- Data Manipulation ---

	// InsertRow adds a new row to the table.
	// Assumes the provided row conforms to the table's schema.
	// Returns ErrCodeConstraintViolation for PK/Unique constraint failures.
	InsertRow(ctx context.Context, tx Transaction, row Row) error
	// DeleteRow removes a row from the table identified by its primary key contained within the row data.
	// The implementation extracts the PK from the row based on the schema.
	// Returns ErrCodeNotFound if the row doesn't exist.
	DeleteRow(ctx context.Context, tx Transaction, row Row) error
	// UpdateRow updates an existing row identified by the primary key in oldRow with data from newRow.
	// Returns ErrCodeNotFound if the row doesn't exist.
	// Returns ErrCodeConstraintViolation for PK/Unique constraint failures on the new data.
	UpdateRow(ctx context.Context, tx Transaction, oldRow Row, newRow Row) error

	// --- Data Retrieval ---

	// GetRowIterator returns an iterator for scanning all rows in the table (or a specific partition).
	// This is the primary way to read data sequentially.
	// The `partitions` argument allows specifying which partitions to scan. If nil or empty, scan all.
	GetRowIterator(ctx context.Context, tx Transaction, partitions []Partition) (RowIterator, error)

	// GetPartitions returns an iterator over the table's partitions.
	// For non-partitioned tables, this might return a single "default" partition.
	// Conceptually similar to sql.Table.Partitions.
	GetPartitions(ctx context.Context, tx Transaction) (PartitionIterator, error)

	// --- Index Management ---

	// CreateIndex creates a new index on the table based on the definition.
	// Returns ErrCodeAlreadyExists if an index with the same name exists.
	CreateIndex(ctx context.Context, tx Transaction, indexDef IndexDefinition) error
	// DropIndex removes an existing index from the table.
	// Returns ErrCodeNotFound if the index does not exist.
	DropIndex(ctx context.Context, tx Transaction, indexName string) error
	// GetIndex retrieves an Index handle by name.
	// Returns ErrCodeNotFound if the index does not exist.
	GetIndex(ctx context.Context, tx Transaction, indexName string) (Index, error)
	// GetAllIndexNames returns the names of all indexes defined on the table.
	GetAllIndexNames(ctx context.Context, tx Transaction) ([]string, error)
}

// --- Row and Iterator Types ---

// Row represents a single row of data, compatible with sql.Row.
type Row = sql.Row // Alias for direct compatibility

// RowIterator iterates over rows returned by a scan or lookup.
// Compatible with sql.RowIter.
type RowIterator interface {
	// Next retrieves the next row. Returns io.EOF when iteration is complete.
	// The context is passed for potential cancellation or transaction checking.
	Next(ctx context.Context) (Row, error)
	// Close releases any resources associated with the iterator.
	Close(ctx context.Context) error
	// Schema returns the schema of the rows returned by this iterator.
	Schema() Schema // Useful for consumers to know the structure
}

// --- Partition Types ---

// Partition represents a physical or logical subdivision of a table's data.
// Compatible with sql.Partition (which is just `interface{ Key() []byte }`).
type Partition interface {
	// Key returns the unique key identifying the partition (can be opaque byte slice).
	Key() []byte
}

// PartitionIterator iterates over the partitions of a table.
// Compatible with sql.PartitionIter.
type PartitionIterator interface {
	// Next retrieves the next partition. Returns io.EOF when iteration is complete.
	Next(ctx context.Context) (Partition, error)
	// Close releases any resources associated with the iterator.
	Close(ctx context.Context) error
}

// --- Index Types ---

// IndexDefinition defines the structure and properties of an index.
// Similar to sql.Index.
type IndexDefinition interface {
	Name() string      // Index name
	TableName() string // Table this index belongs to
	// ColumnNames returns the names of the columns included in the index, in order.
	ColumnNames() []string
	IsUnique() bool
	Comment() string
	IndexType() string // e.g., "BTREE", "HASH", "FULLTEXT" (from sql.IndexType())
	// Maybe storage-specific config?
}

// Index provides accelerated access to rows based on indexed column values.
// Conceptually similar to sql.Index and combines lookup capabilities often
// found in sql.IndexLookup.
type Index interface {
	// Definition returns the structural definition of the index.
	Definition() IndexDefinition

	// Lookup returns a RowIterator that yields rows matching the provided key values.
	// The order and types of `values` must match the index's column definition.
	// This is analogous to using sql.IndexLookup but directly returns rows.
	Lookup(ctx context.Context, tx Transaction, values ...interface{}) (RowIterator, error)

	// --- Optional Advanced Index Capabilities ---
	// Implement these if the storage engine supports them efficiently.

	// RangeScan returns a RowIterator for rows within a specified key range.
	// `startKey` and `endKey` should match the structure of the index keys.
	// nil can be used for open-ended ranges.
	// RangeScan(ctx context.Context, tx Transaction, startKey Row, endKey Row, startInclusive bool, endInclusive bool) (RowIterator, error)

	// AscendRange() / DescendRange() - More specific range scans like Pebble/Badger might offer.
}

// --- Helper Functions ---

// NewStorageError is a helper to create errors specific to the storage layer.
func NewStorageError(code errors.ErrorCode, format string, args ...interface{}) error {
	// You might wrap this further or add specific storage context here if needed
	return errors.Newf(code, format, args...)
}
