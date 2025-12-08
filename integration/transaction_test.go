package integration

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTxn_BasicCommit tests basic transaction commit functionality
func TestTxn_BasicCommit(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)

	// Begin transaction
	client.Exec("BEGIN")

	// Update within transaction
	client.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Verify change is visible within transaction
	var val int
	err := client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)

	// Commit transaction
	client.Exec("COMMIT")

	// Verify change persists after commit
	err = client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)
}

// TestTxn_BasicRollback tests basic transaction rollback functionality
func TestTxn_BasicRollback(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)

	// Begin transaction
	client.Exec("BEGIN")

	// Update within transaction
	client.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Verify change is visible within transaction
	var val int
	err := client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)

	// Rollback transaction
	client.Exec("ROLLBACK")

	// Verify change was rolled back
	err = client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 100, val) // Should be original value
}

// TestTxn_ReadCommitted tests READ COMMITTED isolation level
func TestTxn_ReadCommitted(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	// Client A
	clientA := NewTestClient(t, ts.DSN(""))
	defer clientA.Close()
	
	// Client B
	clientB := NewTestClient(t, ts.DSN(""))
	defer clientB.Close()

	// Setup
	clientA.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)
	clientB.Exec("USE testdb")

	// Client A begins transaction and updates but doesn't commit
	clientA.Exec("BEGIN")
	clientA.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Client B should see the old value (no dirty read)
	var val int
	err := clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 100, val) // Should see original value

	// Client A commits
	clientA.Exec("COMMIT")

	// Client B should now see the new value
	err = clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val) // Should see committed value
}

// TestTxn_DirtyReadPrevented tests that dirty reads are prevented
func TestTxn_DirtyReadPrevented(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	clientA := NewTestClient(t, ts.DSN(""))
	defer clientA.Close()
	
	clientB := NewTestClient(t, ts.DSN(""))
	defer clientB.Close()

	// Setup
	clientA.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)
	clientB.Exec("USE testdb")

	// Client A begins transaction and updates
	clientA.Exec("BEGIN")
	clientA.Exec("UPDATE t SET val = 999 WHERE id = 1")

	// Client B should not see the uncommitted change
	var val int
	err := clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 100, val) // Should see original value, not 999

	// Client A rolls back
	clientA.Exec("ROLLBACK")

	// Client B should still see the original value
	err = clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 100, val)
}

// TestTxn_MultipleOperations tests transactions with multiple operations
func TestTxn_MultipleOperations(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance DECIMAL(10,2))",
		"INSERT INTO accounts (id, balance) VALUES (1, 1000.00), (2, 500.00)",
	)

	// Begin transaction for money transfer
	client.Exec("BEGIN")

	// Transfer $200 from account 1 to account 2
	client.Exec("UPDATE accounts SET balance = balance - 200.00 WHERE id = 1")
	client.Exec("UPDATE accounts SET balance = balance + 200.00 WHERE id = 2")

	// Verify balances within transaction
	var balance1, balance2 float64
	err := client.QueryRow("SELECT balance FROM accounts WHERE id = 1").Scan(&balance1)
	require.NoError(t, err)
	err = client.QueryRow("SELECT balance FROM accounts WHERE id = 2").Scan(&balance2)
	require.NoError(t, err)
	
	assert.Equal(t, 800.00, balance1)
	assert.Equal(t, 700.00, balance2)

	// Commit transaction
	client.Exec("COMMIT")

	// Verify balances persist after commit
	err = client.QueryRow("SELECT balance FROM accounts WHERE id = 1").Scan(&balance1)
	require.NoError(t, err)
	err = client.QueryRow("SELECT balance FROM accounts WHERE id = 2").Scan(&balance2)
	require.NoError(t, err)
	
	assert.Equal(t, 800.00, balance1)
	assert.Equal(t, 700.00, balance2)
}

// TestTxn_RollbackMultipleOperations tests rollback of multiple operations
func TestTxn_RollbackMultipleOperations(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance DECIMAL(10,2))",
		"INSERT INTO accounts (id, balance) VALUES (1, 1000.00), (2, 500.00)",
	)

	// Begin transaction
	client.Exec("BEGIN")

	// Perform multiple operations
	client.Exec("UPDATE accounts SET balance = balance - 300.00 WHERE id = 1")
	client.Exec("UPDATE accounts SET balance = balance + 300.00 WHERE id = 2")
	client.Exec("INSERT INTO accounts (id, balance) VALUES (3, 100.00)")

	// Rollback transaction
	client.Exec("ROLLBACK")

	// Verify all changes were rolled back
	var balance1, balance2 float64
	err := client.QueryRow("SELECT balance FROM accounts WHERE id = 1").Scan(&balance1)
	require.NoError(t, err)
	err = client.QueryRow("SELECT balance FROM accounts WHERE id = 2").Scan(&balance2)
	require.NoError(t, err)
	
	assert.Equal(t, 1000.00, balance1) // Original value
	assert.Equal(t, 500.00, balance2)  // Original value

	// Verify inserted row was rolled back
	var count int
	err = client.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = 3").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestTxn_AutoCommit tests autocommit behavior
func TestTxn_AutoCommit(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)

	// Without explicit transaction, changes should be auto-committed
	client.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Verify change is immediately visible
	var val int
	err := client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)
}

// TestTxn_NestedTransactions tests behavior with nested BEGIN statements
func TestTxn_NestedTransactions(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	// Setup
	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)

	// Begin first transaction
	client.Exec("BEGIN")
	client.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Attempt to begin another transaction (should handle gracefully)
	// Different databases handle this differently - some error, some ignore
	// We'll test that the system doesn't crash
	_, err := client.db.Exec("BEGIN")
	// Don't assert on error - just ensure system is stable

	// Commit should work
	client.Exec("COMMIT")

	// Verify final state
	var val int
	err = client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)
}

// TestTxn_ConcurrentInsert tests concurrent inserts without conflicts
func TestTxn_ConcurrentInsert(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
	)
	setupClient.Close()

	var wg sync.WaitGroup
	numClients := 5
	insertsPerClient := 10

	// Launch concurrent clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()

			for j := 0; j < insertsPerClient; j++ {
				id := clientID*insertsPerClient + j + 1
				client.Exec("INSERT INTO t (id, val) VALUES (?, ?)", id, id*10)
			}
		}(i)
	}

	wg.Wait()

	// Verify all inserts succeeded
	verifyClient := NewTestClient(t, ts.DSN("testdb"))
	defer verifyClient.Close()
	
	var count int
	err := verifyClient.QueryRow("SELECT COUNT(*) FROM t").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, numClients*insertsPerClient, count)
}

// TestTxn_ConcurrentUpdate tests concurrent updates to different rows
func TestTxn_ConcurrentUpdate(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
	)
	
	// Insert test data
	for i := 1; i <= 10; i++ {
		setupClient.Exec("INSERT INTO t (id, val) VALUES (?, ?)", i, i*10)
	}
	setupClient.Close()

	var wg sync.WaitGroup
	numClients := 5

	// Each client updates different rows
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()

			// Each client updates 2 specific rows
			id1 := clientID*2 + 1
			id2 := clientID*2 + 2
			
			client.Exec("BEGIN")
			client.Exec("UPDATE t SET val = val + 1000 WHERE id = ?", id1)
			client.Exec("UPDATE t SET val = val + 1000 WHERE id = ?", id2)
			client.Exec("COMMIT")
		}(i)
	}

	wg.Wait()

	// Verify all updates succeeded
	verifyClient := NewTestClient(t, ts.DSN("testdb"))
	defer verifyClient.Close()
	
	rows := verifyClient.Query("SELECT id, val FROM t ORDER BY id")
	results := CollectRows(t, rows)
	
	assert.Len(t, results, 10)
	for i, result := range results {
		expectedVal := (i+1)*10 + 1000 // Original value + 1000
		assert.Equal(t, int64(expectedVal), result["val"])
	}
}

// TestTxn_LongTransaction tests behavior with long-running transactions
func TestTxn_LongTransaction(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	clientA := NewTestClient(t, ts.DSN(""))
	defer clientA.Close()
	
	clientB := NewTestClient(t, ts.DSN(""))
	defer clientB.Close()

	// Setup
	clientA.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)
	clientB.Exec("USE testdb")

	// Client A starts a long transaction
	clientA.Exec("BEGIN")
	clientA.Exec("UPDATE t SET val = 200 WHERE id = 1")

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Client B should still see the original value
	var val int
	err := clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 100, val)

	// Client A commits after delay
	clientA.Exec("COMMIT")

	// Client B should now see the updated value
	err = clientB.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)
}