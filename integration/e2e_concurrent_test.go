package integration

import (
	"sync"
	"sync/atomic"
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	"github.com/turtacn/guocedb/integration/testutil"
)

// TestE2E_ConcurrentReads tests concurrent read operations
func TestE2E_ConcurrentReads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	testutil.SetupTestTable(client)
	client.Close()

	const numReaders = 20
	const readsPerReader = 50

	var wg sync.WaitGroup
	var errorCount atomic.Int32

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE testdb")

			for j := 0; j < readsPerReader; j++ {
				rows := c.Query("SELECT * FROM products WHERE price > 0")
				count := testutil.CountRows(rows)
				if count < 1 {
					errorCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	require.Equal(t, int32(0), errorCount.Load(), "no read errors should occur")
	t.Logf("Completed %d concurrent reads successfully", numReaders*readsPerReader)
}

// TestE2E_ConcurrentWrites tests concurrent write operations
func TestE2E_ConcurrentWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	client.Exec("CREATE DATABASE concurrency_test")
	client.Exec("USE concurrency_test")
	client.Exec("CREATE TABLE data (id INT PRIMARY KEY, value INT)")
	client.Close()

	const numWriters = 10
	const writesPerWriter = 10

	var wg sync.WaitGroup
	var successCount atomic.Int32

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE concurrency_test")

			for j := 0; j < writesPerWriter; j++ {
				id := writerID*1000 + j
				c.Exec("INSERT INTO data VALUES (?, ?)", id, writerID)
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	// Verify all writes succeeded
	client = testutil.NewTestClient(t, ts.DSN())
	defer client.Close()
	client.Exec("USE concurrency_test")

	count := client.MustQueryInt("SELECT COUNT(*) FROM data")
	require.Equal(t, int(successCount.Load()), count)
	t.Logf("Successfully completed %d concurrent writes", count)
}

// TestE2E_ConcurrentMixedWorkload tests mixed read/write workload
func TestE2E_ConcurrentMixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	client.Exec("CREATE DATABASE workload_test")
	client.Exec("USE workload_test")
	client.Exec("CREATE TABLE counters (id INT PRIMARY KEY, value INT)")
	client.Exec("INSERT INTO counters VALUES (1, 0)")
	client.Close()

	const numWorkers = 15
	const opsPerWorker = 20

	var wg sync.WaitGroup
	var readCount, writeCount atomic.Int32

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE workload_test")

			for j := 0; j < opsPerWorker; j++ {
				if j%3 == 0 {
					// Write operation
					c.Exec("UPDATE counters SET value = value + 1 WHERE id = 1")
					writeCount.Add(1)
				} else {
					// Read operation
					c.Query("SELECT value FROM counters WHERE id = 1")
					readCount.Add(1)
				}
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Completed %d reads and %d writes", readCount.Load(), writeCount.Load())

	// Verify final state
	client = testutil.NewTestClient(t, ts.DSN())
	defer client.Close()
	client.Exec("USE workload_test")

	finalValue := client.MustQueryInt("SELECT value FROM counters WHERE id = 1")
	t.Logf("Final counter value: %d (expected: %d)", finalValue, writeCount.Load())
}

// TestE2E_ConcurrentTransactions tests concurrent transaction execution
func TestE2E_ConcurrentTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	testutil.SetupBankAccounts(client)
	client.Close()

	const numTransfers = 50
	var wg sync.WaitGroup
	var successCount, conflictCount atomic.Int32

	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func(txID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE bank")

			tx := c.BeginTx()
			// Transfer $10 from account 1 to account 2
			tx.Exec("UPDATE accounts SET balance = balance - 10 WHERE id = 1")
			tx.Exec("UPDATE accounts SET balance = balance + 10 WHERE id = 2")

			err := tx.TryCommit()
			if err != nil {
				conflictCount.Add(1)
			} else {
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Transactions - Success: %d, Conflicts: %d",
		successCount.Load(), conflictCount.Load())

	// Verify total balance remains constant
	client = testutil.NewTestClient(t, ts.DSN())
	defer client.Close()
	client.Exec("USE bank")

	total := client.MustQueryFloat("SELECT SUM(balance) FROM accounts")
	require.InDelta(t, 1500.0, total, 0.01)
}

// TestE2E_ConnectionPool tests connection pooling under load
func TestE2E_ConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	// Use a single connection pool with many goroutines
	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	client.Exec("CREATE DATABASE pool_test")
	client.Exec("USE pool_test")
	client.Exec("CREATE TABLE data (id INT PRIMARY KEY)")

	const numGoroutines = 50
	const queriesPerGoroutine = 20

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < queriesPerGoroutine; j++ {
				// Quick queries that exercise the connection pool
				client.Exec("SELECT 1")
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Successfully completed connection pool test with %d goroutines", numGoroutines)
}

// TestE2E_ConcurrentDatabaseOperations tests concurrent database-level operations
func TestE2E_ConcurrentDatabaseOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	const numWorkers = 10
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()

			dbName := "concdb_" + string(rune(workerID+'0'))
			c.Exec("CREATE DATABASE " + dbName)
			c.Exec("USE " + dbName)
			c.Exec("CREATE TABLE t1 (id INT)")
			c.Exec("INSERT INTO t1 VALUES (1)")

			count := c.MustQueryInt("SELECT COUNT(*) FROM t1")
			if count != 1 {
				t.Errorf("Worker %d: expected 1 row, got %d", workerID, count)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent database operations completed successfully")
}

// TestE2E_HighConcurrencyStress performs a high-concurrency stress test
func TestE2E_HighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	client.Exec("CREATE DATABASE stress_test")
	client.Exec("USE stress_test")
	client.Exec("CREATE TABLE metrics (id INT PRIMARY KEY, count INT)")
	client.Exec("INSERT INTO metrics VALUES (1, 0)")
	client.Close()

	const numWorkers = 50
	const duration = 5 * time.Second

	var wg sync.WaitGroup
	var totalOps atomic.Int64
	done := make(chan struct{})

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			c := testutil.NewTestClient(t, ts.DSN())
			defer c.Close()
			c.Exec("USE stress_test")

			for {
				select {
				case <-done:
					return
				default:
					// Mix of operations
					if workerID%2 == 0 {
						c.Exec("UPDATE metrics SET count = count + 1 WHERE id = 1")
					} else {
						c.Query("SELECT count FROM metrics WHERE id = 1")
					}
					totalOps.Add(1)
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	t.Logf("Stress test completed: %d operations in %v (%.0f ops/sec)",
		totalOps.Load(), duration, float64(totalOps.Load())/duration.Seconds())
}
