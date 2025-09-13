package vector

import (
	"github.com/dolthub/go-mysql-server/sql"
)

// Batch is a vector of rows, representing a unit of work in the vectorized engine.
type Batch struct {
	Rows   []sql.Row
	Length int
}

// VectorizedOperator is the interface for a vectorized execution operator.
// Each operator processes data in batches.
type VectorizedOperator interface {
	// Next returns the next batch of rows. It returns an empty batch when done.
	Next(ctx *sql.Context) (*Batch, error)
	// Close closes the operator and releases resources.
	Close(ctx *sql.Context) error
}

// IsVectorizationEnabled is a flag to control whether vectorized execution is used.
var IsVectorizationEnabled = false

// TODO: Implement concrete vectorized operators (e.g., VectorizedFilter, VectorizedProject).
// TODO: Implement logic to convert a standard query plan into a vectorized plan.
// TODO: Explore SIMD (Single Instruction, Multiple Data) optimizations within operators.
