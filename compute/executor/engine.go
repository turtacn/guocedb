// Package executor handles the execution of optimized query plans.
// It takes a query plan (sql.Node) and produces a row iterator.
// It utilizes the go-mysql-server execution logic.
//
// executor 包处理优化后的查询计划的执行。
// 它接收查询计划 (sql.Node) 并生成行迭代器。
// 它利用 go-mysql-server 的执行逻辑。
package executor

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	compute_catalog "github.com/turtacn/guocedb/compute/catalog" // Import compute catalog interface
	"github.com/turtacn/guocedb/interfaces" // Import storage transaction interface
)

// Executor is the interface for executing query plans.
// Executor 是用于执行查询计划的接口。
type Executor interface {
	// Execute takes an optimized plan node, catalog, and transaction, and returns a row iterator.
	// The row iterator can then be used to fetch results.
	//
	// Execute 接收优化后的计划节点、catalog 和事务，并返回一个行迭代器。
	// 然后可以使用该行迭代器获取结果。
	Execute(ctx context.Context, plan sql.Node, cat compute_catalog.Catalog, tx interfaces.Transaction) (sql.RowIter, error)
}

// GMSExecutor is an implementation of the Executor interface that leverages go-mysql-server's execution capabilities.
// GMSExecutor 是 Executor 接口的实现，它利用 go-mysql-server 的执行能力。
type GMSExecutor struct {
	// GMS execution is typically driven by calling the RowIter() method on the root plan node.
	// We might not need a dedicated GMS executor struct field, but rather coordinate context and call methods.
	// GMS 执行通常通过在根计划节点上调用 RowIter() 方法来驱动。
	// 我们可能不需要一个专用的 GMS executor 结构体字段，而是协调 context 并调用方法。
}

// NewGMSExecutor creates a new GMSExecutor instance.
// NewGMSExecutor 创建一个新的 GMSExecutor 实例。
func NewGMSExecutor() Executor {
	log.Info("Initializing GMS SQL executor.") // 初始化 GMS SQL 执行器。
	// No specific initialization needed for a stateless executor wrapper.
	// 对于无状态的 executor 包装器，不需要特定的初始化。
	return &GMSExecutor{}
}

// Execute takes an optimized plan node, catalog, and transaction, and returns a row iterator.
// It sets up the GMS context and calls the RowIter() method on the root plan node.
//
// Execute 接收优化后的计划节点、catalog 和事务，并返回一个行迭代器。
// 它设置 GMS context 并在根计划节点上调用 RowIter() 方法。
func (e *GMSExecutor) Execute(ctx context.Context, plan sql.Node, cat compute_catalog.Catalog, tx interfaces.Transaction) (sql.RowIter, error) {
	log.Debug("Executing SQL plan: %T", plan) // 执行 SQL 计划。

	// GMS execution requires a *sql.Context.
	// This context needs the sql.Catalog, the sql.Transaction, and other session-specific details.
	//
	// GMS 执行需要一个 *sql.Context。
	// 此 context 需要 sql.Catalog、sql.Transaction 和其他会话特定详细信息。

	// Get the GMS-compatible sql.Catalog from our compute catalog
	// 从我们的计算 catalog 获取 GMS 兼容的 sql.Catalog
	sqlCatalog, err := cat.GetCatalogAsSQL(ctx)
	if err != nil {
		log.Error("Failed to get GMS sql.Catalog from compute catalog for execution: %v", err) // 从计算 catalog 获取 GMS sql.Catalog 失败。
		return nil, fmt.Errorf("failed to get sql catalog for execution: %w", err)
	}

	// Get the underlying GMS transaction object from our transaction interface
	// 从我们的事务接口获取底层的 GMS 事务对象
	// The sql.Context needs a sql.Transaction.
	// sql.Context 需要一个 sql.Transaction。
	gmsTx, ok := tx.UnderlyingTx().(sql.Transaction)
	if !ok {
		log.Error("Underlying transaction object is not a GMS sql.Transaction: %T", tx.UnderlyingTx()) // 底层事务对象不是 GMS sql.Transaction。
		// Decide if this is a fatal error or if GMS can work without a transaction for some plans.
		// 决定这是否是致命错误，或者 GMS 是否可以在没有事务的情况下执行某些计划。
		// For now, return an error.
		// 目前返回错误。
		return nil, errors.ErrInternal.New("underlying transaction is not GMS compatible")
	}

	// Create a GMS *sql.Context. Set the catalog and transaction.
	// 创建一个 GMS *sql.Context。设置 catalog 和事务。
	gmsCtx := sql.NewContext(ctx) // Create GMS context from Go context
	gmsCtx.SetCatalog(sqlCatalog) // Set the sql.Catalog
	gmsCtx.SetTx(gmsTx)           // Set the sql.Transaction
	// TODO: Set other session variables, client connection, etc.
	// TODO: 设置其他会话变量、客户端连接等。


	// Execute the plan by calling its RowIter method with the GMS context.
	// 通过使用 GMS context 调用根计划节点的 RowIter 方法来执行计划。
	rowIter, err := plan.RowIter(gmsCtx)
	if err != nil {
		log.Error("Failed to execute SQL plan via RowIter(): %v", err) // 通过 RowIter() 执行 SQL 计划失败。
		// Map GMS execution errors to our error types if needed.
		// 确保 GMS 执行错误被捕获并适当处理。
		return nil, fmt.Errorf("%w: failed to execute plan: %v", errors.ErrInternal, err) // Execution errors are often internal or relate to data access
	}

	log.Debug("SQL plan execution started.") // SQL 计划执行开始。
	return rowIter, nil // Returns GMS sql.RowIter
}