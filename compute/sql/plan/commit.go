package plan

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// CommitTransaction represents a COMMIT statement
type CommitTransaction struct{}

// NewCommitTransaction creates a new COMMIT transaction node
func NewCommitTransaction() *CommitTransaction {
	return &CommitTransaction{}
}

// Schema implements the Node interface
func (c *CommitTransaction) Schema() sql.Schema {
	return nil
}

// Children implements the Node interface
func (c *CommitTransaction) Children() []sql.Node {
	return nil
}

// RowIter implements the Node interface
func (c *CommitTransaction) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Note: Actual transaction commit is handled by the handler layer
	// This plan node is mainly for query parsing and validation
	return sql.RowsToRowIter(), nil
}

// String implements the Stringer interface
func (c *CommitTransaction) String() string {
	return "COMMIT"
}

// Resolved implements the Node interface
func (c *CommitTransaction) Resolved() bool {
	return true
}
