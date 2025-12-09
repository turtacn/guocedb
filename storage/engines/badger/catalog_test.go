package badger

import (
	"sync"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/compute/sql"
)

// TestNewCatalog tests catalog creation
func TestNewCatalog(t *testing.T) {
	catalog := NewCatalog(t.TempDir())
	if catalog == nil {
		t.Fatal("NewCatalog should return a non-nil catalog")
	}
	
	dbs := catalog.AllDatabases(nil)
	if len(dbs) != 0 {
		t.Errorf("New catalog should have 0 databases, got %d", len(dbs))
	}
}

// TestCatalogAddDatabase tests adding databases to catalog
func TestCatalogAddDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	db, err := badger.Open(badger.DefaultOptions(tmpDir + "/db1"))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	catalog.AddDatabase(database)
	
	// Verify database was added
	dbs := catalog.AllDatabases(nil)
	if len(dbs) != 1 {
		t.Errorf("Expected 1 database, got %d", len(dbs))
	}
	
	// Verify we can retrieve it
	retrieved, err := catalog.Database(nil, "testdb")
	if err != nil {
		t.Fatalf("Failed to retrieve database: %v", err)
	}
	if retrieved.Name() != "testdb" {
		t.Errorf("Expected name 'testdb', got %s", retrieved.Name())
	}
}

// TestCatalogDatabaseCaseInsensitive tests case-insensitive database lookup
func TestCatalogDatabaseCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	db, err := badger.Open(badger.DefaultOptions(tmpDir + "/db1"))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("TestDB", db)
	catalog.AddDatabase(database)
	
	testCases := []string{"TestDB", "testdb", "TESTDB", "testDB"}
	for _, name := range testCases {
		retrieved, err := catalog.Database(nil, name)
		if err != nil {
			t.Errorf("Failed to retrieve database with name %s: %v", name, err)
		}
		if retrieved == nil {
			t.Errorf("Expected non-nil database for name %s", name)
		}
	}
}

// TestCatalogHasDatabase tests HasDatabase method
func TestCatalogHasDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	// Test non-existent database
	if catalog.HasDatabase(nil, "nonexistent") {
		t.Error("HasDatabase should return false for non-existent database")
	}
	
	db, err := badger.Open(badger.DefaultOptions(tmpDir + "/db1"))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("mydb", db)
	catalog.AddDatabase(database)
	
	// Test existing database (exact case)
	if !catalog.HasDatabase(nil, "mydb") {
		t.Error("HasDatabase should return true for existing database")
	}
	
	// Test case-insensitive
	if !catalog.HasDatabase(nil, "MYDB") {
		t.Error("HasDatabase should be case-insensitive")
	}
	if !catalog.HasDatabase(nil, "MyDb") {
		t.Error("HasDatabase should be case-insensitive")
	}
}

// TestCatalogDatabaseNotFound tests error handling for missing database
func TestCatalogDatabaseNotFound(t *testing.T) {
	catalog := NewCatalog(t.TempDir())
	
	_, err := catalog.Database(nil, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent database")
	}
	
	// Verify it's the correct error type
	if !sql.ErrDatabaseNotFound.Is(err) {
		t.Errorf("Expected ErrDatabaseNotFound, got %v", err)
	}
}

// TestCatalogAllDatabases tests retrieving all databases
func TestCatalogAllDatabases(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	// Create multiple databases
	dbNames := []string{"db1", "db2", "db3"}
	for i, name := range dbNames {
		db, err := badger.Open(badger.DefaultOptions(tmpDir + "/" + name))
		if err != nil {
			t.Fatalf("Failed to open badger: %v", err)
		}
		defer db.Close()
		
		database := NewDatabase(name, db)
		catalog.AddDatabase(database)
		
		// Verify count after each addition
		dbs := catalog.AllDatabases(nil)
		if len(dbs) != i+1 {
			t.Errorf("Expected %d databases after adding %s, got %d", i+1, name, len(dbs))
		}
	}
	
	// Verify final count
	dbs := catalog.AllDatabases(nil)
	if len(dbs) != len(dbNames) {
		t.Errorf("Expected %d databases, got %d", len(dbNames), len(dbs))
	}
	
	// Verify all databases are present
	dbMap := make(map[string]bool)
	for _, db := range dbs {
		dbMap[db.Name()] = true
	}
	for _, name := range dbNames {
		if !dbMap[name] {
			t.Errorf("Database %s not found in AllDatabases", name)
		}
	}
}

// TestCatalogTables tests retrieving tables from a database via catalog
func TestCatalogTables(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	db, err := badger.Open(badger.DefaultOptions(tmpDir + "/db1"))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	catalog.AddDatabase(database)
	
	// Create a table
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "users"},
		{Name: "name", Type: sql.Text, Nullable: true, Source: "users"},
	}
	
	err = database.Create("users", schema)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	// Get tables via catalog
	tables, err := catalog.Tables(nil, "testdb")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}
	
	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}
	
	if _, ok := tables["users"]; !ok {
		t.Error("Table 'users' not found in catalog.Tables()")
	}
}

// TestCatalogConcurrentAccess tests concurrent access to catalog
func TestCatalogConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	catalog := NewCatalog(tmpDir)
	
	var wg sync.WaitGroup
	numDatabases := 10
	
	// Concurrent writes
	for i := 0; i < numDatabases; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			dbPath := tmpDir + "/concurrent_" + string(rune('a'+id))
			db, err := badger.Open(badger.DefaultOptions(dbPath))
			if err != nil {
				t.Errorf("Failed to open badger: %v", err)
				return
			}
			defer db.Close()
			
			dbName := "db_" + string(rune('a'+id))
			database := NewDatabase(dbName, db)
			catalog.AddDatabase(database)
		}(i)
	}
	
	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = catalog.AllDatabases(nil)
			_ = catalog.HasDatabase(nil, "db_a")
		}()
	}
	
	wg.Wait()
	
	// Verify all databases were added
	dbs := catalog.AllDatabases(nil)
	if len(dbs) != numDatabases {
		t.Errorf("Expected %d databases after concurrent additions, got %d", numDatabases, len(dbs))
	}
}

// TestDatabaseName tests Database.Name() method
func TestDatabaseName(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	if database.Name() != "testdb" {
		t.Errorf("Expected name 'testdb', got %s", database.Name())
	}
}

// TestDatabaseTables tests Database.Tables() method
func TestDatabaseTables(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	
	// Initially empty
	tables := database.Tables()
	if len(tables) != 0 {
		t.Errorf("Expected 0 tables, got %d", len(tables))
	}
	
	// Create a table
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "users"},
	}
	
	err = database.Create("users", schema)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	// Should have 1 table
	tables = database.Tables()
	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}
	
	if table, ok := tables["users"]; !ok {
		t.Error("Table 'users' not found")
	} else if table.Name() != "users" {
		t.Errorf("Expected table name 'users', got %s", table.Name())
	}
}

// TestDatabaseGetTableInsensitive tests case-insensitive table lookup
func TestDatabaseGetTableInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	
	// Create a table
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "Users"},
	}
	
	err = database.Create("Users", schema)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	testCases := []string{"Users", "users", "USERS", "uSeRs"}
	for _, name := range testCases {
		table, found, err := database.GetTableInsensitive(nil, name)
		if err != nil {
			t.Errorf("GetTableInsensitive(%s) returned error: %v", name, err)
		}
		if !found {
			t.Errorf("GetTableInsensitive(%s) should find the table", name)
		}
		if table == nil {
			t.Errorf("GetTableInsensitive(%s) returned nil table", name)
		}
	}
	
	// Test non-existent table
	table, found, err := database.GetTableInsensitive(nil, "nonexistent")
	if err != nil {
		t.Errorf("GetTableInsensitive should not return error for non-existent table: %v", err)
	}
	if found {
		t.Error("GetTableInsensitive should return false for non-existent table")
	}
	if table != nil {
		t.Error("GetTableInsensitive should return nil table for non-existent table")
	}
}

// TestDatabaseGetTableNames tests getting all table names
func TestDatabaseGetTableNames(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	if err != nil {
		t.Fatalf("Failed to open badger: %v", err)
	}
	defer db.Close()
	
	database := NewDatabase("testdb", db)
	
	// Initially empty
	names, err := database.GetTableNames(nil)
	if err != nil {
		t.Fatalf("GetTableNames failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected 0 table names, got %d", len(names))
	}
	
	// Create multiple tables
	tableNames := []string{"users", "posts", "comments"}
	schema := sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: "test"},
	}
	
	for _, name := range tableNames {
		err = database.Create(name, schema)
		if err != nil {
			t.Fatalf("Failed to create table %s: %v", name, err)
		}
	}
	
	// Get all names
	names, err = database.GetTableNames(nil)
	if err != nil {
		t.Fatalf("GetTableNames failed: %v", err)
	}
	if len(names) != len(tableNames) {
		t.Errorf("Expected %d table names, got %d", len(tableNames), len(names))
	}
	
	// Verify all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	for _, expected := range tableNames {
		if !nameMap[expected] {
			t.Errorf("Table name %s not found in GetTableNames", expected)
		}
	}
}
