package badger

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// Catalog implements a catalog of databases.
type Catalog struct {
	dbs map[string]*Database
	path string
}

// NewCatalog creates a new Catalog.
func NewCatalog(path string) *Catalog {
	return &Catalog{
		dbs:  make(map[string]*Database),
		path: path,
	}
}

// Database returns a database by name.
func (c *Catalog) Database(name string) (sql.Database, error) {
	if db, ok := c.dbs[name]; ok {
		return db, nil
	}
	// In a real implementation, we might load it from disk if not in memory.
	// For now, return error if not found.
	return nil, sql.ErrDatabaseNotFound.New(name)
}

// CreateDatabase creates a new database.
// This is not strictly required by sql.Database interface but useful for testing/setup.
func (c *Catalog) CreateDatabase(name string, db *Database) {
	c.dbs[name] = db
}

// Tables returns the tables of a database.
func (c *Catalog) Tables(ctx *sql.Context, dbName string) (map[string]sql.Table, error) {
	db, err := c.Database(dbName)
	if err != nil {
		return nil, err
	}
	return db.Tables(), nil
}
