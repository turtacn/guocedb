// Package transaction provides transaction management for guocedb.
package transaction

import (
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/interfaces"
)

// Manager is responsible for creating and managing transactions.
type Manager struct {
	storage           interfaces.Storage
	db                *badger.DB
	activeTxns        map[string]*Transaction
	mu                sync.RWMutex
	defaultIsolation  IsolationLevel
}

// NewManager creates a new transaction manager.
func NewManager(storage interfaces.Storage) *Manager {
	return &Manager{
		storage:          storage,
		activeTxns:       make(map[string]*Transaction),
		defaultIsolation: LevelReadCommitted,
	}
}

// NewManagerWithDB creates a new transaction manager with direct Badger DB access.
func NewManagerWithDB(db *badger.DB) *Manager {
	return &Manager{
		db:               db,
		activeTxns:       make(map[string]*Transaction),
		defaultIsolation: LevelReadCommitted,
	}
}

// NewTransaction begins a new transaction.
func (m *Manager) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	if m.storage != nil {
		// Delegate to storage engine if available
		return m.storage.NewTransaction(ctx, readOnly)
	}
	
	// Use direct Badger DB access
	if m.db == nil {
		return nil, errors.ErrNotImplemented
	}
	
	opts := TransactionOptions{
		IsolationLevel: m.defaultIsolation,
		ReadOnly:       readOnly,
	}
	
	txn := NewTransaction(m.db, opts)
	
	m.mu.Lock()
	m.activeTxns[txn.ID()] = txn
	m.mu.Unlock()
	
	return txn, nil
}

// Begin creates a new read-write transaction with default options.
func (m *Manager) Begin(opts *TransactionOptions) (*Transaction, error) {
	if m.db == nil {
		return nil, errors.ErrNotImplemented
	}
	
	if opts == nil {
		opts = &TransactionOptions{IsolationLevel: m.defaultIsolation}
	}
	
	txn := NewTransaction(m.db, *opts)
	
	m.mu.Lock()
	m.activeTxns[txn.ID()] = txn
	m.mu.Unlock()
	
	return txn, nil
}

// BeginSQL is an alias for NewTransaction, providing a more standard SQL-like API.
func (m *Manager) BeginSQL(ctx *sql.Context) (interfaces.Transaction, error) {
	// Assuming transactions are read-write by default.
	return m.NewTransaction(ctx, false)
}

// Commit commits a transaction.
func (m *Manager) Commit(txn *Transaction) error {
	if txn == nil {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.activeTxns[txn.ID()]; !exists {
		return ErrTransactionNotFound
	}
	
	err := txn.Commit()
	delete(m.activeTxns, txn.ID())
	return err
}

// Rollback rolls back a transaction.
func (m *Manager) Rollback(txn *Transaction) error {
	if txn == nil {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.activeTxns[txn.ID()]; !exists {
		return ErrTransactionNotFound
	}
	
	err := txn.Rollback()
	delete(m.activeTxns, txn.ID())
	return err
}

// GetTransaction retrieves a transaction by ID.
func (m *Manager) GetTransaction(id string) *Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeTxns[id]
}

// ActiveCount returns the number of active transactions.
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.activeTxns)
}

// Close closes the manager and rolls back all active transactions.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Roll back all active transactions
	for _, txn := range m.activeTxns {
		txn.Rollback()
	}
	m.activeTxns = make(map[string]*Transaction)
	return nil
}

// The following are placeholders for more advanced transaction management features.

// CoordinatedCommit performs a two-phase commit for a distributed transaction.
func (m *Manager) CoordinatedCommit() error {
	return errors.ErrNotImplemented
}

// DetectDeadlocks runs a deadlock detection algorithm.
func (m *Manager) DetectDeadlocks() {
	// Placeholder for deadlock detection logic.
}
