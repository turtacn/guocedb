package persistent

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/compute/catalog"
	"github.com/turtacn/guocedb/interfaces"
)

// PersistentCatalog is a durable implementation of the catalog.Catalog interface.
// It stores all metadata in the underlying storage engine.
type PersistentCatalog struct {
	storage interfaces.Storage
}

var _ catalog.Catalog = (*PersistentCatalog)(nil)

// NewPersistentCatalog creates a new PersistentCatalog.
func NewPersistentCatalog(storage interfaces.Storage) *PersistentCatalog {
	return &PersistentCatalog{
		storage: storage,
	}
}

// The implementation of the interface methods will delegate to the
// underlying storage provider. The storage provider itself (e.g., BadgerStorage)
// is responsible for implementing the sql.DatabaseProvider logic, which
// handles the actual metadata operations. This makes the PersistentCatalog
// a thin wrapper or proxy.

func (p *PersistentCatalog) GetDatabase(ctx *sql.Context, name string) (sql.Database, error) {
	// Assumes the storage engine also implements sql.DatabaseProvider
	provider, ok := p.storage.(sql.DatabaseProvider)
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return provider.Database(ctx, name)
}

func (p *PersistentCatalog) GetAllDatabases(ctx *sql.Context) ([]sql.Database, error) {
	provider, ok := p.storage.(sql.DatabaseProvider)
	if !ok {
		return []sql.Database{}, nil
	}
	return provider.AllDatabases(ctx), nil
}

func (p *PersistentCatalog) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	db, err := p.GetDatabase(ctx, dbName)
	if err != nil {
		return nil, err
	}
	table, ok, err := db.GetTableInsensitive(ctx, tableName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, sql.ErrTableNotFound.New(tableName)
	}
	return table, nil
}

func (p *PersistentCatalog) GetAllTables(ctx *sql.Context, dbName string) ([]sql.Table, error) {
	db, err := p.GetDatabase(ctx, dbName)
	if err != nil {
		return nil, err
	}
	return db.GetTableNames(ctx)
}

func (p *PersistentCatalog) CreateDatabase(ctx *sql.Context, name string) error {
	provider, ok := p.storage.(sql.DatabaseProvider)
	if !ok {
		return sql.ErrUnsupportedFeature.New("create database")
	}
	// This feels redundant. The sql.DatabaseProvider should be the catalog.
	// Refactoring might be needed, but for now, following the plan.
	// This assumes the provider has a CreateDatabase method.
	// Let's assume it's part of the interfaces.Storage for now.
	return p.storage.CreateDatabase(ctx, name)
}

func (p *PersistentCatalog) DropDatabase(ctx *sql.Context, name string) error {
	return p.storage.DropDatabase(ctx, name)
}

func (p *PersistentCatalog) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	// This logic should be on the database object itself.
	db, err := p.GetDatabase(ctx, dbName)
	if err != nil {
		return err
	}
	creator, ok := db.(sql.TableCreator)
	if !ok {
		return sql.ErrUnsupportedFeature.New("create table")
	}
	// The schema is part of the table object
	return creator.CreateTable(ctx, table.Name(), table.Schema(), table.Collation())
}

func (p *PersistentCatalog) DropTable(ctx *sql.Context, dbName, tableName string) error {
	db, err := p.GetDatabase(ctx, dbName)
	if err != nil {
		return err
	}
	dropper, ok := db.(sql.TableDropper)
	if !ok {
		return sql.ErrUnsupportedFeature.New("drop table")
	}
	return dropper.DropTable(ctx, tableName)
}
