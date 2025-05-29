// Package badger implements the BadgerDB specific transaction adapter for Guocedb.
// This file is responsible for adapting the transaction interface defined in interfaces/storage.go
// to BadgerDB's transaction mechanism. It manages BadgerDB's read and write transactions,
// ensuring ACID properties at the Badger layer. It relies on the transaction interface
// from interfaces/storage.go and interacts with the BadgerDB client library.
// storage/engines/badger/badger.go will call this file to create and manage transactions.
package badger

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4" // Import BadgerDB client library

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces
)

// ensure that BadgerTransaction implements the interfaces.Transaction interface.
var _ interfaces.Transaction = (*BadgerTransaction)(nil)

// BadgerTransaction implements the interfaces.Transaction interface for BadgerDB.
type BadgerTransaction struct {
	txn      *badger.Txn // The underlying BadgerDB transaction
	id       interfaces.ID
	readOnly bool
	timeout  time.Duration // Timeout for the transaction
	// Optional: context.Context with timeout to cancel long-running transactions
}

// NewBadgerTransaction creates a new BadgerTransaction.
// It wraps a BadgerDB transaction and provides Guocedb's transaction interface.
func NewBadgerTransaction(badgerDB *badger.DB, readOnly bool, id interfaces.ID) *BadgerTransaction {
	txn := badgerDB.NewTransaction(readOnly)
	return &BadgerTransaction{
		txn:      txn,
		id:       id,
		readOnly: readOnly,
		timeout:  0, // No timeout by default
	}
}

// Commit attempts to commit the BadgerDB transaction.
func (bt *BadgerTransaction) Commit() error {
	err := bt.txn.Commit()
	if err != nil {
		return errors.NewGuocedbError(enum.ErrTransaction, errors.CodeTransactionCommitFailed,
			fmt.Sprintf("failed to commit BadgerDB transaction ID %d", bt.id), err)
	}
	return nil
}

// Rollback aborts the BadgerDB transaction.
func (bt *BadgerTransaction) Rollback() error {
	bt.txn.Discard() // BadgerDB's equivalent of rollback for read-write transactions
	return nil
}

// IsReadOnly returns true if the transaction is read-only.
func (bt *BadgerTransaction) IsReadOnly() bool {
	return bt.readOnly
}

// ID returns the unique identifier for this transaction.
func (bt *BadgerTransaction) ID() interfaces.ID {
	return bt.id
}

// SetTimeout sets a timeout for the transaction.
// Note: BadgerDB's native transaction doesn't directly support a timeout parameter
// in the same way as some other databases (e.g., for automatic rollback).
// This timeout would need to be enforced externally (e.g., via a context.Context
// passed through the transaction's lifetime or a background goroutine).
// For now, we store the timeout duration but don't actively enforce it within this method.
// Enforcement would typically happen in the compute layer's transaction manager.
func (bt *BadgerTransaction) SetTimeout(d time.Duration) {
	bt.timeout = d
	// In a real implementation, you might pass a context.Context with timeout
	// to the BadgerDB transaction operations, or manage it externally.
}

// GetBadgerTxn returns the underlying BadgerDB transaction.
// This is an internal helper to allow direct interaction with BadgerDB methods
// within the badger package (e.g., in badger.go for actual K/V operations).
func (bt *BadgerTransaction) GetBadgerTxn() *badger.Txn {
	return bt.txn
}

//Personal.AI order the ending
