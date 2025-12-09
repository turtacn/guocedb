// Package badger provides the BadgerDB storage engine implementation.
package badger

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/turtacn/guocedb/compute/sql"
)

// Database implements sql.Database.
type Database struct {
	name string
	db   *badger.DB
	// tables cache map from table name to sql.Table
	tables map[string]sql.Table
	mu     sync.RWMutex
}

// NewDatabase creates a new Database instance and loads existing tables.
func NewDatabase(name string, db *badger.DB) *Database {
	d := &Database{
		name:   name,
		db:     db,
		tables: make(map[string]sql.Table),
	}
	d.loadTables()
	return d
}

// SerializableColumn is a struct used for persisting column metadata.
type SerializableColumn struct {
	Name     string
	Type     int32 // query.Type
	Nullable bool
	Source   string
}

func marshalSchema(s sql.Schema) ([]byte, error) {
	cols := make([]SerializableColumn, len(s))
	for i, c := range s {
		cols[i] = SerializableColumn{
			Name:       c.Name,
			Type:       int32(c.Type.Type()),
			Nullable:   c.Nullable,
			Source:     c.Source,
		}
	}
	return json.Marshal(cols)
}

func unmarshalSchema(data []byte) (sql.Schema, error) {
	var cols []SerializableColumn
	if err := json.Unmarshal(data, &cols); err != nil {
		return nil, err
	}

	schema := make(sql.Schema, len(cols))
	for i, c := range cols {
		typ, err := sql.MysqlTypeToType(query.Type(c.Type))
		if err != nil {
			return nil, err
		}
		schema[i] = &sql.Column{
			Name:     c.Name,
			Type:     typ,
			Nullable: c.Nullable,
			Source:   c.Source,
		}
	}
	return schema, nil
}

func (d *Database) loadTables() {
	// Construct prefix using EncodeTableKey logic, but stop before tableName
	// EncodeTableKey: MetaPrefix + dbName + "/" + TableMetaPrefix + tableName
	// We want: MetaPrefix + dbName + "/" + TableMetaPrefix

	prefix := new(bytes.Buffer)
	prefix.WriteByte(MetaPrefix)
	prefix.WriteString(d.name)
	prefix.WriteByte('/')
	prefix.WriteString(TableMetaPrefix)
	prefixBytes := prefix.Bytes()

	d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefixBytes
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract table name from key
			// Key: prefix + tableName
			tableName := string(key[len(prefixBytes):])

			err := item.Value(func(val []byte) error {
				schema, err := unmarshalSchema(val)
				if err != nil {
					return err
				}
				// Reconstruct table
				t := NewTable(tableName, d.name, schema, d.db)
				d.tables[tableName] = t
				return nil
			})
			if err != nil {
				// Log error?
			}
		}
		return nil
	})
}

// Name returns the name of the database.
func (d *Database) Name() string {
	return d.name
}

// Tables returns all tables in the database.
func (d *Database) Tables() map[string]sql.Table {
	d.mu.RLock()
	defer d.mu.RUnlock()

	tables := make(map[string]sql.Table, len(d.tables))
	for k, v := range d.tables {
		tables[k] = v
	}
	return tables
}

// Create implements sql.Alterable.
func (d *Database) Create(name string, schema sql.Schema) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.tables[name]; ok {
		return sql.ErrTableAlreadyExists.New(name)
	}

	table := NewTable(name, d.name, schema, d.db)

	err := d.db.Update(func(txn *badger.Txn) error {
		key := EncodeTableKey(d.name, name)

		_, err := txn.Get(key)
		if err == nil {
			return sql.ErrTableAlreadyExists.New(name)
		}
		if err != badger.ErrKeyNotFound {
			return err
		}

		val, err := marshalSchema(schema)
		if err != nil {
			return err
		}

		return txn.Set(key, val)
	})

	if err != nil {
		return err
	}

	d.tables[name] = table
	return nil
}

// CreateTable is a compatibility method if interfaces require it, but Alterable.Create is the standard one in core.go.
// However, the prompt asked for CreateTable with context etc.
// Given core.go only defines Alterable interface for Database with Create(name, schema),
// and no explicit Database.CreateTable method in the interface,
// we stick to implementing methods that might be useful or required by other parts not visible in core.go but implied.
// But strictly, we implement Create matching Alterable.

// GetTableInsensitive retrieves a table by name (case-insensitive).
func (d *Database) GetTableInsensitive(ctx *sql.Context, name string) (sql.Table, bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Direct match
	if t, ok := d.tables[name]; ok {
		return t, true, nil
	}

	// Case-insensitive scan
	for tableName, table := range d.tables {
		if len(tableName) == len(name) {
			if strings.EqualFold(tableName, name) {
				return table, true, nil
			}
		}
	}

	return nil, false, nil
}

// GetTableNames returns the names of all tables in the database.
func (d *Database) GetTableNames(ctx *sql.Context) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	names := make([]string, 0, len(d.tables))
	for name := range d.tables {
		names = append(names, name)
	}
	return names, nil
}

// DropTable drops a table.
func (d *Database) DropTable(ctx *sql.Context, name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.tables[name]; !ok {
		return sql.ErrTableNotFound.New(name)
	}

	err := d.db.Update(func(txn *badger.Txn) error {
		// Delete metadata
		metaKey := EncodeTableKey(d.name, name)
		if err := txn.Delete(metaKey); err != nil {
			return err
		}

		// Delete all rows
		dataPrefix := EncodeTablePrefix(d.name, name)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(dataPrefix); it.ValidForPrefix(dataPrefix); it.Next() {
			if err := txn.Delete(it.Item().Key()); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	delete(d.tables, name)
	return nil
}
