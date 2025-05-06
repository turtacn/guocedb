// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"context"
	"fmt"
	"bytes"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/interfaces" // Import the interfaces package
)

// BadgerDatabase is an implementation of the interfaces.Database interface
// for the Badger storage engine.
// BadgerDatabase 是 interfaces.Database 接口的实现，用于 Badger 存储引擎。
type BadgerDatabase struct {
	// name is the name of the database.
	// name 是数据库的名称。
	name string

	// engine is a reference to the owning Badger engine.
	// engine 是对所属 Badger 引擎的引用。
	engine *BadgerEngine // Use the actual BadgerEngine type defined in badger.go
}

// NewBadgerDatabase creates a new BadgerDatabase instance.
// NewBadgerDatabase 创建一个新的 BadgerDatabase 实例。
func NewBadgerDatabase(name string, engine *BadgerEngine) *BadgerDatabase {
	return &BadgerDatabase{
		name:   name,
		engine: engine,
	}
}

// Name returns the name of the database.
// Name 返回数据库的名称。
func (d *BadgerDatabase) Name() string {
	return d.name
}

// GetTable returns a table by name within this database.
// It retrieves table schema from the catalog in Badger.
// GetTable 根据名称在此数据库中返回一个表。
// 它从 Badger 中的 catalog 检索表模式。
func (d *BadgerDatabase) GetTable(ctx context.Context, tableName string) (interfaces.Table, error) {
	log.Debug("BadgerDatabase GetTable called for %s.%s", d.name, tableName) // 调用 BadgerDatabase GetTable 获取表。

	catalogKey := EncodeCatalogKey(d.name, tableName, MetadataTypeTable)

	// Use a read-only transaction to get the schema
	// 使用只读事务获取模式
	txn := d.engine.db.NewTransaction(false) // Read-only
	defer txn.Discard()

	item, err := txn.Get(catalogKey)
	if err == badger.ErrKeyNotFound {
		log.Debug("Table '%s' not found in database '%s'", tableName, d.name) // 表 '%s' 在数据库 '%s' 中未找到。
		return nil, errors.ErrTableNotFound.New(tableName)
	} else if err != nil {
		log.Error("Failed to get table schema from Badger for %s.%s: %v", d.name, tableName, err) // 从 Badger 获取表模式失败。
		return nil, fmt.Errorf("%w: %v", errors.ErrBadgerOperationFailed, err)
	}

	// Retrieve the schema bytes
	// 检索模式字节
	schemaBytes, err := item.ValueCopy(nil)
	if err != nil {
		log.Error("Failed to get schema value from Badger item for %s.%s: %v", d.name, tableName, err) // 从 Badger item 获取模式值失败。
		return nil, fmt.Errorf("%w: %v", errors.ErrBadgerOperationFailed, err)
	}

	// TODO: Decode schemaBytes back into sql.Schema.
	// This requires a proper schema encoding/decoding mechanism.
	// For now, use a placeholder or a dummy schema.
	//
	// TODO: 将 schemaBytes 解码回 sql.Schema。
	// 这需要一个适当的模式编码/解码机制。
	// 目前使用占位符或虚拟模式。
	log.Warn("Decoding table schema from bytes is a placeholder.") // 从字节解码表模式是占位符。
	decodedSchema := sql.Schema{} // Placeholder schema

	// Return a new BadgerTable instance (which needs to be implemented)
	// 返回一个新的 BadgerTable 实例（需要实现）
	// TODO: Implement BadgerTable
	log.Warn("Returning placeholder BadgerTable.") // 返回占位符 BadgerTable。
	return &BadgerTablePlaceholder{
		schemaName: d.name,
		tableName: tableName,
		tableSchema: decodedSchema, // Use the decoded schema
		engine: d.engine,
	}, nil // Return placeholder
}

// ListTables returns a list of all table names in this database.
// It scans the catalog keys for tables in this database.
// ListTables 返回此数据库中所有表名的列表。
// 它扫描此数据库中表的 catalog key。
func (d *BadgerDatabase) ListTables(ctx context.Context) ([]string, error) {
	log.Debug("BadgerDatabase ListTables called for %s", d.name) // 调用 BadgerDatabase ListTables 获取表列表。

	var tableNames []string
	// The prefix for catalog keys for this database is catalog:<db_name>:
	// 此数据库的 catalog key 前缀是 catalog:<db_name>:
	prefix := bytes.Join([][]byte{
		NamespaceCatalogBytes,
		[]byte(d.name),
	}, []byte{NsSep, Sep})
	// Append another separator to match keys like catalog:<db_name>:<table_name>:<metadata_type>
	// 追加另一个分隔符以匹配像 catalog:<db_name>:<table_name>:<metadata_type> 这样的 key
	prefix = append(prefix, Sep)

	txn := d.engine.db.NewTransaction(false) // Read-only
	defer txn.Discard()

	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix // Set the prefix for the scan
	iterator := txn.NewIterator(opts)
	defer iterator.Close()

	for iterator.Rewind(); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		key := item.Key()

		// Decode the catalog key to extract table name and metadata type
		// 解码 catalog key 提取表名和元数据类型
		_, tableName, metadataType, ok := DecodeCatalogKey(key)
		if !ok {
			log.Warn("Skipping invalid catalog key during ListTables scan: %v", key) // 在 ListTables 扫描期间跳过无效的 catalog key。
			continue // Skip invalid keys
		}

		// Only list keys that are table schema definitions
		// 只列出作为表模式定义的 key
		if metadataType == MetadataTypeTable {
			tableNames = append(tableNames, tableName)
		}
	}

	log.Debug("Found %d tables in database '%s': %v", len(tableNames), d.name, tableNames) // 在数据库 '%s' 中找到 %d 个表。
	return tableNames, nil
}

// CreateTable creates a new table in this database.
// It stores the table schema in the catalog in Badger.
// CreateTable 在此数据库中创建一个新表。
// 它将表模式存储在 Badger 的 catalog 中。
func (d *BadgerDatabase) CreateTable(ctx context.Context, tableName string, schema sql.Schema) error {
	log.Info("BadgerDatabase CreateTable called for %s.%s", d.name, tableName) // 调用 BadgerDatabase CreateTable 创建表。

	// First, check if the table already exists
	// 首先，检查表是否已存在
	_, err := d.GetTable(ctx, tableName)
	if err == nil {
		log.Warn("Table '%s' already exists in database '%s'", tableName, d.name) // 表 '%s' 在数据库 '%s' 中已存在。
		return errors.ErrTableAlreadyExists.New(tableName)
	}
	if !errors.Is(err, errors.ErrTableNotFound) {
		// An error occurred that is not "not found"
		// 发生了一个不是“未找到”的错误
		return fmt.Errorf("failed to check for existing table %s.%s: %w", d.name, tableName, err) // 检查现有表失败。
	}

	// TODO: Encode the schema into bytes for storage.
	// This requires a proper schema encoding mechanism.
	// For now, use a placeholder or a dummy byte slice.
	//
	// TODO: 将模式编码为字节用于存储。
	// 这需要一个适当的模式编码机制。
	// 目前使用占位符或虚拟字节切片。
	log.Warn("Encoding table schema to bytes is a placeholder.") // 将表模式编码为字节是占位符。
	schemaBytes := []byte(fmt.Sprintf("dummy_schema_for_%s", tableName)) // Placeholder encoding

	catalogKey := EncodeCatalogKey(d.name, tableName, MetadataTypeTable)

	// Use a read-write transaction to write the schema
	// 使用读写事务写入模式
	txn := d.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	err = txn.Set(catalogKey, schemaBytes)
	if err != nil {
		log.Error("Failed to set table schema in Badger for %s.%s: %v", d.name, tableName, err) // 在 Badger 中设置表模式失败。
		return fmt.Errorf("%w: failed to set table schema: %v", errors.ErrBadgerOperationFailed, err)
	}

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for creating table %s.%s: %v", d.name, tableName, err) // 提交创建表事务失败。
		return fmt.Errorf("%w: failed to commit table creation: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Info("Table '%s' created successfully in database '%s'.", tableName, d.name) // 表 '%s' 在数据库 '%s' 中创建成功。
	return nil
}

// DropTable drops an existing table in this database.
// It deletes the table's schema and all its data and index entries from Badger.
// DropTable 删除此数据库中的现有表。
// 它从 Badger 中删除表的模式以及所有数据和索引条目。
func (d *BadgerDatabase) DropTable(ctx context.Context, tableName string) error {
	log.Info("BadgerDatabase DropTable called for %s.%s", d.name, tableName) // 调用 BadgerDatabase DropTable 删除表。

	// First, check if the table exists
	// 首先，检查表是否存在
	_, err := d.GetTable(ctx, tableName)
	if errors.Is(err, errors.ErrTableNotFound) {
		log.Warn("Table '%s' not found in database '%s' for dropping.", tableName, d.name) // 删除表时，表 '%s' 在数据库 '%s' 中未找到。
		return errors.ErrTableNotFound.New(tableName)
	} else if err != nil {
		return fmt.Errorf("failed to check for existing table %s.%s: %w", d.name, tableName, err) // 检查现有表失败。
	}

	// Use a read-write transaction for all deletions
	// 使用读写事务进行所有删除操作
	txn := d.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	// 1. Delete catalog key
	// 1. 删除 catalog key
	catalogKey := EncodeCatalogKey(d.name, tableName, MetadataTypeTable)
	if err := txn.Delete(catalogKey); err != nil {
		log.Error("Failed to delete catalog key for %s.%s: %v", d.name, tableName, err) // 删除 catalog key 失败。
		return fmt.Errorf("%w: failed to delete table catalog entry: %v", errors.ErrBadgerOperationFailed, err)
	}
	log.Debug("Deleted catalog key for %s.%s", d.name, tableName) // 删除表的 catalog key。

	// 2. Delete all data keys for this table
	// 2. 删除此表的所有数据 key
	dataPrefix := buildDataKeyPrefix(d.name, tableName)
	log.Debug("Deleting data with prefix: %v", dataPrefix) // 删除带有前缀的数据。
	// Badger offers DeleteRange or iterating and deleting. Iterating is safer if range deletion is tricky.
	// Badger 提供 DeleteRange 或迭代删除。如果范围删除棘手，迭代更安全。
	// TODO: Use Badger's DeleteAll or iterate and delete.
	// TODO: 使用 Badger 的 DeleteAll 或迭代删除。
	log.Warn("Deleting all data and index keys for table %s.%s is a placeholder (requires iteration/range deletion).", d.name, tableName) // 删除表的所有数据和索引 key 是占位符。

	// Example of iterating and deleting (less efficient for large tables, but safer than risky range delete)
	// 迭代删除示例（对于大表效率较低，但比冒险的范围删除更安全）
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false // Don't need values, just keys
	it := txn.NewIterator(opts)
	defer it.Close()

	// Delete data keys
	// 删除数据 key
	log.Debug("Starting data key deletion iteration for %s.%s", d.name, tableName) // 开始数据 key 删除迭代。
	for it.Seek(dataPrefix); it.ValidForPrefix(dataPrefix); it.Next() {
		keyToDelete := it.Item().KeyCopy(nil)
		if err := txn.Delete(keyToDelete); err != nil {
			log.Error("Failed to delete data key %v for %s.%s: %v", keyToDelete, d.name, tableName, err) // 删除数据 key 失败。
			// Decide if this is a fatal error or if we can continue?
			// 决定这是致命错误还是可以继续？
		}
		log.Debug("Deleted data key: %v", keyToDelete) // 删除数据 key。
	}
	log.Debug("Finished data key deletion iteration for %s.%s", d.name, tableName) // 完成数据 key 删除迭代。


	// 3. Delete all index keys for this table
	// 3. 删除此表的所有索引 key
	indexPrefix := buildIndexKeyPrefix(d.name, tableName) // Need to implement buildIndexKeyPrefix
	log.Debug("Deleting index data with prefix: %v", indexPrefix) // 删除带有前缀的索引数据。
	// Example of iterating and deleting index keys
	// 迭代删除索引 key 示例
	opts = badger.DefaultIteratorOptions
	opts.PrefetchValues = false // Don't need values, just keys
	it = txn.NewIterator(opts) // Create a new iterator for index keys
	defer it.Close()

	log.Debug("Starting index key deletion iteration for %s.%s", d.name, tableName) // 开始索引 key 删除迭代。
	for it.Seek(indexPrefix); it.ValidForPrefix(indexPrefix); it.Next() {
		keyToDelete := it.Item().KeyCopy(nil)
		if err := txn.Delete(keyToDelete); err != nil {
			log.Error("Failed to delete index key %v for %s.%s: %v", keyToDelete, d.name, tableName, err) // 删除索引 key 失败。
			// Decide if this is a fatal error
			// 决定这是否是致命错误
		}
		log.Debug("Deleted index key: %v", keyToDelete) // 删除索引 key。
	}
	log.Debug("Finished index key deletion iteration for %s.%s", d.name, tableName) // 完成索引 key 删除迭代。

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for dropping table %s.%s: %v", d.name, tableName, err) // 提交删除表事务失败。
		return fmt.Errorf("%w: failed to commit table drop: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Info("Table '%s' dropped successfully from database '%s'.", tableName, d.name) // 表 '%s' 已成功从数据库 '%s' 中删除。
	return nil
}

// buildIndexKeyPrefix constructs the key prefix for a table's index data.
// TODO: Move this to badger/encoding.go
// buildIndexKeyPrefix 构造表索引数据的 key 前缀。
// TODO: 将此移动到 badger/encoding.go
func buildIndexKeyPrefix(dbName, tableName string) []byte {
	prefix := bytes.Join([][]byte{
		NamespaceIndexBytes,
		[]byte(dbName),
		[]byte(tableName),
	}, []byte{NsSep, Sep, Sep})
	// Append separator to ensure it's a prefix matching only this table's index keys
	// 追加分隔符以确保它是仅匹配此表索引 key 的前缀
	return append(prefix, Sep)
}


// RenameTable renames a table in this database.
// This is a complex operation in a KV store and is not implemented yet.
// RenameTable 在此数据库中重命名表。
// 这在 KV 存储中是一个复杂的操作，尚未实现。
func (d *BadgerDatabase) RenameTable(ctx context.Context, oldTableName, newTableName string) error {
	log.Warn("BadgerDatabase RenameTable called for %s.%s -> %s (Not Implemented)", d.name, oldTableName, newTableName) // 调用 BadgerDatabase RenameTable（未实现）。
	return errors.ErrNotImplemented.New("rename table for Badger engine")
}