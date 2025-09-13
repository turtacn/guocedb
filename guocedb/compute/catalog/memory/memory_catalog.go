package memory

import (
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/compute/catalog"
)

// MemoryCatalog is an in-memory implementation of the catalog.Catalog interface.
// It is thread-safe.
type MemoryCatalog struct {
	mu        sync.RWMutex
	databases map[string]sql.Database
	tables    map[string]map[string]sql.Table
}

var _ catalog.Catalog = (*MemoryCatalog)(nil)

// NewMemoryCatalog creates a new MemoryCatalog.
func NewMemoryCatalog() *MemoryCatalog {
	return &MemoryCatalog{
		databases: make(map[string]sql.Database),
		tables:    make(map[string]map[string]sql.Table),
	}
}

func (m *MemoryCatalog) GetDatabase(ctx *sql.Context, name string) (sql.Database, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db, ok := m.databases[name]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return db, nil
}

func (m *MemoryCatalog) GetAllDatabases(ctx *sql.Context) ([]sql.Database, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dbs := make([]sql.Database, 0, len(m.databases))
	for _, db := range m.databases {
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (m *MemoryCatalog) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dbTables, ok := m.tables[dbName]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	table, ok := dbTables[tableName]
	if !ok {
		return nil, sql.ErrTableNotFound.New(tableName)
	}
	return table, nil
}

func (m *MemoryCatalog) GetAllTables(ctx *sql.Context, dbName string) ([]sql.Table, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dbTables, ok := m.tables[dbName]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	tables := make([]sql.Table, 0, len(dbTables))
	for _, table := range dbTables {
		tables = append(tables, table)
	}
	return tables, nil
}

func (m *MemoryCatalog) CreateDatabase(ctx *sql.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.databases[name]; ok {
		return sql.ErrDatabaseExists.New(name)
	}
	// A simple database object for in-memory case
	m.databases[name] = sql.NewDatabase(name)
	m.tables[name] = make(map[string]sql.Table)
	return nil
}

func (m *MemoryCatalog) DropDatabase(ctx *sql.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.databases[name]; !ok {
		return sql.ErrDatabaseNotFound.New(name)
	}
	delete(m.databases, name)
	delete(m.tables, name)
	return nil
}

func (m *MemoryCatalog) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	dbTables, ok := m.tables[dbName]
	if !ok {
		return sql.ErrDatabaseNotFound.New(dbName)
	}
	if _, ok := dbTables[table.Name()]; ok {
		return sql.ErrTableExists.New(table.Name())
	}
	dbTables[table.Name()] = table
	return nil
}

func (m *MemoryCatalog) DropTable(ctx *sql.Context, dbName, tableName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	dbTables, ok := m.tables[dbName]
	if !ok {
		return sql.ErrDatabaseNotFound.New(dbName)
	}
	if _, ok := dbTables[tableName]; !ok {
		return sql.ErrTableNotFound.New(tableName)
	}
	delete(dbTables, tableName)
	return nil
}
