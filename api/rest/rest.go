// Package rest contains the REST API server implementation.
// It provides an HTTP interface for interacting with the database,
// potentially for management tasks or data access.
//
// rest 包包含 REST API 服务器实现。
// 它提供一个 HTTP 接口用于与数据库交互，
// 可能用于管理任务或数据访问。
package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/engine" // Need access to the core engine
)

// Server is the REST API server instance.
// Server 是 REST API 服务器实例。
type Server struct {
	// httpServer is the underlying standard library HTTP server.
	// httpServer 是底层的标准库 HTTP 服务器。
	httpServer *http.Server

	// engine is the core database engine to delegate requests to.
	// engine 是要将请求委托到的核心数据库引擎。
	engine *engine.Engine

	// config is the server configuration.
	// config 是服务器配置。
	config config.Config

	// mux is the HTTP request multiplexer.
	// mux 是 HTTP 请求多路复用器。
	mux *http.ServeMux // Use a standard mux for simplicity
}

// NewServer creates a new REST API Server instance.
// It initializes the HTTP server and sets up basic routes.
//
// NewServer 创建一个新的 REST API Server 实例。
// 它初始化 HTTP 服务器并设置基本路由。
func NewServer(cfg config.Config, eng *engine.Engine) (*Server, error) {
	log.Info("Initializing REST API server.") // 初始化 REST API 服务器。

	// Get REST API server configuration from config
	// 从配置获取 REST API 服务器配置
	bindHost := cfg.GetString(enum.ConfigKeyRESTBindHost)
	bindPort := cfg.GetInt(enum.ConfigKeyRESTBindPort)
	address := fmt.Sprintf("%s:%d", bindHost, bindPort)

	log.Info("REST API server address: %s", address) // REST API 服务器地址。


	// Create the HTTP multiplexer and server
	// 创建 HTTP 多路复用器和服务器
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    address,
		Handler: mux, // Use our mux
		// TODO: Add timeouts and other server configurations from config.
		// TODO: 从配置添加超时和其他服务器配置。
	}

	server := &Server{
		httpServer: httpServer,
		engine:     eng, // Store engine reference
		config:     cfg,
		mux:        mux, // Store mux reference
	}

	// Register basic routes (placeholders)
	// 注册基本路由（占位符）
	server.registerRoutes()

	log.Info("REST API server initialized successfully.") // REST API 服务器初始化成功。
	return server, nil
}

// registerRoutes sets up the HTTP routes for the REST API.
// registerRoutes 设置 REST API 的 HTTP 路由。
func (s *Server) registerRoutes() {
	log.Debug("Registering REST API routes.") // 注册 REST API 路由。

	// Example route: A simple health check endpoint
	// 示例路由：简单的健康检查端点
	s.mux.HandleFunc("/health", s.handleHealthCheck)

	// TODO: Register routes for management operations (databases, tables, users).
	// These handlers will call methods on the s.engine.
	// TODO: 注册管理操作（数据库、表、用户）的路由。
	// 这些处理程序将调用 s.engine 上的方法。
	// Example: s.mux.HandleFunc("/databases", s.handleListDatabases)
	// Example: s.mux.HandleFunc("/databases/", s.handleDatabaseOperations) // For GET, POST, DELETE on specific DBs
	// Example: s.mux.HandleFunc("/databases/{dbName}/tables", s.handleListTables)

	// TODO: Register routes for data access (e.g., querying data via POST with SQL).
	// This requires parsing/executing SQL received over HTTP.
	// TODO: 注册数据访问路由（例如，通过 POST 和 SQL 查询数据）。
	// 这需要解析/执行通过 HTTP 接收的 SQL。

	log.Debug("REST API routes registered.") // REST API 路由已注册。
}

// handleHealthCheck is a simple handler for the /health endpoint.
// It reports the server's basic health status.
//
// handleHealthCheck 是 /health 端点的一个简单处理程序。
// 它报告服务器的基本健康状态。
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request for /health.") // 收到 /health 请求。
	// In a real implementation, check the status of core components (engine, storage).
	// 在真实实现中，检查核心组件（engine, storage）的状态。
	// Could use the maintenance/status subsystem.
	// 可以使用 maintenance/status 子系统。

	// Simple placeholder health check: Is the engine reference non-nil?
	// 简单的占位符健康检查：engine 引用是否非 nil？
	if s.engine == nil {
		http.Error(w, "Internal server error: Engine not initialized", http.StatusInternalServerError) // 内部服务器错误：引擎未初始化。
		log.Error("Health check failed: Engine is nil.") // 健康检查失败：引擎为 nil。
		return
	}

	// TODO: Call s.engine.StatusReporter.GetStatus(r.Context()) for a more thorough check.
	// TODO: 调用 s.engine.StatusReporter.GetStatus(r.Context()) 进行更彻底的检查。

	w.WriteHeader(http.StatusOK) // Return 200 OK
	_, err := w.Write([]byte("OK")) // Write response body
	if err != nil {
		log.Error("Failed to write health check response: %v", err) // 写入健康检查响应失败。
	}
	log.Debug("Responded to /health with OK.") // 回复 /health OK。
}


// Start starts the REST API HTTP listener.
// Start 启动 REST API HTTP 监听器。
func (s *Server) Start() error {
	log.Info("Starting REST API server listener on %s", s.httpServer.Addr) // 启动 REST API 服务器监听器。
	// ListenAndServe blocks, so run in a goroutine in a real application.
	// For simplicity in this placeholder, it blocks.
	//
	// ListenAndServe 会阻塞，因此在实际应用中应在 goroutine 中运行。
	// 为了简化此占位符，它会阻塞。
	// In network/server, we will run this in a goroutine.
	// 在 network/server 中，我们将在 goroutine 中运行此方法。
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error("REST API server listener failed: %v", err) // REST API 服务器监听器失败。
		return fmt.Errorf("failed to start REST API server: %w", err)
	}
	log.Info("REST API server listener stopped.") // REST API 服务器监听器已停止。
	return nil // Return nil if shutdown was graceful
}

// Stop stops the REST API HTTP listener gracefully.
// Stop 优雅地停止 REST API HTTP 监听器。
func (s *Server) Stop(ctx context.Context) error {
	log.Info("Stopping REST API server listener.") // 停止 REST API 服务器监听器。
	// Use Shutdown for graceful shutdown
	// 使用 Shutdown 进行优雅关机
	if s.httpServer != nil {
		// Provide a context for shutdown timeout if needed: context.WithTimeout(context.Background(), 5*time.Second)
		// 如果需要关机超时，提供 context：context.WithTimeout(context.Background(), 5*time.Second)
		err := s.httpServer.Shutdown(ctx) // Shutdown takes a context
		if err != nil {
			log.Error("Failed to stop REST API server gracefully: %v", err) // 优雅停止 REST API 服务器失败。
			return fmt.Errorf("failed to stop REST API server gracefully: %w", err)
		}
		log.Info("REST API server listener stopped gracefully.") // REST API 服务器监听器已优雅停止。
	}
	log.Warn("REST API server was not started or already stopped.") // REST API 服务器未启动或已停止。
	return nil
}