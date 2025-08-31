// Package memory provides an in-memory implementation of the catalog.
package memory

import (
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/compute/catalog"
	"github.com/turtacn/guocedb/interfaces"
)

// Database is an in-memory representation of a database.
type Database struct {
	name   string
	tables map[string]sql.Table
}

// Name returns the database name.
func (d *Database) Name() string {
	return d.name
}

// GetTableInsensitive retrieves a table by name, case-insensitive.
func (d *Database) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	// In-memory is case-sensitive for simplicity.
	tbl, ok := d.tables[tblName]
	return tbl, ok, nil
}

// GetTableNames returns the names of all tables in the database.
func (d *Database) GetTableNames(ctx *sql.Context) ([]string, error) {
	var names []string
	for name := range d.tables {
		names = append(names, name)
	}
	return names, nil
}

// MemoryCatalog is an in-memory implementation of the catalog.Catalog interface.
type MemoryCatalog struct {
	mu        sync.RWMutex
	databases map[string]*Database
	storage   interfaces.Storage
}

// NewMemoryCatalog creates a new MemoryCatalog.
func NewMemoryCatalog(storage interfaces.Storage) *MemoryCatalog {
	return &MemoryCatalog{
		databases: make(map[string]*Database),
		storage:   storage,
	}
}

func (c *MemoryCatalog) GetStorage() interfaces.Storage {
	return c.storage
}

func (c *MemoryCatalog) GetDatabase(ctx *sql.Context, name string) (sql.Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	db, ok := c.databases[name]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return db, nil
}

func (c *MemoryCatalog) ListDatabases(ctx *sql.Context) ([]sql.Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var dbs []sql.Database
	for _, db := range c.databases {
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (c *MemoryCatalog) CreateDatabase(ctx *sql.Context, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.databases[name]; ok {
		return sql.ErrDatabaseExists.New(name)
	}
	c.databases[name] = &Database{
		name:   name,
		tables: make(map[string]sql.Table),
	}
	return c.storage.CreateDatabase(ctx, name)
}

func (c *MemoryCatalog) DropDatabase(ctx *sql.Context, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.databases[name]; !ok {
		return sql.ErrDatabaseNotFound.New(name)
	}
	delete(c.databases, name)
	return c.storage.DropDatabase(ctx, name)
}

func (c *MemoryCatalog) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	db, ok := c.databases[dbName]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	tbl, ok := db.tables[tableName]
	if !ok {
		return nil, sql.ErrTableNotFound.New(tableName)
	}
	return tbl, nil
}

func (c *MemoryCatalog) ListTables(ctx *sql.Context, dbName string) ([]sql.Table, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	db, ok := c.databases[dbName]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	var tables []sql.Table
	for _, tbl := range db.tables {
		tables = append(tables, tbl)
	}
	return tables, nil
}

func (c *MemoryCatalog) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	db, ok := c.databases[dbName]
	if !ok {
		return sql.ErrDatabaseNotFound.New(dbName)
	}
	if _, ok := db.tables[table.Name()]; ok {
		return sql.ErrTableExists.New(table.Name())
	}
	db.tables[table.Name()] = table
	return c.storage.CreateTable(ctx, dbName, table)
}

func (c *MemoryCatalog) DropTable(ctx *sql.Context, dbName, tableName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	db, ok := c.databases[dbName]
	if !ok {
		return sql.ErrDatabaseNotFound.New(dbName)
	}
	if _, ok := db.tables[tableName]; !ok {
		return sql.ErrTableNotFound.New(tableName)
	}
	delete(db.tables, tableName)
	return c.storage.DropTable(ctx, dbName, tableName)
}

func (c *MemoryCatalog) RegisterIndex(ctx *sql.Context, dbName, tableName, indexName string) error {
	// Placeholder
	return nil
}

func (c *MemoryCatalog) DropIndex(ctx *sql.Context, dbName, tableName, indexName string) error {
	// Placeholder
	return nil
}

// Enforce interface compliance
var _ catalog.Catalog = (*MemoryCatalog)(nil)
var _ sql.Database = (*Database)(nil)
