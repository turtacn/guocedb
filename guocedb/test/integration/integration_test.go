package integration

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/executor"
	mysql_server "github.com/turtacn/guocedb/protocol/mysql"
	"github.com/turtacn/guocedb/storage/engines/badger"
	"github.com/turtacn/guocedb/common/config"
)

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// setupTestServer starts an in-memory guocedb server for testing.
func setupTestServer(t *testing.T) (dsn string, cleanup func()) {
	require := require.New(t)

	// Use a temporary directory for BadgerDB
	tempDir := t.TempDir()

	storageCfg := &config.StorageConfig{
		Engine: "badger",
		Path:   tempDir,
	}
	storage, err := badger.NewBadgerStorage(storageCfg)
	require.NoError(err)

	engine := executor.NewEngine(storage)

	port, err := getFreePort()
	require.NoError(err)

	srv, err := mysql_server.NewServer("127.0.0.1", port, engine)
	require.NoError(err)

	go func() {
		// This will error on server close, so we ignore it in the test.
		_ = srv.Start()
	}()

	// Wait for the server to be ready
	time.Sleep(100 * time.Millisecond)

	dsn = fmt.Sprintf("root:password@tcp(127.0.0.1:%d)/?multiStatements=true", port)

	cleanup = func() {
		srv.Stop()
		storage.Close()
	}

	return dsn, cleanup
}

func TestSimpleQuery(t *testing.T) {
	dsn, cleanup := setupTestServer(t)
	defer cleanup()

	require := require.New(t)

	db, err := sql.Open("mysql", dsn)
	require.NoError(err)
	defer db.Close()

	// Wait for connection to be ready
	require.NoError(db.Ping())

	// Simple SELECT query
	rows, err := db.Query("SELECT 1")
	require.NoError(err)

	require.True(rows.Next())
	var result int
	err = rows.Scan(&result)
	require.NoError(err)
	require.Equal(1, result)

	require.False(rows.Next())
	require.NoError(rows.Close())
}

func TestDDLAndDML(t *testing.T) {
	dsn, cleanup := setupTestServer(t)
	defer cleanup()

	require := require.New(t)

	db, err := sql.Open("mysql", dsn)
	require.NoError(err)
	defer db.Close()

	require.NoError(db.Ping())

	_, err = db.Exec("CREATE DATABASE testdb")
	require.NoError(err)

	_, err = db.Exec("USE testdb")
	require.NoError(err)

	_, err = db.Exec("CREATE TABLE test (id INT PRIMARY KEY, name VARCHAR(100))")
	require.NoError(err)

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'jules')")
	require.NoError(err)

	rows, err := db.Query("SELECT name FROM test WHERE id = 1")
	require.NoError(err)

	require.True(rows.Next())
	var name string
	err = rows.Scan(&name)
	require.NoError(err)
	require.Equal("jules", name)

	require.NoError(rows.Close())
}
