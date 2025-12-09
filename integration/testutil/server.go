package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/turtacn/guocedb/config"
	"github.com/turtacn/guocedb/server"
)

// TestServer wraps a GuoceDB server instance for testing
type TestServer struct {
	t       *testing.T
	srv     *server.Server
	cfg     *config.Config
	dataDir string
	port    int
}

// TestServerOption configures a TestServer
type TestServerOption func(*TestServer)

// WithPort sets a specific port (0 for random)
func WithPort(port int) TestServerOption {
	return func(ts *TestServer) {
		ts.port = port
	}
}

// WithAuth enables authentication with a root password
func WithAuth(enabled bool, rootPass string) TestServerOption {
	return func(ts *TestServer) {
		ts.cfg.Security.Enabled = enabled
		ts.cfg.Security.RootPassword = rootPass
	}
}

// WithDataDir sets a specific data directory
func WithDataDir(dir string) TestServerOption {
	return func(ts *TestServer) {
		ts.dataDir = dir
	}
}

// NewTestServer creates and configures a test server
func NewTestServer(t *testing.T, opts ...TestServerOption) *TestServer {
	t.Helper()

	// Create temporary directory if not specified
	dataDir := t.TempDir()

	// Find free port if not specified
	port := findFreePort(t)

	ts := &TestServer{
		t:       t,
		dataDir: dataDir,
		port:    port,
	}

	// Apply options
	for _, opt := range opts {
		opt(ts)
	}

	// Create configuration
	ts.cfg = &config.Config{
		Server: config.ServerConfig{
			Host:            "127.0.0.1",
			Port:            ts.port,
			MaxConnections:  100,
			ConnectTimeout:  10 * time.Second,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     8 * time.Hour,
			ShutdownTimeout: 5 * time.Second,
		},
		Storage: config.StorageConfig{
			DataDir:         ts.dataDir,
			MaxMemTableSize: 64 << 20, // 64MB - meets minimum 1MB requirement
			NumCompactors:   2,         // Positive number required
			SyncWrites:      false,     // Faster for testing
		},
		Observability: config.ObservabilityConfig{
			Enabled: false, // Disable for tests
		},
		Logging: config.LoggingConfig{
			Level:  "error", // Less noise in tests
			Format: "text",
			Output: "stdout",
		},
	}

	return ts
}

// Start launches the test server
func (ts *TestServer) Start() *TestServer {
	ts.t.Helper()

	var err error
	ts.srv, err = server.New(ts.cfg)
	if err != nil {
		ts.t.Fatalf("failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		if err := ts.srv.Start(); err != nil {
			ts.t.Logf("server stopped: %v", err)
		}
	}()

	// Wait for server to be ready
	if !ts.waitReady(10 * time.Second) {
		ts.t.Fatal("server not ready in time")
	}

	return ts
}

// Stop gracefully shuts down the test server
func (ts *TestServer) Stop() {
	if ts.srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.srv.Shutdown(ctx)
	}
}

// DSN returns a MySQL DSN for connecting to the test server
func (ts *TestServer) DSN() string {
	if ts.cfg.Security.Enabled {
		return fmt.Sprintf("root:%s@tcp(127.0.0.1:%d)/",
			ts.cfg.Security.RootPassword, ts.port)
	}
	// Use root with no password (native auth compatible)
	return fmt.Sprintf("root@tcp(127.0.0.1:%d)/", ts.port)
}

// DB creates a new database connection
func (ts *TestServer) DB() *sql.DB {
	db, err := sql.Open("mysql", ts.DSN())
	if err != nil {
		ts.t.Fatalf("failed to open connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		ts.t.Fatalf("failed to ping database: %v", err)
	}
	return db
}

// MustExec executes a query and fails the test on error
func (ts *TestServer) MustExec(query string, args ...interface{}) {
	db := ts.DB()
	defer db.Close()
	if _, err := db.Exec(query, args...); err != nil {
		ts.t.Fatalf("exec %q failed: %v", query, err)
	}
}

// Port returns the server port
func (ts *TestServer) Port() int {
	return ts.port
}

// DataDir returns the data directory
func (ts *TestServer) DataDir() string {
	return ts.dataDir
}

// Restart stops and restarts the server
func (ts *TestServer) Restart() {
	ts.Stop()
	time.Sleep(100 * time.Millisecond)
	ts.Start()
}

// waitReady waits for the server to accept connections
func (ts *TestServer) waitReady(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp",
			fmt.Sprintf("127.0.0.1:%d", ts.port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			// Give server a bit more time to fully initialize
			time.Sleep(100 * time.Millisecond)
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// findFreePort finds an available TCP port
func findFreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	// Small delay to ensure port is released
	time.Sleep(10 * time.Millisecond)
	return port
}
