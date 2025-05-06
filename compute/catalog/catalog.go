// Package catalog manages database metadata for the compute layer.
// It provides interfaces for accessing database schemas, tables, and other metadata
// required by the query analysis and execution components.
//
// catalog 包为计算层管理数据库元数据。
// 它提供了访问查询分析和执行组件所需的数据库模式、表及其他元数据的接口。
package catalog

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
	// Assuming interfaces.StorageEngine might be used by implementations, but not required by the interface itself.
	// 假设 implementations 可能使用 interfaces.StorageEngine，但接口本身不要求。
)

// Catalog is the interface for the compute layer's metadata access.
// It is responsible for providing database and table information
// to the analyzer and executor.
//
// Catalog 是计算层元数据访问的接口。
// 它负责向分析器和执行器提供数据库和表信息。
// This interface aligns with go-mysql-server's sql.Catalog, returning GMS types,
// allowing seamless integration with GMS query processing components.
// It uses standard context.Context for Go idiomatic style.
//
// 此接口与 go-mysql-server 的 sql.Catalog 对齐，返回 GMS 类型，
// 允许与 GMS 查询处理组件无缝集成。
// 它使用标准的 context.Context 以符合 Go 惯例。
type Catalog interface {
	// Database returns a specific database by name.
	// It should handle both user and system databases.
	//
	// Database 根据名称返回特定的数据库。
	// 它应处理用户数据库和系统数据库。
	Database(ctx context.Context, dbName string) (sql.Database, error) // Return GMS sql.Database

	// AllDatabases returns a list of all databases known to the catalog.
	// AllDatabases 返回 catalog 已知的所有数据库列表。
	AllDatabases(ctx context.Context) ([]sql.Database, error) // Return GMS sql.Database list

	// GetTable returns a specific table from a database by name.
	// GetTable 根据名称从数据库中返回特定的表。
	GetTable(ctx context.Context, dbName, tableName string) (sql.Table, error) // Return GMS sql.Table

	// GetTableInsensitive returns a specific table from a database by name, case-insensitively.
	// GetTableInsensitive 根据名称从数据库中不区分大小写地返回特定的表。
	GetTableInsensitive(ctx context.Context, dbName, tableName string) (sql.Table, error) // Return GMS sql.Table

	// CreateDatabase adds a new database to the catalog.
	// This might involve persisting the database metadata if it's a persistent catalog.
	//
	// CreateDatabase 向 catalog 添加新数据库。
	// 如果是持久化 catalog，这可能涉及持久化数据库元数据。
	CreateDatabase(ctx context.Context, dbName string) error

	// DropDatabase removes a database from the catalog.
	// This might involve deleting persisted metadata.
	//
	// DropDatabase 从 catalog 中移除数据库。
	// 这可能涉及删除持久化的元数据。
	DropDatabase(ctx context.Context, dbName string) error

	// CreateTable adds a new table definition to a database in the catalog.
	// This might involve persisting the table schema.
	//
	// CreateTable 向 catalog 中数据库添加新的表定义。
	// 这可能涉及持久化表模式。
	CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error

	// DropTable removes a table definition from a database in the catalog.
	// This might involve deleting persisted schema.
	//
	// DropTable 从 catalog 中数据库移除表定义。
	// 这可能涉及删除持久化模式。
	DropTable(ctx context.Context, dbName, tableName string) error

	// RenameTable renames a table in the catalog.
	// This might involve updating persisted metadata.
	//
	// RenameTable 在 catalog 中重命名表。
	// 这可能涉及更新持久化的元数据。
	RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error


	// GetCatalogAsSQL returns the underlying go-mysql-server sql.Catalog.
	// This method exposes the GMS-compatible view of the catalog, necessary
	// for direct integration with GMS analyzer and executor components.
	//
	// GetCatalogAsSQL 返回底层的 go-mysql-server sql.Catalog。
	// 此方法暴露了 catalog 的 GMS 兼容视图，这是与 GMS 分析器和执行器组件
	// 直接集成所必需的。
	GetCatalogAsSQL(ctx context.Context) (sql.Catalog, error)

	// TODO: Add methods for statistics, caching invalidation, etc., as compute layer needs evolve.
	// TODO: 根据计算层需求演进，添加用于统计信息、缓存失效等的方法。
}