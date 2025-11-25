// Package analyzer provides query analysis capabilities for guocedb.
package analyzer

import (
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/sql/analyzer"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/constants"
)

// Analyzer is a wrapper around the go-mysql-server analyzer.
type Analyzer struct {
	// The underlying GMS analyzer.
	*analyzer.Analyzer
}

// NewAnalyzer creates a new query analyzer.
func NewAnalyzer(c *sql.Catalog) *Analyzer {
	// Create a new GMS analyzer using the provided catalog.
	a := analyzer.NewDefault(c)

	return &Analyzer{
		Analyzer: a,
	}
}

// Analyze performs semantic analysis on the given query plan.
func (a *Analyzer) Analyze(ctx *sql.Context, n sql.Node) (sql.Node, error) {
	// Use the underlying GMS analyzer to process the node.
	analyzedNode, err := a.Analyzer.Analyze(ctx, n)
	if err != nil {
		return nil, errors.Wrapf(err, constants.ErrCodeRuntime, "failed to analyze query")
	}

	// Here we can add custom post-analysis checks, like for permissions.
	if err := a.checkPermissions(ctx, analyzedNode); err != nil {
		return nil, err
	}

	return analyzedNode, nil
}

// checkPermissions is a placeholder for security and access control checks.
func (a *Analyzer) checkPermissions(ctx *sql.Context, n sql.Node) error {
	return nil
}
