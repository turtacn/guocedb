package transaction

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"github.com/turtacn/guocedb/interfaces"
)

// TransactionOptions contains options for creating a new transaction
type TransactionOptions struct {
	IsolationLevel IsolationLevel
	ReadOnly       bool
}

// Transaction implements sql.Transaction and wraps a Badger transaction
type Transaction struct {
	id             string
	startTime      time.Time
	isolationLevel IsolationLevel
	readOnly       bool
	badgerTxn      *badger.Txn
	db             *badger.DB
	committed      bool
	rolledBack     bool
}

// NewTransaction creates a new transaction with the given options
func NewTransaction(db *badger.DB, opts TransactionOptions) *Transaction {
	id := uuid.New().String()
	badgerTxn := db.NewTransaction(!opts.ReadOnly) // update=true for read-write
	return &Transaction{
		id:             id,
		startTime:      time.Now(),
		isolationLevel: opts.IsolationLevel,
		readOnly:       opts.ReadOnly,
		badgerTxn:      badgerTxn,
		db:             db,
	}
}

// String returns the string representation of the transaction
func (t *Transaction) String() string {
	return fmt.Sprintf("Transaction(%s)", t.id)
}

// IsReadOnly returns true if the transaction is read-only
func (t *Transaction) IsReadOnly() bool {
	return t.readOnly
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if t.committed || t.rolledBack {
		return ErrTransactionClosed
	}
	err := t.badgerTxn.Commit()
	if err != nil {
		// Check for BadgerDB conflict errors
		if err == badger.ErrConflict {
			t.rolledBack = true
			return ErrTransactionConflict
		}
		t.rolledBack = true
		return err
	}
	t.committed = true
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if t.committed || t.rolledBack {
		return ErrTransactionClosed
	}
	t.badgerTxn.Discard()
	t.rolledBack = true
	return nil
}

// BadgerTxn returns the underlying Badger transaction for storage layer use
func (t *Transaction) BadgerTxn() *badger.Txn {
	return t.badgerTxn
}

// ID returns the transaction ID
func (t *Transaction) ID() string {
	return t.id
}

// StartTime returns the transaction start time
func (t *Transaction) StartTime() time.Time {
	return t.startTime
}

// IsolationLevel returns the transaction isolation level
func (t *Transaction) IsolationLevel() IsolationLevel {
	return t.isolationLevel
}

// IsClosed returns true if the transaction has been committed or rolled back
func (t *Transaction) IsClosed() bool {
	return t.committed || t.rolledBack
}

// Get retrieves a value for a given key within the transaction
func (t *Transaction) Get(key []byte) ([]byte, error) {
	if t.committed || t.rolledBack {
		return nil, ErrTransactionClosed
	}
	item, err := t.badgerTxn.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return item.ValueCopy(nil)
}

// Set stores a key-value pair within the transaction
func (t *Transaction) Set(key, value []byte) error {
	if t.committed || t.rolledBack {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.badgerTxn.Set(key, value)
}

// Delete removes a key within the transaction
func (t *Transaction) Delete(key []byte) error {
	if t.committed || t.rolledBack {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.badgerTxn.Delete(key)
}

// Iterator returns an iterator for a given key prefix within the transaction
func (t *Transaction) Iterator(prefix []byte) (interfaces.Iterator, error) {
	if t.committed || t.rolledBack {
		return nil, ErrTransactionClosed
	}
	
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	iter := t.badgerTxn.NewIterator(opts)
	
	return &transactionIterator{
		badgerIter: iter,
		prefix:     prefix,
	}, nil
}

// transactionIterator wraps a Badger iterator to implement interfaces.Iterator
type transactionIterator struct {
	badgerIter *badger.Iterator
	prefix     []byte
	valid      bool
	err        error
}

// Next moves the iterator to the next key-value pair
func (it *transactionIterator) Next() bool {
	if it.badgerIter == nil {
		return false
	}
	
	if !it.valid {
		// First call to Next
		it.badgerIter.Rewind()
		it.valid = it.badgerIter.Valid()
	} else {
		// Subsequent calls
		it.badgerIter.Next()
		it.valid = it.badgerIter.Valid()
	}
	
	return it.valid
}

// Key returns the current key
func (it *transactionIterator) Key() []byte {
	if !it.valid || it.badgerIter == nil {
		return nil
	}
	return it.badgerIter.Item().KeyCopy(nil)
}

// Value returns the current value
func (it *transactionIterator) Value() []byte {
	if !it.valid || it.badgerIter == nil {
		return nil
	}
	val, err := it.badgerIter.Item().ValueCopy(nil)
	if err != nil {
		it.err = err
		return nil
	}
	return val
}

// Error returns any error that occurred during iteration
func (it *transactionIterator) Error() error {
	return it.err
}

// Close closes the iterator
func (it *transactionIterator) Close() error {
	if it.badgerIter != nil {
		it.badgerIter.Close()
		it.badgerIter = nil
	}
	return nil
}