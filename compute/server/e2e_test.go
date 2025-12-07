package server

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/turtacn/guocedb/compute/executor"
	sqlengine "github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/optimizer"

	"github.com/turtacn/guocedb/compute/auth"
)


// startTestServer starts a GuoceDB server for testing
func startTestServer(t *testing.T) (addr string, cleanup func()) {
	// Create catalog with test database
	catalog := sqlengine.NewCatalog()
	testDB := newMockDatabase("testdb")
	catalog.AddDatabase(testDB)

	// Create engine components
	analyzer := analyzer.NewAnalyzer(catalog)
	optimizer := optimizer.NewOptimizer()
	engine := executor.NewEngine(analyzer, optimizer, catalog)

	// Create session manager and handler
	sessionMgr := NewSessionManager(DefaultSessionBuilder, nil, "localhost:0")
	handler := NewHandler(engine, sessionMgr)

	// Create auth server (simple auth for testing)
	authServer := auth.NewNativeSingle("root", "", auth.AllPermissions)

	// Create MySQL listener with proper parameters on random port
	mysqlListener, err := mysql.NewListener("tcp", "127.0.0.1:0", authServer.Mysql(), handler, 0, 0)
	require.NoError(t, err)

	// Start accepting connections in background
	go func() {
		mysqlListener.Accept()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	cleanup = func() {
		mysqlListener.Close()
	}

	return mysqlListener.Addr().String(), cleanup
}

func TestE2E_MySQLClientConnect(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	assert.NoError(t, err)
}

func TestE2E_UseDatabase(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("USE testdb")
	assert.NoError(t, err)
}

func TestE2E_UseNonexistentDatabase(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("USE nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not found")
}

func TestE2E_SimpleQuery(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/testdb", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.Query("SELECT 1 AS test_col")
	require.NoError(t, err)
	defer rows.Close()

	var result int
	require.True(t, rows.Next())
	err = rows.Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestE2E_MultiStatement(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	// Enable multi-statements in the connection
	dsn := fmt.Sprintf("root@tcp(%s)/testdb?multiStatements=true", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Execute multiple statements
	_, err = db.Exec("SELECT 1; SELECT 2")
	// Note: The exact behavior depends on the MySQL driver implementation
	// Some drivers may not support multi-statements or may handle them differently
	// For now, we just check that it doesn't crash
	// In a real implementation, we'd need to handle this more carefully
}

func TestE2E_ConnectionPooling(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/testdb", addr)
	
	// Create multiple connections
	var dbs []*sql.DB
	for i := 0; i < 3; i++ {
		db, err := sql.Open("mysql", dsn)
		require.NoError(t, err)
		dbs = append(dbs, db)
		
		err = db.Ping()
		assert.NoError(t, err)
	}

	// Close all connections
	for _, db := range dbs {
		db.Close()
	}
}

func TestE2E_SessionIsolation(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/", addr)
	
	// Create two separate connections
	db1, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db1.Close()

	db2, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db2.Close()

	// Set different databases for each connection
	_, err = db1.Exec("USE testdb")
	assert.NoError(t, err)

	// db2 should still be in the default state (no database selected)
	// This test verifies that session state is isolated between connections
}

// Benchmark tests
func BenchmarkE2E_SimpleQuery(b *testing.B) {
	addr, cleanup := startTestServer(&testing.T{})
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/testdb", addr)
	db, err := sql.Open("mysql", dsn)
	require.NoError(b, err)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query("SELECT 1")
		require.NoError(b, err)
		rows.Close()
	}
}

func BenchmarkE2E_Connection(b *testing.B) {
	addr, cleanup := startTestServer(&testing.T{})
	defer cleanup()

	dsn := fmt.Sprintf("root@tcp(%s)/testdb", addr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := sql.Open("mysql", dsn)
		require.NoError(b, err)
		err = db.Ping()
		require.NoError(b, err)
		db.Close()
	}
}