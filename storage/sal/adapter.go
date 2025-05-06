// Package sal implements the Storage Abstraction Layer.
// It acts as an adapter between the compute layer and pluggable storage engines.
// sal 包实现了存储抽象层。
// 它充当计算层和可插拔存储引擎之间的适配器。
package sal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/turtacn/guocedb/common/constants"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/interfaces" // Import the interfaces package
)

// SALAdapter is the implementation of the Storage Abstraction Layer.
// It manages the active storage engine.
// SALAdapter 是存储抽象层的实现。
// 它管理活跃的存储引擎。
type SALAdapter struct {
	activeEngineType enum.StorageEngineType
	activeEngine     interfaces.StorageEngine
	mu               sync.RWMutex // Protects access to activeEngine
}

// NewSALAdapter creates a new SALAdapter.
// NewSALAdapter 创建一个新的 SALAdapter。
func NewSALAdapter() *SALAdapter {
	return &SALAdapter{}
}

// Init initializes the SALAdapter with a specific storage engine.
// Init 使用特定的存储引擎初始化 SALAdapter。
func (s *SALAdapter) Init(ctx context.Context, engineType enum.StorageEngineType, config map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeEngine != nil {
		log.Warn("SALAdapter already initialized. Closing existing engine.") // SALAdapter 已经初始化。关闭现有引擎。
		if err := s.activeEngine.Close(ctx); err != nil {
			log.Error("Failed to close existing storage engine: %v", err) // 关闭现有存储引擎失败。
			// Continue with initialization of new engine despite close error? Decide on policy.
			// 尽管关闭错误，是否继续初始化新引擎？决定策略。
		}
		s.activeEngine = nil
		s.activeEngineType = ""
	}

	log.Info("Initializing storage engine: %s", engineType) // 初始化存储引擎。
	engine, err := s.getEngineInstance(engineType)
	if err != nil {
		return errors.ErrStorageEngineNotFound.New(engineType)
	}

	if err := engine.Init(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize storage engine %s: %w", engineType, err) // 初始化存储引擎失败。
	}

	s.activeEngine = engine
	s.activeEngineType = engineType
	log.Info("Storage engine '%s' initialized successfully.", engineType) // 存储引擎 '%s' 初始化成功。
	return nil
}

// getEngineInstance creates a new instance of a specific storage engine based on type.
// TODO: Implement actual engine instantiation logic.
// getEngineInstance 根据类型创建一个特定存储引擎的新实例。
// TODO: 实现实际的引擎实例化逻辑。
func (s *SALAdapter) getEngineInstance(engineType enum.StorageEngineType) (interfaces.StorageEngine, error) {
	switch engineType {
	case enum.StorageEngineBadger:
		// TODO: Return a new instance of the Badger engine implementation.
		// 返回 Badger 引擎实现的新的实例。
		log.Info("Creating placeholder Badger engine instance.") // 创建占位符 Badger 引擎实例。
		return &BadgerEnginePlaceholder{}, nil // Placeholder
	case enum.StorageEngineMemory:
		// This might represent the GMS in-memory database directly.
		// TODO: Adapt GMS in-memory database to StorageEngine interface or use a wrapper.
		// 这可能直接代表 GMS 的内存数据库。
		// TODO: 将 GMS 内存数据库适配到 StorageEngine 接口或使用包装器。
		log.Info("Creating placeholder Memory engine instance.") // 创建占位符 Memory 引擎实例。
		return &MemoryEnginePlaceholder{}, nil // Placeholder
	// Add other engine types here
	// 在此处添加其他引擎类型
	default:
		return nil, errors.ErrStorageEngineNotFound.New(engineType) // 未找到存储引擎。
	}
}

// Close closes the active storage engine.
// Close 关闭活跃的存储引擎。
func (s *SALAdapter) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeEngine == nil {
		log.Warn("SALAdapter is not initialized.") // SALAdapter 未初始化。
		return nil
	}

	log.Info("Closing storage engine: %s", s.activeEngineType) // 关闭存储引擎。
	err := s.activeEngine.Close(ctx)
	s.activeEngine = nil
	s.activeEngineType = ""
	if err != nil {
		log.Error("Failed to close storage engine %s: %v", s.activeEngineType, err) // 关闭存储引擎失败。
		return fmt.Errorf("failed to close storage engine: %w", err)
	}
	log.Info("Storage engine closed successfully.") // 存储引擎关闭成功。
	return nil
}

// GetActiveEngine returns the currently active storage engine.
// GetActiveEngine 返回当前活跃的存储引擎。
func (s *SALAdapter) GetActiveEngine() (interfaces.StorageEngine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeEngine == nil {
		return nil, errors.ErrInternal.New("SALAdapter is not initialized") // SALAdapter 未初始化。
	}
	return s.activeEngine, nil
}

// --- Pass-through methods to the active storage engine ---
// These methods simply delegate calls to the active storage engine.
// --- 传递到活跃存储引擎的方法 ---
// 这些方法仅将调用委托给活跃存储引擎。

// CreateDatabase delegates the call to the active engine.
// CreateDatabase 将调用委托给活跃引擎。
func (s *SALAdapter) CreateDatabase(ctx context.Context, dbName string) error {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return err
	}
	return engine.CreateDatabase(ctx, dbName)
}

// DropDatabase delegates the call to the active engine.
// DropDatabase 将调用委托给活跃引擎。
func (s *SALAdapter) DropDatabase(ctx context.Context, dbName string) error {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return err
	}
	return engine.DropDatabase(ctx, dbName)
}

// GetDatabase delegates the call to the active engine.
// GetDatabase 将调用委托给活跃引擎。
func (s *SALAdapter) GetDatabase(ctx context.Context, dbName string) (interfaces.Database, error) {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return nil, err
	}
	return engine.GetDatabase(ctx, dbName)
}

// ListDatabases delegates the call to the active engine.
// ListDatabases 将调用委托给活跃引擎。
func (s *SALAdapter) ListDatabases(ctx context.Context) ([]string, error) {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return nil, err
	}
	return engine.ListDatabases(ctx)
}

// GetTable delegates the call to the active engine.
// GetTable 将调用委托给活跃引擎。
func (s *SALAdapter) GetTable(ctx context.Context, dbName, tableName string) (interfaces.Table, error) {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return nil, err
	}
	return engine.GetTable(ctx, dbName, tableName)
}

// CreateTable delegates the call to the active engine.
// CreateTable 将调用委托给活跃引擎。
func (s *SALAdapter) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return err
	}
	return engine.CreateTable(ctx, dbName, tableName, schema)
}

// DropTable delegates the call to the active engine.
// DropTable 将调用委托给活跃引擎。
func (s *SALAdapter) DropTable(ctx context.Context, dbName, tableName string) error {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return err
	}
	return engine.DropTable(ctx, dbName, tableName)
}

// RenameTable delegates the call to the active engine.
// RenameTable 将调用委托给活跃引擎。
func (s *SALAdapter) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return err
	}
	return engine.RenameTable(ctx, dbName, oldTableName, newTableName)
}

// GetCatalog delegates the call to the active engine.
// GetCatalog 将调用委托给活跃引擎。
func (s *SALAdapter) GetCatalog(ctx context.Context) (sql.Catalog, error) {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return nil, err
	}
	return engine.GetCatalog(ctx)
}

// BeginTransaction delegates the call to the active engine.
// BeginTransaction 将调用委托给活跃引擎。
func (s *SALAdapter) BeginTransaction(ctx context.Context) (interfaces.Transaction, error) {
	engine, err := s.GetActiveEngine()
	if err != nil {
		return nil, err
	}
	return engine.BeginTransaction(ctx)
}

// --- Placeholder implementations for engine instances ---
// TODO: Replace these with actual engine implementations.
// --- 引擎实例的占位符实现 ---
// TODO: 用实际的引擎实现替换这些占位符。

type BadgerEnginePlaceholder struct {
	// Add necessary fields for Badger configuration/DB handle
	// 添加 Badger 配置/DB 句柄的必要字段
}

func (p *BadgerEnginePlaceholder) Init(ctx context.Context, config map[string]string) error {
	log.Info("BadgerEnginePlaceholder Init called with config: %v", config)
	// TODO: Initialize Badger DB here using config.
	return nil
}
func (p *BadgerEnginePlaceholder) Close(ctx context.Context) error {
	log.Info("BadgerEnginePlaceholder Close called.")
	// TODO: Close Badger DB here.
	return nil
}
func (p *BadgerEnginePlaceholder) CreateDatabase(ctx context.Context, dbName string) error {
	log.Info("BadgerEnginePlaceholder CreateDatabase called for: %s", dbName)
	// TODO: Implement database creation logic in Badger.
	// This might involve writing metadata to KV store.
	return nil
}
func (p *BadgerEnginePlaceholder) DropDatabase(ctx context.Context, dbName string) error {
	log.Info("BadgerEnginePlaceholder DropDatabase called for: %s", dbName)
	// TODO: Implement database dropping logic in Badger.
	// This involves deleting all related keys (catalog, data, index).
	return nil
}
func (p *BadgerEnginePlaceholder) GetDatabase(ctx context.Context, dbName string) (interfaces.Database, error) {
	log.Info("BadgerEnginePlaceholder GetDatabase called for: %s", dbName)
	// TODO: Implement database retrieval logic.
	// Check if database metadata exists in KV store.
	return &BadgerDatabasePlaceholder{name: dbName, engine: p}, nil // Return placeholder
}
func (p *BadgerEnginePlaceholder) ListDatabases(ctx context.Context) ([]string, error) {
	log.Info("BadgerEnginePlaceholder ListDatabases called.")
	// TODO: Implement logic to list databases from KV store metadata.
	return []string{constants.DefaultDatabaseName, constants.SystemDatabaseName}, nil // Dummy list
}
func (p *BadgerEnginePlaceholder) GetTable(ctx context.Context, dbName, tableName string) (interfaces.Table, error) {
	log.Info("BadgerEnginePlaceholder GetTable called for: %s.%s", dbName, tableName)
	// TODO: Implement table retrieval logic.
	// Check if table metadata exists in KV store.
	// Need schema to create the Table object.
	return nil, errors.ErrTableNotFound.New(tableName) // Placeholder
}
func (p *BadgerEnginePlaceholder) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("BadgerEnginePlaceholder CreateTable called for: %s.%s", dbName, tableName)
	// TODO: Implement table creation logic.
	// Store table schema metadata in KV store.
	return nil
}
func (p *BadgerEnginePlaceholder) DropTable(ctx context.Context, dbName, tableName string) error {
	log.Info("BadgerEnginePlaceholder DropTable called for: %s.%s", dbName, tableName)
	// TODO: Implement table dropping logic.
	// Delete all related keys (metadata, data, index) for this table.
	return nil
}
func (p *BadgerEnginePlaceholder) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	log.Info("BadgerEnginePlaceholder RenameTable called for: %s.%s -> %s", dbName, oldTableName, newTableName)
	// TODO: Implement table rename logic.
	// This is complex in KV: needs to move/copy all keys with the old table name prefix to the new one.
	return errors.ErrNotImplemented.New("table rename for Badger")
}

func (p *BadgerEnginePlaceholder) GetCatalog(ctx context.Context) (sql.Catalog, error) {
	log.Info("BadgerEnginePlaceholder GetCatalog called.")
	// TODO: Provide access to the catalog managed by this engine.
	// This might involve initializing a PersistentCatalog.
	return nil, errors.ErrNotImplemented.New("Badger engine catalog access") // Placeholder
}

func (p *BadgerEnginePlaceholder) BeginTransaction(ctx context.Context) (interfaces.Transaction, error) {
	log.Info("BadgerEnginePlaceholder BeginTransaction called.")
	// TODO: Start a Badger transaction and wrap it in a Transaction interface.
	return &BadgerTransactionPlaceholder{}, nil // Placeholder
}

// BadgerDatabasePlaceholder is a placeholder for Badger database implementation.
// BadgerDatabasePlaceholder 是 Badger 数据库实现的占位符。
type BadgerDatabasePlaceholder struct {
	name   string
	engine *BadgerEnginePlaceholder // Back reference to the engine
}

func (d *BadgerDatabasePlaceholder) Name() string { return d.name }
func (d *BadgerDatabasePlaceholder) GetTable(ctx context.Context, tableName string) (interfaces.Table, error) {
	log.Info("BadgerDatabasePlaceholder GetTable called for: %s.%s", d.name, tableName)
	return d.engine.GetTable(ctx, d.name, tableName) // Delegate back to engine
}
func (d *BadgerDatabasePlaceholder) ListTables(ctx context.Context) ([]string, error) {
	log.Info("BadgerDatabasePlaceholder ListTables called for: %s", d.name)
	// TODO: Implement logic to list tables within this database from KV metadata.
	return []string{}, nil // Dummy list
}
func (d *BadgerDatabasePlaceholder) CreateTable(ctx context.Context, tableName string, schema sql.Schema) error {
	log.Info("BadgerDatabasePlaceholder CreateTable called for: %s.%s", d.name, tableName)
	return d.engine.CreateTable(ctx, d.name, tableName, schema) // Delegate back to engine
}
func (d *BadgerDatabasePlaceholder) DropTable(ctx context.Context, tableName string) error {
	log.Info("BadgerDatabasePlaceholder DropTable called for: %s.%s", d.name, tableName)
	return d.engine.DropTable(ctx, d.name, tableName) // Delegate back to engine
}
func (d *BadgerDatabasePlaceholder) RenameTable(ctx context.Context, oldTableName, newTableName string) error {
	log.Info("BadgerDatabasePlaceholder RenameTable called for: %s.%s -> %s", d.name, oldTableName, newTableName)
	return d.engine.RenameTable(ctx, d.name, oldTableName, newTableName) // Delegate back to engine
}

// MemoryEnginePlaceholder is a placeholder for the Memory storage engine implementation.
// MemoryEnginePlaceholder 是内存存储引擎实现的占位符。
type MemoryEnginePlaceholder struct {
	// TODO: Embed or manage a GMS in-memory database here.
	// 在此处嵌入或管理 GMS 的内存数据库。
}

func (p *MemoryEnginePlaceholder) Init(ctx context.Context, config map[string]string) error {
	log.Info("MemoryEnginePlaceholder Init called with config: %v", config)
	// TODO: Initialize GMS in-memory database.
	return nil
}
func (p *MemoryEnginePlaceholder) Close(ctx context.Context) error {
	log.Info("MemoryEnginePlaceholder Close called.")
	// TODO: Close GMS in-memory database if needed.
	return nil
}
func (p *MemoryEnginePlaceholder) CreateDatabase(ctx context.Context, dbName string) error {
	log.Info("MemoryEnginePlaceholder CreateDatabase called for: %s", dbName)
	// TODO: Implement memory database creation.
	return nil
}
func (p *MemoryEnginePlaceholder) DropDatabase(ctx context.Context, dbName string) error {
	log.Info("MemoryEnginePlaceholder DropDatabase called for: %s", dbName)
	// TODO: Implement memory database dropping.
	return nil
}
func (p *MemoryEnginePlaceholder) GetDatabase(ctx context.Context, dbName string) (interfaces.Database, error) {
	log.Info("MemoryEnginePlaceholder GetDatabase called for: %s", dbName)
	// TODO: Implement memory database retrieval.
	return nil, errors.ErrDatabaseNotFound.New(dbName) // Placeholder
}
func (p *MemoryEnginePlaceholder) ListDatabases(ctx context.Context) ([]string, error) {
	log.Info("MemoryEnginePlaceholder ListDatabases called.")
	// TODO: Implement logic to list memory databases.
	return []string{}, nil // Dummy list
}
func (p *MemoryEnginePlaceholder) GetTable(ctx context.Context, dbName, tableName string) (interfaces.Table, error) {
	log.Info("MemoryEnginePlaceholder GetTable called for: %s.%s", dbName, tableName)
	// TODO: Implement memory table retrieval.
	// Needs to wrap GMS sql.Table in interfaces.Table.
	return nil, errors.ErrTableNotFound.New(tableName) // Placeholder
}
func (p *MemoryEnginePlaceholder) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("MemoryEnginePlaceholder CreateTable called for: %s.%s", dbName, tableName)
	// TODO: Implement memory table creation.
	return nil
}
func (p *MemoryEnginePlaceholder) DropTable(ctx context.Context, dbName, tableName string) error {
	log.Info("MemoryEnginePlaceholder DropTable called for: %s.%s", dbName, tableName)
	// TODO: Implement memory table dropping.
	return nil
}
func (p *MemoryEnginePlaceholder) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	log.Info("MemoryEnginePlaceholder RenameTable called for: %s.%s -> %s", dbName, oldTableName, newTableName)
	// TODO: Implement memory table rename.
	return nil
}
func (p *MemoryEnginePlaceholder) GetCatalog(ctx context.Context) (sql.Catalog, error) {
	log.Info("MemoryEnginePlaceholder GetCatalog called.")
	// TODO: Provide access to the GMS memory catalog.
	return nil, errors.ErrNotImplemented.New("Memory engine catalog access") // Placeholder
}
func (p *MemoryEnginePlaceholder) BeginTransaction(ctx context.Context) (interfaces.Transaction, error) {
	log.Info("MemoryEnginePlaceholder BeginTransaction called.")
	// TODO: Implement memory transaction (might not be true ACID depending on GMS memory).
	return &MemoryTransactionPlaceholder{}, nil // Placeholder
}

// BadgerTransactionPlaceholder is a placeholder for Badger transaction implementation.
// BadgerTransactionPlaceholder 是 Badger 事务实现的占位符。
type BadgerTransactionPlaceholder struct {
	// TODO: Embed or hold a reference to Badger's transaction object.
	// 在此处嵌入或持有对 Badger 事务对象的引用。
}

func (t *BadgerTransactionPlaceholder) Commit(ctx context.Context) error {
	log.Info("BadgerTransactionPlaceholder Commit called.")
	// TODO: Call Badger transaction commit.
	return nil
}
func (t *BadgerTransactionPlaceholder) Rollback(ctx context.Context) error {
	log.Info("BadgerTransactionPlaceholder Rollback called.")
	// TODO: Call Badger transaction rollback.
	return nil
}
func (t *BadgerTransactionPlaceholder) UnderlyingTx() interface{} {
	log.Info("BadgerTransactionPlaceholder UnderlyingTx called.")
	// TODO: Return Badger transaction object.
	return nil // Placeholder
}

// MemoryTransactionPlaceholder is a placeholder for Memory transaction implementation.
// MemoryTransactionPlaceholder 是内存事务实现的占位符。
type MemoryTransactionPlaceholder struct {
	// TODO: Implement memory transaction logic.
	// 实现内存事务逻辑。
}

func (t *MemoryTransactionPlaceholder) Commit(ctx context.Context) error {
	log.Info("MemoryTransactionPlaceholder Commit called.")
	// TODO: Commit memory transaction logic.
	return nil
}
func (t *MemoryTransactionPlaceholder) Rollback(ctx context.Context) error {
	log.Info("MemoryTransactionPlaceholder Rollback called.")
	// TODO: Rollback memory transaction logic.
	return nil
}
func (t *MemoryTransactionPlaceholder) UnderlyingTx() interface{} {
	log.Info("MemoryTransactionPlaceholder UnderlyingTx called.")
	// TODO: Return memory transaction object if any.
	return nil
}

// --- Placeholder for Badger specific Table and RowIterator ---
// TODO: Move these to storage/engines/badger package later.
// --- Badger 特定 Table 和 RowIterator 的占位符 ---
// TODO: 稍后将这些移动到 storage/engines/badger 包中。

// BadgerTablePlaceholder is a placeholder for Badger table implementation.
// BadgerTablePlaceholder 是 Badger 表实现的占位符。
type BadgerTablePlaceholder struct {
	// Add necessary fields like schema, database name, table name, engine reference.
	// 添加必要字段，如模式、数据库名、表名、引擎引用。
	schemaName string
	tableName  string
	tableSchema sql.Schema
	engine *BadgerEnginePlaceholder // Back reference to the engine
}

func (t *BadgerTablePlaceholder) Name() string { return t.tableName }
func (t *BadgerTablePlaceholder) Schema() sql.Schema { return t.tableSchema }
func (t *BadgerTablePlaceholder) Collation() sql.CollationID { return sql.Collation_Default } // Placeholder
func (t *BadgerTablePlaceholder) Comment() string { return "" } // Placeholder
func (t *BadgerTablePlaceholder) Partitions(ctx context.Context) (sql.PartitionIter, error) {
	log.Info("BadgerTablePlaceholder Partitions called for %s.%s", t.schemaName, t.tableName)
	// For Badger, treat the whole table as a single partition initially.
	// The Partition object might just contain the table name or a identifier.
	//
	// 对于 Badger，初步将整个表视为单个 Partition。
	// Partition 对象可能只包含表名或标识符。
	// TODO: Implement proper partitioning if needed for larger datasets.
	// TODO: 如果需要处理更大的数据集，实现适当的分区。
	return &BadgerPartitionIterPlaceholder{tableName: t.tableName, dbName: t.schemaName}, nil // Return placeholder
}
func (t *BadgerTablePlaceholder) PartitionRows(ctx context.Context, partition sql.Partition) (sql.RowIter, error) {
	log.Info("BadgerTablePlaceholder PartitionRows called for %s.%s", t.schemaName, t.tableName)
	// The partition object needs to contain info to start scan in Badger.
	// For a single partition, it means scanning the whole table's data key range.
	//
	// partition 对象需要包含信息，以便在 Badger 中开始扫描。
	// 对于单个 Partition，这意味着扫描整个表的数据 key 范围。
	// TODO: Implement Badger specific RowIter.
	// TODO: 实现 Badger 特定的 RowIter。
	return &BadgerRowIterPlaceholder{}, nil // Return placeholder
}

func (t *BadgerTablePlaceholder) Insert(ctx context.Context, row sql.Row) error {
	log.Info("BadgerTablePlaceholder Insert called for %s.%s with row: %v", t.schemaName, t.tableName, row)
	// TODO: Implement data insertion into Badger.
	// Encode row to value, encode primary key to key. Write KV pair.
	return errors.ErrNotImplemented.New("Badger table insert")
}
func (t *BadgerTablePlaceholder) Update(ctx context.Context, oldRow, newRow sql.Row) error {
	log.Info("BadgerTablePlaceholder Update called for %s.%s old: %v, new: %v", t.schemaName, t.tableName, oldRow, newRow)
	// TODO: Implement data update in Badger.
	// Delete old key, insert new key/value. Requires transaction.
	return errors.ErrNotImplemented.New("Badger table update")
}
func (t *BadgerTablePlaceholder) Delete(ctx context.Context, row sql.Row) error {
	log.Info("BadgerTablePlaceholder Delete called for %s.%s with row: %v", t.schemaName, t.tableName, row)
	// TODO: Implement data deletion from Badger.
	// Encode primary key, delete KV pair. Requires transaction.
	return errors.ErrNotImplemented.New("Badger table delete")
}
func (t *BadgerTablePlaceholder) Truncate(ctx context.Context) error {
	log.Info("BadgerTablePlaceholder Truncate called for %s.%s", t.schemaName, t.tableName)
	// TODO: Implement table truncation in Badger.
	// Delete all keys with this table's data prefix. Requires careful range deletion.
	return errors.ErrNotImplemented.New("Badger table truncate")
}
func (t *BadgerTablePlaceholder) CreateIndex(ctx context.Context, indexDef sql.IndexDef) error {
	log.Info("BadgerTablePlaceholder CreateIndex called for %s.%s index: %s", t.schemaName, t.tableName, indexDef.Name)
	// TODO: Implement index creation.
	// Store index metadata. Build index data by scanning the table and writing index keys.
	return errors.ErrNotImplemented.New("Badger table create index")
}
func (t *BadgerTablePlaceholder) DropIndex(ctx context.Context, indexName string) error {
	log.Info("BadgerTablePlaceholder DropIndex called for %s.%s index: %s", t.schemaName, t.tableName, indexName)
	// TODO: Implement index dropping.
	// Delete index metadata. Delete all index keys for this index.
	return errors.ErrNotImplemented.New("Badger table drop index")
}
func (t *BadgerTablePlaceholder) GetIndex(ctx context.Context, indexName string) (sql.Index, error) {
	log.Info("BadgerTablePlaceholder GetIndex called for %s.%s index: %s", t.schemaName, t.tableName, indexName)
	// TODO: Retrieve index metadata from KV store and return a GMS sql.Index object.
	return nil, errors.ErrIndexNotFound.New(indexName) // Placeholder
}
func (t *BadgerTablePlaceholder) GetIndexes(ctx context.Context) ([]sql.Index, error) {
	log.Info("BadgerTablePlaceholder GetIndexes called for %s.%s", t.schemaName, t.tableName)
	// TODO: Retrieve all index metadata for this table from KV store.
	return []sql.Index{}, nil // Placeholder
}

// BadgerPartitionPlaceholder is a placeholder for Badger partition.
// BadgerPartitionPlaceholder 是 Badger Partition 的占位符。
type BadgerPartitionPlaceholder struct {
	tableName string
	dbName string
	// Add necessary partition info if any. For a single partition, just table name might suffice.
	// 如果需要，添加必要的 Partition 信息。对于单个 Partition，只需要表名即可。
}

func (p *BadgerPartitionPlaceholder) Key() []byte {
	// Return a key that GMS can use to identify this partition.
	// For a table, this could be the encoded table name prefix for data keys.
	//
	// 返回一个 GMS 可以用来识别此 Partition 的 key。
	// 对于表来说，这可以是数据 key 的编码表名前缀。
	prefix := bytes.Join([][]byte{
		NamespaceDataBytes,
		[]byte(p.dbName),
		[]byte(p.tableName),
	}, []byte{NsSep, Sep, Sep})
	// Add a separator at the end to make it a prefix
	return append(prefix, Sep)
}

// BadgerPartitionIterPlaceholder is a placeholder for Badger partition iterator.
// BadgerPartitionIterPlaceholder 是 Badger Partition 迭代器的占位符。
type BadgerPartitionIterPlaceholder struct {
	tableName string
	dbName string
	sent bool // To simulate returning one partition
}

func (i *BadgerPartitionIterPlaceholder) Next(ctx context.Context) (sql.Partition, error) {
	if i.sent {
		return nil, nil // No more partitions
	}
	i.sent = true
	// Return the single partition representing the whole table.
	// 返回代表整个表的单个 Partition。
	return &BadgerPartitionPlaceholder{tableName: i.tableName, dbName: i.dbName}, nil
}
func (i *BadgerPartitionIterPlaceholder) Close(ctx context.Context) error {
	log.Info("BadgerPartitionIterPlaceholder Close called.")
	return nil // Nothing to close for this placeholder
}

// BadgerRowIterPlaceholder is a placeholder for Badger row iterator.
// BadgerRowIterPlaceholder 是 Badger 行迭代器的占位符。
type BadgerRowIterPlaceholder struct {
	// TODO: Hold a reference to a Badger iterator and scan state.
	// 持有对 Badger 迭代器和扫描状态的引用。
}

func (i *BadgerRowIterPlaceholder) Next(ctx context.Context) (sql.Row, error) {
	// TODO: Use the Badger iterator to read the next KV pair.
	// Decode the value into a sql.Row using the table schema.
	//
	// 使用 Badger 迭代器读取下一个 KV 对。
	// 使用表模式将 value 解码为 sql.Row。
	log.Info("BadgerRowIterPlaceholder Next called.")
	return nil, nil // No rows for this placeholder
}

func (i *BadgerRowIterPlaceholder) Close(ctx context.Context) error {
	log.Info("BadgerRowIterPlaceholder Close called.")
	// TODO: Close the underlying Badger iterator.
	return nil
}