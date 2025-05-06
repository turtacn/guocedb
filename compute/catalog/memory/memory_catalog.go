// Package memory provides an in-memory implementation of the compute/catalog.Catalog interface.
// It is suitable for testing, session-specific temporary objects, or simple scenarios
// where data persistence is not required.
//
// memory 包提供了 compute/catalog.Catalog 接口的内存实现。
// 它适用于不需要数据持久化的测试、会话特定的临时对象或简单场景。
package memory

import (
	"context"
	"strings" // For case-insensitive lookup
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/memory" // Import GMS memory database/table implementations
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/catalog" // Import the compute catalog interface
)

// MemoryCatalog implements the compute/catalog.Catalog interface using an in-memory sql.Catalog.
// It primarily wraps the go-mysql-server in-memory catalog functionality.
//
// MemoryCatalog 使用内存 sql.Catalog 实现 compute/catalog.Catalog 接口。
// 它主要包装 go-mysql-server 的内存 catalog 功能。
type MemoryCatalog struct {
	// sqlCatalog is the underlying go-mysql-server in-memory catalog.
	// sqlCatalog 是底层的 go-mysql-server 内存 catalog。
	sqlCatalog *sql.Catalog
}

// NewMemoryCatalog creates a new MemoryCatalog instance.
// It initializes an empty go-mysql-server in-memory catalog. System databases
// are typically added to the GMS engine's main catalog, which this might be part of.
// For a standalone MemoryCatalog, system DBs might be added here if needed.
//
// NewMemoryCatalog 创建一个新的 MemoryCatalog 实例。
// 它初始化一个空的 go-mysql-server 内存 catalog。系统数据库
// 通常在 GMS 引擎的主 catalog 初始化期间添加，此 catalog 可能属于其中。
// 对于独立的 MemoryCatalog，如果需要，可以在此处添加系统 DB。
func NewMemoryCatalog() catalog.Catalog {
	log.Info("Initializing in-memory catalog.") // 初始化内存 catalog。
	// GMS provides a NewCatalog function for creating an in-memory catalog.
	// GMS 提供了 NewCatalog 函数用于创建内存 catalog。
	gmsCatalog := sql.NewCatalog()
	// Add system databases if this is a standalone catalog, otherwise the composite catalog handles it.
	// If needed, add:
	// gmsCatalog.AddDatabase(mysql_db.NewMySQLDb())
	// gmsCatalog.AddDatabase(information_schema.NewInformationSchemaDatabase())

	return &MemoryCatalog{
		sqlCatalog: gmsCatalog,
	}
}

// Database returns a specific database by name from the in-memory catalog.
// Database 从内存 catalog 中根据名称返回特定的数据库。
// Delegates to the underlying GMS sql.Catalog.
// 委托给底层的 GMS sql.Catalog。
func (c *MemoryCatalog) Database(ctx context.Context, dbName string) (sql.Database, error) {
	log.Debug("MemoryCatalog Database called for: %s", dbName) // 调用 MemoryCatalog Database。
	// GMS catalog methods take *sql.Context. Need to convert or use sql.NewEmptyContext.
	// For simplicity in placeholders, use EmptyContext. Real implementation might adapt context.
	//
	// GMS catalog 方法接受 *sql.Context。需要转换或使用 sql.NewEmptyContext。
	// 为了简化占位符，使用 EmptyContext。实际实现可能会适配 context。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		// Map GMS errors to our errors if necessary, or return GMS error directly.
		// sql.ErrDatabaseNotFound is a common GMS error.
		//
		// 如果需要，将 GMS 错误映射到我们的错误，或直接返回 GMS 错误。
		// sql.ErrDatabaseNotFound 是常见的 GMS 错误。
		if sql.ErrDatabaseNotFound.Is(err) {
			log.Debug("Database '%s' not found in memory catalog.", dbName) // 数据库 '%s' 在内存 catalog 中未找到。
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Wrap or map the error
		}
		log.Error("Error from GMS memory catalog when getting database '%s': %v", dbName, err) // 从 GMS 内存 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from GMS memory catalog: %w", err)
	}
	log.Debug("Database '%s' found in memory catalog.", dbName) // 在内存 catalog 中找到数据库 '%s'。
	return db, nil // Returns GMS sql.Database
}

// AllDatabases returns a list of all databases from the in-memory catalog.
// AllDatabases 从内存 catalog 中返回所有数据库列表。
// Delegates to the underlying GMS sql.Catalog.
// 委托给底层的 GMS sql.Catalog。
func (c *MemoryCatalog) AllDatabases(ctx context.Context) ([]sql.Database, error) {
	log.Debug("MemoryCatalog AllDatabases called.") // 调用 MemoryCatalog AllDatabases。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context
	dbs, err := c.sqlCatalog.AllDatabases(gmsCtx)
	if err != nil {
		log.Error("Failed to get all databases from GMS memory catalog: %v", err) // 从 GMS 内存 catalog 获取所有数据库失败。
		// Decide on error handling policy.
		// 决定错误处理策略。
		return nil, fmt.Errorf("failed to get all databases from memory catalog: %w", err)
	}
	log.Debug("Found %d databases in memory catalog.", len(dbs)) // 在内存 catalog 中找到 %d 个数据库。
	return dbs, nil // Returns GMS sql.Database list
}

// GetTable returns a specific table from a database by name from the in-memory catalog.
// Delegates to the underlying GMS sql.Catalog.
// GetTable 从内存 catalog 中根据名称从数据库返回特定的表。
// 委托给底层的 GMS sql.Catalog。
func (c *MemoryCatalog) GetTable(ctx context.Context, dbName, tableName string) (sql.Table, error) {
	log.Debug("MemoryCatalog GetTable called for %s.%s", dbName, tableName) // 调用 MemoryCatalog GetTable。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying GMS catalog
	// 先使用底层的 GMS catalog 获取数据库
	// We need to get the GMS sql.Database here to call its GetTable method.
	// 我们需要在此处获取 GMS sql.Database 来调用其 GetTable 方法。
	db, err := c.sqlCatalog.Database(gmsCtx, dbName) // Use underlying GMS catalog
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from GMS memory catalog when getting database '%s' for GetTable: %v", dbName, err) // 从 GMS 内存 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from GMS memory catalog (get db for table): %w", err)
	}

	// Get the table from the database using the underlying GMS sql.Database object
	// 使用底层的 GMS sql.Database 对象从数据库获取表
	table, err := db.GetTable(gmsCtx, tableName) // GMS sql.Database.GetTable takes *sql.Context
	if err != nil {
		// Map GMS sql.ErrTableNotFound if necessary
		// 如果需要，映射 GMS sql.ErrTableNotFound
		if sql.ErrTableNotFound.Is(err) {
			log.Debug("Table '%s' not found in memory database '%s'.", tableName, dbName) // 表 '%s' 在内存数据库 '%s' 中未找到。
			return nil, errors.ErrTableNotFound.New(tableName) // Return our error type
		}
		log.Error("Error from GMS memory database '%s' when getting table '%s': %v", dbName, tableName, err) // 从 GMS 内存数据库获取表出错。
		return nil, fmt.Errorf("error from GMS memory database (get table): %w", err)
	}
	log.Debug("Table '%s.%s' found in memory catalog.", dbName, tableName) // 在内存 catalog 中找到表 '%s.%s'。
	return table, nil // Returns GMS sql.Table
}

// GetTableInsensitive returns a specific table from a database by name, case-insensitively, from the in-memory catalog.
// Delegates to the underlying GMS sql.Catalog.
// GetTableInsensitive 从数据库根据名称不区分大小写地返回特定的表，委托给底层的 GMS sql.Catalog。
// 委托给底层的 GMS sql.Catalog。
func (c *MemoryCatalog) GetTableInsensitive(ctx context.Context, dbName, tableName string) (sql.Table, error) {
	log.Debug("MemoryCatalog GetTableInsensitive called for %s.%s", dbName, tableName) // 调用 MemoryCatalog GetTableInsensitive。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying GMS catalog
	// 先使用底层的 GMS catalog 获取数据库
	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from GMS memory catalog when getting database '%s' for GetTableInsensitive: %v", dbName, err) // 从 GMS 内存 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from GMS memory catalog (get db for insensitive table): %w", err)
	}

	// Get the table case-insensitively from the database using the underlying GMS sql.Database object
	// 使用底层的 GMS sql.Database 对象从数据库不区分大小写地获取表
	// GMS sql.Database has GetTableInsensitive
	// GMS sql.Database 有 GetTableInsensitive
	table, err := db.GetTableInsensitive(gmsCtx, tableName)
	if err != nil {
		if sql.ErrTableNotFound.Is(err) {
			log.Debug("Case-insensitive table '%s' not found in memory database '%s'.", tableName, dbName) // 不区分大小写的表 '%s' 在内存数据库 '%s' 中未找到。
			return nil, errors.ErrTableNotFound.New(tableName) // Return our error type
		}
		log.Error("Error from GMS memory database '%s' when getting table insensitive '%s': %v", dbName, tableName, err) // 从 GMS 内存数据库获取不区分大小写的表出错。
		return nil, fmt.Errorf("error from GMS memory database (get insensitive table): %w", err)
	}
	log.Debug("Case-insensitive table '%s.%s' found in memory catalog.", dbName, tableName) // 在内存 catalog 中找到不区分大小写的表 '%s.%s'。
	return table, nil // Returns GMS sql.Table
}


// GetCatalogAsSQL returns the underlying go-mysql-server sql.Catalog.
// GetCatalogAsSQL 返回底层的 go-mysql-server sql.Catalog。
func (c *MemoryCatalog) GetCatalogAsSQL(ctx context.Context) (sql.Catalog, error) {
	log.Debug("MemoryCatalog GetCatalogAsSQL called.") // 调用 MemoryCatalog GetCatalogAsSQL。
	// Return the underlying GMS catalog directly.
	// 返回底层的 GMS catalog。
	return c.sqlCatalog, nil
}


// --- DDL Methods ---
// Implement DDL methods by interacting with the underlying GMS sql.Catalog directly.
//
// --- DDL 方法 ---
// 通过直接与底层的 GMS sql.Catalog 交互来实现 DDL 方法。

// CreateDatabase adds a new database to the in-memory catalog.
// CreateDatabase 向内存 catalog 添加新数据库。
// Creates a GMS memory.Database and adds it to the underlying GMS catalog.
// 创建一个 GMS memory.Database 并将其添加到底层的 GMS catalog。
func (c *MemoryCatalog) CreateDatabase(ctx context.Context, dbName string) error {
	log.Info("MemoryCatalog CreateDatabase called for: %s", dbName) // 调用 MemoryCatalog CreateDatabase。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Check if database already exists using the underlying GMS catalog
	// 使用底层的 GMS catalog 检查数据库是否已存在
	_, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err == nil {
		log.Warn("Database '%s' already exists in memory catalog.", dbName) // 数据库 '%s' 在内存 catalog 中已存在。
		return errors.ErrDatabaseAlreadyExists.New(dbName)
	}
	if !sql.ErrDatabaseNotFound.Is(err) {
		log.Error("Failed to check database existence in memory catalog for '%s': %v", dbName, err) // 检查内存 catalog 中数据库存在性失败。
		return fmt.Errorf("failed to check database existence in memory catalog: %w", err)
	}

	// Database does not exist, create a GMS memory.Database and add it.
	// 数据库不存在，创建一个 GMS memory.Database 并添加。
	newDB := memory.NewDatabase(dbName) // Create GMS memory database
	c.sqlCatalog.AddDatabase(newDB)     // Add to underlying GMS catalog

	log.Info("Database '%s' created in memory catalog.", dbName) // 数据库 '%s' 在内存 catalog 中创建成功。
	return nil
}

// DropDatabase removes a database from the in-memory catalog.
// DropDatabase 从内存 catalog 中移除数据库。
// Removes the database from the underlying GMS catalog.
// 从底层的 GMS catalog 中移除数据库。
func (c *MemoryCatalog) DropDatabase(ctx context.Context, dbName string) error {
	log.Info("MemoryCatalog DropDatabase called for: %s", dbName) // 调用 MemoryCatalog DropDatabase。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Check if database exists using the underlying GMS catalog
	// 使用底层的 GMS catalog 检查数据库是否存在
	_, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if sql.ErrDatabaseNotFound.Is(err) {
		log.Warn("Database '%s' not found in memory catalog for dropping.", dbName) // 删除数据库时，在内存 catalog 中未找到。
		return errors.ErrDatabaseNotFound.New(dbName)
	} else if err != nil {
		log.Error("Failed to check database existence in memory catalog for drop '%s': %v", dbName, err) // 检查内存 catalog 中数据库存在性失败。
		return fmt.Errorf("failed to check database existence in memory catalog for drop: %w", err)
	}

	// Drop the database from the GMS catalog
	// 从 GMS catalog 删除数据库
	c.sqlCatalog.RemoveDatabase(dbName) // GMS catalog has RemoveDatabase
	log.Info("Database '%s' removed from memory catalog.", dbName) // 数据库 '%s' 已从内存 catalog 中移除。
	return nil
}

// CreateTable adds a new table definition to a database in the in-memory catalog.
// CreateTable 向内存 catalog 中数据库添加新的表定义。
// Creates a GMS memory.Table and adds it to the specified memory.Database.
// 创建一个 GMS memory.Table 并将其添加到指定的 memory.Database。
func (c *MemoryCatalog) CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error {
	log.Info("MemoryCatalog CreateTable called for %s.%s", dbName, tableName) // 调用 MemoryCatalog CreateTable。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying GMS catalog
	// 先使用底层的 GMS catalog 获取数据库
	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from GMS memory catalog when getting database '%s' for CreateTable: %v", dbName, err) // 从 GMS 内存 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from GMS memory catalog (get db for create table): %w", err)
	}

	// Check if the database is a GMS memory.Database (or a wrapper that supports adding tables)
	// 检查数据库是否是 GMS memory.Database（或支持添加表的包装器）
	memoryDB, ok := db.(sql.Database) // Should be sql.Database already
	if !ok {
		// This should not happen if using GMS memory.Database, but check for safety.
		// 如果使用 GMS memory.Database，这不应该发生，但为了安全起见检查。
		log.Error("Database '%s' in memory catalog is not a GMS sql.Database or wrapper supporting table creation.", dbName) // 内存 catalog 中的数据库不是 GMS sql.Database 或支持表创建的包装器。
		return errors.ErrInternal.New(fmt.Sprintf("database '%s' does not support table creation", dbName))
	}

	// Check if table already exists using the underlying GMS sql.Database
	// 使用底层的 GMS sql.Database 检查表是否已存在
	_, err = memoryDB.GetTable(gmsCtx, tableName)
	if err == nil {
		log.Warn("Table '%s' already exists in memory database '%s'.", tableName, dbName) // 表 '%s' 在内存数据库 '%s' 中已存在。
		return errors.ErrTableAlreadyExists.New(tableName)
	}
	if !sql.ErrTableNotFound.Is(err) {
		log.Error("Failed to check table existence in memory database '%s' for CreateTable '%s': %v", dbName, tableName, err) // 检查内存数据库中表存在性失败。
		return fmt.Errorf("failed to check table existence in memory database: %w", err)
	}


	// Create a GMS memory.Table and add it to the database
	// 创建一个 GMS memory.Table 并将其添加到数据库
	newTable := memory.NewTable(tableName, schema) // Create GMS memory table
	memoryDB.AddTable(tableName, newTable)        // Add to GMS memory database

	log.Info("Table '%s' created in memory database '%s'.", tableName, dbName) // 表 '%s' 在内存数据库 '%s' 中创建成功。
	return nil
}

// DropTable removes a table definition from a database in the in-memory catalog.
// DropTable 从内存 catalog 中数据库移除表定义。
// Removes the table from the specified memory.Database.
// 从指定的 memory.Database 中移除表。
func (c *MemoryCatalog) DropTable(ctx context.Context, dbName, tableName string) error {
	log.Info("MemoryCatalog DropTable called for %s.%s", dbName, tableName) // 调用 MemoryCatalog DropTable。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first using the underlying GMS catalog
	// 先使用底层的 GMS catalog 获取数据库
	db, err := c.sqlCatalog.Database(gmsCtx, dbName)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, errors.ErrDatabaseNotFound.New(dbName) // Return our error type
		}
		log.Error("Error from GMS memory catalog when getting database '%s' for DropTable: %v", dbName, err) // 从 GMS 内存 catalog 获取数据库出错。
		return nil, fmt.Errorf("error from GMS memory catalog (get db for drop table): %w", err)
	}

	// Check if the database is a GMS memory.Database (or a wrapper that supports dropping tables)
	// 检查数据库是否是 GMS memory.Database（或支持删除表的包装器）
	memoryDB, ok := db.(sql.Database) // Should be sql.Database already
	if !ok {
		log.Error("Database '%s' in memory catalog is not a GMS sql.Database or wrapper supporting table dropping.", dbName) // 内存 catalog 中的数据库不是 GMS sql.Database 或支持表删除的包装器。
		return errors.ErrInternal.New(fmt.Sprintf("database '%s' does not support table dropping", dbName))
	}

	// Check if table exists using the underlying GMS sql.Database
	// 使用底层的 GMS sql.Database 检查表是否已存在
	_, err = memoryDB.GetTable(gmsCtx, tableName)
	if sql.ErrTableNotFound.Is(err) {
		log.Warn("Table '%s' not found in memory database '%s' for dropping.", tableName, dbName) // 删除表时，在内存数据库中未找到表。
		return errors.ErrTableNotFound.New(tableName)
	} else if err != nil {
		log.Error("Failed to check table existence in memory database '%s' for DropTable '%s': %v", dbName, tableName, err) // 检查内存数据库中表存在性失败。
		return fmt.Errorf("failed to check table existence in memory database for drop: %w", err)
	}


	// Drop the table from the database
	// 从数据库删除表
	memoryDB.DropTable(tableName) // GMS sql.Database has DropTable

	log.Info("Table '%s' dropped from memory database '%s'.", tableName, dbName) // 表 '%s' 已从内存数据库 '%s' 中删除。
	return nil
}

// RenameTable renames a table in the in-memory catalog.
// Delegates to the underlying GMS memory.Database if it supports renaming.
//
// RenameTable 在内存 catalog 中重命名表。
// 如果底层的 GMS memory.Database 支持重命名，则委托。
func (c *MemoryCatalog) RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error {
	log.Info("MemoryCatalog RenameTable called for %s.%s -> %s", dbName, oldTableName, newTableName) // 调用 MemoryCatalog RenameTable。
	gmsCtx := sql.NewEmptyContext() // Use a basic GMS context

	// Get the database first
	// 先获取数据库
	db, err := c.Database(ctx, dbName) // Use our Database method
	if err != nil {
		return err // Database not found or error
	}

	// Check if the database supports renaming tables (e.g., GMS memory.Database does not directly)
	// Check if the database is a memory.Database and cast it.
	//
	// 检查数据库是否支持重命名表（例如，GMS memory.Database 不直接支持）
	// 检查数据库是否为 memory.Database 并进行类型转换。
	memoryDB, ok := db.(*memory.Database) // Check for the specific GMS memory.Database type
	if !ok {
		log.Warn("Database '%s' in memory catalog is not a GMS memory.Database supporting table renaming.", dbName) // 内存 catalog 中的数据库不是支持表重命名的 GMS memory.Database。
		return errors.ErrNotImplemented.New("rename table in non-memory database")
	}

	// GMS memory.Database does not have a RenameTable method.
	// If this is for session temporary tables, renaming might involve complex AddTable/DropTable logic.
	// For now, mark as not implemented for this MemoryCatalog.
	//
	// GMS memory.Database 没有 RenameTable 方法。
	// 如果这是用于会话临时表，重命名可能涉及复杂的 AddTable/DropTable 逻辑。
	// 目前，将此 MemoryCatalog 中的重命名标记为未实现。
	log.Warn("GMS memory.Database does not support RenameTable. Operation failed for %s.%s -> %s", dbName, oldTableName, newTableName) // GMS memory.Database 不支持 RenameTable。操作失败。
	return errors.ErrNotImplemented.New("rename table in memory database")
}