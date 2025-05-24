// Package config 提供了 GuoceDB 的配置管理系统
// Package config provides configuration management system for GuoceDB
package config

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/guocedb/guocedb/common/errors"
	//"github.com/guocedb/guocedb/common/utils"
)

// ===== 配置结构体定义 Configuration Structure Definition =====

// Config 主配置结构体
// Config main configuration structure
type Config struct {
	// Server 服务器配置
	// Server server configuration
	Server ServerConfig `yaml:"server" json:"server"`

	// Storage 存储引擎配置
	// Storage storage engine configuration
	Storage StorageConfig `yaml:"storage" json:"storage"`

	// Cache 缓存配置
	// Cache cache configuration
	Cache CacheConfig `yaml:"cache" json:"cache"`

	// Query 查询执行配置
	// Query query execution configuration
	Query QueryConfig `yaml:"query" json:"query"`

	// Transaction 事务配置
	// Transaction transaction configuration
	Transaction TransactionConfig `yaml:"transaction" json:"transaction"`

	// Replication 复制配置
	// Replication replication configuration
	Replication ReplicationConfig `yaml:"replication" json:"replication"`

	// Security 安全配置
	// Security security configuration
	Security SecurityConfig `yaml:"security" json:"security"`

	// Monitoring 监控配置
	// Monitoring monitoring configuration
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`

	// Logging 日志配置
	// Logging logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// Performance 性能调优配置
	// Performance performance tuning configuration
	Performance PerformanceConfig `yaml:"performance" json:"performance"`

	// Cluster 集群配置
	// Cluster cluster configuration
	Cluster ClusterConfig `yaml:"cluster" json:"cluster"`

	// Development 开发配置
	// Development development configuration
	Development DevelopmentConfig `yaml:"development" json:"development"`
}

// ServerConfig 服务器配置
// ServerConfig server configuration
type ServerConfig struct {
	// Host 监听地址
	// Host listen address
	Host string `yaml:"host" json:"host" default:"localhost"`

	// Port 监听端口
	// Port listen port
	Port int `yaml:"port" json:"port" default:"3306"`

	// MaxConnections 最大连接数
	// MaxConnections maximum connections
	MaxConnections int `yaml:"max_connections" json:"max_connections" default:"1000"`

	// ConnectionTimeout 连接超时时间
	// ConnectionTimeout connection timeout
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout" default:"30s"`

	// ReadTimeout 读取超时时间
	// ReadTimeout read timeout
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout" default:"300s"`

	// WriteTimeout 写入超时时间
	// WriteTimeout write timeout
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" default:"300s"`

	// KeepAlive TCP保活时间
	// KeepAlive TCP keep-alive time
	KeepAlive time.Duration `yaml:"keep_alive" json:"keep_alive" default:"60s"`

	// MaxPacketSize 最大数据包大小
	// MaxPacketSize maximum packet size
	MaxPacketSize int `yaml:"max_packet_size" json:"max_packet_size" default:"67108864"` // 64MB

	// SocketPath Unix socket路径
	// SocketPath Unix socket path
	SocketPath string `yaml:"socket_path" json:"socket_path" default:"/tmp/guocedb.sock"`

	// EnableTCP 是否启用TCP监听
	// EnableTCP enable TCP listening
	EnableTCP bool `yaml:"enable_tcp" json:"enable_tcp" default:"true"`

	// EnableUnixSocket 是否启用Unix socket
	// EnableUnixSocket enable Unix socket
	EnableUnixSocket bool `yaml:"enable_unix_socket" json:"enable_unix_socket" default:"true"`

	// ServerID 服务器ID（用于复制）
	// ServerID server ID (for replication)
	ServerID uint32 `yaml:"server_id" json:"server_id" default:"1"`

	// ServerUUID 服务器UUID
	// ServerUUID server UUID
	ServerUUID string `yaml:"server_uuid" json:"server_uuid"`
}

// StorageConfig 存储引擎配置
// StorageConfig storage engine configuration
type StorageConfig struct {
	// Engine 存储引擎类型
	// Engine storage engine type
	Engine string `yaml:"engine" json:"engine" default:"innodb"`

	// DataDir 数据目录
	// DataDir data directory
	DataDir string `yaml:"data_dir" json:"data_dir" default:"./data"`

	// TempDir 临时目录
	// TempDir temporary directory
	TempDir string `yaml:"temp_dir" json:"temp_dir" default:"./tmp"`

	// MaxOpenFiles 最大打开文件数
	// MaxOpenFiles maximum open files
	MaxOpenFiles int `yaml:"max_open_files" json:"max_open_files" default:"10000"`

	// PageSize 页大小
	// PageSize page size
	PageSize int `yaml:"page_size" json:"page_size" default:"16384"` // 16KB

	// BufferPoolSize 缓冲池大小
	// BufferPoolSize buffer pool size
	BufferPoolSize int64 `yaml:"buffer_pool_size" json:"buffer_pool_size" default:"134217728"` // 128MB

	// LogBufferSize 日志缓冲区大小
	// LogBufferSize log buffer size
	LogBufferSize int64 `yaml:"log_buffer_size" json:"log_buffer_size" default:"16777216"` // 16MB

	// FlushLogAtCommit 提交时刷新日志
	// FlushLogAtCommit flush log at commit
	FlushLogAtCommit int `yaml:"flush_log_at_commit" json:"flush_log_at_commit" default:"1"`

	// DoubleWrite 双写缓冲
	// DoubleWrite double write buffer
	DoubleWrite bool `yaml:"double_write" json:"double_write" default:"true"`

	// Compression 压缩算法
	// Compression compression algorithm
	Compression string `yaml:"compression" json:"compression" default:"zstd"`

	// ChecksumAlgorithm 校验和算法
	// ChecksumAlgorithm checksum algorithm
	ChecksumAlgorithm string `yaml:"checksum_algorithm" json:"checksum_algorithm" default:"crc32"`

	// FilePerTable 每表一个文件
	// FilePerTable file per table
	FilePerTable bool `yaml:"file_per_table" json:"file_per_table" default:"true"`

	// AutoExtendIncrement 自动扩展增量
	// AutoExtendIncrement auto extend increment
	AutoExtendIncrement int64 `yaml:"auto_extend_increment" json:"auto_extend_increment" default:"67108864"` // 64MB

	// SyncBinlog 同步binlog
	// SyncBinlog sync binlog
	SyncBinlog int `yaml:"sync_binlog" json:"sync_binlog" default:"1"`

	// BinlogFormat binlog格式
	// BinlogFormat binlog format
	BinlogFormat string `yaml:"binlog_format" json:"binlog_format" default:"ROW"`

	// BinlogCacheSize binlog缓存大小
	// BinlogCacheSize binlog cache size
	BinlogCacheSize int64 `yaml:"binlog_cache_size" json:"binlog_cache_size" default:"32768"` // 32KB

	// ExpireDays 数据过期天数
	// ExpireDays data expire days
	ExpireDays int `yaml:"expire_days" json:"expire_days" default:"0"`
}

// CacheConfig 缓存配置
// CacheConfig cache configuration
type CacheConfig struct {
	// QueryCacheEnabled 查询缓存是否启用
	// QueryCacheEnabled query cache enabled
	QueryCacheEnabled bool `yaml:"query_cache_enabled" json:"query_cache_enabled" default:"true"`

	// QueryCacheSize 查询缓存大小
	// QueryCacheSize query cache size
	QueryCacheSize int64 `yaml:"query_cache_size" json:"query_cache_size" default:"67108864"` // 64MB

	// QueryCacheLimit 单个查询结果最大缓存大小
	// QueryCacheLimit single query result maximum cache size
	QueryCacheLimit int64 `yaml:"query_cache_limit" json:"query_cache_limit" default:"1048576"` // 1MB

	// TableCacheSize 表缓存大小
	// TableCacheSize table cache size
	TableCacheSize int `yaml:"table_cache_size" json:"table_cache_size" default:"2000"`

	// TableDefinitionCacheSize 表定义缓存大小
	// TableDefinitionCacheSize table definition cache size
	TableDefinitionCacheSize int `yaml:"table_definition_cache_size" json:"table_definition_cache_size" default:"1400"`

	// KeyBufferSize 键缓冲区大小
	// KeyBufferSize key buffer size
	KeyBufferSize int64 `yaml:"key_buffer_size" json:"key_buffer_size" default:"8388608"` // 8MB

	// SortBufferSize 排序缓冲区大小
	// SortBufferSize sort buffer size
	SortBufferSize int64 `yaml:"sort_buffer_size" json:"sort_buffer_size" default:"262144"` // 256KB

	// JoinBufferSize 连接缓冲区大小
	// JoinBufferSize join buffer size
	JoinBufferSize int64 `yaml:"join_buffer_size" json:"join_buffer_size" default:"262144"` // 256KB

	// ReadBufferSize 读缓冲区大小
	// ReadBufferSize read buffer size
	ReadBufferSize int64 `yaml:"read_buffer_size" json:"read_buffer_size" default:"131072"` // 128KB

	// ReadRndBufferSize 随机读缓冲区大小
	// ReadRndBufferSize random read buffer size
	ReadRndBufferSize int64 `yaml:"read_rnd_buffer_size" json:"read_rnd_buffer_size" default:"262144"` // 256KB

	// PreparedStmtCount 预处理语句数量
	// PreparedStmtCount prepared statement count
	PreparedStmtCount int `yaml:"prepared_stmt_count" json:"prepared_stmt_count" default:"16382"`

	// CacheTTL 缓存过期时间
	// CacheTTL cache TTL
	CacheTTL time.Duration `yaml:"cache_ttl" json:"cache_ttl" default:"1h"`

	// CacheEvictionPolicy 缓存淘汰策略
	// CacheEvictionPolicy cache eviction policy
	CacheEvictionPolicy string `yaml:"cache_eviction_policy" json:"cache_eviction_policy" default:"lru"`
}

// QueryConfig 查询执行配置
// QueryConfig query execution configuration
type QueryConfig struct {
	// MaxExecutionTime 最大执行时间
	// MaxExecutionTime maximum execution time
	MaxExecutionTime time.Duration `yaml:"max_execution_time" json:"max_execution_time" default:"0"`

	// MaxJoinSize 最大连接大小
	// MaxJoinSize maximum join size
	MaxJoinSize int64 `yaml:"max_join_size" json:"max_join_size" default:"18446744073709551615"`

	// MaxLengthForSortData 排序数据最大长度
	// MaxLengthForSortData maximum length for sort data
	MaxLengthForSortData int `yaml:"max_length_for_sort_data" json:"max_length_for_sort_data" default:"1024"`

	// MaxSortLength 最大排序长度
	// MaxSortLength maximum sort length
	MaxSortLength int `yaml:"max_sort_length" json:"max_sort_length" default:"1024"`

	// TmpTableSize 临时表大小
	// TmpTableSize temporary table size
	TmpTableSize int64 `yaml:"tmp_table_size" json:"tmp_table_size" default:"16777216"` // 16MB

	// MaxHeapTableSize 最大堆表大小
	// MaxHeapTableSize maximum heap table size
	MaxHeapTableSize int64 `yaml:"max_heap_table_size" json:"max_heap_table_size" default:"16777216"` // 16MB

	// GroupConcatMaxLen GROUP_CONCAT最大长度
	// GroupConcatMaxLen GROUP_CONCAT maximum length
	GroupConcatMaxLen int `yaml:"group_concat_max_len" json:"group_concat_max_len" default:"1024"`

	// OptimizerSearchDepth 优化器搜索深度
	// OptimizerSearchDepth optimizer search depth
	OptimizerSearchDepth int `yaml:"optimizer_search_depth" json:"optimizer_search_depth" default:"62"`

	// OptimizerPruneLevel 优化器修剪级别
	// OptimizerPruneLevel optimizer prune level
	OptimizerPruneLevel int `yaml:"optimizer_prune_level" json:"optimizer_prune_level" default:"1"`

	// EnableQueryCache 启用查询缓存
	// EnableQueryCache enable query cache
	EnableQueryCache bool `yaml:"enable_query_cache" json:"enable_query_cache" default:"true"`

	// EnableSlowQueryLog 启用慢查询日志
	// EnableSlowQueryLog enable slow query log
	EnableSlowQueryLog bool `yaml:"enable_slow_query_log" json:"enable_slow_query_log" default:"true"`

	// SlowQueryThreshold 慢查询阈值
	// SlowQueryThreshold slow query threshold
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold" json:"slow_query_threshold" default:"10s"`

	// EnableQueryProfiling 启用查询分析
	// EnableQueryProfiling enable query profiling
	EnableQueryProfiling bool `yaml:"enable_query_profiling" json:"enable_query_profiling" default:"false"`

	// ParallelDegree 并行度
	// ParallelDegree parallel degree
	ParallelDegree int `yaml:"parallel_degree" json:"parallel_degree" default:"0"`

	// EnableCBO 启用基于成本的优化器
	// EnableCBO enable cost-based optimizer
	EnableCBO bool `yaml:"enable_cbo" json:"enable_cbo" default:"true"`

	// EnableIndexMerge 启用索引合并
	// EnableIndexMerge enable index merge
	EnableIndexMerge bool `yaml:"enable_index_merge" json:"enable_index_merge" default:"true"`

	// EnableHashJoin 启用哈希连接
	// EnableHashJoin enable hash join
	EnableHashJoin bool `yaml:"enable_hash_join" json:"enable_hash_join" default:"true"`

	// EnableSortMergeJoin 启用排序合并连接
	// EnableSortMergeJoin enable sort merge join
	EnableSortMergeJoin bool `yaml:"enable_sort_merge_join" json:"enable_sort_merge_join" default:"true"`
}

// TransactionConfig 事务配置
// TransactionConfig transaction configuration
type TransactionConfig struct {
	// IsolationLevel 默认隔离级别
	// IsolationLevel default isolation level
	IsolationLevel string `yaml:"isolation_level" json:"isolation_level" default:"REPEATABLE-READ"`

	// AutoCommit 自动提交
	// AutoCommit auto commit
	AutoCommit bool `yaml:"auto_commit" json:"auto_commit" default:"true"`

	// LockWaitTimeout 锁等待超时
	// LockWaitTimeout lock wait timeout
	LockWaitTimeout time.Duration `yaml:"lock_wait_timeout" json:"lock_wait_timeout" default:"50s"`

	// DeadlockDetect 死锁检测
	// DeadlockDetect deadlock detect
	DeadlockDetect bool `yaml:"deadlock_detect" json:"deadlock_detect" default:"true"`

	// DeadlockDetectInterval 死锁检测间隔
	// DeadlockDetectInterval deadlock detect interval
	DeadlockDetectInterval time.Duration `yaml:"deadlock_detect_interval" json:"deadlock_detect_interval" default:"100ms"`

	// MaxTransactionSize 最大事务大小
	// MaxTransactionSize maximum transaction size
	MaxTransactionSize int64 `yaml:"max_transaction_size" json:"max_transaction_size" default:"104857600"` // 100MB

	// MaxTransactionTime 最大事务时间
	// MaxTransactionTime maximum transaction time
	MaxTransactionTime time.Duration `yaml:"max_transaction_time" json:"max_transaction_time" default:"3600s"`

	// MaxUndoLogSize 最大undo日志大小
	// MaxUndoLogSize maximum undo log size
	MaxUndoLogSize int64 `yaml:"max_undo_log_size" json:"max_undo_log_size" default:"1073741824"` // 1GB

	// PurgeInterval 清理间隔
	// PurgeInterval purge interval
	PurgeInterval time.Duration `yaml:"purge_interval" json:"purge_interval" default:"10s"`

	// MaxPurgeLag 最大清理延迟
	// MaxPurgeLag maximum purge lag
	MaxPurgeLag int64 `yaml:"max_purge_lag" json:"max_purge_lag" default:"0"`

	// EnableMVCC 启用MVCC
	// EnableMVCC enable MVCC
	EnableMVCC bool `yaml:"enable_mvcc" json:"enable_mvcc" default:"true"`

	// ReadViewCacheSize 读视图缓存大小
	// ReadViewCacheSize read view cache size
	ReadViewCacheSize int `yaml:"read_view_cache_size" json:"read_view_cache_size" default:"1000"`
}

// ReplicationConfig 复制配置
// ReplicationConfig replication configuration
type ReplicationConfig struct {
	// Enable 是否启用复制
	// Enable enable replication
	Enable bool `yaml:"enable" json:"enable" default:"false"`

	// Role 复制角色 (master/slave)
	// Role replication role (master/slave)
	Role string `yaml:"role" json:"role" default:"master"`

	// MasterHost 主服务器地址
	// MasterHost master host
	MasterHost string `yaml:"master_host" json:"master_host"`

	// MasterPort 主服务器端口
	// MasterPort master port
	MasterPort int `yaml:"master_port" json:"master_port" default:"3306"`

	// MasterUser 主服务器用户
	// MasterUser master user
	MasterUser string `yaml:"master_user" json:"master_user"`

	// MasterPassword 主服务器密码
	// MasterPassword master password
	MasterPassword string `yaml:"master_password" json:"master_password"`

	// BinlogDoDb 需要复制的数据库
	// BinlogDoDb databases to replicate
	BinlogDoDb []string `yaml:"binlog_do_db" json:"binlog_do_db"`

	// BinlogIgnoreDb 忽略复制的数据库
	// BinlogIgnoreDb databases to ignore
	BinlogIgnoreDb []string `yaml:"binlog_ignore_db" json:"binlog_ignore_db"`

	// SlaveNetTimeout 从服务器网络超时
	// SlaveNetTimeout slave network timeout
	SlaveNetTimeout time.Duration `yaml:"slave_net_timeout" json:"slave_net_timeout" default:"60s"`

	// ConnectRetry 连接重试间隔
	// ConnectRetry connect retry interval
	ConnectRetry time.Duration `yaml:"connect_retry" json:"connect_retry" default:"60s"`

	// MaxRelayLogSize 最大中继日志大小
	// MaxRelayLogSize maximum relay log size
	MaxRelayLogSize int64 `yaml:"max_relay_log_size" json:"max_relay_log_size" default:"1073741824"` // 1GB

	// RelayLogPurge 自动清理中继日志
	// RelayLogPurge auto purge relay log
	RelayLogPurge bool `yaml:"relay_log_purge" json:"relay_log_purge" default:"true"`

	// SemiSyncEnabled 半同步复制
	// SemiSyncEnabled semi-sync replication
	SemiSyncEnabled bool `yaml:"semi_sync_enabled" json:"semi_sync_enabled" default:"false"`

	// SemiSyncTimeout 半同步超时
	// SemiSyncTimeout semi-sync timeout
	SemiSyncTimeout time.Duration `yaml:"semi_sync_timeout" json:"semi_sync_timeout" default:"10s"`

	// ParallelWorkers 并行复制工作线程数
	// ParallelWorkers parallel replication workers
	ParallelWorkers int `yaml:"parallel_workers" json:"parallel_workers" default:"4"`

	// PreserveCommitOrder 保持提交顺序
	// PreserveCommitOrder preserve commit order
	PreserveCommitOrder bool `yaml:"preserve_commit_order" json:"preserve_commit_order" default:"true"`
}

// SecurityConfig 安全配置
// SecurityConfig security configuration
type SecurityConfig struct {
	// RequireSecureTransport 要求安全传输
	// RequireSecureTransport require secure transport
	RequireSecureTransport bool `yaml:"require_secure_transport" json:"require_secure_transport" default:"false"`

	// TLSEnabled TLS是否启用
	// TLSEnabled TLS enabled
	TLSEnabled bool `yaml:"tls_enabled" json:"tls_enabled" default:"false"`

	// TLSCert TLS证书文件
	// TLSCert TLS certificate file
	TLSCert string `yaml:"tls_cert" json:"tls_cert"`

	// TLSKey TLS密钥文件
	// TLSKey TLS key file
	TLSKey string `yaml:"tls_key" json:"tls_key"`

	// TLSCA TLS CA文件
	// TLSCA TLS CA file
	TLSCA string `yaml:"tls_ca" json:"tls_ca"`

	// TLSMinVersion TLS最小版本
	// TLSMinVersion TLS minimum version
	TLSMinVersion string `yaml:"tls_min_version" json:"tls_min_version" default:"1.2"`

	// PasswordValidationEnabled 密码验证启用
	// PasswordValidationEnabled password validation enabled
	PasswordValidationEnabled bool `yaml:"password_validation_enabled" json:"password_validation_enabled" default:"true"`

	// PasswordMinLength 密码最小长度
	// PasswordMinLength password minimum length
	PasswordMinLength int `yaml:"password_min_length" json:"password_min_length" default:"8"`

	// PasswordRequireUppercase 密码需要大写字母
	// PasswordRequireUppercase password require uppercase
	PasswordRequireUppercase bool `yaml:"password_require_uppercase" json:"password_require_uppercase" default:"true"`

	// PasswordRequireLowercase 密码需要小写字母
	// PasswordRequireLowercase password require lowercase
	PasswordRequireLowercase bool `yaml:"password_require_lowercase" json:"password_require_lowercase" default:"true"`

	// PasswordRequireNumbers 密码需要数字
	// PasswordRequireNumbers password require numbers
	PasswordRequireNumbers bool `yaml:"password_require_numbers" json:"password_require_numbers" default:"true"`

	// PasswordRequireSpecial 密码需要特殊字符
	// PasswordRequireSpecial password require special
	PasswordRequireSpecial bool `yaml:"password_require_special" json:"password_require_special" default:"true"`

	// MaxFailedLogins 最大失败登录次数
	// MaxFailedLogins maximum failed logins
	MaxFailedLogins int `yaml:"max_failed_logins" json:"max_failed_logins" default:"3"`

	// AccountLockTime 账户锁定时间
	// AccountLockTime account lock time
	AccountLockTime time.Duration `yaml:"account_lock_time" json:"account_lock_time" default:"30m"`

	// AuditEnabled 审计是否启用
	// AuditEnabled audit enabled
	AuditEnabled bool `yaml:"audit_enabled" json:"audit_enabled" default:"false"`

	// AuditLogFile 审计日志文件
	// AuditLogFile audit log file
	AuditLogFile string `yaml:"audit_log_file" json:"audit_log_file" default:"./logs/audit.log"`

	// AuditEvents 审计事件
	// AuditEvents audit events
	AuditEvents []string `yaml:"audit_events" json:"audit_events"`

	// IPWhitelist IP白名单
	// IPWhitelist IP whitelist
	IPWhitelist []string `yaml:"ip_whitelist" json:"ip_whitelist"`

	// IPBlacklist IP黑名单
	// IPBlacklist IP blacklist
	IPBlacklist []string `yaml:"ip_blacklist" json:"ip_blacklist"`
}

// MonitoringConfig 监控配置
// MonitoringConfig monitoring configuration
type MonitoringConfig struct {
	// Enable 是否启用监控
	// Enable enable monitoring
	Enable bool `yaml:"enable" json:"enable" default:"true"`

	// MetricsEndpoint 指标端点
	// MetricsEndpoint metrics endpoint
	MetricsEndpoint string `yaml:"metrics_endpoint" json:"metrics_endpoint" default:"/metrics"`

	// MetricsPort 指标端口
	// MetricsPort metrics port
	MetricsPort int `yaml:"metrics_port" json:"metrics_port" default:"9090"`

	// CollectInterval 收集间隔
	// CollectInterval collect interval
	CollectInterval time.Duration `yaml:"collect_interval" json:"collect_interval" default:"15s"`

	// RetentionPeriod 保留期限
	// RetentionPeriod retention period
	RetentionPeriod time.Duration `yaml:"retention_period" json:"retention_period" default:"168h"` // 7 days

	// EnablePrometheus 启用Prometheus
	// EnablePrometheus enable Prometheus
	EnablePrometheus bool `yaml:"enable_prometheus" json:"enable_prometheus" default:"true"`

	// EnableStatsD 启用StatsD
	// EnableStatsD enable StatsD
	EnableStatsD bool `yaml:"enable_statsd" json:"enable_statsd" default:"false"`

	// StatsDHost StatsD主机
	// StatsDHost StatsD host
	StatsDHost string `yaml:"statsd_host" json:"statsd_host" default:"localhost:8125"`

	// EnableTracing 启用追踪
	// EnableTracing enable tracing
	EnableTracing bool `yaml:"enable_tracing" json:"enable_tracing" default:"false"`

	// TracingEndpoint 追踪端点
	// TracingEndpoint tracing endpoint
	TracingEndpoint string `yaml:"tracing_endpoint" json:"tracing_endpoint"`

	// TracingSampleRate 追踪采样率
	// TracingSampleRate tracing sample rate
	TracingSampleRate float64 `yaml:"tracing_sample_rate" json:"tracing_sample_rate" default:"0.1"`

	// AlertRules 告警规则
	// AlertRules alert rules
	AlertRules []AlertRule `yaml:"alert_rules" json:"alert_rules"`

	// WebhookURL Webhook URL
	// WebhookURL webhook URL
	WebhookURL string `yaml:"webhook_url" json:"webhook_url"`
}

// AlertRule 告警规则
// AlertRule alert rule
type AlertRule struct {
	// Name 规则名称
	// Name rule name
	Name string `yaml:"name" json:"name"`

	// Metric 指标名称
	// Metric metric name
	Metric string `yaml:"metric" json:"metric"`

	// Operator 操作符
	// Operator operator
	Operator string `yaml:"operator" json:"operator"`

	// Threshold 阈值
	// Threshold threshold
	Threshold float64 `yaml:"threshold" json:"threshold"`

	// Duration 持续时间
	// Duration duration
	Duration time.Duration `yaml:"duration" json:"duration"`

	// Action 动作
	// Action action
	Action string `yaml:"action" json:"action"`
}

// LoggingConfig 日志配置
// LoggingConfig logging configuration
type LoggingConfig struct {
	// Level 日志级别
	// Level log level
	Level string `yaml:"level" json:"level" default:"info"`

	// Format 日志格式 (json/text)
	// Format log format (json/text)
	Format string `yaml:"format" json:"format" default:"json"`

	// Output 输出位置
	// Output output location
	Output []string `yaml:"output" json:"output" default:"[\"stdout\"]"`

	// Directory 日志目录
	// Directory log directory
	Directory string `yaml:"directory" json:"directory" default:"./logs"`

	// Filename 日志文件名
	// Filename log filename
	Filename string `yaml:"filename" json:"filename" default:"guocedb.log"`

	// MaxSize 最大文件大小(MB)
	// MaxSize maximum file size(MB)
	MaxSize int `yaml:"max_size" json:"max_size" default:"100"`

	// MaxBackups 最大备份数
	// MaxBackups maximum backups
	MaxBackups int `yaml:"max_backups" json:"max_backups" default:"30"`

	// MaxAge 最大保留天数
	// MaxAge maximum age in days
	MaxAge int `yaml:"max_age" json:"max_age" default:"30"`

	// Compress 是否压缩
	// Compress compress
	Compress bool `yaml:"compress" json:"compress" default:"true"`

	// EnableErrorLog 启用错误日志
	// EnableErrorLog enable error log
	EnableErrorLog bool `yaml:"enable_error_log" json:"enable_error_log" default:"true"`

	// ErrorLogFile 错误日志文件
	// ErrorLogFile error log file
	ErrorLogFile string `yaml:"error_log_file" json:"error_log_file" default:"error.log"`

	// EnableSlowLog 启用慢日志
	// EnableSlowLog enable slow log
	EnableSlowLog bool `yaml:"enable_slow_log" json:"enable_slow_log" default:"true"`

	// SlowLogFile 慢日志文件
	// SlowLogFile slow log file
	SlowLogFile string `yaml:"slow_log_file" json:"slow_log_file" default:"slow.log"`

	// EnableGeneralLog 启用通用日志
	// EnableGeneralLog enable general log
	EnableGeneralLog bool `yaml:"enable_general_log" json:"enable_general_log" default:"false"`

	// GeneralLogFile 通用日志文件
	// GeneralLogFile general log file
	GeneralLogFile string `yaml:"general_log_file" json:"general_log_file" default:"general.log"`

	// EnableAuditLog 启用审计日志
	// EnableAuditLog enable audit log
	EnableAuditLog bool `yaml:"enable_audit_log" json:"enable_audit_log" default:"false"`

	// AuditLogFile 审计日志文件
	// AuditLogFile audit log file
	AuditLogFile string `yaml:"audit_log_file" json:"audit_log_file" default:"audit.log"`

	// SamplingRate 采样率
	// SamplingRate sampling rate
	SamplingRate float64 `yaml:"sampling_rate" json:"sampling_rate" default:"1.0"`

	// EnableCaller 启用调用者信息
	// EnableCaller enable caller info
	EnableCaller bool `yaml:"enable_caller" json:"enable_caller" default:"false"`

	// EnableStacktrace 启用堆栈追踪
	// EnableStacktrace enable stacktrace
	EnableStacktrace bool `yaml:"enable_stacktrace" json:"enable_stacktrace" default:"false"`
}

// PerformanceConfig 性能配置
// PerformanceConfig performance configuration
type PerformanceConfig struct {
	// MaxCPU 最大CPU核心数
	// MaxCPU maximum CPU cores
	MaxCPU int `yaml:"max_cpu" json:"max_cpu" default:"0"`

	// ThreadPoolSize 线程池大小
	// ThreadPoolSize thread pool size
	ThreadPoolSize int `yaml:"thread_pool_size" json:"thread_pool_size" default:"16"`

	// IOThreads IO线程数
	// IOThreads IO threads
	IOThreads int `yaml:"io_threads" json:"io_threads" default:"4"`

	// WorkerThreads 工作线程数
	// WorkerThreads worker threads
	WorkerThreads int `yaml:"worker_threads" json:"worker_threads" default:"16"`

	// EnableAdaptiveHashIndex 启用自适应哈希索引
	// EnableAdaptiveHashIndex enable adaptive hash index
	EnableAdaptiveHashIndex bool `yaml:"enable_adaptive_hash_index" json:"enable_adaptive_hash_index" default:"true"`

	// EnableChangeBuffer 启用变更缓冲
	// EnableChangeBuffer enable change buffer
	EnableChangeBuffer bool `yaml:"enable_change_buffer" json:"enable_change_buffer" default:"true"`

	// ChangeBufferMaxSize 变更缓冲最大大小
	// ChangeBufferMaxSize change buffer maximum size
	ChangeBufferMaxSize int `yaml:"change_buffer_max_size" json:"change_buffer_max_size" default:"25"`

	// SpinWaitDelay 自旋等待延迟
	// SpinWaitDelay spin wait delay
	SpinWaitDelay int `yaml:"spin_wait_delay" json:"spin_wait_delay" default:"6"`

	// ConcurrencyTickets 并发票据数
	// ConcurrencyTickets concurrency tickets
	ConcurrencyTickets int `yaml:"concurrency_tickets" json:"concurrency_tickets" default:"5000"`

	// ThreadConcurrency 线程并发数
	// ThreadConcurrency thread concurrency
	ThreadConcurrency int `yaml:"thread_concurrency" json:"thread_concurrency" default:"10"`

	// ReadIOThreads 读IO线程数
	// ReadIOThreads read IO threads
	ReadIOThreads int `yaml:"read_io_threads" json:"read_io_threads" default:"4"`

	// WriteIOThreads 写IO线程数
	// WriteIOThreads write IO threads
	WriteIOThreads int `yaml:"write_io_threads" json:"write_io_threads" default:"4"`

	// MaxPreparedStmtCount 最大预处理语句数
	// MaxPreparedStmtCount maximum prepared statement count
	MaxPreparedStmtCount int `yaml:"max_prepared_stmt_count" json:"max_prepared_stmt_count" default:"16382"`

	// EnableQueryResponseTimeStats 启用查询响应时间统计
	// EnableQueryResponseTimeStats enable query response time stats
	EnableQueryResponseTimeStats bool `yaml:"enable_query_response_time_stats" json:"enable_query_response_time_stats" default:"false"`

	// EnableMemoryProfiling 启用内存分析
	// EnableMemoryProfiling enable memory profiling
	EnableMemoryProfiling bool `yaml:"enable_memory_profiling" json:"enable_memory_profiling" default:"false"`

	// EnableCPUProfiling 启用CPU分析
	// EnableCPUProfiling enable CPU profiling
	EnableCPUProfiling bool `yaml:"enable_cpu_profiling" json:"enable_cpu_profiling" default:"false"`

	// ProfilingPath 分析文件路径
	// ProfilingPath profiling path
	ProfilingPath string `yaml:"profiling_path" json:"profiling_path" default:"./profiling"`
}

// ClusterConfig 集群配置
// ClusterConfig cluster configuration
type ClusterConfig struct {
	// Enable 是否启用集群
	// Enable enable cluster
	Enable bool `yaml:"enable" json:"enable" default:"false"`

	// NodeID 节点ID
	// NodeID node ID
	NodeID string `yaml:"node_id" json:"node_id"`

	// ClusterName 集群名称
	// ClusterName cluster name
	ClusterName string `yaml:"cluster_name" json:"cluster_name" default:"guocedb-cluster"`

	// BindAddr 绑定地址
	// BindAddr bind address
	BindAddr string `yaml:"bind_addr" json:"bind_addr"`

	// AdvertiseAddr 广播地址
	// AdvertiseAddr advertise address
	AdvertiseAddr string `yaml:"advertise_addr" json:"advertise_addr"`

	// Seeds 种子节点
	// Seeds seed nodes
	Seeds []string `yaml:"seeds" json:"seeds"`

	// GossipPort Gossip端口
	// GossipPort gossip port
	GossipPort int `yaml:"gossip_port" json:"gossip_port" default:"7000"`

	// RaftPort Raft端口
	// RaftPort raft port
	RaftPort int `yaml:"raft_port" json:"raft_port" default:"7001"`

	// DataShards 数据分片数
	// DataShards data shards
	DataShards int `yaml:"data_shards" json:"data_shards" default:"128"`

	// ReplicationFactor 复制因子
	// ReplicationFactor replication factor
	ReplicationFactor int `yaml:"replication_factor" json:"replication_factor" default:"3"`

	// ConsistencyLevel 一致性级别
	// ConsistencyLevel consistency level
	ConsistencyLevel string `yaml:"consistency_level" json:"consistency_level" default:"QUORUM"`

	// HeartbeatInterval 心跳间隔
	// HeartbeatInterval heartbeat interval
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" json:"heartbeat_interval" default:"1s"`

	// ElectionTimeout 选举超时
	// ElectionTimeout election timeout
	ElectionTimeout time.Duration `yaml:"election_timeout" json:"election_timeout" default:"3s"`

	// SnapshotInterval 快照间隔
	// SnapshotInterval snapshot interval
	SnapshotInterval time.Duration `yaml:"snapshot_interval" json:"snapshot_interval" default:"5m"`

	// CompactionInterval 压缩间隔
	// CompactionInterval compaction interval
	CompactionInterval time.Duration `yaml:"compaction_interval" json:"compaction_interval" default:"1h"`

	// EnableAutoFailover 启用自动故障转移
	// EnableAutoFailover enable auto failover
	EnableAutoFailover bool `yaml:"enable_auto_failover" json:"enable_auto_failover" default:"true"`

	// EnableLoadBalancing 启用负载均衡
	// EnableLoadBalancing enable load balancing
	EnableLoadBalancing bool `yaml:"enable_load_balancing" json:"enable_load_balancing" default:"true"`

	// LoadBalancingPolicy 负载均衡策略
	// LoadBalancingPolicy load balancing policy
	LoadBalancingPolicy string `yaml:"load_balancing_policy" json:"load_balancing_policy" default:"round-robin"`
}

// DevelopmentConfig 开发配置
// DevelopmentConfig development configuration
type DevelopmentConfig struct {
	// DebugMode 调试模式
	// DebugMode debug mode
	DebugMode bool `yaml:"debug_mode" json:"debug_mode" default:"false"`

	// EnablePprof 启用pprof
	// EnablePprof enable pprof
	EnablePprof bool `yaml:"enable_pprof" json:"enable_pprof" default:"false"`

	// PprofPort pprof端口
	// PprofPort pprof port
	PprofPort int `yaml:"pprof_port" json:"pprof_port" default:"6060"`

	// EnablePlayground 启用SQL playground
	// EnablePlayground enable SQL playground
	EnablePlayground bool `yaml:"enable_playground" json:"enable_playground" default:"false"`

	// PlaygroundPort playground端口
	// PlaygroundPort playground port
	PlaygroundPort int `yaml:"playground_port" json:"playground_port" default:"8080"`

	// MockDataEnabled 启用模拟数据
	// MockDataEnabled enable mock data
	MockDataEnabled bool `yaml:"mock_data_enabled" json:"mock_data_enabled" default:"false"`

	// MockDataSize 模拟数据大小
	// MockDataSize mock data size
	MockDataSize int `yaml:"mock_data_size" json:"mock_data_size" default:"1000"`

	// EnableExplainAnalyze 启用EXPLAIN ANALYZE
	// EnableExplainAnalyze enable EXPLAIN ANALYZE
	EnableExplainAnalyze bool `yaml:"enable_explain_analyze" json:"enable_explain_analyze" default:"true"`

	// TestMode 测试模式
	// TestMode test mode
	TestMode bool `yaml:"test_mode" json:"test_mode" default:"false"`

	// TestDataDir 测试数据目录
	// TestDataDir test data directory
	TestDataDir string `yaml:"test_data_dir" json:"test_data_dir" default:"./test_data"`
}

// ===== 全局变量 Global Variables =====

var (
	// globalConfig 全局配置实例
	// globalConfig global configuration instance
	globalConfig *Config

	// configMutex 配置互斥锁
	// configMutex configuration mutex
	configMutex sync.RWMutex

	// configFile 配置文件路径
	// configFile configuration file path
	configFile string

	// configWatcher 配置文件监视器
	// configWatcher configuration file watcher
	configWatcher *viper.Viper
)

// ===== 初始化函数 Initialization Functions =====

// Init 初始化配置系统
// Init initializes configuration system
func Init(cfgFile string) error {
	configFile = cfgFile

	// 创建默认配置
	// Create default configuration
	globalConfig = NewDefaultConfig()

	// 如果指定了配置文件，加载它
	// If configuration file is specified, load it
	if configFile != "" {
		if err := LoadFromFile(configFile); err != nil {
			return errors.Wrap(err, errors.ErrConfigLoad, "failed to load config file")
		}
	}

	// 从环境变量加载配置
	// Load configuration from environment variables
	if err := LoadFromEnv(); err != nil {
		return errors.Wrap(err, errors.ErrConfigLoad, "failed to load config from env")
	}

	// 验证配置
	// Validate configuration
	if err := Validate(); err != nil {
		return errors.Wrap(err, errors.ErrConfigValidation, "configuration validation failed")
	}

	// 初始化配置监视器
	// Initialize configuration watcher
	if configFile != "" && !globalConfig.Development.TestMode {
		if err := initWatcher(); err != nil {
			return errors.Wrap(err, errors.ErrConfigLoad, "failed to init config watcher")
		}
	}

	return nil
}

// NewDefaultConfig 创建默认配置
// NewDefaultConfig creates default configuration
func NewDefaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:              "localhost",
			Port:              3306,
			MaxConnections:    1000,
			ConnectionTimeout: 30 * time.Second,
			ReadTimeout:       300 * time.Second,
			WriteTimeout:      300 * time.Second,
			KeepAlive:         60 * time.Second,
			MaxPacketSize:     67108864, // 64MB
			SocketPath:        "/tmp/guocedb.sock",
			EnableTCP:         true,
			EnableUnixSocket:  runtime.GOOS != "windows",
			ServerID:          1,
			ServerUUID:        utils.GenerateUUID(),
		},
		Storage: StorageConfig{
			Engine:              "innodb",
			DataDir:             "./data",
			TempDir:             "./tmp",
			MaxOpenFiles:        10000,
			PageSize:            16384,     // 16KB
			BufferPoolSize:      134217728, // 128MB
			LogBufferSize:       16777216,  // 16MB
			FlushLogAtCommit:    1,
			DoubleWrite:         true,
			Compression:         "zstd",
			ChecksumAlgorithm:   "crc32",
			FilePerTable:        true,
			AutoExtendIncrement: 67108864, // 64MB
			SyncBinlog:          1,
			BinlogFormat:        "ROW",
			BinlogCacheSize:     32768, // 32KB
			ExpireDays:          0,
		},
		Cache: CacheConfig{
			QueryCacheEnabled:        true,
			QueryCacheSize:           67108864, // 64MB
			QueryCacheLimit:          1048576,  // 1MB
			TableCacheSize:           2000,
			TableDefinitionCacheSize: 1400,
			KeyBufferSize:            8388608, // 8MB
			SortBufferSize:           262144,  // 256KB
			JoinBufferSize:           262144,  // 256KB
			ReadBufferSize:           131072,  // 128KB
			ReadRndBufferSize:        262144,  // 256KB
			PreparedStmtCount:        16382,
			CacheTTL:                 time.Hour,
			CacheEvictionPolicy:      "lru",
		},
		Query: QueryConfig{
			MaxExecutionTime:     0,
			MaxJoinSize:          18446744073709551615,
			MaxLengthForSortData: 1024,
			MaxSortLength:        1024,
			TmpTableSize:         16777216, // 16MB
			MaxHeapTableSize:     16777216, // 16MB
			GroupConcatMaxLen:    1024,
			OptimizerSearchDepth: 62,
			OptimizerPruneLevel:  1,
			EnableQueryCache:     true,
			EnableSlowQueryLog:   true,
			SlowQueryThreshold:   10 * time.Second,
			EnableQueryProfiling: false,
			ParallelDegree:       0,
			EnableCBO:            true,
			EnableIndexMerge:     true,
			EnableHashJoin:       true,
			EnableSortMergeJoin:  true,
		},
		Transaction: TransactionConfig{
			IsolationLevel:         "REPEATABLE-READ",
			AutoCommit:             true,
			LockWaitTimeout:        50 * time.Second,
			DeadlockDetect:         true,
			DeadlockDetectInterval: 100 * time.Millisecond,
			MaxTransactionSize:     104857600, // 100MB
			MaxTransactionTime:     3600 * time.Second,
			MaxUndoLogSize:         1073741824, // 1GB
			PurgeInterval:          10 * time.Second,
			MaxPurgeLag:            0,
			EnableMVCC:             true,
			ReadViewCacheSize:      1000,
		},
		Replication: ReplicationConfig{
			Enable:              false,
			Role:                "master",
			MasterPort:          3306,
			SlaveNetTimeout:     60 * time.Second,
			ConnectRetry:        60 * time.Second,
			MaxRelayLogSize:     1073741824, // 1GB
			RelayLogPurge:       true,
			SemiSyncEnabled:     false,
			SemiSyncTimeout:     10 * time.Second,
			ParallelWorkers:     4,
			PreserveCommitOrder: true,
		},
		Security: SecurityConfig{
			RequireSecureTransport:    false,
			TLSEnabled:                false,
			TLSMinVersion:             "1.2",
			PasswordValidationEnabled: true,
			PasswordMinLength:         8,
			PasswordRequireUppercase:  true,
			PasswordRequireLowercase:  true,
			PasswordRequireNumbers:    true,
			PasswordRequireSpecial:    true,
			MaxFailedLogins:           3,
			AccountLockTime:           30 * time.Minute,
			AuditEnabled:              false,
			AuditLogFile:              "./logs/audit.log",
		},
		Monitoring: MonitoringConfig{
			Enable:            true,
			MetricsEndpoint:   "/metrics",
			MetricsPort:       9090,
			CollectInterval:   15 * time.Second,
			RetentionPeriod:   168 * time.Hour, // 7 days
			EnablePrometheus:  true,
			EnableStatsD:      false,
			StatsDHost:        "localhost:8125",
			EnableTracing:     false,
			TracingSampleRate: 0.1,
		},
		Logging: LoggingConfig{
			Level:            "info",
			Format:           "json",
			Output:           []string{"stdout"},
			Directory:        "./logs",
			Filename:         "guocedb.log",
			MaxSize:          100,
			MaxBackups:       30,
			MaxAge:           30,
			Compress:         true,
			EnableErrorLog:   true,
			ErrorLogFile:     "error.log",
			EnableSlowLog:    true,
			SlowLogFile:      "slow.log",
			EnableGeneralLog: false,
			GeneralLogFile:   "general.log",
			EnableAuditLog:   false,
			AuditLogFile:     "audit.log",
			SamplingRate:     1.0,
			EnableCaller:     false,
			EnableStacktrace: false,
		},
		Performance: PerformanceConfig{
			MaxCPU:                       0,
			ThreadPoolSize:               16,
			IOThreads:                    4,
			WorkerThreads:                16,
			EnableAdaptiveHashIndex:      true,
			EnableChangeBuffer:           true,
			ChangeBufferMaxSize:          25,
			SpinWaitDelay:                6,
			ConcurrencyTickets:           5000,
			ThreadConcurrency:            10,
			ReadIOThreads:                4,
			WriteIOThreads:               4,
			MaxPreparedStmtCount:         16382,
			EnableQueryResponseTimeStats: false,
			EnableMemoryProfiling:        false,
			EnableCPUProfiling:           false,
			ProfilingPath:                "./profiling",
		},
		Cluster: ClusterConfig{
			Enable:              false,
			ClusterName:         "guocedb-cluster",
			GossipPort:          7000,
			RaftPort:            7001,
			DataShards:          128,
			ReplicationFactor:   3,
			ConsistencyLevel:    "QUORUM",
			HeartbeatInterval:   time.Second,
			ElectionTimeout:     3 * time.Second,
			SnapshotInterval:    5 * time.Minute,
			CompactionInterval:  time.Hour,
			EnableAutoFailover:  true,
			EnableLoadBalancing: true,
			LoadBalancingPolicy: "round-robin",
		},
		Development: DevelopmentConfig{
			DebugMode:            false,
			EnablePprof:          false,
			PprofPort:            6060,
			EnablePlayground:     false,
			PlaygroundPort:       8080,
			MockDataEnabled:      false,
			MockDataSize:         1000,
			EnableExplainAnalyze: true,
			TestMode:             false,
			TestDataDir:          "./test_data",
		},
	}

	// 设置一些依赖于操作系统的默认值
	// Set some OS-dependent defaults
	if runtime.GOOS == "windows" {
		cfg.Storage.DataDir = "C:\\guocedb\\data"
		cfg.Storage.TempDir = "C:\\guocedb\\tmp"
		cfg.Logging.Directory = "C:\\guocedb\\logs"
	}

	// 根据可用CPU核心数调整默认值
	// Adjust defaults based on available CPU cores
	numCPU := runtime.NumCPU()
	if numCPU > 0 {
		cfg.Performance.MaxCPU = numCPU
		cfg.Performance.ThreadPoolSize = numCPU * 2
		cfg.Performance.WorkerThreads = numCPU * 2
		cfg.Performance.IOThreads = numCPU
		if numCPU > 4 {
			cfg.Performance.ReadIOThreads = numCPU / 2
			cfg.Performance.WriteIOThreads = numCPU / 2
		}
	}

	return cfg
}

// LoadFromFile 从文件加载配置
// LoadFromFile loads configuration from file
func LoadFromFile(filename string) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 检查文件是否存在
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New(errors.ErrFileNotFound, "config file not found: %s", filename)
	}

	// 根据文件扩展名选择解析器
	// Choose parser based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	var err error

	switch ext {
	case ".yaml", ".yml":
		err = loadFromYAML(filename)
	case ".json":
		err = loadFromJSON(filename)
	case ".toml":
		err = loadFromTOML(filename)
	default:
		return errors.New(errors.ErrUnsupported, "unsupported config file format: %s", ext)
	}

	if err != nil {
		return err
	}

	// 保存配置文件路径
	// Save configuration file path
	configFile = filename
	return nil
}

// loadFromYAML 从YAML文件加载配置
// loadFromYAML loads configuration from YAML file
func loadFromYAML(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileRead, "failed to read YAML file")
	}

	if err := yaml.Unmarshal(data, globalConfig); err != nil {
		return errors.Wrap(err, errors.ErrConfigParse, "failed to parse YAML")
	}

	return nil
}

// loadFromJSON 从JSON文件加载配置
// loadFromJSON loads configuration from JSON file
func loadFromJSON(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileRead, "failed to read JSON file")
	}

	if err := json.Unmarshal(data, globalConfig); err != nil {
		return errors.Wrap(err, errors.ErrConfigParse, "failed to parse JSON")
	}

	return nil
}

// loadFromTOML 从TOML文件加载配置
// loadFromTOML loads configuration from TOML file
func loadFromTOML(filename string) error {
	v := viper.New()
	v.SetConfigFile(filename)

	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, errors.ErrFileRead, "failed to read TOML file")
	}

	if err := v.Unmarshal(globalConfig); err != nil {
		return errors.Wrap(err, errors.ErrConfigParse, "failed to parse TOML")
	}

	return nil
}

// LoadFromEnv 从环境变量加载配置
// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	v := viper.New()
	v.SetEnvPrefix("GUOCEDB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 映射环境变量到配置字段
	// Map environment variables to configuration fields
	envMappings := map[string]string{
		"SERVER_HOST":               "server.host",
		"SERVER_PORT":               "server.port",
		"SERVER_MAX_CONNECTIONS":    "server.max_connections",
		"STORAGE_ENGINE":            "storage.engine",
		"STORAGE_DATA_DIR":          "storage.data_dir",
		"STORAGE_BUFFER_POOL_SIZE":  "storage.buffer_pool_size",
		"CACHE_QUERY_CACHE_ENABLED": "cache.query_cache_enabled",
		"CACHE_QUERY_CACHE_SIZE":    "cache.query_cache_size",
		"LOGGING_LEVEL":             "logging.level",
		"LOGGING_FORMAT":            "logging.format",
		"SECURITY_TLS_ENABLED":      "security.tls_enabled",
		"SECURITY_TLS_CERT":         "security.tls_cert",
		"SECURITY_TLS_KEY":          "security.tls_key",
		"CLUSTER_ENABLE":            "cluster.enable",
		"CLUSTER_NODE_ID":           "cluster.node_id",
		"DEVELOPMENT_DEBUG_MODE":    "development.debug_mode",
	}

	for envKey, configKey := range envMappings {
		if value := os.Getenv("GUOCEDB_" + envKey); value != "" {
			v.Set(configKey, value)
		}
	}

	// 合并环境变量配置
	// Merge environment variable configuration
	if err := v.Unmarshal(globalConfig); err != nil {
		return errors.Wrap(err, errors.ErrConfigParse, "failed to parse env config")
	}

	return nil
}

// Validate 验证配置
// Validate validates configuration
func Validate() error {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 验证服务器配置
	// Validate server configuration
	if err := validateServerConfig(&globalConfig.Server); err != nil {
		return err
	}

	// 验证存储配置
	// Validate storage configuration
	if err := validateStorageConfig(&globalConfig.Storage); err != nil {
		return err
	}

	// 验证缓存配置
	// Validate cache configuration
	if err := validateCacheConfig(&globalConfig.Cache); err != nil {
		return err
	}

	// 验证查询配置
	// Validate query configuration
	if err := validateQueryConfig(&globalConfig.Query); err != nil {
		return err
	}

	// 验证事务配置
	// Validate transaction configuration
	if err := validateTransactionConfig(&globalConfig.Transaction); err != nil {
		return err
	}

	// 验证安全配置
	// Validate security configuration
	if err := validateSecurityConfig(&globalConfig.Security); err != nil {
		return err
	}

	// 验证集群配置
	// Validate cluster configuration
	if err := validateClusterConfig(&globalConfig.Cluster); err != nil {
		return err
	}

	return nil
}

// validateServerConfig 验证服务器配置
// validateServerConfig validates server configuration
func validateServerConfig(cfg *ServerConfig) error {
	// 验证端口范围
	// Validate port range
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New(errors.ErrConfigValidation, "invalid server port: %d", cfg.Port)
	}

	// 验证最大连接数
	// Validate maximum connections
	if cfg.MaxConnections < 1 {
		return errors.New(errors.ErrConfigValidation, "max_connections must be at least 1")
	}

	// 验证超时时间
	// Validate timeouts
	if cfg.ConnectionTimeout <= 0 {
		return errors.New(errors.ErrConfigValidation, "connection_timeout must be positive")
	}

	// 验证监听地址
	// Validate listen address
	if cfg.Host != "" && cfg.Host != "localhost" {
		if ip := net.ParseIP(cfg.Host); ip == nil {
			// 尝试解析为主机名
			// Try to resolve as hostname
			if _, err := net.LookupHost(cfg.Host); err != nil {
				return errors.New(errors.ErrConfigValidation, "invalid server host: %s", cfg.Host)
			}
		}
	}

	// 验证数据包大小
	// Validate packet size
	if cfg.MaxPacketSize < 1024 {
		return errors.New(errors.ErrConfigValidation, "max_packet_size too small: %d", cfg.MaxPacketSize)
	}

	// 生成UUID如果没有设置
	// Generate UUID if not set
	if cfg.ServerUUID == "" {
		cfg.ServerUUID = utils.GenerateUUID()
	}

	return nil
}

// validateStorageConfig 验证存储配置
// validateStorageConfig validates storage configuration
func validateStorageConfig(cfg *StorageConfig) error {
	// 验证存储引擎
	// Validate storage engine
	validEngines := map[string]bool{
		"innodb": true,
		"memory": true,
		"csv":    true,
	}
	if !validEngines[strings.ToLower(cfg.Engine)] {
		return errors.New(errors.ErrConfigValidation, "invalid storage engine: %s", cfg.Engine)
	}

	// 验证数据目录
	// Validate data directory
	if cfg.DataDir == "" {
		return errors.New(errors.ErrConfigValidation, "data_dir cannot be empty")
	}

	// 验证页大小
	// Validate page size
	validPageSizes := []int{4096, 8192, 16384, 32768, 65536}
	validPageSize := false
	for _, size := range validPageSizes {
		if cfg.PageSize == size {
			validPageSize = true
			break
		}
	}
	if !validPageSize {
		return errors.New(errors.ErrConfigValidation, "invalid page_size: %d", cfg.PageSize)
	}

	// 验证缓冲池大小
	// Validate buffer pool size
	if cfg.BufferPoolSize < 5*1024*1024 { // 最小5MB
		return errors.New(errors.ErrConfigValidation, "buffer_pool_size too small: %d", cfg.BufferPoolSize)
	}

	// 验证压缩算法
	// Validate compression algorithm
	validCompressions := map[string]bool{
		"none": true,
		"zlib": true,
		"lz4":  true,
		"zstd": true,
	}
	if !validCompressions[strings.ToLower(cfg.Compression)] {
		return errors.New(errors.ErrConfigValidation, "invalid compression: %s", cfg.Compression)
	}

	// 验证binlog格式
	// Validate binlog format
	validBinlogFormats := map[string]bool{
		"ROW":       true,
		"STATEMENT": true,
		"MIXED":     true,
	}
	if !validBinlogFormats[strings.ToUpper(cfg.BinlogFormat)] {
		return errors.New(errors.ErrConfigValidation, "invalid binlog_format: %s", cfg.BinlogFormat)
	}

	return nil
}

// validateCacheConfig 验证缓存配置
// validateCacheConfig validates cache configuration
func validateCacheConfig(cfg *CacheConfig) error {
	// 验证缓存大小
	// Validate cache sizes
	if cfg.QueryCacheSize < 0 {
		return errors.New(errors.ErrConfigValidation, "query_cache_size cannot be negative")
	}

	if cfg.QueryCacheLimit > cfg.QueryCacheSize {
		return errors.New(errors.ErrConfigValidation, "query_cache_limit cannot exceed query_cache_size")
	}

	// 验证缓存淘汰策略
	// Validate cache eviction policy
	validPolicies := map[string]bool{
		"lru":    true,
		"lfu":    true,
		"fifo":   true,
		"random": true,
	}
	if !validPolicies[strings.ToLower(cfg.CacheEvictionPolicy)] {
		return errors.New(errors.ErrConfigValidation, "invalid cache_eviction_policy: %s", cfg.CacheEvictionPolicy)
	}

	// 验证缓冲区大小
	// Validate buffer sizes
	if cfg.SortBufferSize < 32*1024 { // 最小32KB
		return errors.New(errors.ErrConfigValidation, "sort_buffer_size too small: %d", cfg.SortBufferSize)
	}

	return nil
}

// validateQueryConfig 验证查询配置
// validateQueryConfig validates query configuration
func validateQueryConfig(cfg *QueryConfig) error {
	// 验证优化器参数
	// Validate optimizer parameters
	if cfg.OptimizerSearchDepth < 0 || cfg.OptimizerSearchDepth > 62 {
		return errors.New(errors.ErrConfigValidation, "optimizer_search_depth must be between 0 and 62")
	}

	// 验证并行度
	// Validate parallel degree
	if cfg.ParallelDegree < 0 {
		return errors.New(errors.ErrConfigValidation, "parallel_degree cannot be negative")
	}

	// 验证临时表大小
	// Validate temporary table size
	if cfg.TmpTableSize < 1024 {
		return errors.New(errors.ErrConfigValidation, "tmp_table_size too small: %d", cfg.TmpTableSize)
	}

	return nil
}

// validateTransactionConfig 验证事务配置
// validateTransactionConfig validates transaction configuration
func validateTransactionConfig(cfg *TransactionConfig) error {
	// 验证隔离级别
	// Validate isolation level
	validLevels := map[string]bool{
		"READ-UNCOMMITTED": true,
		"READ-COMMITTED":   true,
		"REPEATABLE-READ":  true,
		"SERIALIZABLE":     true,
	}
	if !validLevels[strings.ToUpper(cfg.IsolationLevel)] {
		return errors.New(errors.ErrConfigValidation, "invalid isolation_level: %s", cfg.IsolationLevel)
	}

	// 验证超时时间
	// Validate timeouts
	if cfg.LockWaitTimeout <= 0 {
		return errors.New(errors.ErrConfigValidation, "lock_wait_timeout must be positive")
	}

	// 验证事务大小
	// Validate transaction size
	if cfg.MaxTransactionSize < 1024*1024 { // 最小1MB
		return errors.New(errors.ErrConfigValidation, "max_transaction_size too small: %d", cfg.MaxTransactionSize)
	}

	return nil
}

// validateSecurityConfig 验证安全配置
// validateSecurityConfig validates security configuration
func validateSecurityConfig(cfg *SecurityConfig) error {
	// 如果启用TLS，验证证书文件
	// If TLS is enabled, validate certificate files
	if cfg.TLSEnabled {
		if cfg.TLSCert == "" || cfg.TLSKey == "" {
			return errors.New(errors.ErrConfigValidation, "TLS cert and key must be specified when TLS is enabled")
		}

		// 验证TLS版本
		// Validate TLS version
		validVersions := map[string]bool{
			"1.0": true,
			"1.1": true,
			"1.2": true,
			"1.3": true,
		}
		if !validVersions[cfg.TLSMinVersion] {
			return errors.New(errors.ErrConfigValidation, "invalid tls_min_version: %s", cfg.TLSMinVersion)
		}
	}

	// 验证密码策略
	// Validate password policy
	if cfg.PasswordValidationEnabled {
		if cfg.PasswordMinLength < 4 {
			return errors.New(errors.ErrConfigValidation, "password_min_length too short: %d", cfg.PasswordMinLength)
		}
	}

	// 验证IP列表格式
	// Validate IP list format
	for _, ip := range cfg.IPWhitelist {
		if _, _, err := net.ParseCIDR(ip); err != nil {
			if net.ParseIP(ip) == nil {
				return errors.New(errors.ErrConfigValidation, "invalid IP in whitelist: %s", ip)
			}
		}
	}

	return nil
}

// validateClusterConfig 验证集群配置
// validateClusterConfig validates cluster configuration
func validateClusterConfig(cfg *ClusterConfig) error {
	if !cfg.Enable {
		return nil
	}

	// 验证节点ID
	// Validate node ID
	if cfg.NodeID == "" {
		return errors.New(errors.ErrConfigValidation, "node_id cannot be empty in cluster mode")
	}

	// 验证端口
	// Validate ports
	if cfg.GossipPort < 1 || cfg.GossipPort > 65535 {
		return errors.New(errors.ErrConfigValidation, "invalid gossip_port: %d", cfg.GossipPort)
	}

	if cfg.RaftPort < 1 || cfg.RaftPort > 65535 {
		return errors.New(errors.ErrConfigValidation, "invalid raft_port: %d", cfg.RaftPort)
	}

	// 验证复制因子
	// Validate replication factor
	if cfg.ReplicationFactor < 1 {
		return errors.New(errors.ErrConfigValidation, "replication_factor must be at least 1")
	}

	// 验证一致性级别
	// Validate consistency level
	validLevels := map[string]bool{
		"ONE":    true,
		"QUORUM": true,
		"ALL":    true,
	}
	if !validLevels[strings.ToUpper(cfg.ConsistencyLevel)] {
		return errors.New(errors.ErrConfigValidation, "invalid consistency_level: %s", cfg.ConsistencyLevel)
	}

	return nil
}

// ===== 配置访问函数 Configuration Access Functions =====

// Get 获取全局配置
// Get returns global configuration
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// GetServer 获取服务器配置
// GetServer returns server configuration
func GetServer() ServerConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Server
}

// GetStorage 获取存储配置
// GetStorage returns storage configuration
func GetStorage() StorageConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Storage
}

// GetCache 获取缓存配置
// GetCache returns cache configuration
func GetCache() CacheConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Cache
}

// GetQuery 获取查询配置
// GetQuery returns query configuration
func GetQuery() QueryConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Query
}

// GetTransaction 获取事务配置
// GetTransaction returns transaction configuration
func GetTransaction() TransactionConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Transaction
}

// GetSecurity 获取安全配置
// GetSecurity returns security configuration
func GetSecurity() SecurityConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Security
}

// GetLogging 获取日志配置
// GetLogging returns logging configuration
func GetLogging() LoggingConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Logging
}

// GetCluster 获取集群配置
// GetCluster returns cluster configuration
func GetCluster() ClusterConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Cluster
}

// IsDebugMode 判断是否为调试模式
// IsDebugMode checks if debug mode is enabled
func IsDebugMode() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Development.DebugMode
}

// IsClusterMode 判断是否为集群模式
// IsClusterMode checks if cluster mode is enabled
func IsClusterMode() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig.Cluster.Enable
}

// ===== 配置更新函数 Configuration Update Functions =====

// Update 更新配置
// Update updates configuration
func Update(updater func(*Config)) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 创建配置副本
	// Create configuration copy
	newConfig := *globalConfig

	// 应用更新
	// Apply updates
	updater(&newConfig)

	// 验证新配置
	// Validate new configuration
	tempConfig := globalConfig
	globalConfig = &newConfig
	if err := Validate(); err != nil {
		globalConfig = tempConfig
		return err
	}

	// 触发配置更新回调
	// Trigger configuration update callbacks
	notifyConfigChange(&newConfig)

	return nil
}

// Reload 重新加载配置
// Reload reloads configuration
func Reload() error {
	if configFile == "" {
		return errors.New(errors.ErrConfigLoad, "no config file specified")
	}

	return LoadFromFile(configFile)
}

// ===== 配置监视器 Configuration Watcher =====

// initWatcher 初始化配置文件监视器
// initWatcher initializes configuration file watcher
func initWatcher() error {
	configWatcher = viper.New()
	configWatcher.SetConfigFile(configFile)

	configWatcher.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		if err := Reload(); err != nil {
			fmt.Printf("Failed to reload config: %v\n", err)
		}
	})

	configWatcher.WatchConfig()
	return nil
}

// ===== 配置导出函数 Configuration Export Functions =====

// Export 导出配置
// Export exports configuration
func Export(format string) ([]byte, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	switch strings.ToLower(format) {
	case "yaml", "yml":
		return yaml.Marshal(globalConfig)
	case "json":
		return json.MarshalIndent(globalConfig, "", "  ")
	default:
		return nil, errors.New(errors.ErrUnsupported, "unsupported export format: %s", format)
	}
}

// Save 保存配置到文件
// Save saves configuration to file
func Save(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	format := strings.TrimPrefix(ext, ".")

	data, err := Export(format)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// ===== 配置变更通知 Configuration Change Notification =====

var (
	// configChangeCallbacks 配置变更回调
	// configChangeCallbacks configuration change callbacks
	configChangeCallbacks []func(*Config)
	callbackMutex         sync.RWMutex
)

// OnConfigChange 注册配置变更回调
// OnConfigChange registers configuration change callback
func OnConfigChange(callback func(*Config)) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	configChangeCallbacks = append(configChangeCallbacks, callback)
}

// notifyConfigChange 通知配置变更
// notifyConfigChange notifies configuration change
func notifyConfigChange(cfg *Config) {
	callbackMutex.RLock()
	defer callbackMutex.RUnlock()

	for _, callback := range configChangeCallbacks {
		go callback(cfg)
	}
}

// ===== 测试辅助函数 Test Helper Functions =====

// SetTestConfig 设置测试配置
// SetTestConfig sets test configuration
func SetTestConfig(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}

// ResetToDefault 重置为默认配置
// ResetToDefault resets to default configuration
func ResetToDefault() {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = NewDefaultConfig()
}
