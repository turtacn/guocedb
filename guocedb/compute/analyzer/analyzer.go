package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/turtacn/guocedb/compute/catalog"
)

// Analyzer is a wrapper around the go-mysql-server query analyzer.
// It is responsible for semantic analysis, name resolution, and type checking.
type Analyzer struct {
	instance *analyzer.Analyzer
	catalog  catalog.Catalog
}

// NewAnalyzer creates a new Analyzer instance.
// It requires a catalog to resolve database objects.
func NewAnalyzer(catalog catalog.Catalog) *Analyzer {
	// Create a new GMS analyzer with a default set of rules.
	// The DatabaseProvider is the key to resolving tables. The catalog can act as one.
	provider, ok := catalog.(sql.DatabaseProvider)
	if !ok {
		// This is a panic because the catalog *must* provide this interface.
		// Our persistent catalog is designed to do this.
		panic("catalog does not implement sql.DatabaseProvider")
	}

	a := analyzer.NewDefault(provider)

	return &Analyzer{
		instance: a,
		catalog:  catalog,
	}
}

// Analyze takes a parsed SQL node and returns an analyzed and resolved query plan.
func (a *Analyzer) Analyze(ctx *sql.Context, n sql.Node, scope *sql.Scope) (sql.Node, error) {
	return a.instance.Analyze(ctx, n, scope)
}

// TODO: Add custom analysis rules for guocedb-specific features.
// TODO: Integrate permission checks (authz) into the analysis phase.
// For example, a rule could check if the user has SELECT privileges on a table.
