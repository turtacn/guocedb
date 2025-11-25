package parser

import (
	"context"

	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/sql/parse"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/constants"
)

// Parser parses SQL queries into execution plans.
type Parser interface {
	Parse(ctx context.Context, sqlStr string) (sql.Node, error)
}

// GMSParser is a parser implementation using go-mysql-server.
type GMSParser struct {
}

// NewParser creates a new GMSParser.
func NewParser() *GMSParser {
	return &GMSParser{}
}

// Parse implements the Parser interface.
func (p *GMSParser) Parse(ctx context.Context, sqlStr string) (sql.Node, error) {
	// Ensure we have a *sql.Context
	var sqlCtx *sql.Context
	if c, ok := ctx.(*sql.Context); ok {
		sqlCtx = c
	} else {
		// If not, we create one.
		sqlCtx = sql.NewContext(ctx)
	}

	node, err := parse.Parse(sqlCtx, sqlStr)
	if err != nil {
		return nil, errors.New(constants.ErrCodeSyntax, err.Error())
	}
	return node, nil
}
