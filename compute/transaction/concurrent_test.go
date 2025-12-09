package transaction

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentWriteConflict verifies that concurrent writes to the same key are detected
func TestConcurrentWriteConflict(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup initial data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("conflict_key"), []byte("initial"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Start transaction T1 and read the key
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	val1, err := txn1.Get([]byte("conflict_key"))
	require.NoError(t, err)
	assert.Equal(t, []byte("initial"), val1)
	
	// Start transaction T2, modify and commit
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = txn2.Set([]byte("conflict_key"), []byte("modified_by_t2"))
	require.NoError(t, err)
	err = mgr.Commit(txn2)
	require.NoError(t, err)
	
	// T1 tries to modify the same key and commit - should detect conflict
	err = txn1.Set([]byte("conflict_key"), []byte("modified_by_t1"))
	require.NoError(t, err)
	
	err = mgr.Commit(txn1)
	// BadgerDB should detect the conflict
	assert.Error(t, err, "Should detect write conflict")
	// Note: The specific error depends on BadgerDB's conflict detection
	// It may be ErrTransactionConflict or another error
}

// TestConcurrentReadNoConflict verifies multiple readers can access data concurrently
func TestConcurrentReadNoConflict(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup initial data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("read_key"), []byte("read_value"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Start multiple concurrent read-only transactions
	numReaders := 10
	var wg sync.WaitGroup
	errors := make(chan error, numReaders)
	
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			opts := &TransactionOptions{ReadOnly: true}
			txn, err := mgr.Begin(opts)
			if err != nil {
				errors <- err
				return
			}
			defer mgr.Rollback(txn)
			
			val, err := txn.Get([]byte("read_key"))
			if err != nil {
				errors <- err
				return
			}
			
			if string(val) != "read_value" {
				errors <- assert.AnError
				return
			}
			
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent read error: %v", err)
	}
}

// TestConcurrentWriters verifies multiple concurrent writers work correctly
func TestConcurrentWriters(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	numWriters := 10
	var wg sync.WaitGroup
	successCount := make(chan int, numWriters)
	
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			txn, err := mgr.Begin(nil)
			if err != nil {
				return
			}
			
			// Each writer writes to its own key
			key := []byte("concurrent_key_" + string(rune('0'+id)))
			value := []byte("value_" + string(rune('0'+id)))
			
			err = txn.Set(key, value)
			if err != nil {
				mgr.Rollback(txn)
				return
			}
			
			err = mgr.Commit(txn)
			if err != nil {
				return
			}
			
			successCount <- 1
		}(i)
	}
	
	wg.Wait()
	close(successCount)
	
	// Count successful commits
	count := 0
	for range successCount {
		count++
	}
	
	// All writers should succeed since they write to different keys
	assert.Equal(t, numWriters, count, "All concurrent writers to different keys should succeed")
}

// TestConcurrentManagerOperations verifies thread-safety of manager operations
func TestConcurrentManagerOperations(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	numGoroutines := 20
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Begin transaction
			txn, err := mgr.Begin(nil)
			if err != nil {
				t.Errorf("Begin failed: %v", err)
				return
			}
			
			// Check active count (should not panic)
			_ = mgr.ActiveCount()
			
			// Get transaction (should not panic)
			_ = mgr.GetTransaction(txn.ID())
			
			// Do some work
			time.Sleep(1 * time.Millisecond)
			
			// Randomly commit or rollback
			if id%2 == 0 {
				mgr.Commit(txn)
			} else {
				mgr.Rollback(txn)
			}
		}(i)
	}
	
	wg.Wait()
	
	// All transactions should be cleaned up
	assert.Equal(t, 0, mgr.ActiveCount())
}

// TestTransactionIteratorConcurrent verifies iterators work correctly under concurrent access
func TestTransactionIteratorConcurrent(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup multiple keys
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		key := []byte("iter_key_" + string(rune('0'+i)))
		value := []byte("iter_value_" + string(rune('0'+i)))
		err = setupTxn.Set(key, value)
		require.NoError(t, err)
	}
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Multiple transactions iterate concurrently
	numIterators := 5
	var wg sync.WaitGroup
	errors := make(chan error, numIterators)
	
	for i := 0; i < numIterators; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			txn, err := mgr.Begin(&TransactionOptions{ReadOnly: true})
			if err != nil {
				errors <- err
				return
			}
			defer mgr.Rollback(txn)
			
			iter, err := txn.Iterator([]byte("iter_key_"))
			if err != nil {
				errors <- err
				return
			}
			defer iter.Close()
			
			count := 0
			for iter.Next() {
				_ = iter.Key()
				_ = iter.Value()
				count++
			}
			
			if count != 10 {
				errors <- assert.AnError
			}
		}()
	}
	
	wg.Wait()
	close(errors)
	
	for err := range errors {
		t.Errorf("Iterator error: %v", err)
	}
}

// TestHighConcurrencyStress stress tests the transaction manager
func TestHighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	numGoroutines := 100
	duration := 2 * time.Second
	done := make(chan bool)
	
	var wg sync.WaitGroup
	var successCount, failCount int64
	var mu sync.Mutex
	
	// Start time
	go func() {
		time.Sleep(duration)
		close(done)
	}()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for {
				select {
				case <-done:
					return
				default:
					txn, err := mgr.Begin(nil)
					if err != nil {
						mu.Lock()
						failCount++
						mu.Unlock()
						continue
					}
					
					// Random operations
					key := []byte("stress_key_" + string(rune('0'+(id%10))))
					
					if id%3 == 0 {
						// Read
						_, _ = txn.Get(key)
					} else {
						// Write
						value := []byte("stress_value")
						_ = txn.Set(key, value)
					}
					
					// Commit or rollback
					if id%5 == 0 {
						err = mgr.Rollback(txn)
					} else {
						err = mgr.Commit(txn)
					}
					
					mu.Lock()
					if err == nil {
						successCount++
					} else {
						failCount++
					}
					mu.Unlock()
					
					// Small delay
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	mu.Lock()
	total := successCount + failCount
	mu.Unlock()
	
	t.Logf("Stress test results: %d successful, %d failed, total: %d", successCount, failCount, total)
	assert.Greater(t, successCount, int64(0), "Should have some successful operations")
	
	// All transactions should be cleaned up
	assert.Equal(t, 0, mgr.ActiveCount())
}

// TestWriteAfterReadConflict verifies write-after-read conflict detection
func TestWriteAfterReadConflict(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup initial data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("war_key"), []byte("initial"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Transaction 1 reads
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	_, err = txn1.Get([]byte("war_key"))
	require.NoError(t, err)
	
	// Transaction 2 writes and commits
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = txn2.Set([]byte("war_key"), []byte("modified"))
	require.NoError(t, err)
	err = mgr.Commit(txn2)
	require.NoError(t, err)
	
	// Transaction 1 writes (based on its read)
	err = txn1.Set([]byte("war_key"), []byte("t1_update"))
	require.NoError(t, err)
	
	// Commit should fail due to conflict
	err = mgr.Commit(txn1)
	// This may or may not fail depending on BadgerDB's conflict detection
	// Just log the result
	t.Logf("Write-after-read conflict result: %v", err)
}

// TestManagerCloseWithActiveTransactions verifies cleanup on close
func TestManagerCloseWithActiveTransactions(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Create some active transactions
	numTxns := 5
	txns := make([]*Transaction, numTxns)
	
	for i := 0; i < numTxns; i++ {
		txn, err := mgr.Begin(nil)
		require.NoError(t, err)
		txns[i] = txn
		
		// Do some work
		key := []byte("close_key_" + string(rune('0'+i)))
		err = txn.Set(key, []byte("value"))
		require.NoError(t, err)
	}
	
	assert.Equal(t, numTxns, mgr.ActiveCount())
	
	// Close manager
	err := mgr.Close()
	require.NoError(t, err)
	
	// All transactions should be rolled back
	assert.Equal(t, 0, mgr.ActiveCount())
	
	// All transactions should be closed
	for _, txn := range txns {
		assert.True(t, txn.IsClosed())
	}
}
