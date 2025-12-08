package transaction

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database
func setupTestDB(t *testing.T) (*badger.DB, *Manager) {
	tmpDir, err := ioutil.TempDir("", "integration-test")
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
	
	txnManager := NewManagerWithDB(db)
	return db, txnManager
}

func TestTxn_BasicRollback(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// Begin transaction
	txn, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Set data within transaction
	err = txn.Set([]byte("key1"), []byte("value1"))
	require.NoError(t, err)
	err = txn.Set([]byte("key2"), []byte("value2"))
	require.NoError(t, err)
	
	// Rollback transaction
	err = txnManager.Rollback(txn)
	require.NoError(t, err)
	
	// Verify no data exists
	err = db.View(func(badgerTxn *badger.Txn) error {
		_, err := badgerTxn.Get([]byte("key1"))
		assert.Equal(t, badger.ErrKeyNotFound, err)
		_, err = badgerTxn.Get([]byte("key2"))
		assert.Equal(t, badger.ErrKeyNotFound, err)
		return nil
	})
	assert.NoError(t, err)
}

func TestTxn_BasicCommit(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// Begin transaction
	txn, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Set data within transaction
	err = txn.Set([]byte("key1"), []byte("value1"))
	require.NoError(t, err)
	err = txn.Set([]byte("key2"), []byte("value2"))
	require.NoError(t, err)
	
	// Commit transaction
	err = txnManager.Commit(txn)
	require.NoError(t, err)
	
	// Verify data exists
	err = db.View(func(badgerTxn *badger.Txn) error {
		item, err := badgerTxn.Get([]byte("key1"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), val)
		
		item, err = badgerTxn.Get([]byte("key2"))
		assert.NoError(t, err)
		val, err = item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value2"), val)
		return nil
	})
	assert.NoError(t, err)
}

func TestTxn_UpdateRollback(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// First, insert some data outside of transaction
	err := db.Update(func(badgerTxn *badger.Txn) error {
		return badgerTxn.Set([]byte("key1"), []byte("original_value"))
	})
	require.NoError(t, err)
	
	// Begin transaction for update
	txn, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Update data within transaction
	err = txn.Set([]byte("key1"), []byte("updated_value"))
	require.NoError(t, err)
	
	// Rollback transaction
	err = txnManager.Rollback(txn)
	require.NoError(t, err)
	
	// Verify original data is preserved
	err = db.View(func(badgerTxn *badger.Txn) error {
		item, err := badgerTxn.Get([]byte("key1"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("original_value"), val)
		return nil
	})
	assert.NoError(t, err)
}

func TestTxn_DeleteRollback(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// First, insert some data outside of transaction
	err := db.Update(func(badgerTxn *badger.Txn) error {
		err := badgerTxn.Set([]byte("key1"), []byte("value1"))
		if err != nil {
			return err
		}
		return badgerTxn.Set([]byte("key2"), []byte("value2"))
	})
	require.NoError(t, err)
	
	// Begin transaction for delete
	txn, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Delete data within transaction
	err = txn.Delete([]byte("key1"))
	require.NoError(t, err)
	
	// Rollback transaction
	err = txnManager.Rollback(txn)
	require.NoError(t, err)
	
	// Verify original data is preserved
	err = db.View(func(badgerTxn *badger.Txn) error {
		item, err := badgerTxn.Get([]byte("key1"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), val)
		
		item, err = badgerTxn.Get([]byte("key2"))
		assert.NoError(t, err)
		val, err = item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value2"), val)
		return nil
	})
	assert.NoError(t, err)
}

func TestTxn_MixedOperationsRollback(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// First, insert some initial data
	err := db.Update(func(badgerTxn *badger.Txn) error {
		return badgerTxn.Set([]byte("existing_key"), []byte("existing_value"))
	})
	require.NoError(t, err)
	
	// Begin transaction for mixed operations
	txn, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Insert new data
	err = txn.Set([]byte("new_key"), []byte("new_value"))
	require.NoError(t, err)
	
	// Update existing data
	err = txn.Set([]byte("existing_key"), []byte("updated_value"))
	require.NoError(t, err)
	
	// Delete some data
	err = txn.Delete([]byte("existing_key"))
	require.NoError(t, err)
	
	// Rollback transaction
	err = txnManager.Rollback(txn)
	require.NoError(t, err)
	
	// Verify only original data exists
	err = db.View(func(badgerTxn *badger.Txn) error {
		// Original key should still exist with original value
		item, err := badgerTxn.Get([]byte("existing_key"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("existing_value"), val)
		
		// New key should not exist
		_, err = badgerTxn.Get([]byte("new_key"))
		assert.Equal(t, badger.ErrKeyNotFound, err)
		return nil
	})
	assert.NoError(t, err)
}

func TestTxn_ReadOnlyTransaction(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// First, insert some data
	err := db.Update(func(badgerTxn *badger.Txn) error {
		return badgerTxn.Set([]byte("key1"), []byte("value1"))
	})
	require.NoError(t, err)
	
	// Begin read-only transaction
	opts := &TransactionOptions{ReadOnly: true}
	txn, err := txnManager.Begin(opts)
	require.NoError(t, err)
	
	// Read operations should work
	val, err := txn.Get([]byte("key1"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)
	
	// Write operations should be prevented
	err = txn.Set([]byte("key2"), []byte("value2"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	err = txn.Delete([]byte("key1"))
	assert.Equal(t, ErrReadOnlyTransaction, err)
	
	err = txnManager.Commit(txn)
	require.NoError(t, err)
}

func TestTxn_IsolationBetweenTransactions(t *testing.T) {
	db, txnManager := setupTestDB(t)
	
	// Begin two transactions
	txn1, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	txn2, err := txnManager.Begin(nil)
	require.NoError(t, err)
	
	// Insert data in transaction 1
	err = txn1.Set([]byte("key1"), []byte("value1"))
	require.NoError(t, err)
	
	// Transaction 2 should not see uncommitted data from transaction 1
	_, err = txn2.Get([]byte("key1"))
	assert.Equal(t, badger.ErrKeyNotFound, err)
	
	// Commit transaction 1
	err = txnManager.Commit(txn1)
	require.NoError(t, err)
	
	// Rollback transaction 2
	err = txnManager.Rollback(txn2)
	require.NoError(t, err)
	
	// Verify data from transaction 1 is committed
	err = db.View(func(badgerTxn *badger.Txn) error {
		item, err := badgerTxn.Get([]byte("key1"))
		assert.NoError(t, err)
		val, err := item.ValueCopy(nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), val)
		return nil
	})
	assert.NoError(t, err)
}