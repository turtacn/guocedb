// Package plan defines the query plan structure for guocedb.
package plan

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/errors"
)

// Node is an alias for sql.Node for clarity within the project.
// guocedb uses the query plan nodes provided by the go-mysql-server library.
// This file can be extended to include custom plan nodes if needed.
type Node sql.Node

// Serialize converts a query plan to a byte slice.
// This is a placeholder for a real implementation.
func Serialize(n Node) ([]byte, error) {
	// A real implementation would use a library like gob, json, or protobuf
	// to serialize the plan tree. This is complex because the nodes can be
	// of many different types and contain cycles.
	return nil, errors.ErrNotImplemented
}

// Deserialize converts a byte slice back to a query plan.
// This is a placeholder for a real implementation.
func Deserialize(data []byte) (Node, error) {
	return nil, errors.ErrNotImplemented
}

// Decompose breaks down a plan for distributed execution.
// This is a placeholder for a real implementation.
func Decompose(n Node) ([]Node, error) {
	// This would identify parts of the plan that can be pushed down to
	// data nodes (e.g., filters, partial aggregations) and create sub-plans.
	return nil, errors.ErrNotImplemented
}

// IsDDL returns true if the plan node is a DDL statement.
func IsDDL(n Node) bool {
	switch n.(type) {
	case *sql.CreateTable, *sql.DropTable, *sql.AlterTable,
		*sql.CreateDatabase, *sql.DropDatabase,
		*sql.CreateIndex, *sql.DropIndex:
		return true
	default:
		return false
	}
}
