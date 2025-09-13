package badger

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/interfaces"
	"github.com/turtacn/guocedb/storage/sal"
)

// BadgerStorage is the main struct for the BadgerDB storage engine.
// It implements the interfaces.Storage and sql.DatabaseProvider interfaces.
type BadgerStorage struct {
	db   *badger.DB
	path string
	mu   sync.RWMutex
	// In-memory cache for database objects to avoid repeated lookups.
	databases map[string]*BadgerDatabase
}

var _ interfaces.Storage = (*BadgerStorage)(nil)
var _ sql.DatabaseProvider = (*BadgerStorage)(nil)

// init registers the BadgerDB storage engine with the SAL.
func init() {
	sal.RegisterStorageEngine(constants.StorageEngineBadger, NewBadgerStorage)
}

// NewBadgerStorage creates a new instance of the BadgerDB storage engine.
func NewBadgerStorage(cfg *config.StorageConfig) (interfaces.Storage, error) {
	opts := badger.DefaultOptions(cfg.Path)
	// TODO: Configure Badger options from the config (e.g., logger, cache sizes).
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	storage := &BadgerStorage{
		db:        db,
		path:      cfg.Path,
		databases: make(map[string]*BadgerDatabase),
	}
	// TODO: Load existing databases from metadata.
	return storage, nil
}

// NewTransaction starts a new transaction.
func (b *BadgerStorage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	txn := b.db.NewTransaction(!readOnly)
	return newBadgerTxn(txn), nil
}

// Database returns the database with the given name.
func (b *BadgerStorage) Database(ctx *sql.Context, name string) (sql.Database, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	db, ok := b.databases[name]
	if !ok {
		// TODO: Check metadata to see if this database actually exists.
		// For now, we assume it does if requested.
		return NewBadgerDatabase(b, name), nil
	}
	return db, nil
}

// HasDatabase checks if a database with the given name exists.
func (b *BadgerStorage) HasDatabase(ctx *sql.Context, name string) bool {
	_, err := b.Database(ctx, name)
	return err == nil
}

// AllDatabases returns all databases known to the provider.
func (b *BadgerStorage) AllDatabases(ctx *sql.Context) []sql.Database {
	b.mu.RLock()
	defer b.mu.RUnlock()
	dbs := make([]sql.Database, 0, len(b.databases))
	for _, db := range b.databases {
		dbs = append(dbs, db)
	}
	return dbs
}

// CreateDatabase creates a new database.
func (b *BadgerStorage) CreateDatabase(ctx *sql.Context, name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.databases[name]; exists {
		return sql.ErrDatabaseExists.New(name)
	}
	// TODO: Persist database metadata to Badger.
	b.databases[name] = NewBadgerDatabase(b, name)
	return nil
}

// DropDatabase drops a database.
func (b *BadgerStorage) DropDatabase(ctx *sql.Context, name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.databases[name]; !exists {
		return sql.ErrDatabaseNotFound.New(name)
	}
	// TODO: Remove database metadata and all associated tables/rows from Badger.
	delete(b.databases, name)
	return nil
}

// CreateTable is a placeholder and should be handled by the Database object.
func (b *BadgerStorage) CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error {
	// TODO: Persist table metadata.
	return nil
}

// DropTable is a placeholder.
func (b *BadgerStorage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	// TODO: Remove table metadata and rows.
	return nil
}

// GetTable is a placeholder.
func (b *BadgerStorage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	// TODO: Load table metadata and return a BadgerTable object.
	return nil, fmt.Errorf("GetTable not implemented yet")
}

// Close gracefully shuts down the BadgerDB instance.
func (b *BadgerStorage) Close() error {
	return b.db.Close()
}
