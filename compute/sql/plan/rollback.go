package plan

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// RollbackTransaction represents a ROLLBACK statement
type RollbackTransaction struct{}

// NewRollbackTransaction creates a new ROLLBACK transaction node
func NewRollbackTransaction() *RollbackTransaction {
	return &RollbackTransaction{}
}

// Schema implements the Node interface
func (r *RollbackTransaction) Schema() sql.Schema {
	return nil
}

// Children implements the Node interface
func (r *RollbackTransaction) Children() []sql.Node {
	return nil
}

// RowIter implements the Node interface
func (r *RollbackTransaction) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Note: Actual transaction rollback is handled by the handler layer
	// This plan node is mainly for query parsing and validation
	return sql.RowsToRowIter(), nil
}

// String implements the Stringer interface
func (r *RollbackTransaction) String() string {
	return "ROLLBACK"
}

// Resolved implements the Node interface
func (r *RollbackTransaction) Resolved() bool {
	return true
}
