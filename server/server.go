// Package server provides unified server lifecycle management for GuoceDB.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	commonConfig "github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/compute/optimizer"
	mysql "github.com/turtacn/guocedb/compute/server"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/config"
	"github.com/turtacn/guocedb/observability"
	"github.com/turtacn/guocedb/observability/health"
	"github.com/turtacn/guocedb/storage/sal"
)

// Server state constants.
const (
	stateNew = iota
	stateStarting
	stateRunning
	stateStopping
	stateStopped
)

// Server manages the lifecycle of all GuoceDB components.
type Server struct {
	cfg *config.Config

	// Core components
	storage     *sal.Adapter
	catalog     *sql.Catalog
	analyzer    *analyzer.Analyzer
	optimizer   optimizer.Optimizer
	engine      *executor.Engine
	mysqlServer *mysql.Server
	obsServer   *observability.Server

	// State management
	state     atomic.Int32
	startTime time.Time
	logger    *slog.Logger

	// Lifecycle
	hooks    *LifecycleHooks
	stopChan chan struct{}
	doneChan chan struct{}
}

// New creates a new server instance with the given configuration.
func New(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	srv := &Server{
		cfg:      cfg,
		hooks:    NewLifecycleHooks(),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	srv.state.Store(stateNew)

	return srv, nil
}

// Config returns the server configuration.
func (s *Server) Config() *config.Config {
	return s.cfg
}

// Hooks returns the lifecycle hooks manager.
func (s *Server) Hooks() *LifecycleHooks {
	return s.hooks
}

// Start initializes and starts all server components.
func (s *Server) Start() error {
	if !s.state.CompareAndSwap(stateNew, stateStarting) {
		return errors.New("server already started")
	}

	s.startTime = time.Now()
	s.hooks.RunPreStart(s)

	// Initialize storage
	if err := s.initStorage(); err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	// Initialize catalog
	if err := s.initCatalog(); err != nil {
		return fmt.Errorf("init catalog: %w", err)
	}

	// Initialize compute engine
	if err := s.initEngine(); err != nil {
		return fmt.Errorf("init engine: %w", err)
	}

	// Initialize observability
	if err := s.initObservability(); err != nil {
		return fmt.Errorf("init observability: %w", err)
	}

	// Initialize MySQL server
	if err := s.initMySQLServer(); err != nil {
		return fmt.Errorf("init mysql server: %w", err)
	}

	s.state.Store(stateRunning)
	s.hooks.RunPostStart(s)

	// Wait for stop signal
	<-s.stopChan

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.state.CompareAndSwap(stateRunning, stateStopping) {
		return errors.New("server not running")
	}

	defer func() {
		s.state.Store(stateStopped)
		close(s.doneChan)
	}()

	s.hooks.RunPreStop(s)

	// Stop accepting new connections
	if s.mysqlServer != nil {
		s.logger.Info("Stopping MySQL server...")
		if err := s.mysqlServer.Close(); err != nil {
			s.logger.Error("Error closing MySQL server", "error", err)
		}
	}

	// Wait for active connections with timeout
	if err := s.drainConnections(ctx); err != nil {
		s.logger.Warn("Drain connections timeout", "error", err)
	}

	// Stop observability server
	if s.obsServer != nil {
		s.logger.Info("Stopping observability server...")
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.obsServer.Stop(stopCtx); err != nil {
			s.logger.Error("Error stopping observability server", "error", err)
		}
	}

	// Close storage
	if s.storage != nil {
		s.logger.Info("Closing storage...")
		if err := s.storage.Close(); err != nil {
			s.logger.Error("Error closing storage", "error", err)
		}
	}

	s.hooks.RunPostStop(s)

	// Signal Start() to return
	close(s.stopChan)

	return nil
}

// IsReady returns true if the server is ready to serve requests.
func (s *Server) IsReady() bool {
	return s.state.Load() == stateRunning
}

// Uptime returns the server uptime.
func (s *Server) Uptime() time.Duration {
	if s.startTime.IsZero() {
		return 0
	}
	return time.Since(s.startTime)
}

// DSN returns a MySQL connection string for the server.
func (s *Server) DSN() string {
	return fmt.Sprintf("root@tcp(%s:%d)/", s.cfg.Server.Host, s.cfg.Server.Port)
}

// initStorage initializes the storage layer.
func (s *Server) initStorage() error {
	s.logger.Info("Initializing storage", "data_dir", s.cfg.Storage.DataDir)

	// Convert new config to old common/config format for compatibility
	legacyCfg := &commonConfig.Config{
		Storage: commonConfig.StorageConfig{
			Engine:  "badger",
			DataDir: s.cfg.Storage.DataDir,
			Badger: commonConfig.BadgerConfig{
				SyncWrites: s.cfg.Storage.SyncWrites,
			},
		},
	}

	storage, err := sal.NewAdapter(legacyCfg)
	if err != nil {
		return err
	}

	s.storage = storage
	return nil
}

// initCatalog initializes the catalog service.
func (s *Server) initCatalog() error {
	s.logger.Info("Initializing catalog")
	s.catalog = sql.NewCatalog()
	return nil
}

// initEngine initializes the SQL engine.
func (s *Server) initEngine() error {
	s.logger.Info("Initializing SQL engine")
	s.analyzer = analyzer.NewAnalyzer(s.catalog)
	s.optimizer = optimizer.NewOptimizer()
	s.engine = executor.NewEngine(s.analyzer, s.optimizer, s.catalog)
	return nil
}

// initObservability initializes observability components.
func (s *Server) initObservability() error {
	if !s.cfg.Observability.Enabled {
		s.logger.Info("Observability disabled")
		return nil
	}

	s.logger.Info("Initializing observability", "address", s.cfg.Observability.Address)

	// Create health checker
	checker := health.NewChecker()

	// Start observability server (health + metrics)
	obsCfg := observability.ServerConfig{
		Enabled:     s.cfg.Observability.Enabled,
		Address:     s.cfg.Observability.Address,
		MetricsPath: s.cfg.Observability.MetricsPath,
		EnablePprof: s.cfg.Observability.EnablePprof,
	}

	s.obsServer = observability.NewServer(obsCfg, checker, nil)

	if err := s.obsServer.Start(); err != nil {
		return err
	}

	return nil
}

// initMySQLServer initializes the MySQL protocol server.
func (s *Server) initMySQLServer() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.logger.Info("Initializing MySQL server", "address", addr)

	// For now, use no authentication (can be extended later)
	serverCfg := mysql.Config{
		Protocol: "tcp",
		Address:  addr,
		Auth:     nil, // Will use auth.NewNone() internally
	}

	mysqlSrv, err := mysql.NewDefaultServer(serverCfg, s.engine)
	if err != nil {
		return err
	}

	s.mysqlServer = mysqlSrv

	// Start MySQL server in goroutine
	go func() {
		s.mysqlServer.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// drainConnections waits for active connections to complete.
func (s *Server) drainConnections(ctx context.Context) error {
	// Simple implementation - wait for a grace period
	// In a full implementation, this would query the MySQL server for active connection count
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(s.cfg.Server.ShutdownTimeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check if we've exceeded the shutdown timeout
			if time.Now().After(deadline) {
				return fmt.Errorf("shutdown timeout exceeded")
			}
			// TODO: Query actual connection count from MySQL server
			// For now, wait a minimal period
			return nil
		}
	}
}


