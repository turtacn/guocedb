// Package interfaces 定义GuoceDB存储抽象层的核心接口
// Package interfaces defines core interfaces for GuoceDB storage abstraction layer
package interfaces

import (
	"context"
	"io"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types"
)

// StorageEngine 存储引擎接口
// StorageEngine storage engine interface
type StorageEngine interface {
	// 引擎管理 Engine management
	Name() string                                           // 获取引擎名称 Get engine name
	Version() string                                        // 获取引擎版本 Get engine version
	Open(path string, options map[string]interface{}) error // 打开存储引擎 Open storage engine
	Close() error                                           // 关闭存储引擎 Close storage engine
	IsOpen() bool                                           // 检查引擎是否已打开 Check if engine is open

	// 数据库管理 Database management
	CreateDatabase(ctx context.Context, name string, options *DatabaseOptions) error // 创建数据库 Create database
	DropDatabase(ctx context.Context, name string) error                             // 删除数据库 Drop database
	ListDatabases(ctx context.Context) ([]string, error)                             // 列出所有数据库 List all databases
	DatabaseExists(ctx context.Context, name string) (bool, error)                   // 检查数据库是否存在 Check if database exists

	// 表管理 Table management
	CreateTable(ctx context.Context, database string, schema *TableSchema) error         // 创建表 Create table
	DropTable(ctx context.Context, database, table string) error                         // 删除表 Drop table
	AlterTable(ctx context.Context, database, table string, changes *TableChanges) error // 修改表结构 Alter table structure
	ListTables(ctx context.Context, database string) ([]string, error)                   // 列出表 List tables
	TableExists(ctx context.Context, database, table string) (bool, error)               // 检查表是否存在 Check if table exists
	GetTableSchema(ctx context.Context, database, table string) (*TableSchema, error)    // 获取表结构 Get table schema

	// 数据操作 Data operations
	Insert(ctx context.Context, database, table string, rows []Row) error                                             // 插入数据 Insert data
	Update(ctx context.Context, database, table string, filter Filter, updates map[string]interface{}) (int64, error) // 更新数据 Update data
	Delete(ctx context.Context, database, table string, filter Filter) (int64, error)                                 // 删除数据 Delete data
	Select(ctx context.Context, database, table string, query *SelectQuery) (Iterator, error)                         // 查询数据 Select data
	Count(ctx context.Context, database, table string, filter Filter) (int64, error)                                  // 统计行数 Count rows

	// 索引管理 Index management
	CreateIndex(ctx context.Context, database, table string, index *IndexDefinition) error // 创建索引 Create index
	DropIndex(ctx context.Context, database, table, index string) error                    // 删除索引 Drop index
	ListIndexes(ctx context.Context, database, table string) ([]*IndexDefinition, error)   // 列出索引 List indexes
	IndexExists(ctx context.Context, database, table, index string) (bool, error)          // 检查索引是否存在 Check if index exists

	// 事务支持 Transaction support
	BeginTransaction(ctx context.Context, options *TransactionOptions) (Transaction, error) // 开始事务 Begin transaction
	SupportsTransactions() bool                                                             // 是否支持事务 Whether supports transactions

	// 统计信息 Statistics
	GetStats(ctx context.Context) (*EngineStats, error)                             // 获取引擎统计信息 Get engine statistics
	GetTableStats(ctx context.Context, database, table string) (*TableStats, error) // 获取表统计信息 Get table statistics

	// 维护操作 Maintenance operations
	Vacuum(ctx context.Context, database, table string) error                             // 清理表空间 Vacuum table space
	Analyze(ctx context.Context, database, table string) error                            // 分析表统计信息 Analyze table statistics
	CheckIntegrity(ctx context.Context, database, table string) (*IntegrityReport, error) // 检查数据完整性 Check data integrity

	// 备份恢复 Backup and restore
	Backup(ctx context.Context, writer io.Writer, options *BackupOptions) error   // 备份数据 Backup data
	Restore(ctx context.Context, reader io.Reader, options *RestoreOptions) error // 恢复数据 Restore data

	// 监控和诊断 Monitoring and diagnostics
	GetMetrics(ctx context.Context) (*EngineMetrics, error)        // 获取监控指标 Get monitoring metrics
	GetConnections(ctx context.Context) ([]*ConnectionInfo, error) // 获取连接信息 Get connection information
	KillConnection(ctx context.Context, connectionID uint64) error // 终止连接 Kill connection
}

// Transaction 事务接口
// Transaction transaction interface
type Transaction interface {
	// 事务基本操作 Basic transaction operations
	ID() uint64                         // 获取事务ID Get transaction ID
	Commit(ctx context.Context) error   // 提交事务 Commit transaction
	Rollback(ctx context.Context) error // 回滚事务 Rollback transaction
	IsActive() bool                     // 检查事务是否活跃 Check if transaction is active
	IsReadOnly() bool                   // 检查是否只读事务 Check if read-only transaction

	// 数据操作 Data operations (within transaction)
	Insert(ctx context.Context, database, table string, rows []Row) error                                             // 插入数据 Insert data
	Update(ctx context.Context, database, table string, filter Filter, updates map[string]interface{}) (int64, error) // 更新数据 Update data
	Delete(ctx context.Context, database, table string, filter Filter) (int64, error)                                 // 删除数据 Delete data
	Select(ctx context.Context, database, table string, query *SelectQuery) (Iterator, error)                         // 查询数据 Select data

	// 锁管理 Lock management
	LockTable(ctx context.Context, database, table string, lockType LockType) error            // 锁表 Lock table
	LockRow(ctx context.Context, database, table string, rowID RowID, lockType LockType) error // 锁行 Lock row
	ReleaseLock(ctx context.Context, lockID uint64) error                                      // 释放锁 Release lock
	GetLocks(ctx context.Context) ([]*LockInfo, error)                                         // 获取锁信息 Get lock information

	// 保存点 Savepoints
	CreateSavepoint(ctx context.Context, name string) (*Savepoint, error) // 创建保存点 Create savepoint
	ReleaseSavepoint(ctx context.Context, savepoint *Savepoint) error     // 释放保存点 Release savepoint
	RollbackToSavepoint(ctx context.Context, savepoint *Savepoint) error  // 回滚到保存点 Rollback to savepoint

	// 事务信息 Transaction information
	GetStartTime() time.Time           // 获取事务开始时间 Get transaction start time
	GetIsolationLevel() IsolationLevel // 获取隔离级别 Get isolation level
	GetStats() *TransactionStats       // 获取事务统计信息 Get transaction statistics
}

// Iterator 数据迭代器接口
// Iterator data iterator interface
type Iterator interface {
	// 迭代控制 Iteration control
	Next() bool    // 移动到下一行 Move to next row
	HasNext() bool // 检查是否有下一行 Check if has next row
	Reset() error  // 重置迭代器 Reset iterator
	Close() error  // 关闭迭代器 Close iterator

	// 数据访问 Data access
	Row() (Row, error)                        // 获取当前行 Get current row
	Value(column string) (interface{}, error) // 获取指定列的值 Get value of specified column
	Values() (map[string]interface{}, error)  // 获取所有列的值 Get values of all columns

	// 位置控制 Position control
	Seek(key interface{}) error // 定位到指定键 Seek to specified key
	SeekToFirst() error         // 定位到第一行 Seek to first row
	SeekToLast() error          // 定位到最后一行 Seek to last row
	Position() (int64, error)   // 获取当前位置 Get current position

	// 元数据 Metadata
	Schema() *TableSchema  // 获取表结构 Get table schema
	Count() (int64, error) // 获取总行数 Get total row count

	// 错误处理 Error handling
	Error() error // 获取迭代过程中的错误 Get iteration error
}

// Schema 表结构接口
// Schema table schema interface
type Schema interface {
	// 基本信息 Basic information
	Name() string         // 获取表名 Get table name
	Database() string     // 获取数据库名 Get database name
	Version() int64       // 获取版本号 Get version number
	CreatedAt() time.Time // 获取创建时间 Get creation time
	UpdatedAt() time.Time // 获取更新时间 Get update time

	// 列管理 Column management
	AddColumn(column *ColumnDefinition) error                 // 添加列 Add column
	DropColumn(name string) error                             // 删除列 Drop column
	ModifyColumn(name string, column *ColumnDefinition) error // 修改列 Modify column
	RenameColumn(oldName, newName string) error               // 重命名列 Rename column
	GetColumn(name string) (*ColumnDefinition, error)         // 获取列定义 Get column definition
	GetColumns() []*ColumnDefinition                          // 获取所有列 Get all columns
	HasColumn(name string) bool                               // 检查列是否存在 Check if column exists

	// 约束管理 Constraint management
	AddConstraint(constraint *ConstraintDefinition) error     // 添加约束 Add constraint
	DropConstraint(name string) error                         // 删除约束 Drop constraint
	GetConstraint(name string) (*ConstraintDefinition, error) // 获取约束定义 Get constraint definition
	GetConstraints() []*ConstraintDefinition                  // 获取所有约束 Get all constraints
	HasConstraint(name string) bool                           // 检查约束是否存在 Check if constraint exists

	// 索引管理 Index management
	AddIndex(index *IndexDefinition) error          // 添加索引 Add index
	DropIndex(name string) error                    // 删除索引 Drop index
	GetIndex(name string) (*IndexDefinition, error) // 获取索引定义 Get index definition
	GetIndexes() []*IndexDefinition                 // 获取所有索引 Get all indexes
	HasIndex(name string) bool                      // 检查索引是否存在 Check if index exists

	// 验证 Validation
	Validate() error           // 验证表结构 Validate table schema
	ValidateRow(row Row) error // 验证行数据 Validate row data

	// 序列化 Serialization
	ToJSON() ([]byte, error)    // 序列化为JSON Serialize to JSON
	FromJSON(data []byte) error // 从JSON反序列化 Deserialize from JSON
	ToSQL() string              // 转换为SQL语句 Convert to SQL statement

	// 克隆 Clone
	Clone() Schema // 克隆表结构 Clone table schema
}

// Catalog 目录接口（元数据管理）
// Catalog interface (metadata management)
type Catalog interface {
	// 数据库管理 Database management
	CreateDatabase(ctx context.Context, name string, options *DatabaseOptions) error // 创建数据库 Create database
	DropDatabase(ctx context.Context, name string) error                             // 删除数据库 Drop database
	GetDatabase(ctx context.Context, name string) (*DatabaseInfo, error)             // 获取数据库信息 Get database information
	ListDatabases(ctx context.Context) ([]*DatabaseInfo, error)                      // 列出所有数据库 List all databases
	DatabaseExists(ctx context.Context, name string) (bool, error)                   // 检查数据库是否存在 Check if database exists

	// 表管理 Table management
	CreateTable(ctx context.Context, database string, schema *TableSchema) error         // 创建表 Create table
	DropTable(ctx context.Context, database, table string) error                         // 删除表 Drop table
	AlterTable(ctx context.Context, database, table string, changes *TableChanges) error // 修改表 Alter table
	GetTable(ctx context.Context, database, table string) (*TableInfo, error)            // 获取表信息 Get table information
	ListTables(ctx context.Context, database string) ([]*TableInfo, error)               // 列出表 List tables
	TableExists(ctx context.Context, database, table string) (bool, error)               // 检查表是否存在 Check if table exists

	// 列管理 Column management
	AddColumn(ctx context.Context, database, table string, column *ColumnDefinition) error    // 添加列 Add column
	DropColumn(ctx context.Context, database, table, column string) error                     // 删除列 Drop column
	ModifyColumn(ctx context.Context, database, table string, column *ColumnDefinition) error // 修改列 Modify column
	GetColumn(ctx context.Context, database, table, column string) (*ColumnDefinition, error) // 获取列信息 Get column information
	ListColumns(ctx context.Context, database, table string) ([]*ColumnDefinition, error)     // 列出列 List columns

	// 索引管理 Index management
	CreateIndex(ctx context.Context, database, table string, index *IndexDefinition) error // 创建索引 Create index
	DropIndex(ctx context.Context, database, table, index string) error                    // 删除索引 Drop index
	GetIndex(ctx context.Context, database, table, index string) (*IndexDefinition, error) // 获取索引信息 Get index information
	ListIndexes(ctx context.Context, database, table string) ([]*IndexDefinition, error)   // 列出索引 List indexes

	// 约束管理 Constraint management
	AddConstraint(ctx context.Context, database, table string, constraint *ConstraintDefinition) error    // 添加约束 Add constraint
	DropConstraint(ctx context.Context, database, table, constraint string) error                         // 删除约束 Drop constraint
	GetConstraint(ctx context.Context, database, table, constraint string) (*ConstraintDefinition, error) // 获取约束信息 Get constraint information
	ListConstraints(ctx context.Context, database, table string) ([]*ConstraintDefinition, error)         // 列出约束 List constraints

	// 统计信息 Statistics
	UpdateStats(ctx context.Context, database, table string) error             // 更新统计信息 Update statistics
	GetStats(ctx context.Context, database, table string) (*TableStats, error) // 获取统计信息 Get statistics

	// 权限管理 Permission management
	GrantPermission(ctx context.Context, user, database, table string, permissions []Permission) error      // 授权 Grant permission
	RevokePermission(ctx context.Context, user, database, table string, permissions []Permission) error     // 撤销权限 Revoke permission
	CheckPermission(ctx context.Context, user, database, table string, permission Permission) (bool, error) // 检查权限 Check permission
	ListPermissions(ctx context.Context, user string) ([]*PermissionInfo, error)                            // 列出权限 List permissions

	// 版本管理 Version management
	GetVersion(ctx context.Context) (int64, error)                               // 获取目录版本 Get catalog version
	GetSchemaVersion(ctx context.Context, database, table string) (int64, error) // 获取表结构版本 Get schema version

	// 备份恢复 Backup and restore
	ExportMetadata(ctx context.Context, writer io.Writer) error // 导出元数据 Export metadata
	ImportMetadata(ctx context.Context, reader io.Reader) error // 导入元数据 Import metadata

	// 监听器 Listeners
	AddChangeListener(listener CatalogChangeListener)    // 添加变更监听器 Add change listener
	RemoveChangeListener(listener CatalogChangeListener) // 移除变更监听器 Remove change listener
}

// CatalogChangeListener 目录变更监听器
// CatalogChangeListener catalog change listener
type CatalogChangeListener interface {
	OnDatabaseCreated(database string)                            // 数据库创建事件 Database created event
	OnDatabaseDropped(database string)                            // 数据库删除事件 Database dropped event
	OnTableCreated(database, table string)                        // 表创建事件 Table created event
	OnTableDropped(database, table string)                        // 表删除事件 Table dropped event
	OnTableAltered(database, table string, changes *TableChanges) // 表修改事件 Table altered event
	OnIndexCreated(database, table, index string)                 // 索引创建事件 Index created event
	OnIndexDropped(database, table, index string)                 // 索引删除事件 Index dropped event
}

// 数据类型定义 Data type definitions

// Row 行数据类型
// Row row data type
type Row map[string]interface{}

// RowID 行标识符
// RowID row identifier
type RowID interface {
	String() string          // 转换为字符串 Convert to string
	Bytes() []byte           // 转换为字节数组 Convert to byte array
	Compare(other RowID) int // 比较 Compare
	IsValid() bool           // 检查是否有效 Check if valid
}

// Filter 过滤条件
// Filter filter condition
type Filter interface {
	Match(row Row) (bool, error) // 匹配行数据 Match row data
	ToSQL() string               // 转换为SQL条件 Convert to SQL condition
	Optimize() Filter            // 优化过滤条件 Optimize filter condition
	GetColumns() []string        // 获取涉及的列 Get involved columns
}

// SelectQuery 查询对象
// SelectQuery query object
type SelectQuery struct {
	Columns []string          `json:"columns"`  // 查询列 Query columns
	Filter  Filter            `json:"filter"`   // 过滤条件 Filter condition
	OrderBy []*OrderBy        `json:"order_by"` // 排序条件 Order by condition
	GroupBy []string          `json:"group_by"` // 分组条件 Group by condition
	Having  Filter            `json:"having"`   // Having条件 Having condition
	Limit   int64             `json:"limit"`    // 限制数量 Limit count
	Offset  int64             `json:"offset"`   // 偏移量 Offset
	Hints   map[string]string `json:"hints"`    // 查询提示 Query hints
}

// OrderBy 排序条件
// OrderBy order by condition
type OrderBy struct {
	Column    string         `json:"column"`    // 排序列 Order column
	Direction OrderDirection `json:"direction"` // 排序方向 Order direction
}

// OrderDirection 排序方向
// OrderDirection order direction
type OrderDirection int

const (
	OrderAsc  OrderDirection = iota // 升序 Ascending
	OrderDesc                       // 降序 Descending
)

// TableSchema 表结构定义
// TableSchema table schema definition
type TableSchema struct {
	Name        string                  `json:"name"`        // 表名 Table name
	Database    string                  `json:"database"`    // 数据库名 Database name
	Columns     []*ColumnDefinition     `json:"columns"`     // 列定义 Column definitions
	Constraints []*ConstraintDefinition `json:"constraints"` // 约束定义 Constraint definitions
	Indexes     []*IndexDefinition      `json:"indexes"`     // 索引定义 Index definitions
	Options     map[string]interface{}  `json:"options"`     // 表选项 Table options
	Version     int64                   `json:"version"`     // 版本号 Version number
	CreatedAt   time.Time               `json:"created_at"`  // 创建时间 Creation time
	UpdatedAt   time.Time               `json:"updated_at"`  // 更新时间 Update time
}

// ColumnDefinition 列定义
// ColumnDefinition column definition
type ColumnDefinition struct {
	Name          string                 `json:"name"`           // 列名 Column name
	Type          types.DataType         `json:"type"`           // 数据类型 Data type
	Length        int                    `json:"length"`         // 长度 Length
	Precision     int                    `json:"precision"`      // 精度 Precision
	Scale         int                    `json:"scale"`          // 标度 Scale
	Nullable      bool                   `json:"nullable"`       // 是否可空 Whether nullable
	DefaultValue  interface{}            `json:"default_value"`  // 默认值 Default value
	AutoIncrement bool                   `json:"auto_increment"` // 是否自增 Whether auto increment
	Primary       bool                   `json:"primary"`        // 是否主键 Whether primary key
	Unique        bool                   `json:"unique"`         // 是否唯一 Whether unique
	Comment       string                 `json:"comment"`        // 注释 Comment
	Options       map[string]interface{} `json:"options"`        // 列选项 Column options
}

// ConstraintDefinition 约束定义
// ConstraintDefinition constraint definition
type ConstraintDefinition struct {
	Name              string          `json:"name"`               // 约束名 Constraint name
	Type              ConstraintType  `json:"type"`               // 约束类型 Constraint type
	Columns           []string        `json:"columns"`            // 涉及的列 Involved columns
	RefTable          string          `json:"ref_table"`          // 引用表 Reference table
	RefColumns        []string        `json:"ref_columns"`        // 引用列 Reference columns
	OnUpdate          ReferenceAction `json:"on_update"`          // 更新动作 Update action
	OnDelete          ReferenceAction `json:"on_delete"`          // 删除动作 Delete action
	CheckExpr         string          `json:"check_expr"`         // 检查表达式 Check expression
	Deferrable        bool            `json:"deferrable"`         // 是否可延迟 Whether deferrable
	InitiallyDeferred bool            `json:"initially_deferred"` // 初始是否延迟 Initially deferred
}

// ConstraintType 约束类型
// ConstraintType constraint type
type ConstraintType int

const (
	ConstraintPrimaryKey ConstraintType = iota // 主键约束 Primary key constraint
	ConstraintForeignKey                       // 外键约束 Foreign key constraint
	ConstraintUnique                           // 唯一约束 Unique constraint
	ConstraintCheck                            // 检查约束 Check constraint
	ConstraintNotNull                          // 非空约束 Not null constraint
)

// ReferenceAction 引用动作
// ReferenceAction reference action
type ReferenceAction int

const (
	ReferenceNoAction   ReferenceAction = iota // 无动作 No action
	ReferenceRestrict                          // 限制 Restrict
	ReferenceCascade                           // 级联 Cascade
	ReferenceSetNull                           // 设为空 Set null
	ReferenceSetDefault                        // 设为默认值 Set default
)

// IndexDefinition 索引定义
// IndexDefinition index definition
type IndexDefinition struct {
	Name      string                 `json:"name"`       // 索引名 Index name
	Table     string                 `json:"table"`      // 表名 Table name
	Columns   []*IndexColumn         `json:"columns"`    // 索引列 Index columns
	Type      IndexType              `json:"type"`       // 索引类型 Index type
	Unique    bool                   `json:"unique"`     // 是否唯一 Whether unique
	Partial   string                 `json:"partial"`    // 部分索引条件 Partial index condition
	Method    string                 `json:"method"`     // 索引方法 Index method
	Options   map[string]interface{} `json:"options"`    // 索引选项 Index options
	CreatedAt time.Time              `json:"created_at"` // 创建时间 Creation time
}

// IndexColumn 索引列
// IndexColumn index column
type IndexColumn struct {
	Name      string         `json:"name"`      // 列名 Column name
	Direction OrderDirection `json:"direction"` // 排序方向 Sort direction
	Length    int            `json:"length"`    // 索引长度 Index length
}

// IndexType 索引类型
// IndexType index type
type IndexType int

const (
	IndexBTree  IndexType = iota // B-Tree索引 B-Tree index
	IndexHash                    // 哈希索引 Hash index
	IndexGiST                    // GiST索引 GiST index
	IndexGIN                     // GIN索引 GIN index
	IndexSPGiST                  // SP-GiST索引 SP-GiST index
	IndexBRIN                    // BRIN索引 BRIN index
)

// TableChanges 表结构变更
// TableChanges table structure changes
type TableChanges struct {
	AddColumns      []*ColumnDefinition     `json:"add_columns"`      // 添加的列 Added columns
	DropColumns     []string                `json:"drop_columns"`     // 删除的列 Dropped columns
	ModifyColumns   []*ColumnDefinition     `json:"modify_columns"`   // 修改的列 Modified columns
	RenameColumns   map[string]string       `json:"rename_columns"`   // 重命名的列 Renamed columns
	AddConstraints  []*ConstraintDefinition `json:"add_constraints"`  // 添加的约束 Added constraints
	DropConstraints []string                `json:"drop_constraints"` // 删除的约束 Dropped constraints
	AddIndexes      []*IndexDefinition      `json:"add_indexes"`      // 添加的索引 Added indexes
	DropIndexes     []string                `json:"drop_indexes"`     // 删除的索引 Dropped indexes
	RenameTable     string                  `json:"rename_table"`     // 新表名 New table name
	Options         map[string]interface{}  `json:"options"`          // 新选项 New options
}

// DatabaseOptions 数据库选项
// DatabaseOptions database options
type DatabaseOptions struct {
	Charset  string                 `json:"charset"`  // 字符集 Character set
	Collate  string                 `json:"collate"`  // 排序规则 Collation
	Owner    string                 `json:"owner"`    // 所有者 Owner
	Template string                 `json:"template"` // 模板 Template
	Encoding string                 `json:"encoding"` // 编码 Encoding
	Options  map[string]interface{} `json:"options"`  // 其他选项 Other options
}

// TransactionOptions 事务选项
// TransactionOptions transaction options
type TransactionOptions struct {
	Isolation IsolationLevel         `json:"isolation"` // 隔离级别 Isolation level
	ReadOnly  bool                   `json:"read_only"` // 是否只读 Whether read-only
	Timeout   time.Duration          `json:"timeout"`   // 超时时间 Timeout
	Options   map[string]interface{} `json:"options"`   // 其他选项 Other options
}

// IsolationLevel 隔离级别
// IsolationLevel isolation level
type IsolationLevel int

const (
	IsolationReadUncommitted IsolationLevel = iota // 读未提交 Read uncommitted
	IsolationReadCommitted                         // 读已提交 Read committed
	IsolationRepeatableRead                        // 可重复读 Repeatable read
	IsolationSerializable                          // 串行化 Serializable
)

// LockType 锁类型
// LockType lock type
type LockType int

const (
	LockShared    LockType = iota // 共享锁 Shared lock
	LockExclusive                 // 排他锁 Exclusive lock
	LockUpdate                    // 更新锁 Update lock
	LockIntention                 // 意向锁 Intention lock
)

// LockInfo 锁信息
// LockInfo lock information
type LockInfo struct {
	ID            uint64        `json:"id"`             // 锁ID Lock ID
	Type          LockType      `json:"type"`           // 锁类型 Lock type
	Database      string        `json:"database"`       // 数据库 Database
	Table         string        `json:"table"`          // 表 Table
	RowID         RowID         `json:"row_id"`         // 行ID Row ID
	TransactionID uint64        `json:"transaction_id"` // 事务ID Transaction ID
	AcquiredAt    time.Time     `json:"acquired_at"`    // 获得时间 Acquired time
	WaitingTime   time.Duration `json:"waiting_time"`   // 等待时间 Waiting time
}

// Savepoint 保存点
// Savepoint savepoint
type Savepoint struct {
	ID            uint64    `json:"id"`             // 保存点ID Savepoint ID
	Name          string    `json:"name"`           // 保存点名称 Savepoint name
	TransactionID uint64    `json:"transaction_id"` // 事务ID Transaction ID
	CreatedAt     time.Time `json:"created_at"`     // 创建时间 Creation time
}

// Permission 权限类型
// Permission permission type
type Permission int

const (
	PermissionSelect Permission = iota // 查询权限 Select permission
	PermissionInsert                   // 插入权限 Insert permission
	PermissionUpdate                   // 更新权限 Update permission
	PermissionDelete                   // 删除权限 Delete permission
	PermissionCreate                   // 创建权限 Create permission
	PermissionDrop                     // 删除权限 Drop permission
	PermissionAlter                    // 修改权限 Alter permission
	PermissionIndex                    // 索引权限 Index permission
	PermissionAdmin                    // 管理权限 Admin permission
)

// 信息类型定义 Information type definitions

// DatabaseInfo 数据库信息
// DatabaseInfo database information
type DatabaseInfo struct {
	Name       string                 `json:"name"`        // 数据库名 Database name
	Owner      string                 `json:"owner"`       // 所有者 Owner
	Charset    string                 `json:"charset"`     // 字符集 Character set
	Collate    string                 `json:"collate"`     // 排序规则 Collation
	Size       int64                  `json:"size"`        // 大小 Size
	TableCount int                    `json:"table_count"` // 表数量 Table count
	CreatedAt  time.Time              `json:"created_at"`  // 创建时间 Creation time
	UpdatedAt  time.Time              `json:"updated_at"`  // 更新时间 Update time
	Options    map[string]interface{} `json:"options"`     // 选项 Options
}

// TableInfo 表信息
// TableInfo table information
type TableInfo struct {
	Name          string                 `json:"name"`           // 表名 Table name
	Database      string                 `json:"database"`       // 数据库名 Database name
	Schema        *TableSchema           `json:"schema"`         // 表结构 Table schema
	RowCount      int64                  `json:"row_count"`      // 行数 Row count
	Size          int64                  `json:"size"`           // 大小 Size
	IndexSize     int64                  `json:"index_size"`     // 索引大小 Index size
	AutoIncrement int64                  `json:"auto_increment"` // 自增值 Auto increment value
	Charset       string                 `json:"charset"`        // 字符集 Character set
	Collate       string                 `json:"collate"`        // 排序规则 Collation
	Engine        string                 `json:"engine"`         // 存储引擎 Storage engine
	Comment       string                 `json:"comment"`        // 注释 Comment
	CreatedAt     time.Time              `json:"created_at"`     // 创建时间 Creation time
	UpdatedAt     time.Time              `json:"updated_at"`     // 更新时间 Update time
	Options       map[string]interface{} `json:"options"`        // 选项 Options
}

// PermissionInfo 权限信息
// PermissionInfo permission information
type PermissionInfo struct {
	User       string     `json:"user"`       // 用户 User
	Database   string     `json:"database"`   // 数据库 Database
	Table      string     `json:"table"`      // 表 Table
	Permission Permission `json:"permission"` // 权限 Permission
	GrantedBy  string     `json:"granted_by"` // 授权者 Granted by
	GrantedAt  time.Time  `json:"granted_at"` // 授权时间 Granted time
}

// 统计信息类型定义 Statistics type definitions

// EngineStats 引擎统计信息
// EngineStats engine statistics
type EngineStats struct {
	Name              string                 `json:"name"`               // 引擎名称 Engine name
	Version           string                 `json:"version"`            // 版本 Version
	Uptime            time.Duration          `json:"uptime"`             // 运行时间 Uptime
	DatabaseCount     int                    `json:"database_count"`     // 数据库数量 Database count
	TableCount        int                    `json:"table_count"`        // 表数量 Table count
	ConnectionCount   int                    `json:"connection_count"`   // 连接数量 Connection count
	ActiveTransaction int                    `json:"active_transaction"` // 活跃事务数 Active transaction count
	TotalSize         int64                  `json:"total_size"`         // 总大小 Total size
	CacheHitRate      float64                `json:"cache_hit_rate"`     // 缓存命中率 Cache hit rate
	QPS               float64                `json:"qps"`                // 每秒查询数 Queries per second
	TPS               float64                `json:"tps"`                // 每秒事务数 Transactions per second
	AvgResponseTime   time.Duration          `json:"avg_response_time"`  // 平均响应时间 Average response time
	SlowQueryCount    int64                  `json:"slow_query_count"`   // 慢查询数量 Slow query count
	ErrorCount        int64                  `json:"error_count"`        // 错误数量 Error count
	Metrics           map[string]interface{} `json:"metrics"`            // 其他指标 Other metrics
}

// TableStats 表统计信息
// TableStats table statistics
type TableStats struct {
	Database     string           `json:"database"`       // 数据库名 Database name
	Table        string           `json:"table"`          // 表名 Table name
	RowCount     int64            `json:"row_count"`      // 行数 Row count
	Size         int64            `json:"size"`           // 数据大小 Data size
	IndexSize    int64            `json:"index_size"`     // 索引大小 Index size
	AvgRowLength int64            `json:"avg_row_length"` // 平均行长度 Average row length
	SelectCount  int64            `json:"select_count"`   // 查询次数 Select count
	InsertCount  int64            `json:"insert_count"`   // 插入次数 Insert count
	UpdateCount  int64            `json:"update_count"`   // 更新次数 Update count
	DeleteCount  int64            `json:"delete_count"`   // 删除次数 Delete count
	LastAccessed time.Time        `json:"last_accessed"`  // 最后访问时间 Last accessed time
	LastUpdated  time.Time        `json:"last_updated"`   // 最后更新时间 Last updated time
	ColumnStats  []*ColumnStats   `json:"column_stats"`   // 列统计信息 Column statistics
	IndexStats   []*IndexStats    `json:"index_stats"`    // 索引统计信息 Index statistics
	Cardinality  map[string]int64 `json:"cardinality"`    // 基数统计 Cardinality statistics
}

// ColumnStats 列统计信息
// ColumnStats column statistics
type ColumnStats struct {
	Name          string            `json:"name"`           // 列名 Column name
	Type          types.DataType    `json:"type"`           // 数据类型 Data type
	NullCount     int64             `json:"null_count"`     // 空值数量 Null count
	DistinctCount int64             `json:"distinct_count"` // 不同值数量 Distinct count
	MinValue      interface{}       `json:"min_value"`      // 最小值 Minimum value
	MaxValue      interface{}       `json:"max_value"`      // 最大值 Maximum value
	AvgLength     float64           `json:"avg_length"`     // 平均长度 Average length
	Histogram     []HistogramBucket `json:"histogram"`      // 直方图 Histogram
	TopValues     []interface{}     `json:"top_values"`     // 高频值 Top values
	UpdatedAt     time.Time         `json:"updated_at"`     // 更新时间 Updated time
}

// IndexStats 索引统计信息
// IndexStats index statistics
type IndexStats struct {
	Name        string    `json:"name"`         // 索引名 Index name
	Type        IndexType `json:"type"`         // 索引类型 Index type
	Size        int64     `json:"size"`         // 索引大小 Index size
	Pages       int64     `json:"pages"`        // 页数 Page count
	Cardinality int64     `json:"cardinality"`  // 基数 Cardinality
	SelectCount int64     `json:"select_count"` // 使用次数 Usage count
	LastUsed    time.Time `json:"last_used"`    // 最后使用时间 Last used time
	CreatedAt   time.Time `json:"created_at"`   // 创建时间 Creation time
}

// HistogramBucket 直方图桶
// HistogramBucket histogram bucket
type HistogramBucket struct {
	LowerBound interface{} `json:"lower_bound"` // 下界 Lower bound
	UpperBound interface{} `json:"upper_bound"` // 上界 Upper bound
	Count      int64       `json:"count"`       // 数量 Count
	Frequency  float64     `json:"frequency"`   // 频率 Frequency
}

// TransactionStats 事务统计信息
// TransactionStats transaction statistics
type TransactionStats struct {
	ID             uint64         `json:"id"`              // 事务ID Transaction ID
	StartTime      time.Time      `json:"start_time"`      // 开始时间 Start time
	Duration       time.Duration  `json:"duration"`        // 持续时间 Duration
	IsolationLevel IsolationLevel `json:"isolation_level"` // 隔离级别 Isolation level
	ReadOnly       bool           `json:"read_only"`       // 是否只读 Whether read-only
	RowsRead       int64          `json:"rows_read"`       // 读取行数 Rows read
	RowsInserted   int64          `json:"rows_inserted"`   // 插入行数 Rows inserted
	RowsUpdated    int64          `json:"rows_updated"`    // 更新行数 Rows updated
	RowsDeleted    int64          `json:"rows_deleted"`    // 删除行数 Rows deleted
	LocksAcquired  int            `json:"locks_acquired"`  // 获得锁数 Locks acquired
	SavepointCount int            `json:"savepoint_count"` // 保存点数 Savepoint count
	QueryCount     int            `json:"query_count"`     // 查询次数 Query count
	BytesRead      int64          `json:"bytes_read"`      // 读取字节数 Bytes read
	BytesWritten   int64          `json:"bytes_written"`   // 写入字节数 Bytes written
}

// IntegrityReport 完整性报告
// IntegrityReport integrity report
type IntegrityReport struct {
	Database  string              `json:"database"`   // 数据库名 Database name
	Table     string              `json:"table"`      // 表名 Table name
	CheckedAt time.Time           `json:"checked_at"` // 检查时间 Checked time
	Status    IntegrityStatus     `json:"status"`     // 状态 Status
	Errors    []*IntegrityError   `json:"errors"`     // 错误列表 Error list
	Warnings  []*IntegrityWarning `json:"warnings"`   // 警告列表 Warning list
	Summary   *IntegritySummary   `json:"summary"`    // 摘要 Summary
}

// IntegrityStatus 完整性状态
// IntegrityStatus integrity status
type IntegrityStatus int

const (
	IntegrityOK        IntegrityStatus = iota // 正常 OK
	IntegrityWarning                          // 警告 Warning
	IntegrityError                            // 错误 Error
	IntegrityCorrupted                        // 损坏 Corrupted
)

// IntegrityError 完整性错误
// IntegrityError integrity error
type IntegrityError struct {
	Type       string    `json:"type"`        // 错误类型 Error type
	Message    string    `json:"message"`     // 错误消息 Error message
	Table      string    `json:"table"`       // 表名 Table name
	Column     string    `json:"column"`      // 列名 Column name
	RowID      RowID     `json:"row_id"`      // 行ID Row ID
	Severity   string    `json:"severity"`    // 严重程度 Severity
	DetectedAt time.Time `json:"detected_at"` // 检测时间 Detected time
}

// IntegrityWarning 完整性警告
// IntegrityWarning integrity warning
type IntegrityWarning struct {
	Type       string    `json:"type"`        // 警告类型 Warning type
	Message    string    `json:"message"`     // 警告消息 Warning message
	Table      string    `json:"table"`       // 表名 Table name
	Column     string    `json:"column"`      // 列名 Column name
	Suggestion string    `json:"suggestion"`  // 建议 Suggestion
	DetectedAt time.Time `json:"detected_at"` // 检测时间 Detected time
}

// IntegritySummary 完整性摘要
// IntegritySummary integrity summary
type IntegritySummary struct {
	TablesChecked     int           `json:"tables_checked"`     // 检查的表数 Tables checked
	RowsChecked       int64         `json:"rows_checked"`       // 检查的行数 Rows checked
	ErrorCount        int           `json:"error_count"`        // 错误数量 Error count
	WarningCount      int           `json:"warning_count"`      // 警告数量 Warning count
	Duration          time.Duration `json:"duration"`           // 检查耗时 Check duration
	RecommendedAction string        `json:"recommended_action"` // 推荐操作 Recommended action
}

// 备份恢复类型定义 Backup and restore type definitions

// BackupOptions 备份选项
// BackupOptions backup options
type BackupOptions struct {
	IncludeData    bool                   `json:"include_data"`    // 包含数据 Include data
	IncludeSchema  bool                   `json:"include_schema"`  // 包含结构 Include schema
	IncludeIndexes bool                   `json:"include_indexes"` // 包含索引 Include indexes
	Databases      []string               `json:"databases"`       // 数据库列表 Database list
	Tables         []string               `json:"tables"`          // 表列表 Table list
	Compression    bool                   `json:"compression"`     // 是否压缩 Whether compress
	Encryption     bool                   `json:"encryption"`      // 是否加密 Whether encrypt
	Parallel       int                    `json:"parallel"`        // 并行度 Parallelism
	ChunkSize      int64                  `json:"chunk_size"`      // 块大小 Chunk size
	Options        map[string]interface{} `json:"options"`         // 其他选项 Other options
}

// RestoreOptions 恢复选项
// RestoreOptions restore options
type RestoreOptions struct {
	OverwriteExisting bool                   `json:"overwrite_existing"` // 覆盖已存在 Overwrite existing
	SkipErrors        bool                   `json:"skip_errors"`        // 跳过错误 Skip errors
	Databases         []string               `json:"databases"`          // 数据库列表 Database list
	Tables            []string               `json:"tables"`             // 表列表 Table list
	Parallel          int                    `json:"parallel"`           // 并行度 Parallelism
	CheckIntegrity    bool                   `json:"check_integrity"`    // 检查完整性 Check integrity
	Options           map[string]interface{} `json:"options"`            // 其他选项 Other options
}

// 监控类型定义 Monitoring type definitions

// EngineMetrics 引擎监控指标
// EngineMetrics engine monitoring metrics
type EngineMetrics struct {
	Timestamp time.Time `json:"timestamp"` // 时间戳 Timestamp

	// 连接指标 Connection metrics
	ConnectionsActive int `json:"connections_active"` // 活跃连接数 Active connections
	ConnectionsTotal  int `json:"connections_total"`  // 总连接数 Total connections
	ConnectionsMax    int `json:"connections_max"`    // 最大连接数 Max connections

	// 查询指标 Query metrics
	QueriesPerSecond float64       `json:"queries_per_second"` // 每秒查询数 Queries per second
	SlowQueries      int64         `json:"slow_queries"`       // 慢查询数 Slow queries
	AvgQueryTime     time.Duration `json:"avg_query_time"`     // 平均查询时间 Average query time

	// 事务指标 Transaction metrics
	TransactionsActive    int     `json:"transactions_active"`     // 活跃事务数 Active transactions
	TransactionsPerSecond float64 `json:"transactions_per_second"` // 每秒事务数 Transactions per second
	TransactionCommits    int64   `json:"transaction_commits"`     // 事务提交数 Transaction commits
	TransactionRollbacks  int64   `json:"transaction_rollbacks"`   // 事务回滚数 Transaction rollbacks

	// 存储指标 Storage metrics
	DataSize  int64 `json:"data_size"`  // 数据大小 Data size
	IndexSize int64 `json:"index_size"` // 索引大小 Index size
	FreeSpace int64 `json:"free_space"` // 空闲空间 Free space

	// 缓存指标 Cache metrics
	CacheHitRate float64 `json:"cache_hit_rate"` // 缓存命中率 Cache hit rate
	CacheSize    int64   `json:"cache_size"`     // 缓存大小 Cache size
	CacheUsed    int64   `json:"cache_used"`     // 已用缓存 Cache used

	// 锁指标 Lock metrics
	LocksWaiting      int   `json:"locks_waiting"`      // 等待锁数 Waiting locks
	LocksAcquired     int   `json:"locks_acquired"`     // 已获得锁数 Acquired locks
	DeadlocksDetected int64 `json:"deadlocks_detected"` // 检测到的死锁数 Detected deadlocks

	// I/O指标 I/O metrics
	DiskReadsPerSecond  float64 `json:"disk_reads_per_second"`  // 每秒磁盘读取数 Disk reads per second
	DiskWritesPerSecond float64 `json:"disk_writes_per_second"` // 每秒磁盘写入数 Disk writes per second
	BytesRead           int64   `json:"bytes_read"`             // 读取字节数 Bytes read
	BytesWritten        int64   `json:"bytes_written"`          // 写入字节数 Bytes written

	// 错误指标 Error metrics
	Errors   int64 `json:"errors"`   // 错误数 Errors
	Warnings int64 `json:"warnings"` // 警告数 Warnings

	// 自定义指标 Custom metrics
	CustomMetrics map[string]interface{} `json:"custom_metrics"` // 自定义指标 Custom metrics
}

// ConnectionInfo 连接信息
// ConnectionInfo connection information
type ConnectionInfo struct {
	ID              uint64    `json:"id"`               // 连接ID Connection ID
	RemoteAddr      string    `json:"remote_addr"`      // 远程地址 Remote address
	User            string    `json:"user"`             // 用户 User
	Database        string    `json:"database"`         // 数据库 Database
	State           string    `json:"state"`            // 状态 State
	ConnectedAt     time.Time `json:"connected_at"`     // 连接时间 Connected time
	LastActivity    time.Time `json:"last_activity"`    // 最后活动时间 Last activity time
	QueriesExecuted int64     `json:"queries_executed"` // 执行的查询数 Queries executed
	BytesSent       int64     `json:"bytes_sent"`       // 发送字节数 Bytes sent
	BytesReceived   int64     `json:"bytes_received"`   // 接收字节数 Bytes received
	CurrentQuery    string    `json:"current_query"`    // 当前查询 Current query
	TransactionID   uint64    `json:"transaction_id"`   // 事务ID Transaction ID
	LockCount       int       `json:"lock_count"`       // 锁数量 Lock count
}

// 事件类型定义 Event type definitions

// StorageEvent 存储事件
// StorageEvent storage event
type StorageEvent struct {
	Type      StorageEventType       `json:"type"`      // 事件类型 Event type
	Timestamp time.Time              `json:"timestamp"` // 时间戳 Timestamp
	Database  string                 `json:"database"`  // 数据库 Database
	Table     string                 `json:"table"`     // 表 Table
	User      string                 `json:"user"`      // 用户 User
	Data      map[string]interface{} `json:"data"`      // 事件数据 Event data
	Error     error                  `json:"error"`     // 错误信息 Error information
}

// StorageEventType 存储事件类型
// StorageEventType storage event type
type StorageEventType int

const (
	EventDatabaseCreated         StorageEventType = iota // 数据库创建 Database created
	EventDatabaseDropped                                 // 数据库删除 Database dropped
	EventTableCreated                                    // 表创建 Table created
	EventTableDropped                                    // 表删除 Table dropped
	EventTableAltered                                    // 表修改 Table altered
	EventIndexCreated                                    // 索引创建 Index created
	EventIndexDropped                                    // 索引删除 Index dropped
	EventDataInserted                                    // 数据插入 Data inserted
	EventDataUpdated                                     // 数据更新 Data updated
	EventDataDeleted                                     // 数据删除 Data deleted
	EventTransactionStarted                              // 事务开始 Transaction started
	EventTransactionCommitted                            // 事务提交 Transaction committed
	EventTransactionRollback                             // 事务回滚 Transaction rollback
	EventLockAcquired                                    // 锁获得 Lock acquired
	EventLockReleased                                    // 锁释放 Lock released
	EventDeadlockDetected                                // 死锁检测 Deadlock detected
	EventBackupStarted                                   // 备份开始 Backup started
	EventBackupCompleted                                 // 备份完成 Backup completed
	EventRestoreStarted                                  // 恢复开始 Restore started
	EventRestoreCompleted                                // 恢复完成 Restore completed
	EventIntegrityCheckStarted                           // 完整性检查开始 Integrity check started
	EventIntegrityCheckCompleted                         // 完整性检查完成 Integrity check completed
	EventError                                           // 错误事件 Error event
)

// StorageEventListener 存储事件监听器
// StorageEventListener storage event listener
type StorageEventListener interface {
	OnEvent(event *StorageEvent) // 事件处理 Event handling
}

// 高级接口定义 Advanced interface definitions

// QueryOptimizer 查询优化器接口
// QueryOptimizer query optimizer interface
type QueryOptimizer interface {
	Optimize(ctx context.Context, query *SelectQuery, schema *TableSchema) (*QueryPlan, error) // 优化查询 Optimize query
	EstimateCost(ctx context.Context, plan *QueryPlan) (*QueryCost, error)                     // 估算成本 Estimate cost
	GetStats(ctx context.Context) (*OptimizerStats, error)                                     // 获取统计信息 Get statistics
}

// QueryPlan 查询计划
// QueryPlan query plan
type QueryPlan struct {
	ID            string                 `json:"id"`             // 计划ID Plan ID
	Query         *SelectQuery           `json:"query"`          // 原始查询 Original query
	Steps         []*QueryStep           `json:"steps"`          // 执行步骤 Execution steps
	EstimatedCost *QueryCost             `json:"estimated_cost"` // 估算成本 Estimated cost
	CreatedAt     time.Time              `json:"created_at"`     // 创建时间 Creation time
	Options       map[string]interface{} `json:"options"`        // 选项 Options
}

// QueryStep 查询步骤
// QueryStep query step
type QueryStep struct {
	Type    QueryStepType          `json:"type"`     // 步骤类型 Step type
	Table   string                 `json:"table"`    // 表名 Table name
	Index   string                 `json:"index"`    // 索引名 Index name
	Filter  Filter                 `json:"filter"`   // 过滤条件 Filter condition
	Columns []string               `json:"columns"`  // 列名 Column names
	OrderBy []*OrderBy             `json:"order_by"` // 排序 Order by
	Limit   int64                  `json:"limit"`    // 限制 Limit
	Offset  int64                  `json:"offset"`   // 偏移 Offset
	Cost    *QueryCost             `json:"cost"`     // 成本 Cost
	Options map[string]interface{} `json:"options"`  // 选项 Options
}

// QueryStepType 查询步骤类型
// QueryStepType query step type
type QueryStepType int

const (
	StepTableScan  QueryStepType = iota // 表扫描 Table scan
	StepIndexScan                       // 索引扫描 Index scan
	StepIndexSeek                       // 索引查找 Index seek
	StepFilter                          // 过滤 Filter
	StepSort                            // 排序 Sort
	StepGroupBy                         // 分组 Group by
	StepJoin                            // 连接 Join
	StepUnion                           // 联合 Union
	StepLimit                           // 限制 Limit
	StepProjection                      // 投影 Projection
)

// QueryCost 查询成本
// QueryCost query cost
type QueryCost struct {
	CPUCost    float64 `json:"cpu_cost"`    // CPU成本 CPU cost
	IOCost     float64 `json:"io_cost"`     // IO成本 IO cost
	MemoryCost float64 `json:"memory_cost"` // 内存成本 Memory cost
	TotalCost  float64 `json:"total_cost"`  // 总成本 Total cost
	Rows       int64   `json:"rows"`        // 预计行数 Estimated rows
}

// OptimizerStats 优化器统计信息
// OptimizerStats optimizer statistics
type OptimizerStats struct {
	QueriesOptimized int64         `json:"queries_optimized"` // 已优化查询数 Queries optimized
	AvgOptimizeTime  time.Duration `json:"avg_optimize_time"` // 平均优化时间 Average optimize time
	CacheHitRate     float64       `json:"cache_hit_rate"`    // 缓存命中率 Cache hit rate
	PlansCached      int           `json:"plans_cached"`      // 缓存的计划数 Cached plans
}

// PartitionManager 分区管理器接口
// PartitionManager partition manager interface
type PartitionManager interface {
	CreatePartition(ctx context.Context, database, table string, partition *PartitionDefinition) error           // 创建分区 Create partition
	DropPartition(ctx context.Context, database, table, partition string) error                                  // 删除分区 Drop partition
	ListPartitions(ctx context.Context, database, table string) ([]*PartitionInfo, error)                        // 列出分区 List partitions
	GetPartition(ctx context.Context, database, table, partition string) (*PartitionInfo, error)                 // 获取分区信息 Get partition info
	PrunePartitions(ctx context.Context, database, table string, condition Filter) ([]string, error)             // 分区裁剪 Partition pruning
	MergePartitions(ctx context.Context, database, table string, partitions []string, newPartition string) error // 合并分区 Merge partitions
	SplitPartition(ctx context.Context, database, table, partition string, splitPoint interface{}) error         // 分割分区 Split partition
}

// PartitionDefinition 分区定义
// PartitionDefinition partition definition
type PartitionDefinition struct {
	Name          string                 `json:"name"`           // 分区名 Partition name
	Type          PartitionType          `json:"type"`           // 分区类型 Partition type
	Columns       []string               `json:"columns"`        // 分区列 Partition columns
	Values        []interface{}          `json:"values"`         // 分区值 Partition values
	SubPartitions []*PartitionDefinition `json:"sub_partitions"` // 子分区 Sub partitions
	Options       map[string]interface{} `json:"options"`        // 选项 Options
}

// PartitionType 分区类型
// PartitionType partition type
type PartitionType int

const (
	PartitionRange     PartitionType = iota // 范围分区 Range partition
	PartitionHash                           // 哈希分区 Hash partition
	PartitionList                           // 列表分区 List partition
	PartitionComposite                      // 复合分区 Composite partition
)

// PartitionInfo 分区信息
// PartitionInfo partition information
type PartitionInfo struct {
	Name         string                 `json:"name"`          // 分区名 Partition name
	Table        string                 `json:"table"`         // 表名 Table name
	Database     string                 `json:"database"`      // 数据库名 Database name
	Type         PartitionType          `json:"type"`          // 分区类型 Partition type
	Columns      []string               `json:"columns"`       // 分区列 Partition columns
	Values       []interface{}          `json:"values"`        // 分区值 Partition values
	RowCount     int64                  `json:"row_count"`     // 行数 Row count
	Size         int64                  `json:"size"`          // 大小 Size
	IndexSize    int64                  `json:"index_size"`    // 索引大小 Index size
	LastAccessed time.Time              `json:"last_accessed"` // 最后访问时间 Last accessed time
	CreatedAt    time.Time              `json:"created_at"`    // 创建时间 Creation time
	Options      map[string]interface{} `json:"options"`       // 选项 Options
}

// ReplicationManager 复制管理器接口
// ReplicationManager replication manager interface
type ReplicationManager interface {
	// 主从复制 Master-slave replication
	StartReplication(ctx context.Context, config *ReplicationConfig) error // 启动复制 Start replication
	StopReplication(ctx context.Context) error                             // 停止复制 Stop replication
	GetReplicationStatus(ctx context.Context) (*ReplicationStatus, error)  // 获取复制状态 Get replication status

	// 主节点操作 Master operations
	RegisterSlave(ctx context.Context, slave *SlaveInfo) error // 注册从节点 Register slave
	UnregisterSlave(ctx context.Context, slaveID string) error // 注销从节点 Unregister slave
	GetSlaves(ctx context.Context) ([]*SlaveInfo, error)       // 获取从节点列表 Get slave list

	// 从节点操作 Slave operations
	SynchronizeFromMaster(ctx context.Context) error            // 从主节点同步 Synchronize from master
	ApplyBinlog(ctx context.Context, binlog *BinlogEntry) error // 应用二进制日志 Apply binlog
	GetLag(ctx context.Context) (time.Duration, error)          // 获取复制延迟 Get replication lag

	// 故障转移 Failover
	PromoteToMaster(ctx context.Context) error                  // 提升为主节点 Promote to master
	DemoteToSlave(ctx context.Context, masterAddr string) error // 降级为从节点 Demote to slave

	// 监控 Monitoring
	GetMetrics(ctx context.Context) (*ReplicationMetrics, error) // 获取复制指标 Get replication metrics
	AddEventListener(listener ReplicationEventListener)          // 添加事件监听器 Add event listener
	RemoveEventListener(listener ReplicationEventListener)       // 移除事件监听器 Remove event listener
}

// ReplicationConfig 复制配置
// ReplicationConfig replication configuration
type ReplicationConfig struct {
	Mode          ReplicationMode        `json:"mode"`           // 复制模式 Replication mode
	MasterAddr    string                 `json:"master_addr"`    // 主节点地址 Master address
	SlaveAddr     string                 `json:"slave_addr"`     // 从节点地址 Slave address
	Username      string                 `json:"username"`       // 用户名 Username
	Password      string                 `json:"password"`       // 密码 Password
	Databases     []string               `json:"databases"`      // 复制的数据库 Replicated databases
	Tables        []string               `json:"tables"`         // 复制的表 Replicated tables
	BinlogFormat  BinlogFormat           `json:"binlog_format"`  // 二进制日志格式 Binlog format
	SyncMode      SyncMode               `json:"sync_mode"`      // 同步模式 Sync mode
	BatchSize     int                    `json:"batch_size"`     // 批处理大小 Batch size
	RetryInterval time.Duration          `json:"retry_interval"` // 重试间隔 Retry interval
	MaxRetries    int                    `json:"max_retries"`    // 最大重试次数 Max retries
	SSL           *SSLConfig             `json:"ssl"`            // SSL配置 SSL configuration
	Options       map[string]interface{} `json:"options"`        // 其他选项 Other options
}

// ReplicationMode 复制模式
// ReplicationMode replication mode
type ReplicationMode int

const (
	ReplicationAsync    ReplicationMode = iota // 异步复制 Async replication
	ReplicationSync                            // 同步复制 Sync replication
	ReplicationSemiSync                        // 半同步复制 Semi-sync replication
)

// BinlogFormat 二进制日志格式
// BinlogFormat binlog format
type BinlogFormat int

const (
	BinlogStatement BinlogFormat = iota // 语句格式 Statement format
	BinlogRow                           // 行格式 Row format
	BinlogMixed                         // 混合格式 Mixed format
)

// SyncMode 同步模式
// SyncMode sync mode
type SyncMode int

const (
	SyncModeAsync SyncMode = iota // 异步同步 Async sync
	SyncModeSync                  // 同步同步 Sync sync
	SyncModeBatch                 // 批量同步 Batch sync
)

// ReplicationStatus 复制状态
// ReplicationStatus replication status
type ReplicationStatus struct {
	Role           ReplicationRole  `json:"role"`             // 角色 Role
	State          ReplicationState `json:"state"`            // 状态 State
	MasterAddr     string           `json:"master_addr"`      // 主节点地址 Master address
	SlaveCount     int              `json:"slave_count"`      // 从节点数量 Slave count
	LastBinlogFile string           `json:"last_binlog_file"` // 最后二进制日志文件 Last binlog file
	LastBinlogPos  int64            `json:"last_binlog_pos"`  // 最后二进制日志位置 Last binlog position
	Lag            time.Duration    `json:"lag"`              // 复制延迟 Replication lag
	BytesSent      int64            `json:"bytes_sent"`       // 发送字节数 Bytes sent
	BytesReceived  int64            `json:"bytes_received"`   // 接收字节数 Bytes received
	EventsApplied  int64            `json:"events_applied"`   // 已应用事件数 Events applied
	ErrorCount     int64            `json:"error_count"`      // 错误数量 Error count
	LastError      string           `json:"last_error"`       // 最后错误 Last error
	StartedAt      time.Time        `json:"started_at"`       // 开始时间 Started time
	LastEventTime  time.Time        `json:"last_event_time"`  // 最后事件时间 Last event time
}

// ReplicationRole 复制角色
// ReplicationRole replication role
type ReplicationRole int

const (
	RoleMaster     ReplicationRole = iota // 主节点 Master
	RoleSlave                             // 从节点 Slave
	RoleStandalone                        // 独立节点 Standalone
)

// ReplicationState 复制状态
// ReplicationState replication state
type ReplicationState int

const (
	StateRunning    ReplicationState = iota // 运行中 Running
	StateStopped                            // 已停止 Stopped
	StateError                              // 错误 Error
	StateConnecting                         // 连接中 Connecting
	StateSyncing                            // 同步中 Syncing
)

// SlaveInfo 从节点信息
// SlaveInfo slave information
type SlaveInfo struct {
	ID           string           `json:"id"`            // 从节点ID Slave ID
	Addr         string           `json:"addr"`          // 地址 Address
	State        ReplicationState `json:"state"`         // 状态 State
	Lag          time.Duration    `json:"lag"`           // 延迟 Lag
	LastContact  time.Time        `json:"last_contact"`  // 最后联系时间 Last contact time
	BytesSent    int64            `json:"bytes_sent"`    // 发送字节数 Bytes sent
	EventsSent   int64            `json:"events_sent"`   // 发送事件数 Events sent
	RegisteredAt time.Time        `json:"registered_at"` // 注册时间 Registered time
}

// BinlogEntry 二进制日志条目
// BinlogEntry binlog entry
type BinlogEntry struct {
	File      string                 `json:"file"`       // 文件名 File name
	Position  int64                  `json:"position"`   // 位置 Position
	Timestamp time.Time              `json:"timestamp"`  // 时间戳 Timestamp
	EventType BinlogEventType        `json:"event_type"` // 事件类型 Event type
	Database  string                 `json:"database"`   // 数据库 Database
	Table     string                 `json:"table"`      // 表 Table
	Data      map[string]interface{} `json:"data"`       // 数据 Data
	SQL       string                 `json:"sql"`        // SQL语句 SQL statement
}

// BinlogEventType 二进制日志事件类型
// BinlogEventType binlog event type
type BinlogEventType int

const (
	BinlogInsert      BinlogEventType = iota // 插入事件 Insert event
	BinlogUpdate                             // 更新事件 Update event
	BinlogDelete                             // 删除事件 Delete event
	BinlogDDL                                // DDL事件 DDL event
	BinlogTransaction                        // 事务事件 Transaction event
)

// ReplicationMetrics 复制指标
// ReplicationMetrics replication metrics
type ReplicationMetrics struct {
	Timestamp             time.Time     `json:"timestamp"`               // 时间戳 Timestamp
	BytesPerSecond        float64       `json:"bytes_per_second"`        // 每秒字节数 Bytes per second
	EventsPerSecond       float64       `json:"events_per_second"`       // 每秒事件数 Events per second
	AverageLag            time.Duration `json:"average_lag"`             // 平均延迟 Average lag
	MaxLag                time.Duration `json:"max_lag"`                 // 最大延迟 Max lag
	ConnectedSlaves       int           `json:"connected_slaves"`        // 连接的从节点数 Connected slaves
	TotalBytesReplicated  int64         `json:"total_bytes_replicated"`  // 总复制字节数 Total bytes replicated
	TotalEventsReplicated int64         `json:"total_events_replicated"` // 总复制事件数 Total events replicated
	ErrorRate             float64       `json:"error_rate"`              // 错误率 Error rate
	ReconnectCount        int64         `json:"reconnect_count"`         // 重连次数 Reconnect count
}

// ReplicationEventListener 复制事件监听器
// ReplicationEventListener replication event listener
type ReplicationEventListener interface {
	OnSlaveConnected(slave *SlaveInfo)  // 从节点连接事件 Slave connected event
	OnSlaveDisconnected(slaveID string) // 从节点断开事件 Slave disconnected event
	OnReplicationStarted()              // 复制开始事件 Replication started event
	OnReplicationStopped()              // 复制停止事件 Replication stopped event
	OnReplicationError(err error)       // 复制错误事件 Replication error event
	OnLagExceeded(lag time.Duration)    // 延迟超限事件 Lag exceeded event
}

// SSLConfig SSL配置
// SSLConfig SSL configuration
type SSLConfig struct {
	Enabled    bool   `json:"enabled"`     // 是否启用 Whether enabled
	CertFile   string `json:"cert_file"`   // 证书文件 Certificate file
	KeyFile    string `json:"key_file"`    // 密钥文件 Key file
	CAFile     string `json:"ca_file"`     // CA文件 CA file
	SkipVerify bool   `json:"skip_verify"` // 跳过验证 Skip verification
}

// 扩展接口 Extended interfaces

// FullTextSearchEngine 全文搜索引擎接口
// FullTextSearchEngine full-text search engine interface
type FullTextSearchEngine interface {
	// 索引管理 Index management
	CreateIndex(ctx context.Context, database, table string, config *FullTextIndexConfig) error // 创建全文索引 Create full-text index
	DropIndex(ctx context.Context, database, table, index string) error                         // 删除全文索引 Drop full-text index
	RebuildIndex(ctx context.Context, database, table, index string) error                      // 重建全文索引 Rebuild full-text index

	// 搜索 Search
	Search(ctx context.Context, database, table string, query *FullTextQuery) (*SearchResult, error)        // 全文搜索 Full-text search
	Suggest(ctx context.Context, database, table string, query *SuggestionQuery) (*SuggestionResult, error) // 搜索建议 Search suggestion

	// 统计 Statistics
	GetIndexStats(ctx context.Context, database, table, index string) (*FullTextIndexStats, error) // 获取索引统计 Get index statistics
}

// FullTextIndexConfig 全文索引配置
// FullTextIndexConfig full-text index configuration
type FullTextIndexConfig struct {
	Name     string                 `json:"name"`     // 索引名 Index name
	Columns  []string               `json:"columns"`  // 索引列 Index columns
	Language string                 `json:"language"` // 语言 Language
	Analyzer string                 `json:"analyzer"` // 分析器 Analyzer
	Options  map[string]interface{} `json:"options"`  // 选项 Options
}

// FullTextQuery 全文查询
// FullTextQuery full-text query
type FullTextQuery struct {
	Text      string                 `json:"text"`      // 查询文本 Query text
	Columns   []string               `json:"columns"`   // 搜索列 Search columns
	Mode      SearchMode             `json:"mode"`      // 搜索模式 Search mode
	Limit     int                    `json:"limit"`     // 限制数量 Limit
	Offset    int                    `json:"offset"`    // 偏移量 Offset
	Highlight bool                   `json:"highlight"` // 是否高亮 Whether highlight
	Options   map[string]interface{} `json:"options"`   // 选项 Options
}

// SearchMode 搜索模式
// SearchMode search mode
type SearchMode int

const (
	SearchModeNatural SearchMode = iota // 自然语言模式 Natural language mode
	SearchModeBoolean                   // 布尔模式 Boolean mode
	SearchModeFuzzy                     // 模糊模式 Fuzzy mode
	SearchModePhrase                    // 短语模式 Phrase mode
)

// SearchResult 搜索结果
// SearchResult search result
type SearchResult struct {
	Total     int64                     `json:"total"`      // 总数 Total count
	Hits      []*SearchHit              `json:"hits"`       // 命中结果 Hits
	TimeTaken time.Duration             `json:"time_taken"` // 耗时 Time taken
	Facets    map[string][]*FacetResult `json:"facets"`     // 分面结果 Facet results
}

// SearchHit 搜索命中
// SearchHit search hit
type SearchHit struct {
	Score      float64             `json:"score"`      // 得分 Score
	Row        Row                 `json:"row"`        // 行数据 Row data
	Highlights map[string][]string `json:"highlights"` // 高亮片段 Highlights
	Explain    *ScoreExplanation   `json:"explain"`    // 得分解释 Score explanation
}

// ScoreExplanation 得分解释
// ScoreExplanation score explanation
type ScoreExplanation struct {
	Value       float64             `json:"value"`       // 得分值 Score value
	Description string              `json:"description"` // 描述 Description
	Details     []*ScoreExplanation `json:"details"`     // 详细信息 Details
}

// SuggestionQuery 建议查询
// SuggestionQuery suggestion query
type SuggestionQuery struct {
	Text    string                 `json:"text"`    // 查询文本 Query text
	Field   string                 `json:"field"`   // 字段 Field
	Size    int                    `json:"size"`    // 建议数量 Suggestion count
	Options map[string]interface{} `json:"options"` // 选项 Options
}

// SuggestionResult 建议结果
// SuggestionResult suggestion result
type SuggestionResult struct {
	Suggestions []*Suggestion `json:"suggestions"` // 建议列表 Suggestion list
	TimeTaken   time.Duration `json:"time_taken"`  // 耗时 Time taken
}

// Suggestion 建议
// Suggestion suggestion
type Suggestion struct {
	Text  string  `json:"text"`  // 建议文本 Suggestion text
	Score float64 `json:"score"` // 得分 Score
	Freq  int64   `json:"freq"`  // 频率 Frequency
}

// FacetResult 分面结果
// FacetResult facet result
type FacetResult struct {
	Value string `json:"value"` // 值 Value
	Count int64  `json:"count"` // 数量 Count
}

// FullTextIndexStats 全文索引统计
// FullTextIndexStats full-text index statistics
type FullTextIndexStats struct {
	Name          string        `json:"name"`            // 索引名 Index name
	Size          int64         `json:"size"`            // 索引大小 Index size
	DocumentCount int64         `json:"document_count"`  // 文档数量 Document count
	TermCount     int64         `json:"term_count"`      // 词条数量 Term count
	LastUpdated   time.Time     `json:"last_updated"`    // 最后更新时间 Last updated time
	SearchCount   int64         `json:"search_count"`    // 搜索次数 Search count
	AvgSearchTime time.Duration `json:"avg_search_time"` // 平均搜索时间 Average search time
}

// VectorSearchEngine 向量搜索引擎接口
// VectorSearchEngine vector search engine interface
type VectorSearchEngine interface {
	// 向量索引管理 Vector index management
	CreateVectorIndex(ctx context.Context, database, table string, config *VectorIndexConfig) error // 创建向量索引 Create vector index
	DropVectorIndex(ctx context.Context, database, table, index string) error                       // 删除向量索引 Drop vector index

	// 向量搜索 Vector search
	VectorSearch(ctx context.Context, database, table string, query *VectorQuery) (*VectorSearchResult, error)             // 向量搜索 Vector search
	SimilaritySearch(ctx context.Context, database, table string, vector []float64, topK int) (*VectorSearchResult, error) // 相似性搜索 Similarity search

	// 向量操作 Vector operations
	InsertVectors(ctx context.Context, database, table string, vectors []*VectorData) error // 插入向量 Insert vectors
	UpdateVectors(ctx context.Context, database, table string, vectors []*VectorData) error // 更新向量 Update vectors
	DeleteVectors(ctx context.Context, database, table string, ids []string) error          // 删除向量 Delete vectors
}

// VectorIndexConfig 向量索引配置
// VectorIndexConfig vector index configuration
type VectorIndexConfig struct {
	Name       string                 `json:"name"`       // 索引名 Index name
	Column     string                 `json:"column"`     // 向量列 Vector column
	Dimension  int                    `json:"dimension"`  // 维度 Dimension
	Algorithm  VectorIndexAlgorithm   `json:"algorithm"`  // 算法 Algorithm
	Metric     DistanceMetric         `json:"metric"`     // 距离度量 Distance metric
	Parameters map[string]interface{} `json:"parameters"` // 参数 Parameters
	Options    map[string]interface{} `json:"options"`    // 选项 Options
}

// VectorIndexAlgorithm 向量索引算法
// VectorIndexAlgorithm vector index algorithm
type VectorIndexAlgorithm int

const (
	VectorIndexIVF   VectorIndexAlgorithm = iota // IVF算法 IVF algorithm
	VectorIndexHNSW                              // HNSW算法 HNSW algorithm
	VectorIndexLSH                               // LSH算法 LSH algorithm
	VectorIndexAnnoy                             // Annoy算法 Annoy algorithm
)

// DistanceMetric 距离度量
// DistanceMetric distance metric
type DistanceMetric int

const (
	DistanceEuclidean  DistanceMetric = iota // 欧几里得距离 Euclidean distance
	DistanceCosine                           // 余弦距离 Cosine distance
	DistanceDotProduct                       // 点积距离 Dot product distance
	DistanceManhattan                        // 曼哈顿距离 Manhattan distance
)

// VectorQuery 向量查询
// VectorQuery vector query
type VectorQuery struct {
	Vector  []float64              `json:"vector"`  // 查询向量 Query vector
	TopK    int                    `json:"top_k"`   // 返回数量 Return count
	Filter  Filter                 `json:"filter"`  // 过滤条件 Filter condition
	Options map[string]interface{} `json:"options"` // 选项 Options
}

// VectorSearchResult 向量搜索结果
// VectorSearchResult vector search result
type VectorSearchResult struct {
	Total     int64         `json:"total"`      // 总数 Total count
	Hits      []*VectorHit  `json:"hits"`       // 命中结果 Hits
	TimeTaken time.Duration `json:"time_taken"` // 耗时 Time taken
}

// VectorHit 向量命中
// VectorHit vector hit
type VectorHit struct {
	ID       string    `json:"id"`       // ID
	Score    float64   `json:"score"`    // 得分 Score
	Distance float64   `json:"distance"` // 距离 Distance
	Vector   []float64 `json:"vector"`   // 向量 Vector
	Row      Row       `json:"row"`      // 行数据 Row data
}

// VectorData 向量数据
// VectorData vector data
type VectorData struct {
	ID     string    `json:"id"`     // ID
	Vector []float64 `json:"vector"` // 向量 Vector
	Data   Row       `json:"data"`   // 关联数据 Associated data
}

// 工厂接口 Factory interfaces

// StorageEngineFactory 存储引擎工厂接口
// StorageEngineFactory storage engine factory interface
type StorageEngineFactory interface {
	CreateEngine(name string, config map[string]interface{}) (StorageEngine, error) // 创建存储引擎 Create storage engine
	GetSupportedEngines() []string                                                  // 获取支持的引擎列表 Get supported engines
	GetEngineInfo(name string) (*EngineInfo, error)                                 // 获取引擎信息 Get engine information
}

// EngineInfo 引擎信息
// EngineInfo engine information
type EngineInfo struct {
	Name         string                 `json:"name"`         // 引擎名 Engine name
	Version      string                 `json:"version"`      // 版本 Version
	Description  string                 `json:"description"`  // 描述 Description
	Features     []string               `json:"features"`     // 特性 Features
	Capabilities map[string]bool        `json:"capabilities"` // 能力 Capabilities
	Options      map[string]interface{} `json:"options"`      // 选项 Options
}

// 辅助函数和常量 Helper functions and constants

// StorageEngineCapability 存储引擎能力
// StorageEngineCapability storage engine capability
type StorageEngineCapability string

const (
	CapabilityTransactions   StorageEngineCapability = "transactions"    // 事务支持 Transaction support
	CapabilityIndexes        StorageEngineCapability = "indexes"         // 索引支持 Index support
	CapabilityForeignKeys    StorageEngineCapability = "foreign_keys"    // 外键支持 Foreign key support
	CapabilityFullTextSearch StorageEngineCapability = "fulltext_search" // 全文搜索 Full-text search
	CapabilityVectorSearch   StorageEngineCapability = "vector_search"   // 向量搜索 Vector search
	CapabilityPartitioning   StorageEngineCapability = "partitioning"    // 分区支持 Partitioning support
	CapabilityReplication    StorageEngineCapability = "replication"     // 复制支持 Replication support
	CapabilityCompression    StorageEngineCapability = "compression"     // 压缩支持 Compression support
	CapabilityEncryption     StorageEngineCapability = "encryption"      // 加密支持 Encryption support
	CapabilityMVCC           StorageEngineCapability = "mvcc"            // MVCC支持 MVCC support
	CapabilityWAL            StorageEngineCapability = "wal"             // WAL支持 WAL support
	CapabilityPointInTime    StorageEngineCapability = "point_in_time"   // 时点恢复 Point-in-time recovery
	CapabilityCluster        StorageEngineCapability = "cluster"         // 集群支持 Cluster support
)

// EngineFeature 引擎特性
// EngineFeature engine feature
type EngineFeature string

const (
	FeatureACID                  EngineFeature = "acid"                   // ACID事务 ACID transactions
	FeatureHighAvailability      EngineFeature = "high_availability"      // 高可用 High availability
	FeatureHorizontalScaling     EngineFeature = "horizontal_scaling"     // 水平扩展 Horizontal scaling
	FeatureInMemory              EngineFeature = "in_memory"              // 内存存储 In-memory storage
	FeatureColumnOriented        EngineFeature = "column_oriented"        // 列式存储 Column-oriented storage
	FeatureRowOriented           EngineFeature = "row_oriented"           // 行式存储 Row-oriented storage
	FeatureTimeSeriesOptimized   EngineFeature = "timeseries_optimized"   // 时序优化 Time-series optimized
	FeatureAnalyticalWorkload    EngineFeature = "analytical_workload"    // 分析工作负载 Analytical workload
	FeatureTransactionalWorkload EngineFeature = "transactional_workload" // 事务工作负载 Transactional workload
)

// 错误定义 Error definitions
var (
	ErrEngineNotSupported    = errors.NewError(errors.ErrCodeNotSupported, "Storage engine not supported")
	ErrDatabaseNotFound      = errors.NewError(errors.ErrCodeNotFound, "Database not found")
	ErrTableNotFound         = errors.NewError(errors.ErrCodeNotFound, "Table not found")
	ErrColumnNotFound        = errors.NewError(errors.ErrCodeNotFound, "Column not found")
	ErrIndexNotFound         = errors.NewError(errors.ErrCodeNotFound, "Index not found")
	ErrConstraintNotFound    = errors.NewError(errors.ErrCodeNotFound, "Constraint not found")
	ErrDatabaseExists        = errors.NewError(errors.ErrCodeAlreadyExists, "Database already exists")
	ErrTableExists           = errors.NewError(errors.ErrCodeAlreadyExists, "Table already exists")
	ErrColumnExists          = errors.NewError(errors.ErrCodeAlreadyExists, "Column already exists")
	ErrIndexExists           = errors.NewError(errors.ErrCodeAlreadyExists, "Index already exists")
	ErrConstraintExists      = errors.NewError(errors.ErrCodeAlreadyExists, "Constraint already exists")
	ErrTransactionNotActive  = errors.NewError(errors.ErrCodeInvalidState, "Transaction not active")
	ErrTransactionReadOnly   = errors.NewError(errors.ErrCodePermissionDenied, "Transaction is read-only")
	ErrDeadlockDetected      = errors.NewError(errors.ErrCodeDeadlock, "Deadlock detected")
	ErrLockTimeout           = errors.NewError(errors.ErrCodeTimeout, "Lock timeout")
	ErrIteratorClosed        = errors.NewError(errors.ErrCodeInvalidState, "Iterator is closed")
	ErrInvalidSchema         = errors.NewError(errors.ErrCodeInvalidParameter, "Invalid schema")
	ErrSchemaVersionMismatch = errors.NewError(errors.ErrCodeVersionMismatch, "Schema version mismatch")
	ErrPermissionDenied      = errors.NewError(errors.ErrCodePermissionDenied, "Permission denied")
	ErrQuotaExceeded         = errors.NewError(errors.ErrCodeQuotaExceeded, "Quota exceeded")
	ErrStorageCorrupted      = errors.NewError(errors.ErrCodeDataCorruption, "Storage corrupted")
)
