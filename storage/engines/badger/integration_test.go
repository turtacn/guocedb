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

// TestCatalogTableLifecycle tests the full lifecycle of catalog and table operations
func TestCatalogTableLifecycle(t *testing.T) {
	dir, err := ioutil.TempDir("", "catalog-lifecycle-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	catalog := NewCatalog(dir)

	// Create database
	db, err := badger.Open(badger.DefaultOptions(dir + "/db1"))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("testdb", db)
	catalog.AddDatabase(database)

	// Verify database exists
	assert.True(t, catalog.HasDatabase(nil, "testdb"))

	// Create table
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "products"},
		{Name: "name", Type: sql.Text, Nullable: false, Source: "products"},
		{Name: "price", Type: sql.Float64, Nullable: false, Source: "products"},
	}

	err = database.Create("products", schema)
	require.NoError(t, err)

	// Verify Tables() includes new table
	tables := database.Tables()
	assert.Equal(t, 1, len(tables))
	assert.Contains(t, tables, "products")

	// Verify GetTableInsensitive() can find it
	table, found, err := database.GetTableInsensitive(nil, "PRODUCTS")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "products", table.Name())

	// Verify GetTableNames() includes it
	names, err := database.GetTableNames(nil)
	require.NoError(t, err)
	assert.Equal(t, 1, len(names))

	// Drop table
	err = database.DropTable(nil, "products")
	require.NoError(t, err)

	// Verify Tables() no longer includes it
	tables = database.Tables()
	assert.Equal(t, 0, len(tables))

	// Verify GetTableInsensitive() doesn't find it
	_, found, err = database.GetTableInsensitive(nil, "products")
	require.NoError(t, err)
	assert.False(t, found)
}

// TestCatalogBadgerIntegration tests complete integration from Catalog to Row operations
func TestCatalogBadgerIntegration(t *testing.T) {
	dir, err := ioutil.TempDir("", "catalog-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	catalog := NewCatalog(dir)

	// Create BadgerDB storage
	db, err := badger.Open(badger.DefaultOptions(dir + "/integration"))
	require.NoError(t, err)
	defer db.Close()

	// Create database in catalog
	database := NewDatabase("mydb", db)
	catalog.AddDatabase(database)

	// Create table
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "orders"},
		{Name: "customer", Type: sql.Text, Nullable: false, Source: "orders"},
		{Name: "total", Type: sql.Float64, Nullable: false, Source: "orders"},
	}

	err = database.Create("orders", schema)
	require.NoError(t, err)

	// Get table through catalog
	retrievedDB, err := catalog.Database(nil, "mydb")
	require.NoError(t, err)

	tables := retrievedDB.Tables()
	table, ok := tables["orders"]
	require.True(t, ok)

	// Insert rows
	ctx := sql.NewEmptyContext()
	rows := []sql.Row{
		sql.NewRow(int64(1), "Alice", float64(99.99)),
		sql.NewRow(int64(2), "Bob", float64(149.50)),
		sql.NewRow(int64(3), "Charlie", float64(75.25)),
	}

	for _, row := range rows {
		err := table.(sql.Inserter).Insert(ctx, row)
		require.NoError(t, err)
	}

	// Read back rows through table.PartitionRows()
	partitions, err := table.Partitions(ctx)
	require.NoError(t, err)
	defer partitions.Close()

	partition, err := partitions.Next()
	require.NoError(t, err)

	iter, err := table.PartitionRows(ctx, partition)
	require.NoError(t, err)
	defer iter.Close()

	var retrievedRows []sql.Row
	for {
		row, err := iter.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		retrievedRows = append(retrievedRows, row)
	}

	assert.Equal(t, len(rows), len(retrievedRows))

	// Verify data integrity
	for i, row := range retrievedRows {
		assert.Equal(t, 3, len(row), "Row %d should have 3 columns", i)
	}
}

// TestTableInserterUpdaterDeleter tests the extended table interfaces
func TestTableInserterUpdaterDeleter(t *testing.T) {
	dir, err := ioutil.TempDir("", "table-editor-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "test"},
		{Name: "value", Type: sql.Text, Nullable: false, Source: "test"},
	}

	table := NewTable("test", "testdb", schema, db)
	ctx := sql.NewEmptyContext()

	// Test Inserter
	inserter := table.Inserter(ctx)
	inserter.StatementBegin(ctx)

	testRows := []sql.Row{
		sql.NewRow(int64(1), "first"),
		sql.NewRow(int64(2), "second"),
		sql.NewRow(int64(3), "third"),
	}

	for _, row := range testRows {
		err := inserter.Insert(ctx, row)
		require.NoError(t, err)
	}

	err = inserter.StatementComplete(ctx)
	require.NoError(t, err)
	inserter.Close(ctx)

	// Verify inserts
	partitions, _ := table.Partitions(ctx)
	defer partitions.Close()
	partition, _ := partitions.Next()
	iter, _ := table.PartitionRows(ctx, partition)
	defer iter.Close()

	count := 0
	for {
		_, err := iter.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		count++
	}

	assert.Equal(t, len(testRows), count)

	// Test Updater
	updater := table.Updater(ctx)
	updater.StatementBegin(ctx)

	oldRow := sql.NewRow(int64(1), "first")
	newRow := sql.NewRow(int64(1), "updated")

	err = updater.Update(ctx, oldRow, newRow)
	require.NoError(t, err)

	err = updater.StatementComplete(ctx)
	require.NoError(t, err)
	updater.Close(ctx)

	// Test Deleter
	deleter := table.Deleter(ctx)
	deleter.StatementBegin(ctx)

	deleteRow := sql.NewRow(int64(2), "second")
	err = deleter.Delete(ctx, deleteRow)
	require.NoError(t, err)

	err = deleter.StatementComplete(ctx)
	require.NoError(t, err)
	deleter.Close(ctx)

	// Verify final state (should have 2 rows remaining)
	partitions2, _ := table.Partitions(ctx)
	defer partitions2.Close()
	partition2, _ := partitions2.Next()
	iter2, _ := table.PartitionRows(ctx, partition2)
	defer iter2.Close()

	finalCount := 0
	for {
		_, err := iter2.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		finalCount++
	}

	assert.Equal(t, 2, finalCount)
}
