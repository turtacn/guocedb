package badger

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/compute/sql"
)

func TestDatabase_Name(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("testdb", db)
	assert.Equal(t, "testdb", database.Name())
}

func TestDatabase_Tables(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("testdb", db)
	tables := database.Tables()
	assert.Empty(t, tables)

	schema := sql.Schema{}
	err = database.Create("table1", schema)
	require.NoError(t, err)

	tables = database.Tables()
	assert.Len(t, tables, 1)
	assert.Contains(t, tables, "table1")
}

func TestDatabase_GetTableInsensitive(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("testdb", db)
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{}
	err = database.Create("TestTable", schema)
	require.NoError(t, err)

	tbl, ok, err := database.GetTableInsensitive(ctx, "testtable")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "TestTable", tbl.Name())

	tbl, ok, err = database.GetTableInsensitive(ctx, "TESTTABLE")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "TestTable", tbl.Name())

	_, ok, err = database.GetTableInsensitive(ctx, "NonExistent")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestDatabase_DropTable(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	require.NoError(t, err)
	defer db.Close()

	database := NewDatabase("testdb", db)
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{}
	err = database.Create("table1", schema)
	require.NoError(t, err)

	err = database.DropTable(ctx, "table1")
	require.NoError(t, err)

	tables := database.Tables()
	assert.Empty(t, tables)

	err = database.DropTable(ctx, "table1")
	assert.Error(t, err)
	assert.True(t, sql.ErrTableNotFound.Is(err))
}
