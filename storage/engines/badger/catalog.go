package badger

import (
	"strings"
	"sync"

	"github.com/turtacn/guocedb/compute/sql"
)

// DatabaseProvider implements a provider for sql databases.
// This allows managing multiple databases through a single interface.
type DatabaseProvider interface {
	Database(ctx *sql.Context, name string) (sql.Database, error)
	HasDatabase(ctx *sql.Context, name string) bool
	AllDatabases(ctx *sql.Context) []sql.Database
}

// Catalog implements a catalog of databases and the DatabaseProvider interface.
type Catalog struct {
	mu   sync.RWMutex
	dbs  map[string]*Database
	path string
}

// NewCatalog creates a new Catalog.
func NewCatalog(path string) *Catalog {
	return &Catalog{
		dbs:  make(map[string]*Database),
		path: path,
	}
}

// Database returns a database by name (case-insensitive).
func (c *Catalog) Database(ctx *sql.Context, name string) (sql.Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lowerName := strings.ToLower(name)
	for dbName, db := range c.dbs {
		if strings.ToLower(dbName) == lowerName {
			return db, nil
		}
	}
	return nil, sql.ErrDatabaseNotFound.New(name)
}

// HasDatabase checks if a database exists (case-insensitive).
func (c *Catalog) HasDatabase(ctx *sql.Context, name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lowerName := strings.ToLower(name)
	for dbName := range c.dbs {
		if strings.ToLower(dbName) == lowerName {
			return true
		}
	}
	return false
}

// AllDatabases returns all databases in the catalog.
func (c *Catalog) AllDatabases(ctx *sql.Context) []sql.Database {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dbs := make([]sql.Database, 0, len(c.dbs))
	for _, db := range c.dbs {
		dbs = append(dbs, db)
	}
	return dbs
}

// AddDatabase adds a database to the catalog.
func (c *Catalog) AddDatabase(db *Database) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dbs[db.Name()] = db
}

// CreateDatabase creates a new database.
func (c *Catalog) CreateDatabase(name string, db *Database) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dbs[name] = db
}

// Tables returns the tables of a database.
func (c *Catalog) Tables(ctx *sql.Context, dbName string) (map[string]sql.Table, error) {
	db, err := c.Database(ctx, dbName)
	if err != nil {
		return nil, err
	}
	return db.Tables(), nil
}
