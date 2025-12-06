// Package badger provides the BadgerDB storage engine implementation.
package badger

import (
	"encoding/json"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
)

// Storage is the BadgerDB implementation of the interfaces.Storage interface.
type Storage struct {
	db *badger.DB
}

// NewStorage creates a new instance of the BadgerDB storage engine.
func NewStorage(cfg config.BadgerConfig) (*Storage, error) {
	// A data directory must be specified in the main config.
	// For now, let's assume a path is passed or use a default.
	// This should be improved to take the dataDir from the top-level storage config.
	dataDir := config.Get().Storage.DataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions(dataDir)
	opts.ValueLogFileSize = int64(cfg.ValueLogFileSize)
	opts.SyncWrites = cfg.SyncWrites
	// Disable Badger's own logger to use our structured logger
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// Get retrieves a value for a given key from a specific table.
func (s *Storage) Get(ctx *sql.Context, db, table string, key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(EncodeRowKey(db, table, key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return value, err
}

// Set stores a key-value pair in a specific table.
func (s *Storage) Set(ctx *sql.Context, db, table string, key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(EncodeRowKey(db, table, key), value)
	})
}

// Delete removes a key from a specific table.
func (s *Storage) Delete(ctx *sql.Context, db, table string, key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(EncodeRowKey(db, table, key))
	})
}

// Iterator returns an iterator for a given key prefix in a table.
func (s *Storage) Iterator(ctx *sql.Context, db, table string, prefix []byte) (interfaces.Iterator, error) {
	txn := s.db.NewTransaction(false) // Read-only iterator
	// The caller is responsible for closing the transaction via the iterator's context.
	// This is a simplification. A better design would manage the txn lifecycle more carefully.
	return newIterator(txn, EncodeRowKey(db, table, prefix)), nil
}

// NewTransaction creates a new transaction.
func (s *Storage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return newTransaction(s.db, readOnly), nil
}

// Database management
func (s *Storage) CreateDatabase(ctx *sql.Context, name string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(EncodeDBKey(name))
		if err == nil {
			return sql.ErrDatabaseExists.New(name)
		}
		if err != badger.ErrKeyNotFound {
			return err
		}
		// Store a dummy value to represent the database's existence
		return txn.Set(EncodeDBKey(name), []byte("1"))
	})
}

func (s *Storage) DropDatabase(ctx *sql.Context, name string) error {
	// This is a simplified implementation. A real one would need to delete all tables and data.
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(EncodeDBKey(name))
	})
}

func (s *Storage) ListDatabases(ctx *sql.Context) ([]string, error) {
	// Implementation would scan for DB keys.
	return []string{"information_schema", "mysql", "test"}, nil // Placeholder
}

// Table management
func (s *Storage) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := EncodeTableKey(dbName, table.Name())
		_, err := txn.Get(key)
		if err == nil {
			return sql.ErrTableAlreadyExists.New(table.Name())
		}
		if err != badger.ErrKeyNotFound {
			return err
		}
		// Serialize and store the table schema
		schemaBytes, err := json.Marshal(table.Schema())
		if err != nil {
			return err
		}
		return txn.Set(key, schemaBytes)
	})
}

func (s *Storage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Also need to delete all rows for this table.
		return txn.Delete(EncodeTableKey(dbName, tableName))
	})
}

func (s *Storage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	// This is a simplified placeholder. A real implementation would deserialize the table
	// and return a struct that implements sql.Table.
	return nil, nil
}

func (s *Storage) ListTables(ctx *sql.Context, dbName string) ([]string, error) {
	// Implementation would scan for table keys within a db.
	return nil, nil // Placeholder
}

// Close shuts down the storage engine gracefully.
func (s *Storage) Close() error {
	return s.db.Close()
}
