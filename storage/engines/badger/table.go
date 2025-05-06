// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"context"
	"fmt"
	"bytes"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/interfaces" // Import the interfaces package
)

// BadgerTable is an implementation of the interfaces.Table interface
// for the Badger storage engine.
// BadgerTable 是 interfaces.Table 接口的实现，用于 Badger 存储引擎。
type BadgerTable struct {
	// dbName is the name of the database this table belongs to.
	// dbName 是此表所属数据库的名称。
	dbName string

	// tableName is the name of the table.
	// tableName 是表的名称。
	tableName string

	// tableSchema is the schema of the table.
	// tableSchema 是表的模式。
	tableSchema sql.Schema

	// engine is a reference to the owning Badger engine.
	// engine 是对所属 Badger 引擎的引用。
	engine *BadgerEngine // Use the actual BadgerEngine type defined in badger.go

	// indexes stores the list of indexes for this table.
	// indexes 存储此表的索引列表。
	// TODO: This should ideally be loaded from the catalog.
	// TODO: 这理想情况下应该从 catalog 加载。
	indexes []sql.Index
}

// NewBadgerTable creates a new BadgerTable instance.
// NewBadgerTable 创建一个新的 BadgerTable 实例。
// Note: Schema and indexes should be loaded from the catalog when the table is retrieved.
// 注意：检索表时应从 catalog 加载模式和索引。
func NewBadgerTable(dbName, tableName string, schema sql.Schema, engine *BadgerEngine) *BadgerTable {
	// TODO: Load indexes from catalog here or in GetTable.
	// TODO: 在此处或 GetTable 中从 catalog 加载索引。
	return &BadgerTable{
		dbName: dbName,
		tableName: tableName,
		tableSchema: schema,
		engine: engine,
		indexes: []sql.Index{}, // Placeholder
	}
}

// Name returns the name of the table.
// Name 返回表的名称。
func (t *BadgerTable) Name() string {
	return t.tableName
}

// Schema returns the schema of the table.
// Schema 返回表的模式。
func (t *BadgerTable) Schema() sql.Schema {
	return t.tableSchema
}

// Collation returns the collation of the table.
// Collation 返回表的排序规则。
// Placeholder implementation.
// 占位符实现。
func (t *BadgerTable) Collation() sql.CollationID {
	return sql.Collation_Default
}

// Comment returns the comment of the table.
// Comment 返回表的注释。
// Placeholder implementation.
// 占位符实现。
func (t *BadgerTable) Comment() string {
	return ""
}

// Partitions returns partitions of the table.
// For Badger, we treat the whole table as a single partition initially.
// Partitions 返回表的 Partition。
// 对于 Badger，我们初步将整个表视为单个 Partition。
func (t *BadgerTable) Partitions(ctx context.Context) (sql.PartitionIter, error) {
	log.Debug("BadgerTable Partitions called for %s.%s", t.dbName, t.tableName) // 调用 BadgerTable Partitions。
	// Return an iterator that yields a single partition representing the whole table.
	// 返回一个迭代器，该迭代器产生一个代表整个表的单个 Partition。
	return &BadgerPartitionIterator{dbName: t.dbName, tableName: t.tableName, sent: false}, nil
}

// BadgerPartitionIterator is an iterator that yields Badger partitions.
// BadgerPartitionIterator 是一个产生 Badger Partition 的迭代器。
type BadgerPartitionIterator struct {
	dbName string
	tableName string
	sent bool // Flag to indicate if the single partition has been sent
	// 标记，指示单个 Partition 是否已发送
}

// Next returns the next partition.
// Next 返回下一个 Partition。
func (i *BadgerPartitionIterator) Next(ctx context.Context) (sql.Partition, error) {
	if i.sent {
		return nil, nil // No more partitions
	}
	i.sent = true
	// Return the single partition representing the whole table.
	// 返回代表整个表的单个 Partition。
	return &BadgerPartition{dbName: i.dbName, tableName: i.tableName}, nil
}

// Close closes the iterator.
// Close 关闭迭代器。
func (i *BadgerPartitionIterator) Close(ctx context.Context) error {
	log.Debug("BadgerPartitionIterator Close called for %s.%s", i.dbName, i.tableName) // 关闭 BadgerPartitionIterator。
	return nil // Nothing to close for this simple iterator
}

// BadgerPartition is a partition for a Badger table.
// For Badger, a partition represents the entire table data range.
// BadgerPartition 是 Badger 表的 Partition。
// 对于 Badger，一个 Partition 代表整个表的数据范围。
type BadgerPartition struct {
	dbName string
	tableName string
}

// Key returns the key identifying this partition.
// Key 返回识别此 Partition 的 key。
func (p *BadgerPartition) Key() []byte {
	// The key should be the prefix for this table's data keys.
	// key 应该是此表数据 key 的前缀。
	return buildDataKeyPrefix(p.dbName, p.tableName)
}

// PartitionRows returns a RowIter for a specific partition.
// PartitionRows 返回特定 Partition 的 RowIter。
// This iterator scans the data keys for the given partition (which is the whole table for now).
// 此迭代器扫描给定 Partition 的数据 key（目前是整个表）。
func (t *BadgerTable) PartitionRows(ctx context.Context, partition sql.Partition) (sql.RowIter, error) {
	log.Debug("BadgerTable PartitionRows called for %s.%s with partition key: %v", t.dbName, t.tableName, partition.Key()) // 调用 BadgerTable PartitionRows。

	// Ensure the partition is a BadgerPartition
	// 确保 Partition 是 BadgerPartition
	badgerPartition, ok := partition.(*BadgerPartition)
	if !ok {
		return nil, fmt.Errorf("invalid partition type for BadgerTable: %T", partition) // BadgerTable 的 Partition 类型无效。
	}

	// Create and return a BadgerRowIterator
	// 创建并返回一个 BadgerRowIterator
	// The iterator needs the Badger DB handle from the engine.
	// 迭代器需要引擎中的 Badger DB 句柄。
	iter, err := NewBadgerRowIterator(t.engine.db, t.tableSchema, badgerPartition.dbName, badgerPartition.tableName)
	if err != nil {
		log.Error("Failed to create BadgerRowIterator for %s.%s: %v", t.dbName, t.tableName, err) // 创建 BadgerRowIterator 失败。
		return nil, fmt.Errorf("%w: failed to create row iterator: %v", errors.ErrBadgerOperationFailed, err)
	}

	return iter, nil
}

// Insert inserts a new row into the table.
// It writes the data and index entries to Badger within a transaction.
// Insert 向表中插入新行。
// 它在事务中将数据和索引条目写入 Badger。
func (t *BadgerTable) Insert(ctx context.Context, row sql.Row) error {
	log.Debug("BadgerTable Insert called for %s.%s with row: %v", t.dbName, t.tableName, row) // 调用 BadgerTable Insert 插入行。

	// Get the primary key from the row
	// 从行中获取主键
	pkRow, err := sql.GetPrimaryKeyValues(row, t.tableSchema)
	if err != nil {
		log.Error("Failed to get primary key from row for insert: %v", err) // 从行中获取主键失败。
		// Consider wrapping in a specific error type
		return fmt.Errorf("failed to get primary key from row: %w", err)
	}
	if len(pkRow) == 0 {
		log.Error("Primary key is required for insert but missing in row: %v", row) // Insert 需要主键，但行中缺少主键。
		return errors.ErrPrimaryKeyRequired.New(t.tableName)
	}

	// Start a read-write transaction
	// 开启读写事务
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	// 1. Encode and write the data key-value pair
	// 1. 编码并写入数据 key-value 对
	dataKey, err := EncodeDataKey(t.dbName, t.tableName, pkRow)
	if err != nil {
		log.Error("Failed to encode data key for insert: %v", err) // 编码数据 key 失败。
		return fmt.Errorf("%w: failed to encode data key: %v", errors.ErrEncodingFailed, err)
	}

	// Check for primary key conflicts (optional, but good practice)
	// 检查主键冲突（可选，但推荐）
	_, err = txn.Get(dataKey)
	if err == nil {
		log.Warn("Primary key conflict during insert for %s.%s, key: %v", t.dbName, t.tableName, dataKey) // Insert 期间主键冲突。
		return errors.ErrTableAlreadyExists.New(fmt.Sprintf("row with primary key %v", pkRow)) // Re-use ErrTableAlreadyExists for row conflict? Or define ErrRowAlreadyExists.
	} else if err != badger.ErrKeyNotFound {
		log.Error("Failed to check for primary key existence during insert for %s.%s, key %v: %v", t.dbName, t.tableName, dataKey, err) // 检查主键存在性失败。
		return fmt.Errorf("%w: failed to check primary key existence: %v", errors.ErrBadgerOperationFailed, err)
	}

	rowData, err := EncodeRow(row)
	if err != nil {
		log.Error("Failed to encode row data for insert: %v", err) // 编码行数据失败。
		return fmt.Errorf("%w: failed to encode row data: %v", errors.ErrEncodingFailed, err)
	}

	if err := txn.Set(dataKey, rowData); err != nil {
		log.Error("Failed to set data key-value in Badger for insert %v: %v", dataKey, err) // 在 Badger 中设置数据 key-value 失败。
		return fmt.Errorf("%w: failed to set data key: %v", errors.ErrBadgerOperationFailed, err)
	}
	log.Debug("Set data key for insert: %v", dataKey) // 设置用于 insert 的数据 key。

	// 2. Encode and write index key-value pairs for all indexes
	// 2. 为所有索引编码并写入索引 key-value 对
	for _, index := range t.indexes {
		// Get the values for the index columns from the row
		// 从行中获取索引列的值
		indexValues, err := expression.GetValues(row, index.Expressions()...)
		if err != nil {
			log.Error("Failed to get index values from row for index '%s' insert: %v", index.Name(), err) // 从行中获取索引值失败。
			// Decide if this is a fatal error during insert
			// 决定这是否是 insert 期间的致命错误
			txn.Discard() // Rollback the transaction
			return fmt.Errorf("failed to get index values for index '%s': %w", index.Name(), err)
		}

		// Encode the index key
		// 编码索引 key
		// Note: EncodeIndexKey needs the index values and the primary key (for uniqueness if index is not unique)
		// 注意：EncodeIndexKey 需要索引值和主键（如果索引不唯一，则需要主键保证唯一性）
		indexKey, err := EncodeIndexKey(t.dbName, t.tableName, index.Name(), indexValues, pkRow)
		if err != nil {
			log.Error("Failed to encode index key for index '%s' insert: %v", index.Name(), err) // 编码索引 key 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode index key '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}

		// The value for an index entry can be empty or just the primary key encoded bytes.
		// Storing primary key bytes allows retrieving the PK directly from index scan.
		// 索引条目的 value 可以是空的，或者只是主键编码字节。
		// 存储主键字节允许从索引扫描直接检索 PK。
		indexValue := pkRow // Store primary key as value for index entry (can be []byte or encoded)
		pkEncodedForIndexValue, err := encoding.EncodeRowForKV(indexValue) // Assuming EncodeRowForKV gives bytes
		if err != nil {
			log.Error("Failed to encode primary key for index value for index '%s': %v", index.Name(), err) // 编码主键用于索引值失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode primary key for index value '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}


		if err := txn.Set(indexKey, pkEncodedForIndexValue); err != nil { // Or just empty []byte{} if PK is in key
			log.Error("Failed to set index key-value in Badger for index '%s' insert %v: %v", index.Name(), indexKey, err) // 在 Badger 中设置索引 key-value 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to set index key '%s': %v", errors.ErrBadgerOperationFailed, index.Name(), err)
		}
		log.Debug("Set index key for insert: %v", indexKey) // 设置用于 insert 的索引 key。
	}

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for inserting row into %s.%s: %v", t.dbName, t.tableName, err) // 提交将行插入表事务失败。
		return fmt.Errorf("%w: failed to commit row insert: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Debug("Row inserted successfully into %s.%s", t.dbName, t.tableName) // 行成功插入到表。
	return nil
}

// Update updates an existing row in the table.
// It deletes the old data/index entries and writes the new ones within a transaction.
// Update 更新表中的现有行。
// 它在事务中删除旧的数据/索引条目并写入新的。
func (t *BadgerTable) Update(ctx context.Context, oldRow, newRow sql.Row) error {
	log.Debug("BadgerTable Update called for %s.%s old: %v, new: %v", t.dbName, t.tableName, oldRow, newRow) // 调用 BadgerTable Update。

	// Get primary keys for old and new rows
	// 获取旧行和新行的主键
	oldPkRow, err := sql.GetPrimaryKeyValues(oldRow, t.tableSchema)
	if err != nil {
		log.Error("Failed to get primary key from old row for update: %v", err) // 从旧行获取主键失败。
		return fmt.Errorf("failed to get primary key from old row: %w", err)
	}
	newPkRow, err := sql.GetPrimaryKeyValues(newRow, t.tableSchema)
	if err != nil {
		log.Error("Failed to get primary key from new row for update: %v", err) // 从新行获取主键失败。
		return fmt.Errorf("failed to get primary key from new row: %w", err)
	}

	// Start a read-write transaction
	// 开启读写事务
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	// 1. Delete old data key-value pair
	// 1. 删除旧的数据 key-value 对
	oldDataKey, err := EncodeDataKey(t.dbName, t.tableName, oldPkRow)
	if err != nil {
		log.Error("Failed to encode old data key for update: %v", err) // 编码旧数据 key 失败。
		return fmt.Errorf("%w: failed to encode old data key: %v", errors.ErrEncodingFailed, err)
	}
	// Check if old data key exists before deleting (optional)
	// 在删除前检查旧数据 key 是否存在（可选）
	if err := txn.Delete(oldDataKey); err != nil {
		if err == badger.ErrKeyNotFound {
			log.Warn("Old data key not found during update for %s.%s, key: %v. Row might not exist?", t.dbName, t.tableName, oldDataKey) // 更新期间旧数据 key 未找到。行可能不存在？
			// Depending on requirements, this might be an error or ignorable.
			// 根据需求，这可能是一个错误或可忽略。
		} else {
			log.Error("Failed to delete old data key in Badger for update %v: %v", oldDataKey, err) // 在 Badger 中删除旧数据 key 失败。
			return fmt.Errorf("%w: failed to delete old data key: %v", errors.ErrBadgerOperationFailed, err)
		}
	}
	log.Debug("Deleted old data key for update: %v", oldDataKey) // 删除用于 update 的旧数据 key。


	// 2. Encode and write the new data key-value pair
	// 2. 编码并写入新的数据 key-value 对
	newDataKey, err := EncodeDataKey(t.dbName, t.tableName, newPkRow)
	if err != nil {
		log.Error("Failed to encode new data key for update: %v", err) // 编码新数据 key 失败。
		return fmt.Errorf("%w: failed to encode new data key: %v", errors.ErrEncodingFailed, err)
	}
	newRowData, err := EncodeRow(newRow)
	if err != nil {
		log.Error("Failed to encode new row data for update: %v", err) // 编码新行数据失败。
		return fmt.Errorf("%w: failed to encode new row data: %v", errors.ErrEncodingFailed, err)
	}
	if err := txn.Set(newDataKey, newRowData); err != nil {
		log.Error("Failed to set new data key-value in Badger for update %v: %v", newDataKey, err) // 在 Badger 中设置新数据 key-value 失败。
		return fmt.Errorf("%w: failed to set new data key: %v", errors.ErrBadgerOperationFailed, err)
	}
	log.Debug("Set new data key for update: %v", newDataKey) // 设置用于 update 的新数据 key。


	// 3. Delete old index entries and write new index entries for all indexes
	// 3. 为所有索引删除旧的索引条目并写入新的索引条目
	for _, index := range t.indexes {
		// Get old and new index values
		// 获取旧的和新的索引值
		oldIndexValues, err := expression.GetValues(oldRow, index.Expressions()...)
		if err != nil {
			log.Error("Failed to get old index values from row for index '%s' update: %v", index.Name(), err) // 从行中获取旧索引值失败。
			txn.Discard()
			return fmt.Errorf("failed to get old index values for index '%s': %w", index.Name(), err)
		}
		newIndexValues, err := expression.GetValues(newRow, index.Expressions()...)
		if err != nil {
			log.Error("Failed to get new index values from row for index '%s' update: %v", index.Name(), err) // 从行中获取新索引值失败。
			txn.Discard()
			return fmt.Errorf("failed to get new index values for index '%s': %w", index.Name(), err)
		}

		// Encode old and new index keys
		// 编码旧的和新的索引 key
		oldIndexKey, err := EncodeIndexKey(t.dbName, t.tableName, index.Name(), oldIndexValues, oldPkRow)
		if err != nil {
			log.Error("Failed to encode old index key for index '%s' update: %v", index.Name(), err) // 编码旧索引 key 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode old index key '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}
		newIndexKey, err := EncodeIndexKey(t.dbName, t.tableName, index.Name(), newIndexValues, newPkRow)
		if err != nil {
			log.Error("Failed to encode new index key for index '%s' update: %v", index.Name(), err) // 编码新索引 key 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode new index key '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}

		// Delete the old index key
		// 删除旧的索引 key
		if err := txn.Delete(oldIndexKey); err != nil {
			if err == badger.ErrKeyNotFound {
				log.Warn("Old index key not found during update for index '%s', key: %v. Index entry might not exist?", index.Name(), oldIndexKey) // 更新期间旧索引 key 未找到。索引条目可能不存在？
			} else {
				log.Error("Failed to delete old index key in Badger for index '%s' update %v: %v", index.Name(), oldIndexKey, err) // 在 Badger 中删除旧索引 key 失败。
				txn.Discard()
				return fmt.Errorf("%w: failed to delete old index key '%s': %v", errors.ErrBadgerOperationFailed, index.Name(), err)
			}
		}
		log.Debug("Deleted old index key for update: %v", oldIndexKey) // 删除用于 update 的旧索引 key。


		// Write the new index key-value pair
		// 写入新的索引 key-value 对
		newIndexValue := newPkRow // Store primary key as value
		pkEncodedForNewIndexValue, err := encoding.EncodeRowForKV(newIndexValue)
		if err != nil {
			log.Error("Failed to encode primary key for new index value for index '%s': %v", index.Name(), err) // 编码主键用于新索引值失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode primary key for new index value '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}
		if err := txn.Set(newIndexKey, pkEncodedForNewIndexValue); err != nil {
			log.Error("Failed to set new index key-value in Badger for index '%s' update %v: %v", index.Name(), newIndexKey, err) // 在 Badger 中设置新索引 key-value 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to set new index key '%s': %v", errors.ErrBadgerOperationFailed, index.Name(), err)
		}
		log.Debug("Set new index key for update: %v", newIndexKey) // 设置用于 update 的新索引 key。
	}


	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for updating row in %s.%s: %v", t.dbName, t.tableName, err) // 提交更新行事务失败。
		return fmt.Errorf("%w: failed to commit row update: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Debug("Row updated successfully in %s.%s", t.dbName, t.tableName) // 行在表成功更新。
	return nil
}

// Delete deletes a row from the table.
// It deletes the data and index entries from Badger within a transaction.
// Delete 从表中删除行。
// 它在事务中从 Badger 删除数据和索引条目。
func (t *BadgerTable) Delete(ctx context.Context, row sql.Row) error {
	log.Debug("BadgerTable Delete called for %s.%s with row: %v", t.dbName, t.tableName, row) // 调用 BadgerTable Delete 删除行。

	// Get the primary key from the row
	// 从行中获取主键
	pkRow, err := sql.GetPrimaryKeyValues(row, t.tableSchema)
	if err != nil {
		log.Error("Failed to get primary key from row for delete: %v", err) // 从行中获取主键失败。
		return fmt.Errorf("failed to get primary key from row: %w", err)
	}
	if len(pkRow) == 0 {
		log.Error("Primary key is required for delete but missing in row: %v", row) // Delete 需要主键，但行中缺少主键。
		return errors.ErrPrimaryKeyRequired.New(t.tableName)
	}

	// Start a read-write transaction
	// 开启读写事务
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	// 1. Delete the data key-value pair
	// 1. 删除数据 key-value 对
	dataKey, err := EncodeDataKey(t.dbName, t.tableName, pkRow)
	if err != nil {
		log.Error("Failed to encode data key for delete: %v", err) // 编码数据 key 失败。
		return fmt.Errorf("%w: failed to encode data key: %v", errors.ErrEncodingFailed, err)
	}
	if err := txn.Delete(dataKey); err != nil {
		if err == badger.ErrKeyNotFound {
			log.Warn("Data key not found during delete for %s.%s, key: %v. Row might not exist?", t.dbName, t.tableName, dataKey) // 删除期间数据 key 未找到。行可能不存在？
			// Depending on requirements, this might be an error or ignorable.
		} else {
			log.Error("Failed to delete data key in Badger for delete %v: %v", dataKey, err) // 在 Badger 中删除数据 key 失败。
			return fmt.Errorf("%w: failed to delete data key: %v", errors.ErrBadgerOperationFailed, err)
		}
	}
	log.Debug("Deleted data key for delete: %v", dataKey) // 删除用于 delete 的数据 key。


	// 2. Delete index entries for all indexes
	// 2. 为所有索引删除索引条目
	for _, index := range t.indexes {
		// Get the values for the index columns from the row
		// 从行中获取索引列的值
		indexValues, err := expression.GetValues(row, index.Expressions()...)
		if err != nil {
			log.Error("Failed to get index values from row for index '%s' delete: %v", index.Name(), err) // 从行中获取索引值失败。
			txn.Discard()
			return fmt.Errorf("failed to get index values for index '%s': %w", index.Name(), err)
		}

		// Encode the index key
		// 编码索引 key
		indexKey, err := EncodeIndexKey(t.dbName, t.tableName, index.Name(), indexValues, pkRow)
		if err != nil {
			log.Error("Failed to encode index key for index '%s' delete: %v", index.Name(), err) // 编码索引 key 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode index key '%s': %v", errors.ErrEncodingFailed, index.Name(), err)
		}

		// Delete the index key
		// 删除索引 key
		if err := txn.Delete(indexKey); err != nil {
			if err == badger.ErrKeyNotFound {
				log.Warn("Index key not found during delete for index '%s', key: %v. Index entry might not exist?", index.Name(), indexKey) // 删除期间索引 key 未找到。索引条目可能不存在？
			} else {
				log.Error("Failed to delete index key in Badger for index '%s' delete %v: %v", index.Name(), indexKey, err) // 在 Badger 中删除索引 key 失败。
				txn.Discard()
				return fmt.Errorf("%w: failed to delete index key '%s': %v", errors.ErrBadgerOperationFailed, index.Name(), err)
			}
		}
		log.Debug("Deleted index key for delete: %v", indexKey) // 删除用于 delete 的索引 key。
	}

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for deleting row from %s.%s: %v", t.dbName, t.tableName, err) // 提交从表删除行事务失败。
		return fmt.Errorf("%w: failed to commit row delete: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Debug("Row deleted successfully from %s.%s", t.dbName, t.tableName) // 行成功从表删除。
	return nil
}

// Truncate removes all rows from the table.
// It deletes all data and index entries for this table from Badger.
// Truncate 删除表中的所有行。
// 它从 Badger 中删除此表的所有数据和索引条目。
func (t *BadgerTable) Truncate(ctx context.Context) error {
	log.Info("BadgerTable Truncate called for %s.%s", t.dbName, t.tableName) // 调用 BadgerTable Truncate。

	// Use a read-write transaction for deletions
	// 使用读写事务进行删除操作
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard() // Discard on failure

	// Delete all data keys for this table
	// 删除此表的所有数据 key
	dataPrefix := buildDataKeyPrefix(t.dbName, t.tableName)
	log.Debug("Deleting data with prefix for truncate: %v", dataPrefix) // Truncate 删除数据前缀。
	// Use iterators and Delete as range deletion might not be atomic or efficient depending on Badger version/use.
	// 使用迭代器和 Delete 进行删除，因为范围删除可能不是原子或高效的，取决于 Badger 版本/使用方式。
	// TODO: Implement efficient range deletion or iteration.
	// TODO: 实现高效的范围删除或迭代。
	log.Warn("Truncating data for table %s.%s is a placeholder (requires iteration/range deletion).", t.dbName, t.tableName) // Truncate 表数据是占位符。

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false // Don't need values, just keys
	it := txn.NewIterator(opts)
	defer it.Close()

	// Delete data keys
	// 删除数据 key
	log.Debug("Starting data key deletion iteration for truncate %s.%s", t.dbName, t.tableName) // 开始 Truncate 数据 key 删除迭代。
	for it.Seek(dataPrefix); it.ValidForPrefix(dataPrefix); it.Next() {
		keyToDelete := it.Item().KeyCopy(nil)
		if err := txn.Delete(keyToDelete); err != nil {
			log.Error("Failed to delete data key %v during truncate for %s.%s: %v", keyToDelete, t.dbName, t.tableName, err) // Truncate 期间删除数据 key 失败。
		}
		log.Debug("Deleted data key during truncate: %v", keyToDelete) // Truncate 期间删除数据 key。
	}
	log.Debug("Finished data key deletion iteration for truncate %s.%s", t.dbName, t.tableName) // 完成 Truncate 数据 key 删除迭代。

	// Delete all index keys for this table
	// 删除此表的所有索引 key
	indexPrefix := buildIndexKeyPrefix(t.dbName, t.tableName)
	log.Debug("Deleting index data with prefix for truncate: %v", indexPrefix) // Truncate 删除索引数据前缀。
	// Use a new iterator for index keys
	// 为索引 key 使用新的迭代器
	it = txn.NewIterator(opts)
	defer it.Close()

	log.Debug("Starting index key deletion iteration for truncate %s.%s", t.dbName, t.tableName) // 开始 Truncate 索引 key 删除迭代。
	for it.Seek(indexPrefix); it.ValidForPrefix(indexPrefix); it.Next() {
		keyToDelete := it.Item().KeyCopy(nil)
		if err := txn.Delete(keyToDelete); err != nil {
			log.Error("Failed to delete index key %v during truncate for %s.%s: %v", keyToDelete, t.dbName, t.tableName, err) // Truncate 期间删除索引 key 失败。
		}
		log.Debug("Deleted index key during truncate: %v", keyToDelete) // Truncate 期间删除索引 key。
	}
	log.Debug("Finished index key deletion iteration for truncate %s.%s", t.dbName, t.tableName) // 完成 Truncate 索引 key 删除迭代。


	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for truncating table %s.%s: %v", t.dbName, t.tableName, err) // 提交 Truncate 表事务失败。
		return fmt.Errorf("%w: failed to commit table truncate: %v", errors.ErrTransactionCommitFailed, err)
	}

	log.Info("Table '%s' truncated successfully in database '%s'.", t.dbName, t.tableName) // 表 '%s' 在数据库 '%s' 中成功清空。
	return nil
}


// CreateIndex creates a new index on the table.
// It stores the index definition and builds the index data in Badger.
// CreateIndex 在表上创建一个新索引。
// 它在 Badger 中存储索引定义并构建索引数据。
func (t *BadgerTable) CreateIndex(ctx context.Context, indexDef sql.IndexDef) error {
	log.Info("BadgerTable CreateIndex called for %s.%s index: %s", t.dbName, t.tableName, indexDef.Name) // 调用 BadgerTable CreateIndex。

	// TODO: Check if index already exists in catalog.
	// TODO: Store index definition metadata in catalog.

	// Build the index by scanning all existing data and writing index entries.
	// 通过扫描所有现有数据并写入索引条目来构建索引。
	log.Info("Building index '%s' for table %s.%s", indexDef.Name, t.dbName, t.tableName) // 为表构建索引。

	// Use a transaction for the entire index build process.
	// 使用事务进行整个索引构建过程。
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard()

	// Get a RowIterator for the table data.
	// 获取表数据的 RowIterator。
	// We need to iterate through all existing rows.
	// 我们需要遍历所有现有行。
	partitionIter, err := t.Partitions(ctx)
	if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to get partitions for index build: %w", err) // 获取 Partition 失败。
	}
	defer partitionIter.Close(ctx)

	// Assume single partition for now
	// 假设只有一个 Partition
	partition, err := partitionIter.Next(ctx)
	if err != nil || partition == nil {
		txn.Discard()
		return fmt.Errorf("failed to get partition for index build: %w", err) // 获取 Partition 失败。
	}

	rowIter, err := t.PartitionRows(ctx, partition)
	if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to get row iterator for index build: %w", err) // 获取行迭代器失败。
	}
	defer rowIter.Close(ctx)

	// Iterate through rows and create index entries
	// 遍历行并创建索引条目
	count := 0
	for {
		row, err := rowIter.Next(ctx)
		if err != nil {
			log.Error("Error iterating rows during index build for index '%s' on %s.%s: %v", indexDef.Name, t.dbName, t.tableName, err) // 在索引构建期间迭代行出错。
			txn.Discard()
			return fmt.Errorf("error iterating rows during index build: %w", err)
		}
		if row == nil {
			break // End of iteration
		}
		count++

		// Get primary key and index values for the row
		// 获取行的主键和索引值
		pkRow, err := sql.GetPrimaryKeyValues(row, t.tableSchema)
		if err != nil {
			log.Error("Failed to get primary key for row during index build for index '%s': %v", indexDef.Name, err) // 在索引构建期间获取主键失败。
			txn.Discard()
			return fmt.Errorf("failed to get primary key for row during index build: %w", err)
		}
		indexValues, err := expression.GetValues(row, indexDef.Expressions...) // Note: using indexDef.Expressions
		if err != nil {
			log.Error("Failed to get index values for row during index build for index '%s': %v", indexDef.Name, err) // 在索引构建期间获取索引值失败。
			txn.Discard()
			return fmt.Errorf("failed to get index values for row during index build: %w", err)
		}

		// Encode the index key
		// 编码索引 key
		indexKey, err := EncodeIndexKey(t.dbName, t.tableName, indexDef.Name, indexValues, pkRow)
		if err != nil {
			log.Error("Failed to encode index key for row during index build for index '%s': %v", indexDef.Name, err) // 在索引构建期间编码索引 key 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode index key for index build '%s': %v", errors.ErrEncodingFailed, indexDef.Name, err)
		}

		// The value for an index entry is typically the primary key encoded bytes.
		// 索引条目的 value 通常是主键编码字节。
		pkEncodedForIndexValue, err := encoding.EncodeRowForKV(pkRow) // Assuming EncodeRowForKV gives bytes
		if err != nil {
			log.Error("Failed to encode primary key for index value during index build for index '%s': %v", indexDef.Name, err) // 编码主键用于索引值失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to encode primary key for index value during index build '%s': %v", errors.ErrEncodingFailed, indexDef.Name, err)
		}

		// Write the index key-value pair
		// 写入索引 key-value 对
		if err := txn.Set(indexKey, pkEncodedForIndexValue); err != nil {
			log.Error("Failed to set index key-value in Badger during index build for index '%s' %v: %v", indexDef.Name, indexKey, err) // 在 Badger 中设置索引 key-value 失败。
			txn.Discard()
			return fmt.Errorf("%w: failed to set index key during index build '%s': %v", errors.ErrBadgerOperationFailed, indexDef.Name, err)
		}
		// Batch commits for large index builds might be necessary for performance
		// 对于大型索引构建，可能需要批量提交以提高性能
	}

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for building index '%s' on %s.%s: %v", indexDef.Name, t.dbName, t.tableName, err) // 提交构建索引事务失败。
		return fmt.Errorf("%w: failed to commit index build: %v", errors.ErrTransactionCommitFailed, err)
	}

	// TODO: Add the new index to the table's index list (in memory and persistent catalog).
	// TODO: 将新索引添加到表的索引列表（内存和持久化 catalog）。

	log.Info("Index '%s' built successfully for table %s.%s (%d rows processed).", indexDef.Name, t.dbName, t.tableName, count) // 索引构建成功。
	return nil
}


// DropIndex drops an existing index from the table.
// It deletes the index definition and all index data from Badger.
// DropIndex 从表中删除现有索引。
// 它从 Badger 中删除索引定义和所有索引数据。
func (t *BadgerTable) DropIndex(ctx context.Context, indexName string) error {
	log.Info("BadgerTable DropIndex called for %s.%s index: %s", t.dbName, t.tableName, indexName) // 调用 BadgerTable DropIndex。

	// TODO: Check if index exists in catalog.
	// TODO: Delete index definition metadata from catalog.

	// Delete all index keys belonging to this index.
	// 删除属于此索引的所有索引 key。
	indexPrefix := buildIndexKeyPrefixWithIndexName(t.dbName, t.tableName, indexName) // Need to implement buildIndexKeyPrefixWithIndexName
	log.Debug("Deleting index data with prefix for drop index: %v", indexPrefix) // Drop index 删除索引数据前缀。

	// Use a read-write transaction for deletions
	// 使用读写事务进行删除操作
	txn := t.engine.db.NewTransaction(true) // Read-write
	defer txn.Discard()

	// Use iterator and Delete to remove keys under the index prefix.
	// 使用迭代器和 Delete 删除索引前缀下的 key。
	// TODO: Implement efficient range deletion or iteration.
	// TODO: 实现高效的范围删除或迭代。
	log.Warn("Deleting index data for index %s on table %s.%s is a placeholder (requires iteration/range deletion).", indexName, t.dbName, t.tableName) // 删除索引数据是占位符。

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false // Don't need values, just keys
	it := txn.NewIterator(opts)
	defer it.Close()

	log.Debug("Starting index key deletion iteration for drop index '%s' on %s.%s", indexName, t.dbName, t.tableName) // 开始 Drop index 索引 key 删除迭代。
	for it.Seek(indexPrefix); it.ValidForPrefix(indexPrefix); it.Next() {
		keyToDelete := it.Item().KeyCopy(nil)
		if err := txn.Delete(keyToDelete); err != nil {
			log.Error("Failed to delete index key %v during drop index '%s': %v", keyToDelete, indexName, err) // Drop index 期间删除索引 key 失败。
		}
		log.Debug("Deleted index key during drop index: %v", keyToDelete) // Drop index 期间删除索引 key。
	}
	log.Debug("Finished index key deletion iteration for drop index '%s' on %s.%s", indexName, t.dbName, t.tableName) // 完成 Drop index 索引 key 删除迭代。

	// Commit the transaction
	// 提交事务
	if err := txn.Commit(); err != nil {
		log.Error("Failed to commit transaction for dropping index '%s' on %s.%s: %v", indexName, t.dbName, t.tableName, err) // 提交删除索引事务失败。
		return fmt.Errorf("%w: failed to commit index drop: %v", errors.ErrTransactionCommitFailed, err)
	}

	// TODO: Remove the index from the table's index list (in memory and persistent catalog).
	// TODO: 从表的索引列表（内存和持久化 catalog）中移除索引。

	log.Info("Index '%s' dropped successfully from table %s.%s.", indexName, t.dbName, t.tableName) // 索引 '%s' 已成功从表 '%s.%s' 中删除。
	return nil
}

// buildIndexKeyPrefixWithIndexName constructs the key prefix for a specific index's data.
// TODO: Move this to badger/encoding.go
// buildIndexKeyPrefixWithIndexName 构造特定索引数据的 key 前缀。
// TODO: 将此移动到 badger/encoding.go
func buildIndexKeyPrefixWithIndexName(dbName, tableName, indexName string) []byte {
	prefix := bytes.Join([][]byte{
		NamespaceIndexBytes,
		[]byte(dbName),
		[]byte(tableName),
		[]byte(indexName),
	}, []byte{NsSep, Sep, Sep, Sep})
	// Append separator to ensure it's a prefix matching only this index's keys
	// 追加分隔符以确保它是仅匹配此索引 key 的前缀
	return append(prefix, Sep)
}


// GetIndex returns an index by name.
// It retrieves the index definition from the catalog.
// GetIndex 根据名称返回一个索引。
// 它从 catalog 检索索引定义。
func (t *BadgerTable) GetIndex(ctx context.Context, indexName string) (sql.Index, error) {
	log.Debug("BadgerTable GetIndex called for %s.%s index: %s", t.dbName, t.tableName, indexName) // 调用 BadgerTable GetIndex。

	// TODO: Retrieve index definition metadata from catalog in Badger.
	// Need catalog key like catalog:<db_name>:<table_name>:<index_name>:<metadata_type_index>
	//
	// TODO: 从 Badger 的 catalog 中检索索引定义元数据。
	// 需要 catalog key，如 catalog:<db_name>:<table_name>:<index_name>:<metadata_type_index>
	log.Warn("GetIndex implementation is a placeholder.", indexName) // GetIndex 实现是占位符。
	// If found, decode it into a sql.Index object and return.
	// 如果找到，将其解码为 sql.Index 对象并返回。

	// For now, simulate finding an index if it exists in the in-memory list (placeholder logic)
	// 目前，如果索引存在于内存列表（占位符逻辑）中，则模拟找到索引。
	for _, idx := range t.indexes {
		if idx.Name() == indexName {
			log.Debug("Found index '%s' in in-memory list.", indexName) // 在内存列表中找到索引。
			return idx, nil
		}
	}

	log.Debug("Index '%s' not found in table %s.%s", indexName, t.dbName, t.tableName) // 索引 '%s' 在表 '%s.%s' 中未找到。
	return nil, errors.ErrIndexNotFound.New(indexName)
}

// GetIndexes returns a list of all indexes on the table.
// It retrieves index definitions from the catalog.
// GetIndexes 返回表上所有索引的列表。
// 它从 catalog 检索索引定义。
func (t *BadgerTable) GetIndexes(ctx context.Context) ([]sql.Index, error) {
	log.Debug("BadgerTable GetIndexes called for %s.%s", t.dbName, t.tableName) // 调用 BadgerTable GetIndexes。

	// TODO: Retrieve all index definition metadata for this table from catalog in Badger.
	// Scan catalog keys with prefix like catalog:<db_name>:<table_name>:* and filter for index metadata type.
	//
	// TODO: 从 Badger 的 catalog 中检索此表的所有索引定义元数据。
	// 扫描前缀如 catalog:<db_name>:<table_name>:* 的 catalog key，并过滤索引元数据类型。
	log.Warn("GetIndexes implementation is a placeholder.", t.dbName, t.tableName) // GetIndexes 实现是占位符。

	// For now, return the in-memory list (placeholder logic)
	// 目前，返回内存列表（占位符逻辑）。
	return t.indexes, nil
}