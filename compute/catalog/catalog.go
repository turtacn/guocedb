// Package catalog defines the core interface for the metadata catalog in Guocedb.
// This includes management operations for metadata objects like databases, tables, columns, and indexes.
// This interface is crucial for the compute layer's semantic analysis, query optimization, and execution.
// compute/catalog/memory/memory_catalog.go and compute/catalog/persistent/persistent_catalog.go
// will implement this interface. It will rely on data types defined in common/types/value/value.go
// and potentially common/errors.
package catalog

import (
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/interfaces" // Assuming interfaces.ID and interfaces.RowID etc. are here
)

// ColumnDefinition represents the schema of a single column in a table.
type ColumnDefinition struct {
	ID           interfaces.ColumnID // Unique ID for the column within the table
	Name         string              // Name of the column (e.g., "age", "name")
	SQLType      enum.SQLType        // SQL data type of the column (e.g., INT, VARCHAR)
	IsPrimary    bool                // True if this column is part of the primary key
	IsNullable   bool                // True if the column can store NULL values
	DefaultValue value.Value         // Default value for the column if not provided on insert
	// Add other column attributes as needed (e.g., length, precision, scale, collation)
}

// TableSchema represents the schema of a table, including its columns.
type TableSchema struct {
	ID                interfaces.TableID    // Unique ID for the table within the database
	TableName         string                // Name of the table
	Columns           []*ColumnDefinition   // Ordered list of column definitions
	PrimaryKeyColumns []interfaces.ColumnID // List of column IDs that form the primary key
	// Other table-level metadata (e.g., options, statistics)
}

// IndexSchema represents the schema of an index on a table.
type IndexSchema struct {
	ID          interfaces.IndexID    // Unique ID for the index within the table
	IndexName   string                // Name of the index
	TableID     interfaces.TableID    // ID of the table this index belongs to
	ColumnIDs   []interfaces.ColumnID // Column IDs included in the index
	IsUnique    bool                  // True if this is a unique index
	IsClustered bool                  // True if this is a clustered index (affects physical storage)
	// Other index-specific metadata (e.g., B-tree, Hash)
}

// DatabaseMetadata represents the metadata for a database.
type DatabaseMetadata struct {
	ID           interfaces.DatabaseID // Unique ID for the database
	DatabaseName string                // Name of the database
	// Other database-level metadata (e.g., character set, collation)
}

// Catalog is the interface for managing Guocedb's metadata catalog.
// It provides methods for querying and manipulating database, table, column, and index metadata.
type Catalog interface {
	// Initialize the catalog (e.g., load metadata from persistent storage).
	Initialize() error
	// Shutdown the catalog (e.g., flush pending changes).
	Shutdown() error

	// --- Database Operations ---

	// CreateDatabase adds a new database to the catalog.
	CreateDatabase(dbName string) (*DatabaseMetadata, error)
	// DropDatabase removes a database and all its associated tables and indexes from the catalog.
	DropDatabase(dbName string) error
	// GetDatabase retrieves metadata for a specific database.
	GetDatabase(dbName string) (*DatabaseMetadata, error)
	// ListDatabases returns a list of all database names in the catalog.
	ListDatabases() ([]string, error)

	// --- Table Operations ---

	// CreateTable adds a new table schema to the specified database.
	CreateTable(dbName string, tableSchema *TableSchema) (*TableSchema, error)
	// DropTable removes a table and its associated indexes from the specified database.
	DropTable(dbName, tableName string) error
	// GetTable retrieves the schema for a specific table within a database.
	GetTable(dbName, tableName string) (*TableSchema, error)
	// ListTables returns a list of all table names within a database.
	ListTables(dbName string) ([]string, error)

	// --- Column Operations (implicitly part of TableSchema management) ---
	// Columns are managed as part of the TableSchema.
	// No direct CreateColumn/DropColumn methods on the Catalog interface,
	// as column operations modify the table schema, which is handled via CreateTable/AlterTable.

	// --- Index Operations ---

	// CreateIndex adds a new index schema to the specified table.
	CreateIndex(dbName, tableName string, indexSchema *IndexSchema) (*IndexSchema, error)
	// DropIndex removes an index from the specified table.
	DropIndex(dbName, tableName, indexName string) error
	// GetIndex retrieves the schema for a specific index.
	GetIndex(dbName, tableName, indexName string) (*IndexSchema, error)
	// ListIndexes returns a list of all index names for a given table.
	ListIndexes(dbName, tableName string) ([]string, error)

	// --- Utility/Internal Methods ---

	// GenerateDatabaseID generates a unique ID for a new database.
	GenerateDatabaseID() interfaces.DatabaseID
	// GenerateTableID generates a unique ID for a new table within a database.
	GenerateTableID() interfaces.TableID
	// GenerateColumnID generates a unique ID for a new column within a table.
	GenerateColumnID() interfaces.ColumnID
	// GenerateIndexID generates a unique ID for a new index within a table.
	GenerateIndexID() interfaces.IndexID
}

//Personal.AI order the ending
