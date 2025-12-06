package parse

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/sql/plan"
	"github.com/turtacn/guocedb/compute/sql/expression"
)

func TestParseSetOp(t *testing.T) {
	testCases := []struct {
		name         string
		query        string
		expectedNode sql.Node
	}{
		{
			"Union",
			"SELECT 1 UNION SELECT 2",
			plan.NewUnion(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				true,
			),
		},
		{
			"Union All",
			"SELECT 1 UNION ALL SELECT 2",
			plan.NewUnion(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				false,
			),
		},
		{
			"Intersect",
			"SELECT 1 INTERSECT SELECT 2",
			plan.NewIntersect(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				true,
			),
		},
		{
			"Intersect All",
			"SELECT 1 INTERSECT ALL SELECT 2",
			plan.NewIntersect(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				false,
			),
		},
		{
			"Except",
			"SELECT 1 EXCEPT SELECT 2",
			plan.NewExcept(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				true,
			),
		},
        {
			"Except All",
			"SELECT 1 EXCEPT ALL SELECT 2",
			plan.NewExcept(
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(1), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				plan.NewProject(
					[]sql.Expression{expression.NewLiteral(int64(2), sql.Int64)},
					plan.NewUnresolvedTable("dual", ""),
				),
				false,
			),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctx := sql.NewEmptyContext()
			node, err := Parse(ctx, tt.query)
			require.NoError(err)
			require.Equal(tt.expectedNode, node)
		})
	}
}
