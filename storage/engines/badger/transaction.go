// Package badger provides the BadgerDB storage engine implementation.
package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/interfaces"
)

// Transaction is a wrapper around a badger.Txn that implements the interfaces.Transaction.
type Transaction struct {
	txn      *badger.Txn
	readOnly bool
}

// newTransaction creates a new transaction.
func newTransaction(db *badger.DB, readOnly bool) *Transaction {
	txn := db.NewTransaction(!readOnly)
	return &Transaction{
		txn:      txn,
		readOnly: readOnly,
	}
}

// Get retrieves a value for a given key.
func (t *Transaction) Get(key []byte) ([]byte, error) {
	item, err := t.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, nil // Return nil, nil for not found, as per many KV store conventions
	}
	if err != nil {
		return nil, err
	}

	var valCopy []byte
	err = item.Value(func(val []byte) error {
		valCopy = append([]byte{}, val...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return valCopy, nil
}

// Set stores a key-value pair.
func (t *Transaction) Set(key, value []byte) error {
	if t.readOnly {
		return badger.ErrUpdateInReadOnlyTxn
	}
	return t.txn.Set(key, value)
}

// Delete removes a key.
func (t *Transaction) Delete(key []byte) error {
	if t.readOnly {
		return badger.ErrUpdateInReadOnlyTxn
	}
	return t.txn.Delete(key)
}

// Iterator returns an iterator for a given key prefix.
func (t *Transaction) Iterator(prefix []byte) (interfaces.Iterator, error) {
	// Note: The transaction must remain open for the iterator to be valid.
	// The user of the iterator is responsible for managing the transaction's lifecycle.
	return newIterator(t.txn, prefix), nil
}

// Commit finalizes the transaction.
func (t *Transaction) Commit() error {
	return t.txn.Commit()
}

// Rollback aborts the transaction.
func (t *Transaction) Rollback() error {
	t.txn.Discard()
	return nil // Discard never returns an error.
}

// IsReadOnly returns true if the transaction is read-only.
func (t *Transaction) IsReadOnly() bool {
	return t.readOnly
}
