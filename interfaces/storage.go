// Package interfaces defines common interfaces used across the guocedb project.
// This file, storage.go, specifies the interfaces for various storage components,
// providing an abstraction layer over different storage implementations
// (e.g., in-memory, disk-based, key-value stores).
// This abstraction promotes modularity, testability, and allows for
// easy swapping of underlying storage technologies.
package interfaces

// KeyType represents the type used for keys in the storage system.
// KeyType 表示存储系统中键的类型。
type KeyType []byte

// ValueType represents the type used for values in the storage system.
// ValueType 表示存储系统中值的类型。
type ValueType []byte

// StorageEngine defines the core interface for interacting with the underlying data storage mechanism.
// It provides fundamental CRUD (Create, Read, Update, Delete) operations at a key-value level.
// StorageEngine 定义了与底层数据存储机制交互的核心接口。
// 它在键值级别提供基本的 CRUD（创建、读取、更新、删除）操作。
type StorageEngine interface {
	// Name returns the name of the storage engine (e.g., "badger", "leveldb", "in-memory").
	// Name 返回存储引擎的名称（例如，“badger”、“leveldb”、“in-memory”）。
	Name() string

	// Open initializes and opens the storage engine. This typically involves
	// opening database files or establishing connections.
	// Open 初始化并打开存储引擎。这通常涉及打开数据库文件或建立连接。
	Open(path string) error

	// Close shuts down the storage engine, releasing resources.
	// Close 关闭存储引擎，释放资源。
	Close() error

	// Get retrieves the value associated with the given key.
	// Returns (nil, nil) if the key does not exist.
	// Get 检索与给定键关联的值。
	// 如果键不存在，则返回 (nil, nil)。
	Get(key KeyType) (ValueType, error)

	// Set stores the given key-value pair. If the key already exists, its value is updated.
	// Set 存储给定的键值对。如果键已存在，则更新其值。
	Set(key KeyType, value ValueType) error

	// Delete removes the key-value pair associated with the given key.
	// Delete 删除与给定键关联的键值对。
	Delete(key KeyType) error

	// NewTransaction creates a new transaction for atomic operations.
	// The transaction object provides methods for read and write operations that are
	// isolated until the transaction is committed.
	// NewTransaction 创建一个新的事务，用于原子操作。
	// 事务对象提供读取和写入操作的方法，这些操作在事务提交之前是隔离的。
	NewTransaction(update bool) (Transaction, error)

	// Iterator returns a new Iterator for traversing the key-value pairs in lexicographical order.
	// The caller is responsible for closing the iterator.
	// Iterator 返回一个新的迭代器，用于按字典顺序遍历键值对。
	// 调用者负责关闭迭代器。
	Iterator(prefix KeyType) (Iterator, error)

	// Sync ensures that all buffered writes are flushed to durable storage.
	// This might be a no-op for some engines that write directly to disk or have their
	// own durability mechanisms (e.g., WAL).
	// Sync 确保所有缓冲写入都刷新到持久存储。
	// 对于某些直接写入磁盘或拥有自己的持久性机制（例如 WAL）的引擎，这可能是一个空操作。
	Sync() error

	// Size returns the approximate size of the data stored by the engine in bytes.
	// Size 返回引擎存储的数据的大致大小（字节）。
	Size() (int64, error)
}

// Transaction defines the interface for database transactions, enabling
// atomic, isolated, and durable operations.
// Transaction 定义了数据库事务的接口，实现原子性、隔离性和持久性操作。
type Transaction interface {
	// Get retrieves the value associated with the given key within this transaction's scope.
	// Returns (nil, nil) if the key does not exist.
	// Get 在此事务范围内检索与给定键关联的值。
	// 如果键不存在，则返回 (nil, nil)。
	Get(key KeyType) (ValueType, error)

	// Set stores the given key-value pair within this transaction's scope.
	// Set 在此事务范围内存储给定的键值对。
	Set(key KeyType, value ValueType) error

	// Delete removes the key-value pair associated with the given key within this transaction's scope.
	// Delete 在此事务范围内删除与给定键关联的键值对。
	Delete(key KeyType) error

	// Commit applies all changes made within the transaction to the underlying storage.
	// Commit 将事务中发生的所有更改应用到底层存储。
	Commit() error

	// Discard discards all changes made within the transaction.
	// Discard 放弃事务中发生的所有更改。
	Discard()
}

// Iterator defines the interface for iterating over key-value pairs in a storage engine.
// Iterators provide a way to efficiently scan through ranges of keys.
// Iterator 定义了在存储引擎中迭代键值对的接口。
// 迭代器提供了一种有效扫描键范围的方法。
type Iterator interface {
	// Rewind seeks the iterator to the first key in the iteration range.
	// Rewind 将迭代器定位到迭代范围内的第一个键。
	Rewind()

	// Seek seeks the iterator to the first key that is greater than or equal to the given key.
	// Seek 将迭代器定位到大于或等于给定键的第一个键。
	Seek(key KeyType)

	// Next moves the iterator to the next key-value pair.
	// Returns true if there is a next element, false otherwise.
	// Next 将迭代器移动到下一个键值对。
	// 如果存在下一个元素，则返回 true，否则返回 false。
	Next() bool

	// Valid returns true if the iterator is currently positioned at a valid key-value pair.
	// Valid 返回 true 如果迭代器当前定位在一个有效的键值对。
	Valid() bool

	// Key returns the current key.
	// Key 返回当前键。
	Key() KeyType

	// Value returns the current value.
	// Value 返回当前值。
	Value() ValueType

	// Close closes the iterator and releases any associated resources.
	// Close 关闭迭代器并释放任何关联的资源。
	Close() error
}

// WAL represents the Write-Ahead Log interface.
// The WAL is crucial for ensuring data durability and atomicity.
// Every write operation to the database should first be logged to the WAL.
// WAL 代表预写日志接口。
// WAL 对于确保数据持久性和原子性至关重要。
// 对数据库的每次写入操作都应首先记录到 WAL。
type WAL interface {
	// WriteEntry appends a log entry to the WAL.
	// WriteEntry 将日志条目追加到 WAL。
	WriteEntry(entry []byte) error

	// ReadEntries reads log entries from the WAL starting from a given offset.
	// It's typically used for recovery after a crash.
	// ReadEntries 从给定偏移量开始从 WAL 读取日志条目。
	// 它通常用于崩溃后的恢复。
	ReadEntries(offset int64) ([][]byte, error)

	// Truncate discards log entries up to a certain offset.
	// This is important for WAL compaction and recycling space.
	// Truncate 丢弃直到某个偏移量的日志条目。
	// 这对于 WAL 压缩和回收空间很重要。
	Truncate(offset int64) error

	// Flush ensures that all buffered WAL entries are written to durable storage.
	// Flush 确保所有缓冲的 WAL 条目都写入持久存储。
	Flush() error

	// Close closes the WAL, releasing resources.
	// Close 关闭 WAL，释放资源。
	Close() error

	// Size returns the current size of the WAL in bytes.
	// Size 返回 WAL 的当前大小（字节）。
	Size() (int64, error)
}

// Index represents a generic interface for indexing data within guocedb.
// An index allows for efficient lookup and retrieval of records based on key values,
// beyond primary key lookups provided by the StorageEngine.
// Index 代表 guocedb 中索引数据的通用接口。
// 索引允许基于键值高效地查找和检索记录，超越 StorageEngine 提供的 Primary Key 查找。
type Index interface {
	// Name returns the name of the index (e.g., "btree", "hash").
	// Name 返回索引的名称（例如，“btree”、“hash”）。
	Name() string

	// Put adds or updates an entry in the index.
	// Key is the index key, and Value is typically the primary key or a pointer to the actual data.
	// Put 在索引中添加或更新条目。
	// Key 是索引键，Value 通常是主键或指向实际数据的指针。
	Put(key KeyType, value ValueType) error

	// Get retrieves the value (e.g., primary key) associated with the given index key.
	// Get 检索与给定索引键关联的值（例如，主键）。
	Get(key KeyType) (ValueType, error)

	// Delete removes an entry from the index.
	// Delete 从索引中删除条目。
	Delete(key KeyType) error

	// Close closes the index and releases any associated resources.
	// Close 关闭索引并释放任何关联的资源。
	Close() error

	// Count returns the number of entries in the index.
	// Count 返回索引中的条目数。
	Count() (int64, error)

	// Iterator returns a new Iterator for traversing the index entries.
	// The caller is responsible for closing the iterator.
	// Iterator 返回一个新的迭代器，用于遍历索引条目。
	// 调用者负责关闭迭代器。
	Iterator(prefix KeyType) (Iterator, error) // Reuses the generic Iterator interface
}
