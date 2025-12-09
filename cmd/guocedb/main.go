// The main entry point for the GuoceDB database server.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/turtacn/guocedb/config"
	"github.com/turtacn/guocedb/server"
)

var (
	// Version information (set via ldflags during build)
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	rootCmd := buildRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runServer is the main server execution function.
func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadWithFlags(cfgFile, cmd.Flags())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logging
	logger, err := initLogging(cfg.Logging)
	if err != nil {
		return fmt.Errorf("init logging: %w", err)
	}

	logger.Info("Starting GuoceDB",
		"version", Version,
		"commit", GitCommit,
		"build_time", BuildTime,
		"port", cfg.Server.Port,
		"data_dir", cfg.Storage.DataDir,
	)

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	// Set the logger
	server.WithLogger(logger)(srv)

	// Register lifecycle hooks
	srv.Hooks().OnPostStart(func(s *server.Server) {
		logger.Info("GuoceDB started successfully",
			"address", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			"metrics", cfg.Observability.Address,
		)
	})

	srv.Hooks().OnPreStop(func(s *server.Server) {
		logger.Info("GuoceDB shutting down...", "uptime", s.Uptime())
	})

	srv.Hooks().OnPostStop(func(s *server.Server) {
		logger.Info("GuoceDB stopped")
	})

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleSignals(ctx, srv, logger)

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// handleSignals handles OS signals for graceful shutdown.
func handleSignals(ctx context.Context, srv *server.Server, logger *slog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigChan:
			logger.Info("Received shutdown signal", "signal", sig)

			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				srv.Config().Server.ShutdownTimeout,
			)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("Shutdown error", "error", err)
			}
			return
		}
	}
}

// initLogging initializes the logging system.
func initLogging(cfg config.LoggingConfig) (*slog.Logger, error) {
	// Parse log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Determine output
	var output io.Writer
	if cfg.Output == "stdout" || cfg.Output == "" {
		output = os.Stdout
	} else {
		// For file output, we would use lumberjack here
		// Since it's not in dependencies yet, fall back to stdout
		fmt.Fprintf(os.Stderr, "Warning: File logging not yet implemented, using stdout\n")
		output = os.Stdout
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler), nil
}

// printVersion prints version information.
func printVersion() {
	fmt.Printf("GuoceDB %s\n", Version)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
}

// checkConfig validates the configuration.
func checkConfig() error {
	cfg, err := config.LoadWithFlags(cfgFile, nil)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	fmt.Println("Configuration is valid")
	fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  Data Dir: %s\n", cfg.Storage.DataDir)
	fmt.Printf("  Log Level: %s\n", cfg.Logging.Level)
	fmt.Printf("  Security: %v\n", cfg.Security.Enabled)
	fmt.Printf("  Observability: %v (%s)\n", cfg.Observability.Enabled, cfg.Observability.Address)

	return nil
}
