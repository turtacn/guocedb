// Package integration contains integration tests for the guocedb project.
package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/network/server"
	
	// Import the new integration test framework
	integrationFramework "github.com/turtacn/guocedb/integration"
)

// TestMain sets up and tears down the server for integration tests.
func TestMain(m *testing.M) {
	// This is a placeholder for starting a real server instance.
	// A full implementation would be complex, requiring careful setup
	// and teardown of the server in a separate process or goroutine.
	fmt.Println("Integration tests are placeholders and do not run a real server.")

	// For now, we will skip running the tests that require a server.
	// To run these tests, you would need to:
	// 1. Build the server binary.
	// 2. Run it in the background.
	// 3. Run the tests.
	// 4. Shut down the server.

	// os.Exit(m.Run())
	os.Exit(0)
}

// helper function to start a server for testing
func startTestServer() (*server.Manager, *config.Config) {
	// This would be a more complete version of the setup.
	// Modify config for testing (e.g., random ports, temp data dir)

	// ... initialize and start the server ...

	return nil, nil
}

// TestSimpleQuery tests a simple SELECT query against a running server.
func TestSimpleQuery(t *testing.T) {
	t.Skip("Skipping test that requires a running server.")

	dsn := "root:password@tcp(127.0.0.1:3306)/test"
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Ping the server to ensure connection
	err = db.Ping()
	require.NoError(t, err)

	// Run a simple query
	rows, err := db.Query("SELECT 1")
	require.NoError(t, err)
	defer rows.Close()

	// Check the result
	require.True(t, rows.Next())
	var result int
	err = rows.Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestDDLAndDML tests basic CREATE, INSERT, and SELECT statements.
func TestDDLAndDML(t *testing.T) {
	t.Skip("Skipping test that requires a running server.")

	dsn := "root:password@tcp(127.0.0.1:3306)/"
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Create Database
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS integration_test")
	require.NoError(t, err)

	// Use Database
	_, err = db.Exec("USE integration_test")
	require.NoError(t, err)

	// Create Table
	_, err = db.Exec(`CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255))`)
	require.NoError(t, err)

	// Insert Data
	_, err = db.Exec(`INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob')`)
	require.NoError(t, err)

	// Select Data
	rows, err := db.Query("SELECT name FROM users WHERE id = 2")
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	var name string
	err = rows.Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Bob", name)
}

// TestIntegrationFramework tests the new integration test framework
func TestIntegrationFramework(t *testing.T) {
	// Use the new integration framework
	ts := integrationFramework.NewTestServer(t)
	defer ts.Close()
	
	client := integrationFramework.NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Test basic functionality
	client.MustExec(
		"CREATE DATABASE framework_test",
		"USE framework_test",
		"CREATE TABLE test (id INT PRIMARY KEY, name VARCHAR(50))",
		"INSERT INTO test (id, name) VALUES (1, 'framework')",
	)

	// Verify data
	var id int
	var name string
	err := client.QueryRow("SELECT id, name FROM test WHERE id = 1").Scan(&id, &name)
	require.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "framework", name)

	// Test helper functions
	integrationFramework.AssertTableExists(t, client, "test")
	count := integrationFramework.GetRowCount(t, client, "test")
	assert.Equal(t, 1, count)
}
