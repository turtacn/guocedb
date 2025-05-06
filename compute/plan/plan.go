// Package plan defines representations for logical and physical query execution plans.
// In Guocedb, query plans are primarily represented by the go-mysql-server sql.Node interface
// after parsing, analysis, and optimization.
//
// plan 包定义了逻辑和物理查询执行计划的表示。
// 在 Guocedb 中，查询计划在解析、分析和优化后主要由 go-mysql-server 的 sql.Node 接口表示。
package plan

import (
	"github.com/dolthub/go-mysql-server/sql" // Import GMS sql types
)

// Node is a type alias for the go-mysql-server sql.Node interface.
// It represents a node in the query execution plan tree.
//
// Node 是 go-mysql-server 的 sql.Node 接口的类型别名。
// 它表示查询执行计划树中的一个节点。
// The sql.Node interface provides methods like Schema(), Children(), RowIter(), Resolved(), etc.,
// which define the structure and behavior of the execution plan.
//
// sql.Node 接口提供了 Schema()、Children()、RowIter()、Resolved() 等方法，
// 这些方法定义了执行计划的结构和行为。
type Node = sql.Node

// TODO: If custom plan structures or extensions are needed in the future,
// define them in this package. For now, the GMS sql.Node is sufficient.
//
// TODO: 如果未来需要自定义计划结构或扩展，
// 在此包中定义它们。目前，GMS 的 sql.Node 足够了。