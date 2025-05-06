// Package server contains the network server implementation for Guocedb.
// It is responsible for listening for incoming client connections on various protocols.
//
// server 包包含 Guocedb 的网络服务器实现。
// 它负责监听各种协议上的传入客户端连接。
package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/engine" // Import the main engine package
	"github.com/turtacn/guocedb/protocol/mysql" // Import the MySQL protocol server
)

// Server is the main network server component.
// It orchestrates different protocol servers.
//
// Server 是主要的网络服务器组件。
// 它协调不同的协议服务器。
type Server struct {
	// engine is the core database engine providing compute and storage capabilities.
	// engine 是提供计算和存储能力的核心数据库引擎。
	engine *engine.Engine

	// mysqlServer is the instance of the MySQL protocol server.
	// mysqlServer 是 MySQL 协议服务器的实例。
	mysqlServer *mysql.Server

	// Add fields for other protocol servers here (e.g., postgreSQL, HTTP API).
	// 在此处添加其他协议服务器的字段（例如，postgreSQL, HTTP API）。

	// wg is used to wait for all protocol servers to stop.
	// wg 用于等待所有协议服务器停止。
	wg sync.WaitGroup
	// cancel context for stopping the server.
	// 用于停止服务器的取消 context。
	cancel context.CancelFunc
	ctx    context.Context
}

// NewServer creates a new network Server instance.
// It initializes protocol-specific servers based on the provided configuration.
//
// NewServer 创建一个新的网络 Server 实例。
// 它根据提供的配置初始化协议特定的服务器。
func NewServer(ctx context.Context, cfg config.Config, eng *engine.Engine) (*Server, error) {
	log.Info("Initializing network server.") // 初始化网络服务器。

	// The main server context should be cancellable
	// 主服务器 context 应该是可取消的
	serverCtx, cancel := context.WithCancel(ctx)


	// Initialize protocol-specific servers based on config
	// 根据配置初始化协议特定的服务器
	// Example: Initialize MySQL server if enabled in config
	// 示例：如果配置中启用，初始化 MySQL 服务器
	// Assuming config has a way to enable/disable protocols and their settings.
	// 假设配置有一种方式来启用/禁用协议及其设置。
	// For now, assume MySQL is always enabled and configured via cfg.
	// 目前，假设 MySQL 总是启用并通过 cfg 配置。

	// The MySQL server needs the configuration and the main engine.
	// MySQL 服务器需要配置和主引擎。
	mysqlServer, err := mysql.NewServer(serverCtx, cfg, eng.StorageEngine) // Pass config and storage engine (needed by mysql server for catalog/tx)
	if err != nil {
		cancel() // Cancel context on initialization error
		log.Error("Failed to initialize MySQL server: %v", err) // 初始化 MySQL 服务器失败。
		return nil, fmt.Errorf("failed to initialize MySQL server: %w", err)
	}
	log.Info("MySQL server initialized successfully within network server.") // 网络服务器内 MySQL 服务器初始化成功。


	// TODO: Initialize other protocol servers here based on config.
	// TODO: 在此处根据配置初始化其他协议服务器。

	server := &Server{
		engine:      eng, // Store reference to the main engine
		mysqlServer: mysqlServer,
		ctx:         serverCtx,
		cancel:      cancel,
	}

	log.Info("Network server initialized successfully.") // 网络服务器初始化成功。
	return server, nil
}

// Start starts all configured protocol server listeners.
// This method blocks until all servers are stopped or an error occurs.
//
// Start 启动所有配置的协议服务器监听器。
// 此方法会阻塞，直到所有服务器停止或发生错误。
func (s *Server) Start() error {
	log.Info("Starting network server listeners...") // 启动网络服务器监听器。

	// Start MySQL server in a goroutine
	// 在 goroutine 中启动 MySQL 服务器
	if s.mysqlServer != nil {
		s.wg.Add(1) // Add to wait group
		go func() {
			defer s.wg.Done()
			log.Info("Starting MySQL server listener.") // 启动 MySQL 服务器监听器。
			if err := s.mysqlServer.Start(); err != nil {
				log.Error("MySQL server listener failed: %v", err) // MySQL 服务器监听器失败。
				// Signal main server to stop? Or handle error internally?
				// Consider propagating fatal errors to the main context/channel.
				//
				// 通知主服务器停止？还是内部处理错误？
				// 考虑将致命错误传播到主 context/channel。
				// For now, just log the error.
				// 目前，只记录错误。
			}
			log.Info("MySQL server listener stopped.") // MySQL 服务器监听器已停止。
		}()
	}

	// TODO: Start other protocol servers in goroutines here.
	// TODO: 在此处在 goroutine 中启动其他协议服务器。


	log.Info("All network server listeners started. Waiting for shutdown signal.") // 所有网络服务器监听器已启动。等待关机信号。

	// Wait for the server context to be cancelled (signaling shutdown)
	// 等待服务器 context 被取消（发出关机信号）
	<-s.ctx.Done()

	log.Info("Shutdown signal received. Stopping network server listeners.") // 收到关机信号。停止网络服务器监听器。

	// Signal all protocol servers to stop (via context cancellation)
	// 通知所有协议服务器停止（通过 context 取消）
	// The Start method of protocol servers should ideally respect the context.
	// 协议服务器的 Start 方法理想情况下应尊重 context。
	// However, if their Start method blocks and doesn't watch the context,
	// we might need to call their explicit Stop methods.
	//
	// 然而，如果它们的 Start 方法阻塞且不监视 context，
	// 我们可能需要调用它们的显式 Stop 方法。

	// Call explicit Stop methods on protocol servers if they don't respect context
	// 如果协议服务器不尊重 context，调用其显式 Stop 方法
	if s.mysqlServer != nil {
		// We are already in the shutdown path triggered by context.
		// Calling s.cancel() above should signal servers respecting ctx.Done().
		// If mysqlServer.Start blocks and doesn't watch ctx.Done(), call mysqlServer.Stop() here.
		// Let's assume mysqlServer.Start respects ctx.Done() for now.
		// 如果 mysqlServer.Start 阻塞且不监视 ctx.Done()，在此处调用 mysqlServer.Stop()。
		// 目前，假设 mysqlServer.Start 尊重 ctx.Done()。
		log.Info("Calling stop on MySQL server (assumes it respects context or has a Stop method).") // 在 MySQL 服务器上调用停止（假设它尊重 context 或有 Stop 方法）。
		// If mysqlServer.Start doesn't return on ctx.Done(), uncomment the line below:
		// 如果 mysqlServer.Start 不在 ctx.Done() 上返回，取消注释以下行：
		// if err := s.mysqlServer.Stop(); err != nil {
		// 	log.Error("Failed to explicitly stop MySQL server: %v", err)
		// }
	}

	// TODO: Call Stop on other protocol servers here if needed.
	// TODO: 如果需要，在此处调用其他协议服务器上的 Stop。


	// Wait for all goroutines (protocol servers) to finish
	// 等待所有 goroutine（协议服务器）完成
	s.wg.Wait()

	log.Info("All network server listeners stopped.") // 所有网络服务器监听器已停止。
	return nil // Return nil if shutdown was successful
}

// Stop signals the network server to stop listening and shuts down all protocol servers.
// Stop 通知网络服务器停止监听并关闭所有协议服务器。
func (s *Server) Stop() {
	log.Info("Sending shutdown signal to network server.") // 向网络服务器发送关机信号。
	// Cancel the server context, which signals goroutines to stop.
	// 取消服务器 context，向 goroutine 发送停止信号。
	if s.cancel != nil {
		s.cancel()
	}
	// The Start method is waiting for this context to be done.
	// The Start method will then call Stop on individual servers if needed and wait for the wait group.
	// Start 方法正在等待此 context 完成。
	// 然后 Start 方法将根据需要调用各个服务器上的 Stop，并等待 wait group。
}