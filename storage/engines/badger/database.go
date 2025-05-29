// Package badger implements database-level operations for Badger storage engine
// badger包，实现Badger存储引擎的数据库级别操作
package badger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/turtacn/guocedb/common/errors"
	logging "github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types"
	"github.com/turtacn/guocedb/storage/engines/badger"
)

// DatabaseManager 数据库管理器 Database manager
type DatabaseManager struct {
	mu           sync.RWMutex         // 读写锁 Read-write lock
	basePath     string               // 基础路径 Base path
	databases    map[string]*Database // 数据库映射 Database mapping
	encoder      *KeyEncoder          // 键编码器 Key encoder
	decoder      *KeyDecoder          // 键解码器 Key decoder
	valueEncoder *ValueEncoder        // 值编码器 Value encoder
	logger       logging.Logger       // 日志记录器 Logger
	config       *DatabaseConfig      // 配置 Configuration
}

// Database 数据库实例 Database instance
type Database struct {
	name         string            // 数据库名 Database name
	path         string            // 存储路径 Storage path
	db           *badger.DB        // Badger实例 Badger instance
	metadata     *DatabaseMetadata // 元数据 Metadata
	encoder      *KeyEncoder       // 键编码器 Key encoder
	decoder      *KeyDecoder       // 键解码器 Key decoder
	valueEncoder *ValueEncoder     // 值编码器 Value encoder
	logger       logging.Logger    // 日志记录器 Logger
	mu           sync.RWMutex      // 读写锁 Read-write lock
	closed       bool              // 关闭状态 Closed status
}

// DatabaseMetadata 数据库元数据 Database metadata
type DatabaseMetadata struct {
	Name        string            `json:"name"`        // 数据库名 Database name
	CreatedAt   time.Time         `json:"created_at"`  // 创建时间 Creation time
	UpdatedAt   time.Time         `json:"updated_at"`  // 更新时间 Update time
	Version     int64             `json:"version"`     // 版本号 Version number
	TableCount  int               `json:"table_count"` // 表数量 Table count
	IndexCount  int               `json:"index_count"` // 索引数量 Index count
	Size        int64             `json:"size"`        // 数据库大小 Database size
	Description string            `json:"description"` // 描述 Description
	Properties  map[string]string `json:"properties"`  // 属性 Properties
	Statistics  *DatabaseStats    `json:"statistics"`  // 统计信息 Statistics
}

// DatabaseStats 数据库统计信息 Database statistics
type DatabaseStats struct {
	RowCount       int64     `json:"row_count"`       // 行数 Row count
	KeyCount       int64     `json:"key_count"`       // 键数 Key count
	ValueSize      int64     `json:"value_size"`      // 值大小 Value size
	LastCompaction time.Time `json:"last_compaction"` // 最近压缩时间 Last compaction time
	ReadOps        int64     `json:"read_ops"`        // 读操作数 Read operations
	WriteOps       int64     `json:"write_ops"`       // 写操作数 Write operations
}

// DatabaseConfig 数据库配置 Database configuration
type DatabaseConfig struct {
	MaxOpenDatabases int           `json:"max_open_databases"` // 最大打开数据库数 Max open databases
	SyncWrites       bool          `json:"sync_writes"`        // 同步写入 Sync writes
	Compression      bool          `json:"compression"`        // 压缩 Compression
	MemTableSize     int64         `json:"mem_table_size"`     // 内存表大小 MemTable size
	ValueLogSize     int64         `json:"value_log_size"`     // 值日志大小 Value log size
	GCInterval       time.Duration `json:"gc_interval"`        // GC间隔 GC interval
	BackupInterval   time.Duration `json:"backup_interval"`    // 备份间隔 Backup interval
}

// DefaultDatabaseConfig 默认数据库配置 Default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MaxOpenDatabases: 100,
		SyncWrites:       true,
		Compression:      true,
		MemTableSize:     64 << 20, // 64MB
		ValueLogSize:     1 << 30,  // 1GB
		GCInterval:       30 * time.Minute,
		BackupInterval:   24 * time.Hour,
	}
}

// NewDatabaseManager 创建数据库管理器 Create database manager
func NewDatabaseManager(basePath string, config *DatabaseConfig, logger logging.Logger) (*DatabaseManager, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	dm := &DatabaseManager{
		basePath:     basePath,
		databases:    make(map[string]*Database),
		encoder:      NewKeyEncoder(""),
		decoder:      NewKeyDecoder(),
		valueEncoder: NewValueEncoder(),
		logger:       logger,
		config:       config,
	}

	// 加载已存在的数据库 Load existing databases
	if err := dm.loadExistingDatabases(); err != nil {
		return nil, fmt.Errorf("failed to load existing databases: %w", err)
	}

	return dm, nil
}

// CreateDatabase 创建数据库 Create database
func (dm *DatabaseManager) CreateDatabase(ctx context.Context, name string, options map[string]interface{}) error {
	if err := dm.validateDatabaseName(name); err != nil {
		return fmt.Errorf("invalid database name: %w", err)
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查数据库是否已存在 Check if database already exists
	if _, exists := dm.databases[name]; exists {
		return errors.NewDatabaseExistsError(name)
	}

	// 检查最大数据库数限制 Check max databases limit
	if len(dm.databases) >= dm.config.MaxOpenDatabases {
		return fmt.Errorf("maximum number of databases (%d) reached", dm.config.MaxOpenDatabases)
	}

	// 创建数据库目录 Create database directory
	dbPath := filepath.Join(dm.basePath, name)

	// 创建Badger配置 Create Badger configuration
	badgerOpts := dm.createBadgerOptions(dbPath)

	// 打开Badger数据库 Open Badger database
	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %w", err)
	}

	// 创建数据库元数据 Create database metadata
	metadata := &DatabaseMetadata{
		Name:        name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
		TableCount:  0,
		IndexCount:  0,
		Size:        0,
		Description: getString(options, "description", ""),
		Properties:  getStringMap(options, "properties"),
		Statistics:  &DatabaseStats{},
	}

	// 创建数据库实例 Create database instance
	database := &Database{
		name:         name,
		path:         dbPath,
		db:           badgerDB,
		metadata:     metadata,
		encoder:      NewKeyEncoder(name),
		decoder:      dm.decoder,
		valueEncoder: dm.valueEncoder,
		logger:       dm.logger,
	}

	// 保存数据库元数据 Save database metadata
	if err := database.saveMetadata(); err != nil {
		badgerDB.Close()
		return fmt.Errorf("failed to save database metadata: %w", err)
	}

	// 添加到管理器 Add to manager
	dm.databases[name] = database

	dm.logger.Info("Database created successfully",
		"database", name,
		"path", dbPath)

	return nil
}

// DropDatabase 删除数据库 Drop database
func (dm *DatabaseManager) DropDatabase(ctx context.Context, name string, cascade bool) error {
	if err := dm.validateDatabaseName(name); err != nil {
		return fmt.Errorf("invalid database name: %w", err)
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查数据库是否存在 Check if database exists
	database, exists := dm.databases[name]
	if !exists {
		return errors.NewDatabaseNotFoundError(name)
	}

	// 检查是否有表存在 Check if tables exist
	if !cascade && database.metadata.TableCount > 0 {
		return fmt.Errorf("database %s contains tables, use CASCADE to force deletion", name)
	}

	// 关闭数据库 Close database
	if err := database.Close(); err != nil {
		dm.logger.Warn("Failed to close database during drop",
			"database", name,
			"error", err)
	}

	// 删除数据库文件 Delete database files
	if err := dm.removeDatabaseFiles(database.path); err != nil {
		return fmt.Errorf("failed to remove database files: %w", err)
	}

	// 从管理器中移除 Remove from manager
	delete(dm.databases, name)

	dm.logger.Info("Database dropped successfully",
		"database", name,
		"cascade", cascade)

	return nil
}

// GetDatabase 获取数据库 Get database
func (dm *DatabaseManager) GetDatabase(name string) (*Database, error) {
	if err := dm.validateDatabaseName(name); err != nil {
		return nil, fmt.Errorf("invalid database name: %w", err)
	}

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	database, exists := dm.databases[name]
	if !exists {
		return nil, errors.NewDatabaseNotFoundError(name)
	}

	return database, nil
}

// ListDatabases 列出数据库 List databases
func (dm *DatabaseManager) ListDatabases(ctx context.Context) ([]*DatabaseMetadata, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	databases := make([]*DatabaseMetadata, 0, len(dm.databases))
	for _, db := range dm.databases {
		db.mu.RLock()
		// 复制元数据 Copy metadata
		metadata := *db.metadata
		db.mu.RUnlock()
		databases = append(databases, &metadata)
	}

	return databases, nil
}

// DatabaseExists 检查数据库是否存在 Check if database exists
func (dm *DatabaseManager) DatabaseExists(name string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	_, exists := dm.databases[name]
	return exists
}

// GetDatabaseStats 获取数据库统计信息 Get database statistics
func (dm *DatabaseManager) GetDatabaseStats(name string) (*DatabaseStats, error) {
	database, err := dm.GetDatabase(name)
	if err != nil {
		return nil, err
	}

	return database.GetStats()
}

// RefreshStats 刷新统计信息 Refresh statistics
func (dm *DatabaseManager) RefreshStats(ctx context.Context) error {
	dm.mu.RLock()
	databases := make([]*Database, 0, len(dm.databases))
	for _, db := range dm.databases {
		databases = append(databases, db)
	}
	dm.mu.RUnlock()

	for _, db := range databases {
		if err := db.RefreshStats(ctx); err != nil {
			dm.logger.Warn("Failed to refresh database stats",
				"database", db.name,
				"error", err)
		}
	}

	return nil
}

// Close 关闭数据库管理器 Close database manager
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var errs []error
	for name, db := range dm.databases {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database %s: %w", name, err))
		}
	}

	dm.databases = make(map[string]*Database)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}

// loadExistingDatabases 加载已存在的数据库 Load existing databases
func (dm *DatabaseManager) loadExistingDatabases() error {
	// 扫描数据库目录 Scan database directories
	entries, err := os.ReadDir(dm.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建基础目录 Create base directory
			if err := os.MkdirAll(dm.basePath, 0755); err != nil {
				return fmt.Errorf("failed to create base directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dbName := entry.Name()
		if !dm.isValidDatabaseName(dbName) {
			continue
		}

		dbPath := filepath.Join(dm.basePath, dbName)

		// 尝试打开数据库 Try to open database
		if err := dm.openExistingDatabase(dbName, dbPath); err != nil {
			dm.logger.Warn("Failed to open existing database",
				"database", dbName,
				"path", dbPath,
				"error", err)
			continue
		}
	}

	return nil
}

// openExistingDatabase 打开已存在的数据库 Open existing database
func (dm *DatabaseManager) openExistingDatabase(name, path string) error {
	// 创建Badger配置 Create Badger configuration
	badgerOpts := dm.createBadgerOptions(path)

	// 打开Badger数据库 Open Badger database
	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %w", err)
	}

	// 创建数据库实例 Create database instance
	database := &Database{
		name:         name,
		path:         path,
		db:           badgerDB,
		encoder:      NewKeyEncoder(name),
		decoder:      dm.decoder,
		valueEncoder: dm.valueEncoder,
		logger:       dm.logger,
	}

	// 加载元数据 Load metadata
	if err := database.loadMetadata(); err != nil {
		badgerDB.Close()
		return fmt.Errorf("failed to load database metadata: %w", err)
	}

	// 添加到管理器 Add to manager
	dm.databases[name] = database

	dm.logger.Info("Existing database loaded successfully",
		"database", name,
		"path", path)

	return nil
}

// createBadgerOptions 创建Badger配置 Create Badger options
func (dm *DatabaseManager) createBadgerOptions(path string) badger.Options {
	opts := badger.DefaultOptions(path)

	// 基础配置 Basic configuration
	opts.SyncWrites = dm.config.SyncWrites
	opts.CompactL0OnClose = true
	opts.DetectConflicts = false

	// 内存配置 Memory configuration
	opts.MemTableSize = dm.config.MemTableSize
	opts.ValueLogFileSize = dm.config.ValueLogSize

	// 压缩配置 Compression configuration
	if dm.config.Compression {
		opts.Compression = options.ZSTD
	} else {
		opts.Compression = options.None
	}

	// 日志配置 Logging configuration
	opts.Logger = &badgerLogger{logger: dm.logger}

	return opts
}

// validateDatabaseName 验证数据库名 Validate database name
func (dm *DatabaseManager) validateDatabaseName(name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if len(name) > 64 {
		return fmt.Errorf("database name too long (max 64 characters)")
	}

	if !dm.isValidDatabaseName(name) {
		return fmt.Errorf("invalid database name format")
	}

	return nil
}

// isValidDatabaseName 检查数据库名是否有效 Check if database name is valid
func (dm *DatabaseManager) isValidDatabaseName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	// 只允许字母数字和下划线 Only allow alphanumeric and underscore
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_') {
			return false
		}
	}

	return true
}

// removeDatabaseFiles 删除数据库文件 Remove database files
func (dm *DatabaseManager) removeDatabaseFiles(path string) error {
	return os.RemoveAll(path)
}

// Database methods 数据库方法

// GetName 获取数据库名 Get database name
func (db *Database) GetName() string {
	return db.name
}

// GetPath 获取数据库路径 Get database path
func (db *Database) GetPath() string {
	return db.path
}

// GetBadgerDB 获取Badger数据库实例 Get Badger database instance
func (db *Database) GetBadgerDB() *badger.DB {
	return db.db
}

// GetMetadata 获取数据库元数据 Get database metadata
func (db *Database) GetMetadata() *DatabaseMetadata {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// 返回副本 Return copy
	metadata := *db.metadata
	return &metadata
}

// UpdateMetadata 更新数据库元数据 Update database metadata
func (db *Database) UpdateMetadata(updates map[string]interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 更新字段 Update fields
	if desc, ok := updates["description"]; ok {
		if descStr, ok := desc.(string); ok {
			db.metadata.Description = descStr
		}
	}

	if props, ok := updates["properties"]; ok {
		if propsMap, ok := props.(map[string]string); ok {
			if db.metadata.Properties == nil {
				db.metadata.Properties = make(map[string]string)
			}
			for k, v := range propsMap {
				db.metadata.Properties[k] = v
			}
		}
	}

	db.metadata.UpdatedAt = time.Now()
	db.metadata.Version++

	// 保存元数据 Save metadata
	return db.saveMetadata()
}

// GetStats 获取数据库统计信息 Get database statistics
func (db *Database) GetStats() (*DatabaseStats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, fmt.Errorf("database is closed")
	}

	// 获取Badger统计信息 Get Badger statistics
	lsm, vlog := db.db.Size()

	stats := &DatabaseStats{
		RowCount:       db.metadata.Statistics.RowCount,
		KeyCount:       db.metadata.Statistics.KeyCount,
		ValueSize:      lsm + vlog,
		LastCompaction: db.metadata.Statistics.LastCompaction,
		ReadOps:        db.metadata.Statistics.ReadOps,
		WriteOps:       db.metadata.Statistics.WriteOps,
	}

	return stats, nil
}

// RefreshStats 刷新统计信息 Refresh statistics
func (db *Database) RefreshStats(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	stats := &DatabaseStats{}

	// 统计键值对数量 Count key-value pairs
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		rowCount := int64(0)
		keyCount := int64(0)

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			keyType, err := db.decoder.DecodeKeyType(key)
			if err != nil {
				continue
			}

			keyCount++
			if keyType == KeyTypeRow {
				rowCount++
			}
		}

		stats.RowCount = rowCount
		stats.KeyCount = keyCount
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to count keys: %w", err)
	}

	// 获取大小信息 Get size information
	lsm, vlog := db.db.Size()
	stats.ValueSize = lsm + vlog

	// 更新元数据统计信息 Update metadata statistics
	db.metadata.Statistics = stats
	db.metadata.Size = stats.ValueSize
	db.metadata.UpdatedAt = time.Now()

	return db.saveMetadata()
}

// saveMetadata 保存元数据 Save metadata
func (db *Database) saveMetadata() error {
	// 编码元数据 Encode metadata
	data, err := db.valueEncoder.encoder.EncodeStruct(db.metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	// 生成元数据键 Generate metadata key
	key := db.encoder.EncodeConfigKey("__metadata__")

	// 保存到数据库 Save to database
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// loadMetadata 加载元数据 Load metadata
func (db *Database) loadMetadata() error {
	// 生成元数据键 Generate metadata key
	key := db.encoder.EncodeConfigKey("__metadata__")

	var metadata DatabaseMetadata

	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// 创建默认元数据 Create default metadata
				metadata = DatabaseMetadata{
					Name:        db.name,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Version:     1,
					TableCount:  0,
					IndexCount:  0,
					Size:        0,
					Description: "",
					Properties:  make(map[string]string),
					Statistics:  &DatabaseStats{},
				}
				return nil
			}
			return err
		}

		return item.Value(func(data []byte) error {
			return db.valueEncoder.encoder.DecodeStruct(data, &metadata)
		})
	})

	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	db.metadata = &metadata

	// 如果是新创建的元数据，保存它 If it's newly created metadata, save it
	if metadata.Version == 1 && metadata.CreatedAt.Equal(metadata.UpdatedAt) {
		return db.saveMetadata()
	}

	return nil
}

// IncrementTableCount 增加表数量 Increment table count
func (db *Database) IncrementTableCount() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.metadata.TableCount++
	db.metadata.UpdatedAt = time.Now()
	db.metadata.Version++

	return db.saveMetadata()
}

// DecrementTableCount 减少表数量 Decrement table count
func (db *Database) DecrementTableCount() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.metadata.TableCount > 0 {
		db.metadata.TableCount--
	}
	db.metadata.UpdatedAt = time.Now()
	db.metadata.Version++

	return db.saveMetadata()
}

// IncrementIndexCount 增加索引数量 Increment index count
func (db *Database) IncrementIndexCount() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.metadata.IndexCount++
	db.metadata.UpdatedAt = time.Now()
	db.metadata.Version++

	return db.saveMetadata()
}

// DecrementIndexCount 减少索引数量 Decrement index count
func (db *Database) DecrementIndexCount() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.metadata.IndexCount > 0 {
		db.metadata.IndexCount--
	}
	db.metadata.UpdatedAt = time.Now()
	db.metadata.Version++

	return db.saveMetadata()
}

// Compact 压缩数据库 Compact database
func (db *Database) Compact(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	// 运行压缩 Run compaction
	err := db.db.RunValueLogGC(0.5)
	if err != nil && err != badger.ErrNoRewrite {
		return fmt.Errorf("failed to run value log GC: %w", err)
	}

	// 更新统计信息 Update statistics
	db.metadata.Statistics.LastCompaction = time.Now()
	db.metadata.UpdatedAt = time.Now()

	return db.saveMetadata()
}

// Backup 备份数据库 Backup database
func (db *Database) Backup(ctx context.Context, writer io.Writer) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	// 创建备份 Create backup
	_, err := db.db.Backup(writer, 0)
	return err
}

// IsClosed 检查数据库是否关闭 Check if database is closed
func (db *Database) IsClosed() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.closed
}

// Close 关闭数据库 Close database
func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil
	}

	err := db.db.Close()
	db.closed = true

	if err != nil {
		return fmt.Errorf("failed to close badger database: %w", err)
	}

	db.logger.Info("Database closed successfully",
		"database", db.name)

	return nil
}

// 辅助函数 Helper functions

// getString 从选项中获取字符串 Get string from options
func getString(options map[string]interface{}, key, defaultValue string) string {
	if value, ok := options[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getStringMap 从选项中获取字符串映射 Get string map from options
func getStringMap(options map[string]interface{}, key string) map[string]string {
	if value, ok := options[key]; ok {
		if m, ok := value.(map[string]string); ok {
			return m
		}
		if m, ok := value.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
			return result
		}
	}
	return make(map[string]string)
}

// badgerLogger Badger日志适配器 Badger logger adapter
type badgerLogger struct {
	logger logging.Logger
}

func (bl *badgerLogger) Errorf(format string, args ...interface{}) {
	bl.logger.Error(fmt.Sprintf(format, args...))
}

func (bl *badgerLogger) Warningf(format string, args ...interface{}) {
	bl.logger.Warn(fmt.Sprintf(format, args...))
}

func (bl *badgerLogger) Infof(format string, args ...interface{}) {
	bl.logger.Info(fmt.Sprintf(format, args...))
}

func (bl *badgerLogger) Debugf(format string, args ...interface{}) {
	bl.logger.Debug(fmt.Sprintf(format, args...))
}
