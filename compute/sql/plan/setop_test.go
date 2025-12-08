package plan

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/mem"
	"github.com/turtacn/guocedb/compute/sql"
)

func TestIntersect(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	// Schema: [col1 int]
	schema := sql.Schema{
		{Name: "col1", Type: sql.Int64, Nullable: true},
	}

	// Left: [1, 1, 2, 3]
	leftTable := mem.NewTable("left", schema)
	_ = leftTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(2)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(3)))

	// Right: [1, 2, 4]
	rightTable := mem.NewTable("right", schema)
	_ = rightTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = rightTable.Insert(ctx, sql.NewRow(int64(2)))
	_ = rightTable.Insert(ctx, sql.NewRow(int64(4)))

	// Intersect Distinct: [1, 2]
	node := NewIntersect(NewResolvedTable(leftTable), NewResolvedTable(rightTable), true)
	rows := getRows(t, ctx, node)
	require.ElementsMatch([]sql.Row{{int64(1)}, {int64(2)}}, rows)
}

func TestExcept(t *testing.T) {
	require := require.New(t)
	ctx := sql.NewEmptyContext()

	schema := sql.Schema{
		{Name: "col1", Type: sql.Int64, Nullable: true},
	}

	// Left: [1, 1, 2, 2, 3]
	leftTable := mem.NewTable("left", schema)
	_ = leftTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(2)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(2)))
	_ = leftTable.Insert(ctx, sql.NewRow(int64(3)))

	// Right: [1, 2]
	rightTable := mem.NewTable("right", schema)
	_ = rightTable.Insert(ctx, sql.NewRow(int64(1)))
	_ = rightTable.Insert(ctx, sql.NewRow(int64(2)))

	// Except Distinct: [3]
	// 1 is in Right -> remove all 1s from Left.
	// 2 is in Right -> remove all 2s from Left.
	// 3 is not -> keep 3.
	node := NewExcept(NewResolvedTable(leftTable), NewResolvedTable(rightTable), true)
	rows := getRows(t, ctx, node)
	require.ElementsMatch([]sql.Row{{int64(3)}}, rows)
}

func getRows(t *testing.T, ctx *sql.Context, n sql.Node) []sql.Row {
	iter, err := n.RowIter(ctx)
	require.NoError(t, err)
	var rows []sql.Row
	for {
		row, err := iter.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}
	iter.Close()
	return rows
}
