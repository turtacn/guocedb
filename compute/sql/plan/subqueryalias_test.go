package plan

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/mem"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/sql/expression"
)

func TestSubqueryAliasSchema(t *testing.T) {
	require := require.New(t)

	tableSchema := sql.Schema{
		{Name: "foo", Type: sql.Text, Nullable: false, Source: "bar"},
		{Name: "baz", Type: sql.Text, Nullable: false, Source: "bar"},
	}

	subquerySchema := sql.Schema{
		{Name: "foo", Type: sql.Text, Nullable: false, Source: "alias"},
		{Name: "baz", Type: sql.Text, Nullable: false, Source: "alias"},
	}

	table := mem.NewTable("bar", tableSchema)

	subquery := NewProject(
		[]sql.Expression{
			expression.NewGetField(0, sql.Text, "foo", false),
			expression.NewGetField(1, sql.Text, "baz", false),
		},
		NewResolvedTable(table),
	)

	require.Equal(
		subquerySchema,
		NewSubqueryAlias("alias", subquery).Schema(),
	)
}
