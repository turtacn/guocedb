// Package sal provides the Storage Abstraction Layer for guocedb.
package sal

import (
	"fmt"
	"sync"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
	"github.com/turtacn/guocedb/storage/engines/badger"
	// Placeholders for future engines
	// "github.com/turtacn/guocedb/storage/engines/kvd"
	// "github.com/turtacn/guocedb/storage/engines/mdd"
	// "github.com/turtacn/guocedb/storage/engines/mdi"
)

var (
	// mu protects the drivers map
	mu sync.RWMutex
	// drivers stores the registered storage engine drivers
	drivers = make(map[string]interfaces.Storage)
)

// Register makes a storage driver available by name.
// If Register is called twice with the same name or if the driver is nil,
// it panics.
func Register(name string, driver interfaces.Storage) {
	mu.Lock()
	defer mu.Unlock()
	if driver == nil {
		panic("storage: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("storage: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// GetStorageEngine initializes and returns the storage engine specified in the configuration.
func GetStorageEngine(cfg *config.Config) (interfaces.Storage, error) {
	mu.RLock()
	defer mu.RUnlock()

	engineName := cfg.Storage.Engine
	driver, ok := drivers[engineName]
	if !ok {
		return nil, fmt.Errorf("storage: unknown driver %q (forgotten import?)", engineName)
	}
	return driver, nil
}

// init registers the default drivers.
func init() {
	// Register the BadgerDB engine
	// In a real application, the badger instance would be initialized based on config.
	// For now, we register a placeholder that would be configured later.
	// A better approach would be a factory function.
	// Let's assume a factory `badger.NewStorage` exists.

	// This is a bit of a chicken-and-egg problem. We need the config to init the engine,
	// but we register the *type* of engine here. Let's adjust the pattern.
	// We'll register factory functions instead.

	// Let's simplify for now: we will initialize the engines in the server's main function
	// and the adapter will simply hold the active engine.
}

// Adapter is a concrete implementation of the Storage interface that delegates
// calls to a specific, underlying storage engine.
type Adapter struct {
	engine interfaces.Storage
}

// NewAdapter creates a new storage adapter for the configured engine.
func NewAdapter(cfg *config.Config) (*Adapter, error) {
	var engine interfaces.Storage
	var err error

	switch cfg.Storage.Engine {
	case "badger":
		engine, err = badger.NewStorage(cfg.Storage.Badger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize badger storage: %w", err)
		}
	// case "kvd":
	// 	engine, err = kvd.NewStorage(cfg.Storage.KVD)
	// 	...
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", cfg.Storage.Engine)
	}

	return &Adapter{engine: engine}, nil
}

// Forward all the interface methods to the underlying engine.
// This is boilerplate but ensures the Adapter satisfies the interface.

func (a *Adapter) Get(ctx *sql.Context, db, table string, key []byte) ([]byte, error) {
	return a.engine.Get(ctx, db, table, key)
}

func (a *Adapter) Set(ctx *sql.Context, db, table string, key, value []byte) error {
	return a.engine.Set(ctx, db, table, key, value)
}

func (a *Adapter) Delete(ctx *sql.Context, db, table string, key []byte) error {
	return a.engine.Delete(ctx, db, table, key)
}

func (a *Adapter) Iterator(ctx *sql.Context, db, table string, prefix []byte) (interfaces.Iterator, error) {
	return a.engine.Iterator(ctx, db, table, prefix)
}

func (a *Adapter) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return a.engine.NewTransaction(ctx, readOnly)
}

func (a *Adapter) CreateDatabase(ctx *sql.Context, name string) error {
	return a.engine.CreateDatabase(ctx, name)
}

func (a *Adapter) DropDatabase(ctx *sql.Context, name string) error {
	return a.engine.DropDatabase(ctx, name)
}

func (a *Adapter) ListDatabases(ctx *sql.Context) ([]string, error) {
	return a.engine.ListDatabases(ctx)
}

func (a *Adapter) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	return a.engine.CreateTable(ctx, dbName, table)
}

func (a *Adapter) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return a.engine.DropTable(ctx, dbName, tableName)
}

func (a *Adapter) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	return a.engine.GetTable(ctx, dbName, tableName)
}

func (a *Adapter) ListTables(ctx *sql.Context, dbName string) ([]string, error) {
	return a.engine.ListTables(ctx, dbName)
}

func (a *Adapter) Close() error {
	return a.engine.Close()
}
