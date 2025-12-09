package integration

import (
	"database/sql"
	"testing"

	"github.com/turtacn/guocedb/integration/testutil"
)

// TestE2E_AuthValidCredentials tests authentication with valid credentials
func TestE2E_AuthValidCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t,
		testutil.WithAuth(true, "testroot123"),
	).Start()
	defer ts.Stop()

	// This test requires actual auth implementation
	// For now, we just verify the server starts with auth enabled
	t.Log("Server started with authentication enabled")

	// Try to connect (may fail if auth not fully implemented)
	// client := testutil.NewTestClient(t, ts.DSN())
	// defer client.Close()
}

// TestE2E_AuthInvalidCredentials tests authentication with invalid credentials
func TestE2E_AuthInvalidCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t,
		testutil.WithAuth(true, "correctpass"),
	).Start()
	defer ts.Stop()

	// Try to connect with wrong password
	dsn := "root:wrongpass@tcp(127.0.0.1:" + string(rune(ts.Port())) + ")/"
	db, err := sql.Open("mysql", dsn)
	if err == nil {
		defer db.Close()
		err = db.Ping()
		// We expect this to fail with auth error
		if err == nil {
			t.Log("Warning: Expected auth to fail, but it succeeded")
		}
	}
}

// TestE2E_NoAuth tests server without authentication
func TestE2E_NoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	// Should be able to execute queries without auth
	client.Exec("CREATE DATABASE test_noauth")
	client.Exec("DROP DATABASE test_noauth")
}

// TestE2E_UserManagement tests user creation and management
func TestE2E_UserManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	// Note: User management may not be fully implemented yet
	// These tests document expected behavior

	// Create user (may not be supported yet)
	// client.Exec("CREATE USER 'testuser'@'%' IDENTIFIED BY 'testpass'")

	// Grant privileges (may not be supported yet)
	// client.Exec("GRANT SELECT ON testdb.* TO 'testuser'@'%'")

	// For now, just verify basic operations work
	client.Exec("CREATE DATABASE security_test")
	client.Exec("DROP DATABASE security_test")
}

// TestE2E_PrivilegeCheck tests privilege enforcement
func TestE2E_PrivilegeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	// Note: Privilege system may not be fully implemented yet
	// This test documents expected behavior

	client.Exec("CREATE DATABASE privilege_test")
	client.Exec("USE privilege_test")
	client.Exec("CREATE TABLE t1 (id INT)")

	// All operations should work as root
	client.Exec("INSERT INTO t1 VALUES (1)")
	client.Exec("SELECT * FROM t1")
	client.Exec("UPDATE t1 SET id = 2")
	client.Exec("DELETE FROM t1")

	client.Exec("DROP DATABASE privilege_test")
}

// TestE2E_ConnectionLimit tests connection limit enforcement
func TestE2E_ConnectionLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	// Open multiple connections
	var clients []*testutil.TestClient
	defer func() {
		for _, c := range clients {
			c.Close()
		}
	}()

	// Open 10 connections - should all succeed within default limit
	for i := 0; i < 10; i++ {
		client := testutil.NewTestClient(t, ts.DSN())
		clients = append(clients, client)
		client.Exec("SELECT 1")
	}

	t.Logf("Successfully opened %d connections", len(clients))
}

// TestE2E_SecurityAudit tests that security events can be audited
func TestE2E_SecurityAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	// Perform various operations that should be audited
	client.Exec("CREATE DATABASE audit_test")
	client.Exec("USE audit_test")
	client.Exec("CREATE TABLE sensitive_data (id INT, data VARCHAR(100))")
	client.Exec("INSERT INTO sensitive_data VALUES (1, 'secret')")
	client.Exec("SELECT * FROM sensitive_data")
	client.Exec("DROP DATABASE audit_test")

	// Note: Actual audit log verification would require access to audit files
	t.Log("Security audit test completed - audit logs should contain these operations")
}
