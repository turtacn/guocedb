package catalog

import "github.com/dolthub/go-mysql-server/sql"

// Catalog is the interface for the metadata catalog.
// It provides a unified way to access schema information for databases, tables,
// columns, indexes, and other database objects.
type Catalog interface {
	// GetDatabase retrieves a database by name.
	GetDatabase(ctx *sql.Context, name string) (sql.Database, error)

	// GetAllDatabases retrieves all databases in the catalog.
	GetAllDatabases(ctx *sql.Context) ([]sql.Database, error)

	// GetTable retrieves a table from a specific database by name.
	GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error)

	// GetAllTables retrieves all tables from a specific database.
	GetAllTables(ctx *sql.Context, dbName string) ([]sql.Table, error)

	// CreateDatabase creates a new database.
	CreateDatabase(ctx *sql.Context, name string) error

	// DropDatabase removes a database.
	DropDatabase(ctx *sql.Context, name string) error

	// CreateTable creates a new table in a specific database.
	CreateTable(ctx *sql.Context, dbName string, table sql.Table) error

	// DropTable removes a table from a specific database.
	DropTable(ctx *sql.Context, dbName, tableName string) error
}
