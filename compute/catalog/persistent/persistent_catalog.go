// Package persistent provides a persistent implementation of the compute/catalog.Catalog interface.
// It delegates metadata management (including DDL operations) to the storage layer's sql.Catalog
// and the underlying interfaces.StorageEngine.
//
// persistent 包提供了 compute/catalog.Catalog 接口的持久化实现。
// 它将元数据管理（包括 DDL 操作）委托给存储层的 sql.Catalog
// 和底层的 interfaces.StorageEngine。
package persistent

import (
	"context"
	"fmt"
	"strings" // For case-insensitive lookup

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/catalog" // Import the compute catalog interface
	"github.com/turtacn/guocedb/interfaces" // Import storage interfaces
)

// PersistentCatalog implements the compute/catalog.Catalog interface by delegating to the storage layer's sql.Catalog and interfaces.StorageEngine.
// PersistentCatalog 通过委托给存储层的 sql.Catalog 和 interfaces.StorageEngine 来实现 compute/catalog.Catalog 接口。
type PersistentCatalog struct {
	// storageEngine is the underlying storage engine providing the persistent catalog.
	// It is used for DDL operations (Create/Drop/Rename).
	//
	// storageEngine 是提供持久化 catalog 的底层存储引擎。
	// 它用于 DDL 操作（创建/删除/重命名）。
	storageEngine interfaces.StorageEngine

	// sqlCatalog is the sql.Catalog obtained from the storage engine.
	// It is used for read operations (Database/AllDatabases/GetTable/GetTableInsensitive).
	//
	// sqlCatalog 是从存储引擎获取的 sql.Catalog。
	// 它用于读取操作（Database/AllDatabases/GetTable/GetTableInsensitive）。
	sqlCatalog sql.Catalog
}

// NewPersistentCatalog creates a new PersistentCatalog instance.
// It obtains the sql.Catalog from the provided storage engine.
// NewPersistentCatalog 创建一个新的 PersistentCatalog 实例。
// 它从提供的存储引擎获取 sql.Catalog。
func NewPersistentCatalog(ctx context.Context, storageEngine interfaces.StorageEngine) (catalog.Catalog, error) {
	log.Info("Initializing persistent catalog from storage engine.") // 从存储引擎初始化持久化 catalog。

	// Get the sql.Catalog from the storage engine
	// This catalog will handle mixing user (Badger-backed) and system (GMS-provided) databases.
	//
	// 从存储引擎获取 sql.Catalog。
	// 此 catalog 将处理混合用户（Badger 支持的）和系统（GMS 提供的）数据库。
	sqlCatalog, err := storageEngine.GetCatalog(ctx)
	if err != nil {
		log.Error("Failed to get sql.Catalog from storage engine: %v", err) // 从存储引擎获取 sql.Catalog 失败。
		return nil, fmt.Errorf("failed to get sql.Catalog from storage engine: %w", err)
	}

	return &PersistentCatalog{
		storageEngine: storageEngine,
		sqlCatalog: sqlCatalog,
	}, nil
}

// Database returns a specific database by name, delegating to the underlying sql.Catalog.
// Database 根据名称返回特定的数据库，委托给底层的 sql.Catalog。
func (c *PersistentCatalog) Database(ctx context.Context, dbName string) (sql.Database, error) {
	log.Debug("PersistentCatalog Database called for: %s", dbName) // 调用 PersistentCatalog Database。
	// Delegate to the underlying sql.Catalog.
	// GMS sql.Catalog methods take *sql.Context. Need to convert or adapt.
	//
	// 委托给底层的 sql.Catalog。
	// GMS sql.Catalog 方法接受 *sql.Context。需要转换或适配。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		// Map GMS errors to our errors if necessary.
		// 如果需要，将 GMS 错误映射到我们的错误。
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from underlying sql.Catalog when getting database '%s': %v", dbName, err) // 从底层 sql.Catalog 获取数据库出错。
		return nil, fmt.Errorf("error from underlying sql.Catalog: %w", err)
	}
	log.Debug("Database '%s' found by persistent catalog.", dbName) // 持久化 catalog 找到数据库 '%s'。
	return db, nil // Returns GMS sql.Database (which might be a wrapper from the storage layer)
}

// AllDatabases returns a list of all databases, delegating to the underlying sql.Catalog.
// AllDatabases 返回所有数据库列表，委托给底层的 sql.Catalog。
func (c *PersistentCatalog) AllDatabases(ctx context.Context) ([]sql.Database, error) {
	log.Debug("PersistentCatalog AllDatabases called.") // 调用 PersistentCatalog AllDatabases。
	// Delegate to the underlying sql.Catalog.
	// GMS sql.Catalog.AllDatabases takes *sql.Context. Need to convert or adapt.
	// 委托给底层的 sql.Catalog。
	// GMS sql.Catalog.AllDatabases 接受 *sql.Context。需要转换或适配。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	dbs, err := c.sqlCatalog.AllDatabases(gmsCtx)
	if err != nil {
		log.Error("Failed to get all databases from underlying sql.Catalog: %v", err) // 从底层 sql.Catalog 获取所有数据库失败。
		return nil, fmt.Errorf("failed to get all databases from underlying sql.Catalog: %w", err)
	}
	log.Debug("Found %d databases by persistent catalog.", len(dbs)) // 持久化 catalog 找到 %d 个数据库。
	return dbs, nil // Returns GMS sql.Database list
}

// GetTable returns a specific table from a database by name, delegating to the underlying sql.Catalog.
// GetTable 从数据库根据名称返回特定的表，委托给底层的 sql.Catalog。
func (c *PersistentCatalog) GetTable(ctx context.Context, dbName, tableName string) (sql.Table, error) {
	log.Debug("PersistentCatalog GetTable called for %s.%s", dbName, tableName) // 调用 PersistentCatalog GetTable。
	// Delegate to the underlying sql.Catalog.
	// GMS sql.Catalog.Database takes *sql.Context. GetTable takes *sql.Context.
	//
	// 委托给底层的 sql.Catalog。
	// GMS sql.Catalog.Database 接受 *sql.Context。GetTable 接受 *sql.Context。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying catalog
	// 先使用底层 catalog 获取数据库
	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from underlying sql.Catalog when getting database '%s' for GetTable: %v", dbName, err) // 从底层 sql.Catalog 获取数据库出错。
		return nil, fmt.Errorf("error from underlying sql.Catalog (get db for table): %w", err)
	}

	// Get the table from the database using the underlying catalog's db object
	// 使用底层 catalog 的 db 对象从数据库获取表
	table, err := db.GetTable(gmsCtx, tableName)
	if err != nil {
		if sql.ErrTableNotFound.Is(err) {
			return nil, errors.ErrTableNotFound.New(tableName) // Return our error type
		}
		log.Error("Error from underlying sql.Database when getting table '%s.%s': %v", dbName, tableName, err) // 从底层 sql.Database 获取表出错。
		return nil, fmt.Errorf("error from underlying sql.Database (get table): %w", err)
	}
	log.Debug("Table '%s.%s' found by persistent catalog.", dbName, tableName) // 持久化 catalog 找到表 '%s.%s'。
	return table, nil // Returns GMS sql.Table (which might be a wrapper from the storage layer)
}

// GetTableInsensitive returns a specific table from a database by name, case-insensitively, delegating to the underlying sql.Catalog.
// GetTableInsensitive 从数据库根据名称不区分大小写地返回特定的表，委托给底层的 sql.Catalog。
func (c *PersistentCatalog) GetTableInsensitive(ctx context.Context, dbName, tableName string) (sql.Table, error) {
	log.Debug("PersistentCatalog GetTableInsensitive called for %s.%s", dbName, tableName) // 调用 PersistentCatalog GetTableInsensitive。
	// Delegate to the underlying sql.Catalog.
	// GMS sql.Catalog.Database takes *sql.Context. GetTableInsensitive takes *sql.Context.
	//
	// 委托给底层的 sql.Catalog。
	// GMS sql.Catalog.Database 接受 *sql.Context。GetTableInsensitive 接受 *sql.Context。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying catalog
	// 先使用底层 catalog 获取数据库
	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from underlying sql.Catalog when getting database '%s' for GetTableInsensitive: %v", dbName, err) // 从底层 sql.Catalog 获取数据库出错。
		return nil, fmt.Errorf("error from underlying sql.Catalog (get db for insensitive table): %w", err)
	}

	// Get the table case-insensitively from the database using the underlying catalog's db object
	// 使用底层 catalog 的 db 对象从数据库不区分大小写地获取表
	// GMS sql.Database has GetTableInsensitive
	// GMS sql.Database 有 GetTableInsensitive
	table, err := db.GetTableInsensitive(gmsCtx, tableName)
	if err != nil {
		if sql.ErrTableNotFound.Is(err) {
			return nil, errors.ErrTableNotFound.New(tableName) // Return our error type
		}
		log.Error("Error from underlying sql.Database when getting table insensitive '%s.%s': %v", dbName, tableName, err) // 从底层 sql.Database 获取不区分大小写的表出错。
		return nil, fmt.Errorf("error from underlying sql.Database (get insensitive table): %w", err)
	}
	log.Debug("Case-insensitive table '%s.%s' found by persistent catalog.", dbName, tableName) // 持久化 catalog 找到不区分大小写的表 '%s.%s'。
	return table, nil // Returns GMS sql.Table
}

// GetCatalogAsSQL returns the underlying go-mysql-server sql.Catalog.
// GetCatalogAsSQL 返回底层的 go-mysql-server sql.Catalog。
func (c *PersistentCatalog) GetCatalogAsSQL(ctx context.Context) (sql.Catalog, error) {
	log.Debug("PersistentCatalog GetCatalogAsSQL called.") // 调用 PersistentCatalog GetCatalogAsSQL。
	// Return the underlying sql.Catalog obtained from the storage engine.
	// 返回从存储引擎获取的底层 sql.Catalog。
	return c.sqlCatalog, nil
}

// --- DDL Methods ---
// Delegate DDL methods to the underlying storage engine.
// The storage engine is responsible for persisting these changes and ensuring
// the sqlCatalog (obtained from the storage engine) reflects them.
//
// --- DDL 方法 ---
// 将 DDL 方法委托给底层的存储引擎。
// 存储引擎负责持久化这些更改，并确保
// sqlCatalog（从存储引擎获取的）反映这些更改。

// CreateDatabase delegates the call to the underlying storage engine.
// CreateDatabase 将调用委托给底层的存储引擎。
func (c *PersistentCatalog) CreateDatabase(ctx context.Context, dbName string) error {
	log.Info("PersistentCatalog CreateDatabase called for: %s", dbName) // 调用 PersistentCatalog CreateDatabase。
	// Delegate to the storage engine's interfaces.StorageEngine method
	// 委托给存储引擎的 interfaces.StorageEngine 方法
	return c.storageEngine.CreateDatabase(ctx, dbName)
}

// DropDatabase delegates the call to the underlying storage engine.
// DropDatabase 将调用委托给底层的存储引擎。
func (c *PersistentCatalog) DropDatabase(ctx context.Context, dbName string) error {
	log.Info("PersistentCatalog DropDatabase called for: %s", dbName) // 调用 PersistentCatalog DropDatabase。
	// Delegate to the storage engine's interfaces.StorageEngine method
	// 委托给存储引擎的 interfaces.StorageEngine 方法
	return c.storageEngine.DropDatabase(ctx, dbName)
}

// CreateTable delegates the call to the underlying storage engine.
// CreateTable 将调用委托给底层的存储引擎。
func (c *PersistentCatalog) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("PersistentCatalog CreateTable called for %s.%s", dbName, tableName) // 调用 PersistentCatalog CreateTable。
	// Delegate to the storage engine's interfaces.StorageEngine method
	// 委托给存储引擎的 interfaces.StorageEngine 方法
	return c.storageEngine.CreateTable(ctx, dbName, tableName, schema)
}

// DropTable delegates the call to the underlying storage engine.
// DropTable 将调用委托给底层的存储引擎。
func (c *PersistentCatalog) DropTable(ctx context.Context, dbName, tableName string) error {
	log.Info("PersistentCatalog DropTable called for %s.%s", dbName, tableName) // 调用 PersistentCatalog DropTable。
	// Delegate to the storage engine's interfaces.StorageEngine method
	// 委托给存储引擎的 interfaces.StorageEngine 方法
	return c.storageEngine.DropTable(ctx, dbName, tableName)
}

// RenameTable delegates the call to the underlying storage engine.
// RenameTable 将调用委托给底层的存储引擎。
func (c *PersistentCatalog) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	log.Info("PersistentCatalog RenameTable called for %s.%s -> %s", dbName, oldTableName, newTableName) // 调用 PersistentCatalog RenameTable。
	// Delegate to the storage engine's interfaces.StorageEngine method
	// 委托给存储引擎的 interfaces.StorageEngine 方法
	return c.storageEngine.RenameTable(ctx, dbName, oldTableName, newTableName)
}