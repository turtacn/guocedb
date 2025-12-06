// The main entry point for the guocedb server.
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/compute/optimizer"
	"github.com/turtacn/guocedb/maintenance/metrics"
	"github.com/turtacn/guocedb/network/server"
	"github.com/turtacn/guocedb/compute/auth"
	mysql "github.com/turtacn/guocedb/compute/server"
	"github.com/turtacn/guocedb/storage/sal"
	"fmt"
)

func main() {
	// 1. Parse command-line flags
	var configPath string
	flag.StringVar(&configPath, "config", "configs/config.yaml.example", "Path to the configuration file.")
	flag.Parse()

	// 2. Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.GetLogger().Fatalf("Failed to load configuration: %v", err)
	}

	// 3. Initialize logger
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting guocedb server...")
	logger.Infof("Version: %s", config.Get().Server.Version)

	// 4. Initialize components
	// Storage Layer
	storage, err := sal.NewAdapter(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize storage engine: %v", err)
	}

	// Compute Layer
	catalog := sql.NewCatalog()
	analyzer := analyzer.NewAnalyzer(catalog)
	optimizer := optimizer.NewOptimizer()
	engine := executor.NewEngine(analyzer, optimizer, catalog)

	// Protocol Layer
	// Use auth.None for now or load from config/file
	authMethod := auth.NewNone()

	serverConfig := mysql.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Auth:     authMethod,
	}

	mysqlServer, err := mysql.NewDefaultServer(serverConfig, engine)
	if err != nil {
		logger.Fatalf("Failed to initialize MySQL server: %v", err)
	}

	// Maintenance Layer
	metricsRegistry := metrics.NewRegistry()

	// Network Server Manager
	serverMgr := server.NewManager()
	serverMgr.Register(mysqlServer)
	// serverMgr.Register(grpcServer)
	// serverMgr.Register(httpGateway)

	// 5. Start servers
	serverMgr.StartAll()
	metricsRegistry.Serve(&cfg.Metrics)

	logger.Info("guocedb server is ready to accept connections.")

	// 6. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down guocedb server...")
	serverMgr.CloseAll()
	if err := storage.Close(); err != nil {
		logger.Errorf("Error closing storage: %v", err)
	}
	logger.Info("Server shut down gracefully.")
}
