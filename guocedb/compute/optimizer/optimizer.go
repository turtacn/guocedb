package optimizer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/optimizer"
)

// Optimizer is a placeholder for the query optimizer.
// It is responsible for transforming a logical query plan into a more
// efficient physical query plan.
type Optimizer struct {
	instance *optimizer.Optimizer
}

// NewOptimizer creates a new Optimizer instance.
func NewOptimizer() *Optimizer {
	// Create a new GMS optimizer with a default set of rules.
	// GMS's default optimizer is a logical optimizer.
	o := optimizer.NewDefault()

	return &Optimizer{
		instance: o,
	}
}

// Optimize takes an analyzed query plan and returns an optimized plan.
func (o *Optimizer) Optimize(ctx *sql.Context, n sql.Node, scope *sql.Scope) (sql.Node, error) {
	return o.instance.Optimize(ctx, n, scope)
}

// TODO: Implement a cost-based optimizer (CBO).
// This would involve:
// - Collecting statistics from tables (e.g., row count, cardinality).
// - Estimating the cost of different plan alternatives (e.g., join orders).
// - Choosing the lowest-cost plan.
// TODO: Add custom optimization rules for guocedb-specific features or storage engines.
