// Package parser provides SQL parsing capabilities for guocedb.
package parser

import (
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/parse"
	"github.com/turtacn/guocedb/common/errors"
)

// Parser is a wrapper around the go-mysql-server parser.
type Parser struct {
	mu    sync.RWMutex
	cache map[string]sql.Node // A simple cache for parsed queries
}

// NewParser creates a new SQL parser.
func NewParser() *Parser {
	return &Parser{
		cache: make(map[string]sql.Node),
	}
}

// Parse parses a SQL query string and returns the abstract syntax tree (AST).
func (p *Parser) Parse(ctx *sql.Context, query string) (sql.Node, error) {
	// Check cache first
	p.mu.RLock()
	node, ok := p.cache[query]
	p.mu.RUnlock()
	if ok {
		// Return a copy to avoid concurrent modification issues with the AST
		return node.Copy(), nil
	}

	// Parse the query
	parsedNode, err := parse.Parse(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrCodeSyntax, "failed to parse query: %s", query)
	}

	// Add to cache
	p.mu.Lock()
	p.cache[query] = parsedNode.Copy()
	p.mu.Unlock()

	return parsedNode, nil
}

// ClearCache clears the parser's internal cache.
func (p *Parser) ClearCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = make(map[string]sql.Node)
}

// AddCustomSyntax would be a placeholder to extend the parser.
func (p *Parser) AddCustomSyntax() error {
	return errors.ErrNotImplemented
}
