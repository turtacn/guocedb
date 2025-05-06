// Package parser handles the parsing of SQL queries.
// It utilizes the go-mysql-server SQL parser.
//
// parser 包处理 SQL 查询的解析。
// 它利用 go-mysql-server 的 SQL 解析器。
package parser

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/parse" // Import GMS parser functions
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
)

// Parser is the interface for parsing SQL queries.
// Parser 是用于解析 SQL 查询的接口。
type Parser interface {
	// Parse takes a SQL query string and returns a slice of SQL nodes representing the parsed statements.
	// Parse 接收 SQL 查询字符串，并返回表示已解析语句的 SQL 节点切片。
	// Context might be used for session-specific settings or cancellation.
	// Context 可用于会话特定设置或取消。
	Parse(ctx context.Context, query string) ([]sql.Node, error)
}

// GMSParser is an implementation of the Parser interface that wraps the go-mysql-server parser.
// GMSParser 是 Parser 接口的实现，它包装了 go-mysql-server 解析器。
type GMSParser struct {
	// The underlying GMS parser instance. GMS parse package provides functions directly,
	// so we might not need a struct field unless wrapping a stateful parser object if available.
	// GMS parse 包直接提供了函数，因此我们可能不需要结构体字段，除非包装一个有状态的解析器对象（如果可用）。
	// For now, just wrap the parse.Parse function.
	// 目前，只包装 parse.Parse 函数。
}

// NewGMSParser creates a new GMSParser instance.
// NewGMSParser 创建一个新的 GMSParser 实例。
func NewGMSParser() Parser {
	log.Info("Initializing GMS SQL parser.") // 初始化 GMS SQL 解析器。
	// GMS parser is stateless for basic parsing, so no specific initialization needed beyond creating the struct.
	// GMS 解析器对于基本解析是无状态的，因此除了创建结构体之外，不需要特定的初始化。
	return &GMSParser{}
}

// Parse takes a SQL query string and returns a slice of SQL nodes.
// It delegates the parsing to the go-mysql-server parse.Parse function.
//
// Parse 接收 SQL 查询字符串，并返回 SQL 节点切片。
// 它将解析委托给 go-mysql-server 的 parse.Parse 函数。
func (p *GMSParser) Parse(ctx context.Context, query string) ([]sql.Node, error) {
	log.Debug("Parsing query: %s", query) // 解析查询。

	// GMS parse.Parse requires a *sql.Context. Need to create or adapt one.
	// GMS parse.Parse 需要一个 *sql.Context。需要创建一个或适配一个。
	// For simplicity, create a new empty context. In a real application,
	// this context should likely come from the session or request context.
	//
	// 为了简化，创建一个新的空 context。在实际应用中，
	// 此 context 可能应来自会话或请求 context。
	gmsCtx := sql.NewEmptyContext() // Or sql.NewContext(ctx) if GMS supports adapting

	nodes, _, err := parse.Parse(gmsCtx, query) // GMS parse.Parse returns nodes, remaining, error
	if err != nil {
		log.Error("Failed to parse query '%s': %v", query, err) // 解析查询失败。
		// Map GMS parser errors to our error types if needed, or return GMS error directly.
		// Check for common parsing errors.
		// 如果需要，将 GMS 解析器错误映射到我们的错误类型，或直接返回 GMS 错误。
		// 检查常见的解析错误。
		// GMS parser errors are usually specific types or wrapped errors.
		// GMS 解析器错误通常是特定的类型或包装的错误。
		// For now, assume GMS errors are informative enough or wrap generically.
		// 目前，假设 GMS 错误足够有信息量，或进行通用包装。
		return nil, fmt.Errorf("%w: failed to parse query: %v", errors.ErrInvalidSQL, err) // Wrap with our error type
	}

	log.Debug("Query parsed successfully.") // 查询解析成功。
	return nodes, nil // Returns slice of GMS sql.Node
}