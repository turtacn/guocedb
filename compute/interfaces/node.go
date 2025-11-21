// Copyright 2024 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfaces

import "github.com/turtacn/guocedb/compute/types"
// Context is the context for a query execution.
type Context struct {
	// TODO: add session, tracer, pid, etc.
}

// RowIter is an iterator over rows.
type RowIter interface {
	// Next returns the next row.
	Next(ctx *Context) (types.Row, error)
	// Close closes the iterator.
	Close() error
}

// Node is a node in the query execution tree.
type Node interface {
	// Schema returns the schema of the output of the node.
	Schema() types.Schema
	// Children returns the children of the node.
	Children() []Node
	// RowIter returns a row iterator for the node.
	RowIter(ctx *Context) (RowIter, error)
	// Resolved returns true if the node is resolved.
	Resolved() bool
	// TransformUp transforms the node tree from the bottom up.
	TransformUp(f TransformNodeFunc) (Node, error)
}

// Expression is an expression in the query execution tree.
type Expression interface {
	// Eval evaluates the expression.
	Eval(ctx *Context, row types.Row) (interface{}, error)
	// Type returns the type of the expression.
	Type() types.Type
	// IsNullable returns true if the expression can be null.
	IsNullable() bool
	// Resolved returns true if the expression is resolved.
	Resolved() bool
	// TransformUp transforms the expression tree from the bottom up.
	TransformUp(f TransformExprFunc) (Expression, error)
}

// TransformNodeFunc is a function that transforms a node.
type TransformNodeFunc func(Node) (Node, error)

// TransformExprFunc is a function that transforms an expression.
type TransformExprFunc func(Expression) (Expression, error)
