package transaction

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openTestBadger creates a temporary Badger database for testing
func openTestBadger(t *testing.T) *badger.DB {
	tmpDir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	
	opts := badger.DefaultOptions(tmpDir).WithLogger(nil)
	db, err := badger.Open(opts)
	require.NoError(t, err)
	
	t.Cleanup(func() {
		db.Close()
	})
	
	return db
}

func TestManager_Begin(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, 1, mgr.ActiveCount())
	
	mgr.Rollback(txn)
	assert.Equal(t, 0, mgr.ActiveCount())
}

func TestManager_BeginWithOptions(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	opts := &TransactionOptions{
		IsolationLevel: LevelSerializable,
		ReadOnly:       true,
	}
	
	txn, err := mgr.Begin(opts)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.True(t, txn.IsReadOnly())
	assert.Equal(t, LevelSerializable, txn.IsolationLevel())
	
	mgr.Rollback(txn)
}

func TestManager_CommitPersists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "badger-commit-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create first database instance
	opts := badger.DefaultOptions(tmpDir).WithLogger(nil)
	db1, err := badger.Open(opts)
	require.NoError(t, err)
	
	mgr1 := NewManagerWithDB(db1)
	
	txn, err := mgr1.Begin(nil)
	require.NoError(t, err)
	
	err = txn.Set([]byte("key"), []byte("value"))
	require.NoError(t, err)
	
	err = mgr1.Commit(txn)
	require.NoError(t, err)
	
	db1.Close()
	
	// Reopen database and verify data persists
	db2, err := badger.Open(opts)
	require.NoError(t, err)
	defer db2.Close()
	
	err = db2.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("key"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value"), val)
		return nil
	})
	assert.NoError(t, err)
}

func TestManager_RollbackDiscards(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = txn.Set([]byte("key"), []byte("value"))
	require.NoError(t, err)
	
	err = mgr.Rollback(txn)
	require.NoError(t, err)
	
	// Verify data doesn't exist
	err = db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("key"))
		assert.Equal(t, badger.ErrKeyNotFound, err)
		return nil
	})
	assert.NoError(t, err)
}

func TestManager_MultipleTransactions(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Start multiple transactions
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	assert.Equal(t, 2, mgr.ActiveCount())
	
	// Commit one, rollback the other
	err = mgr.Commit(txn1)
	assert.NoError(t, err)
	
	err = mgr.Rollback(txn2)
	assert.NoError(t, err)
	
	assert.Equal(t, 0, mgr.ActiveCount())
}

func TestManager_GetTransaction(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	retrieved := mgr.GetTransaction(txn.ID())
	assert.Equal(t, txn, retrieved)
	
	mgr.Rollback(txn)
	
	retrieved = mgr.GetTransaction(txn.ID())
	assert.Nil(t, retrieved)
}

func TestManager_Close(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	// Start multiple transactions
	txn1, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	txn2, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	assert.Equal(t, 2, mgr.ActiveCount())
	
	// Close should rollback all active transactions
	err = mgr.Close()
	assert.NoError(t, err)
	assert.Equal(t, 0, mgr.ActiveCount())
	
	// Transactions should be rolled back
	assert.True(t, txn1.rolledBack)
	assert.True(t, txn2.rolledBack)
}

func TestTransaction_DoubleCommit(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = mgr.Commit(txn)
	assert.NoError(t, err)
	
	// Second commit should fail
	err = txn.Commit()
	assert.Equal(t, ErrTransactionClosed, err)
}

func TestTransaction_DoubleRollback(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = mgr.Rollback(txn)
	assert.NoError(t, err)
	
	// Second rollback should fail
	err = txn.Rollback()
	assert.Equal(t, ErrTransactionClosed, err)
}

func TestTransaction_ReadOnlyViolation(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	opts := &TransactionOptions{ReadOnly: true}
	txn, err := mgr.Begin(opts)
	require.NoError(t, err)
	
	// Write operations should fail on read-only transaction
	err = txn.Set([]byte("key"), []byte("value"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	err = txn.Delete([]byte("key"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	mgr.Rollback(txn)
}

func TestTransaction_OperationsAfterClose(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	
	err = mgr.Commit(txn)
	require.NoError(t, err)
	
	// Operations after commit should fail
	err = txn.Set([]byte("key"), []byte("value"))
	assert.Equal(t, ErrTransactionClosed, err)
	
	_, err = txn.Get([]byte("key"))
	assert.Equal(t, ErrTransactionClosed, err)
	
	err = txn.Delete([]byte("key"))
	assert.Equal(t, ErrTransactionClosed, err)
}

func TestTransaction_String(t *testing.T) {
	db := openTestBadger(t)
	mgr := NewManagerWithDB(db)
	
	txn, err := mgr.Begin(nil)
	require.NoError(t, err)
	defer mgr.Rollback(txn)
	
	str := txn.String()
	assert.Contains(t, str, "Transaction(")
	assert.Contains(t, str, txn.ID())
}