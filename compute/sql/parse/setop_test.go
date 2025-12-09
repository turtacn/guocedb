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

func TestParseNestedSetOp(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	// Test (A UNION B) EXCEPT C
	query := "(SELECT 1 UNION SELECT 2) EXCEPT SELECT 3"
	node, err := Parse(ctx, query)
	require.NoError(err)

	// The node should be an Except
	exceptNode, ok := node.(*plan.Except)
	require.True(ok, "Expected *plan.Except, got %T", node)
	require.True(exceptNode.Distinct)

	// Left side should be a Union
	unionNode, ok := exceptNode.Left.(*plan.Union)
	require.True(ok, "Expected left child to be *plan.Union, got %T", exceptNode.Left)
	require.True(unionNode.Distinct)

	// Right side should be a Project
	_, ok = exceptNode.Right.(*plan.Project)
	require.True(ok, "Expected right child to be *plan.Project, got %T", exceptNode.Right)
}

func TestParseSetOpWithOrderBy(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	query := "SELECT 1 UNION SELECT 2 ORDER BY 1"
	node, err := Parse(ctx, query)
	require.NoError(err)

	// The node should be a Sort wrapping a Union
	sortNode, ok := node.(*plan.Sort)
	require.True(ok, "Expected *plan.Sort, got %T", node)

	// The child should be a Union
	unionNode, ok := sortNode.Child.(*plan.Union)
	require.True(ok, "Expected child to be *plan.Union, got %T", sortNode.Child)
	require.True(unionNode.Distinct)
}

func TestParseSetOpWithLimit(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	query := "SELECT 1 UNION SELECT 2 LIMIT 10"
	node, err := Parse(ctx, query)
	require.NoError(err)

	// The node should be a Limit wrapping a Union
	limitNode, ok := node.(*plan.Limit)
	require.True(ok, "Expected *plan.Limit, got %T", node)

	// The child should be a Union
	unionNode, ok := limitNode.Child.(*plan.Union)
	require.True(ok, "Expected child to be *plan.Union, got %T", limitNode.Child)
	require.True(unionNode.Distinct)
}

func TestParseSetOpWithLimitAndOffset(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	query := "SELECT 1 UNION SELECT 2 LIMIT 10 OFFSET 5"
	node, err := Parse(ctx, query)
	require.NoError(err)

	// The node should be an Offset wrapping a Limit wrapping a Union
	offsetNode, ok := node.(*plan.Offset)
	require.True(ok, "Expected *plan.Offset, got %T", node)

	limitNode, ok := offsetNode.Child.(*plan.Limit)
	require.True(ok, "Expected child to be *plan.Limit, got %T", offsetNode.Child)

	unionNode, ok := limitNode.Child.(*plan.Union)
	require.True(ok, "Expected child to be *plan.Union, got %T", limitNode.Child)
	require.True(unionNode.Distinct)
}
