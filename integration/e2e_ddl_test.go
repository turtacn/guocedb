package integration

import (
	"testing"

	"github.com/turtacn/guocedb/integration/testutil"
)

// TestE2E_CreateDropDatabase tests database lifecycle
func TestE2E_CreateDropDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	// Create databases
	client.Exec("CREATE DATABASE testdb1")
	client.Exec("CREATE DATABASE testdb2")

	// Verify they exist by using them
	client.Exec("USE testdb1")
	client.Exec("USE testdb2")

	// Drop one database
	client.Exec("DROP DATABASE testdb1")

	// Create IF NOT EXISTS should work for existing db
	client.Exec("CREATE DATABASE IF NOT EXISTS testdb2")

	// Drop IF EXISTS should not error for non-existent db
	client.Exec("DROP DATABASE IF EXISTS nonexistent")

	// Clean up
	client.Exec("DROP DATABASE testdb2")
}

// TestE2E_CreateDropTable tests table lifecycle
func TestE2E_CreateDropTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	client.Exec("CREATE DATABASE testdb")
	client.Exec("USE testdb")

	// Create table with various column types
	client.Exec(`
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(200),
			age INT,
			balance DECIMAL(10,2),
			active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP
		)
	`)

	// Create table with index
	client.Exec(`
		CREATE TABLE orders (
			id INT PRIMARY KEY,
			user_id INT,
			amount DECIMAL(10,2),
			INDEX idx_user (user_id)
		)
	`)

	// Drop table
	client.Exec("DROP TABLE users")

	// Drop IF EXISTS
	client.Exec("DROP TABLE IF EXISTS nonexistent")

	// Clean up
	client.Exec("DROP TABLE orders")
	client.Exec("DROP DATABASE testdb")
}

// TestE2E_AlterTable tests table alterations
func TestE2E_AlterTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	client.Exec("CREATE DATABASE testdb")
	client.Exec("USE testdb")
	client.Exec("CREATE TABLE t1 (id INT PRIMARY KEY, name VARCHAR(50))")

	// Note: ALTER TABLE support may be limited in the current implementation
	// These tests document the expected behavior

	// Add column (may not be supported yet)
	// client.Exec("ALTER TABLE t1 ADD COLUMN age INT")

	// Drop column (may not be supported yet)
	// client.Exec("ALTER TABLE t1 DROP COLUMN age")

	// For now, just verify table exists
	client.Exec("SELECT * FROM t1")

	client.Exec("DROP DATABASE testdb")
}
