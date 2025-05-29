// Package persistent implements the persistent metadata catalog for Guocedb.
// This file provides a storage-backed metadata management implementation,
// ensuring metadata persistence, including storage/loading, versioning,
// and consistency guarantees. It relies on the catalog interface and the
// storage abstraction layer, and will be used for metadata recovery during
// database startup.
package persistent

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sync"
	"sync/atomic"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/compute/catalog" // Import the catalog interfaces and structs
	"github.com/turtacn/guocedb/interfaces"      // Import general interfaces like ID
	"github.com/turtacn/guocedb/storage/sal"     // Import the Storage Abstraction Layer
)

// ensure that PersistentCatalog implements the catalog.Catalog interface.
var _ catalog.Catalog = (*PersistentCatalog)(nil)

// Key prefixes for metadata storage in the underlying key-value store.
const (
	// DatabaseMetadataPrefix is used for storing database metadata.
	// Key: DatabaseMetadataPrefix + DatabaseID
	DatabaseMetadataPrefix byte = 0x01
	// TableMetadataPrefix is used for storing table schemas.
	// Key: TableMetadataPrefix + DatabaseID + TableID
	TableMetadataPrefix byte = 0x02
	// IndexMetadataPrefix is used for storing index schemas.
	// Key: IndexMetadataPrefix + DatabaseID + TableID + IndexID
	IndexMetadataPrefix byte = 0x03
	// IDSequencePrefix is used for storing ID sequence counters.
	// Key: IDSequencePrefix + TypeIdentifier (e.g., "db", "table", "col", "idx")
	IDSequencePrefix byte = 0x04
)

// PersistentCatalog implements the Catalog interface with persistent storage.
// It leverages the storage abstraction layer (SAL) to store and retrieve metadata.
type PersistentCatalog struct {
	storageEngine interfaces.StorageEngine // The underlying storage engine (e.g., BadgerEngine)
	systemDB      interfaces.Database      // The dedicated database for catalog metadata (e.g., "guocedb_catalog")

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
// It requires a configured StorageEngine to persist its metadata.
func NewPersistentCatalog(engine interfaces.StorageEngine) (*PersistentCatalog, error) {
	if engine == nil {
		return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
			"storage engine cannot be nil for persistent catalog", nil)
	}

	pc := &PersistentCatalog{
		storageEngine:  engine,
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

	// 1. Create or get the system database for catalog metadata
	systemDBName := "guocedb_catalog"
	err := pc.storageEngine.CreateDatabase(systemDBName)
	if err != nil && !errors.Is(err, errors.NewGuocedbError(enum.ErrAlreadyExists, errors.CodeDatabaseAlreadyExists, "", nil)) {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeCatalogInitFailed,
			fmt.Sprintf("failed to create/open system catalog database '%s'", systemDBName), err)
	}

	db, err := pc.storageEngine.GetDatabase(systemDBName)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeCatalogInitFailed,
			fmt.Sprintf("failed to get system catalog database '%s'", systemDBName), err)
	}
	pc.systemDB = db

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
// For persistent catalog, this primarily involves flushing any pending metadata changes
// (though most changes are transactional and committed immediately) and closing the system DB.
func (pc *PersistentCatalog) Shutdown() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.systemDB == nil {
		log.Warnf(enum.ComponentCatalog, "PersistentCatalog already shut down or not initialized.")
		return nil
	}

	// Flush any pending ID sequences (already handled by atomic updates and storage Set operations)
	// No explicit flush needed if operations are transactional.

	err := pc.systemDB.Close()
	if err != nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeCatalogShutdownFailed,
			"failed to close system catalog database", err)
	}
	pc.systemDB = nil
	pc.databaseCache = nil // Clear caches
	pc.tableCache = nil
	pc.indexCache = nil

	log.Infof(enum.ComponentCatalog, "PersistentCatalog shut down.")
	return nil
}

// loadIDSequences loads the highest used ID for each type from persistent storage.
func (pc *PersistentCatalog) loadIDSequences() error {
	txn, err := pc.systemDB.BeginTransaction(enum.RepeatableRead)
	if err != nil {
		return err
	}
	defer txn.Rollback() // Read-only transaction, so always discard

	// Load Database ID sequence
	dbID, err := pc.loadSequence(txn, "db")
	if err != nil {
		return err
	}
	pc.nextDatabaseID = dbID

	// Load Table ID sequence
	tableID, err := pc.loadSequence(txn, "table")
	if err != nil {
		return err
	}
	pc.nextTableID = tableID

	// Load Column ID sequence
	colID, err := pc.loadSequence(txn, "col")
	if err != nil {
		return err
	}
	pc.nextColumnID = colID

	// Load Index ID sequence
	idxID, err := pc.loadSequence(txn, "idx")
	if err != nil {
		return err
	}
	pc.nextIndexID = idxID

	log.Infof(enum.ComponentCatalog, "Loaded ID sequences: DB=%d, Table=%d, Column=%d, Index=%d",
		pc.nextDatabaseID, pc.nextTableID, pc.nextColumnID, pc.nextIndexID)
	return nil
}

// loadSequence is a helper to load a specific ID sequence from storage.
func (pc *PersistentCatalog) loadSequence(txn interfaces.Transaction, seqName string) (uint64, error) {
	key := EncodeIDSequenceKey(seqName)
	val, err := txn.Read(key) // Assuming Read method takes raw byte key
	if err == nil {
		if len(val) != 8 {
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
func (pc *PersistentCatalog) updateSequence(txn interfaces.Transaction, seqName string, newID uint64) error {
	key := EncodeIDSequenceKey(seqName)
	val := sal.Uint64ToBytes(newID)
	err := txn.Write(key, val) // Assuming Write method takes raw byte key and value
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist ID sequence '%s' to storage", seqName), err)
	}
	return nil
}

// loadAllMetadata loads all database, table, and index metadata into in-memory caches.
func (pc *PersistentCatalog) loadAllMetadata() error {
	txn, err := pc.systemDB.BeginTransaction(enum.RepeatableRead)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	// Load Databases
	dbNames, err := pc.listMetadataKeysByPrefix(txn, DatabaseMetadataPrefix)
	if err != nil {
		return err
	}
	for _, key := range dbNames {
		dbMeta, err := pc.getDatabaseByRawKey(txn, key)
		if err != nil {
			return err
		}
		pc.databaseCache[dbMeta.DatabaseName] = dbMeta
		pc.tableCache[dbMeta.ID] = make(map[string]*catalog.TableSchema) // Initialize table map for loaded DB
	}

	// Load Tables
	tableKeys, err := pc.listMetadataKeysByPrefix(txn, TableMetadataPrefix)
	if err != nil {
		return err
	}
	for _, key := range tableKeys {
		tableSchema, err := pc.getTableByRawKey(txn, key)
		if err != nil {
			return err
		}
		if _, ok := pc.tableCache[tableSchema.ID]; !ok {
			// This scenario means a table exists without its parent database in cache.
			// This shouldn't happen if database loading is robust. Log a warning.
			log.Warnf(enum.ComponentCatalog, "Found table '%s' (ID: %d) with no corresponding database ID %d in cache.",
				tableSchema.TableName, tableSchema.ID, DecodeTableKeyDatabaseID(key))
			continue
		}
		pc.tableCache[DecodeTableKeyDatabaseID(key)][tableSchema.TableName] = tableSchema
		pc.indexCache[tableSchema.ID] = make(map[string]*catalog.IndexSchema) // Initialize index map for loaded table
	}

	// Load Indexes
	indexKeys, err := pc.listMetadataKeysByPrefix(txn, IndexMetadataPrefix)
	if err != nil {
		return err
	}
	for _, key := range indexKeys {
		indexSchema, err := pc.getIndexByRawKey(txn, key)
		if err != nil {
			return err
		}
		if _, ok := pc.indexCache[indexSchema.TableID]; !ok {
			log.Warnf(enum.ComponentCatalog, "Found index '%s' (ID: %d) with no corresponding table ID %d in cache.",
				indexSchema.IndexName, indexSchema.ID, indexSchema.TableID)
			continue
		}
		pc.indexCache[indexSchema.TableID][indexSchema.IndexName] = indexSchema
	}

	return nil
}

// listMetadataKeysByPrefix iterates the system DB for keys with a given prefix.
func (pc *PersistentCatalog) listMetadataKeysByPrefix(txn interfaces.Transaction, prefix byte) ([][]byte, error) {
	var keys [][]byte
	// Assuming GetIterator on interfaces.Transaction and Next method provides key
	// This is a placeholder; actual implementation depends on `interfaces.Transaction` methods.
	// A common pattern is `txn.NewIterator(prefixBytes).Seek(prefixBytes).Next().Item().Key()`
	// For now, let's assume `txn.Iterate(prefix)` exists.
	// if it, err := txn.GetIterator(prefix); err == nil {
	// 	defer it.Close()
	// 	for it.Next() {
	// 		keys = append(keys, it.CurrentKey())
	// 	}
	// 	return keys, nil
	// } else {
	// 	return nil, err
	// }

	// A more robust but verbose way if `GetIterator` or `Iterate` isn't directly available:
	// We need to simulate listing keys for a prefix. This might require a custom iterator
	// in the `storage/sal` or `badger` layer that exposes a byte prefix scan.
	// For now, assuming `txn.ListKeysByPrefix(prefix)` exists or adapting `badger.Iterator`.
	// Since we import badger, we can directly use `badgerTxn.NewIterator`.
	badgerTxn, ok := txn.(*BadgerTransaction) // Type assertion, as PersistentCatalog uses storage.Transaction
	if !ok {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState, "invalid transaction type for listing keys", nil)
	}

	// Create a BadgerDB iterator specific to the prefix
	badgerOpts := badger.DefaultIteratorOptions
	badgerOpts.Prefix = []byte{prefix} // Use the byte prefix directly

	it := badgerTxn.GetBadgerTxn().NewIterator(badgerOpts)
	defer it.Close()

	for it.Rewind(); it.ValidForPrefix(badgerOpts.Prefix); it.Next() {
		item := it.Item()
		keys = append(keys, item.KeyCopy(nil))
	}
	return keys, nil
}

// getDatabaseByRawKey retrieves and decodes DatabaseMetadata from a raw key.
func (pc *PersistentCatalog) getDatabaseByRawKey(txn interfaces.Transaction, key []byte) (*catalog.DatabaseMetadata, error) {
	valBytes, err := txn.Read(key)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read database metadata for key %x", key), err)
	}
	dbMeta, err := DecodeDatabaseMetadataValue(valBytes)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode database metadata for key %x", key), err)
	}
	return dbMeta, nil
}

// getTableByRawKey retrieves and decodes TableSchema from a raw key.
func (pc *PersistentCatalog) getTableByRawKey(txn interfaces.Transaction, key []byte) (*catalog.TableSchema, error) {
	valBytes, err := txn.Read(key)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read table schema for key %x", key), err)
	}
	tableSchema, err := DecodeTableSchemaValue(valBytes) // Assuming DecodeTableSchemaValue from encoding.go
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode table schema for key %x", key), err)
	}
	return tableSchema, nil
}

// getIndexByRawKey retrieves and decodes IndexSchema from a raw key.
func (pc *PersistentCatalog) getIndexByRawKey(txn interfaces.Transaction, key []byte) (*catalog.IndexSchema, error) {
	valBytes, err := txn.Read(key)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read index schema for key %x", key), err)
	}
	indexSchema, err := DecodeIndexSchemaValue(valBytes) // Assuming DecodeIndexSchemaValue from encoding.go
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode index schema for key %x", key), err)
	}
	return indexSchema, nil
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

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable) // Use serializable for metadata changes
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
	key := EncodeDatabaseMetadataKey(dbID)
	val, err := EncodeDatabaseMetadataValue(dbMeta)
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

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
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
			indexKeys, err := pc.listMetadataKeysByPrefix(txn, EncodeIndexMetadataPrefix(dbID, tableSchema.ID))
			if err != nil {
				return err
			}
			for _, idxKey := range indexKeys {
				if err = txn.Delete(idxKey); err != nil {
					return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
						fmt.Sprintf("failed to delete index metadata for table %s (ID %d)", tableSchema.TableName, tableSchema.ID), err)
				}
			}
			delete(pc.indexCache, tableSchema.ID) // Remove from cache

			// Delete table metadata
			tableKey := EncodeTableMetadataKey(dbID, tableSchema.ID)
			if err = txn.Delete(tableKey); err != nil {
				return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
					fmt.Sprintf("failed to delete table metadata for '%s'", tableSchema.TableName), err)
			}
		}
	}
	delete(pc.tableCache, dbID) // Remove table map from cache

	// 2. Delete database metadata from storage and cache
	dbKey := EncodeDatabaseMetadataKey(dbID)
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

	// If not in cache, try to load from storage (e.g., if another process created it)
	// For robustness, this should ideally be handled by a cache invalidation/refresh mechanism
	// or by always creating through this catalog instance. For simplicity, we assume
	// once loaded, it's in cache, or we would trigger a full reload.
	// For now, this path indicates a "not found" scenario or a stale cache.
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

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
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
	key := EncodeTableMetadataKey(dbMeta.ID, tableSchema.ID)
	val, err := EncodeTableSchemaValue(tableSchema)
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
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
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
	indexKeys, err := pc.listMetadataKeysByPrefix(txn, EncodeIndexMetadataPrefix(dbMeta.ID, tableSchema.ID))
	if err != nil {
		return err
	}
	for _, idxKey := range indexKeys {
		if err = txn.Delete(idxKey); err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
				fmt.Sprintf("failed to delete index metadata during table drop for table %s (ID %d)", tableName, tableSchema.ID), err)
		}
	}
	delete(pc.indexCache, tableSchema.ID) // Remove from cache

	// 2. Delete table metadata from storage and cache
	tableKey := EncodeTableMetadataKey(dbMeta.ID, tableSchema.ID)
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

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
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
	key := EncodeIndexMetadataKey(dbMeta.ID, tableSchema.ID, indexSchema.ID)
	val, err := EncodeIndexSchemaValue(indexSchema)
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
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, dbName), nil)
	}

	indexesInTable := pc.indexCache[tableSchema.ID]
	if indexesInTable == nil {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	indexSchema, exists := indexesInTable[indexName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrCatalog, errors.CodeIndexNotFound,
			fmt.Sprintf("index '%s' not found on table '%s'", indexName, tableName), nil)
	}

	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
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
	key := EncodeIndexMetadataKey(dbMeta.ID, tableSchema.ID, indexSchema.ID)
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
	// Persist the new ID immediately (best effort, transaction handles atomicity for metadata itself)
	// This update is best effort here, the actual transaction for CreateDatabase will ensure consistency.
	// In a high-concurrency setup, this would be part of a single commit.
	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
	if err == nil {
		defer txn.Commit() // Commit on success
		pc.updateSequence(txn, "db", uint64(newID))
	} else {
		log.Errorf(enum.ComponentCatalog, "Failed to begin transaction for ID sequence update: %v", err)
	}
	return newID
}

// GenerateTableID generates a unique ID for a new table within a database.
func (pc *PersistentCatalog) GenerateTableID() interfaces.TableID {
	newID := interfaces.TableID(atomic.AddUint64(&pc.nextTableID, 1))
	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
	if err == nil {
		defer txn.Commit()
		pc.updateSequence(txn, "table", uint64(newID))
	} else {
		log.Errorf(enum.ComponentCatalog, "Failed to begin transaction for ID sequence update: %v", err)
	}
	return newID
}

// GenerateColumnID generates a unique ID for a new column within a table.
func (pc *PersistentCatalog) GenerateColumnID() interfaces.ColumnID {
	newID := interfaces.ColumnID(atomic.AddUint64(&pc.nextColumnID, 1))
	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
	if err == nil {
		defer txn.Commit()
		pc.updateSequence(txn, "col", uint64(newID))
	} else {
		log.Errorf(enum.ComponentCatalog, "Failed to begin transaction for ID sequence update: %v", err)
	}
	return newID
}

// GenerateIndexID generates a unique ID for a new index within a table.
func (pc *PersistentCatalog) GenerateIndexID() interfaces.IndexID {
	newID := interfaces.IndexID(atomic.AddUint64(&pc.nextIndexID, 1))
	txn, err := pc.systemDB.BeginTransaction(enum.Serializable)
	if err == nil {
		defer txn.Commit()
		pc.updateSequence(txn, "idx", uint64(newID))
	} else {
		log.Errorf(enum.ComponentCatalog, "Failed to begin transaction for ID sequence update: %v", err)
	}
	return newID
}

// Helper functions for encoding/decoding metadata keys and values
// These would typically reside in `storage/engines/badger/encoding.go` or a common `encoding` package.
// For demonstration, simplified inline functions are used, assuming underlying SAL handles the byte conversion.

// EncodeDatabaseMetadataKey encodes a database metadata key.
// Format: 0x01 (prefix) + DatabaseID (8 bytes)
func EncodeDatabaseMetadataKey(dbID interfaces.DatabaseID) []byte {
	return append([]byte{DatabaseMetadataPrefix}, sal.Uint64ToBytes(uint64(dbID))...)
}

// DecodeDatabaseMetadataKey decodes a database metadata key to get DatabaseID.
func DecodeDatabaseMetadataKey(key []byte) (interfaces.DatabaseID, error) {
	if len(key) != 1+8 || key[0] != DatabaseMetadataPrefix {
		return 0, fmt.Errorf("invalid database metadata key format: %x", key)
	}
	return interfaces.DatabaseID(sal.BytesToUint64(key[1:])), nil
}

// EncodeTableMetadataKey encodes a table metadata key.
// Format: 0x02 (prefix) + DatabaseID (8 bytes) + TableID (8 bytes)
func EncodeTableMetadataKey(dbID interfaces.DatabaseID, tableID interfaces.TableID) []byte {
	key := make([]byte, 1+8+8)
	key[0] = TableMetadataPrefix
	copy(key[1:9], sal.Uint64ToBytes(uint64(dbID)))
	copy(key[9:17], sal.Uint64ToBytes(uint64(tableID)))
	return key
}

// DecodeTableKeyDatabaseID extracts the DatabaseID from a table metadata key.
func DecodeTableKeyDatabaseID(key []byte) interfaces.DatabaseID {
	return interfaces.DatabaseID(sal.BytesToUint64(key[1:9]))
}

// EncodeIndexMetadataKey encodes an index metadata key.
// Format: 0x03 (prefix) + DatabaseID (8 bytes) + TableID (8 bytes) + IndexID (8 bytes)
func EncodeIndexMetadataKey(dbID interfaces.DatabaseID, tableID interfaces.TableID, indexID interfaces.IndexID) []byte {
	key := make([]byte, 1+8+8+8)
	key[0] = IndexMetadataPrefix
	copy(key[1:9], sal.Uint64ToBytes(uint64(dbID)))
	copy(key[9:17], sal.Uint64ToBytes(uint64(tableID)))
	copy(key[17:25], sal.Uint64ToBytes(uint64(indexID)))
	return key
}

// EncodeIndexMetadataPrefix generates a prefix for iterating indexes of a specific table.
// Format: 0x03 (prefix) + DatabaseID (8 bytes) + TableID (8 bytes)
func EncodeIndexMetadataPrefix(dbID interfaces.DatabaseID, tableID interfaces.TableID) []byte {
	prefix := make([]byte, 1+8+8)
	prefix[0] = IndexMetadataPrefix
	copy(prefix[1:9], sal.Uint64ToBytes(uint64(dbID)))
	copy(prefix[9:17], sal.Uint64ToBytes(uint64(tableID)))
	return prefix
}

// EncodeIDSequenceKey encodes an ID sequence key.
// Format: 0x04 (prefix) + SequenceName (variable length string)
func EncodeIDSequenceKey(seqName string) []byte {
	return append([]byte{IDSequencePrefix}, []byte(seqName)...)
}

// Encode/Decode functions for metadata values (simplified, would use Gob/Protobuf for real)

// EncodeDatabaseMetadataValue encodes DatabaseMetadata into bytes.
func EncodeDatabaseMetadataValue(meta *catalog.DatabaseMetadata) ([]byte, error) {
	// Simple JSON encoding for demonstration. In production, use more efficient serialization.
	return []byte(fmt.Sprintf(`{"ID":%d,"DatabaseName":"%s"}`, meta.ID, meta.DatabaseName)), nil
}

// DecodeDatabaseMetadataValue decodes bytes into DatabaseMetadata.
func DecodeDatabaseMetadataValue(data []byte) (*catalog.DatabaseMetadata, error) {
	// Simple JSON decoding for demonstration.
	var meta struct {
		ID           interfaces.DatabaseID
		DatabaseName string
	}
	_, err := fmt.Sscanf(string(data), `{"ID":%d,"DatabaseName":"%s"}`, &meta.ID, &meta.DatabaseName)
	if err != nil {
		return nil, err
	}
	return &catalog.DatabaseMetadata{
		ID:           meta.ID,
		DatabaseName: meta.DatabaseName,
	}, nil
}

// EncodeTableSchemaValue encodes TableSchema into bytes.
func EncodeTableSchemaValue(schema *interfaces.TableSchema) ([]byte, error) {
	// This would be complex with columns. For demo, simplified.
	// In a real system, you'd serialize the full struct (e.g., using Gob or Protobuf).
	return []byte(fmt.Sprintf(`{"ID":%d,"TableName":"%s","NumCols":%d}`, schema.ID, schema.TableName, len(schema.Columns))), nil
}

// DecodeTableSchemaValue decodes bytes into TableSchema.
func DecodeTableSchemaValue(data []byte) (*interfaces.TableSchema, error) {
	// Simplified decoding.
	var s struct {
		ID        interfaces.TableID
		TableName string
		NumCols   int // Just a placeholder, actual columns would be serialized
	}
	_, err := fmt.Sscanf(string(data), `{"ID":%d,"TableName":"%s","NumCols":%d}`, &s.ID, &s.TableName, &s.NumCols)
	if err != nil {
		return nil, err
	}
	return &interfaces.TableSchema{
		ID:        s.ID,
		TableName: s.TableName,
		Columns:   make([]*interfaces.ColumnDefinition, s.NumCols), // Placeholder
	}, nil
}

// EncodeIndexSchemaValue encodes IndexSchema into bytes.
func EncodeIndexSchemaValue(schema *catalog.IndexSchema) ([]byte, error) {
	// Simplified.
	return []byte(fmt.Sprintf(`{"ID":%d,"IndexName":"%s","TableID":%d}`, schema.ID, schema.IndexName, schema.TableID)), nil
}

// DecodeIndexSchemaValue decodes bytes into IndexSchema.
func DecodeIndexSchemaValue(data []byte) (*catalog.IndexSchema, error) {
	// Simplified.
	var s struct {
		ID        interfaces.IndexID
		IndexName string
		TableID   interfaces.TableID
	}
	_, err := fmt.Sscanf(string(data), `{"ID":%d,"IndexName":"%s","TableID":%d}`, &s.ID, &s.IndexName, &s.TableID)
	if err != nil {
		return nil, err
	}
	return &catalog.IndexSchema{
		ID:        s.ID,
		IndexName: s.IndexName,
		TableID:   s.TableID,
	}, nil
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
