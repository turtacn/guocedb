// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"bytes"
	"context"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
)

// BadgerRowIterator is an implementation of sql.RowIter that iterates over rows in a Badger database.
// BadgerRowIterator 是 sql.RowIter 的一个实现，用于遍历 Badger 数据库中的行。
type BadgerRowIterator struct {
	// db represents the Badger database handle.
	// db 代表 Badger 数据库句柄。
	db *badger.DB

	// txn represents the active Badger transaction.
	// txn 代表活跃的 Badger 事务。
	txn *badger.Txn

	// iterator is the underlying Badger iterator.
	// iterator 是底层的 Badger 迭代器。
	iterator *badger.Iterator

	// schema is the schema of the table being iterated.
	// schema 是正在迭代的表的模式。
	schema sql.Schema

	// prefix is the key prefix for the table data (data:<db>:<table>:<pk>).
	// prefix 是表数据（data:<db>:<table>:<pk>）的 key 前缀。
	prefix []byte
}

// NewBadgerRowIterator creates a new BadgerRowIterator.
// NewBadgerRowIterator 创建一个新的 BadgerRowIterator。
// It starts a new transaction and creates a new iterator.
// 它开始一个新的事务并创建一个新的迭代器。
func NewBadgerRowIterator(db *badger.DB, schema sql.Schema, dbName, tableName string) (*BadgerRowIterator, error) {
	// Start a read-only transaction for the iterator
	// 为迭代器开始一个只读事务
	txn := db.NewTransaction(false) // false for read-only
	opts := badger.DefaultIteratorOptions
	// Optionally set PrefetchSize, etc.
	// 可选地设置 PrefetchSize 等。
	iterator := txn.NewIterator(opts)

	// Construct the key prefix for the table data
	// 构造表数据的 key 前缀
	prefix := buildDataKeyPrefix(dbName, tableName)

	// Seek to the start of the table data range
	// 寻求到表数据范围的开始
	iterator.Seek(prefix)

	return &BadgerRowIterator{
		db:       db,
		txn:      txn,
		iterator: iterator,
		schema:   schema,
		prefix:   prefix,
	}, nil
}

// buildDataKeyPrefix constructs the key prefix for a table's data.
// buildDataKeyPrefix 构造表数据的 key 前缀。
func buildDataKeyPrefix(dbName, tableName string) []byte {
	prefix := bytes.Join([][]byte{
		NamespaceDataBytes,
		[]byte(dbName),
		[]byte(tableName),
	}, []byte{NsSep, Sep, Sep})
	// Append separator to ensure it's a prefix matching only this table's data keys
	// 追加分隔符以确保它是仅匹配此表数据 key 的前缀
	return append(prefix, Sep)
}

// Next retrieves the next row from the iterator.
// Next 从迭代器中检索下一行。
// It checks if the current key is within the table's data prefix.
// 它检查当前 key 是否在表数据的范围内。
func (i *BadgerRowIterator) Next(ctx context.Context) (sql.Row, error) {
	// Check if the iterator is valid and the current key is within the desired prefix
	// 检查迭代器是否有效且当前 key 是否在期望的前缀范围内
	if !i.iterator.ValidForPrefix(i.prefix) {
		return nil, nil // No more rows
	}

	item := i.iterator.Item()
	key := item.Key()

	// Decode the key to get primary key encoded bytes
	// 解码 key 获取主键编码字节
	_, _, pkEncoded, ok := DecodeDataKey(key)
	if !ok {
		// This indicates a corrupted or unexpected key format
		// 这表示 key 格式损坏或意外
		log.Error("Failed to decode data key: %v", key) // 解码数据 key 失败。
		return nil, errors.ErrDecodingFailed.New("data key decoding failed")
	}

	// Retrieve the value (row data)
	// 检索 value（行数据）
	val, err := item.ValueCopy(nil) // Use ValueCopy to get the value bytes
	if err != nil {
		log.Error("Failed to get value from Badger item for key %v: %v", key, err) // 从 Badger item 获取 value 失败。
		return nil, errors.ErrBadgerOperationFailed.New("failed to get value from Badger")
	}

	// Decode the row data using the table schema
	// 使用表模式解码行数据
	row, err := DecodeRow(val, i.schema)
	if err != nil {
		log.Error("Failed to decode row data for key %v: %v", key, err) // 解码行数据失败。
		return nil, errors.ErrDecodingFailed.New("row data decoding failed")
	}

	// Move the iterator to the next item
	// 将迭代器移动到下一个 item
	i.iterator.Next()

	return row, nil
}

// Close closes the iterator and the underlying transaction.
// Close 关闭迭代器和底层事务。
func (i *BadgerRowIterator) Close(ctx context.Context) error {
	log.Info("Closing BadgerRowIterator for table with prefix %v", i.prefix) // 关闭 BadgerRowIterator。
	i.iterator.Close()
	i.txn.Discard() // Discard the read-only transaction
	return nil
}

// Note: For index iteration, a separate iterator implementation will be needed,
// seeking on the index key prefix and decoding index keys.
//
// 注意：对于索引迭代，需要一个单独的迭代器实现，
// 寻求索引 key 前缀并解码索引 key。