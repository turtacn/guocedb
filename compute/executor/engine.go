// Package executor provides query execution capabilities for guocedb.
package executor

import (
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/optimizer"
	"github.com/turtacn/guocedb/compute/parser"
	"github.com/turtacn/guocedb/compute/auth"
	"github.com/turtacn/guocedb/common/constants"
)

// Engine is the main query execution engine for guocedb.
// It orchestrates the parsing, analysis, optimization, and execution of queries.
type Engine struct {
	parser    parser.Parser
	analyzer  *analyzer.Analyzer
	optimizer optimizer.Optimizer
	Catalog   *sql.Catalog
	Auth      auth.Auth
}

// NewEngine creates a new query execution engine.
func NewEngine(a *analyzer.Analyzer, o optimizer.Optimizer, c *sql.Catalog) *Engine {
	return &Engine{
		parser:    parser.NewParser(),
		analyzer:  a,
		optimizer: o,
		Catalog:   c,
		Auth:      auth.NewNativeSingle("root", "", auth.AllPermissions), // Default auth
	}
}

// Query executes a SQL query and returns the resulting rows and schema.
func (e *Engine) Query(ctx *sql.Context, query string) (sql.Schema, sql.RowIter, error) {
	// 1. Parse the query to get the AST
	parsedNode, err := e.parser.Parse(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	// 2. Analyze the AST to create a logical plan
	analyzedNode, err := e.analyzer.Analyze(ctx, parsedNode)
	if err != nil {
		return nil, nil, err
	}

	// 3. Optimize the logical plan to create a physical plan
	optimizedNode, err := e.optimizer.Optimize(ctx, analyzedNode)
	if err != nil {
		return nil, nil, err
	}

	// 4. Execute the physical plan
	// The GMS plan nodes have an Execute method that returns a RowIter.
	rowIter, err := optimizedNode.RowIter(ctx)
	if err != nil {
		return nil, nil, errors.Wrapf(err, constants.ErrCodeRuntime, "failed to execute query")
	}

	return optimizedNode.Schema(), rowIter, nil
}
