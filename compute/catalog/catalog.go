// Package catalog provides metadata management for guocedb.
package catalog

import (
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/interfaces"
)

// Catalog is the interface for the metadata catalog.
// It provides access to databases, tables, and other schema objects.
type Catalog interface {
	// GetDatabase returns a database from the catalog.
	GetDatabase(ctx *sql.Context, name string) (sql.Database, error)

	// ListDatabases returns all databases in the catalog.
	ListDatabases(ctx *sql.Context) ([]sql.Database, error)

	// CreateDatabase creates a new database.
	CreateDatabase(ctx *sql.Context, name string) error

	// DropDatabase drops a database.
	DropDatabase(ctx *sql.Context, name string) error

	// GetTable returns a table from a specific database.
	GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error)

	// ListTables returns all tables in a database.
	ListTables(ctx *sql.Context, dbName string) ([]sql.Table, error)

	// CreateTable creates a new table in a database.
	CreateTable(ctx *sql.Context, dbName string, table sql.Table) error

	// DropTable drops a table from a database.
	DropTable(ctx *sql.Context, dbName, tableName string) error

	// RegisterIndex registers a new index.
	RegisterIndex(ctx *sql.Context, dbName, tableName, indexName string) error

	// DropIndex drops an index.
	DropIndex(ctx *sql.Context, dbName, tableName, indexName string) error

	// GetStorage an instance of the storage engine
	GetStorage() interfaces.Storage
}
