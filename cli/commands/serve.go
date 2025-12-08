package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/turtacn/guocedb/cli/config"
	commonConfig "github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/auth"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/compute/optimizer"
	mysql "github.com/turtacn/guocedb/compute/server"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/maintenance/metrics"
	"github.com/turtacn/guocedb/network/server"
	"github.com/turtacn/guocedb/storage/sal"
)

// NewServeCmd creates the serve command.
func NewServeCmd(cfgFile *string) *cobra.Command {
	var (
		host    string
		port    int
		dataDir string
	)
	
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the GuoceDB server",
		Long:  `Start the GuoceDB server with MySQL protocol compatibility.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), *cfgFile, host, port, dataDir)
		},
	}
	
	cmd.Flags().StringVar(&host, "host", "", "listen host (overrides config)")
	cmd.Flags().IntVar(&port, "port", 0, "listen port (overrides config)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "data directory (overrides config)")
	
	return cmd
}

func runServe(ctx context.Context, cfgFile, host string, port int, dataDir string) error {
	// 1. Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Command line arguments override config file
	if host != "" {
		cfg.Server.Host = host
	}
	if port != 0 {
		cfg.Server.Port = port
	}
	if dataDir != "" {
		cfg.Storage.DataDir = dataDir
	}
	
	// 2. Initialize logger
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting GuoceDB server...")
	logger.Infof("Version: %s", Version)
	logger.Infof("Listen address: %s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Data directory: %s", cfg.Storage.DataDir)
	
	// 3. Ensure data directory exists
	if err := os.MkdirAll(cfg.Storage.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// 4. Initialize storage layer
	commonCfg := convertToCommonConfig(cfg)
	storage, err := sal.NewAdapter(commonCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage engine: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Errorf("Error closing storage: %v", err)
		}
	}()
	
	// 5. Initialize compute layer
	catalog := sql.NewCatalog()
	analyzer := analyzer.NewAnalyzer(catalog)
	optimizer := optimizer.NewOptimizer()
	engine := executor.NewEngine(analyzer, optimizer, catalog)
	
	// 6. Initialize authentication
	var authMethod auth.Auth
	if cfg.Auth.Enabled {
		// TODO: Implement proper authentication with users from config
		authMethod = auth.NewNone()
	} else {
		authMethod = auth.NewNone()
	}
	
	// 7. Initialize MySQL server
	serverConfig := mysql.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Auth:     authMethod,
	}
	
	mysqlServer, err := mysql.NewDefaultServer(serverConfig, engine)
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL server: %w", err)
	}
	
	// 8. Initialize metrics (if enabled)
	var metricsRegistry *metrics.Registry
	if cfg.Monitoring.Enabled {
		metricsRegistry = metrics.NewRegistry()
		go func() {
			metricsRegistry.Serve(&commonConfig.MetricsConfig{
				Enable: cfg.Monitoring.Enabled,
				Port:   cfg.Monitoring.Port,
			})
		}()
	}
	
	// 9. Initialize server manager
	serverMgr := server.NewManager()
	serverMgr.Register(mysqlServer)
	
	// 10. Write PID file
	if err := writePIDFile(cfg.Storage.DataDir); err != nil {
		logger.Warnf("Failed to write PID file: %v", err)
	}
	
	// 11. Start servers
	serverMgr.StartAll()
	logger.Info("GuoceDB server is ready to accept connections.")
	
	// 12. Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-sigCh:
		logger.Info("Received shutdown signal, shutting down...")
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down...")
	}
	
	// 13. Graceful shutdown
	logger.Info("Shutting down GuoceDB server...")
	serverMgr.CloseAll()
	
	// Clean up PID file
	removePIDFile(cfg.Storage.DataDir)
	
	logger.Info("Server shut down gracefully.")
	return nil
}

// writePIDFile writes the current process ID to a file.
func writePIDFile(dataDir string) error {
	pidPath := filepath.Join(dataDir, "guocedb.pid")
	pid := strconv.Itoa(os.Getpid())
	return os.WriteFile(pidPath, []byte(pid), 0644)
}

// removePIDFile removes the PID file.
func removePIDFile(dataDir string) {
	pidPath := filepath.Join(dataDir, "guocedb.pid")
	os.Remove(pidPath) // Ignore errors
}

// convertToCommonConfig converts CLI config to common config format.
// This is a temporary adapter until we unify the config structures.
func convertToCommonConfig(cfg *config.Config) *commonConfig.Config {
	return &commonConfig.Config{
		Storage: commonConfig.StorageConfig{
			Engine:  "badger",
			DataDir: cfg.Storage.DataDir,
			Badger: commonConfig.BadgerConfig{
				ValueLogFileSize: cfg.Storage.Badger.ValueLogFileSize,
				SyncWrites:       cfg.Storage.SyncWrites,
			},
		},
	}
}