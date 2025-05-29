package badger

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/turtacn/guocedb/common/types"
	encoding "github.com/turtacn/guocedb/storage/engines/badger"
)

// IteratorDirection 迭代方向 Iterator direction
type IteratorDirection int

const (
	// Forward 前向迭代 Forward iteration
	Forward IteratorDirection = iota
	// Backward 后向迭代 Backward iteration
	Backward
)

// IteratorType 迭代器类型 Iterator type
type IteratorType int

const (
	// TableIterator 表迭代器 Table iterator
	TableIterator IteratorType = iota
	// IndexIterator 索引迭代器 Index iterator
	IndexIterator
	// MetadataIterator 元数据迭代器 Metadata iterator
	MetadataIterator
	// PrefixIterator 前缀迭代器 Prefix iterator
	PrefixIterator
)

// IteratorOptions 迭代器选项 Iterator options
type IteratorOptions struct {
	// Direction 迭代方向 Iteration direction
	Direction IteratorDirection
	// PrefetchSize 预取大小 Prefetch size
	PrefetchSize int
	// PrefetchValues 是否预取值 Whether to prefetch values
	PrefetchValues bool
	// StartKey 起始键 Start key
	StartKey []byte
	// EndKey 结束键 End key
	EndKey []byte
	// Prefix 键前缀 Key prefix
	Prefix []byte
	// Filter 过滤器 Filter
	Filter FilterFunc
	// Limit 限制数量 Limit count
	Limit int64
	// Offset 偏移量 Offset
	Offset int64
	// BufferSize 缓冲区大小 Buffer size
	BufferSize int
	// Timeout 超时时间 Timeout
	Timeout time.Duration
	// ReadOnly 只读模式 Read-only mode
	ReadOnly bool
	// IgnoreDeleted 忽略已删除项 Ignore deleted items
	IgnoreDeleted bool
}

// FilterFunc 过滤器函数 Filter function
type FilterFunc func(key, value []byte) bool

// IteratorStats 迭代器统计信息 Iterator statistics
type IteratorStats struct {
	// ItemsRead 读取项目数 Items read count
	ItemsRead int64
	// ItemsFiltered 过滤项目数 Items filtered count
	ItemsFiltered int64
	// ItemsSkipped 跳过项目数 Items skipped count
	ItemsSkipped int64
	// BytesRead 读取字节数 Bytes read count
	BytesRead int64
	// SeekCount 定位次数 Seek count
	SeekCount int64
	// NextCount 下一项次数 Next count
	NextCount int64
	// PrevCount 上一项次数 Previous count
	PrevCount int64
	// BufferHits 缓冲区命中 Buffer hits
	BufferHits int64
	// BufferMisses 缓冲区未命中 Buffer misses
	BufferMisses int64
	// ElapsedTime 耗时 Elapsed time
	ElapsedTime time.Duration
	// StartTime 开始时间 Start time
	StartTime time.Time
	// EndTime 结束时间 End time
	EndTime time.Time
	// LastAccessTime 最后访问时间 Last access time
	LastAccessTime time.Time
}

// Iterator Badger数据迭代器 Badger data iterator
type Iterator struct {
	// 基本属性 Basic properties
	id      string                 // 迭代器ID Iterator ID
	txn     *badger.Txn            // Badger事务 Badger transaction
	it      *badger.Iterator       // Badger迭代器 Badger iterator
	encoder *encoding.KeyEncoder   // 键编码器 Key encoder
	decoder *encoding.ValueEncoder // 值解码器 Value decoder
	logger  *zap.Logger            // 日志记录器 Logger

	// 配置选项 Configuration options
	options   *IteratorOptions // 迭代器选项 Iterator options
	iterType  IteratorType     // 迭代器类型 Iterator type
	tableName string           // 表名 Table name
	indexName string           // 索引名 Index name

	// 状态管理 State management
	mu         sync.RWMutex // 读写锁 Read-write lock
	started    bool         // 是否已开始 Whether started
	closed     bool         // 是否已关闭 Whether closed
	valid      bool         // 是否有效 Whether valid
	eof        bool         // 是否到达末尾 Whether reached EOF
	currentKey []byte       // 当前键 Current key
	currentVal []byte       // 当前值 Current value
	position   int64        // 当前位置 Current position

	// 缓冲和预取 Buffering and prefetching
	buffer     []*IteratorItem // 缓冲区 Buffer
	bufferIdx  int             // 缓冲区索引 Buffer index
	prefetched bool            // 是否已预取 Whether prefetched
	bufferPos  int64           // 缓冲区位置 Buffer position

	// 统计信息 Statistics
	stats *IteratorStats // 统计信息 Statistics

	// 上下文管理 Context management
	ctx    context.Context    // 上下文 Context
	cancel context.CancelFunc // 取消函数 Cancel function

	// 高级功能 Advanced features
	snapshot *IteratorSnapshot // 快照 Snapshot
	bookmark []byte            // 书签 Bookmark
}

// IteratorItem 迭代器项目 Iterator item
type IteratorItem struct {
	Key       []byte    // 键 Key
	Value     []byte    // 值 Value
	Version   uint64    // 版本 Version
	ExpiresAt uint64    // 过期时间 Expiration time
	UserMeta  byte      // 用户元数据 User metadata
	Timestamp time.Time // 时间戳 Timestamp
	IsDeleted bool      // 是否已删除 Whether deleted
	Size      int       // 项目大小 Item size
}

// IteratorSnapshot 迭代器快照 Iterator snapshot
type IteratorSnapshot struct {
	Key       []byte            // 键 Key
	Position  int64             // 位置 Position
	Direction IteratorDirection // 方向 Direction
	Timestamp time.Time         // 时间戳 Timestamp
}

// IteratorManager 迭代器管理器 Iterator manager
type IteratorManager struct {
	mu        sync.RWMutex           // 读写锁 Read-write lock
	iterators map[string]*Iterator   // 活动迭代器 Active iterators
	db        *badger.DB             // Badger数据库 Badger database
	encoder   *encoding.KeyEncoder   // 键编码器 Key encoder
	decoder   *encoding.ValueEncoder // 值解码器 Value decoder
	logger    *zap.Logger            // 日志记录器 Logger

	// 配置 Configuration
	maxIterators    int           // 最大迭代器数量 Maximum iterator count
	defaultTimeout  time.Duration // 默认超时时间 Default timeout
	cleanupInterval time.Duration // 清理间隔 Cleanup interval
	enableMetrics   bool          // 启用指标 Enable metrics

	// 统计 Statistics
	totalCreated int64     // 总创建数 Total created count
	totalClosed  int64     // 总关闭数 Total closed count
	totalErrors  int64     // 总错误数 Total error count
	lastCleanup  time.Time // 最后清理时间 Last cleanup time

	// 关闭控制 Shutdown control
	closeOnce sync.Once     // 关闭一次 Close once
	closed    chan struct{} // 关闭信号 Close signal
}

// NewIteratorManager 创建迭代器管理器 Create iterator manager
func NewIteratorManager(db *badger.DB, encoder *encoding.KeyEncoder, decoder *encoding.ValueEncoder, logger *zap.Logger) *IteratorManager {
	im := &IteratorManager{
		iterators:       make(map[string]*Iterator),
		db:              db,
		encoder:         encoder,
		decoder:         decoder,
		logger:          logger,
		maxIterators:    1000,
		defaultTimeout:  30 * time.Minute,
		cleanupInterval: 5 * time.Minute,
		enableMetrics:   true,
		closed:          make(chan struct{}),
	}

	// 启动清理协程 Start cleanup goroutine
	go im.cleanupExpiredIterators()

	return im
}

// CreateTableIterator 创建表迭代器 Create table iterator
func (im *IteratorManager) CreateTableIterator(ctx context.Context, txn *badger.Txn, tableName string, options *IteratorOptions) (*Iterator, error) {
	return im.createIterator(ctx, txn, TableIterator, tableName, "", options)
}

// CreateIndexIterator 创建索引迭代器 Create index iterator
func (im *IteratorManager) CreateIndexIterator(ctx context.Context, txn *badger.Txn, tableName, indexName string, options *IteratorOptions) (*Iterator, error) {
	return im.createIterator(ctx, txn, IndexIterator, tableName, indexName, options)
}

// CreateMetadataIterator 创建元数据迭代器 Create metadata iterator
func (im *IteratorManager) CreateMetadataIterator(ctx context.Context, txn *badger.Txn, options *IteratorOptions) (*Iterator, error) {
	return im.createIterator(ctx, txn, MetadataIterator, "", "", options)
}

// CreatePrefixIterator 创建前缀迭代器 Create prefix iterator
func (im *IteratorManager) CreatePrefixIterator(ctx context.Context, txn *badger.Txn, prefix []byte, options *IteratorOptions) (*Iterator, error) {
	if options == nil {
		options = &IteratorOptions{}
	}
	options.Prefix = prefix
	return im.createIterator(ctx, txn, PrefixIterator, "", "", options)
}

// createIterator 创建迭代器 Create iterator
func (im *IteratorManager) createIterator(ctx context.Context, txn *badger.Txn, iterType IteratorType, tableName, indexName string, options *IteratorOptions) (*Iterator, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// 检查管理器是否已关闭 Check if manager is closed
	select {
	case <-im.closed:
		return nil, fmt.Errorf("iterator manager is closed")
	default:
	}

	// 检查迭代器数量限制 Check iterator count limit
	if len(im.iterators) >= im.maxIterators {
		return nil, fmt.Errorf("maximum number of iterators (%d) reached", im.maxIterators)
	}

	// 设置默认选项 Set default options
	if options == nil {
		options = &IteratorOptions{}
	}
	im.setDefaultOptions(options)

	// 创建上下文 Create context
	iterCtx, cancel := context.WithCancel(ctx)
	if options.Timeout > 0 {
		iterCtx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	// 生成迭代器ID Generate iterator ID
	iterID := im.generateIteratorID(iterType, tableName, indexName)

	// 创建迭代器 Create iterator
	iterator := &Iterator{
		id:        iterID,
		txn:       txn,
		encoder:   im.encoder,
		decoder:   im.decoder,
		logger:    im.logger,
		options:   options,
		iterType:  iterType,
		tableName: tableName,
		indexName: indexName,
		stats: &IteratorStats{
			StartTime:      time.Now(),
			LastAccessTime: time.Now(),
		},
		ctx:    iterCtx,
		cancel: cancel,
	}

	// 初始化迭代器 Initialize iterator
	if err := iterator.initialize(); err != nil {
		cancel()
		im.totalErrors++
		return nil, fmt.Errorf("failed to initialize iterator: %w", err)
	}

	// 注册迭代器 Register iterator
	im.iterators[iterID] = iterator
	im.totalCreated++

	im.logger.Debug("Iterator created",
		"id", iterID,
		"type", iterType,
		"table", tableName,
		"index", indexName,
		"active_count", len(im.iterators))

	return iterator, nil
}

// setDefaultOptions 设置默认选项 Set default options
func (im *IteratorManager) setDefaultOptions(options *IteratorOptions) {
	if options.PrefetchSize == 0 {
		options.PrefetchSize = 100
	}
	if options.BufferSize == 0 {
		options.BufferSize = 1000
	}
	if options.Timeout == 0 {
		options.Timeout = im.defaultTimeout
	}
	// 默认预取值 Default prefetch values
	options.PrefetchValues = true
}

// generateIteratorID 生成迭代器ID Generate iterator ID
func (im *IteratorManager) generateIteratorID(iterType IteratorType, tableName, indexName string) string {
	timestamp := time.Now().UnixNano()
	switch iterType {
	case TableIterator:
		return fmt.Sprintf("table_%s_%d", tableName, timestamp)
	case IndexIterator:
		return fmt.Sprintf("index_%s_%s_%d", tableName, indexName, timestamp)
	case MetadataIterator:
		return fmt.Sprintf("metadata_%d", timestamp)
	case PrefixIterator:
		return fmt.Sprintf("prefix_%d", timestamp)
	default:
		return fmt.Sprintf("unknown_%d", timestamp)
	}
}

// initialize 初始化迭代器 Initialize iterator
func (iter *Iterator) initialize() error {
	// 创建Badger迭代器选项 Create Badger iterator options
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = iter.options.PrefetchSize
	opts.PrefetchValues = iter.options.PrefetchValues
	opts.Reverse = iter.options.Direction == Backward

	// 创建Badger迭代器 Create Badger iterator
	iter.it = iter.txn.NewIterator(opts)

	// 设置键前缀 Set key prefix
	if iter.options.Prefix == nil {
		iter.options.Prefix = iter.getDefaultPrefix()
	}

	// 初始化缓冲区 Initialize buffer
	if iter.options.BufferSize > 0 {
		iter.buffer = make([]*IteratorItem, 0, iter.options.BufferSize)
	}

	return nil
}

// getDefaultPrefix 获取默认前缀 Get default prefix
func (iter *Iterator) getDefaultPrefix() []byte {
	switch iter.iterType {
	case TableIterator:
		return iter.encoder.EncodeRowPrefix(iter.tableName)
	case IndexIterator:
		return iter.encoder.EncodeIndexPrefix(iter.tableName, iter.indexName)
	case MetadataIterator:
		return iter.encoder.EncodeMetadataPrefix()
	case PrefixIterator:
		return iter.options.Prefix
	default:
		return []byte{}
	}
}

// Seek 定位到指定键 Seek to specified key
func (iter *Iterator) Seek(key []byte) bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return false
	}

	iter.stats.SeekCount++
	iter.stats.LastAccessTime = time.Now()

	// 清空缓冲区 Clear buffer
	iter.clearBuffer()

	// 执行定位 Perform seek
	if key == nil {
		if iter.options.StartKey != nil {
			iter.it.Seek(iter.options.StartKey)
		} else {
			iter.it.Seek(iter.options.Prefix)
		}
	} else {
		iter.it.Seek(key)
		// 更新书签 Update bookmark
		iter.bookmark = append([]byte(nil), key...)
	}

	// 更新状态 Update state
	iter.started = true
	iter.updateCurrentItem()

	return iter.valid
}

// SeekToFirst 定位到第一项 Seek to first item
func (iter *Iterator) SeekToFirst() bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return false
	}

	iter.stats.SeekCount++
	iter.stats.LastAccessTime = time.Now()
	iter.clearBuffer()

	if iter.options.Direction == Forward {
		if iter.options.StartKey != nil {
			iter.it.Seek(iter.options.StartKey)
		} else {
			iter.it.Seek(iter.options.Prefix)
		}
	} else {
		iter.it.SeekToLast()
		// 反向迭代时需要确保在前缀范围内 For backward iteration, ensure within prefix range
		for iter.it.Valid() && !iter.hasValidPrefix() {
			iter.it.Prev()
		}
	}

	iter.started = true
	iter.position = 0
	iter.updateCurrentItem()

	return iter.valid
}

// SeekToLast 定位到最后一项 Seek to last item
func (iter *Iterator) SeekToLast() bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return false
	}

	iter.stats.SeekCount++
	iter.stats.LastAccessTime = time.Now()
	iter.clearBuffer()

	if iter.options.Direction == Backward {
		if iter.options.EndKey != nil {
			iter.it.Seek(iter.options.EndKey)
			if !iter.it.Valid() {
				iter.it.SeekToLast()
			}
		} else {
			iter.it.SeekToLast()
		}
		// 确保在前缀范围内 Ensure within prefix range
		for iter.it.Valid() && !iter.hasValidPrefix() {
			iter.it.Prev()
		}
	} else {
		// 前向迭代时找到前缀的最后一个键 For forward iteration, find last key of prefix
		prefix := iter.options.Prefix
		lastKey := iter.getLastKeyForPrefix(prefix)
		iter.it.Seek(lastKey)
		if !iter.it.Valid() {
			iter.it.SeekToLast()
		}
		// 向前找到前缀范围内的最后一个键 Move back to find last key within prefix range
		for iter.it.Valid() && !iter.hasValidPrefix() {
			iter.it.Prev()
		}
	}

	iter.started = true
	iter.updateCurrentItem()

	return iter.valid
}

// Next 移动到下一项 Move to next item
func (iter *Iterator) Next() bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed || !iter.started {
		return false
	}

	iter.stats.NextCount++
	iter.stats.LastAccessTime = time.Now()

	// 如果有缓冲区数据，先使用缓冲区 If buffer has data, use buffer first
	if iter.hasBufferedItems() {
		iter.bufferIdx++
		if iter.bufferIdx < len(iter.buffer) {
			item := iter.buffer[iter.bufferIdx]
			iter.currentKey = item.Key
			iter.currentVal = item.Value
			iter.valid = true
			iter.position++
			iter.stats.ItemsRead++
			iter.stats.BufferHits++
			return true
		}
		// 缓冲区用完，清空并继续从迭代器读取 Buffer exhausted, clear and continue reading from iterator
		iter.clearBuffer()
		iter.stats.BufferMisses++
	}

	// 移动迭代器 Move iterator
	if iter.options.Direction == Forward {
		iter.it.Next()
	} else {
		iter.it.Prev()
	}

	iter.position++
	iter.updateCurrentItem()

	// 尝试预取数据到缓冲区 Try to prefetch data to buffer
	if iter.valid && iter.options.BufferSize > 0 && !iter.prefetched {
		iter.prefetchToBuffer()
	}

	return iter.valid
}

// Prev 移动到上一项 Move to previous item
func (iter *Iterator) Prev() bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed || !iter.started {
		return false
	}

	iter.stats.PrevCount++
	iter.stats.LastAccessTime = time.Now()

	// 如果有缓冲区数据且正在反向遍历 If buffer has data and iterating backward
	if iter.hasBufferedItems() && iter.options.Direction == Backward {
		iter.bufferIdx++
		if iter.bufferIdx < len(iter.buffer) {
			item := iter.buffer[iter.bufferIdx]
			iter.currentKey = item.Key
			iter.currentVal = item.Value
			iter.valid = true
			iter.position--
			iter.stats.ItemsRead++
			iter.stats.BufferHits++
			return true
		}
		iter.clearBuffer()
		iter.stats.BufferMisses++
	}

	// 移动迭代器 Move iterator
	if iter.options.Direction == Forward {
		iter.it.Prev()
	} else {
		iter.it.Next()
	}

	iter.position--
	iter.updateCurrentItem()

	return iter.valid
}

// Valid 检查迭代器是否有效 Check if iterator is valid
func (iter *Iterator) Valid() bool {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	return iter.valid && !iter.closed && !iter.eof
}

// Key 获取当前键 Get current key
func (iter *Iterator) Key() []byte {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	if iter.valid {
		return append([]byte(nil), iter.currentKey...)
	}
	return nil
}

// Value 获取当前值 Get current value
func (iter *Iterator) Value() []byte {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	if iter.valid {
		return append([]byte(nil), iter.currentVal...)
	}
	return nil
}

// Item 获取当前项目 Get current item
func (iter *Iterator) Item() *IteratorItem {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	if !iter.valid {
		return nil
	}

	item := iter.it.Item()
	return &IteratorItem{
		Key:       append([]byte(nil), iter.currentKey...),
		Value:     append([]byte(nil), iter.currentVal...),
		Version:   item.Version(),
		ExpiresAt: item.ExpiresAt(),
		UserMeta:  item.UserMeta(),
		Timestamp: time.Unix(int64(item.Version()), 0),
		IsDeleted: item.IsDeletedOrExpired(),
		Size:      len(iter.currentKey) + len(iter.currentVal),
	}
}

// DecodeRow 解码当前行 Decode current row
func (iter *Iterator) DecodeRow() (*types.Row, error) {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	if !iter.valid || iter.iterType != TableIterator {
		return nil, fmt.Errorf("invalid iterator or not a table iterator")
	}

	return iter.decoder.DecodeRow(iter.currentVal)
}

// DecodeKey 解码当前键 Decode current key
func (iter *Iterator) DecodeKey() ([][]byte, error) {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	if !iter.valid {
		return nil, fmt.Errorf("invalid iterator")
	}

	switch iter.iterType {
	case TableIterator:
		return iter.encoder.DecodeRowKey(iter.currentKey)
	case IndexIterator:
		return iter.encoder.DecodeIndexKey(iter.currentKey)
	default:
		return [][]byte{iter.currentKey}, nil
	}
}

// Collect 收集所有项目 Collect all items
func (iter *Iterator) Collect(maxItems int) ([]*IteratorItem, error) {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return nil, fmt.Errorf("iterator is closed")
	}

	var items []*IteratorItem
	count := 0

	// 如果还未开始，先定位到第一项 If not started, seek to first item
	if !iter.started {
		iter.SeekToFirst()
	}

	for iter.valid && (maxItems == 0 || count < maxItems) {
		// 检查上下文是否已取消 Check if context is cancelled
		select {
		case <-iter.ctx.Done():
			return items, iter.ctx.Err()
		default:
		}

		item := &IteratorItem{
			Key:   append([]byte(nil), iter.currentKey...),
			Value: append([]byte(nil), iter.currentVal...),
		}

		// 应用过滤器 Apply filter
		if iter.options.Filter == nil || iter.options.Filter(item.Key, item.Value) {
			items = append(items, item)
			count++
		} else {
			iter.stats.ItemsFiltered++
		}

		// 移动到下一项 Move to next item
		if iter.options.Direction == Forward {
			iter.it.Next()
		} else {
			iter.it.Prev()
		}
		iter.updateCurrentItem()
	}

	return items, nil
}

// Count 统计符合条件的项目数量 Count items that match conditions
func (iter *Iterator) Count() (int64, error) {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return 0, fmt.Errorf("iterator is closed")
	}

	var count int64

	// 保存当前状态 Save current state
	currentSnapshot := iter.createSnapshot()
	defer func() {
		// 恢复状态 Restore state
		iter.restoreFromSnapshot(currentSnapshot)
	}()

	// 如果还未开始，先定位到第一项 If not started, seek to first item
	if !iter.started {
		iter.SeekToFirst()
	}

	for iter.valid {
		// 检查上下文是否已取消 Check if context is cancelled
		select {
		case <-iter.ctx.Done():
			return count, iter.ctx.Err()
		default:
		}

		// 应用过滤器 Apply filter
		if iter.options.Filter == nil || iter.options.Filter(iter.currentKey, iter.currentVal) {
			count++
		}

		// 移动到下一项 Move to next item
		if iter.options.Direction == Forward {
			iter.it.Next()
		} else {
			iter.it.Prev()
		}
		iter.updateCurrentItem()
	}

	return count, nil
}

// Skip 跳过指定数量的项目 Skip specified number of items
func (iter *Iterator) Skip(count int64) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return fmt.Errorf("iterator is closed")
	}

	var skipped int64
	for iter.valid && skipped < count {
		// 检查上下文是否已取消 Check if context is cancelled
		select {
		case <-iter.ctx.Done():
			return iter.ctx.Err()
		default:
		}

		// 移动到下一项 Move to next item
		if iter.options.Direction == Forward {
			iter.it.Next()
		} else {
			iter.it.Prev()
		}
		iter.updateCurrentItem()
		skipped++
		iter.stats.ItemsSkipped++
	}

	return nil
}

// Reset 重置迭代器 Reset iterator
func (iter *Iterator) Reset() error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return fmt.Errorf("iterator is closed")
	}

	// 清空缓冲区 Clear buffer
	iter.clearBuffer()

	// 重置状态 Reset state
	iter.started = false
	iter.valid = false
	iter.eof = false
	iter.prefetched = false
	iter.currentKey = nil
	iter.currentVal = nil
	iter.position = 0
	iter.bookmark = nil

	// 重置统计信息 Reset statistics
	iter.stats = &IteratorStats{
		StartTime:      time.Now(),
		LastAccessTime: time.Now(),
	}

	return nil
}

// GetStats 获取统计信息 Get statistics
func (iter *Iterator) GetStats() *IteratorStats {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	stats := *iter.stats
	if iter.stats.EndTime.IsZero() {
		stats.ElapsedTime = time.Since(iter.stats.StartTime)
	}

	return &stats
}

// GetPosition 获取当前位置 Get current position
func (iter *Iterator) GetPosition() int64 {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	return iter.position
}

// CreateSnapshot 创建快照 Create snapshot
func (iter *Iterator) CreateSnapshot() *IteratorSnapshot {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	return iter.createSnapshot()
}

// createSnapshot 内部创建快照 Internal create snapshot
func (iter *Iterator) createSnapshot() *IteratorSnapshot {
	return &IteratorSnapshot{
		Key:       append([]byte(nil), iter.currentKey...),
		Position:  iter.position,
		Direction: iter.options.Direction,
		Timestamp: time.Now(),
	}
}

// RestoreFromSnapshot 从快照恢复 Restore from snapshot
func (iter *Iterator) RestoreFromSnapshot(snapshot *IteratorSnapshot) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return fmt.Errorf("iterator is closed")
	}

	return iter.restoreFromSnapshot(snapshot)
}

// restoreFromSnapshot 内部从快照恢复 Internal restore from snapshot
func (iter *Iterator) restoreFromSnapshot(snapshot *IteratorSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot is nil")
	}

	// 清空缓冲区 Clear buffer
	iter.clearBuffer()

	// 恢复位置 Restore position
	if len(snapshot.Key) > 0 {
		iter.it.Seek(snapshot.Key)
		iter.updateCurrentItem()
	}

	iter.position = snapshot.Position
	iter.started = true

	return nil
}

// SetBookmark 设置书签 Set bookmark
func (iter *Iterator) SetBookmark() {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.valid {
		iter.bookmark = append([]byte(nil), iter.currentKey...)
	}
}

// SeekToBookmark 定位到书签 Seek to bookmark
func (iter *Iterator) SeekToBookmark() bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed || iter.bookmark == nil {
		return false
	}

	iter.clearBuffer()
	iter.it.Seek(iter.bookmark)
	iter.updateCurrentItem()

	return iter.valid
}

// Close 关闭迭代器 Close iterator
func (iter *Iterator) Close() error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return nil
	}

	// 关闭Badger迭代器 Close Badger iterator
	if iter.it != nil {
		iter.it.Close()
	}

	// 取消上下文 Cancel context
	if iter.cancel != nil {
		iter.cancel()
	}

	// 清空缓冲区 Clear buffer
	iter.clearBuffer()

	// 更新统计信息 Update statistics
	iter.stats.EndTime = time.Now()
	iter.stats.ElapsedTime = iter.stats.EndTime.Sub(iter.stats.StartTime)

	iter.closed = true

	iter.logger.Debug("Iterator closed",
		"id", iter.id,
		"items_read", iter.stats.ItemsRead,
		"elapsed_time", iter.stats.ElapsedTime)

	return nil
}

// updateCurrentItem 更新当前项目 Update current item
func (iter *Iterator) updateCurrentItem() {
	iter.valid = iter.it.Valid() && iter.hasValidPrefix() && iter.withinRange() && iter.withinLimit()

	if iter.valid {
		item := iter.it.Item()
		iter.currentKey = item.KeyCopy(nil)

		// 检查是否忽略已删除项 Check if ignoring deleted items
		if iter.options.IgnoreDeleted && item.IsDeletedOrExpired() {
			// 跳过已删除项 Skip deleted items
			if iter.options.Direction == Forward {
				iter.it.Next()
			} else {
				iter.it.Prev()
			}
			iter.updateCurrentItem() // 递归调用 Recursive call
			return
		}

		// 根据选项决定是否获取值 Based on options, decide whether to get value
		if iter.options.PrefetchValues {
			err := item.Value(func(val []byte) error {
				iter.currentVal = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				iter.valid = false
				return
			}
		}

		// 应用过滤器 Apply filter
		if iter.options.Filter != nil && !iter.options.Filter(iter.currentKey, iter.currentVal) {
			iter.stats.ItemsFiltered++
			// 如果不匹配过滤器，移动到下一项 If doesn't match filter, move to next item
			if iter.options.Direction == Forward {
				iter.it.Next()
			} else {
				iter.it.Prev()
			}
			iter.updateCurrentItem() // 递归调用 Recursive call
			return
		}

		iter.stats.ItemsRead++
		iter.stats.BytesRead += int64(len(iter.currentKey) + len(iter.currentVal))
	} else {
		iter.eof = true
	}
}

// hasValidPrefix 检查是否有有效前缀 Check if has valid prefix
func (iter *Iterator) hasValidPrefix() bool {
	if iter.options.Prefix == nil || len(iter.options.Prefix) == 0 {
		return true
	}

	key := iter.it.Item().Key()
	return bytes.HasPrefix(key, iter.options.Prefix)
}

// withinRange 检查是否在范围内 Check if within range
func (iter *Iterator) withinRange() bool {
	key := iter.it.Item().Key()

	// 检查起始键 Check start key
	if iter.options.StartKey != nil && bytes.Compare(key, iter.options.StartKey) < 0 {
		return false
	}

	// 检查结束键 Check end key
	if iter.options.EndKey != nil && bytes.Compare(key, iter.options.EndKey) >= 0 {
		return false
	}

	return true
}

// withinLimit 检查是否在限制内 Check if within limit
func (iter *Iterator) withinLimit() bool {
	if iter.options.Limit > 0 && iter.stats.ItemsRead >= iter.options.Limit {
		return false
	}

	// 检查偏移量 Check offset
	if iter.options.Offset > 0 && iter.stats.ItemsRead < iter.options.Offset {
		return false
	}

	return true
}

// hasBufferedItems 检查是否有缓冲项目 Check if has buffered items
func (iter *Iterator) hasBufferedItems() bool {
	return len(iter.buffer) > 0 && iter.bufferIdx < len(iter.buffer)-1
}

// clearBuffer 清空缓冲区 Clear buffer
func (iter *Iterator) clearBuffer() {
	iter.buffer = iter.buffer[:0]
	iter.bufferIdx = -1
	iter.prefetched = false
	iter.bufferPos = 0
}

// prefetchToBuffer 预取数据到缓冲区 Prefetch data to buffer
func (iter *Iterator) prefetchToBuffer() {
	if iter.options.BufferSize <= 0 {
		return
	}

	iter.clearBuffer()
	count := 0

	// 保存当前位置 Save current position
	currentKey := append([]byte(nil), iter.currentKey...)

	// 预取后续数据 Prefetch subsequent data
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = iter.options.PrefetchValues
	tempIt := iter.txn.NewIterator(opts)
	defer tempIt.Close()

	// 定位到当前位置的下一个 Seek to next position from current
	tempIt.Seek(currentKey)
	if tempIt.Valid() {
		if iter.options.Direction == Forward {
			tempIt.Next()
		} else {
			tempIt.Prev()
		}
	}

	for tempIt.Valid() && count < iter.options.BufferSize {
		if !iter.hasValidPrefixForKey(tempIt.Item().Key()) {
			break
		}

		if !iter.withinRangeForKey(tempIt.Item().Key()) {
			break
		}

		item := &IteratorItem{
			Key: tempIt.Item().KeyCopy(nil),
		}

		// 获取值 Get value
		if iter.options.PrefetchValues {
			err := tempIt.Item().Value(func(val []byte) error {
				item.Value = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				break
			}
		}

		// 应用过滤器 Apply filter
		if iter.options.Filter == nil || iter.options.Filter(item.Key, item.Value) {
			iter.buffer = append(iter.buffer, item)
			count++
		}

		if iter.options.Direction == Forward {
			tempIt.Next()
		} else {
			tempIt.Prev()
		}
	}

	iter.prefetched = true
	iter.bufferPos = iter.position
}

// hasValidPrefixForKey 检查指定键是否有有效前缀 Check if specified key has valid prefix
func (iter *Iterator) hasValidPrefixForKey(key []byte) bool {
	if iter.options.Prefix == nil || len(iter.options.Prefix) == 0 {
		return true
	}

	return bytes.HasPrefix(key, iter.options.Prefix)
}

// withinRangeForKey 检查指定键是否在范围内 Check if specified key is within range
func (iter *Iterator) withinRangeForKey(key []byte) bool {
	// 检查起始键 Check start key
	if iter.options.StartKey != nil && bytes.Compare(key, iter.options.StartKey) < 0 {
		return false
	}

	// 检查结束键 Check end key
	if iter.options.EndKey != nil && bytes.Compare(key, iter.options.EndKey) >= 0 {
		return false
	}

	return true
}

// getLastKeyForPrefix 获取前缀的最后一个键 Get last key for prefix
func (iter *Iterator) getLastKeyForPrefix(prefix []byte) []byte {
	if len(prefix) == 0 {
		return []byte{0xFF}
	}

	// 复制前缀并增加最后一个字节 Copy prefix and increment last byte
	lastKey := make([]byte, len(prefix))
	copy(lastKey, prefix)

	// 从末尾开始递增 Increment from the end
	for i := len(lastKey) - 1; i >= 0; i-- {
		if lastKey[i] < 0xFF {
			lastKey[i]++
			break
		}
		lastKey[i] = 0
		if i == 0 {
			// 所有字节都是 0xFF，添加一个字节 All bytes are 0xFF, add one byte
			lastKey = append(lastKey, 0x00)
		}
	}

	return lastKey
}

// ForEach 遍历所有项目 Iterate over all items
func (iter *Iterator) ForEach(fn func(key, value []byte) error) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	if iter.closed {
		return fmt.Errorf("iterator is closed")
	}

	// 如果还未开始，先定位到第一项 If not started, seek to first item
	if !iter.started {
		iter.SeekToFirst()
	}

	for iter.valid {
		// 检查上下文是否已取消 Check if context is cancelled
		select {
		case <-iter.ctx.Done():
			return iter.ctx.Err()
		default:
		}

		// 调用处理函数 Call processing function
		if err := fn(iter.currentKey, iter.currentVal); err != nil {
			return err
		}

		// 移动到下一项 Move to next item
		if iter.options.Direction == Forward {
			iter.it.Next()
		} else {
			iter.it.Prev()
		}
		iter.updateCurrentItem()
	}

	return nil
}

// String 字符串表示 String representation
func (iter *Iterator) String() string {
	iter.mu.RLock()
	defer iter.mu.RUnlock()

	return fmt.Sprintf("Iterator{id=%s, type=%d, table=%s, index=%s, valid=%t, closed=%t, position=%d}",
		iter.id, iter.iterType, iter.tableName, iter.indexName, iter.valid, iter.closed, iter.position)
}

// IsTableIterator 检查是否为表迭代器 Check if is table iterator
func (iter *Iterator) IsTableIterator() bool {
	return iter.iterType == TableIterator
}

// IsIndexIterator 检查是否为索引迭代器 Check if is index iterator
func (iter *Iterator) IsIndexIterator() bool {
	return iter.iterType == IndexIterator
}

// IsMetadataIterator 检查是否为元数据迭代器 Check if is metadata iterator
func (iter *Iterator) IsMetadataIterator() bool {
	return iter.iterType == MetadataIterator
}

// IsPrefixIterator 检查是否为前缀迭代器 Check if is prefix iterator
func (iter *Iterator) IsPrefixIterator() bool {
	return iter.iterType == PrefixIterator
}

// GetTableName 获取表名 Get table name
func (iter *Iterator) GetTableName() string {
	return iter.tableName
}

// GetIndexName 获取索引名 Get index name
func (iter *Iterator) GetIndexName() string {
	return iter.indexName
}

// GetID 获取迭代器ID Get iterator ID
func (iter *Iterator) GetID() string {
	return iter.id
}

// GetDirection 获取迭代方向 Get iteration direction
func (iter *Iterator) GetDirection() IteratorDirection {
	return iter.options.Direction
}

// GetIterator 获取指定ID的迭代器 Get iterator by ID
func (im *IteratorManager) GetIterator(id string) (*Iterator, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	iter, exists := im.iterators[id]
	return iter, exists
}

// CloseIterator 关闭指定迭代器 Close specified iterator
func (im *IteratorManager) CloseIterator(id string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	iter, exists := im.iterators[id]
	if !exists {
		return fmt.Errorf("iterator %s not found", id)
	}

	if err := iter.Close(); err != nil {
		im.totalErrors++
		return fmt.Errorf("failed to close iterator %s: %w", id, err)
	}

	delete(im.iterators, id)
	im.totalClosed++

	return nil
}

// CloseAllIterators 关闭所有迭代器 Close all iterators
func (im *IteratorManager) CloseAllIterators() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	var errors []string

	for id, iter := range im.iterators {
		if err := iter.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close iterator %s: %v", id, err))
			im.totalErrors++
		} else {
			im.totalClosed++
		}
	}

	// 清空迭代器映射 Clear iterator map
	im.iterators = make(map[string]*Iterator)

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing iterators: %v", errors)
	}

	return nil
}

// GetActiveIteratorCount 获取活动迭代器数量 Get active iterator count
func (im *IteratorManager) GetActiveIteratorCount() int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return len(im.iterators)
}

// GetIteratorStats 获取迭代器统计信息 Get iterator statistics
func (im *IteratorManager) GetIteratorStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"active_count":    len(im.iterators),
		"total_created":   im.totalCreated,
		"total_closed":    im.totalClosed,
		"total_errors":    im.totalErrors,
		"max_iterators":   im.maxIterators,
		"default_timeout": im.defaultTimeout,
		"last_cleanup":    im.lastCleanup,
	}
}

// ListActiveIterators 列出活动迭代器 List active iterators
func (im *IteratorManager) ListActiveIterators() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var ids []string
	for id := range im.iterators {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	return ids
}

// cleanupExpiredIterators 清理过期迭代器 Cleanup expired iterators
func (im *IteratorManager) cleanupExpiredIterators() {
	ticker := time.NewTicker(im.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			im.performCleanup()
		case <-im.closed:
			return
		}
	}
}

// performCleanup 执行清理 Perform cleanup
func (im *IteratorManager) performCleanup() {
	im.mu.Lock()
	defer im.mu.Unlock()

	var expiredIDs []string
	now := time.Now()

	for id, iter := range im.iterators {
		// 检查迭代器是否过期 Check if iterator is expired
		select {
		case <-iter.ctx.Done():
			expiredIDs = append(expiredIDs, id)
		default:
			// 检查是否超时 Check if timed out
			if iter.options.Timeout > 0 && now.Sub(iter.stats.StartTime) > iter.options.Timeout {
				expiredIDs = append(expiredIDs, id)
			}
			// 检查是否长时间未访问 Check if not accessed for long time
			if now.Sub(iter.stats.LastAccessTime) > im.defaultTimeout {
				expiredIDs = append(expiredIDs, id)
			}
		}
	}

	// 关闭过期迭代器 Close expired iterators
	for _, id := range expiredIDs {
		if iter, exists := im.iterators[id]; exists {
			if err := iter.Close(); err != nil {
				im.logger.Warn("Failed to close expired iterator",
					"id", id,
					"error", err)
				im.totalErrors++
			} else {
				delete(im.iterators, id)
				im.totalClosed++

				im.logger.Debug("Expired iterator closed",
					"id", id)
			}
		}
	}

	im.lastCleanup = now

	if len(expiredIDs) > 0 {
		im.logger.Info("Cleanup completed",
			"expired_count", len(expiredIDs),
			"active_count", len(im.iterators))
	}
}

// CreateRangeIterator 创建范围迭代器 Create range iterator
func (im *IteratorManager) CreateRangeIterator(ctx context.Context, txn *badger.Txn, tableName string, startKey, endKey []byte, options *IteratorOptions) (*Iterator, error) {
	if options == nil {
		options = &IteratorOptions{}
	}

	options.StartKey = startKey
	options.EndKey = endKey

	return im.CreateTableIterator(ctx, txn, tableName, options)
}

// CreateFilteredIterator 创建带过滤器的迭代器 Create filtered iterator
func (im *IteratorManager) CreateFilteredIterator(ctx context.Context, txn *badger.Txn, tableName string, filter FilterFunc, options *IteratorOptions) (*Iterator, error) {
	if options == nil {
		options = &IteratorOptions{}
	}

	options.Filter = filter

	return im.CreateTableIterator(ctx, txn, tableName, options)
}

// CreateLimitedIterator 创建限制数量的迭代器 Create limited iterator
func (im *IteratorManager) CreateLimitedIterator(ctx context.Context, txn *badger.Txn, tableName string, limit, offset int64, options *IteratorOptions) (*Iterator, error) {
	if options == nil {
		options = &IteratorOptions{}
	}

	options.Limit = limit
	options.Offset = offset

	return im.CreateTableIterator(ctx, txn, tableName, options)
}

// CreateBatchIterator 创建批量迭代器 Create batch iterator
func (im *IteratorManager) CreateBatchIterator(ctx context.Context, txn *badger.Txn, tableName string, batchSize int, options *IteratorOptions) (*BatchIterator, error) {
	if options == nil {
		options = &IteratorOptions{}
	}

	options.BufferSize = batchSize

	baseIter, err := im.CreateTableIterator(ctx, txn, tableName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create base iterator: %w", err)
	}

	return &BatchIterator{
		Iterator:  baseIter,
		batchSize: batchSize,
		logger:    im.logger,
	}, nil
}

// BatchIterator 批量迭代器 Batch iterator
type BatchIterator struct {
	*Iterator
	batchSize int
	logger    *zap.Logger
}

// NextBatch 获取下一批数据 Get next batch
func (bi *BatchIterator) NextBatch() ([]*IteratorItem, error) {
	if !bi.started {
		bi.SeekToFirst()
	}

	var batch []*IteratorItem
	count := 0

	for bi.Valid() && count < bi.batchSize {
		// 检查上下文 Check context
		select {
		case <-bi.ctx.Done():
			return batch, bi.ctx.Err()
		default:
		}

		item := bi.Item()
		if item != nil {
			batch = append(batch, item)
			count++
		}

		if !bi.Next() {
			break
		}
	}

	return batch, nil
}

// HasNextBatch 检查是否有下一批 Check if has next batch
func (bi *BatchIterator) HasNextBatch() bool {
	return bi.Valid()
}

// GetBatchSize 获取批次大小 Get batch size
func (bi *BatchIterator) GetBatchSize() int {
	return bi.batchSize
}

// SetBatchSize 设置批次大小 Set batch size
func (bi *BatchIterator) SetBatchSize(size int) {
	if size > 0 {
		bi.batchSize = size
		bi.options.BufferSize = size
	}
}

// Close 关闭迭代器管理器 Close iterator manager
func (im *IteratorManager) Close() error {
	var err error
	im.closeOnce.Do(func() {
		im.logger.Info("Closing iterator manager")

		// 发送关闭信号 Send close signal
		close(im.closed)

		// 关闭所有迭代器 Close all iterators
		if closeErr := im.CloseAllIterators(); closeErr != nil {
			im.logger.Warn("Error closing iterators", "error", closeErr)
			err = closeErr
		}

		im.logger.Info("Iterator manager closed",
			"total_created", im.totalCreated,
			"total_closed", im.totalClosed,
			"total_errors", im.totalErrors)
	})

	return err
}

// SetMaxIterators 设置最大迭代器数量 Set maximum iterator count
func (im *IteratorManager) SetMaxIterators(max int) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.maxIterators = max
}

// SetDefaultTimeout 设置默认超时时间 Set default timeout
func (im *IteratorManager) SetDefaultTimeout(timeout time.Duration) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.defaultTimeout = timeout
}

// SetCleanupInterval 设置清理间隔 Set cleanup interval
func (im *IteratorManager) SetCleanupInterval(interval time.Duration) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.cleanupInterval = interval
}

// EnableMetrics 启用指标收集 Enable metrics collection
func (im *IteratorManager) EnableMetrics(enable bool) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.enableMetrics = enable
}
