// Package memory implements the in-memory metadata catalog for Guocedb.
// This file provides a memory-based metadata management implementation,
// primarily for development, testing, or scenarios with low persistence requirements.
// It relies on the interfaces defined in compute/catalog/catalog.go,
// data types from common/types/value/value.go, and common/errors and common/log.
// In the early development stages of Guocedb, it will serve as the primary metadata management approach.
package memory

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/compute/catalog" // Import the catalog interfaces and structs
	"github.com/turtacn/guocedb/interfaces"      // Import general interfaces like ID
)

// ensure that MemoryCatalog implements the catalog.Catalog interface.
var _ catalog.Catalog = (*MemoryCatalog)(nil)

// MemoryCatalog implements the Catalog interface using in-memory data structures.
// It uses maps to store database, table, and index metadata.
type MemoryCatalog struct {
	databases map[string]*catalog.DatabaseMetadata                      // Map: dbName -> DatabaseMetadata
	tables    map[interfaces.DatabaseID]map[string]*catalog.TableSchema // Map: dbID -> (tableName -> TableSchema)
	indexes   map[interfaces.TableID]map[string]*catalog.IndexSchema    // Map: tableID -> (indexName -> IndexSchema)

	// Atomic counters for generating unique IDs
	nextDatabaseID uint64
	nextTableID    uint64
	nextColumnID   uint64
	nextIndexID    uint64

	mu sync.RWMutex // Mutex for protecting access to catalog data
}

// NewMemoryCatalog creates and returns a new initialized MemoryCatalog.
func NewMemoryCatalog() *MemoryCatalog {
	return &MemoryCatalog{
		databases: make(map[string]*catalog.DatabaseMetadata),
		tables:    make(map[interfaces.DatabaseID]map[string]*catalog.TableSchema),
		indexes:   make(map[interfaces.TableID]map[string]*catalog.IndexSchema),

		nextDatabaseID: 0,
		nextTableID:    0,
		nextColumnID:   0,
		nextIndexID:    0,
	}
}

// Initialize prepares the MemoryCatalog for use.
// For an in-memory catalog, this primarily involves logging its readiness.
func (mc *MemoryCatalog) Initialize() error {
	log.Infof(enum.ComponentCatalog, "MemoryCatalog initialized successfully.")
	return nil
}

// Shutdown gracefully closes the MemoryCatalog.
// For an in-memory catalog, this involves clearing data (optional) and logging.
func (mc *MemoryCatalog) Shutdown() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Clear maps to release memory (optional, as GC will handle it anyway)
	mc.databases = make(map[string]*catalog.DatabaseMetadata)
	mc.tables = make(map[interfaces.DatabaseID]map[string]*catalog.TableSchema)
	mc.indexes = make(map[interfaces.TableID]map[string]*catalog.IndexSchema)

	log.Infof(enum.ComponentCatalog, "MemoryCatalog shut down.")
	return nil
}

// --- Database Operations ---

// CreateDatabase adds a new database to the catalog.
func (mc *MemoryCatalog) CreateDatabase(dbName string) (*catalog.DatabaseMetadata, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.databases[dbName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseAlreadyExists,
			fmt.Sprintf("database '%s' already exists", dbName), nil)
	}

	dbID := mc.GenerateDatabaseID()
	dbMeta := &catalog.DatabaseMetadata{
		ID:           interfaces.DatabaseID(dbID),
		DatabaseName: dbName,
	}
	mc.databases[dbName] = dbMeta
	mc.tables[interfaces.DatabaseID(dbID)] = make(map[string]*catalog.TableSchema) // Initialize table map for new DB

	log.Infof(enum.ComponentCatalog, "Database '%s' (ID: %d) created in MemoryCatalog.", dbName, dbID)
	return dbMeta, nil
}

// DropDatabase removes a database and all its associated tables and indexes from the catalog.
func (mc *MemoryCatalog) DropDatabase(dbName string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	dbID := dbMeta.ID

	// Delete all tables and their indexes within this database
	if tablesInDb, ok := mc.tables[dbID]; ok {
		for _, tableSchema := range tablesInDb {
			delete(mc.indexes, tableSchema.ID) // Remove all indexes for this table
		}
	}
	delete(mc.tables, dbID)      // Remove table map for this database
	delete(mc.databases, dbName) // Remove the database itself

	log.Infof(enum.ComponentCatalog, "Database '%s' (ID: %d) and all its contents dropped from MemoryCatalog.", dbName, dbID)
	return nil
}

// GetDatabase retrieves metadata for a specific database.
func (mc *MemoryCatalog) GetDatabase(dbName string) (*catalog.DatabaseMetadata, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}
	return dbMeta, nil
}

// ListDatabases returns a list of all database names in the catalog.
func (mc *MemoryCatalog) ListDatabases() ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbNames := make([]string, 0, len(mc.databases))
	for name := range mc.databases {
		dbNames = append(dbNames, name)
	}
	return dbNames, nil
}

// --- Table Operations ---

// CreateTable adds a new table schema to the specified database.
func (mc *MemoryCatalog) CreateTable(dbName string, tableSchema *catalog.TableSchema) (*catalog.TableSchema, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		// This should not happen if tables map is initialized upon db creation, but as a safeguard:
		tablesInDb = make(map[string]*catalog.TableSchema)
		mc.tables[dbMeta.ID] = tablesInDb
	}

	if _, exists := tablesInDb[tableSchema.TableName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableAlreadyExists,
			fmt.Sprintf("table '%s' already exists in database '%s'", tableSchema.TableName, dbName), nil)
	}

	// Assign unique IDs to the table and its columns
	tableSchema.ID = interfaces.TableID(mc.GenerateTableID())
	for _, colDef := range tableSchema.Columns {
		colDef.ID = interfaces.ColumnID(mc.GenerateColumnID())
	}

	tablesInDb[tableSchema.TableName] = tableSchema
	mc.indexes[tableSchema.ID] = make(map[string]*catalog.IndexSchema) // Initialize index map for new table

	log.Infof(enum.ComponentCatalog, "Table '%s' (ID: %d) created in database '%s' (ID: %d).",
		tableSchema.TableName, tableSchema.ID, dbName, dbMeta.ID)
	return tableSchema, nil
}

// DropTable removes a table and its associated indexes from the specified database.
func (mc *MemoryCatalog) DropTable(dbName, tableName string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	delete(mc.indexes, tableSchema.ID) // Remove all indexes associated with this table
	delete(tablesInDb, tableName)      // Remove the table itself

	log.Infof(enum.ComponentCatalog, "Table '%s' (ID: %d) dropped from database '%s'.", tableName, tableSchema.ID, dbName)
	return nil
}

// GetTable retrieves the schema for a specific table within a database.
func (mc *MemoryCatalog) GetTable(dbName, tableName string) (*catalog.TableSchema, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}
	return tableSchema, nil
}

// ListTables returns a list of all table names within a database.
func (mc *MemoryCatalog) ListTables(dbName string) ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return []string{}, nil // No tables found, return empty slice
	}

	tableNames := make([]string, 0, len(tablesInDb))
	for name := range tablesInDb {
		tableNames = append(tableNames, name)
	}
	return tableNames, nil
}

// --- Index Operations ---

// CreateIndex adds a new index schema to the specified table.
func (mc *MemoryCatalog) CreateIndex(dbName, tableName string, indexSchema *catalog.IndexSchema) (*catalog.IndexSchema, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := mc.indexes[tableSchema.ID]
	if indexesInTable == nil {
		// This should have been initialized when the table was created, but as a safeguard
		indexesInTable = make(map[string]*catalog.IndexSchema)
		mc.indexes[tableSchema.ID] = indexesInTable
	}

	if _, exists := indexesInTable[indexSchema.IndexName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexAlreadyExists,
			fmt.Sprintf("index '%s' already exists on table '%s'", indexSchema.IndexName, tableName), nil)
	}

	// Assign unique ID to the index
	indexSchema.ID = interfaces.IndexID(mc.GenerateIndexID())
	indexSchema.TableID = tableSchema.ID // Ensure index links to the correct table ID

	indexesInTable[indexSchema.IndexName] = indexSchema

	log.Infof(enum.ComponentCatalog, "Index '%s' (ID: %d) created on table '%s' (ID: %d) in database '%s'.",
		indexSchema.IndexName, indexSchema.ID, tableName, tableSchema.ID, dbName)
	return indexSchema, nil
}

// DropIndex removes an index from the specified table.
func (mc *MemoryCatalog) DropIndex(dbName, tableName, indexName string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := mc.indexes[tableSchema.ID]
	if indexesInTable == nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	indexSchema, exists := indexesInTable[indexName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	delete(indexesInTable, indexName)

	log.Infof(enum.ComponentCatalog, "Index '%s' (ID: %d) dropped from table '%s'.", indexName, indexSchema.ID, tableName)
	return nil
}

// GetIndex retrieves the schema for a specific index.
func (mc *MemoryCatalog) GetIndex(dbName, tableName, indexName string) (*catalog.IndexSchema, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := mc.indexes[tableSchema.ID]
	if indexesInTable == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	indexSchema, exists := indexesInTable[indexName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}
	return indexSchema, nil
}

// ListIndexes returns a list of all index names for a given table.
func (mc *MemoryCatalog) ListIndexes(dbName, tableName string) ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	dbMeta, exists := mc.databases[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := mc.tables[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := mc.indexes[tableSchema.ID]
	if indexesInTable == nil {
		return []string{}, nil // No indexes found, return empty slice
	}

	indexNames := make([]string, 0, len(indexesInTable))
	for name := range indexesInTable {
		indexNames = append(indexNames, name)
	}
	return indexNames, nil
}

// --- Utility/Internal Methods for ID Generation ---

// GenerateDatabaseID generates a unique ID for a new database.
func (mc *MemoryCatalog) GenerateDatabaseID() interfaces.DatabaseID {
	return interfaces.DatabaseID(atomic.AddUint64(&mc.nextDatabaseID, 1))
}

// GenerateTableID generates a unique ID for a new table within a database.
func (mc *MemoryCatalog) GenerateTableID() interfaces.TableID {
	return interfaces.TableID(atomic.AddUint64(&mc.nextTableID, 1))
}

// GenerateColumnID generates a unique ID for a new column within a table.
func (mc *MemoryCatalog) GenerateColumnID() interfaces.ColumnID {
	return interfaces.ColumnID(atomic.AddUint64(&mc.nextColumnID, 1))
}

// GenerateIndexID generates a unique ID for a new index within a table.
func (mc *MemoryCatalog) GenerateIndexID() interfaces.IndexID {
	return interfaces.IndexID(atomic.AddUint64(&mc.nextIndexID, 1))
}

//Personal.AI order the ending
