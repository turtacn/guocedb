package badger

import (
	"context"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/interfaces"
)

// badgerTxn implements the interfaces.Transaction for the BadgerDB engine.
// It wraps a badger.Txn to provide transactional guarantees.
type badgerTxn struct {
	txn *badger.Txn
}

// newBadgerTxn creates a new transaction wrapper.
func newBadgerTxn(txn *badger.Txn) interfaces.Transaction {
	return &badgerTxn{txn: txn}
}

// Get retrieves the value for a given key within the transaction.
func (t *badgerTxn) Get(ctx context.Context, key []byte) ([]byte, error) {
	item, err := t.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, nil // Return nil for not found
	}
	if err != nil {
		return nil, err
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Set sets the value for a given key within the transaction.
func (t *badgerTxn) Set(ctx context.Context, key, value []byte) error {
	return t.txn.Set(key, value)
}

// Delete removes a key within the transaction.
func (t *badgerTxn) Delete(ctx context.TBD, key []byte) error {
	return t.txn.Delete(key)
}

// Commit commits the transaction, making all changes visible and durable.
func (t *badgerTxn) Commit(ctx context.Context) error {
	return t.txn.Commit()
}

// Rollback discards all changes made in the transaction.
func (t *badgerTxn) Rollback(ctx context.Context) error {
	t.txn.Discard()
	return nil // Badger's Discard() doesn't return an error
}

// NewIterator creates a new iterator for the transaction.
func (t *badgerTxn) NewIterator(ctx context.Context, prefix []byte) (interfaces.Iterator, error) {
	return newBadgerIterator(t.txn, prefix), nil
}
