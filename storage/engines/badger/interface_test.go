package badger

import (
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/compute/sql"
)

// Compile-time interface verification - if interfaces are not fully implemented, compilation will fail

// Verify Database implements sql.Database
var _ sql.Database = (*Database)(nil)

// Verify Catalog implements DatabaseProvider
var _ DatabaseProvider = (*Catalog)(nil)

// Verify Table implements sql.Table
var _ sql.Table = (*Table)(nil)

// Verify Table implements sql.Inserter
var _ sql.Inserter = (*Table)(nil)

// Verify Table implements InsertableTable
var _ InsertableTable = (*Table)(nil)

// Verify Table implements UpdatableTable
var _ UpdatableTable = (*Table)(nil)

// Verify Table implements DeletableTable
var _ DeletableTable = (*Table)(nil)

// Verify rowEditor implements RowInserter
var _ RowInserter = (*rowEditor)(nil)

// Verify rowEditor implements RowUpdater
var _ RowUpdater = (*rowEditor)(nil)

// Verify rowEditor implements RowDeleter
var _ RowDeleter = (*rowEditor)(nil)

// TestDatabaseImplementsDatabase verifies Database interface implementation at runtime
func TestDatabaseImplementsDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()

	database := NewDatabase("testdb", db)
	
	// Verify it can be assigned to sql.Database
	var sqlDB sql.Database = database
	if sqlDB.Name() != "testdb" {
		t.Errorf("Expected name 'testdb', got %s", sqlDB.Name())
	}
	
	// Verify Tables() returns a map
	tables := sqlDB.Tables()
	if tables == nil {
		t.Error("Tables() should return a non-nil map")
	}
}

// TestCatalogImplementsDatabaseProvider verifies Catalog interface implementation
func TestCatalogImplementsDatabaseProvider(t *testing.T) {
	catalog := NewCatalog(t.TempDir())
	
	// Verify it can be assigned to DatabaseProvider
	var provider DatabaseProvider = catalog
	
	// Test HasDatabase on empty catalog
	if provider.HasDatabase(nil, "nonexistent") {
		t.Error("HasDatabase should return false for nonexistent database")
	}
	
	// Test AllDatabases on empty catalog
	dbs := provider.AllDatabases(nil)
	if len(dbs) != 0 {
		t.Errorf("Expected 0 databases, got %d", len(dbs))
	}
}

// TestTableImplementsInterfaces verifies Table interface implementations
func TestTableImplementsInterfaces(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()

	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "test"},
		{Name: "name", Type: sql.Text, Nullable: true, Source: "test"},
	}
	
	table := NewTable("test", "testdb", schema, db)
	
	// Verify sql.Table
	var sqlTable sql.Table = table
	if sqlTable.Name() != "test" {
		t.Errorf("Expected name 'test', got %s", sqlTable.Name())
	}
	
	// Verify sql.Inserter
	var inserter sql.Inserter = table
	if inserter == nil {
		t.Error("Table should implement sql.Inserter")
	}
	
	// Verify InsertableTable
	var insertable InsertableTable = table
	rowInserter := insertable.Inserter(nil)
	if rowInserter == nil {
		t.Error("Inserter() should return a non-nil RowInserter")
	}
	
	// Verify UpdatableTable
	var updatable UpdatableTable = table
	rowUpdater := updatable.Updater(nil)
	if rowUpdater == nil {
		t.Error("Updater() should return a non-nil RowUpdater")
	}
	
	// Verify DeletableTable
	var deletable DeletableTable = table
	rowDeleter := deletable.Deleter(nil)
	if rowDeleter == nil {
		t.Error("Deleter() should return a non-nil RowDeleter")
	}
}
