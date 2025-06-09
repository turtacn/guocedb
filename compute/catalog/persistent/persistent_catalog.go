// Package persistent implements the persistent metadata catalog for Guocedb.
// This file is responsible for persisting metadata to the storage layer (via storage/sal/adapter.go),
// ensuring that metadata is not lost after database restarts.
// It relies on the interfaces defined in compute/catalog/catalog.go,
// data types from common/types/value/value.go, storage/sal/adapter.go for data read/write,
// and internal/encoding/encoding.go for metadata serialization.
// This file is a crucial step in Guocedb's evolution from an in-memory database to a general-purpose centralized database.
package persistent

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/compute/catalog"         // Import the catalog interfaces and structs
	"github.com/turtacn/guocedb/interfaces"              // Import general interfaces like ID
	"github.com/turtacn/guocedb/internal/encoding"       // Import for metadata serialization/deserialization
	"github.com/turtacn/guocedb/storage/sal"             // Import the Storage Abstraction Layer (for adapter and byte conversions)
)

// ensure that PersistentCatalog implements the catalog.Catalog interface.
var _ catalog.Catalog = (*PersistentCatalog)(nil)

// Key prefixes for metadata storage in the underlying key-value store.
const (
	// DatabaseMetadataPrefix is used for storing database metadata.
	// Key: DatabaseMetadataPrefix + DatabaseID (8 bytes)
	DatabaseMetadataPrefix byte = 0x01
	// TableMetadataPrefix is used for storing table schemas.
	// Key: TableMetadataPrefix + DatabaseID (8 bytes) + TableID (8 bytes)
	TableMetadataPrefix byte = 0x02
	// IndexMetadataPrefix is used for storing index schemas.
	// Key: IndexMetadataPrefix + DatabaseID (8 bytes) + TableID (8 bytes) + IndexID (8 bytes)
	IndexMetadataPrefix byte = 0x03
	// IDSequencePrefix is used for storing ID sequence counters.
	// Key: IDSequencePrefix + TypeIdentifier (e.g., "db", "table", "col", "idx")
	IDSequencePrefix byte = 0x04
)

// PersistentCatalog implements the Catalog interface with persistent storage.
// It leverages the storage abstraction layer (SAL) to store and retrieve metadata.
type PersistentCatalog struct {
	storageAdapter interfaces.StorageAdapter // The storage abstraction layer adapter
	systemDBName   string                    // Name of the dedicated system database for catalog metadata

	// In-memory cache for frequently accessed metadata to reduce storage I/O.
	// These caches are updated on metadata changes and invalidated/reloaded
	// during initialization or on explicit refresh.
	databaseCache map[string]*catalog.DatabaseMetadata
	tableCache    map[interfaces.DatabaseID]map[string]*catalog.TableSchema
	indexCache    map[interfaces.TableID]map[string]*catalog.IndexSchema

	// Atomic counters for generating unique IDs, persisted in storage.
	nextDatabaseID uint64
	nextTableID    uint64
	nextColumnID   uint64
	nextIndexID    uint64

	mu sync.RWMutex // Mutex for protecting access to catalog data and caches
}

// NewPersistentCatalog creates a new PersistentCatalog instance.
// It requires a configured StorageAdapter to persist its metadata.
func NewPersistentCatalog(adapter interfaces.StorageAdapter) (*PersistentCatalog, error) {
	if adapter == nil {
		return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
			"storage adapter cannot be nil for persistent catalog", nil)
	}

	pc := &PersistentCatalog{
		storageAdapter: adapter,
		systemDBName:   "guocedb_catalog", // Fixed name for the system catalog database
		databaseCache:  make(map[string]*catalog.DatabaseMetadata),
		tableCache:     make(map[interfaces.DatabaseID]map[string]*catalog.TableSchema),
		indexCache:     make(map[interfaces.TableID]map[string]*catalog.IndexSchema),
		nextDatabaseID: 0, // Will be loaded from storage during Initialize
		nextTableID:    0,
		nextColumnID:   0,
		nextIndexID:    0,
	}

	return pc, nil
}

// Initialize prepares the PersistentCatalog for use.
// It creates/opens a dedicated system database for metadata and loads existing metadata.
func (pc *PersistentCatalog) Initialize() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 1. Ensure the system database exists.
	// The StorageAdapter's OpenDatabase method should handle creation if it doesn't exist.
	// This implicitly handles `guocedb_catalog` database.

	// 2. Load ID sequence counters
	if err := pc.loadIDSequences(); err != nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeCatalogInitFailed,
			"failed to load ID sequences", err)
	}

	// 3. Load all existing metadata into memory caches
	if err := pc.loadAllMetadata(); err != nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeCatalogInitFailed,
			"failed to load all existing metadata", err)
	}

	log.Infof(enum.ComponentCatalog, "PersistentCatalog initialized successfully. Loaded %d databases, %d tables, %d indexes.",
		len(pc.databaseCache), countTotalTables(pc.tableCache), countTotalIndexes(pc.indexCache))
	return nil
}

// Shutdown gracefully closes the PersistentCatalog.
func (pc *PersistentCatalog) Shutdown() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// For a persistent catalog, shutdown typically involves flushing any pending metadata changes
	// and ensuring the underlying storage connection is properly closed.
	// Since most operations are transactional and commit immediately, explicit flush here might be minimal.
	// The StorageAdapter.Close() method should handle underlying storage closure.

	// Clear caches to release memory
	pc.databaseCache = nil
	pc.tableCache = nil
	pc.indexCache = nil

	log.Infof(enum.ComponentCatalog, "PersistentCatalog shut down.")
	return nil
}

// loadIDSequences loads the highest used ID for each type from persistent storage.
func (pc *PersistentCatalog) loadIDSequences() error {
	// Use the storage adapter to read ID sequences
	// This will typically involve a read-only transaction or direct reads.
	var err error

	pc.nextDatabaseID, err = pc.loadSequence("db")
	if err != nil {
		return err
	}

	pc.nextTableID, err = pc.loadSequence("table")
	if err != nil {
		return err
	}

	pc.nextColumnID, err = pc.loadSequence("col")
	if err != nil {
		return err
	}

	pc.nextIndexID, err = pc.loadSequence("idx")
	if err != nil {
		return err
	}

	log.Infof(enum.ComponentCatalog, "Loaded ID sequences: DB=%d, Table=%d, Column=%d, Index=%d",
		pc.nextDatabaseID, pc.nextTableID, pc.nextColumnID, pc.nextIndexID)
	return nil
}

// loadSequence is a helper to load a specific ID sequence from storage.
func (pc *PersistentCatalog) loadSequence(seqName string) (uint64, error) {
	key := encoding.EncodeIDSequenceKey(seqName)
	val, err := pc.storageAdapter.Read(pc.systemDBName, key) // Read from system catalog DB
	if err == nil {
		if len(val) != 8 { // uint64 is 8 bytes
			return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("invalid length for ID sequence '%s' value: %d bytes", seqName, len(val)), nil)
		}
		return sal.BytesToUint64(val), nil
	} else if errors.Is(err, errors.NewGuocedbError(enum.ErrNotFound, errors.CodeKeyNotFound, "", nil)) {
		return 0, nil // Not found, start from 0
	}
	return 0, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
		fmt.Sprintf("failed to read ID sequence '%s' from storage", seqName), err)
}

// updateSequence is a helper to persist an updated ID sequence value.
// This is typically called within a transaction that also commits the metadata change.
func (pc *PersistentCatalog) updateSequence(seqName string, newID uint64) error {
	key := encoding.EncodeIDSequenceKey(seqName)
	val := sal.Uint64ToBytes(newID)
	// We'll use a direct write here, assuming `Write` for a single key is fine,
	// or that the caller (`Generate*ID`) ensures transactional consistency.
	// For ID sequences, a separate, fast, non-transactional write might be acceptable
	// if the risk of losing the *very latest* ID in a crash (and thus reusing it)
	// is mitigated by other mechanisms (e.g., ID generation reserving blocks).
	// For simplicity, we'll assume `pc.storageAdapter.Write` handles it.
	err := pc.storageAdapter.Write(pc.systemDBName, key, val)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist ID sequence '%s' to storage", seqName), err)
	}
	return nil
}

// loadAllMetadata loads all database, table, and index metadata into in-memory caches.
func (pc *PersistentCatalog) loadAllMetadata() error {
	// Use the storage adapter to iterate and load all metadata.
	// This might involve multiple scans with different prefixes.

	// Load Databases
	dbKeys, err := pc.storageAdapter.ListKeysWithPrefix(pc.systemDBName, []byte{DatabaseMetadataPrefix})
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			"failed to list database metadata keys", err)
	}
	for _, key := range dbKeys {
		val, err := pc.storageAdapter.Read(pc.systemDBName, key)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
				fmt.Sprintf("failed to read database metadata for key %x", key), err)
		}
		dbMeta, err := encoding.DecodeDatabaseMetadataValue(val)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to decode database metadata for key %x", key), err)
		}
		pc.databaseCache[dbMeta.DatabaseName] = dbMeta
		pc.tableCache[dbMeta.ID] = make(map[string]*catalog.TableSchema) // Initialize table map for loaded DB
	}

	// Load Tables
	tableKeys, err := pc.storageAdapter.ListKeysWithPrefix(pc.systemDBName, []byte{TableMetadataPrefix})
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			"failed to list table metadata keys", err)
	}
	for _, key := range tableKeys {
		val, err := pc.storageAdapter.Read(pc.systemDBName, key)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
				fmt.Sprintf("failed to read table schema for key %x", key), err)
		}
		tableSchema, err := encoding.DecodeTableSchemaValue(val)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to decode table schema for key %x", key), err)
		}

		dbID := encoding.DecodeTableKeyDatabaseID(key) // Extract DB ID from key
		if _, ok := pc.tableCache[dbID]; !ok {
			log.Warnf(enum.ComponentCatalog, "Found table '%s' (ID: %d) with no corresponding database ID %d in cache. Database likely dropped without full cleanup.",
				tableSchema.TableName, tableSchema.ID, dbID)
			continue // Skip this table as its parent DB is missing from cache
		}
		pc.tableCache[dbID][tableSchema.TableName] = tableSchema
		pc.indexCache[tableSchema.ID] = make(map[string]*catalog.IndexSchema) // Initialize index map for loaded table
	}

	// Load Indexes
	indexKeys, err := pc.storageAdapter.ListKeysWithPrefix(pc.systemDBName, []byte{IndexMetadataPrefix})
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			"failed to list index metadata keys", err)
	}
	for _, key := range indexKeys {
		val, err := pc.storageAdapter.Read(pc.systemDBName, key)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
				fmt.Sprintf("failed to read index schema for key %x", key), err)
		}
		indexSchema, err := encoding.DecodeIndexSchemaValue(val)
		if err != nil {
			return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to decode index schema for key %x", key), err)
		}

		// Ensure the table exists in cache for this index
		if _, ok := pc.indexCache[indexSchema.TableID]; !ok {
			log.Warnf(enum.ComponentCatalog, "Found index '%s' (ID: %d) with no corresponding table ID %d in cache. Table likely dropped without full cleanup.",
				indexSchema.IndexName, indexSchema.ID, indexSchema.TableID)
			continue // Skip this index as its parent table is missing from cache
		}
		pc.indexCache[indexSchema.TableID][indexSchema.IndexName] = indexSchema
	}

	return nil
}

// --- Database Operations ---

// CreateDatabase adds a new database to the catalog.
func (pc *PersistentCatalog) CreateDatabase(dbName string) (*catalog.DatabaseMetadata, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if _, exists := pc.databaseCache[dbName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseAlreadyExists,
			fmt.Sprintf("database '%s' already exists", dbName), nil)
	}

	// Use a storage transaction for atomicity
	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable) // Use serializable for metadata changes
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	dbID := pc.GenerateDatabaseID() // Use the atomic ID generator
	dbMeta := &catalog.DatabaseMetadata{
		ID:           dbID,
		DatabaseName: dbName,
	}

	// Persist to storage
	key := encoding.EncodeDatabaseMetadataKey(dbID)
	val, err := encoding.EncodeDatabaseMetadataValue(dbMeta)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode database metadata for '%s'", dbName), err)
	}
	if err = txn.Write(key, val); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist database '%s' metadata", dbName), err)
	}

	// Update in-memory cache
	pc.databaseCache[dbName] = dbMeta
	pc.tableCache[dbID] = make(map[string]*catalog.TableSchema)

	log.Infof(enum.ComponentCatalog, "Database '%s' (ID: %d) created and persisted.", dbName, dbID)
	return dbMeta, nil
}

// DropDatabase removes a database and all its associated tables and indexes from the catalog.
func (pc *PersistentCatalog) DropDatabase(dbName string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	dbID := dbMeta.ID

	// 1. Delete all tables and their indexes associated with this database from storage and cache
	if tablesInDb, ok := pc.tableCache[dbID]; ok {
		for _, tableSchema := range tablesInDb {
			// Delete indexes for this table
			indexPrefix := encoding.EncodeIndexMetadataPrefix(dbID, tableSchema.ID)
			indexKeysToDelete, err := txn.ListKeysWithPrefix(indexPrefix)
			if err != nil {
				return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
					fmt.Sprintf("failed to list index keys for table %s (ID %d)", tableSchema.TableName, tableSchema.ID), err)
			}
			for _, idxKey := range indexKeysToDelete {
				if err = txn.Delete(idxKey); err != nil {
					return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
						fmt.Sprintf("failed to delete index metadata for table %s (ID %d)", tableSchema.TableName, tableSchema.ID), err)
				}
			}
			delete(pc.indexCache, tableSchema.ID) // Remove from cache

			// Delete table metadata
			tableKey := encoding.EncodeTableMetadataKey(dbID, tableSchema.ID)
			if err = txn.Delete(tableKey); err != nil {
				return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
					fmt.Sprintf("failed to delete table metadata for '%s'", tableSchema.TableName), err)
			}
		}
	}
	delete(pc.tableCache, dbID) // Remove table map from cache

	// 2. Delete database metadata from storage and cache
	dbKey := encoding.EncodeDatabaseMetadataKey(dbID)
	if err = txn.Delete(dbKey); err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
			fmt.Sprintf("failed to delete database metadata for '%s'", dbName), err)
	}
	delete(pc.databaseCache, dbName)

	log.Infof(enum.ComponentCatalog, "Database '%s' (ID: %d) and all its contents dropped and persisted.", dbName, dbID)
	return nil
}

// GetDatabase retrieves metadata for a specific database.
func (pc *PersistentCatalog) GetDatabase(dbName string) (*catalog.DatabaseMetadata, error) {
	pc.mu.RLock()
	dbMeta, exists := pc.databaseCache[dbName]
	pc.mu.RUnlock()

	if exists {
		return dbMeta, nil
	}

	// In a production persistent catalog, if not in cache, one might attempt to read from storage
	// directly, or implement a more sophisticated cache refresh/invalidation strategy.
	// For this exercise, we assume the cache is the source of truth after initialization.
	return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
		fmt.Sprintf("database '%s' not found in catalog cache", dbName), nil)
}

// ListDatabases returns a list of all database names in the catalog.
func (pc *PersistentCatalog) ListDatabases() ([]string, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dbNames := make([]string, 0, len(pc.databaseCache))
	for name := range pc.databaseCache {
		dbNames = append(dbNames, name)
	}
	return dbNames, nil
}

// --- Table Operations ---

// CreateTable adds a new table schema to the specified database.
func (pc *PersistentCatalog) CreateTable(dbName string, tableSchema *catalog.TableSchema) (*catalog.TableSchema, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		// This should not happen if pc.tableCache[dbID] is initialized on db creation
		tablesInDb = make(map[string]*catalog.TableSchema)
		pc.tableCache[dbMeta.ID] = tablesInDb
	}

	if _, exists := tablesInDb[tableSchema.TableName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableAlreadyExists,
			fmt.Sprintf("table '%s' already exists in database '%s'", tableSchema.TableName, dbName), nil)
	}

	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	// Assign unique IDs to the table and its columns
	tableSchema.ID = pc.GenerateTableID()
	for _, colDef := range tableSchema.Columns {
		colDef.ID = pc.GenerateColumnID()
	}

	// Persist to storage
	key := encoding.EncodeTableMetadataKey(dbMeta.ID, tableSchema.ID)
	val, err := encoding.EncodeTableSchemaValue(tableSchema)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode table schema for '%s'", tableSchema.TableName), err)
	}
	if err = txn.Write(key, val); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist table '%s' metadata", tableSchema.TableName), err)
	}

	// Update in-memory cache
	tablesInDb[tableSchema.TableName] = tableSchema
	pc.indexCache[tableSchema.ID] = make(map[string]*catalog.IndexSchema)

	log.Infof(enum.ComponentCatalog, "Table '%s' (ID: %d) created and persisted in database '%s'.",
		tableSchema.TableName, tableSchema.ID, dbName)
	return tableSchema, nil
}

// DropTable removes a table and its associated indexes from the specified database.
func (pc *PersistentCatalog) DropTable(dbName, tableName string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	// 1. Delete all indexes associated with this table from storage and cache
	indexPrefix := encoding.EncodeIndexMetadataPrefix(dbMeta.ID, tableSchema.ID)
	indexKeysToDelete, err := txn.ListKeysWithPrefix(indexPrefix)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to list index keys for table %s (ID %d)", tableSchema.TableName, tableSchema.ID), err)
	}
	for _, idxKey := range indexKeysToDelete {
		if err = txn.Delete(idxKey); err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
				fmt.Sprintf("failed to delete index metadata during table drop for table %s (ID %d)", tableName, tableSchema.ID), err)
		}
	}
	delete(pc.indexCache, tableSchema.ID) // Remove from cache

	// 2. Delete table metadata from storage and cache
	tableKey := encoding.EncodeTableMetadataKey(dbMeta.ID, tableSchema.ID)
	if err = txn.Delete(tableKey); err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
			fmt.Sprintf("failed to delete table metadata for '%s'", tableName), err)
	}
	delete(tablesInDb, tableName)

	log.Infof(enum.ComponentCatalog, "Table '%s' (ID: %d) dropped and persisted from database '%s'.", tableName, tableSchema.ID, dbName)
	return nil
}

// GetTable retrieves the schema for a specific table within a database.
func (pc *PersistentCatalog) GetTable(dbName, tableName string) (*catalog.TableSchema, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
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
func (pc *PersistentCatalog) ListTables(dbName string) ([]string, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
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
func (pc *PersistentCatalog) CreateIndex(dbName, tableName string, indexSchema *catalog.IndexSchema) (*catalog.IndexSchema, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := pc.indexCache[tableSchema.ID]
	if indexesInTable == nil {
		indexesInTable = make(map[string]*catalog.IndexSchema)
		pc.indexCache[tableSchema.ID] = indexesInTable
	}

	if _, exists := indexesInTable[indexSchema.IndexName]; exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexAlreadyExists,
			fmt.Sprintf("index '%s' already exists on table '%s'", indexSchema.IndexName, tableName), nil)
	}

	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	// Assign unique ID to the index
	indexSchema.ID = pc.GenerateIndexID()
	indexSchema.TableID = tableSchema.ID // Ensure index links to the correct table ID

	// Persist to storage
	key := encoding.EncodeIndexMetadataKey(dbMeta.ID, tableSchema.ID, indexSchema.ID)
	val, err := encoding.EncodeIndexSchemaValue(indexSchema)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode index schema for '%s'", indexSchema.IndexName), err)
	}
	if err = txn.Write(key, val); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist index '%s' metadata", indexSchema.IndexName), err)
	}

	// Update in-memory cache
	indexesInTable[indexSchema.IndexName] = indexSchema

	log.Infof(enum.ComponentCatalog, "Index '%s' (ID: %d) created and persisted on table '%s' (ID: %d) in database '%s'.",
		indexSchema.IndexName, indexSchema.ID, tableName, tableSchema.ID, dbName)
	return indexSchema, nil
}

// DropIndex removes an index from the specified table.
func (pc *PersistentCatalog) DropIndex(dbName, tableName, indexName string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := pc.indexCache[tableSchema.ID]
	if indexesInTable == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	indexSchema, exists := indexesInTable[indexName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	txn, err := pc.storageAdapter.BeginTransaction(pc.systemDBName, enum.Serializable)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	// Delete from storage
	key := encoding.EncodeIndexMetadataKey(dbMeta.ID, tableSchema.ID, indexSchema.ID)
	if err = txn.Delete(key); err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
			fmt.Sprintf("failed to delete index metadata for '%s'", indexName), err)
	}

	// Delete from cache
	delete(indexesInTable, indexName)

	log.Infof(enum.ComponentCatalog, "Index '%s' (ID: %d) dropped and persisted from table '%s'.", indexName, indexSchema.ID, tableName)
	return nil
}

// GetIndex retrieves the schema for a specific index.
func (pc *PersistentCatalog) GetIndex(dbName, tableName, indexName string) (*catalog.IndexSchema, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := pc.indexCache[tableSchema.ID]
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
func (pc *PersistentCatalog) ListIndexes(dbName, tableName string) ([]string, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dbMeta, exists := pc.databaseCache[dbName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	tablesInDb := pc.tableCache[dbMeta.ID]
	if tablesInDb == nil {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	tableSchema, exists := tablesInDb[tableName]
	if !exists {
		return nil, errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := pc.indexCache[tableSchema.ID]
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
// It atomically increments the counter and persists the new value.
func (pc *PersistentCatalog) GenerateDatabaseID() interfaces.DatabaseID {
	newID := interfaces.DatabaseID(atomic.AddUint64(&pc.nextDatabaseID, 1))
	// Persist the new ID immediately. In a real system, this might be batched or
	// part of the transaction for the actual metadata creation. For now, direct write.
	if err := pc.updateSequence("db", uint64(newID)); err != nil {
		log.Errorf(enum.ComponentCatalog, "Failed to persist new database ID: %v", err)
	}
	return newID
}

// GenerateTableID generates a unique ID for a new table within a database.
func (pc *PersistentCatalog) GenerateTableID() interfaces.TableID {
	newID := interfaces.TableID(atomic.AddUint64(&pc.nextTableID, 1))
	if err := pc.updateSequence("table", uint64(newID)); err != nil {
		log.Errorf(enum.ComponentCatalog, "Failed to persist new table ID: %v", err)
	}
	return newID
}

// GenerateColumnID generates a unique ID for a new column within a table.
func (pc *PersistentCatalog) GenerateColumnID() interfaces.ColumnID {
	newID := interfaces.ColumnID(atomic.AddUint64(&pc.nextColumnID, 1))
	if err := pc.updateSequence("col", uint64(newID)); err != nil {
		log.Errorf(enum.ComponentCatalog, "Failed to persist new column ID: %v", err)
	}
	return newID
}

// GenerateIndexID generates a unique ID for a new index within a table.
func (pc *PersistentCatalog) GenerateIndexID() interfaces.IndexID {
	newID := interfaces.IndexID(atomic.AddUint64(&pc.nextIndexID, 1))
	if err := pc.updateSequence("idx", uint64(newID)); err != nil {
		log.Errorf(enum.ComponentCatalog, "Failed to persist new index ID: %v", err)
	}
	return newID
}

// countTotalTables is a helper for logging.
func countTotalTables(cache map[interfaces.DatabaseID]map[string]*catalog.TableSchema) int {
	count := 0
	for _, tables := range cache {
		count += len(tables)
	}
	return count
}

// countTotalIndexes is a helper for logging.
func countTotalIndexes(cache map[interfaces.TableID]map[string]*catalog.IndexSchema) int {
	count := 0
	for _, indexes := range cache {
		count += len(indexes)
	}
	return count
}

//Personal.AI order the ending
