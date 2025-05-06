// cmd/guocedb-server/main.go

// The main package for the Guocedb database server executable.
// It handles configuration loading, initialization of the database engine
// and network server, and graceful shutdown.
//
// Guocedb 数据库服务器可执行文件的主包。
// 它处理配置加载、数据库引擎和网络服务器的初始化以及优雅关机。
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/constants" // For default config path
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/engine"
	"github.com/turtacn/guocedb/network/server"
	// Import other top-level components as needed for initialization
	// 例如：_ "github.com/turtacn/guocedb/storage/engines/badger" // Register Badger engine
)

var (
	// configFile is the path to the configuration file.
	// configFile 是配置文件的路径。
	configFile = flag.String("config", constants.DefaultConfigPath, "Path to the configuration file") // 配置文件路径
)

func main() {
	flag.Parse() // Parse command-line flags / 解析命令行标志

	// 1. Load Configuration
	// 1. 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration from %s: %v\n", *configFile, err) // 从 %s 加载配置失败。
		os.Exit(1) // Exit on configuration error / 配置错误时退出
	}

	// 2. Initialize Logging
	// 2. 初始化日志记录
	// Logging configuration should be available in the loaded config.
	// 日志记录配置应在加载的配置中可用。
	err = log.InitLogger(cfg)
	if err != nil {
		// If logging initialization fails, fall back to stderr or a basic logger.
		// 如果日志记录初始化失败，回退到 stderr 或基本日志记录器。
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err) // 初始化日志记录器失败。
		// Continue execution, but logging might be limited.
		// 继续执行，但日志记录可能受限。
	}
	log.Info("Guocedb server starting...") // Guocedb 服务器正在启动...


	// Create a root context for the server lifetime.
	// 为服务器生命周期创建一个根 context。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is called on exit / 确保在退出时调用 cancel


	// 3. Initialize Core Engine
	// 3. 初始化核心引擎
	// The engine initialization requires configuration and potentially registers components.
	// 引擎初始化需要配置，并可能注册组件。
	eng, err := engine.NewEngine(ctx, cfg)
	if err != nil {
		log.Error("Failed to initialize database engine: %v", err) // 初始化数据库引擎失败。
		cancel() // Signal shutdown / 发出关机信号
		os.Exit(1) // Exit on engine initialization error / 引擎初始化错误时退出
	}
	log.Info("Database engine initialized.") // 数据库引擎初始化完成。


	// 4. Initialize Network Server
	// 4. 初始化网络服务器
	// The network server needs configuration and the core engine to handle requests.
	// 网络服务器需要配置和核心引擎来处理请求。
	netServer, err := server.NewServer(ctx, cfg, eng) // Pass config and engine
	if err != nil {
		log.Error("Failed to initialize network server: %v", err) // 初始化网络服务器失败。
		// Clean up engine if necessary before exiting.
		// 如果需要，在退出前清理引擎。
		if eng != nil {
			// Add a context with timeout for engine shutdown during error handling.
			// 在错误处理期间为引擎关机添加带有超时的 context。
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second) // Example timeout
			defer shutdownCancel()
			if stopErr := eng.Stop(shutdownCtx); stopErr != nil {
				log.Error("Failed to stop engine during network server init error: %v", stopErr) // 初始化网络服务器出错时停止引擎失败。
			}
		}
		cancel() // Signal shutdown / 发出关机信号
		os.Exit(1) // Exit on network server initialization error / 网络服务器初始化错误时退出
	}
	log.Info("Network server initialized.") // 网络服务器初始化完成。


	// 5. Start Network Server
	// 5. 启动网络服务器
	// Start the network server listener in a goroutine so main can wait for signals.
	// 在 goroutine 中启动网络服务器监听器，以便 main 可以等待信号。
	go func() {
		if startErr := netServer.Start(); startErr != nil {
			// Log the error, but the main function's signal handling will manage shutdown.
			// 记录错误，但 main 函数的信号处理将管理关机。
			log.Error("Network server failed to start or run: %v", startErr) // 网络服务器启动或运行失败。
			// Potentially signal a fatal error back to main if needed.
			// 如果需要，可能向 main 发回致命错误信号。
			// For now, assume logging the error is sufficient.
			// 目前，假设记录错误就足够了。
		}
	}()
	log.Info("Network server started in background.") // 网络服务器已在后台启动。


	// 6. Set up Graceful Shutdown
	// 6. 设置优雅关机
	// Listen for OS signals to trigger a graceful shutdown.
	// 监听 OS 信号以触发优雅关机。
	signalChan := make(chan os.Signal, 1)
	// Register signals for interrupt (Ctrl+C) and termination.
	// 注册中断 (Ctrl+C) 和终止信号。
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Guocedb server is running. Press Ctrl+C to stop.") // Guocedb 服务器正在运行。按 Ctrl+C 停止。

	// Wait for a shutdown signal.
	// 等待关机信号。
	<-signalChan
	log.Info("Shutdown signal received. Initiating graceful shutdown.") // 收到关机信号。正在启动优雅关机。


	// 7. Graceful Shutdown Process
	// 7. 优雅关机过程
	// Stop the network server first to stop accepting new connections.
	// 先停止网络服务器，停止接受新连接。
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second) // Provide a timeout for shutdown
	defer shutdownCancel()

	log.Info("Stopping network server...") // 正在停止网络服务器...
	if stopErr := netServer.Stop(shutdownCtx); stopErr != nil { // Pass shutdown context to Stop
		log.Error("Failed to stop network server gracefully: %v", stopErr) // 优雅停止网络服务器失败。
		// Continue with engine shutdown even if network server stop failed.
		// 即使网络服务器停止失败，也继续进行引擎关机。
	} else {
		log.Info("Network server stopped.") // 网络服务器已停止。
	}

	// Then stop the core engine (flushing data, closing storage, etc.).
	// 然后停止核心引擎（刷新数据、关闭存储等）。
	log.Info("Stopping database engine...") // 正在停止数据库引擎...
	if stopErr := eng.Stop(shutdownCtx); stopErr != nil { // Pass shutdown context to Stop
		log.Error("Failed to stop database engine gracefully: %v", stopErr) // 优雅停止数据库引擎失败。
		os.Exit(1) // Exit with error code if engine shutdown fails
	} else {
		log.Info("Database engine stopped.") // 数据库引擎已停止。
	}


	log.Info("Guocedb server stopped gracefully.") // Guocedb 服务器已优雅停止。
	os.Exit(0) // Exit successfully / 成功退出
}