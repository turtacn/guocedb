package plan

import "github.com/dolthub/go-mysql-server/sql/plan"

// This package re-exports key plan node types from go-mysql-server for convenience
// and provides a place to define custom plan nodes for guocedb-specific features.

// For convenience, we can alias the most commonly used types.
type Node = plan.Node
type UnaryNode = plan.UnaryNode
type BinaryNode = plan.BinaryNode

// Plan nodes for DQL
type Project = plan.Project
type Filter = plan.Filter
type ResolvedTable = plan.ResolvedTable
type Join = plan.Join

// Plan nodes for DML
type InsertInto = plan.InsertInto
type Update = plan.Update
type DeleteFrom = plan.DeleteFrom

// Plan nodes for DDL
type CreateTable = plan.CreateTable
type DropTable = plan.DropTable
type CreateDatabase = plan.CreateDatabase
type DropDatabase = plan.DropDatabase

// CustomPlanNode is an example of a custom plan node for a feature specific to guocedb.
// For example, this could be a node for a distributed operation.
type CustomPlanNode struct {
	UnaryNode
	// Custom fields would go here
}

// TODO: Define custom plan nodes for distributed query execution.
// e.g., ExchangeNode, RemoteScanNode.
// TODO: Define custom nodes for specific storage engine optimizations.
