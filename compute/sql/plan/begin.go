package plan

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// BeginTransaction represents a BEGIN statement
type BeginTransaction struct {
	ReadOnly bool
}

// NewBeginTransaction creates a new BEGIN transaction node
func NewBeginTransaction(readOnly bool) *BeginTransaction {
	return &BeginTransaction{ReadOnly: readOnly}
}

// Schema implements the Node interface
func (b *BeginTransaction) Schema() sql.Schema {
	return nil
}

// Children implements the Node interface
func (b *BeginTransaction) Children() []sql.Node {
	return nil
}

// RowIter implements the Node interface
func (b *BeginTransaction) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Check if transaction is already active
	if ctx.GetTransaction() != nil {
		return nil, sql.ErrTransactionAlreadyStarted
	}
	
	// Note: Actual transaction creation is handled by the handler layer
	// This plan node is mainly for query parsing and validation
	return sql.RowsToRowIter(), nil
}

// String implements the Stringer interface
func (b *BeginTransaction) String() string {
	if b.ReadOnly {
		return "BEGIN READ ONLY"
	}
	return "BEGIN"
}

// Resolved implements the Node interface
func (b *BeginTransaction) Resolved() bool {
	return true
}
