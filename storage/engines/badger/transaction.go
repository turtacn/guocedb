// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"context"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/interfaces" // Import the interfaces package
)

// BadgerTransaction is an implementation of the interfaces.Transaction interface
// wrapping a Badger transaction.
// BadgerTransaction 是 interfaces.Transaction 接口的一个实现，
// 包装了一个 Badger 事务。
type BadgerTransaction struct {
	// txn is the underlying Badger transaction object.
	// txn 是底层的 Badger 事务对象。
	txn *badger.Txn

	// closed indicates if the transaction has been committed or rolled back.
	// closed 表示事务是否已提交或回滚。
	closed bool
}

// NewBadgerTransaction creates a new BadgerTransaction wrapping a Badger transaction.
// NewBadgerTransaction 创建一个新的 BadgerTransaction，包装一个 Badger 事务。
func NewBadgerTransaction(txn *badger.Txn) *BadgerTransaction {
	return &BadgerTransaction{
		txn:    txn,
		closed: false,
	}
}

// Commit commits the underlying Badger transaction.
// Commit 提交底层的 Badger 事务。
func (t *BadgerTransaction) Commit(ctx context.Context) error {
	if t.closed {
		return errors.ErrTransactionCommitFailed.New("transaction already closed") // 事务已关闭。
	}

	log.Debug("Committing Badger transaction") // 提交 Badger 事务。
	err := t.txn.Commit()
	t.closed = true
	if err != nil {
		log.Error("Failed to commit Badger transaction: %v", err) // 提交 Badger 事务失败。
		return fmt.Errorf("%w: %v", errors.ErrTransactionCommitFailed, err)
	}
	log.Debug("Badger transaction committed successfully") // Badger 事务提交成功。
	return nil
}

// Rollback rolls back the underlying Badger transaction.
// Rollback 回滚底层的 Badger 事务。
func (t *BadgerTransaction) Rollback(ctx context.Context) error {
	if t.closed {
		log.Warn("Rollback called on a closed transaction. Ignoring.") // 在已关闭的事务上调用 Rollback。忽略。
		return nil // Or return a specific error? Badger's Discard is idempotent.
	}

	log.Debug("Rolling back Badger transaction") // 回滚 Badger 事务。
	t.txn.Discard() // Badger Discard is safe to call multiple times
	t.closed = true
	log.Debug("Badger transaction rolled back") // Badger 事务已回滚。
	return nil
}

// UnderlyingTx returns the underlying Badger transaction object.
// This is exposed for cases where the storage engine implementation
// needs direct access (e.g., putting/getting within the same transaction).
// UnderlyingTx 返回底层的 Badger 事务对象。
// 这是为了存储引擎实现需要直接访问（例如，在同一事务内进行 put/get 操作）的情况而暴露的。
func (t *BadgerTransaction) UnderlyingTx() interface{} {
	return t.txn
}

// IsClosed checks if the transaction has been committed or rolled back.
// IsClosed 检查事务是否已提交或回滚。
func (t *BadgerTransaction) IsClosed() bool {
	return t.closed
}