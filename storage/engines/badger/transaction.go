package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid" // For generating transaction IDs

	"github.com/turtacn/guocedb/common/errors" // Core sortable encoding functions
	"github.com/turtacn/guocedb/interfaces"    // For Schema, ColumnDefinition, IndexDefinition
)

// Compile-time check to ensure badgerTransaction implements interfaces.Transaction.
var _ interfaces.Transaction = (*badgerTransaction)(nil)

// badgerTransaction wraps a Badger transaction (*badger.Txn) to implement
// the interfaces.Transaction interface required by the Storage Abstraction Layer.
type badgerTransaction struct {
	// The underlying Badger read-write or read-only transaction.
	txn *badger.Txn
	// A unique identifier generated for this transaction wrapper.
	id string
	// Indicates if the underlying Badger transaction was created as read-only.
	readOnly bool
	// Reference to the engine, potentially useful for accessing configuration or metrics.
	// engine *badgerEngine // Uncomment if needed later
}

// newBadgerTransaction creates a new wrapper instance for a given Badger transaction.
// It assigns a unique ID and stores the read-only status.
func newBadgerTransaction(txn *badger.Txn, readOnly bool /*, engine *badgerEngine*/) *badgerTransaction {
	return &badgerTransaction{
		txn:      txn,
		id:       uuid.NewString(), // Generate a unique ID for this SAL transaction
		readOnly: readOnly,
		// engine: engine, // Uncomment if needed later
	}
}

// ID returns the unique identifier assigned to this transaction wrapper.
func (bt *badgerTransaction) ID() string {
	return bt.id
}

// IsReadOnly returns true if the underlying Badger transaction is read-only.
func (bt *badgerTransaction) IsReadOnly() bool {
	return bt.readOnly
}

// Commit attempts to commit the underlying Badger transaction.
// It wraps any Badger error into a GuoceDB storage error.
// It respects context cancellation before attempting the commit.
func (bt *badgerTransaction) Commit(ctx context.Context) error {
	// Check for context cancellation before attempting the potentially blocking commit.
	select {
	case <-ctx.Done():
		// Context was cancelled, ensure the transaction is discarded.
		bt.txn.Discard() // Discard best-effort on cancellation
		return errors.Wrapf(ctx.Err(), errors.ErrCodeTransactionAborted, "transaction %s commit aborted due to context cancellation", bt.id)
	default:
		// Context is still valid, proceed with commit.
	}

	// Attempt to commit the Badger transaction.
	err := bt.txn.Commit()
	if err != nil {
		// Don't call Discard here, Commit failed, transaction is already aborted by Badger.
		// Wrap the Badger error. Common errors include ErrConflict.
		code := errors.ErrCodeTransactionCommitFailed
		if err == badger.ErrConflict {
			code = errors.ErrCodeTransactionConflict
		}
		return errors.Wrapf(err, code, "failed to commit badger transaction %s", bt.id)
	}

	// Commit successful. The underlying bt.txn is no longer valid after Commit().
	return nil
}

// Rollback aborts the underlying Badger transaction using Discard().
// Badger's Discard() does not return an error.
// It respects context cancellation minimally (checks before discard).
func (bt *badgerTransaction) Rollback(ctx context.Context) error {
	// Check for context cancellation, although Discard is usually very fast.
	select {
	case <-ctx.Done():
		// Log maybe? Discard should still be called for cleanup.
		// Discarding anyway as it's a cleanup operation.
	default:
		// Proceed
	}

	// Discard the Badger transaction. This cleans up resources.
	// It's safe to call Discard() multiple times or after Commit()/Discard().
	bt.txn.Discard()

	// The underlying bt.txn is no longer valid after Discard().
	// Our interface expects an error return, but Badger's Discard doesn't provide one.
	return nil
}

// --- Helper for accessing the underlying transaction ---

// badgerTxn returns the underlying *badger.Txn.
// This is an internal helper used by other parts of the badger engine adapter
// (like badgerTable, badgerIndex) to perform operations within the transaction context.
// It assumes the caller knows the transaction is valid (not yet committed or rolled back).
func (bt *badgerTransaction) badgerTxn() *badger.Txn {
	return bt.txn
}
