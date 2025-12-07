package badger

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/sql"
)

func TestTable_InsertAndRead(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Source: "table1"},
		{Name: "name", Type: sql.Text, Source: "table1"},
	}

	table := NewTable("table1", "testdb", schema, db)
	ctx := sql.NewEmptyContext()

	// Insert rows
	inserter := table.Inserter(ctx)
	row1 := sql.NewRow(int64(1), "alice")
	row2 := sql.NewRow(int64(2), "bob")

	inserter.StatementBegin(ctx)
	err = inserter.Insert(ctx, row1)
	require.NoError(t, err)
	err = inserter.Insert(ctx, row2)
	require.NoError(t, err)
	inserter.StatementComplete(ctx)
	inserter.Close(ctx)

	// Read rows back
	partitions, err := table.Partitions(ctx)
	require.NoError(t, err)

	var rows []sql.Row
	for {
		part, err := partitions.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		iter, err := table.PartitionRows(ctx, part)
		require.NoError(t, err)

		for {
			row, err := iter.Next()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			rows = append(rows, row)
		}
		iter.Close()
	}
	partitions.Close()

	assert.Len(t, rows, 2)

	foundAlice := false
	foundBob := false
	for _, r := range rows {
		if r[0] == int64(1) && r[1] == "alice" {
			foundAlice = true
		}
		if r[0] == int64(2) && r[1] == "bob" {
			foundBob = true
		}
	}
	assert.True(t, foundAlice)
	assert.True(t, foundBob)
}

func TestTable_Update(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Source: "table1"},
		{Name: "name", Type: sql.Text, Source: "table1"},
	}

	table := NewTable("table1", "testdb", schema, db)
	ctx := sql.NewEmptyContext()

	// Insert
	inserter := table.Inserter(ctx)
	inserter.StatementBegin(ctx)
	err = inserter.Insert(ctx, sql.NewRow(int64(1), "alice"))
	require.NoError(t, err)
	inserter.StatementComplete(ctx)
	inserter.Close(ctx)

	// Update
	updater := table.Updater(ctx)
	updater.StatementBegin(ctx)
	err = updater.Update(ctx, sql.NewRow(int64(1), "alice"), sql.NewRow(int64(1), "bob"))
	require.NoError(t, err)
	updater.StatementComplete(ctx)
	updater.Close(ctx)

	// Verify
	partitions, _ := table.Partitions(ctx)
	part, _ := partitions.Next()
	iter, _ := table.PartitionRows(ctx, part)

	row, err := iter.Next()
	require.NoError(t, err)
	assert.Equal(t, "bob", row[1])
	iter.Close()
}

func TestTable_Delete(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Source: "table1"},
	}

	table := NewTable("table1", "testdb", schema, db)
	ctx := sql.NewEmptyContext()

	// Insert
	inserter := table.Inserter(ctx)
	inserter.StatementBegin(ctx)
	err = inserter.Insert(ctx, sql.NewRow(int64(1)))
	require.NoError(t, err)
	inserter.StatementComplete(ctx)
	inserter.Close(ctx)

	// Delete
	deleter := table.Deleter(ctx)
	deleter.StatementBegin(ctx)
	err = deleter.Delete(ctx, sql.NewRow(int64(1)))
	require.NoError(t, err)
	deleter.StatementComplete(ctx)
	deleter.Close(ctx)

	// Verify
	partitions, _ := table.Partitions(ctx)
	part, _ := partitions.Next()
	iter, _ := table.PartitionRows(ctx, part)

	_, err = iter.Next()
	assert.Equal(t, io.EOF, err)
	iter.Close()
}
