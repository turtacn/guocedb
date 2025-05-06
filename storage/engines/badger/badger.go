// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"bytes"
	"strings" // Needed for case-insensitive table lookup

	badger "github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/mysql_db" // Required for GMS system database
	"github.com/dolthub/go-mysql-server/sql/information_schema" // Required for GMS information_schema
	// "github.com/turtacn/guocedb/common/config" // Not used directly here, but in main.go to get config values
	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/interfaces" // Import the interfaces package
)

// BadgerEngine is an implementation of the interfaces.StorageEngine interface
// that uses Badger as the backend storage.
// BadgerEngine 是 interfaces.StorageEngine 接口的实现，
// 它使用 Badger 作为后端存储。
type BadgerEngine struct {
	// db is the Badger database handle.
	// db 代表 Badger 数据库句柄。
	db *badger.DB
	// config holds the configuration for the engine.
	// config 保存引擎的配置。
	config map[string]string
	// mu protects concurrent access to engine fields if needed (less likely for DB handle).
	// mu 保护引擎字段（如果需要）的并发访问（对于 DB 句柄不太可能）。
	mu sync.RWMutex // Maybe needed for catalog or other mutable state
	// catalog is the sql.Catalog provided by this engine.
	// It bridges the GMS catalog interface to the Badger storage logic.
	// catalog 是此引擎提供的 sql.Catalog。
	// 它将 GMS catalog 接口桥接到 Badger 存储逻辑。
	catalog sql.Catalog // This will be a custom implementation
}

// NewBadgerEngine creates a new uninitialized BadgerEngine instance.
// NewBadgerEngine 创建一个新的未初始化的 BadgerEngine 实例。
func NewBadgerEngine() *BadgerEngine {
	return &BadgerEngine{}
}


// Init initializes the Badger storage engine.
// Init 初始化 Badger 存储引擎。
func (e *BadgerEngine) Init(ctx context.Context, config map[string]string) error {
	log.Info("Initializing Badger storage engine...") // 初始化 Badger 存储引擎。

	e.config = config // Store config

	// Load config values for data and WAL directories
	// 加载数据和 WAL 目录的配置值
	dataDir := config[string(enum.ConfigKeyBadgerDataDir)]
	if dataDir == "" {
		dataDir = constants.DefaultBadgerDataDir
		log.Info("Badger data directory not specified, using default: %s", dataDir) // 未指定 Badger 数据目录，使用默认路径。
	}
	walDir := config[string(enum.ConfigKeyBadgerWALDir)]
	if walDir == "" {
		walDir = constants.DefaultBadgerWALDir
		log.Info("Badger WAL directory not specified, using default: %s", walDir) // 未指定 Badger WAL 目录，使用默认路径。
	}

	// Ensure directories exist
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Error("Failed to create Badger data directory %s: %v", dataDir, err) // 创建 Badger 数据目录失败。
		return fmt.Errorf("failed to create Badger data directory %s: %w", dataDir, err)
	}
	if err := os.MkdirAll(walDir, 0755); err != nil {
		log.Error("Failed to create Badger WAL directory %s: %v", walDir, err) // 创建 Badger WAL 目录失败。
		return fmt.Errorf("failed to create Badger WAL directory %s: %w", walDir, err)
	}

	// Open Badger DB
	// 打开 Badger DB
	opts := badger.DefaultOptions(dataDir).WithLogger(NewBadgerLoggerAdapter()).WithSyncWrites(true) // SyncWrites for safety / 为安全起见启用同步写入
	if dataDir != walDir {
		opts = opts.WithValueLogPath(walDir)
	}

	db, err := badger.Open(opts)
	if err != nil {
		log.Error("Failed to open Badger DB at %s (WAL: %s): %v", dataDir, walDir, err) // 打开 Badger DB 失败。
		return fmt.Errorf("%w: failed to open Badger DB: %v", errors.ErrBadgerOperationFailed, err)
	}
	e.db = db
	log.Info("Badger DB opened successfully.") // Badger DB 打开成功。

	// Initialize the catalog
	// The catalog needs access to the Badger DB to read metadata.
	// We'll create a BadgerCatalog that implements sql.Catalog and holds a ref to this engine.
	//
	// 初始化 catalog。
	// catalog 需要访问 Badger DB 读取元数据。
	// 我们将创建一个实现 sql.Catalog 并持有此引擎引用的 BadgerCatalog。
	e.catalog = NewBadgerCatalog(e) // Create the catalog instance, pass engine reference

	// Ensure default database exists on first run
	// 在首次运行时确保默认数据库存在
	// This requires using the catalog to check for the database.
	// This check needs a context.
	// 这需要使用 catalog 来检查数据库。
	// 此检查需要 context。
	checkCtx := context.Background() // Use background context for initialization check
	// Use the catalog to check for the database existence
	// 使用 catalog 检查数据库存在性
	_, err = e.catalog.Database(sql.NewEmptyContext(), constants.DefaultDatabaseName) // Use EmptyContext if no request context available
	if sql.ErrDatabaseNotFound.Is(err) {
		log.Info("Default database '%s' not found, creating it.", constants.DefaultDatabaseName) // 默认数据库 '%s' 未找到，正在创建。
		// Create the default database using the engine's method
		// 使用引擎的方法创建默认数据库
		if createErr := e.CreateDatabase(checkCtx, constants.DefaultDatabaseName); createErr != nil {
			log.Error("Failed to create default database '%s' on initialization: %v", constants.DefaultDatabaseName, createErr) // 初始化时创建默认数据库失败。
			// Decide if this is a fatal error
			// 决定这是否是致命错误
			return fmt.Errorf("failed to create default database '%s': %w", constants.DefaultDatabaseName, createErr)
		}
		log.Info("Default database '%s' created successfully.", constants.DefaultDatabaseName) // 默认数据库 '%s' 创建成功。
	} else if err != nil {
		// Handle other errors from checking database existence
		// 处理检查数据库存在性的其他错误
		log.Error("Error checking for default database '%s': %v", constants.DefaultDatabaseName, err) // 检查默认数据库出错。
		return fmt.Errorf("error checking for default database '%s': %w", constants.DefaultDatabaseName, err)
	} else {
		log.Info("Default database '%s' already exists.", constants.DefaultDatabaseName) // 默认数据库 '%s' 已存在。
	}


	log.Info("Badger storage engine initialized.") // Badger 存储引擎初始化完成。
	return nil
}

// Close closes the Badger storage engine.
// Close 关闭 Badger 存储引擎。
func (e *BadgerEngine) Close(ctx context.Context) error {
	log.Info("Closing Badger storage engine...") // 关闭 Badger 存储引擎。
	if e.db != nil {
		err := e.db.Close()
		if err != nil {
			log.Error("Failed to close Badger DB: %v", err) // 关闭 Badger DB 失败。
			return fmt.Errorf("%w: failed to close Badger DB: %v", errors.ErrBadgerOperationFailed, err)
		}
		e.db = nil
		log.Info("Badger DB closed successfully.") // Badger DB 关闭成功。
	}
	return nil
}

// --- Database Management Methods ---

// CreateDatabase creates a new database in Badger.
// It writes a metadata entry for the database in the catalog.
// CreateDatabase 在 Badger 中创建一个新数据库。
// 它在 catalog 中为数据库写入一个元数据条目。
func (e *BadgerEngine) CreateDatabase(ctx context.Context, dbName string) error {
	log.Info("BadgerEngine CreateDatabase called for: %s", dbName) // 调用 BadgerEngine CreateDatabase。

	// Check if database already exists
	// 检查数据库是否已存在
	// Use the catalog to perform the check
	// 使用 catalog 进行检查
	// Use sql.NewEmptyContext() or adapt the incoming context.
	// 使用 sql.NewEmptyContext() 或适配传入的 context。
	gmsCtx := sql.NewEmptyContext() // Default GMS context
	_, err := e.catalog.Database(gmsCtx, dbName)
	if err == nil {
		log.Warn("Database '%s' already exists.", dbName) // 数据库 '%s' 已存在。
		return errors.ErrDatabaseAlreadyExists.New(dbName)
	}
	// If err is not sql.ErrDatabaseNotFound, return it
	// 如果 err 不是 sql.ErrDatabaseNotFound，则返回
	if !sql.ErrDatabaseNotFound.Is(err) {
		log.Error("Failed to check database existence for '%s' using catalog: %v", dbName, err) // 使用 catalog 检查数据库存在性失败。
		return fmt.Errorf("failed to check database existence using catalog: %w", err)
	}

	// Database does not exist, create it.
	// Write a dummy key to mark database existence in the catalog namespace.
	// 数据库不存在，创建它。
	// 在 catalog 命名空间中写入一个虚拟 key 来标记数据库存在。
	dbMetadataKey := bytes.Join([][]byte{
		NamespaceCatalogBytes,
		[]byte(dbName),
	}, []byte{NsSep}) // Key like catalog:<db_name>

	writeTxn := e.db.NewTransaction(true) // Read-write
	defer writeTxn.Discard()

	// Write a marker key for the database
	// 写入一个标记 key 来标记数据库存在
	if err := writeTxn.Set(dbMetadataKey, []byte("exists")); err != nil { // Value can be anything or empty
		log.Error("Failed to set database metadata key for '%s': %v", dbName, err) // 设置数据库元数据 key 失败。
		return fmt.Errorf("%w: failed to write database metadata: %v", errors.ErrBadgerOperationFailed, err)
	}

	// Commit the transaction
	// 提交事务
	if err := writeTxn.Commit(); err != nil {
		log.Error("Failed to commit transaction for creating database '%s': %v", dbName, err) // 提交创建数据库事务失败。
		return fmt.Errorf("%w: failed to commit database creation: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Info("Database '%s' created successfully.", dbName) // 数据库 '%s' 创建成功。
	return nil
}

// DropDatabase drops an existing database from Badger.
// It deletes the database metadata and all associated data/index/catalog keys.
// DropDatabase 从 Badger 中删除现有数据库。
// 它删除数据库元数据以及所有相关的数据/索引/catalog key。
func (e *BadgerEngine) DropDatabase(ctx context.Context, dbName string) error {
	log.Info("BadgerEngine DropDatabase called for: %s", dbName) // 调用 BadgerEngine DropDatabase。

	// Cannot drop system databases
	// 无法删除系统数据库
	if dbName == constants.SystemDatabaseName || dbName == "information_schema" { // Include GMS system DB
		log.Warn("Attempted to drop system database '%s'. Permission denied.", dbName) // 尝试删除系统数据库。权限被拒绝。
		return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot drop system database '%s'", dbName)) // 无法删除系统数据库 '%s'。
	}

	// Check if database exists using the catalog
	// 使用 catalog 检查数据库是否存在
	// Use sql.NewEmptyContext() or adapt the incoming context.
	gmsCtx := sql.NewEmptyContext() // Default GMS context
	_, err := e.catalog.Database(gmsCtx, dbName)
	if sql.ErrDatabaseNotFound.Is(err) {
		log.Warn("Database '%s' not found for dropping.", dbName) // 删除数据库时未找到。
		return errors.ErrDatabaseNotFound.New(dbName)
	} else if err != nil {
		log.Error("Failed to check database existence for drop '%s' using catalog: %v", dbName, err) // 使用 catalog 检查数据库存在性失败。
		return fmt.Errorf("failed to check database existence for drop using catalog: %w", err)
	}

	// Database exists, proceed with dropping.
	// This is a complex operation: delete all keys starting with *:<dbName>:
	// 数据库存在，继续删除。
	// 这是一个复杂的操作：删除所有以 *:<dbName>: 开头的 key。
	log.Warn("Dropping user database '%s' requires deleting all associated data (catalog, data, index). This is a potentially slow operation.", dbName) // 删除用户数据库需要删除所有相关数据。这可能是一个慢操作。

	// Use a read-write transaction for all deletions
	// 使用读写事务进行所有删除操作
	writeTxn := e.db.NewTransaction(true) // Read-write
	defer writeTxn.Discard() // Discard on failure

	// Delete all keys under the database prefix across all namespaces.
	// This requires scanning keys that start with <Namespace>:<dbName>:.
	// 删除此数据库前缀下所有命名空间中的所有 key。
	// 这需要扫描以 <Namespace>:<dbName>: 开头的 key。

	namespacesToDelete := [][]byte{
		NamespaceCatalogBytes,
		NamespaceDataBytes,
		NamespaceIndexBytes,
	}

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false // Don't need values

	for _, ns := range namespacesToDelete {
		// Prefix for keys in this namespace and database: <ns>:<dbName>:
		// 此命名空间和数据库中 key 的前缀：<ns>:<dbName>:
		prefix := bytes.Join([][]byte{ns, []byte(dbName)}, []byte{NsSep})
		prefix = append(prefix, Sep) // Add separator to match keys within the database

		it := writeTxn.NewIterator(opts)
		defer it.Close()

		log.Debug("Starting deletion iteration for namespace '%s' in database '%s' with prefix: %v", string(ns), dbName, prefix) // 开始删除迭代。
		count := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			keyToDelete := it.Item().KeyCopy(nil)
			if err := writeTxn.Delete(keyToDelete); err != nil {
				log.Error("Failed to delete key %v during drop database '%s' (namespace %s): %v", keyToDelete, dbName, string(ns), err) // 删除 key 失败。
				// Decide if this is a fatal error or if we can continue?
				// 决定这是致命错误还是可以继续？
			}
			count++
			log.Debug("Deleted key: %v", keyToDelete) // 删除 key。
		}
		log.Debug("Finished deletion iteration for namespace '%s' in database '%s'. Deleted %d keys.", string(ns), dbName, count) // 完成删除迭代。删除 %d 个 key。
	}

	// Delete the database marker key last
	// 最后删除数据库标记 key
	dbMetadataKey := bytes.Join([][]byte{NamespaceCatalogBytes, []byte(dbName)}, []byte{NsSep})
	if err := writeTxn.Delete(dbMetadataKey); err != nil {
		log.Error("Failed to delete database marker key for '%s': %v", dbName, err) // 删除数据库标记 key 失败。
		// Decide if this is fatal or ignore if already deleted?
		// 决定这是致命错误还是如果已删除则忽略？
		// For robustness, better return error if it fails.
		return fmt.Errorf("%w: failed to delete database marker key: %v", errors.ErrBadgerOperationFailed, err)
	}
	log.Debug("Deleted database marker key for '%s'", dbName) // 删除数据库标记 key。


	// Commit the transaction
	// 提交事务
	if err := writeTxn.Commit(); err != nil {
		log.Error("Failed to commit transaction for dropping database '%s': %v", dbName, err) // 提交删除数据库事务失败。
		return fmt.Errorf("%w: failed to commit database drop: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Info("Database '%s' dropped successfully.", dbName) // 数据库 '%s' 删除成功。
	return nil
}

// GetDatabase returns a database by name.
// It retrieves database metadata from the catalog.
// GetDatabase 根据名称返回一个数据库。
// 它从 catalog 检索数据库元数据。
func (e *BadgerEngine) GetDatabase(ctx context.Context, dbName string) (interfaces.Database, error) {
	log.Debug("BadgerEngine GetDatabase called for: %s", dbName) // 调用 BadgerEngine GetDatabase。

	// Use the engine's sql.Catalog to find the database.
	// This catalog knows about both user and system databases.
	// 使用引擎的 sql.Catalog 查找数据库。
	// 此 catalog 知道用户数据库和系统数据库。
	sqlCatalog, err := e.GetCatalog(ctx) // Get the GMS-compatible catalog
	if err != nil {
		return nil, fmt.Errorf("failed to get engine catalog: %w", err) // 获取引擎 catalog 失败。
	}

	// Use a GMS context for catalog operations
	// 为 catalog 操作使用 GMS context
	gmsCtx := sql.NewEmptyContext() // Or derive from input ctx if possible

	sqlDB, err := sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		// This error could be sql.ErrDatabaseNotFound or other errors from underlying catalog
		// 此错误可能是 sql.ErrDatabaseNotFound 或底层 catalog 的其他错误
		log.Debug("Database '%s' not found via engine catalog: %v", dbName, err) // 数据库 '%s' 未通过引擎 catalog 找到。
		// Map GMS errors to our errors if needed, or assume compatibility for now.
		// 如果需要，将 GMS 错误映射到我们的错误，或者暂时假设兼容。
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from engine catalog when getting database '%s': %v", dbName, err) // 从引擎 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from engine catalog: %w", err)
	}

	// Found a database (could be GMS system DB or our user DB wrapped)
	// We need to return interfaces.Database, not sql.Database.
	// If it's a Badger-backed DBWrapper, unwrap it to get interfaces.Database.
	// If it's a GMS system DB, we need a wrapper that implements interfaces.Database.
	//
	// 找到一个数据库（可能是 GMS 系统 DB 或我们包装的用户 DB）。
	// 我们需要返回 interfaces.Database，而不是 sql.Database。
	// 如果是 Badger 支持的 DBWrapper，解包它以获取 interfaces.Database。
	// 如果是 GMS 系统 DB，我们需要一个实现 interfaces.Database 的包装器。

	badgerDBWrapper, ok := sqlDB.(*BadgerDatabaseWrapper)
	if ok {
		// It's our wrapped user database, return the underlying interfaces.Database (BadgerDatabase)
		// 它是我们包装的用户数据库，返回底层的 interfaces.Database (BadgerDatabase)
		// The wrapper holds the actual interfaces.Database instance.
		// 包装器持有实际的 interfaces.Database 实例。
		log.Debug("Found user database '%s', returning underlying interfaces.Database instance.", dbName) // 找到用户数据库 '%s'，返回底层 interfaces.Database 实例。
		return badgerDBWrapper.actualDB, nil // Return our concrete implementation
	}

	// It's likely a GMS system database (since it wasn't our wrapper). Wrap it in interfaces.Database.
	// 很可能是 GMS 系统数据库（因为它不是我们的包装器）。将其包装在 interfaces.Database 中。
	log.Debug("Found system database '%s', wrapping it in GmsSystemDatabaseInterfacesWrapper (interfaces.Database).", dbName) // 找到系统数据库 '%s'，将其包装在 GmsSystemDatabaseInterfacesWrapper (interfaces.Database) 中。
	return NewGmsSystemDatabaseInterfacesWrapper(sqlDB), nil // Create this wrapper
}

// ListDatabases returns a list of all database names.
// It gets the list from the engine's sql.Catalog which knows about both user and system databases.
// ListDatabases 返回所有数据库列表。
// 它从引擎的 sql.Catalog 获取列表，该 catalog 知道用户数据库和系统数据库。
func (e *BadgerEngine) ListDatabases(ctx context.Context) ([]string, error) {
	log.Debug("BadgerEngine ListDatabases called.") // 调用 BadgerEngine ListDatabases。

	// Get the list from the engine's sql.Catalog
	// From the perspective of interfaces.StorageEngine, we just need the names.
	// The Catalog implementation should handle getting both user and system names.
	//
	// 从引擎的 sql.Catalog 获取列表。
	// 从 interfaces.StorageEngine 的角度来看，我们只需要名称。
	// Catalog 实现应该负责获取用户和系统名称。

	sqlCatalog, err := e.GetCatalog(ctx) // Get the GMS-compatible catalog
	if err != nil {
		return nil, fmt.Errorf("failed to get engine catalog: %w", err) // 获取引擎 catalog 失败。
	}

	gmsCtx := sql.NewEmptyContext() // Default GMS context
	sqlDBs, err := sqlCatalog.AllDatabases(gmsCtx) // Use catalog to get all DBs (includes system)
	if err != nil {
		log.Error("Failed to get all databases from engine catalog: %v", err) // 从引擎 catalog 获取所有数据库失败。
		// Decide if this is fatal
		// 决定这是否是致命错误
		return nil, fmt.Errorf("failed to list all databases via catalog: %w", err)
	}

	// Extract names from the sql.Database list
	// 从 sql.Database 列表中提取名称
	dbNames := make([]string, len(sqlDBs))
	for i, db := range sqlDBs {
		dbNames[i] = db.Name()
	}

	log.Debug("Total databases found: %d, names: %v", len(dbNames), dbNames) // 找到的总数据库数，名称。
	return dbNames, nil
}


// GetTable returns a table by database and table name.
// Delegates to the respective database instance.
// GetTable 根据数据库名和表名返回一个表。
// 委托给相应的数据库实例。
func (e *BadgerEngine) GetTable(ctx context.Context, dbName, tableName string) (interfaces.Table, error) {
	log.Debug("BadgerEngine GetTable called for %s.%s", dbName, tableName) // 调用 BadgerEngine GetTable。

	// Get the database instance using the engine's GetDatabase (which returns interfaces.Database)
	// 使用引擎的 GetDatabase 获取数据库实例（它返回 interfaces.Database）
	db, err := e.GetDatabase(ctx, dbName)
	if err != nil {
		return nil, err // Database not found or error (already wrapped in our error type)
	}

	// Delegate the GetTable call to the interfaces.Database instance
	// 委托 GetTable 调用给 interfaces.Database 实例
	return db.GetTable(ctx, tableName) // db is already interfaces.Database
}

// CreateTable creates a new table in a database.
// Delegates to the respective database instance.
// CreateTable 在数据库中创建一个新表。
// 委托给相应的数据库实例。
func (e *BadgerEngine) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("BadgerEngine CreateTable called for %s.%s", dbName, tableName) // 调用 BadgerEngine CreateTable。

	// Get the database instance
	// 获取数据库实例
	db, err := e.GetDatabase(ctx, dbName)
	if err != nil {
		return err // Database not found or error
	}

	// Delegate the CreateTable call
	// 委托 CreateTable 调用
	return db.CreateTable(ctx, tableName, schema) // db is interfaces.Database
}

// DropTable drops an existing table in a database.
// Delegates to the respective database instance.
// DropTable 删除数据库中的现有表。
// 委托给相应的数据库实例。
func (e *BadgerEngine) DropTable(ctx context.Context, dbName, tableName string) error {
	log.Info("BadgerEngine DropTable called for %s.%s", dbName, tableName) // 调用 BadgerEngine DropTable。

	// Get the database instance
	// 获取数据库实例
	db, err := e.GetDatabase(ctx, dbName)
	if err != nil {
		return err // Database not found or error
	}

	// Delegate the DropTable call
	// 委托 DropTable 调用
	return db.DropTable(ctx, tableName) // db is interfaces.Database
}

// RenameTable renames a table.
// Delegates to the respective database instance.
// RenameTable 重命名表。
// 委托给相应的数据库实例。
func (e *BadgerEngine) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	log.Info("BadgerEngine RenameTable called for %s.%s -> %s", dbName, oldTableName, newTableName) // 调用 BadgerEngine RenameTable。

	// Get the database instance
	// 获取数据库实例
	db, err := e.GetDatabase(ctx, dbName)
	if err != nil {
		return err // Database not found or error
	}

	// Delegate the RenameTable call
	// 委托 RenameTable 调用
	return db.RenameTable(ctx, oldTableName, newTableName) // db is interfaces.Database
}

// GetCatalog returns the sql.Catalog provided by this engine.
// This catalog allows GMS to access databases and tables managed by Badger.
// GetCatalog 返回此引擎提供的 sql.Catalog。
// 此 catalog 允许 GMS 访问 Badger 管理的数据库和表。
func (e *BadgerEngine) GetCatalog(ctx context.Context) (sql.Catalog, error) {
	log.Debug("BadgerEngine GetCatalog called.") // 调用 BadgerEngine GetCatalog。
	if e.catalog == nil {
		// Should be initialized in Init, but double-check
		// 应该在 Init 中初始化，但要再次检查
		return nil, errors.ErrInternal.New("badger catalog not initialized") // Badger catalog 未初始化。
	}
	return e.catalog, nil
}

// BeginTransaction starts a new transaction.
// It creates a new Badger transaction and wraps it in a BadgerTransaction.
// BeginTransaction 开启一个新的事务。
// 它创建一个新的 Badger 事务并将其包装在 BadgerTransaction 中。
func (e *BadgerEngine) BeginTransaction(ctx context.Context) (interfaces.Transaction, error) {
	log.Debug("BadgerEngine BeginTransaction called.") // 调用 BadgerEngine BeginTransaction。
	if e.db == nil {
		return nil, errors.ErrInternal.New("badger DB not initialized") // Badger DB 未初始化。
	}
	// Badger transactions are created directly from the DB handle
	// Badger 事务直接从 DB 句柄创建
	txn := e.db.NewTransaction(true) // Use read-write transaction by default for DML
	// For SELECT statements, GMS might start a read-only transaction.
	// GMS might handle read-only vs read-write transaction needs.
	// The sql.Handler interface will determine the transaction type.
	// We return a read-write txn here, and GMS/handler decides how to use it.
	//
	// 对于 SELECT 语句，GMS 可能会开始一个只读事务。
	// GMS 可能会处理只读与读写事务的需求。
	// sql.Handler 接口将决定事务类型。
	// 我们在这里返回一个读写事务，GMS/handler 决定如何使用它。
	return NewBadgerTransaction(txn), nil
}

// --- Implement GMS sql.Catalog ---
// --- 实现 GMS sql.Catalog ---

// BadgerCatalog is a wrapper around the Badger DB to implement sql.Catalog.
// It provides access to Badger-backed databases and tables to GMS.
// BadgerCatalog 是围绕 Badger DB 的包装器，用于实现 sql.Catalog。
// 它向 GMS 提供对 Badger 支持的数据库和表的访问。
type BadgerCatalog struct {
	engine *BadgerEngine // Reference to the owning engine
	// systemCatalog is the GMS provided catalog for system databases (mysql, information_schema).
	// systemCatalog 是 GMS 提供的系统数据库 catalog（mysql, information_schema）。
	systemCatalog *sql.Catalog // Pointer to GMS system catalog
}

// NewBadgerCatalog creates a new BadgerCatalog.
// It now takes a reference to the BadgerEngine.
// NewBadgerCatalog 创建一个新的 BadgerCatalog。
// 它现在接受一个 BadgerEngine 的引用。
func NewBadgerCatalog(engine *BadgerEngine) sql.Catalog {
	log.Debug("Creating BadgerCatalog") // 创建 BadgerCatalog。
	// Initialize the GMS system catalog.
	// A real implementation should get this from the GMS engine setup.
	// We need to be able to access mysql_db and information_schema.
	//
	// 初始化 GMS 系统 catalog。
	// 实际实现应该从 GMS 引擎设置中获取。
	// 我们需要能够访问 mysql_db 和 information_schema。
	// Let's create a simple GMS-like memory catalog that includes these.
	// 创建一个包含这些数据库的简单类 GMS 内存 catalog。
	systemCatalog := sql.NewCatalog()
	systemCatalog.AddDatabase(mysql_db.NewMySQLDb()) // Add GMS mysql system database
	systemCatalog.AddDatabase(information_schema.NewInformationSchemaDatabase()) // Add GMS information_schema database


	return &BadgerCatalog{
		engine: engine, // Store engine reference
		systemCatalog: systemCatalog, // Store reference to system catalog
	}
}

// Database returns a sql.Database by name.
// It checks for both user databases in Badger and system databases.
// Database 根据名称返回一个 sql.Database。
// 它检查 Badger 中的用户数据库和系统数据库。
func (c *BadgerCatalog) Database(ctx *sql.Context, dbName string) (sql.Database, error) {
	log.Debug("BadgerCatalog Database called for: %s", dbName) // 调用 BadgerCatalog Database。

	// Check system databases first using the wrapped GMS system catalog
	// 首先使用包装的 GMS 系统 catalog 检查系统数据库
	sysDB, err := c.systemCatalog.Database(ctx, dbName)
	if err == nil {
		log.Debug("Found system database '%s' in GMS catalog.", dbName) // 在 GMS catalog 中找到系统数据库。
		return sysDB, nil // Already implements sql.Database
	}
	// If err is not sql.ErrDatabaseNotFound, return it
	// 如果 err 不是 sql.ErrDatabaseNotFound，则返回
	if !sql.ErrDatabaseNotFound.Is(err) {
		log.Error("Error checking system catalog for database '%s': %v", dbName, err) // 检查系统 catalog 出错。
		return nil, fmt.Errorf("error checking system catalog: %w", err)
	}

	// If not a system database, check user databases via the engine
	// 如果不是系统数据库，通过引擎检查用户数据库
	// Use the engine's GetDatabase method to get interfaces.Database
	// 使用引擎的 GetDatabase 方法获取 interfaces.Database
	// Pass the GMS context down or use context.Background()
	userDB, err := c.engine.GetDatabase(context.Background(), dbName) // Use background context or adapt GMS ctx
	if err != nil {
		if errors.Is(err, errors.ErrDatabaseNotFound) {
			log.Debug("User database '%s' not found via engine.", dbName) // 用户数据库 '%s' 未通过引擎找到。
			return nil, sql.ErrDatabaseNotFound.New(dbName) // Return GMS specific error
		}
		log.Error("Failed to get user database '%s' via engine: %v", dbName, err) // 通过引擎获取用户数据库失败。
		return nil, fmt.Errorf("failed to get user database via engine: %w", err)
	}

	// Found a user database (interfaces.Database). Wrap it in sql.Database.
	// 找到用户数据库 (interfaces.Database)。将其包装在 sql.Database 中。
	log.Debug("Found user database '%s' via engine, wrapping it.", dbName) // 通过引擎找到用户数据库 '%s'，正在包装。
	return NewBadgerDatabaseWrapperFromInterfaces(userDB), nil // Create the wrapper
}

// AllDatabases returns a list of all databases, including user and system databases.
// AllDatabases 返回所有数据库列表，包括用户数据库和系统数据库。
func (c *BadgerCatalog) AllDatabases(ctx *sql.Context) ([]sql.Database, error) {
	log.Debug("BadgerCatalog AllDatabases called.") // 调用 BadgerCatalog AllDatabases。

	var databases []sql.Database

	// Get system databases from the GMS system catalog
	sysDBs, err := c.systemCatalog.AllDatabases(ctx)
	if err != nil {
		log.Error("Failed to get system databases from GMS catalog: %v", err) // 从 GMS catalog 获取系统数据库失败。
		// Continue if possible
	} else {
		databases = append(databases, sysDBs...)
		log.Debug("Added %d system databases.", len(sysDBs)) // 添加 %d 个系统数据库。
	}

	// Get user databases from Badger via the engine's ListDatabases
	// 通过引擎的 ListDatabases 从 Badger 获取用户数据库
	// Pass the GMS context down or use context.Background()
	userDBNames, err := c.engine.ListDatabases(context.Background()) // Use background context or adapt GMS ctx
	if err != nil {
		log.Error("Failed to list user databases via engine: %v", err) // 通过引擎列出用户数据库失败。
		// Decide if this is fatal
		// 决定这是否是致命错误
	} else {
		log.Debug("Found %d user databases via engine: %v", len(userDBNames), userDBNames) // 通过引擎找到 %d 个用户数据库，名称。
		for _, dbName := range userDBNames {
			// Check if it's a system database already added (shouldn't be if engine filters them)
			// 检查它是否是已添加的系统数据库（如果引擎过滤了它们，则不应该出现）
			isSystem := false
			for _, sysDB := range sysDBs {
				if sysDB.Name() == dbName {
					isSystem = true
					break
				}
			}
			if !isSystem {
				// Get the actual database instance via the engine and wrap it
				// 通过引擎获取实际的数据库实例并包装它
				// Use the engine's GetDatabase which returns interfaces.Database
				// 使用引擎的 GetDatabase 获取 interfaces.Database
				userDB, getDBErr := c.engine.GetDatabase(context.Background(), dbName) // Get interfaces.Database
				if getDBErr != nil {
					log.Error("Failed to get user database '%s' during AllDatabases: %v", dbName, getDBErr) // 在 AllDatabases 期间获取用户数据库失败。
					continue // Skip this database on error
				}
				log.Debug("Wrapping user database '%s' for AllDatabases list.", dbName) // 为 AllDatabases 列表包装用户数据库 '%s'。
				databases = append(databases, NewBadgerDatabaseWrapperFromInterfaces(userDB)) // Add wrapped sql.Database
			}
		}
	}


	log.Debug("Total databases found: %d", len(databases)) // 找到的总数据库数。
	return databases, nil
}


// Implement remaining sql.Catalog methods (CreateTable, DropTable, GetTableAsOf)
// by delegating to the engine.
//
// 实现剩余的 sql.Catalog 方法 (CreateTable, DropTable, GetTableAsOf)，
// 通过委托给引擎。

func (c *BadgerCatalog) CreateDatabase(ctx *sql.Context, dbName string) error {
	log.Info("BadgerCatalog CreateDatabase called for: %s", dbName) // 调用 BadgerCatalog CreateDatabase。
	// Delegate to engine's interfaces.StorageEngine method
	// 委托给引擎的 interfaces.StorageEngine 方法
	// Use background context or adapt GMS ctx
	return c.engine.CreateDatabase(context.Background(), dbName)
}

func (c *BadgerCatalog) DropDatabase(ctx *sql.Context, dbName string) error {
	log.Info("BadgerCatalog DropDatabase called for: %s", dbName) // 调用 BadgerCatalog DropDatabase。
	// Delegate to engine's interfaces.StorageEngine method
	// 委托给引擎的 interfaces.StorageEngine 方法
	// Use background context or adapt GMS ctx
	return c.engine.DropDatabase(context.Background(), dbName)
}

func (c *BadgerCatalog) CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("BadgerCatalog CreateTable called for %s.%s", dbName, tableName) // 调用 BadgerCatalog CreateTable。
	// Delegate to engine's interfaces.StorageEngine method
	// 委托给引擎的 interfaces.StorageEngine 方法
	// Use background context or adapt GMS ctx
	return c.engine.CreateTable(context.Background(), dbName, tableName, schema)
}

func (c *BadgerCatalog) DropTable(ctx *sql.Context, dbName, tableName string) error {
	log.Info("BadgerCatalog DropTable called for %s.%s", dbName, tableName) // 调用 BadgerCatalog DropTable。
	// Delegate to engine's interfaces.StorageEngine method
	// 委托给引擎的 interfaces.StorageEngine 方法
	// Use background context or adapt GMS ctx
	return c.engine.DropTable(context.Background(), dbName, tableName)
}

func (c *BadgerCatalog) GetTableAsOf(ctx *sql.Context, dbName, tableName string, asOf sql.Time) (sql.Table, error) {
	log.Warn("BadgerCatalog GetTableAsOf called (Not Implemented).") // 调用 BadgerCatalog GetTableAsOf（未实现）。
	return nil, errors.ErrNotImplemented.New("temporal queries via catalog")
}

// --- Implement GMS sql.Database and sql.Table wrappers ---
// These are needed to expose our interfaces.Database and interfaces.Table
// implementations to the GMS catalog and query engine.
//
// --- 实现 GMS sql.Database 和 sql.Table 包装器 ---
// 需要这些包装器将我们的 interfaces.Database 和 interfaces.Table 实现
// 暴露给 GMS catalog 和查询引擎。

// BadgerDatabaseWrapper is a wrapper around interfaces.Database (BadgerDatabase)
// to implement sql.Database interface expected by GMS Catalog.
// BadgerDatabaseWrapper 是 interfaces.Database (BadgerDatabase) 的包装器，
// 用于实现 GMS Catalog 期望的 sql.Database 接口。
type BadgerDatabaseWrapper struct {
	dbName string // Keep name for logging/debugging
	// actualDB is the underlying interfaces.Database implementation (BadgerDatabase).
	// actualDB 是底层的 interfaces.Database 实现 (BadgerDatabase)。
	actualDB interfaces.Database
}

// NewBadgerDatabaseWrapperFromInterfaces creates a wrapper for interfaces.Database implementing sql.Database.
// NewBadgerDatabaseWrapperFromInterfaces 创建一个实现 sql.Database 的 interfaces.Database 包装器。
func NewBadgerDatabaseWrapperFromInterfaces(db interfaces.Database) sql.Database {
	log.Debug("Creating BadgerDatabaseWrapperFromInterfaces for '%s'", db.Name()) // 创建 BadgerDatabaseWrapperFromInterfaces。
	return &BadgerDatabaseWrapper{
		dbName: db.Name(),
		actualDB: db, // Hold the actual interfaces.Database
	}
}

// Name returns the database name.
// Name 返回数据库名称。
func (w *BadgerDatabaseWrapper) Name() string {
	return w.dbName
}

// GetTable returns a table by name within this database.
// It wraps the interfaces.Table returned by actualDB into sql.Table.
// GetTable 根据名称在此数据库中返回一个表。
// 它将 actualDB 返回的 interfaces.Table 包装为 sql.Table。
func (w *BadgerDatabaseWrapper) GetTable(ctx *sql.Context, tableName string) (sql.Table, error) {
	log.Debug("BadgerDatabaseWrapper GetTable called for %s.%s", w.dbName, tableName) // 调用 BadgerDatabaseWrapper GetTable。

	// Delegate to the actual interfaces.Database instance to get the table
	// 委托给实际的 interfaces.Database 实例获取表
	// Use context.Background() or adapt GMS ctx
	actualTable, err := w.actualDB.GetTable(context.Background(), tableName)
	if err != nil {
		// Propagate the error (e.g., ErrTableNotFound)
		// 传播错误（例如 ErrTableNotFound）
		// Need to check for our error type and return GMS error type if needed.
		// If our ErrTableNotFound wraps GMS ErrTableNotFound, it might work.
		// For now, assume the error is compatible or needs mapping.
		//
		// 需要检查我们的错误类型，并在需要时返回 GMS 错误类型。
		// 如果我们的 ErrTableNotFound 包装了 GMS ErrTableNotFound，它可能会起作用。
		// 目前，假设错误兼容或需要映射。
		if errors.Is(err, errors.ErrTableNotFound) {
			return nil, sql.ErrTableNotFound.New(tableName) // Return GMS specific error
		}
		log.Error("Failed to get table '%s' from actual database '%s': %v", tableName, w.dbName, err) // 从实际数据库获取表失败。
		return nil, fmt.Errorf("failed to get table from actual database: %w", err)
	}

	// Wrap the interfaces.Table (BadgerTable) into a sql.Table (BadgerTableWrapper)
	// 将 interfaces.Table (BadgerTable) 包装为 sql.Table (BadgerTableWrapper)
	return NewBadgerTableWrapperFromInterfaces(actualTable), nil // Create the wrapper
}

// AddTable is part of the sql.Database interface. Not applicable for BadgerDatabaseWrapper.
// AddTable 是 sql.Database 接口的一部分。不适用于 BadgerDatabaseWrapper。
func (w *BadgerDatabaseWrapper) AddTable(tableName string, table sql.Table) {
	log.Warn("BadgerDatabaseWrapper AddTable called for %s.%s (Not Applicable)", w.dbName, tableName) // 调用 BadgerDatabaseWrapper AddTable（不适用）。
	// This method is typically used for in-memory databases.
	// 此方法通常用于内存数据库。
}

// DropTable is part of the sql.Database interface. Not applicable for BadgerDatabaseWrapper.
// DropTable 是 sql.Database 接口的一部分。不适用于 BadgerDatabaseWrapper。
func (w *BadgerDatabaseWrapper) DropTable(tableName string) {
	log.Warn("BadgerDatabaseWrapper DropTable called for %s.%s (Not Applicable)", w.dbName, tableName) // 调用 BadgerDatabaseWrapper DropTable（不适用）。
	// This method is typically used for in-memory databases.
}

// ContainsTable is part of the sql.Database interface.
// Checks if a table exists in this database.
// Delegates to the actual interfaces.Database.
// ContainsTable 是 sql.Database 接口的一部分。
// 检查此数据库中是否存在表。
// 委托给实际的 interfaces.Database。
func (w *BadgerDatabaseWrapper) ContainsTable(name string) (bool, error) {
	log.Debug("BadgerDatabaseWrapper ContainsTable called for %s.%s", w.dbName, name) // 调用 BadgerDatabaseWrapper ContainsTable。
	// Attempt to get the table. If it exists, ContainsTable returns true.
	// 尝试获取表。如果存在，ContainsTable 返回 true。
	// Use context.Background() or adapt GMS ctx
	_, err := w.actualDB.GetTable(context.Background(), name)
	if err != nil {
		if errors.Is(err, errors.ErrTableNotFound) {
			log.Debug("Table '%s' not found by ContainsTable in %s.", name, w.dbName) // ContainsTable 未找到表。
			return false, nil
		}
		log.Error("Failed to check table existence for %s.%s: %v", w.dbName, name, err) // 检查表存在性失败。
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	log.Debug("Table '%s' found by ContainsTable in %s.", name, w.dbName) // ContainsTable 找到表。
	return true, nil
}

// GetTableInsensitive is part of the sql.Database interface. Case-insensitive table lookup.
// Delegates to the actual interfaces.Database if it supports it, or perform case-insensitive search.
// GetTableInsensitive 是 sql.Database 接口的一部分。不区分大小写的表查找。
// 如果实际的 interfaces.Database 支持此功能，则委托，否则执行不区分大小写的搜索。
func (w *BadgerDatabaseWrapper) GetTableInsensitive(ctx *sql.Context, name string) (sql.Table, error) {
	log.Debug("BadgerDatabaseWrapper GetTableInsensitive called for %s.%s", w.dbName, name) // 调用 BadgerDatabaseWrapper GetTableInsensitive。

	// Attempt case-sensitive lookup first using the actual database
	// 首先使用实际数据库尝试区分大小写的查找
	// Use the wrapped GetTable which delegates to actualDB
	table, err := w.GetTable(ctx, name)
	if err == nil {
		return table, nil // Found it case-sensitively
	}
	if !sql.ErrTableNotFound.Is(err) {
		// An error other than not found
		// 不是未找到的错误
		return nil, err
	}

	// If not found case-sensitively, list all tables and search case-insensitively
	// 如果未区分大小写查找未找到，列出所有表并执行不区分大小写搜索
	tableNames, err := w.GetTableNames(ctx) // Use the wrapped GetTableNames
	if err != nil {
		log.Error("Failed to get table names for case-insensitive lookup in %s: %v", w.dbName, err) // 获取表名失败。
		return nil, fmt.Errorf("failed to get table names for case-insensitive lookup: %w", err) // Failed to get table names
	}

	lowerName := strings.ToLower(name)
	for _, tableName := range tableNames {
		if strings.ToLower(tableName) == lowerName {
			// Found a case-insensitive match, get the table case-sensitively
			// 找到不区分大小写的匹配项，以区分大小写的方式获取表
			log.Debug("Found case-insensitive match for '%s': '%s'", name, tableName) // 找到不区分大小写的匹配项。
			return w.GetTable(ctx, tableName) // Get the table using the correct case
		}
	}

	// Not found
	// 未找到
	log.Debug("Case-insensitive table '%s' not found in database '%s'.", name, w.dbName) // 不区分大小写的表未找到。
	return nil, sql.ErrTableNotFound.New(name)
}

// GetTableNames returns a list of table names.
// Delegates to the actual interfaces.Database.
// GetTableNames 返回表名列表。
// 委托给实际的 interfaces.Database。
func (w *BadgerDatabaseWrapper) GetTableNames(ctx *sql.Context) ([]string, error) {
	log.Debug("BadgerDatabaseWrapper GetTableNames called for %s", w.dbName) // 调用 BadgerDatabaseWrapper GetTableNames。
	// Delegate to the actual interfaces.Database instance
	// 委托给实际的 interfaces.Database 实例
	// Use context.Background() or adapt GMS ctx
	return w.actualDB.ListTables(context.Background())
}

// BadgerTableWrapper is a wrapper around interfaces.Table (BadgerTable)
// to implement sql.Table interface expected by GMS.
// BadgerTableWrapper 是 interfaces.Table (BadgerTable) 的包装器，
// 用于实现 GMS 期望的 sql.Table 接口。
type BadgerTableWrapper struct {
	// actualTable is the underlying interfaces.Table implementation (BadgerTable).
	// actualTable 是底层的 interfaces.Table 实现 (BadgerTable)。
	actualTable interfaces.Table
}

// NewBadgerTableWrapperFromInterfaces creates a wrapper for interfaces.Table implementing sql.Table.
// NewBadgerTableWrapperFromInterfaces 创建一个实现 sql.Table 的 interfaces.Table 包装器。
func NewBadgerTableWrapperFromInterfaces(table interfaces.Table) sql.Table {
	log.Debug("Creating BadgerTableWrapperFromInterfaces for table '%s'", table.Name()) // 创建 BadgerTableWrapperFromInterfaces。
	return &BadgerTableWrapper{
		actualTable: table, // Hold the actual interfaces.Table instance
	}
}


// Implement sql.Table methods by delegating to actualTable
// Implementation note: Need to cast ctx to context.Context if actualTable methods use it directly.
//
// 实现 sql.Table 方法，委托给 actualTable。
// 实现说明：如果 actualTable 方法直接使用 context.Context，需要将 ctx 转换为 context.Context。

func (w *BadgerTableWrapper) Name() string { return w.actualTable.Name() }
func (w *BadgerTableWrapper) Schema() sql.Schema { return w.actualTable.Schema() }
func (w *BadgerTableWrapper) Collation() sql.CollationID { return w.actualTable.Collation() }
func (w *BadgerTableWrapper) Comment() string { return w.actualTable.Comment() }

func (w *BadgerTableWrapper) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	log.Debug("BadgerTableWrapper Partitions called for '%s'", w.Name()) // 调用 BadgerTableWrapper Partitions。
	// Delegate to actualTable.Partitions. No need to wrap the result (sql.PartitionIter).
	// The context needs to be passed down or adapted.
	// 委托给 actualTable.Partitions。无需包装结果 (sql.PartitionIter)。
	// 需要将 context 向下传递或适配。
	// Assuming actualTable.Partitions takes context.Context
	return w.actualTable.Partitions(context.Background()) // Use background context or adapt GMS ctx
}

func (w *BadgerTableWrapper) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	log.Debug("BadgerTableWrapper PartitionRows called for '%s' with partition key: %v", w.Name(), partition.Key()) // 调用 BadgerTableWrapper PartitionRows。
	// Delegate to actualTable.PartitionRows. No need to wrap the result (sql.RowIter).
	// The context needs to be passed down or adapted.
	// 委托给 actualTable.PartitionRows。无需包装结果 (sql.RowIter)。
	// 需要将 context 向下传递或适配。
	// Assuming actualTable.PartitionRows takes context.Context
	return w.actualTable.PartitionRows(context.Background(), partition) // Use background context or adapt GMS ctx
}

// IndexedAccess is part of sql.Table interface. Used for index lookups.
// Delegates to the actual interfaces.Table if it implements sql.IndexedTable.
// IndexedAccess 是 sql.Table 接口的一部分。用于索引查找。
// 如果实际的 interfaces.Table 实现 sql.IndexedTable，则委托。
func (w *BadgerTableWrapper) IndexedAccess(lookup sql.IndexLookup) (sql.IndexAccess, error) {
	log.Debug("BadgerTableWrapper IndexedAccess called for '%s'", w.Name()) // 调用 BadgerTableWrapper IndexedAccess。

	// Check if the underlying interfaces.Table implements sql.IndexedTable
	// 检查底层的 interfaces.Table 是否实现 sql.IndexedTable
	indexedTable, ok := w.actualTable.(sql.IndexedTable)
	if !ok {
		log.Warn("Underlying interfaces.Table does not implement sql.IndexedTable for '%s'", w.Name()) // 底层 interfaces.Table 未实现 sql.IndexedTable。
		return nil, sql.ErrTableNotIndexed.New(w.Name()) // Return GMS error
	}
	// Delegate the call
	// 委托调用
	return indexedTable.IndexedAccess(lookup)
}

// GetIndexes returns a list of indexes on the table.
// Delegates to the actual interfaces.Table.
// GetIndexes 返回表上的索引列表。
// 委托给实际的 interfaces.Table。
func (w *BadgerTableWrapper) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	log.Debug("BadgerTableWrapper GetIndexes called for '%s'", w.Name()) // 调用 BadgerTableWrapper GetIndexes。
	// Delegate to actualTable.GetIndexes. The result is already sql.Index list.
	// The context needs to be passed down or adapted.
	// 委托给 actualTable.GetIndexes。结果已经是 sql.Index 列表。
	// 需要将 context 向下传递或适配。
	// Assuming actualTable.GetIndexes takes context.Context
	return w.actualTable.GetIndexes(context.Background()) // Use background context or adapt GMS ctx
}


// --- GMS System Database and Table Wrappers (Actual Implementations) ---
// These are concrete wrappers for GMS internal database types to implement interfaces.Database and interfaces.Table.
//
// --- GMS 系统数据库和表包装器（实际实现） ---
// 这些是 GMS 内部数据库类型的具体包装器，用于实现 interfaces.Database 和 interfaces.Table。

// GmsSystemDatabaseInterfacesWrapper wraps a GMS sql.Database into an interfaces.Database.
// GmsSystemDatabaseInterfacesWrapper 将 GMS sql.Database 包装到 interfaces.Database 中。
type GmsSystemDatabaseInterfacesWrapper struct {
	// gmsDB is the underlying GMS sql.Database (e.g., mysql_db.MySQLDb or information_schema.InformationSchema)
	// gmsDB 是底层的 GMS sql.Database（例如 mysql_db.MySQLDb 或 information_schema.InformationSchema）
	gmsDB sql.Database
}

// NewGmsSystemDatabaseInterfacesWrapper creates a new wrapper.
// NewGmsSystemDatabaseInterfacesWrapper 创建一个新的包装器。
func NewGmsSystemDatabaseInterfacesWrapper(gmsDB sql.Database) interfaces.Database {
	log.Debug("Creating GmsSystemDatabaseInterfacesWrapper for '%s'", gmsDB.Name()) // 创建 GmsSystemDatabaseInterfacesWrapper。
	return &GmsSystemDatabaseInterfacesWrapper{gmsDB: gmsDB}
}

func (w *GmsSystemDatabaseInterfacesWrapper) Name() string { return w.gmsDB.Name() }

// GetTable delegates to the underlying GMS sql.Database and wraps the result.
// GetTable 委托给底层的 GMS sql.Database 并包装结果。
func (w *GmsSystemDatabaseInterfacesWrapper) GetTable(ctx context.Context, tableName string) (interfaces.Table, error) {
	log.Debug("GmsSystemDatabaseInterfacesWrapper GetTable called for %s.%s", w.Name(), tableName) // 调用 GmsSystemDatabaseInterfacesWrapper GetTable。
	// Delegate to GMS sql.Database.GetTable
	// Use sql.NewEmptyContext() or adapt context.Context to *sql.Context.
	// Assuming GMS GetTable takes *sql.Context
	gmsCtx := sql.NewEmptyContext() // Default GMS context
	gmsTable, err := w.gmsDB.GetTable(gmsCtx, tableName)
	if err != nil {
		// Propagate GMS error (e.g., sql.ErrTableNotFound)
		// 传播 GMS 错误（例如 sql.ErrTableNotFound）
		return nil, err // GMS errors are usually compatible with our error checks
	}
	// Wrap the GMS sql.Table into our interfaces.Table
	// 将 GMS sql.Table 包装到我们的 interfaces.Table 中
	return NewGmsSystemTableInterfacesWrapper(gmsTable), nil
}

// ListTables delegates to the underlying GMS sql.Database.GetTableNames.
// ListTables 委托给底层的 GMS sql.Database.GetTableNames。
func (w *GmsSystemDatabaseInterfacesWrapper) ListTables(ctx context.Context) ([]string, error) {
	log.Debug("GmsSystemDatabaseInterfacesWrapper ListTables called for %s", w.Name()) // 调用 GmsSystemDatabaseInterfacesWrapper ListTables。
	// Delegate to GMS sql.Database.GetTableNames
	// Use sql.NewEmptyContext() or adapt context.Context to *sql.Context.
	// Assuming GMS GetTableNames takes *sql.Context
	gmsCtx := sql.NewEmptyContext() // Default GMS context
	return w.gmsDB.GetTableNames(gmsCtx)
}

// CreateTable is not applicable for system databases.
// CreateTable 不适用于系统数据库。
func (w *GmsSystemDatabaseInterfacesWrapper) CreateTable(ctx context.Context, tableName string, schema sql.Schema) error {
	log.Warn("GmsSystemDatabaseInterfacesWrapper CreateTable called for %s.%s (Not Applicable for System DB)", w.Name(), tableName)
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot create tables in system database '%s'", w.Name())) // 无法在系统数据库中创建表。
}

// DropTable is not applicable for system databases.
// DropTable 不适用于系统数据库。
func (w *GmsSystemDatabaseInterfacesWrapper) DropTable(ctx context.Context, tableName string) error {
	log.Warn("GmsSystemDatabaseInterfacesWrapper DropTable called for %s.%s (Not Applicable for System DB)", w.Name(), tableName)
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot drop tables in system database '%s'", w.Name())) // 无法在系统数据库中删除表。
}

// RenameTable is not applicable for system databases.
// RenameTable 不适用于系统数据库。
func (w *GmsSystemDatabaseInterfacesWrapper) RenameTable(ctx context.Context, oldTableName, newTableName string) error {
	log.Warn("GmsSystemDatabaseInterfacesWrapper RenameTable called for %s.%s -> %s (Not Applicable for System DB)", w.Name(), oldTableName, newTableName)
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot rename tables in system database '%s'", w.Name())) // 无法在系统数据库中重命名表。
}


// GmsSystemTableInterfacesWrapper wraps a GMS sql.Table into an interfaces.Table.
// GmsSystemTableInterfacesWrapper 将 GMS sql.Table 包装到 interfaces.Table 中。
type GmsSystemTableInterfacesWrapper struct {
	// gmsTable is the underlying GMS sql.Table.
	// gmsTable 是底层的 GMS sql.Table。
	gmsTable sql.Table
}

// NewGmsSystemTableInterfacesWrapper creates a new wrapper.
// NewGmsSystemTableInterfacesWrapper 创建一个新的包装器。
func NewGmsSystemTableInterfacesWrapper(gmsTable sql.Table) interfaces.Table {
	log.Debug("Creating GmsSystemTableInterfacesWrapper for '%s'", gmsTable.Name()) // 创建 GmsSystemTableInterfacesWrapper。
	return &GmsSystemTableInterfacesWrapper{gmsTable: gmsTable}
}


// Implement interfaces.Table methods by delegating to gmsTable.
// Note: Data manipulation methods (Insert, Update, Delete, Truncate, CreateIndex, DropIndex)
// are not applicable for GMS system tables and should return permission denied errors.
// The read methods (Name, Schema, Collation, Comment, Partitions, PartitionRows,
// IndexedAccess, GetIndexes) should delegate to gmsTable.
//
// 实现 interfaces.Table 方法，委托给 gmsTable。
// 注意：数据操作方法（Insert, Update, Delete, Truncate, CreateIndex, DropIndex）
// 不适用于 GMS 系统表，应返回权限拒绝错误。
// 读取方法（Name, Schema, Collation, Comment, Partitions, PartitionRows,
// IndexedAccess, GetIndexes）应委托给 gmsTable。

func (w *GmsSystemTableInterfacesWrapper) Name() string { return w.gmsTable.Name() }
func (w *GmsSystemTableInterfacesWrapper) Schema() sql.Schema { return w.gmsTable.Schema() }
func (w *GmsSystemTableInterfacesWrapper) Collation() sql.CollationID { return w.gmsTable.Collation() }
func (w *GmsSystemTableInterfacesWrapper) Comment() string { return w.gmsTable.Comment() }

// Data manipulation methods (Not applicable for system tables)
// 数据操作方法（不适用于系统表）
func (w *GmsSystemTableInterfacesWrapper) Insert(ctx context.Context, row sql.Row) error {
	log.Warn("GmsSystemTableInterfacesWrapper Insert called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot insert into system table '%s'", w.Name())) // 无法插入到系统表。
}
func (w *GmsSystemTableInterfacesWrapper) Update(ctx context.Context, oldRow, newRow sql.Row) error {
	log.Warn("GmsSystemTableInterfacesWrapper Update called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot update system table '%s'", w.Name())) // 无法更新系统表。
}
func (w *GmsSystemTableInterfacesWrapper) Delete(ctx context.Context, row sql.Row) error {
	log.Warn("GmsSystemTableInterfacesWrapper Delete called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot delete from system table '%s'", w.Name())) // 无法从系统表删除。
}
func (w *GmsSystemTableInterfacesWrapper) Truncate(ctx context.Context) error {
	log.Warn("GmsSystemTableInterfacesWrapper Truncate called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot truncate system table '%s'", w.Name())) // 无法清空系统表。
}
func (w *GmsSystemTableInterfacesWrapper) CreateIndex(ctx context.Context, indexDef sql.IndexDef) error {
	log.Warn("GmsSystemTableInterfacesWrapper CreateIndex called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot create indexes on system table '%s'", w.Name())) // 无法在系统表上创建索引。
}
func (w *GmsSystemTableInterfacesWrapper) DropIndex(ctx context.Context, indexName string) error {
	log.Warn("GmsSystemTableInterfacesWrapper DropIndex called on system table '%s'", w.Name())
	return errors.ErrPermissionDenied.New(fmt.Sprintf("cannot drop indexes on system table '%s'", w.Name())) // 无法在系统表上删除索引。
}

// Read methods (Delegate to gmsTable)
// Implementation note: Need to cast context.Context to *sql.Context if gmsTable methods use it directly.
//
// 读取方法（委托给 gmsTable）。
// 实现说明：如果 gmsTable 方法直接使用 *sql.Context，需要将 context.Context 转换为 *sql.Context。

func (w *GmsSystemTableInterfacesWrapper) Partitions(ctx context.Context) (sql.PartitionIter, error) {
	log.Debug("GmsSystemTableInterfacesWrapper Partitions called for '%s'", w.Name())
	// Delegate to gmsTable.Partitions. Use sql.NewEmptyContext() or adapt context.Context.
	return w.gmsTable.Partitions(sql.NewEmptyContext())
}
func (w *GmsSystemTableInterfacesWrapper) PartitionRows(ctx context.Context, partition sql.Partition) (sql.RowIter, error) {
	log.Debug("GmsSystemTableInterfacesWrapper PartitionRows called for '%s' with partition key: %v", w.Name(), partition.Key())
	// Delegate to gmsTable.PartitionRows. Use sql.NewEmptyContext() or adapt context.Context.
	return w.gmsTable.PartitionRows(sql.NewEmptyContext(), partition)
}
func (w *GmsSystemTableInterfacesWrapper) IndexedAccess(lookup sql.IndexLookup) (sql.IndexAccess, error) {
	log.Debug("GmsSystemTableInterfacesWrapper IndexedAccess called for '%s'", w.Name())
	// Delegate to gmsTable if it implements sql.IndexedTable
	indexedTable, ok := w.gmsTable.(sql.IndexedTable)
	if !ok {
		log.Warn("Underlying GMS sql.Table does not implement sql.IndexedTable for '%s'", w.Name())
		return nil, sql.ErrTableNotIndexed.New(w.Name()) // Return GMS error
	}
	// Use sql.NewEmptyContext() or adapt context.Context.
	return indexedTable.IndexedAccess(lookup)
}
func (w *GmsSystemTableInterfacesWrapper) GetIndex(ctx context.Context, indexName string) (sql.Index, error) {
	log.Debug("GmsSystemTableInterfacesWrapper GetIndex called for '%s' index: %s", w.Name(), indexName)
	// Delegate if gmsTable implements sql.IndexedTable
	indexedTable, ok := w.gmsTable.(sql.IndexedTable)
	if !ok {
		return nil, sql.ErrTableNotIndexed.New(w.Name()) // Return GMS error
	}
	// Use sql.NewEmptyContext() or adapt context.Context.
	return indexedTable.GetIndex(sql.NewEmptyContext(), indexName)
}
func (w *GmsSystemTableInterfacesWrapper) GetIndexes(ctx context.Context) ([]sql.Index, error) {
	log.Debug("GmsSystemTableInterfacesWrapper GetIndexes called for '%s'", w.Name())
	// Delegate if gmsTable implements sql.IndexedTable
	indexedTable, ok := w.gmsTable.(sql.IndexedTable)
	if !ok {
		return []sql.Index{}, nil // No indexes if not indexed
	}
	// Use sql.NewEmptyContext() or adapt context.Context.
	return indexedTable.GetIndexes(sql.NewEmptyContext())
}


// --- Badger Logging Adapter ---
// --- Badger 日志适配器 ---
// Badger requires a logger that implements dgraph-io/badger/v4.Logger
// Badger 需要一个实现 dgraph-io/badger/v4.Logger 的日志记录器
type BadgerLoggerAdapter struct{}

// NewBadgerLoggerAdapter creates a new BadgerLoggerAdapter.
// NewBadgerLoggerAdapter 创建一个新的 BadgerLoggerAdapter。
func NewBadgerLoggerAdapter() *BadgerLoggerAdapter {
	return &BadgerLoggerAdapter{}
}

func (l *BadgerLoggerAdapter) Errorf(format string, args ...interface{}) {
	log.Error("[Badger] "+format, args...)
}

func (l *BadgerLoggerAdapter) Warningf(format string, args ...interface{}) {
	log.Warn("[Badger] "+format, args...)
}

func (l *BadgerLoggerAdapter) Infof(format string, args ...interface{}) {
	log.Info("[Badger] "+format, args...)
}

func (l *BadgerLoggerAdapter) Debugf(format string, args ...interface{}) {
	log.Debug("[Badger] "+format, args...)
}

// Disable adds a method required by the interface but not strictly needed by our adapter
// Disable 添加了接口要求的方法，但我们的适配器并不严格需要
func (l *BadgerLoggerAdapter) Disable() {}