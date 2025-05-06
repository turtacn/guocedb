// Package analyzer handles the semantic analysis of SQL queries.
// It resolves table/column names, checks types, and applies rules.
// It utilizes the go-mysql-server SQL analyzer.
//
// analyzer 包处理 SQL 查询的语义分析。
// 它解析表/列名，检查类型，并应用规则。
// 它利用 go-mysql-server 的 SQL 分析器。
package analyzer

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer" // Import GMS analyzer
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	compute_catalog "github.com/turtacn/guocedb/compute/catalog" // Import compute catalog interface
)

// Analyzer is the interface for analyzing SQL query plans.
// Analyzer 是用于分析 SQL 查询计划的接口。
type Analyzer interface {
	// Analyze takes a parsed SQL node and the catalog, and returns an analyzed SQL node.
	// Analysis involves resolving names, type checking, and applying logical rules.
	//
	// Analyze 接收已解析的 SQL 节点和 catalog，并返回已分析的 SQL 节点。
	// 分析涉及名称解析、类型检查和应用逻辑规则。
	Analyze(ctx context.Context, node sql.Node, cat compute_catalog.Catalog) (sql.Node, error)
}

// GMSAnalyzer is an implementation of the Analyzer interface that wraps the go-mysql-server analyzer.
// GMSAnalyzer 是 Analyzer 接口的实现，它包装了 go-mysql-server 分析器。
type GMSAnalyzer struct {
	// gmsAnalyzer is the underlying go-mysql-server analyzer instance.
	// gmsAnalyzer 是底层的 go-mysql-server 分析器实例。
	gmsAnalyzer *analyzer.Analyzer
}

// NewGMSAnalyzer creates a new GMSAnalyzer instance.
// It initializes the go-mysql-server analyzer.
//
// NewGMSAnalyzer 创建一个新的 GMSAnalyzer 实例。
// 它初始化 go-mysql-server 分析器。
func NewGMSAnalyzer() Analyzer {
	log.Info("Initializing GMS SQL analyzer.") // 初始化 GMS SQL 分析器。
	// GMS analyzer requires a catalog and transaction manager during initialization.
	// For now, create a basic analyzer instance. The catalog and transaction manager
	// will be provided during the Analyze call via the sql.Context.
	//
	// GMS 分析器在初始化期间需要 catalog 和事务管理器。
	// 目前，创建一个基本的分析器实例。catalog 和事务管理器将
	// 在 Analyze 调用期间通过 sql.Context 提供。
	// The analyzer constructor takes options, which can include rules, tracing, etc.
	// 分析器构造函数接受选项，可以包括规则、跟踪等。
	// Use default options for now.
	// 目前使用默认选项。
	gmsAnalyzer := analyzer.NewDefault(nil) // NewDefault creates analyzer with default rules

	return &GMSAnalyzer{
		gmsAnalyzer: gmsAnalyzer,
	}
}

// Analyze takes a parsed SQL node and the compute catalog, and returns an analyzed SQL node.
// It delegates the analysis to the go-mysql-server analyzer.
//
// Analyze 接收已解析的 SQL 节点和计算 catalog，并返回已分析的 SQL 节点。
// 它将分析委托给 go-mysql-server 分析器。
func (a *GMSAnalyzer) Analyze(ctx context.Context, node sql.Node, cat compute_catalog.Catalog) (sql.Node, error) {
	log.Debug("Analyzing SQL node: %T", node) // 分析 SQL 节点。

	// GMS analyzer methods take *sql.Context. Need to create or adapt one.
	// This context needs to contain the sql.Catalog and potentially a sql.Transaction.
	//
	// GMS 分析器方法接受 *sql.Context。需要创建一个或适配一个。
	// 此 context 需要包含 sql.Catalog，并可能包含 sql.Transaction。

	// Get the GMS-compatible sql.Catalog from our compute catalog
	// 从我们的计算 catalog 获取 GMS 兼容的 sql.Catalog
	sqlCatalog, err := cat.GetCatalogAsSQL(ctx)
	if err != nil {
		log.Error("Failed to get GMS sql.Catalog from compute catalog for analysis: %v", err) // 从计算 catalog 获取 GMS sql.Catalog 失败。
		return nil, fmt.Errorf("failed to get sql catalog for analysis: %w", err)
	}

	// Create a GMS *sql.Context. It must contain the sql.Catalog.
	// A real implementation would pass the transaction from the session as well.
	//
	// 创建一个 GMS *sql.Context。它必须包含 sql.Catalog。
	// 实际实现也会传递会话中的事务。
	gmsCtx := sql.NewContext(ctx) // Create GMS context from Go context
	gmsCtx.SetCatalog(sqlCatalog) // Set the sql.Catalog


	analyzedNode, err := a.gmsAnalyzer.Analyze(gmsCtx, node, nil) // The third argument is the transaction, pass nil for now
	if err != nil {
		log.Error("Failed to analyze SQL node: %v", err) // 分析 SQL 节点失败。
		// Map GMS analysis errors to our error types if needed.
		// Some GMS errors are specific (e.g., sql.ErrTableNotFound).
		//
		// 如果需要，将 GMS 分析错误映射到我们的错误类型。
		// 一些 GMS 错误是特定的（例如 sql.ErrTableNotFound）。
		// For now, wrap generically.
		// 目前进行通用包装。
		return nil, fmt.Errorf("%w: failed to analyze query: %v", errors.ErrInvalidSQL, err) // Use ErrInvalidSQL for analysis errors
	}

	log.Debug("SQL node analyzed successfully.") // SQL 节点分析成功。
	return analyzedNode, nil // Returns analyzed GMS sql.Node
}