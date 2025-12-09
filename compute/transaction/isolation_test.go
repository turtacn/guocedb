package transaction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotIsolation verifies that transactions see a consistent snapshot
func TestSnapshotIsolation(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup initial data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("key1"), []byte("initial"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Start transaction T1
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	// T1 reads the value
	val1, err := txn1.Get([]byte("key1"))
	require.NoError(t, err)
	assert.Equal(t, []byte("initial"), val1)
	
	// Start and commit transaction T2 that modifies the value
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = txn2.Set([]byte("key1"), []byte("modified"))
	require.NoError(t, err)
	err = mgr.Commit(txn2)
	require.NoError(t, err)
	
	// T1 should still see the initial value (snapshot isolation)
	val1Again, err := txn1.Get([]byte("key1"))
	require.NoError(t, err)
	assert.Equal(t, []byte("initial"), val1Again, "Transaction should see consistent snapshot")
	
	err = mgr.Commit(txn1)
	require.NoError(t, err)
	
	// After T1 commits, a new transaction should see the modified value
	txn3, err := mgr.Begin(nil)
	require.NoError(t, err)
	val3, err := txn3.Get([]byte("key1"))
	require.NoError(t, err)
	assert.Equal(t, []byte("modified"), val3)
	mgr.Rollback(txn3)
}

// TestDirtyReadPrevention verifies uncommitted changes are not visible to other transactions
func TestDirtyReadPrevention(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Transaction T1 writes but doesn't commit
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = txn1.Set([]byte("key1"), []byte("uncommitted"))
	require.NoError(t, err)
	
	// Transaction T2 should not see the uncommitted value
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	_, err = txn2.Get([]byte("key1"))
	assert.Equal(t, ErrKeyNotFound, err, "Should not see uncommitted data (dirty read)")
	
	mgr.Rollback(txn2)
	mgr.Rollback(txn1)
}

// TestRepeatableRead verifies that multiple reads in a transaction return consistent results
func TestRepeatableRead(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup initial data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("counter"), []byte("1"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Start long-running transaction
	txnLong, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	// First read
	val1, err := txnLong.Get([]byte("counter"))
	require.NoError(t, err)
	assert.Equal(t, []byte("1"), val1)
	
	// Another transaction updates the counter
	txnUpdate, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = txnUpdate.Set([]byte("counter"), []byte("2"))
	require.NoError(t, err)
	err = mgr.Commit(txnUpdate)
	require.NoError(t, err)
	
	// Long-running transaction reads again - should see same value
	val2, err := txnLong.Get([]byte("counter"))
	require.NoError(t, err)
	assert.Equal(t, val1, val2, "Repeatable read: should see same value in same transaction")
	
	mgr.Commit(txnLong)
}

// TestReadOnlyTransaction verifies read-only transactions work correctly
func TestReadOnlyTransaction(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Setup data
	setupTxn, err := mgr.Begin(nil)
	require.NoError(t, err)
	err = setupTxn.Set([]byte("readonly_key"), []byte("readonly_value"))
	require.NoError(t, err)
	err = mgr.Commit(setupTxn)
	require.NoError(t, err)
	
	// Read-only transaction
	opts := &TransactionOptions{ReadOnly: true}
	roTxn, err := mgr.Begin(opts)
	require.NoError(t, err)
	
	// Should be able to read
	val, err := roTxn.Get([]byte("readonly_key"))
	require.NoError(t, err)
	assert.Equal(t, []byte("readonly_value"), val)
	
	// Should not be able to write
	err = roTxn.Set([]byte("new_key"), []byte("new_value"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	// Should not be able to delete
	err = roTxn.Delete([]byte("readonly_key"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	mgr.Rollback(roTxn)
}

// TestTransactionTimestamps verifies transaction timing information
func TestTransactionTimestamps(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	before := time.Now()
	time.Sleep(1 * time.Millisecond)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	time.Sleep(1 * time.Millisecond)
	after := time.Now()
	
	startTime := txn.StartTime()
	assert.True(t, startTime.After(before), "Start time should be after 'before' timestamp")
	assert.True(t, startTime.Before(after), "Start time should be before 'after' timestamp")
	
	mgr.Rollback(txn)
}

// TestIsolationLevels verifies different isolation levels can be set
func TestIsolationLevels(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	testCases := []struct {
		name  string
		level IsolationLevel
	}{
		{"ReadUncommitted", LevelReadUncommitted},
		{"ReadCommitted", LevelReadCommitted},
		{"RepeatableRead", LevelRepeatableRead},
		{"Serializable", LevelSerializable},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &TransactionOptions{IsolationLevel: tc.level}
			txn, err := mgr.Begin(opts)
			require.NoError(t, err)
			
			assert.Equal(t, tc.level, txn.IsolationLevel())
			assert.Equal(t, tc.level.String(), txn.IsolationLevel().String())
			
			mgr.Rollback(txn)
		})
	}
}

// TestKeyNotFound verifies proper error handling for missing keys
func TestKeyNotFound(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	defer mgr.Rollback(txn)
	
	_, err = txn.Get([]byte("non_existent_key"))
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestTransactionIsClosed verifies IsClosed method
func TestTransactionIsClosed(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	assert.False(t, txn.IsClosed(), "Transaction should not be closed initially")
	
	err = mgr.Commit(txn)
	require.NoError(t, err)
	
	assert.True(t, txn.IsClosed(), "Transaction should be closed after commit")
}

// TestRollbackAfterCommit verifies rollback on committed transaction
func TestRollbackAfterCommit(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = mgr.Commit(txn)
	require.NoError(t, err)
	
	// Rollback after commit should return error
	err = txn.Rollback()
	assert.Equal(t, ErrTransactionClosed, err)
}

// TestCommitAfterRollback verifies commit on rolled back transaction
func TestCommitAfterRollback(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = mgr.Rollback(txn)
	require.NoError(t, err)
	
	// Commit after rollback should return error
	err = txn.Commit()
	assert.Equal(t, ErrTransactionClosed, err)
}
