package integration

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrent_MultiClient tests multiple clients performing queries simultaneously
func TestConcurrent_MultiClient(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup test data
	setupClient := NewTestClient(t, ts.DSN(""))
	ExecSQLFile(t, setupClient, "testdata/schema.sql")
	ExecSQLFile(t, setupClient, "testdata/seed.sql")
	setupClient.Close()

	numClients := 10
	queriesPerClient := 50
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	// Launch concurrent clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()

			for j := 0; j < queriesPerClient; j++ {
				// Perform various read operations
				queries := []string{
					"SELECT COUNT(*) FROM users",
					"SELECT name FROM users WHERE age > 25",
					"SELECT * FROM products WHERE price < 100",
					"SELECT u.name, COUNT(o.id) FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id",
				}
				
				query := queries[j%len(queries)]
				rows, err := client.db.Query(query)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}
				
				// Consume all rows
				for rows.Next() {
					// Just iterate through results
				}
				rows.Close()
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// Verify most operations succeeded
	totalOps := int64(numClients * queriesPerClient)
	successRate := float64(successCount) / float64(totalOps)
	
	t.Logf("Concurrent operations: %d total, %d success, %d errors (%.2f%% success rate)",
		totalOps, successCount, errorCount, successRate*100)
	
	assert.Greater(t, successRate, 0.95) // At least 95% success rate
}

// TestConcurrent_MixedWorkload tests mixed read/write workload
func TestConcurrent_MixedWorkload(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE items (id INT PRIMARY KEY, data VARCHAR(100), counter INT DEFAULT 0)",
	)
	setupClient.Close()

	var wg sync.WaitGroup
	var insertCount, updateCount, selectCount int64
	var errorCount int64

	// Writers - insert new records
	numWriters := 3
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()
			
			for j := 0; j < 20; j++ {
				id := writerID*1000 + j
				_, err := client.db.Exec("INSERT INTO items (id, data) VALUES (?, ?)", 
					id, fmt.Sprintf("data-%d", id))
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&insertCount, 1)
				}
				time.Sleep(10 * time.Millisecond) // Small delay
			}
		}(i)
	}

	// Updaters - update existing records
	numUpdaters := 2
	for i := 0; i < numUpdaters; i++ {
		wg.Add(1)
		go func(updaterID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()
			
			for j := 0; j < 30; j++ {
				// Update random existing records
				_, err := client.db.Exec("UPDATE items SET counter = counter + 1 WHERE id < 1000")
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&updateCount, 1)
				}
				time.Sleep(15 * time.Millisecond)
			}
		}(i)
	}

	// Readers - perform various queries
	numReaders := 5
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()
			
			for j := 0; j < 50; j++ {
				queries := []string{
					"SELECT COUNT(*) FROM items",
					"SELECT * FROM items WHERE id < 100",
					"SELECT AVG(counter) FROM items",
					"SELECT data FROM items ORDER BY id LIMIT 10",
				}
				
				query := queries[j%len(queries)]
				rows, err := client.db.Query(query)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}
				
				// Consume results
				for rows.Next() {
					// Just iterate
				}
				rows.Close()
				atomic.AddInt64(&selectCount, 1)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	t.Logf("Mixed workload results: %d inserts, %d updates, %d selects, %d errors",
		insertCount, updateCount, selectCount, errorCount)
	
	// Verify operations succeeded
	assert.Greater(t, insertCount, int64(0))
	assert.Greater(t, selectCount, int64(0))
	
	// Error rate should be low
	totalOps := insertCount + updateCount + selectCount + errorCount
	if totalOps > 0 {
		errorRate := float64(errorCount) / float64(totalOps)
		assert.Less(t, errorRate, 0.1) // Less than 10% error rate
	}
}

// TestConcurrent_ConnectionPool tests connection pool behavior under load
func TestConcurrent_ConnectionPool(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE load_test (id INT PRIMARY KEY, data VARCHAR(100))",
	)
	setupClient.Close()

	numClients := 20 // More clients than typical connection pool size
	operationsPerClient := 25
	var wg sync.WaitGroup
	var successCount int64

	// Launch many concurrent clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()

			for j := 0; j < operationsPerClient; j++ {
				// Mix of operations
				if j%3 == 0 {
					// Insert
					id := clientID*1000 + j
					_, err := client.db.Exec("INSERT INTO load_test (id, data) VALUES (?, ?)", 
						id, fmt.Sprintf("client-%d-op-%d", clientID, j))
					if err == nil {
						atomic.AddInt64(&successCount, 1)
					}
				} else {
					// Select
					rows, err := client.db.Query("SELECT COUNT(*) FROM load_test")
					if err == nil {
						for rows.Next() {
							// Consume result
						}
						rows.Close()
						atomic.AddInt64(&successCount, 1)
					}
				}
				
				// Small delay to simulate real work
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify high success rate despite connection pressure
	expectedOps := int64(numClients * operationsPerClient)
	successRate := float64(successCount) / float64(expectedOps)
	
	t.Logf("Connection pool test: %d/%d operations succeeded (%.2f%%)",
		successCount, expectedOps, successRate*100)
	
	assert.Greater(t, successRate, 0.90) // At least 90% success rate
}

// TestConcurrent_LongTransaction tests impact of long transactions on concurrent operations
func TestConcurrent_LongTransaction(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE txn_test (id INT PRIMARY KEY, val INT)",
		"INSERT INTO txn_test (id, val) VALUES (1, 100), (2, 200), (3, 300)",
	)
	setupClient.Close()

	var wg sync.WaitGroup
	var shortOpsCount int64

	// Start a long-running transaction
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		longClient := NewTestClient(t, ts.DSN("testdb"))
		defer longClient.Close()
		
		longClient.Exec("BEGIN")
		longClient.Exec("UPDATE txn_test SET val = val + 1000 WHERE id = 1")
		
		// Hold transaction for a while
		time.Sleep(500 * time.Millisecond)
		
		longClient.Exec("COMMIT")
	}()

	// Start multiple short operations
	numShortClients := 5
	for i := 0; i < numShortClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()
			
			for j := 0; j < 10; j++ {
				// Read operations that shouldn't be blocked by the long transaction
				rows, err := client.db.Query("SELECT val FROM txn_test WHERE id IN (2, 3)")
				if err == nil {
					for rows.Next() {
						// Consume results
					}
					rows.Close()
					atomic.AddInt64(&shortOpsCount, 1)
				}
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify short operations weren't significantly impacted
	expectedShortOps := int64(numShortClients * 10)
	shortOpsRate := float64(shortOpsCount) / float64(expectedShortOps)
	
	t.Logf("Long transaction test: %d/%d short operations succeeded (%.2f%%)",
		shortOpsCount, expectedShortOps, shortOpsRate*100)
	
	assert.Greater(t, shortOpsRate, 0.80) // At least 80% of short ops should succeed
}

// TestConcurrent_HighFrequencyInserts tests high-frequency insert operations
func TestConcurrent_HighFrequencyInserts(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE high_freq (id INT PRIMARY KEY, timestamp BIGINT, data VARCHAR(50))",
	)
	setupClient.Close()

	numClients := 8
	insertsPerClient := 100
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	startTime := time.Now().UnixNano()

	// Launch high-frequency inserters
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()

			for j := 0; j < insertsPerClient; j++ {
				id := clientID*10000 + j
				timestamp := time.Now().UnixNano()
				data := fmt.Sprintf("client-%d-insert-%d", clientID, j)
				
				_, err := client.db.Exec("INSERT INTO high_freq (id, timestamp, data) VALUES (?, ?, ?)",
					id, timestamp, data)
				
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
				
				// Very small delay for high frequency
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	
	duration := time.Since(time.Unix(0, startTime))
	totalOps := successCount + errorCount
	opsPerSecond := float64(totalOps) / duration.Seconds()

	t.Logf("High frequency test: %d inserts in %v (%.2f ops/sec), %d errors",
		successCount, duration, opsPerSecond, errorCount)

	// Verify high success rate and reasonable performance
	successRate := float64(successCount) / float64(totalOps)
	assert.Greater(t, successRate, 0.95) // At least 95% success rate
	assert.Greater(t, opsPerSecond, 100.0) // At least 100 ops/second
}

// TestConcurrent_StressTest performs a comprehensive stress test
func TestConcurrent_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := NewTestServer(t)
	defer ts.Close()

	// Setup
	setupClient := NewTestClient(t, ts.DSN(""))
	setupClient.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE stress_test (id INT PRIMARY KEY, category VARCHAR(20), value INT, data TEXT)",
	)
	setupClient.Close()

	var wg sync.WaitGroup
	var totalOps, successOps, errorOps int64
	testDuration := 10 * time.Second
	stopChan := make(chan struct{})

	// Start timer to stop test
	go func() {
		time.Sleep(testDuration)
		close(stopChan)
	}()

	// Mixed workload workers
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			client := NewTestClient(t, ts.DSN("testdb"))
			defer client.Close()
			
			opCounter := 0
			for {
				select {
				case <-stopChan:
					return
				default:
					atomic.AddInt64(&totalOps, 1)
					opCounter++
					
					// Mix of operations
					switch opCounter % 10 {
					case 0, 1, 2: // 30% inserts
						id := workerID*100000 + opCounter
						category := fmt.Sprintf("cat-%d", opCounter%5)
						value := opCounter * 10
						data := fmt.Sprintf("worker-%d-data-%d", workerID, opCounter)
						
						_, err := client.db.Exec("INSERT INTO stress_test (id, category, value, data) VALUES (?, ?, ?, ?)",
							id, category, value, data)
						if err != nil {
							atomic.AddInt64(&errorOps, 1)
						} else {
							atomic.AddInt64(&successOps, 1)
						}
						
					case 3, 4: // 20% updates
						_, err := client.db.Exec("UPDATE stress_test SET value = value + 1 WHERE id < ?", 
							workerID*100000+opCounter-100)
						if err != nil {
							atomic.AddInt64(&errorOps, 1)
						} else {
							atomic.AddInt64(&successOps, 1)
						}
						
					default: // 50% selects
						queries := []string{
							"SELECT COUNT(*) FROM stress_test",
							"SELECT category, AVG(value) FROM stress_test GROUP BY category",
							"SELECT * FROM stress_test WHERE value > ? LIMIT 10",
							"SELECT id, data FROM stress_test ORDER BY id DESC LIMIT 5",
						}
						
						query := queries[opCounter%len(queries)]
						var rows *sql.Rows
						var err error
						
						if query == "SELECT * FROM stress_test WHERE value > ? LIMIT 10" {
							rows, err = client.db.Query(query, opCounter*5)
						} else {
							rows, err = client.db.Query(query)
						}
						
						if err != nil {
							atomic.AddInt64(&errorOps, 1)
						} else {
							// Consume results
							for rows.Next() {
								// Just iterate
							}
							rows.Close()
							atomic.AddInt64(&successOps, 1)
						}
					}
					
					// Small delay
					time.Sleep(2 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate results
	successRate := float64(successOps) / float64(totalOps)
	opsPerSecond := float64(totalOps) / testDuration.Seconds()

	t.Logf("Stress test results over %v:", testDuration)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Successful: %d (%.2f%%)", successOps, successRate*100)
	t.Logf("  Errors: %d", errorOps)
	t.Logf("  Throughput: %.2f ops/second", opsPerSecond)

	// Verify acceptable performance under stress
	assert.Greater(t, successRate, 0.85) // At least 85% success rate under stress
	assert.Greater(t, opsPerSecond, 50.0) // At least 50 ops/second under stress
	assert.Greater(t, totalOps, int64(100)) // Ensure we actually did some work
}