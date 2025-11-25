package optimizer

import (
	"context"

	"github.com/turtacn/guocedb/compute/plan"
)

// Optimizer optimizes the plan.
// In GMS, the analyzer performs both analysis and optimization (rule-based).
// This interface might be redundant if we just use Analyzer, but we keep it for architecture separation.
type Optimizer interface {
	Optimize(ctx context.Context, node plan.Node) (plan.Node, error)
}

// GMSOptimizer is a wrapper. Since GMS Analyzer does optimization, this might just be a pass-through
// or handle specific optimization stages if we separated them.
type GMSOptimizer struct {
}

func NewOptimizer() *GMSOptimizer {
	return &GMSOptimizer{}
}

func (o *GMSOptimizer) Optimize(ctx context.Context, node plan.Node) (plan.Node, error) {
	// In GMS, optimization happens during analysis (Analyzer.Analyze).
	// So if the node is already analyzed, it might be already optimized.
	// However, if we want to add extra optimization steps, we can do it here.
	// For now, we return the node as is.
	return node, nil
}
