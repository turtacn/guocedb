// Package sal implements Storage Abstraction Layer for GuoceDB
// 存储抽象层包，为GuoceDB提供统一的存储接口
package sal

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	logging "github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types"
	storage "github.com/turtacn/guocedb/interfaces"
	"github.com/turtacn/guocedb/maintenance/metrics"
)

// StorageAdapter 存储适配器结构 Storage adapter structure
type StorageAdapter struct {
	engines       map[string]storage.Engine // 存储引擎映射 Storage engine mapping
	engineMutex   sync.RWMutex              // 引擎读写锁 Engine read-write lock
	factory       storage.EngineFactory     // 引擎工厂 Engine factory
	router        *RequestRouter            // 路由器 Router
	monitor       *PerformanceMonitor       // 性能监控器 Performance monitor
	healthChecker *HealthChecker            // 健康检查器 Health checker
	config        *AdapterConfig            // 适配器配置 Adapter configuration
	logger        logging.Logger            // 日志器 Logger
	metrics       metrics.Collector         // 指标收集器 Metrics collector

	// 运行状态 Runtime state
	running      int32     // 运行状态标志 Running state flag
	startTime    time.Time // 启动时间 Start time
	requestCount int64     // 请求计数 Request count
	errorCount   int64     // 错误计数 Error count

	// 事件监听器 Event listeners
	listeners     []EventListener // 事件监听器列表 Event listener list
	listenerMutex sync.RWMutex    // 监听器读写锁 Listener read-write lock
}

// AdapterConfig 适配器配置 Adapter configuration
type AdapterConfig struct {
	// 路由配置 Routing configuration
	RoutingPolicy     RoutingPolicy     `json:"routing_policy"`      // 路由策略 Routing policy
	LoadBalanceMethod LoadBalanceMethod `json:"load_balance_method"` // 负载均衡方法 Load balance method
	FailoverStrategy  FailoverStrategy  `json:"failover_strategy"`   // 故障转移策略 Failover strategy

	// 监控配置 Monitoring configuration
	EnableMonitoring   bool          `json:"enable_monitoring"`   // 启用监控 Enable monitoring
	MonitoringInterval time.Duration `json:"monitoring_interval"` // 监控间隔 Monitoring interval
	MetricsRetention   time.Duration `json:"metrics_retention"`   // 指标保留时间 Metrics retention

	// 健康检查配置 Health check configuration
	HealthCheckEnabled  bool          `json:"health_check_enabled"`  // 启用健康检查 Enable health check
	HealthCheckInterval time.Duration `json:"health_check_interval"` // 健康检查间隔 Health check interval
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`  // 健康检查超时 Health check timeout

	// 性能配置 Performance configuration
	MaxConcurrentRequests int           `json:"max_concurrent_requests"` // 最大并发请求数 Max concurrent requests
	RequestTimeout        time.Duration `json:"request_timeout"`         // 请求超时 Request timeout
	RetryAttempts         int           `json:"retry_attempts"`          // 重试次数 Retry attempts
	RetryDelay            time.Duration `json:"retry_delay"`             // 重试延迟 Retry delay

	// 缓存配置 Cache configuration
	EnableCache bool          `json:"enable_cache"` // 启用缓存 Enable cache
	CacheSize   int           `json:"cache_size"`   // 缓存大小 Cache size
	CacheTTL    time.Duration `json:"cache_ttl"`    // 缓存TTL Cache TTL
}

// RoutingPolicy 路由策略 Routing policy
type RoutingPolicy int

const (
	RoutingPolicyRoundRobin RoutingPolicy = iota // 轮询 Round robin
	RoutingPolicyRandom                          // 随机 Random
	RoutingPolicyWeighted                        // 加权 Weighted
	RoutingPolicyHash                            // 哈希 Hash
	RoutingPolicyLatency                         // 延迟优先 Latency priority
	RoutingPolicyCapacity                        // 容量优先 Capacity priority
)

// LoadBalanceMethod 负载均衡方法 Load balance method
type LoadBalanceMethod int

const (
	LoadBalanceRoundRobin LoadBalanceMethod = iota // 轮询 Round robin
	LoadBalanceLeastConn                           // 最少连接 Least connections
	LoadBalanceWeighted                            // 加权轮询 Weighted round robin
	LoadBalanceResource                            // 资源感知 Resource aware
)

// FailoverStrategy 故障转移策略 Failover strategy
type FailoverStrategy int

const (
	FailoverImmediate      FailoverStrategy = iota // 立即故障转移 Immediate failover
	FailoverGraceful                               // 优雅故障转移 Graceful failover
	FailoverCircuitBreaker                         // 熔断器故障转移 Circuit breaker failover
)

// NewStorageAdapter 创建存储适配器 Create storage adapter
func NewStorageAdapter(config *AdapterConfig, factory storage.EngineFactory) (*StorageAdapter, error) {
	if config == nil {
		return nil, errors.NewError(errors.ErrCodeInvalidParameter, "adapter config is required")
	}

	if factory == nil {
		return nil, errors.NewError(errors.ErrCodeInvalidParameter, "engine factory is required")
	}

	adapter := &StorageAdapter{
		engines:   make(map[string]storage.Engine),
		factory:   factory,
		config:    config,
		startTime: time.Now(),
		logger:    logging.GetLogger("storage.sal.adapter"),
		metrics:   metrics.GetCollector("storage.sal.adapter"),
	}

	// 初始化路由器 Initialize router
	var err error
	adapter.router, err = NewRequestRouter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	// 初始化性能监控器 Initialize performance monitor
	if config.EnableMonitoring {
		adapter.monitor = NewPerformanceMonitor(config.MonitoringInterval, config.MetricsRetention)
	}

	// 初始化健康检查器 Initialize health checker
	if config.HealthCheckEnabled {
		adapter.healthChecker = NewHealthChecker(config.HealthCheckInterval, config.HealthCheckTimeout)
	}

	return adapter, nil
}

// RegisterEngine 注册存储引擎 Register storage engine
func (a *StorageAdapter) RegisterEngine(name string, engine storage.Engine) error {
	if name == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "engine name cannot be empty")
	}

	if engine == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "engine cannot be nil")
	}

	a.engineMutex.Lock()
	defer a.engineMutex.Unlock()

	// 检查引擎是否已存在 Check if engine already exists
	if _, exists := a.engines[name]; exists {
		return errors.NewError(errors.ErrCodeAlreadyExists,
			fmt.Sprintf("engine %s already registered", name))
	}

	// 注册引擎 Register engine
	a.engines[name] = engine

	// 更新路由器 Update router
	if a.router != nil {
		a.router.AddEngine(name, engine)
	}

	// 启动健康检查 Start health check
	if a.healthChecker != nil && atomic.LoadInt32(&a.running) == 1 {
		a.healthChecker.AddEngine(name, engine)
	}

	a.logger.Info("Storage engine registered",
		logging.Field("engine", name),
		logging.Field("type", fmt.Sprintf("%T", engine)))

	// 触发事件 Trigger event
	a.triggerEvent(EventEngineRegistered, map[string]interface{}{
		"engine": name,
		"type":   fmt.Sprintf("%T", engine),
	})

	return nil
}

// UnregisterEngine 注销存储引擎 Unregister storage engine
func (a *StorageAdapter) UnregisterEngine(name string) error {
	if name == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "engine name cannot be empty")
	}

	a.engineMutex.Lock()
	defer a.engineMutex.Unlock()

	// 检查引擎是否存在 Check if engine exists
	engine, exists := a.engines[name]
	if !exists {
		return errors.NewError(errors.ErrCodeNotFound,
			fmt.Sprintf("engine %s not found", name))
	}

	// 停止引擎 Stop engine
	if err := engine.Close(); err != nil {
		a.logger.Warn("Failed to close engine during unregistration",
			logging.Field("engine", name),
			logging.Field("error", err))
	}

	// 从路由器中移除 Remove from router
	if a.router != nil {
		a.router.RemoveEngine(name)
	}

	// 从健康检查中移除 Remove from health checker
	if a.healthChecker != nil {
		a.healthChecker.RemoveEngine(name)
	}

	// 删除引擎 Delete engine
	delete(a.engines, name)

	a.logger.Info("Storage engine unregistered",
		logging.Field("engine", name))

	// 触发事件 Trigger event
	a.triggerEvent(EventEngineUnregistered, map[string]interface{}{
		"engine": name,
	})

	return nil
}

// GetEngine 获取存储引擎 Get storage engine
func (a *StorageAdapter) GetEngine(name string) (storage.Engine, error) {
	a.engineMutex.RLock()
	defer a.engineMutex.RUnlock()

	engine, exists := a.engines[name]
	if !exists {
		return nil, errors.NewError(errors.ErrCodeNotFound,
			fmt.Sprintf("engine %s not found", name))
	}

	return engine, nil
}

// ListEngines 列出所有存储引擎 List all storage engines
func (a *StorageAdapter) ListEngines() []string {
	a.engineMutex.RLock()
	defer a.engineMutex.RUnlock()

	names := make([]string, 0, len(a.engines))
	for name := range a.engines {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Start 启动适配器 Start adapter
func (a *StorageAdapter) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return errors.NewError(errors.ErrCodeInvalidState, "adapter is already running")
	}

	a.logger.Info("Starting storage adapter")

	// 启动所有引擎 Start all engines
	a.engineMutex.RLock()
	engines := make(map[string]storage.Engine)
	for name, engine := range a.engines {
		engines[name] = engine
	}
	a.engineMutex.RUnlock()

	for name, engine := range engines {
		if err := engine.Start(ctx); err != nil {
			a.logger.Error("Failed to start engine",
				logging.Field("engine", name),
				logging.Field("error", err))
			return fmt.Errorf("failed to start engine %s: %w", name, err)
		}
		a.logger.Info("Engine started", logging.Field("engine", name))
	}

	// 启动路由器 Start router
	if a.router != nil {
		if err := a.router.Start(ctx); err != nil {
			return fmt.Errorf("failed to start router: %w", err)
		}
	}

	// 启动性能监控器 Start performance monitor
	if a.monitor != nil {
		a.monitor.Start(ctx)
	}

	// 启动健康检查器 Start health checker
	if a.healthChecker != nil {
		for name, engine := range engines {
			a.healthChecker.AddEngine(name, engine)
		}
		a.healthChecker.Start(ctx)
	}

	a.logger.Info("Storage adapter started successfully")

	// 触发事件 Trigger event
	a.triggerEvent(EventAdapterStarted, nil)

	return nil
}

// Stop 停止适配器 Stop adapter
func (a *StorageAdapter) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&a.running, 1, 0) {
		return errors.NewError(errors.ErrCodeInvalidState, "adapter is not running")
	}

	a.logger.Info("Stopping storage adapter")

	// 停止健康检查器 Stop health checker
	if a.healthChecker != nil {
		a.healthChecker.Stop(ctx)
	}

	// 停止性能监控器 Stop performance monitor
	if a.monitor != nil {
		a.monitor.Stop(ctx)
	}

	// 停止路由器 Stop router
	if a.router != nil {
		a.router.Stop(ctx)
	}

	// 停止所有引擎 Stop all engines
	a.engineMutex.RLock()
	engines := make(map[string]storage.Engine)
	for name, engine := range a.engines {
		engines[name] = engine
	}
	a.engineMutex.RUnlock()

	for name, engine := range engines {
		if err := engine.Stop(ctx); err != nil {
			a.logger.Error("Failed to stop engine",
				logging.Field("engine", name),
				logging.Field("error", err))
		} else {
			a.logger.Info("Engine stopped", logging.Field("engine", name))
		}
	}

	a.logger.Info("Storage adapter stopped successfully")

	// 触发事件 Trigger event
	a.triggerEvent(EventAdapterStopped, nil)

	return nil
}

// CreateDatabase 创建数据库 Create database
func (a *StorageAdapter) CreateDatabase(ctx context.Context, name string, options *types.DatabaseOptions) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("CreateDatabase", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteCreateDatabase(ctx, name, options)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("CreateDatabase", err)
	}

	return engine.CreateDatabase(ctx, name, options)
}

// DropDatabase 删除数据库 Drop database
func (a *StorageAdapter) DropDatabase(ctx context.Context, name string) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("DropDatabase", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteDropDatabase(ctx, name)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("DropDatabase", err)
	}

	return engine.DropDatabase(ctx, name)
}

// CreateTable 创建表 Create table
func (a *StorageAdapter) CreateTable(ctx context.Context, database, name string, schema *types.Schema) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("CreateTable", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteCreateTable(ctx, database, name, schema)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("CreateTable", err)
	}

	return engine.CreateTable(ctx, database, name, schema)
}

// DropTable 删除表 Drop table
func (a *StorageAdapter) DropTable(ctx context.Context, database, name string) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("DropTable", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteDropTable(ctx, database, name)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("DropTable", err)
	}

	return engine.DropTable(ctx, database, name)
}

// Insert 插入数据 Insert data
func (a *StorageAdapter) Insert(ctx context.Context, database, table string, rows []*types.Row) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("Insert", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteInsert(ctx, database, table, rows)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("Insert", err)
	}

	return engine.Insert(ctx, database, table, rows)
}

// Update 更新数据 Update data
func (a *StorageAdapter) Update(ctx context.Context, database, table string, filter *types.Filter, updates map[string]interface{}) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("Update", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteUpdate(ctx, database, table, filter, updates)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("Update", err)
	}

	return engine.Update(ctx, database, table, filter, updates)
}

// Delete 删除数据 Delete data
func (a *StorageAdapter) Delete(ctx context.Context, database, table string, filter *types.Filter) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("Delete", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteDelete(ctx, database, table, filter)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("Delete", err)
	}

	return engine.Delete(ctx, database, table, filter)
}

// Select 查询数据 Select data
func (a *StorageAdapter) Select(ctx context.Context, database, table string, query *types.Query) (*types.ResultSet, error) {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("Select", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteSelect(ctx, database, table, query)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return nil, a.handleError("Select", err)
	}

	return engine.Select(ctx, database, table, query)
}

// CreateIndex 创建索引 Create index
func (a *StorageAdapter) CreateIndex(ctx context.Context, database, table string, index *types.Index) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("CreateIndex", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteCreateIndex(ctx, database, table, index)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("CreateIndex", err)
	}

	return engine.CreateIndex(ctx, database, table, index)
}

// DropIndex 删除索引 Drop index
func (a *StorageAdapter) DropIndex(ctx context.Context, database, table, name string) error {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("DropIndex", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteDropIndex(ctx, database, table, name)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return a.handleError("DropIndex", err)
	}

	return engine.DropIndex(ctx, database, table, name)
}

// BeginTransaction 开始事务 Begin transaction
func (a *StorageAdapter) BeginTransaction(ctx context.Context, options *types.TransactionOptions) (storage.Transaction, error) {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("BeginTransaction", time.Since(startTime))
		}
	}()

	engine, err := a.router.RouteBeginTransaction(ctx, options)
	if err != nil {
		atomic.AddInt64(&a.errorCount, 1)
		return nil, a.handleError("BeginTransaction", err)
	}

	return engine.BeginTransaction(ctx, options)
}

// GetStatistics 获取统计信息 Get statistics
func (a *StorageAdapter) GetStatistics(ctx context.Context) (*storage.Statistics, error) {
	startTime := time.Now()
	atomic.AddInt64(&a.requestCount, 1)

	defer func() {
		if a.monitor != nil {
			a.monitor.RecordRequest("GetStatistics", time.Since(startTime))
		}
	}()

	// 收集所有引擎的统计信息 Collect statistics from all engines
	a.engineMutex.RLock()
	engines := make(map[string]storage.Engine)
	for name, engine := range a.engines {
		engines[name] = engine
	}
	a.engineMutex.RUnlock()

	totalStats := &storage.Statistics{
		Databases:     make(map[string]*storage.DatabaseStatistics),
		TotalRequests: atomic.LoadInt64(&a.requestCount),
		TotalErrors:   atomic.LoadInt64(&a.errorCount),
		Uptime:        time.Since(a.startTime),
	}

	for name, engine := range engines {
		stats, err := engine.GetStatistics(ctx)
		if err != nil {
			a.logger.Warn("Failed to get statistics from engine",
				logging.Field("engine", name),
				logging.Field("error", err))
			continue
		}

		// 合并统计信息 Merge statistics
		totalStats.TotalSize += stats.TotalSize
		totalStats.TotalRows += stats.TotalRows

		for dbName, dbStats := range stats.Databases {
			if existingDbStats, exists := totalStats.Databases[dbName]; exists {
				existingDbStats.TableCount += dbStats.TableCount
				existingDbStats.RowCount += dbStats.RowCount
				existingDbStats.Size += dbStats.Size
				existingDbStats.IndexCount += dbStats.IndexCount
			} else {
				totalStats.Databases[dbName] = dbStats
			}
		}
	}

	return totalStats, nil
}

// handleError 处理错误 Handle error
func (a *StorageAdapter) handleError(operation string, err error) error {
	a.logger.Error("Storage operation failed",
		logging.Field("operation", operation),
		logging.Field("error", err))

	// 记录错误指标 Record error metrics
	if a.metrics != nil {
		a.metrics.Counter("storage_errors_total").
			WithLabels(map[string]string{
				"operation": operation,
				"error":     err.Error(),
			}).Inc()
	}

	// 触发错误事件 Trigger error event
	a.triggerEvent(EventOperationError, map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	})

	return err
}

// RequestRouter 请求路由器 Request router
type RequestRouter struct {
	engines         map[string]storage.Engine // 引擎映射 Engine mapping
	engineMutex     sync.RWMutex              // 引擎读写锁 Engine read-write lock
	config          *AdapterConfig            // 配置 Configuration
	loadBalancer    *LoadBalancer             // 负载均衡器 Load balancer
	circuitBreaker  *CircuitBreaker           // 熔断器 Circuit breaker
	failoverManager *FailoverManager          // 故障转移管理器 Failover manager
	logger          logging.Logger            // 日志器 Logger
	running         int32                     // 运行状态 Running state
}

// NewRequestRouter 创建请求路由器 Create request router
func NewRequestRouter(config *AdapterConfig) (*RequestRouter, error) {
	router := &RequestRouter{
		engines: make(map[string]storage.Engine),
		config:  config,
		logger:  logging.GetLogger("storage.sal.router"),
	}

	// 初始化负载均衡器 Initialize load balancer
	router.loadBalancer = NewLoadBalancer(config.LoadBalanceMethod)

	// 初始化熔断器 Initialize circuit breaker
	if config.FailoverStrategy == FailoverCircuitBreaker {
		router.circuitBreaker = NewCircuitBreaker()
	}

	// 初始化故障转移管理器 Initialize failover manager
	router.failoverManager = NewFailoverManager(config.FailoverStrategy)

	return router, nil
}

// Start 启动路由器 Start router
func (r *RequestRouter) Start(ctx context.Context) error {
	atomic.StoreInt32(&r.running, 1)
	r.logger.Info("Request router started")
	return nil
}

// Stop 停止路由器 Stop router
func (r *RequestRouter) Stop(ctx context.Context) error {
	atomic.StoreInt32(&r.running, 0)
	r.logger.Info("Request router stopped")
	return nil
}

// AddEngine 添加引擎 Add engine
func (r *RequestRouter) AddEngine(name string, engine storage.Engine) {
	r.engineMutex.Lock()
	defer r.engineMutex.Unlock()

	r.engines[name] = engine
	r.loadBalancer.AddEngine(name, engine)
}

// RemoveEngine 移除引擎 Remove engine
func (r *RequestRouter) RemoveEngine(name string) {
	r.engineMutex.Lock()
	defer r.engineMutex.Unlock()

	delete(r.engines, name)
	r.loadBalancer.RemoveEngine(name)
}

// RouteCreateDatabase 路由创建数据库请求 Route create database request
func (r *RequestRouter) RouteCreateDatabase(ctx context.Context, name string, options *types.DatabaseOptions) (storage.Engine, error) {
	return r.selectEngine(ctx, "CreateDatabase", map[string]interface{}{
		"database": name,
		"options":  options,
	})
}

// RouteDropDatabase 路由删除数据库请求 Route drop database request
func (r *RequestRouter) RouteDropDatabase(ctx context.Context, name string) (storage.Engine, error) {
	return r.selectEngine(ctx, "DropDatabase", map[string]interface{}{
		"database": name,
	})
}

// RouteCreateTable 路由创建表请求 Route create table request
func (r *RequestRouter) RouteCreateTable(ctx context.Context, database, name string, schema *types.Schema) (storage.Engine, error) {
	return r.selectEngine(ctx, "CreateTable", map[string]interface{}{
		"database": database,
		"table":    name,
		"schema":   schema,
	})
}

// RouteDropTable 路由删除表请求 Route drop table request
func (r *RequestRouter) RouteDropTable(ctx context.Context, database, name string) (storage.Engine, error) {
	return r.selectEngine(ctx, "DropTable", map[string]interface{}{
		"database": database,
		"table":    name,
	})
}

// RouteInsert 路由插入请求 Route insert request
func (r *RequestRouter) RouteInsert(ctx context.Context, database, table string, rows []*types.Row) (storage.Engine, error) {
	return r.selectEngine(ctx, "Insert", map[string]interface{}{
		"database": database,
		"table":    table,
		"rows":     rows,
	})
}

// RouteUpdate 路由更新请求 Route update request
func (r *RequestRouter) RouteUpdate(ctx context.Context, database, table string, filter *types.Filter, updates map[string]interface{}) (storage.Engine, error) {
	return r.selectEngine(ctx, "Update", map[string]interface{}{
		"database": database,
		"table":    table,
		"filter":   filter,
		"updates":  updates,
	})
}

// RouteDelete 路由删除请求 Route delete request
func (r *RequestRouter) RouteDelete(ctx context.Context, database, table string, filter *types.Filter) (storage.Engine, error) {
	return r.selectEngine(ctx, "Delete", map[string]interface{}{
		"database": database,
		"table":    table,
		"filter":   filter,
	})
}

// RouteSelect 路由查询请求 Route select request
func (r *RequestRouter) RouteSelect(ctx context.Context, database, table string, query *types.Query) (storage.Engine, error) {
	return r.selectEngine(ctx, "Select", map[string]interface{}{
		"database": database,
		"table":    table,
		"query":    query,
	})
}

// RouteCreateIndex 路由创建索引请求 Route create index request
func (r *RequestRouter) RouteCreateIndex(ctx context.Context, database, table string, index *types.Index) (storage.Engine, error) {
	return r.selectEngine(ctx, "CreateIndex", map[string]interface{}{
		"database": database,
		"table":    table,
		"index":    index,
	})
}

// RouteDropIndex 路由删除索引请求 Route drop index request
func (r *RequestRouter) RouteDropIndex(ctx context.Context, database, table, name string) (storage.Engine, error) {
	return r.selectEngine(ctx, "DropIndex", map[string]interface{}{
		"database": database,
		"table":    table,
		"index":    name,
	})
}

// RouteBeginTransaction 路由开始事务请求 Route begin transaction request
func (r *RequestRouter) RouteBeginTransaction(ctx context.Context, options *types.TransactionOptions) (storage.Engine, error) {
	return r.selectEngine(ctx, "BeginTransaction", map[string]interface{}{
		"options": options,
	})
}

// selectEngine 选择存储引擎 Select storage engine
func (r *RequestRouter) selectEngine(ctx context.Context, operation string, params map[string]interface{}) (storage.Engine, error) {
	r.engineMutex.RLock()
	defer r.engineMutex.RUnlock()

	if len(r.engines) == 0 {
		return nil, errors.NewError(errors.ErrCodeNotFound, "no storage engines available")
	}

	// 根据路由策略选择引擎 Select engine based on routing policy
	switch r.config.RoutingPolicy {
	case RoutingPolicyRoundRobin:
		return r.selectByRoundRobin()
	case RoutingPolicyRandom:
		return r.selectByRandom()
	case RoutingPolicyWeighted:
		return r.selectByWeight()
	case RoutingPolicyHash:
		return r.selectByHash(params)
	case RoutingPolicyLatency:
		return r.selectByLatency()
	case RoutingPolicyCapacity:
		return r.selectByCapacity()
	default:
		return r.selectByRoundRobin()
	}
}

// selectByRoundRobin 轮询选择 Round robin selection
func (r *RequestRouter) selectByRoundRobin() (storage.Engine, error) {
	return r.loadBalancer.SelectEngine()
}

// selectByRandom 随机选择 Random selection
func (r *RequestRouter) selectByRandom() (storage.Engine, error) {
	engines := make([]storage.Engine, 0, len(r.engines))
	for _, engine := range r.engines {
		engines = append(engines, engine)
	}

	if len(engines) == 0 {
		return nil, errors.NewError(errors.ErrCodeNotFound, "no engines available")
	}

	index := rand.Intn(len(engines))
	return engines[index], nil
}

// selectByWeight 按权重选择 Weighted selection
func (r *RequestRouter) selectByWeight() (storage.Engine, error) {
	// 这里可以实现基于权重的选择逻辑
	// Here can implement weight-based selection logic
	return r.loadBalancer.SelectEngine()
}

// selectByHash 哈希选择 Hash selection
func (r *RequestRouter) selectByHash(params map[string]interface{}) (storage.Engine, error) {
	// 生成哈希键 Generate hash key
	hashKey := r.generateHashKey(params)

	engines := make([]string, 0, len(r.engines))
	for name := range r.engines {
		engines = append(engines, name)
	}

	if len(engines) == 0 {
		return nil, errors.NewError(errors.ErrCodeNotFound, "no engines available")
	}

	sort.Strings(engines) // 保证一致性 Ensure consistency

	hasher := fnv.New32a()
	hasher.Write([]byte(hashKey))
	hash := hasher.Sum32()

	index := int(hash) % len(engines)
	engineName := engines[index]

	return r.engines[engineName], nil
}

// selectByLatency 按延迟选择 Latency-based selection
func (r *RequestRouter) selectByLatency() (storage.Engine, error) {
	// 这里可以实现基于延迟的选择逻辑
	// Here can implement latency-based selection logic
	return r.loadBalancer.SelectEngine()
}

// selectByCapacity 按容量选择 Capacity-based selection
func (r *RequestRouter) selectByCapacity() (storage.Engine, error) {
	// 这里可以实现基于容量的选择逻辑
	// Here can implement capacity-based selection logic
	return r.loadBalancer.SelectEngine()
}

// generateHashKey 生成哈希键 Generate hash key
func (r *RequestRouter) generateHashKey(params map[string]interface{}) string {
	var keyParts []string

	if database, ok := params["database"].(string); ok {
		keyParts = append(keyParts, database)
	}

	if table, ok := params["table"].(string); ok {
		keyParts = append(keyParts, table)
	}

	return strings.Join(keyParts, ":")
}

// LoadBalancer 负载均衡器 Load balancer
type LoadBalancer struct {
	method  LoadBalanceMethod         // 负载均衡方法 Load balance method
	engines map[string]storage.Engine // 引擎映射 Engine mapping
	counter int64                     // 计数器 Counter
	mutex   sync.RWMutex              // 读写锁 Read-write lock
}

// NewLoadBalancer 创建负载均衡器 Create load balancer
func NewLoadBalancer(method LoadBalanceMethod) *LoadBalancer {
	return &LoadBalancer{
		method:  method,
		engines: make(map[string]storage.Engine),
	}
}

// AddEngine 添加引擎 Add engine
func (lb *LoadBalancer) AddEngine(name string, engine storage.Engine) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.engines[name] = engine
}

// RemoveEngine 移除引擎 Remove engine
func (lb *LoadBalancer) RemoveEngine(name string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	delete(lb.engines, name)
}

// SelectEngine 选择引擎 Select engine
func (lb *LoadBalancer) SelectEngine() (storage.Engine, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if len(lb.engines) == 0 {
		return nil, errors.NewError(errors.ErrCodeNotFound, "no engines available")
	}

	switch lb.method {
	case LoadBalanceRoundRobin:
		return lb.selectRoundRobin()
	case LoadBalanceLeastConn:
		return lb.selectLeastConn()
	case LoadBalanceWeighted:
		return lb.selectWeighted()
	case LoadBalanceResource:
		return lb.selectResource()
	default:
		return lb.selectRoundRobin()
	}
}

// selectRoundRobin 轮询选择 Round robin selection
func (lb *LoadBalancer) selectRoundRobin() (storage.Engine, error) {
	engines := make([]storage.Engine, 0, len(lb.engines))
	for _, engine := range lb.engines {
		engines = append(engines, engine)
	}

	counter := atomic.AddInt64(&lb.counter, 1)
	index := int(counter-1) % len(engines)

	return engines[index], nil
}

// selectLeastConn 最少连接选择 Least connections selection
func (lb *LoadBalancer) selectLeastConn() (storage.Engine, error) {
	// 这里可以实现最少连接选择逻辑
	// Here can implement least connections selection logic
	return lb.selectRoundRobin()
}

// selectWeighted 加权选择 Weighted selection
func (lb *LoadBalancer) selectWeighted() (storage.Engine, error) {
	// 这里可以实现加权选择逻辑
	// Here can implement weighted selection logic
	return lb.selectRoundRobin()
}

// selectResource 资源感知选择 Resource-aware selection
func (lb *LoadBalancer) selectResource() (storage.Engine, error) {
	// 这里可以实现资源感知选择逻辑
	// Here can implement resource-aware selection logic
	return lb.selectRoundRobin()
}

// CircuitBreaker 熔断器 Circuit breaker
type CircuitBreaker struct {
	state        CircuitBreakerState // 状态 State
	failureCount int                 // 失败计数 Failure count
	successCount int                 // 成功计数 Success count
	lastFailTime time.Time           // 最后失败时间 Last failure time
	timeout      time.Duration       // 超时时间 Timeout
	threshold    int                 // 阈值 Threshold
	mutex        sync.RWMutex        // 读写锁 Read-write lock
}

// CircuitBreakerState 熔断器状态 Circuit breaker state
type CircuitBreakerState int

const (
	CircuitBreakerClosed   CircuitBreakerState = iota // 关闭 Closed
	CircuitBreakerOpen                                // 打开 Open
	CircuitBreakerHalfOpen                            // 半开 Half-open
)

// NewCircuitBreaker 创建熔断器 Create circuit breaker
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:     CircuitBreakerClosed,
		timeout:   30 * time.Second,
		threshold: 5,
	}
}

// CanExecute 是否可以执行 Can execute
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return time.Since(cb.lastFailTime) > cb.timeout
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功 Record success
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.successCount++

	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
	}
}

// RecordFailure 记录失败 Record failure
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.threshold {
		cb.state = CircuitBreakerOpen
	}
}

// FailoverManager 故障转移管理器 Failover manager
type FailoverManager struct {
	strategy FailoverStrategy // 策略 Strategy
}

// NewFailoverManager 创建故障转移管理器 Create failover manager
func NewFailoverManager(strategy FailoverStrategy) *FailoverManager {
	return &FailoverManager{
		strategy: strategy,
	}
}

// HandleFailure 处理故障 Handle failure
func (fm *FailoverManager) HandleFailure(ctx context.Context, engine storage.Engine, err error) error {
	switch fm.strategy {
	case FailoverImmediate:
		return fm.immediateFailover(ctx, engine, err)
	case FailoverGraceful:
		return fm.gracefulFailover(ctx, engine, err)
	case FailoverCircuitBreaker:
		return fm.circuitBreakerFailover(ctx, engine, err)
	default:
		return err
	}
}

// immediateFailover 立即故障转移 Immediate failover
func (fm *FailoverManager) immediateFailover(ctx context.Context, engine storage.Engine, err error) error {
	// 实现立即故障转移逻辑
	// Implement immediate failover logic
	return err
}

// gracefulFailover 优雅故障转移 Graceful failover
func (fm *FailoverManager) gracefulFailover(ctx context.Context, engine storage.Engine, err error) error {
	// 实现优雅故障转移逻辑
	// Implement graceful failover logic
	return err
}

// circuitBreakerFailover 熔断器故障转移 Circuit breaker failover
func (fm *FailoverManager) circuitBreakerFailover(ctx context.Context, engine storage.Engine, err error) error {
	// 实现熔断器故障转移逻辑
	// Implement circuit breaker failover logic
	return err
}

// PerformanceMonitor 性能监控器 Performance monitor
type PerformanceMonitor struct {
	interval  time.Duration      // 监控间隔 Monitoring interval
	retention time.Duration      // 保留时间 Retention time
	metrics   map[string]*Metric // 指标映射 Metrics mapping
	mutex     sync.RWMutex       // 读写锁 Read-write lock
	running   int32              // 运行状态 Running state
	logger    logging.Logger     // 日志器 Logger
}

// Metric 指标结构 Metric structure
type Metric struct {
	Name      string            // 名称 Name
	Value     float64           // 值 Value
	Timestamp time.Time         // 时间戳 Timestamp
	Labels    map[string]string // 标签 Labels
}

// NewPerformanceMonitor 创建性能监控器 Create performance monitor
func NewPerformanceMonitor(interval, retention time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		interval:  interval,
		retention: retention,
		metrics:   make(map[string]*Metric),
		logger:    logging.GetLogger("storage.sal.monitor"),
	}
}

// Start 启动监控器 Start monitor
func (pm *PerformanceMonitor) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&pm.running, 0, 1) {
		return
	}

	go pm.monitorLoop(ctx)
}

// Stop 停止监控器 Stop monitor
func (pm *PerformanceMonitor) Stop(ctx context.Context) {
	atomic.StoreInt32(&pm.running, 0)
}

// RecordRequest 记录请求 Record request
func (pm *PerformanceMonitor) RecordRequest(operation string, duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	metric := &Metric{
		Name:      "request_duration",
		Value:     float64(duration.Milliseconds()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"operation": operation,
		},
	}

	key := fmt.Sprintf("%s_%s", metric.Name, operation)
	pm.metrics[key] = metric
}

// monitorLoop 监控循环 Monitoring loop
func (pm *PerformanceMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&pm.running) == 0 {
				return
			}
			pm.cleanupOldMetrics()
		}
	}
}

// cleanupOldMetrics 清理旧指标 Cleanup old metrics
func (pm *PerformanceMonitor) cleanupOldMetrics() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	cutoff := time.Now().Add(-pm.retention)

	for key, metric := range pm.metrics {
		if metric.Timestamp.Before(cutoff) {
			delete(pm.metrics, key)
		}
	}
}

// HealthChecker 健康检查器 Health checker
type HealthChecker struct {
	engines  map[string]storage.Engine // 引擎映射 Engine mapping
	interval time.Duration             // 检查间隔 Check interval
	timeout  time.Duration             // 超时时间 Timeout
	mutex    sync.RWMutex              // 读写锁 Read-write lock
	running  int32                     // 运行状态 Running state
	logger   logging.Logger            // 日志器 Logger
}

// NewHealthChecker 创建健康检查器 Create health checker
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		engines:  make(map[string]storage.Engine),
		interval: interval,
		timeout:  timeout,
		logger:   logging.GetLogger("storage.sal.health"),
	}
}

// Start 启动健康检查器 Start health checker
func (hc *HealthChecker) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&hc.running, 0, 1) {
		return
	}

	go hc.healthCheckLoop(ctx)
}

// Stop 停止健康检查器 Stop health checker
func (hc *HealthChecker) Stop(ctx context.Context) {
	atomic.StoreInt32(&hc.running, 0)
}

// AddEngine 添加引擎 Add engine
func (hc *HealthChecker) AddEngine(name string, engine storage.Engine) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.engines[name] = engine
}

// RemoveEngine 移除引擎 Remove engine
func (hc *HealthChecker) RemoveEngine(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.engines, name)
}

// healthCheckLoop 健康检查循环 Health check loop
func (hc *HealthChecker) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&hc.running) == 0 {
				return
			}
			hc.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck 执行健康检查 Perform health check
func (hc *HealthChecker) performHealthCheck(ctx context.Context) {
	hc.mutex.RLock()
	engines := make(map[string]storage.Engine)
	for name, engine := range hc.engines {
		engines[name] = engine
	}
	hc.mutex.RUnlock()

	for name, engine := range engines {
		go func(engineName string, eng storage.Engine) {
			ctx, cancel := context.WithTimeout(ctx, hc.timeout)
			defer cancel()

			if err := hc.checkEngineHealth(ctx, eng); err != nil {
				hc.logger.Warn("Engine health check failed",
					logging.Field("engine", engineName),
					logging.Field("error", err))
			} else {
				hc.logger.Debug("Engine health check passed",
					logging.Field("engine", engineName))
			}
		}(name, engine)
	}
}

// checkEngineHealth 检查引擎健康状态 Check engine health
func (hc *HealthChecker) checkEngineHealth(ctx context.Context, engine storage.Engine) error {
	// 这里可以实现具体的健康检查逻辑
	// Here can implement specific health check logic
	_, err := engine.GetStatistics(ctx)
	return err
}

// Event 事件类型 Event types
type EventType string

const (
	EventAdapterStarted     EventType = "adapter_started"     // 适配器启动 Adapter started
	EventAdapterStopped     EventType = "adapter_stopped"     // 适配器停止 Adapter stopped
	EventEngineRegistered   EventType = "engine_registered"   // 引擎注册 Engine registered
	EventEngineUnregistered EventType = "engine_unregistered" // 引擎注销 Engine unregistered
	EventOperationError     EventType = "operation_error"     // 操作错误 Operation error
	EventHealthCheckFailed  EventType = "health_check_failed" // 健康检查失败 Health check failed
)

// EventListener 事件监听器接口 Event listener interface
type EventListener interface {
	OnEvent(eventType EventType, data map[string]interface{}) // 事件处理 Event handling
}

// AddEventListener 添加事件监听器 Add event listener
func (a *StorageAdapter) AddEventListener(listener EventListener) {
	a.listenerMutex.Lock()
	defer a.listenerMutex.Unlock()

	a.listeners = append(a.listeners, listener)
}

// RemoveEventListener 移除事件监听器 Remove event listener
func (a *StorageAdapter) RemoveEventListener(listener EventListener) {
	a.listenerMutex.Lock()
	defer a.listenerMutex.Unlock()

	for i, l := range a.listeners {
		if l == listener {
			a.listeners = append(a.listeners[:i], a.listeners[i+1:]...)
			break
		}
	}
}

// triggerEvent 触发事件 Trigger event
func (a *StorageAdapter) triggerEvent(eventType EventType, data map[string]interface{}) {
	a.listenerMutex.RLock()
	listeners := make([]EventListener, len(a.listeners))
	copy(listeners, a.listeners)
	a.listenerMutex.RUnlock()

	for _, listener := range listeners {
		go func(l EventListener) {
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error("Event listener panicked",
						logging.Field("event", string(eventType)),
						logging.Field("panic", r))
				}
			}()
			l.OnEvent(eventType, data)
		}(listener)
	}
}

// DefaultAdapterConfig 默认适配器配置 Default adapter configuration
func DefaultAdapterConfig() *AdapterConfig {
	return &AdapterConfig{
		RoutingPolicy:         RoutingPolicyRoundRobin,
		LoadBalanceMethod:     LoadBalanceRoundRobin,
		FailoverStrategy:      FailoverGraceful,
		EnableMonitoring:      true,
		MonitoringInterval:    time.Minute,
		MetricsRetention:      time.Hour,
		HealthCheckEnabled:    true,
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    10 * time.Second,
		MaxConcurrentRequests: 1000,
		RequestTimeout:        30 * time.Second,
		RetryAttempts:         3,
		RetryDelay:            time.Second,
		EnableCache:           true,
		CacheSize:             1000,
		CacheTTL:              10 * time.Minute,
	}
}
