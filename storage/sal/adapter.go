// Package sal implements the Storage Abstraction Layer (SAL) adapter for Guocedb.
// This adapter is responsible for routing generic storage interface requests, as defined
// in interfaces/storage.go, to concrete storage engine implementations (e.g., Badger).
// It is a crucial internal component of the storage layer, depending on interfaces/storage.go
// and interacting with specific storage engines in storage/engines (e.g., badger/badger.go).
// Its primary role is to provide a unified entry point, allowing the compute layer to
// transparently use different storage engines.
package sal

import (
	"fmt"
	"sync"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces

	// Import concrete storage engines here for registration
	"github.com/turtacn/guocedb/storage/engines/badger" // Example: BadgerDB engine
	// "guocedb/storage/engines/inmemory" // Example: In-memory engine
)

// ensure that concrete engine implementations satisfy the StorageEngine interface.
var _ interfaces.StorageEngine = (*badger.BadgerEngine)(nil)

// var _ interfaces.StorageEngine = (*inmemory.InMemoryEngine)(nil) // Uncomment if inmemory engine is implemented

// StorageAdapter is the main entry point for the compute layer to interact with
// the underlying storage engine. It implements the interfaces.StorageEngine interface.
type StorageAdapter struct {
	activeEngine interfaces.StorageEngine
	mu           sync.RWMutex // Protects activeEngine
}

// NewStorageAdapter creates a new StorageAdapter instance.
// It initializes with a specific storage engine based on the provided type and config.
func NewStorageAdapter(engineType enum.StorageEngineType, config interface{}) (*StorageAdapter, error) {
	adapter := &StorageAdapter{}

	var engine interfaces.StorageEngine
	var err error

	switch engineType {
	case enum.StorageEngineBadger:
		badgerConfig, ok := config.(*badger.BadgerConfig)
		if !ok {
			return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
				"invalid configuration type for Badger engine", nil)
		}
		engine, err = badger.NewBadgerEngine(badgerConfig)
	// case enum.StorageEngineInMemory:
	// 	// In-memory engine might not need a config or have a simple one
	// 	engine, err = inmemory.NewInMemoryEngine()
	default:
		return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
			fmt.Sprintf("unsupported storage engine type: %s", engineType.String()), nil)
	}

	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
			fmt.Sprintf("failed to create storage engine of type %s", engineType.String()), err)
	}

	adapter.activeEngine = engine
	return adapter, nil
}

// Initialize delegates the call to the active storage engine.
func (sa *StorageAdapter) Initialize() error {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	if sa.activeEngine == nil {
		return errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"storage engine not initialized in adapter", nil)
	}
	return sa.activeEngine.Initialize()
}

// Shutdown delegates the call to the active storage engine.
func (sa *StorageAdapter) Shutdown() error {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	if sa.activeEngine == nil {
		return nil // Already shut down or never initialized
	}
	return sa.activeEngine.Shutdown()
}

// CreateDatabase delegates the call to the active storage engine.
func (sa *StorageAdapter) CreateDatabase(dbName string) error {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.activeEngine.CreateDatabase(dbName)
}

// DropDatabase delegates the call to the active storage engine.
func (sa *StorageAdapter) DropDatabase(dbName string) error {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.activeEngine.DropDatabase(dbName)
}

// GetDatabase delegates the call to the active storage engine.
func (sa *StorageAdapter) GetDatabase(dbName string) (interfaces.Database, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.activeEngine.GetDatabase(dbName)
}

// ListDatabases delegates the call to the active storage engine.
func (sa *StorageAdapter) ListDatabases() ([]string, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.activeEngine.ListDatabases()
}

// BeginTransaction delegates the call to the active storage engine.
func (sa *StorageAdapter) BeginTransaction(isolationLevel enum.IsolationLevel) (interfaces.Transaction, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.activeEngine.BeginTransaction(isolationLevel)
}

// GetEngineType returns the type of the active storage engine.
func (sa *StorageAdapter) GetEngineType() enum.StorageEngineType {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	if sa.activeEngine == nil {
		return enum.StorageEngineUnknown
	}
	return sa.activeEngine.GetEngineType()
}

// GetEngineStats delegates the call to the active storage engine.
func (sa *StorageAdapter) GetEngineStats() (interface{}, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	if sa.activeEngine == nil {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"storage engine not initialized in adapter", nil)
	}
	return sa.activeEngine.GetEngineStats()
}

// -----------------------------------------------------------------------------
// DatabaseAdapter: Adapts an interfaces.Database to the StorageAdapter's internal needs
// This might not be strictly necessary if GetDatabase directly returns the engine's Database.
// However, it's useful if the adapter needs to add cross-cutting concerns or
// maintain additional state related to database instances.
// For now, it largely just passes through.
// -----------------------------------------------------------------------------

// databaseAdapter implements the interfaces.Database interface by delegating
// to a concrete database instance from the underlying storage engine.
type databaseAdapter struct {
	db interfaces.Database
}

// newDatabaseAdapter creates a new databaseAdapter.
func newDatabaseAdapter(db interfaces.Database) *databaseAdapter {
	return &databaseAdapter{db: db}
}

// Name delegates the call.
func (da *databaseAdapter) Name() string {
	return da.db.Name()
}

// CreateTable delegates the call.
func (da *databaseAdapter) CreateTable(tableName string, schema *interfaces.TableSchema) error {
	return da.db.CreateTable(tableName, schema)
}

// DropTable delegates the call.
func (da *databaseAdapter) DropTable(tableName string) error {
	return da.db.DropTable(tableName)
}

// GetTable delegates the call.
func (da *databaseAdapter) GetTable(tableName string) (interfaces.Table, error) {
	table, err := da.db.GetTable(tableName)
	if err != nil {
		return nil, err
	}
	// Wrap the concrete table with a tableAdapter if necessary
	return newTableAdapter(table), nil
}

// ListTables delegates the call.
func (da *databaseAdapter) ListTables() ([]string, error) {
	return da.db.ListTables()
}

// -----------------------------------------------------------------------------
// TableAdapter: Adapts an interfaces.Table to the StorageAdapter's internal needs
// Similar to DatabaseAdapter, this wraps the concrete table implementation.
// -----------------------------------------------------------------------------

// tableAdapter implements the interfaces.Table interface.
type tableAdapter struct {
	table interfaces.Table
}

// newTableAdapter creates a new tableAdapter.
func newTableAdapter(table interfaces.Table) *tableAdapter {
	return &tableAdapter{table: table}
}

// Name delegates the call.
func (ta *tableAdapter) Name() string {
	return ta.table.Name()
}

// Schema delegates the call.
func (ta *tableAdapter) Schema() *interfaces.TableSchema {
	return ta.table.Schema()
}

// InsertRow delegates the call.
func (ta *tableAdapter) InsertRow(txn interfaces.Transaction, values []value.Value) (interfaces.RowID, error) {
	return ta.table.InsertRow(txn, values)
}

// ReadRow delegates the call.
func (ta *tableAdapter) ReadRow(txn interfaces.Transaction, rowID interfaces.RowID) ([]value.Value, error) {
	return ta.table.ReadRow(txn, rowID)
}

// UpdateRow delegates the call.
func (ta *tableAdapter) UpdateRow(txn interfaces.Transaction, rowID interfaces.RowID, updates map[interfaces.ColumnID]value.Value) error {
	return ta.table.UpdateRow(txn, rowID, updates)
}

// DeleteRow delegates the call.
func (ta *tableAdapter) DeleteRow(txn interfaces.Transaction, rowID interfaces.RowID) error {
	return ta.table.DeleteRow(txn, rowID)
}

// GetRowIterator delegates the call.
func (ta *tableAdapter) GetRowIterator(txn interfaces.Transaction, opts *interfaces.ScanOptions) (interfaces.RowIterator, error) {
	iter, err := ta.table.GetRowIterator(txn, opts)
	if err != nil {
		return nil, err
	}
	return newRowIteratorAdapter(iter), nil // Wrap the iterator
}

// GetApproxRowCount delegates the call.
func (ta *tableAdapter) GetApproxRowCount() (int64, error) {
	return ta.table.GetApproxRowCount()
}

// GetApproxTableSize delegates the call.
func (ta *tableAdapter) GetApproxTableSize() (int64, error) {
	return ta.table.GetApproxTableSize()
}

// -----------------------------------------------------------------------------
// RowIteratorAdapter: Adapts an interfaces.RowIterator
// -----------------------------------------------------------------------------

// rowIteratorAdapter implements the interfaces.RowIterator interface.
type rowIteratorAdapter struct {
	iter interfaces.RowIterator
}

// newRowIteratorAdapter creates a new rowIteratorAdapter.
func newRowIteratorAdapter(iter interfaces.RowIterator) *rowIteratorAdapter {
	return &rowIteratorAdapter{iter: iter}
}

// Next delegates the call.
func (ria *rowIteratorAdapter) Next() bool {
	return ria.iter.Next()
}

// Current delegates the call.
func (ria *rowIteratorAdapter) Current() (interfaces.RowID, []value.Value, error) {
	return ria.iter.Current()
}

// Close delegates the call.
func (ria *rowIteratorAdapter) Close() error {
	return ria.iter.Close()
}

// -----------------------------------------------------------------------------
// TransactionAdapter: Adapts an interfaces.Transaction
// -----------------------------------------------------------------------------

// transactionAdapter implements the interfaces.Transaction interface.
type transactionAdapter struct {
	txn interfaces.Transaction
}

// newTransactionAdapter creates a new transactionAdapter.
func newTransactionAdapter(txn interfaces.Transaction) *transactionAdapter {
	return &transactionAdapter{txn: txn}
}

// Commit delegates the call.
func (tra *transactionAdapter) Commit() error {
	return tra.txn.Commit()
}

// Rollback delegates the call.
func (tra *transactionAdapter) Rollback() error {
	return tra.txn.Rollback()
}

// IsReadOnly delegates the call.
func (tra *transactionAdapter) IsReadOnly() bool {
	return tra.txn.IsReadOnly()
}

// ID delegates the call.
func (tra *transactionAdapter) ID() interfaces.ID {
	return tra.txn.ID()
}

// SetTimeout delegates the call.
func (tra *transactionAdapter) SetTimeout(d time.Duration) {
	tra.txn.SetTimeout(d)
}

//Personal.AI order the ending
