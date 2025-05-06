// Package optimizer handles the optimization of logical query plans.
// It transforms the analyzed plan into a more efficient execution plan.
// It utilizes the go-mysql-server SQL optimizer.
//
// optimizer 包处理逻辑查询计划的优化。
// 它将已分析的计划转换为更高效的执行计划。
// 它利用 go-mysql-server 的 SQL 优化器。
package optimizer

import (
	"context"
	"fmt"
	"github.com/turtacn/guocedb/compute/analyzer"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan" // Import GMS plan nodes
	"github.com/dolthub/go-mysql-server/sql/stats" // GMS Optimizer can use statistics
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	compute_catalog "github.com/turtacn/guocedb/compute/catalog" // Import compute catalog interface
)

// Optimizer is the interface for optimizing analyzed SQL query plans.
// Optimizer 是用于优化已分析 SQL 查询计划的接口。
type Optimizer interface {
	// Optimize takes an analyzed SQL node and the catalog, and returns an optimized SQL node.
	// Optimization applies transformation rules to improve execution efficiency.
	//
	// Optimize 接收已分析的 SQL 节点和 catalog，并返回已优化的 SQL 节点。
	// 优化应用转换规则以提高执行效率。
	Optimize(ctx context.Context, node sql.Node, cat compute_catalog.Catalog) (sql.Node, error)
}

// GMSOptimizer is an implementation of the Optimizer interface that wraps the go-mysql-server optimizer.
// GMSOptimizer 是 Optimizer 接口的实现，它包装了 go-mysql-server 优化器。
type GMSOptimizer struct {
	// gmsOptimizer is the underlying go-mysql-server optimizer instance.
	// gmsOptimizer 是底层的 go-mysql-server 优化器实例。
	// GMS optimizer is typically created from the analyzer.
	// GMS 优化器通常从分析器创建。
	gmsOptimizer *sql.Optimizer
}

// NewGMSOptimizer creates a new GMSOptimizer instance.
// It wraps the optimizer obtained from the go-mysql-server analyzer.
//
// NewGMSOptimizer 创建一个新的 GMSOptimizer 实例。
// 它包装了从 go-mysql-server 分析器获取的优化器。
func NewGMSOptimizer(gmsAnalyzer *analyzer.Analyzer) Optimizer {
	log.Info("Initializing GMS SQL optimizer.") // 初始化 GMS SQL 优化器。
	// GMS analyzer has a property or method to get its optimizer.
	// GMS 分析器有一个属性或方法来获取其优化器。
	// Assuming the analyzer's optimizer is directly accessible.
	// 假设分析器的优化器是直接可访问的。
	gmsOptimizer := gmsAnalyzer.Optimizer // Access the optimizer from the analyzer

	return &GMSOptimizer{
		gmsOptimizer: gmsOptimizer,
	}
}

// Optimize takes an analyzed SQL node and the compute catalog, and returns an optimized SQL node.
// It delegates the optimization to the go-mysql-server optimizer.
//
// Optimize 接收已分析的 SQL 节点和计算 catalog，并返回已优化的 SQL 节点。
// 它将优化委托给 go-mysql-server 优化器。
func (o *GMSOptimizer) Optimize(ctx context.Context, node sql.Node, cat compute_catalog.Catalog) (sql.Node, error) {
	log.Debug("Optimizing SQL node: %T", node) // 优化 SQL 节点。

	// GMS optimizer methods take *sql.Context. Need to create or adapt one.
	// This context needs to contain the sql.Catalog and potentially a sql.Transaction and statistics.
	//
	// GMS 优化器方法接受 *sql.Context。需要创建一个或适配一个。
	// 此 context 需要包含 sql.Catalog，并可能包含 sql.Transaction 和统计信息。

	// Get the GMS-compatible sql.Catalog from our compute catalog
	// 从我们的计算 catalog 获取 GMS 兼容的 sql.Catalog
	sqlCatalog, err := cat.GetCatalogAsSQL(ctx)
	if err != nil {
		log.Error("Failed to get GMS sql.Catalog from compute catalog for optimization: %v", err) // 从计算 catalog 获取 GMS sql.Catalog 失败。
		return nil, fmt.Errorf("failed to get sql catalog for optimization: %w", err)
	}

	// Create a GMS *sql.Context. It must contain the sql.Catalog.
	// A real implementation would pass the transaction and statistics provider.
	//
	// 创建一个 GMS *sql.Context。它必须包含 sql.Catalog。
	// 实际实现会传递事务和统计信息提供者。
	gmsCtx := sql.NewContext(ctx) // Create GMS context from Go context
	gmsCtx.SetCatalog(sqlCatalog) // Set the sql.Catalog
	// TODO: Set transaction and statistics provider if available
	// TODO: 如果可用，设置事务和统计信息提供者
	// gmsCtx.SetTx(transaction)
	// gmsCtx.SetStatisticProvider(statsProvider)


	optimizedNode, err := o.gmsOptimizer.Optimize(gmsCtx, node, nil) // The third argument is the transaction, pass nil for now
	if err != nil {
		log.Error("Failed to optimize SQL node: %v", err) // 优化 SQL 节点失败。
		// Map GMS optimizer errors to our error types if needed.
		// 确保 GMS 优化器错误被捕获并适当处理。
		return nil, fmt.Errorf("%w: failed to optimize query: %v", errors.ErrInvalidSQL, err) // Optimization errors might still relate to invalid plans
	}

	log.Debug("SQL node optimized successfully.") // SQL 节点优化成功。
	return optimizedNode, nil // Returns optimized GMS sql.Node (which is the execution plan)
}