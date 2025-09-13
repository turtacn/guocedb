package parser

import (
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/parse"
)

// Parser is a wrapper around the go-mysql-server SQL parser.
// It provides a simple interface for parsing queries and can be extended
// with features like caching.
type Parser struct {
	mu    sync.RWMutex
	cache map[string]sql.Node // A simple, non-evicting cache for parsed queries.
}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{
		cache: make(map[string]sql.Node),
	}
}

// Parse takes a SQL query string and returns a parsed query plan (AST).
// It uses a cache to avoid re-parsing identical query strings.
func (p *Parser) Parse(ctx *sql.Context, query string) (sql.Node, error) {
	p.mu.RLock()
	node, ok := p.cache[query]
	p.mu.RUnlock()

	if ok {
		// Return the cached result. Note: a more robust cache would clone the node.
		return node, nil
	}

	// Parse the query using the standard GMS parser.
	parsed, err := parse.Parse(ctx, query)
	if err != nil {
		return nil, err
	}

	// Cache the result.
	p.mu.Lock()
	p.cache[query] = parsed
	p.mu.Unlock()

	return parsed, nil
}

// TODO: Implement a more sophisticated cache with an eviction policy (e.g., LRU).
// TODO: Add support for custom syntax extensions if needed in the future.
