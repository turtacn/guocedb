// Package mysql provides the MySQL server protocol implementation.
// server.go sets up and starts the main MySQL server listener.
//
// mysql 包提供了 MySQL 服务器协议实现。
// server.go 设置并启动主 MySQL 服务器监听器。
package mysql

import (
	"context"
	"fmt"
	"net" // For network listener

	"github.com/dolthub/go-mysql-server/server/mysql" // Import GMS MySQL server types
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	compute_analyzer "github.com/turtacub/guocedb/compute/analyzer" // Import compute analyzer
	compute_catalog "github.com/turtacn/guocedb/compute/catalog"       // Import compute catalog
	compute_executor "github.com/turtacn/guocedb/compute/executor"     // Import compute executor
	compute_optimizer "github.com/turtacub/guocedb/compute/optimizer" // Import compute optimizer
	compute_parser "github.com/turtacn/guocedb/compute/parser"         // Import compute parser
	compute_transaction "github.com/turtacn/guocedb/compute/transaction" // Import compute transaction manager
	"github.com/turtacn/guocedb/interfaces" // Import storage engine interface
)

// Server is the main MySQL server instance.
// It encapsulates the go-mysql-server server and initializes its dependencies.
//
// Server 是主 MySQL 服务器实例。
// 它封装 go-mysql-server 服务器并初始化其依赖项。
type Server struct {
	// gmsServer is the underlying go-mysql-server server instance.
	// gmsServer 是底层的 go-mysql-server 服务器实例。
	gmsServer *mysql.Server

	// storageEngine is the underlying storage engine used by the server.
	// storageEngine 是服务器使用的底层存储引擎。
	storageEngine interfaces.StorageEngine

	// computeComponents holds references to the initialized compute layer components.
	// computeComponents 保存对已初始化的计算层组件的引用。
	computeComponents *ComputeComponents
}

// NewServer creates a new Server instance.
// It initializes the compute layer components and the MySQL handler and authenticator.
//
// NewServer 创建一个新的 Server 实例。
// 它初始化计算层组件以及 MySQL 处理程序和认证器。
func NewServer(ctx context.Context, cfg config.Config, storageEngine interfaces.StorageEngine) (*Server, error) {
	log.Info("Initializing MySQL server.") // 初始化 MySQL 服务器。

	// Get MySQL server configuration from config
	// 从配置获取 MySQL 服务器配置
	bindHost := cfg.GetString(enum.ConfigKeyMySQLBindHost)
	bindPort := cfg.GetInt(enum.ConfigKeyMySQLBindPort)
	address := fmt.Sprintf("%s:%d", bindHost, bindPort)

	log.Info("MySQL server address: %s", address) // MySQL 服务器地址。

	// Initialize compute layer components
	// 初始化计算层组件
	// Note: Parser, Analyzer, Optimizer, Executor, TxManager, Catalog
	// The compute catalog needs the storage engine.
	//
	// 注意：Parser、Analyzer、Optimizer、Executor、TxManager、Catalog
	// 计算 catalog 需要存储引擎。
	computeCatalog, err := compute_catalog.NewPersistentCatalog(ctx, storageEngine) // Use PersistentCatalog backed by the storage engine
	if err != nil {
		log.Error("Failed to initialize compute catalog: %v", err) // 初始化计算 catalog 失败。
		return nil, fmt.Errorf("failed to initialize compute catalog: %w", err)
	}
	log.Info("Compute catalog initialized.") // 计算 catalog 初始化完成。

	computeTxManager := compute_transaction.NewComputeTransactionManager(storageEngine) // TxManager needs storage engine
	log.Info("Compute transaction manager initialized.") // 计算事务管理器初始化完成。

	// Initialize GMS-based compute components
	// These wrap or use GMS functionalities.
	//
	// 初始化基于 GMS 的计算组件。
	// 这些包装或使用 GMS 功能。
	gmsParser := compute_parser.NewGMSParser() // GMS parser wrapper
	log.Info("GMS parser initialized.") // GMS 解析器初始化完成。

	// GMS analyzer requires an Analyzer instance (often tied to a catalog/provider).
	// Our compute_analyzer.GMSAnalyzer wraps the GMS analyzer created with defaults.
	//
	// GMS 分析器需要一个 Analyzer 实例（通常绑定到 catalog/provider）。
	// 我们的 compute_analyzer.GMSAnalyzer 包装了使用默认值创建的 GMS 分析器。
	// Need to get the underlying GMS analyzer instance.
	// 需要获取底层的 GMS 分析器实例。
	// The GMS analyzer is typically created by GMS itself or its factory.
	// Our compute_analyzer.NewGMSAnalyzer creates the GMS analyzer.
	//
	// GMS 分析器通常由 GMS 本身或其工厂创建。
	// 我们的 compute_analyzer.NewGMSAnalyzer 创建 GMS 分析器。
	gmsAnalyzer := compute_analyzer.NewGMSAnalyzer() // Our wrapper
	log.Info("GMS analyzer wrapper initialized.") // GMS 分析器包装器初始化完成。

	// The GMS optimizer is usually part of the GMS analyzer.
	// Our compute_optimizer.NewGMSOptimizer takes the GMS analyzer instance.
	//
	// GMS 优化器通常是 GMS 分析器的一部分。
	// 我们的 compute_optimizer.NewGMSOptimizer 接受 GMS 分析器实例。
	// Need to get the underlying GMS analyzer instance from our wrapper.
	// It seems our compute_analyzer.GMSAnalyzer holds `gmsAnalyzer *analyzer.Analyzer`.
	// Let's pass that to the optimizer factory.
	//
	// 需要从我们的包装器获取底层的 GMS 分析器实例。
	// 看起来我们的 compute_analyzer.GMSAnalyzer 持有 `gmsAnalyzer *analyzer.Analyzer`。
	// 将其传递给优化器工厂。
	gmsOptimizer := compute_optimizer.NewGMSOptimizer(gmsAnalyzer.(*compute_analyzer.GMSAnalyzer).gmsAnalyzer) // Cast to get the internal GMS analyzer
	log.Info("GMS optimizer wrapper initialized.") // GMS 优化器包装器初始化完成。


	gmsExecutor := compute_executor.NewGMSExecutor() // GMS executor wrapper
	log.Info("GMS executor wrapper initialized.") // GMS 执行器包装器初始化完成。


	// Collect compute components
	// 收集计算组件
	computeComponents := &ComputeComponents{
		Parser: gmsParser,       // Our parser wrapper
		Analyzer: gmsAnalyzer,   // Our analyzer wrapper
		Optimizer: gmsOptimizer, // Our optimizer wrapper
		Executor: gmsExecutor,   // Our executor wrapper
		Catalog: computeCatalog, // Our compute catalog (persistent)
		TxManager: computeTxManager, // Our compute tx manager
	}

	// Initialize MySQL protocol handler and authenticator
	// 初始化 MySQL 协议处理程序和认证器
	// The handler needs access to the compute components.
	// 处理程序需要访问计算组件。
	handler := NewHandler(computeComponents) // Pass compute components to handler
	log.Info("MySQL handler initialized.") // MySQL 处理程序初始化完成。

	authenticator := NewAuthenticator() // Basic authenticator
	log.Info("MySQL authenticator initialized.") // MySQL 认证器初始化完成。

	// Create the underlying go-mysql-server server instance
	// 创建底层的 go-mysql-server 服务器实例
	gmsServer, err := mysql.NewServer(address, authenticator, handler) // Pass address, authenticator, handler
	if err != nil {
		log.Error("Failed to create GMS MySQL server instance: %v", err) // 创建 GMS MySQL 服务器实例失败。
		return nil, fmt.Errorf("failed to create GMS MySQL server: %w", err)
	}
	log.Info("GMS MySQL server instance created.") // GMS MySQL 服务器实例创建成功。


	server := &Server{
		gmsServer:     gmsServer,
		storageEngine: storageEngine, // Store storage engine reference
		computeComponents: computeComponents, // Store compute components reference
	}

	log.Info("MySQL server initialized successfully.") // MySQL 服务器初始化成功。
	return server, nil
}

// Start starts the MySQL server listener.
// Start 启动 MySQL 服务器监听器。
func (s *Server) Start() error {
	log.Info("Starting MySQL server listener...") // 启动 MySQL 服务器监听器。
	// The underlying GMS server Start method listens and accepts connections.
	// 底层的 GMS 服务器 Start 方法监听并接受连接。
	// It uses the provided Handler and Authenticator for each connection.
	// 它为每个连接使用提供的 Handler 和 Authenticator。
	err := s.gmsServer.Start()
	if err != nil {
		log.Error("Failed to start GMS MySQL server listener: %v", err) // 启动 GMS MySQL 服务器监听器失败。
		return fmt.Errorf("failed to start MySQL server: %w", err)
	}
	log.Info("MySQL server listener started.") // MySQL 服务器监听器已启动。
	return nil // Start blocks until Stop is called or error
}

// Stop stops the MySQL server listener.
// Stop 停止 MySQL 服务器监听器。
func (s *Server) Stop() error {
	log.Info("Stopping MySQL server listener...") // 停止 MySQL 服务器监听器。
	// The underlying GMS server Stop method gracefully shuts down the listener and connections.
	// 底层的 GMS 服务器 Stop 方法优雅地关闭监听器和连接。
	if s.gmsServer != nil {
		err := s.gmsServer.Close() // GMS uses Close for shutdown
		if err != nil {
			log.Error("Failed to stop GMS MySQL server: %v", err) // 停止 GMS MySQL 服务器失败。
			return fmt.Errorf("failed to stop MySQL server: %w", err)
		}
		log.Info("MySQL server listener stopped.") // MySQL 服务器监听器已停止。
		return nil
	}
	log.Warn("MySQL server was not started or already stopped.") // MySQL 服务器未启动或已停止。
	return nil // Already stopped or never started
}