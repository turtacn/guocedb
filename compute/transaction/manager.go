// Package transaction manages database transactions for the compute layer.
// It provides an interface for starting and managing transactions, delegating
// to the underlying storage engine for transaction capabilities.
//
// transaction 包为计算层管理数据库事务。
// 它提供一个接口，用于开启和管理事务，并将事务能力
// 委托给底层的存储引擎。
package transaction

import (
	"context"
	"fmt"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/interfaces" // Import storage interfaces
)

// TransactionManager is the interface for managing compute-layer transactions.
// It orchestrates transaction creation and potentially tracks active transactions.
//
// TransactionManager 是管理计算层事务的接口。
// 它协调事务创建，并可能跟踪活跃的事务。
type TransactionManager interface {
	// Begin starts a new database transaction.
	// It returns a Transaction object that can be used for commit or rollback.
	//
	// Begin 开启一个新的数据库事务。
	// 它返回一个可用于提交或回滚的 Transaction 对象。
	Begin(ctx context.Context) (interfaces.Transaction, error)

	// // Future methods could include GetTransaction(id) for tracking, RollbackAllActive(), etc.
	// // 未来方法可能包括 GetTransaction(id) 用于跟踪、RollbackAllActive() 等。
}

// ComputeTransactionManager is an implementation of the TransactionManager interface.
// It delegates transaction creation to the underlying storage engine.
//
// ComputeTransactionManager 是 TransactionManager 接口的实现。
// 它将事务创建委托给底层的存储引擎。
type ComputeTransactionManager struct {
	// storageEngine is the underlying storage engine responsible for providing transaction capabilities.
	// storageEngine 是负责提供事务能力的底层存储引擎。
	storageEngine interfaces.StorageEngine
}

// NewComputeTransactionManager creates a new ComputeTransactionManager instance.
// It requires a reference to the storage engine.
//
// NewComputeTransactionManager 创建一个新的 ComputeTransactionManager 实例。
// 它需要对存储引擎的引用。
func NewComputeTransactionManager(storageEngine interfaces.StorageEngine) TransactionManager {
	log.Info("Initializing compute transaction manager.") // 初始化计算事务管理器。
	return &ComputeTransactionManager{
		storageEngine: storageEngine,
	}
}

// Begin starts a new database transaction by delegating to the storage engine.
// Begin 通过委托给存储引擎开启一个新的数据库事务。
func (m *ComputeTransactionManager) Begin(ctx context.Context) (interfaces.Transaction, error) {
	log.Debug("ComputeTransactionManager Begin called.") // 调用 ComputeTransactionManager Begin。

	if m.storageEngine == nil {
		log.Error("Storage engine is not set in ComputeTransactionManager.") // ComputeTransactionManager 中未设置存储引擎。
		return nil, errors.ErrInternal.New("transaction manager not initialized with storage engine")
	}

	// Delegate the BeginTransaction call to the underlying storage engine.
	// 委托 BeginTransaction 调用给底层的存储引擎。
	tx, err := m.storageEngine.BeginTransaction(ctx)
	if err != nil {
		log.Error("Failed to begin transaction via storage engine: %v", err) // 通过存储引擎开启事务失败。
		// Map storage errors to our error types if needed.
		// For now, assume storage engine returns appropriate errors or wrap generically.
		//
		// 如果需要，将存储错误映射到我们的错误类型。
		// 目前，假设存储引擎返回适当的错误或进行通用包装。
		return nil, fmt.Errorf("%w: failed to begin transaction: %v", errors.ErrInternal, err) // Wrap with ErrInternal or a more specific transaction error
	}

	log.Debug("Transaction started successfully via storage engine.") // 通过存储引擎开启事务成功。
	return tx, nil // Returns interfaces.Transaction from the storage engine
}