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

func TestStorageRoundTrip(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("mydb", db)
	ctx := sql.NewEmptyContext()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Source: "items"},
		{Name: "value", Type: sql.Int64, Source: "items"},
	}

	err = database.Create("items", schema)
	require.NoError(t, err)

	table, ok, err := database.GetTableInsensitive(ctx, "items")
	require.NoError(t, err)
	require.True(t, ok)

	insertableTable, ok := table.(InsertableTable)
	require.True(t, ok)

	inserter := insertableTable.Inserter(ctx)
	inserter.StatementBegin(ctx)
	err = inserter.Insert(ctx, sql.NewRow(int64(1), int64(100)))
	require.NoError(t, err)
	inserter.StatementComplete(ctx)
	inserter.Close(ctx)

	// Verify
	partIter, err := table.Partitions(ctx)
	require.NoError(t, err)

	part, err := partIter.Next()
	require.NoError(t, err)

	rowIter, err := table.PartitionRows(ctx, part)
	require.NoError(t, err)

	row, err := rowIter.Next()
	require.NoError(t, err)
	assert.Equal(t, int64(1), row[0])
	assert.Equal(t, int64(100), row[1])

	_, err = rowIter.Next()
	assert.Equal(t, io.EOF, err)

	rowIter.Close()
	partIter.Close()
}

func TestStoragePersistence(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-persistence-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Phase 1: Write data
	{
		db, err := badger.Open(badger.DefaultOptions(dir))
		require.NoError(t, err)

		database := NewDatabase("mydb", db)
		ctx := sql.NewEmptyContext()
		schema := sql.Schema{
			{Name: "id", Type: sql.Int64, Source: "items"},
		}
		err = database.Create("items", schema)
		require.NoError(t, err)

		table, _, _ := database.GetTableInsensitive(ctx, "items")
		insertableTable := table.(InsertableTable)
		inserter := insertableTable.Inserter(ctx)
		inserter.StatementBegin(ctx)
		err = inserter.Insert(ctx, sql.NewRow(int64(1)))
		require.NoError(t, err)
		inserter.StatementComplete(ctx)
		inserter.Close(ctx)

		db.Close()
	}

	// Phase 2: Read back
	{
		db, err := badger.Open(badger.DefaultOptions(dir))
		require.NoError(t, err)
		defer db.Close()

		database := NewDatabase("mydb", db)
		ctx := sql.NewEmptyContext()

		// Metadata should be loaded automatically by NewDatabase
		table, ok, err := database.GetTableInsensitive(ctx, "items")
		require.NoError(t, err)
		require.True(t, ok)

		partIter, _ := table.Partitions(ctx)
		part, _ := partIter.Next()
		rowIter, _ := table.PartitionRows(ctx, part)

		row, err := rowIter.Next()
		require.NoError(t, err)
		assert.Equal(t, int64(1), row[0])
	}
}
