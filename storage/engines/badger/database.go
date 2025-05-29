// Package badger implements the BadgerDB specific database-level operations for Guocedb.
// This file is responsible for managing the lifecycle of BadgerDB instances, such as
// creating, deleting, opening, and closing databases. It may also maintain database-level
// metadata if not managed entirely within the catalog. It relies on common/errors
// for error handling and common/log for logging. storage/engines/badger/badger.go
// will use this file to manage the underlying BadgerDB databases.
package badger

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger/v4" // Import BadgerDB client library

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces
)

// ensure that BadgerDatabase implements the interfaces.Database interface.
var _ interfaces.Database = (*BadgerDatabase)(nil)

// BadgerDatabase represents a logical database managed by the BadgerDB engine.
// It encapsulates a BadgerDB instance and provides methods for table management.
type BadgerDatabase struct {
	name   string
	path   string        // File system path to the BadgerDB database
	db     *badger.DB    // The underlying BadgerDB instance
	engine *BadgerEngine // Reference back to the parent engine
	mu     sync.RWMutex  // Protects access to the db instance and its state

	// Optional: You might cache table schemas here if they are loaded from BadgerDB
	// map[tableName] *interfaces.TableSchema
	// tableSchemaCache map[string]*interfaces.TableSchema
}

// NewBadgerDatabase creates a new BadgerDatabase instance.
// This function is typically called by BadgerEngine to manage a database.
func NewBadgerDatabase(name, path string, engine *BadgerEngine) *BadgerDatabase {
	return &BadgerDatabase{
		name:   name,
		path:   path,
		engine: engine,
		// tableSchemaCache: make(map[string]*interfaces.TableSchema),
	}
}

// Open opens the underlying BadgerDB instance for this database.
// This should be called before any operations on the database.
func (bd *BadgerDatabase) Open() error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	if bd.db != nil {
		log.Warnf(enum.ComponentStorage, "BadgerDB database '%s' is already open.", bd.name)
		return nil // Already open
	}

	opts := badger.DefaultOptions(bd.path)
	// Configure BadgerDB options, e.g., memory usage, value log settings
	// opts.WithLogger(badgerLoggerAdapter{}) // Optional: Integrate Guocedb's logger with Badger's
	opts.WithLoggingLevel(badger.WARNING) // Adjust Badger's logging level if needed

	db, err := badger.Open(opts)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
			fmt.Sprintf("failed to open BadgerDB for database '%s' at path '%s'", bd.name, bd.path), err)
	}
	bd.db = db
	log.Infof(enum.ComponentStorage, "BadgerDB database '%s' opened successfully at '%s'.", bd.name, bd.path)
	return nil
}

// Close closes the underlying BadgerDB instance for this database.
// This should be called when the database is no longer needed.
func (bd *BadgerDatabase) Close() error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	if bd.db == nil {
		log.Warnf(enum.ComponentStorage, "BadgerDB database '%s' is already closed or not open.", bd.name)
		return nil // Already closed or never opened
	}

	err := bd.db.Close()
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageShutdownFailed,
			fmt.Sprintf("failed to close BadgerDB for database '%s'", bd.name), err)
	}
	bd.db = nil // Clear the reference
	log.Infof(enum.ComponentStorage, "BadgerDB database '%s' closed successfully.", bd.name)
	return nil
}

// Name returns the name of the database.
func (bd *BadgerDatabase) Name() string {
	return bd.name
}

// CreateTable creates a new table within this database by storing its schema.
// This involves writing the table metadata to BadgerDB.
func (bd *BadgerDatabase) CreateTable(tableName string, schema *interfaces.TableSchema) error {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	if bd.db == nil {
		return errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			fmt.Sprintf("BadgerDB for database '%s' is not open", bd.name), nil)
	}

	// Begin a transaction to store the table schema
	txn := bd.db.NewTransaction(true) // Read-write transaction
	defer txn.Discard()

	// Check if table already exists (optional, could be handled by catalog)
	// For simplicity, we assume the catalog layer will handle existence checks.

	// Generate a unique TableID. This could be a sequence managed by Badger or catalog.
	// For now, let's derive it from the table name (simple for example, but not robust)
	// A robust system would use a global unique ID generator or a sequence.
	// Let's use a dummy ID for now and assume it comes from catalog later.
	tableID := interfaces.ID(hashString(tableName)) // Dummy ID generation

	// Store the TableSchema using its designated key and encoding
	tableSchemaBytes, err := EncodeTableSchemaValue(schema)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode schema for table '%s'", tableName), err)
	}

	// Key: TableMetadataPrefix + DatabaseIDBytes + TableNameBytes
	// Assuming each database has a unique ID handled by the engine, let's use a dummy ID here.
	// In a real system, bd.engine would provide bd.engine.GetDatabaseID(bd.name)
	dbID := interfaces.ID(hashString(bd.name)) // Dummy DB ID

	key := EncodeTableMetadataKey(dbID, tableName)

	err = txn.Set(key, tableSchemaBytes)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to store schema for table '%s' in BadgerDB", tableName), err)
	}

	err = txn.Commit()
	if err != nil {
		return errors.NewGuocedbError(enum.ErrTransaction, errors.CodeTransactionCommitFailed,
			fmt.Sprintf("failed to commit transaction for creating table '%s'", tableName), err)
	}

	log.Infof(enum.ComponentStorage, "Table '%s' created successfully in database '%s'.", tableName, bd.name)
	return nil
}

// DropTable removes an existing table from this database.
// This involves deleting its schema and all associated row data.
func (bd *BadgerDatabase) DropTable(tableName string) error {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	if bd.db == nil {
		return errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			fmt.Sprintf("BadgerDB for database '%s' is not open", bd.name), nil)
	}

	txn := bd.db.NewTransaction(true)
	defer txn.Discard()

	// Dummy DB ID
	dbID := interfaces.ID(hashString(bd.name))

	// Delete table metadata key
	tableMetaKey := EncodeTableMetadataKey(dbID, tableName)
	err := txn.Delete(tableMetaKey)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
			fmt.Sprintf("failed to delete metadata for table '%s'", tableName), err)
	}

	// TODO: Efficiently delete all row data associated with this table.
	// This would typically involve iterating over keys with the table's prefix
	// and deleting them. For very large tables, this might need batching or
	// a specific BadgerDB feature for prefix deletion.
	// Example (conceptual):
	// rowPrefix := EncodeRowKey(dbID, tableID, 0) // Need to get tableID first
	// it := txn.NewIterator(badger.IteratorOptions{Prefix: rowPrefix})
	// for it.Rewind(); it.ValidForPrefix(rowPrefix); it.Next() {
	//     err = txn.Delete(it.Item().Key())
	//     if err != nil { ... }
	// }
	// it.Close()

	err = txn.Commit()
	if err != nil {
		return errors.NewGuocedbError(enum.ErrTransaction, errors.CodeTransactionCommitFailed,
			fmt.Sprintf("failed to commit transaction for dropping table '%s'", tableName), err)
	}

	log.Infof(enum.ComponentStorage, "Table '%s' dropped successfully from database '%s'.", tableName, bd.name)
	return nil
}

// GetTable returns a Table interface for the given table name.
// This involves loading the table's schema from BadgerDB.
func (bd *BadgerDatabase) GetTable(tableName string) (interfaces.Table, error) {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	if bd.db == nil {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			fmt.Sprintf("BadgerDB for database '%s' is not open", bd.name), nil)
	}

	txn := bd.db.NewTransaction(false) // Read-only transaction
	defer txn.Discard()

	// Dummy DB ID
	dbID := interfaces.ID(hashString(bd.name))

	tableMetaKey := EncodeTableMetadataKey(dbID, tableName)
	item, err := txn.Get(tableMetaKey)
	if err == badger.ErrKeyNotFound {
		return nil, errors.NewGuocedbError(enum.ErrNotFound, errors.CodeTableNotFound,
			fmt.Sprintf("table '%s' not found in database '%s'", tableName, bd.name), nil)
	}
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to retrieve metadata for table '%s'", tableName), err)
	}

	tableSchemaBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to get value for table metadata key '%s'", tableName), err)
	}

	schema, err := DecodeTableSchemaValue(tableSchemaBytes)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode schema for table '%s'", tableName), err)
	}

	// In a real system, we'd also get a consistent TableID (from catalog or sequence)
	// For now, we'll generate a dummy one consistent with creation
	tableID := interfaces.ID(hashString(tableName))

	return NewBadgerTable(bd.db, dbID, tableID, schema), nil
}

// ListTables returns a list of all table names within this database.
func (bd *BadgerDatabase) ListTables() ([]string, error) {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	if bd.db == nil {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			fmt.Sprintf("BadgerDB for database '%s' is not open", bd.name), nil)
	}

	var tableNames []string
	txn := bd.db.NewTransaction(false) // Read-only transaction
	defer txn.Discard()

	// Dummy DB ID
	dbIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dbIDBytes, uint64(hashString(bd.name)))

	// Iterate over keys with the TableMetadataPrefix + dbIDBytes
	prefix := EncodeKey(TableMetadataPrefix, dbIDBytes)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: prefix})
	defer it.Close()

	for it.Rewind(); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := item.Key()
		// The table name is the part after the prefix (1 byte) and dbID (8 bytes)
		_, tableName, err := DecodeTableMetadataKey(key)
		if err != nil {
			log.Warnf(enum.ComponentStorage, "Failed to decode table metadata key '%x': %v", key, err)
			continue // Skip malformed keys
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

// GetBadgerDB returns the underlying BadgerDB instance.
// This is an internal helper for other components within the 'badger' package.
func (bd *BadgerDatabase) GetBadgerDB() *badger.DB {
	bd.mu.RLock()
	defer bd.mu.RUnlock()
	return bd.db
}

// hashString is a simple non-cryptographic hash for demonstration purposes.
// In a real system, unique IDs would be generated more robustly (e.g., UUIDs, atomic counters).
func hashString(s string) uint64 {
	var h uint64 = 14695981039346656037 // FNV offset basis
	const prime uint64 = 1099511628211  // FNV prime
	for i := 0; i < len(s); i++ {
		h = h * prime
		h = h ^ uint64(s[i])
	}
	return h
}

//Personal.AI order the ending
