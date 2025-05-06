// Package interfaces defines core abstraction interface definitions.
// interfaces 包定义了核心抽象接口定义。
package interfaces

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
)

// StorageEngine is the interface that all storage engines must implement.
// It represents the capability to manage databases, tables, and data persistence.
// StorageEngine 是所有存储引擎必须实现的接口。
// 它代表管理数据库、表和数据持久化的能力。
type StorageEngine interface {
	// Init initializes the storage engine.
	// Init 初始化存储引擎。
	Init(ctx context.Context, config map[string]string) error

	// Close closes the storage engine.
	// Close 关闭存储引擎。
	Close(ctx context.Context) error

	// CreateDatabase creates a new database.
	// CreateDatabase 创建一个新数据库。
	CreateDatabase(ctx context.Context, dbName string) error

	// DropDatabase drops an existing database.
	// DropDatabase 删除现有数据库。
	DropDatabase(ctx context.Context, dbName string) error

	// GetDatabase returns a database by name.
	// GetDatabase 根据名称返回一个数据库。
	GetDatabase(ctx context.Context, dbName string) (Database, error)

	// ListDatabases returns a list of all databases.
	// ListDatabases 返回所有数据库的列表。
	ListDatabases(ctx context.Context) ([]string, error)

	// GetTable returns a table by database and table name.
	// GetTable 根据数据库名和表名返回一个表。
	GetTable(ctx context.Context, dbName, tableName string) (Table, error)

	// CreateTable creates a new table in a database.
	// CreateTable 在数据库中创建一个新表。
	CreateTable(ctx context.Context, dbName, tableName string, schema sql.Schema) error

	// DropTable drops an existing table in a database.
	// DropTable 删除数据库中的现有表。
	DropTable(ctx context.Context, dbName, tableName string) error

	// RenameTable renames a table.
	// RenameTable 重命名表。
	RenameTable(ctx context.Context, dbName, oldTableName, newTableName string) error

	// GetCatalog returns the catalog managed by this storage engine.
	// This catalog should provide access to database and table metadata.
	// GetCatalog 返回此存储引擎管理的目录。
	// 此目录应提供对数据库和表元数据的访问。
	// Note: In some designs, the catalog might be managed externally or be part of the Compute Layer.
	// For Guocedb, the Storage Engine is responsible for providing access to its own metadata.
	// 注意：在某些设计中，目录可能由外部管理或属于计算层。
	// 对于 Guocedb，存储引擎负责提供对其自身元数据的访问。
	GetCatalog(ctx context.Context) (sql.Catalog, error) // GMS requires sql.Catalog interface

	// BeginTransaction starts a new transaction.
	// BeginTransaction 开启一个新的事务。
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// Database is the interface representing a database within a storage engine.
// Database 是代表存储引擎内数据库的接口。
type Database interface {
	// Name returns the name of the database.
	// Name 返回数据库的名称。
	Name() string

	// GetTable returns a table by name within this database.
	// GetTable 根据名称在此数据库中返回一个表。
	GetTable(ctx context.Context, tableName string) (Table, error)

	// ListTables returns a list of all table names in this database.
	// ListTables 返回此数据库中所有表名的列表。
	ListTables(ctx context.Context) ([]string, error)

	// CreateTable creates a new table in this database.
	// CreateTable 在此数据库中创建一个新表。
	CreateTable(ctx context.Context, tableName string, schema sql.Schema) error

	// DropTable drops an existing table in this database.
	// DropTable 删除此数据库中的现有表。
	DropTable(ctx context.Context, tableName string) error

	// RenameTable renames a table in this database.
	// RenameTable 在此数据库中重命名表。
	RenameTable(ctx context.Context, oldTableName, newTableName string) error
}

// Table is the interface representing a table within a database.
// Table 是代表数据库内表的接口。
// It extends the GMS sql.Table interface.
// 它扩展了 GMS 的 sql.Table 接口。
type Table interface {
	sql.Table // Inherit methods from GMS sql.Table like Name() and Schema()

	// Insert inserts a new row into the table.
	// Insert 向表中插入新行。
	Insert(ctx context.Context, row sql.Row) error

	// Update updates an existing row in the table.
	// Update 更新表中的现有行。
	Update(ctx context.Context, oldRow, newRow sql.Row) error

	// Delete deletes a row from the table.
	// Delete 从表中删除行。
	Delete(ctx context.Context, row sql.Row) error

	// Truncate removes all rows from the table.
	// Truncate 删除表中的所有行。
	Truncate(ctx context.Context) error

	// CreateIndex creates a new index on the table.
	// CreateIndex 在表上创建一个新索引。
	CreateIndex(ctx context.Context, indexDef sql.IndexDef) error

	// DropIndex drops an existing index from the table.
	// DropIndex 从表中删除现有索引。
	DropIndex(ctx context.Context, indexName string) error

	// GetIndex returns an index by name.
	// GetIndex 根据名称返回一个索引。
	GetIndex(ctx context.Context, indexName string) (sql.Index, error) // GMS requires sql.Index

	// GetIndexes returns a list of all indexes on the table.
	// GetIndexes 返回表上所有索引的列表。
	GetIndexes(ctx context.Context) ([]sql.Index, error) // GMS requires sql.Index list

	// Partitions returns partitions of the table. This is crucial for GMS execution.
	// Partitions 返回表的Partition。这对于 GMS 执行至关重要。
	// For a simple KV store like Badger, a single partition might represent the whole table.
	// 对于像 Badger 这样的简单 KV 存储，一个 Partition 可能代表整个表。
	Partitions(ctx context.Context) (sql.PartitionIter, error)

	// PartitionRows returns a RowIter for a specific partition. This is crucial for GMS execution.
	// PartitionRows 返回特定 Partition 的 RowIter。这对于 GMS 执行至关重要。
	// The RowIter needs to read data from the storage engine.
	// RowIter 需要从存储引擎读取数据。
	PartitionRows(ctx context.Context, partition sql.Partition) (sql.RowIter, error)
}

// Transaction is the interface for database transactions.
// Transaction 是数据库事务的接口。
// It should provide methods for commit and rollback.
// 它应该提供提交和回滚的方法。
type Transaction interface {
	// Commit commits the transaction.
	// Commit 提交事务。
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction.
	// Rollback 回滚事务。
	Rollback(ctx context.Context) error

	// UnderlyingTx returns the underlying storage engine transaction object.
	// This might be needed by the storage engine implementation details.
	// UnderlyingTx 返回底层存储引擎的事务对象。
	// 存储引擎实现细节可能需要此对象。
	UnderlyingTx() interface{}
}

// RowIterator is the interface for iterating over rows.
// It extends the GMS sql.RowIter interface.
// RowIterator 是遍历行的接口。
// 它扩展了 GMS 的 sql.RowIter 接口。
type RowIterator interface {
	sql.RowIter // Inherit methods like Next() and Close()
}