package transaction

import "errors"

var (
	// ErrTransactionClosed is returned when trying to use a closed transaction
	ErrTransactionClosed = errors.New("transaction already closed")
	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrReadOnlyTransaction is returned when trying to write in a read-only transaction
	ErrReadOnlyTransaction = errors.New("cannot write in read-only transaction")
	// ErrNestedTransaction is returned when trying to start a nested transaction
	ErrNestedTransaction = errors.New("nested transactions not supported")
)