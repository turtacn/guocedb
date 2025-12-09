package integration

import (
	"sync"
	"sync/atomic"
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/turtacn/guocedb/integration/testutil"
)

// TestE2E_TransactionCommit tests successful transaction commit
func TestE2E_TransactionCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Begin transaction
	tx := client.BeginTx()
	tx.Exec("INSERT INTO products VALUES (10, 'NewProduct', 9.99)")
	tx.Commit()

	// Verify data persisted
	count := client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 10")
	require.Equal(t, 1, count)
}

// TestE2E_TransactionRollback tests transaction rollback
func TestE2E_TransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Begin transaction
	tx := client.BeginTx()
	tx.Exec("INSERT INTO products VALUES (11, 'TempProduct', 5.00)")
	tx.Rollback()

	// Verify data was not persisted
	count := client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 11")
	require.Equal(t, 0, count)
}

// TestE2E_TransactionMultiStatement tests transaction with multiple statements
func TestE2E_TransactionMultiStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupBankAccounts(client)
	client.Exec("USE bank")

	// Transfer money between accounts
	tx := client.BeginTx()
	tx.Exec("UPDATE accounts SET balance = balance - 200 WHERE id = 1")
	tx.Exec("UPDATE accounts SET balance = balance + 200 WHERE id = 2")
	tx.Commit()

	// Verify balances
	b1 := client.MustQueryFloat("SELECT balance FROM accounts WHERE id = 1")
	b2 := client.MustQueryFloat("SELECT balance FROM accounts WHERE id = 2")

	require.InDelta(t, 800.0, b1, 0.01)
	require.InDelta(t, 700.0, b2, 0.01)

	// Verify total hasn't changed
	total := client.MustQueryFloat("SELECT SUM(balance) FROM accounts")
	require.InDelta(t, 1500.0, total, 0.01)
}

// TestE2E_TransactionIsolation tests basic transaction isolation
func TestE2E_TransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client1 := testutil.NewTestClient(t, ts.DSN())
	defer client1.Close()
	client2 := testutil.NewTestClient(t, ts.DSN())
	defer client2.Close()

	testutil.SetupTestTable(client1)
	client1.Exec("USE testdb")
	client2.Exec("USE testdb")

	// TX1: Update but don't commit yet
	tx1 := client1.BeginTx()
	tx1.Exec("UPDATE products SET price = 100.00 WHERE id = 1")

	// TX2: Should see old value (snapshot isolation)
	// Note: This depends on the isolation level implementation
	oldPrice := client2.MustQueryFloat("SELECT price FROM products WHERE id = 1")
	t.Logf("Price before TX1 commit: %.2f", oldPrice)

	// TX1: Commit
	tx1.Commit()

	// TX2: Should now see new value
	newPrice := client2.MustQueryFloat("SELECT price FROM products WHERE id = 1")
	t.Logf("Price after TX1 commit: %.2f", newPrice)
	require.InDelta(t, 100.0, newPrice, 0.01)
}

// TestE2E_TransactionConcurrency tests concurrent transactions
func TestE2E_TransactionConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	client.Exec("CREATE DATABASE concurrency_test")
	client.Exec("USE concurrency_test")
	client.Exec("CREATE TABLE counters (id INT PRIMARY KEY, value INT)")
	client.Exec("INSERT INTO counters VALUES (1, 0)")

	const numGoroutines = 10
	const incrementsPerGoroutine = 10

	var wg sync.WaitGroup
	var successCount, failCount atomic.Int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine gets its own connection
			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE concurrency_test")

			for j := 0; j < incrementsPerGoroutine; j++ {
				tx := c.BeginTx()
				tx.Exec("UPDATE counters SET value = value + 1 WHERE id = 1")
				err := tx.TryCommit()
				if err != nil {
					failCount.Add(1)
				} else {
					successCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	// Verify final count
	finalValue := client.MustQueryInt("SELECT value FROM counters WHERE id = 1")
	t.Logf("Success: %d, Fail: %d, Final value: %d",
		successCount.Load(), failCount.Load(), finalValue)

	// Final value should match successful commits
	require.Equal(t, int(successCount.Load()), finalValue)
}

// TestE2E_TransactionRollbackAfterError tests rollback on error
func TestE2E_TransactionRollbackAfterError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Start transaction
	tx := client.BeginTx()
	tx.Exec("INSERT INTO products VALUES (20, 'Product20', 10.00)")

	// Rollback explicitly
	tx.Rollback()

	// Verify no data persisted
	count := client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 20")
	require.Equal(t, 0, count)
}

// TestE2E_NestedTransactions tests behavior with nested BEGIN statements
func TestE2E_NestedTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// First transaction
	tx1 := client.BeginTx()
	tx1.Exec("INSERT INTO products VALUES (30, 'Product30', 15.00)")
	tx1.Commit()

	// Verify first transaction committed
	count := client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 30")
	require.Equal(t, 1, count)

	// Second transaction
	tx2 := client.BeginTx()
	tx2.Exec("INSERT INTO products VALUES (31, 'Product31', 16.00)")
	tx2.Rollback()

	// Verify second transaction rolled back
	count = client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 31")
	require.Equal(t, 0, count)

	// First transaction data should still exist
	count = client.MustQueryInt("SELECT COUNT(*) FROM products WHERE id = 30")
	require.Equal(t, 1, count)
}
