package executor

import (
	"context"
	"io"
	"testing"

	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/optimizer"
	"github.com/turtacn/guocedb/compute/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Execute_Simple(t *testing.T) {
	// Create components
	c := sql.NewCatalog()
	// Register a dummy database to avoid "database not found" error if analyzer checks for it
	db := mem.NewDatabase("test_db")
	c.AddDatabase(db)

	// Set current database in catalog (since GMS v0.5.1 seems to use Catalog for this, or at least for default)
	c.SetCurrentDatabase("test_db")

	a := analyzer.NewAnalyzer(c)
	o := optimizer.NewOptimizer()

	e := NewEngine(a, o, c)

	ctx := context.Background()
	// Need a *sql.Context
	sqlCtx := sql.NewContext(ctx)

	sqlStr := "SELECT 1"
	schema, iter, err := e.Query(sqlCtx, sqlStr)
	require.NoError(t, err)
	require.NotNil(t, iter)
	require.NotNil(t, schema)

	// Consume iterator
	rows := []sql.Row{}
	for {
		row, err := iter.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	assert.Equal(t, 1, len(rows))
	// Result of SELECT 1 is likely int8 or int64 depending on parser/analyzer type inference
	// In GMS, literals are usually parsed with specific types.
	// 1 is parsed as IntVal, which `convertVal` in parser converts to int64.
	assert.Equal(t, int64(1), rows[0][0])
}
