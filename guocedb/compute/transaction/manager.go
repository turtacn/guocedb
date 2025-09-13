package transaction

import (
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/interfaces"
)

// Manager is the interface for the transaction manager.
// It is responsible for creating and managing the lifecycle of transactions.
type Manager interface {
	// Begin starts a new transaction for a given session.
	Begin(ctx *sql.Context) (*sql.Transaction, error)
	// Commit commits the transaction for the given session.
	Commit(ctx *sql.Context, tx *sql.Transaction) error
	// Rollback rolls back the transaction for the given session.
	Rollback(ctx *sql.Context, tx *sql.Transaction) error
}

// defaultManager is a standard implementation of the transaction manager.
// It coordinates with the underlying storage engine to manage transactions.
type defaultManager struct {
	storage interfaces.Storage
	// TODO: Add tracking for active transactions for deadlock detection etc.
	mu sync.Mutex
}

// NewManager creates a new transaction manager.
func NewManager(storage interfaces.Storage) Manager {
	return &defaultManager{
		storage: storage,
	}
}

// Begin starts a new transaction.
func (m *defaultManager) Begin(ctx *sql.Context) (*sql.Transaction, error) {
	// The `readOnly` flag should be determined from the query context if possible.
	// Defaulting to read-write for now.
	txn, err := m.storage.NewTransaction(ctx, false)
	if err != nil {
		return nil, err
	}

	// The sql.Transaction object wraps our native transaction interface.
	return &sql.Transaction{
		Txn: txn,
	}, nil
}

// Commit commits the transaction.
func (m *defaultManager) Commit(ctx *sql.Context, tx *sql.Transaction) error {
	if tx == nil || tx.Txn == nil {
		return nil // Nothing to commit
	}
	nativeTx, ok := tx.Txn.(interfaces.Transaction)
	if !ok {
		return sql.ErrInvalidTxForEngine.New()
	}
	return nativeTx.Commit(ctx)
}

// Rollback rolls back the transaction.
func (m *defaultManager) Rollback(ctx *sql.Context, tx *sql.Transaction) error {
	if tx == nil || tx.Txn == nil {
		return nil // Nothing to roll back
	}
	nativeTx, ok := tx.Txn.(interfaces.Transaction)
	if !ok {
		return sql.ErrInvalidTxForEngine.New()
	}
	return nativeTx.Rollback(ctx)
}

// TODO: Implement distributed transaction coordination (2PC).
// TODO: Implement deadlock detection and recovery mechanisms.
