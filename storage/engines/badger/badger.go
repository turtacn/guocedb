// Package badger implements the main BadgerDB storage engine adapter for Guocedb.
// This file wraps BadgerDB to conform to the storage engine interface defined in
// interfaces/storage.go. It serves as the entry point for the 'badger' sub-package,
// responsible for initializing the BadgerDB instance and coordinating functionalities
// across database.go, table.go, transaction.go, iterator.go, and encoding.go.
// It relies on interfaces/storage.go, common/errors, and common/log.
// storage/sal/adapter.go will interact with the Badger engine through this file.
package badger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic" // For atomic ID generation

	"github.com/dgraph-io/badger/v4" // Import BadgerDB client library

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces
)

// ensure that BadgerEngine implements the interfaces.StorageEngine interface.
var _ interfaces.StorageEngine = (*BadgerEngine)(nil)

// BadgerConfig holds configuration parameters for the BadgerDB engine.
type BadgerConfig struct {
	Path             string // Directory path where BadgerDB will store its data.
	InMemory         bool   // If true, run BadgerDB in in-memory mode (for testing/temp data).
	ValueLogFileSize int64  // Maximum size of a value log file in bytes (e.g., 1GB)
	SyncWrites       bool   // Whether to sync writes to disk (true for durability, false for performance)
	// Add more BadgerDB specific options as needed
}

// DefaultBadgerConfig returns a default configuration for BadgerDB.
func DefaultBadgerConfig(path string) *BadgerConfig {
	return &BadgerConfig{
		Path:             path,
		InMemory:         false,
		ValueLogFileSize: 1 << 30, // 1GB
		SyncWrites:       true,
	}
}

// BadgerEngine implements the interfaces.StorageEngine interface using BadgerDB.
// It manages multiple logical databases, each backed by its own BadgerDB instance.
type BadgerEngine struct {
	config         *BadgerConfig
	databases      map[string]*BadgerDatabase // Map of database name to BadgerDatabase instance
	mu             sync.RWMutex               // Protects access to the databases map
	nextDatabaseID uint64                     // Atomic counter for generating unique database IDs
	closed         atomic.Bool                // True if the engine has been closed
}

// NewBadgerEngine creates a new BadgerEngine instance.
func NewBadgerEngine(config *BadgerConfig) (*BadgerEngine, error) {
	if config == nil {
		return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
			"badger engine configuration cannot be nil", nil)
	}
	if !config.InMemory && config.Path == "" {
		return nil, errors.NewGuocedbError(enum.ErrConfig, errors.CodeInvalidConfig,
			"path cannot be empty for non-in-memory BadgerDB engine", nil)
	}

	return &BadgerEngine{
		config:         config,
		databases:      make(map[string]*BadgerDatabase),
		nextDatabaseID: 0, // Will be initialized on startup or first database creation
		closed:         atomic.Bool{},
	}, nil
}

// Initialize prepares the BadgerEngine for use.
// It ensures the base directory exists and initializes the database ID counter.
func (be *BadgerEngine) Initialize() error {
	if be.config.InMemory {
		log.Infof(enum.ComponentStorage, "BadgerDB engine initializing in in-memory mode.")
	} else {
		log.Infof(enum.ComponentStorage, "BadgerDB engine initializing at base path: %s", be.config.Path)
		if err := os.MkdirAll(be.config.Path, 0755); err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
				fmt.Sprintf("failed to create base directory for BadgerDB: %s", be.config.Path), err)
		}
	}

	// Load highest used database ID to ensure uniqueness, or start from 0 if no databases exist
	// This would typically involve scanning for DatabaseMetadataPrefix keys.
	// For simplicity, we'll start at 0 for now and assume ID conflicts are rare in this example.
	// A more robust system would persist and load this counter.
	be.nextDatabaseID = 0
	be.closed.Store(false)

	log.Infof(enum.ComponentStorage, "BadgerDB engine initialized successfully.")
	return nil
}

// Shutdown gracefully closes all open BadgerDB instances managed by this engine.
func (be *BadgerEngine) Shutdown() error {
	if be.closed.Load() {
		log.Warnf(enum.ComponentStorage, "BadgerDB engine already shut down.")
		return nil
	}

	be.mu.Lock()
	defer be.mu.Unlock()

	var firstErr error
	for dbName, db := range be.databases {
		log.Infof(enum.ComponentStorage, "Closing BadgerDB for database '%s'...", dbName)
		if err := db.Close(); err != nil {
			log.Errorf(enum.ComponentStorage, "Failed to close database '%s': %v", dbName, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	be.databases = make(map[string]*BadgerDatabase) // Clear map
	be.closed.Store(true)
	log.Infof(enum.ComponentStorage, "BadgerDB engine shut down.")
	return firstErr // Return the first error encountered, if any
}

// CreateDatabase creates a new logical database.
// This involves creating a new directory for the BadgerDB instance and opening it.
func (be *BadgerEngine) CreateDatabase(dbName string) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.closed.Load() {
		return errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"badger engine is closed", nil)
	}

	if _, exists := be.databases[dbName]; exists {
		return errors.NewGuocedbError(enum.ErrAlreadyExists, errors.CodeDatabaseAlreadyExists,
			fmt.Sprintf("database '%s' already exists", dbName), nil)
	}

	dbPath := filepath.Join(be.config.Path, dbName)
	if !be.config.InMemory {
		// Create the directory for the new database
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
				fmt.Sprintf("failed to create directory for new database '%s' at '%s'", dbName, dbPath), err)
		}
	}

	newDB := NewBadgerDatabase(dbName, dbPath, be)
	if err := newDB.Open(); err != nil {
		// Clean up directory if open fails
		if !be.config.InMemory {
			os.RemoveAll(dbPath)
		}
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
			fmt.Sprintf("failed to open BadgerDB for new database '%s'", dbName), err)
	}

	be.databases[dbName] = newDB
	log.Infof(enum.ComponentStorage, "Database '%s' created and opened successfully.", dbName)
	return nil
}

// DropDatabase removes an existing logical database.
// This involves closing its BadgerDB instance and deleting its directory.
func (be *BadgerEngine) DropDatabase(dbName string) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.closed.Load() {
		return errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"badger engine is closed", nil)
	}

	db, exists := be.databases[dbName]
	if !exists {
		return errors.NewGuocedbError(enum.ErrNotFound, errors.CodeDatabaseNotFound,
			fmt.Sprintf("database '%s' not found", dbName), nil)
	}

	if err := db.Close(); err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageShutdownFailed,
			fmt.Sprintf("failed to close database '%s' before dropping", dbName), err)
	}

	if !be.config.InMemory {
		dbPath := filepath.Join(be.config.Path, dbName)
		if err := os.RemoveAll(dbPath); err != nil {
			return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
				fmt.Sprintf("failed to delete database directory '%s' for database '%s'", dbPath, dbName), err)
		}
	}

	delete(be.databases, dbName)
	log.Infof(enum.ComponentStorage, "Database '%s' dropped successfully.", dbName)
	return nil
}

// GetDatabase returns a Database interface for the given database name.
// If the database is not open, it will attempt to open it.
func (be *BadgerEngine) GetDatabase(dbName string) (interfaces.Database, error) {
	be.mu.RLock()
	db, exists := be.databases[dbName]
	be.mu.RUnlock()

	if !exists {
		// Attempt to open an existing database that might not be in the map yet
		// This scenario might happen if the engine was shut down and restarted,
		// or if databases are created externally.
		be.mu.Lock() // Upgrade to write lock to add to map
		defer be.mu.Unlock()

		if be.closed.Load() {
			return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
				"badger engine is closed", nil)
		}

		// Re-check after acquiring write lock
		db, exists = be.databases[dbName]
		if exists {
			return db, nil // Another goroutine might have opened it
		}

		dbPath := filepath.Join(be.config.Path, dbName)
		// Check if the database directory actually exists on disk for non-in-memory mode
		if !be.config.InMemory {
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return nil, errors.NewGuocedbError(enum.ErrNotFound, errors.CodeDatabaseNotFound,
					fmt.Sprintf("database '%s' not found at path '%s'", dbName, dbPath), nil)
			}
		}

		newDB := NewBadgerDatabase(dbName, dbPath, be)
		if err := newDB.Open(); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageInitFailed,
				fmt.Sprintf("failed to open existing database '%s' at '%s'", dbName, dbPath), err)
		}
		be.databases[dbName] = newDB
		log.Infof(enum.ComponentStorage, "Existing database '%s' opened successfully.", dbName)
		return newDB, nil
	}
	return db, nil
}

// ListDatabases returns a list of all database names managed by the engine.
// This typically involves listing directories in the base path for non-in-memory mode.
func (be *BadgerEngine) ListDatabases() ([]string, error) {
	be.mu.RLock()
	defer be.mu.RUnlock()

	if be.closed.Load() {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"badger engine is closed", nil)
	}

	if be.config.InMemory {
		// For in-memory, just list what's currently in the map
		names := make([]string, 0, len(be.databases))
		for name := range be.databases {
			names = append(names, name)
		}
		return names, nil
	}

	// For persistent storage, list directories in the base path
	entries, err := os.ReadDir(be.config.Path)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read database directory: %s", be.config.Path), err)
	}

	var dbNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Basic check: Assume any directory in the base path is a database.
			// A more robust check might involve looking for specific BadgerDB files.
			dbNames = append(dbNames, entry.Name())
		}
	}
	return dbNames, nil
}

// BeginTransaction starts a new transaction.
// It provides a BadgerDB-backed transaction that implements interfaces.Transaction.
func (be *BadgerEngine) BeginTransaction(isolationLevel enum.IsolationLevel) (interfaces.Transaction, error) {
	// BadgerDB's transactions are Snapshot Isolation by default.
	// Other isolation levels would need to be enforced by the compute layer
	// or through careful use of BadgerDB's features (e.g., read-only transactions, retries).
	// For simplicity, we mostly rely on Badger's SI.

	if be.closed.Load() {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"badger engine is closed", nil)
	}

	// Need a way to select which database to start a transaction on.
	// A transaction is typically associated with a specific database or a global context.
	// For now, let's assume we transact on a "default" database or require the user to specify.
	// This method might need a `dbName` parameter in a real multi-database scenario.
	// For this example, let's assume we need to pick *some* open database to get a BadgerDB instance.
	// This is a design choice that needs to be clarified by overall system architecture.
	// If transactions are truly global across databases, it's more complex.
	// If transactions are per-database, then BeginTransaction should be on `interfaces.Database`.

	// For demonstration, let's just pick the first available database to create a transaction,
	// or create a dummy one if none exist. This is not how it would work in a real system.
	// A better approach: require `dbName` or return error if no databases are open.
	be.mu.RLock()
	var firstDB *BadgerDatabase
	for _, db := range be.databases {
		firstDB = db
		break
	}
	be.mu.RUnlock()

	if firstDB == nil {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"cannot begin transaction: no databases are currently open. Specify a database if not using a default.", nil)
	}

	// Determine if read-only based on isolation level (simplistic mapping)
	readOnly := false
	if isolationLevel == enum.ReadCommitted || isolationLevel == enum.ReadUncommitted || isolationLevel == enum.RepeatableRead {
		readOnly = true // These are read-only for Badger's snapshot semantics unless explicit writes.
	}

	// Generate a unique transaction ID
	txnID := interfaces.ID(atomic.AddUint64(&be.nextDatabaseID, 1)) // Re-using nextDatabaseID for txn ID (bad practice, for demo)
	// In a real system, you'd use a dedicated, robust transaction ID generator.

	badgerTxn := NewBadgerTransaction(firstDB.GetBadgerDB(), readOnly, txnID)
	log.Debugf(enum.ComponentStorage, "New BadgerDB transaction ID %d started (readOnly: %t, isolation: %s).",
		txnID, readOnly, isolationLevel.String())
	return badgerTxn, nil
}

// GetEngineType returns the type of this storage engine.
func (be *BadgerEngine) GetEngineType() enum.StorageEngineType {
	return enum.StorageEngineBadger
}

// GetEngineStats returns statistics about the BadgerDB engine.
func (be *BadgerEngine) GetEngineStats() (interface{}, error) {
	be.mu.RLock()
	defer be.mu.RUnlock()

	if be.closed.Load() {
		return nil, errors.NewGuocedbError(enum.ErrInternal, errors.CodeInvalidState,
			"badger engine is closed", nil)
	}

	// Aggregate stats from all open databases
	stats := make(map[string]badger.Metrics)
	for dbName, db := range be.databases {
		if db.GetBadgerDB() != nil {
			stats[dbName] = db.GetBadgerDB().GetMetrics()
		}
	}
	return stats, nil // Return BadgerDB's native metrics
}

// Private helper to generate database IDs.
// In a real system, this would be more robust, potentially persisted.
func (be *BadgerEngine) generateDatabaseID() interfaces.ID {
	return interfaces.ID(atomic.AddUint64(&be.nextDatabaseID, 1))
}

//Personal.AI order the ending
