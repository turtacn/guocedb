package badger

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/compute/sql"
)

// Interfaces not in core but requested by review/design.
// We implement them here for completeness of the "badger" package contract,
// even if core.go doesn't enforce them yet.

// RowInserter allows inserting rows.
type RowInserter interface {
	StatementBegin(ctx *sql.Context)
	Insert(ctx *sql.Context, row sql.Row) error
	StatementComplete(ctx *sql.Context) error
	DiscardChanges(ctx *sql.Context, err error) error
	Close(ctx *sql.Context) error
}

// RowUpdater allows updating rows.
type RowUpdater interface {
	StatementBegin(ctx *sql.Context)
	Update(ctx *sql.Context, oldRow sql.Row, newRow sql.Row) error
	StatementComplete(ctx *sql.Context) error
	DiscardChanges(ctx *sql.Context, err error) error
	Close(ctx *sql.Context) error
}

// RowDeleter allows deleting rows.
type RowDeleter interface {
	StatementBegin(ctx *sql.Context)
	Delete(ctx *sql.Context, row sql.Row) error
	StatementComplete(ctx *sql.Context) error
	DiscardChanges(ctx *sql.Context, err error) error
	Close(ctx *sql.Context) error
}

// InsertableTable is a table that can be inserted into.
type InsertableTable interface {
	Inserter(ctx *sql.Context) RowInserter
}

// UpdatableTable is a table that can be updated.
type UpdatableTable interface {
	Updater(ctx *sql.Context) RowUpdater
}

// DeletableTable is a table that can be deleted from.
type DeletableTable interface {
	Deleter(ctx *sql.Context) RowDeleter
}

// Table implements sql.Table and sql.Inserter.
// It also implements the extended InsertableTable/UpdatableTable/DeletableTable interfaces
// defined in this package for future compatibility or advanced usage.
type Table struct {
	name   string
	dbName string
	schema sql.Schema
	db     *badger.DB
}

// NewTable creates a new Table.
func NewTable(name, dbName string, schema sql.Schema, db *badger.DB) *Table {
	return &Table{
		name:   name,
		dbName: dbName,
		schema: schema,
		db:     db,
	}
}

// Name returns the table name.
func (t *Table) Name() string {
	return t.name
}

// String returns the table name.
func (t *Table) String() string {
	return t.name
}

// Schema returns the table schema.
func (t *Table) Schema() sql.Schema {
	return t.schema
}

// Partitions returns a PartitionIter for the table.
func (t *Table) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &partitionIter{
		partitions: []*Partition{{key: []byte(t.name)}},
		idx:        0,
	}, nil
}

// PartitionRows returns a RowIter for the given partition.
func (t *Table) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	txn := t.db.NewTransaction(false) // Read-only
	prefix := EncodeTablePrefix(t.dbName, t.name)

	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)

	iter.Seek(prefix)

	return &tableRowIter{
		iter:   iter,
		txn:    txn,
		schema: t.schema,
		prefix: prefix,
	}, nil
}

// Insert implements sql.Inserter (core interface).
// It delegates to a short-lived rowEditor to perform the insert.
// Note: This creates a transaction per row, which is safe but slow.
// If the engine supported StatementBegin/Complete hooks, we could optimize this.
func (t *Table) Insert(ctx *sql.Context, row sql.Row) error {
	inserter := t.Inserter(ctx)
	defer inserter.Close(ctx)

	inserter.StatementBegin(ctx)
	if err := inserter.Insert(ctx, row); err != nil {
		inserter.DiscardChanges(ctx, err)
		return err
	}
	return inserter.StatementComplete(ctx)
}

// Inserter returns a RowInserter for the table.
func (t *Table) Inserter(ctx *sql.Context) RowInserter {
	return &rowEditor{
		table: t,
	}
}

// Updater returns a RowUpdater for the table.
func (t *Table) Updater(ctx *sql.Context) RowUpdater {
	return &rowEditor{
		table: t,
	}
}

// Deleter returns a RowDeleter for the table.
func (t *Table) Deleter(ctx *sql.Context) RowDeleter {
	return &rowEditor{
		table: t,
	}
}

type rowEditor struct {
	table *Table
	txn   *badger.Txn
}

// StatementBegin starts a transaction.
func (re *rowEditor) StatementBegin(ctx *sql.Context) {
	re.txn = re.table.db.NewTransaction(true)
}

// DiscardChanges discards the transaction.
func (re *rowEditor) DiscardChanges(ctx *sql.Context, err error) error {
	if re.txn != nil {
		re.txn.Discard()
		re.txn = nil
	}
	return nil
}

// StatementComplete commits the transaction.
func (re *rowEditor) StatementComplete(ctx *sql.Context) error {
	if re.txn != nil {
		err := re.txn.Commit()
		re.txn = nil
		return err
	}
	return nil
}

// Close closes the editor.
func (re *rowEditor) Close(ctx *sql.Context) error {
	if re.txn != nil {
		re.txn.Discard()
		re.txn = nil
	}
	return nil
}

// Insert inserts a row.
func (re *rowEditor) Insert(ctx *sql.Context, row sql.Row) error {
	key, val, err := re.encodeRow(row)
	if err != nil {
		return err
	}

	if re.txn != nil {
		return re.txn.Set(key, val)
	}

	// Fallback should not be reached if properly used, but just in case:
	return re.table.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

// Update updates a row.
func (re *rowEditor) Update(ctx *sql.Context, oldRow sql.Row, newRow sql.Row) error {
	oldKey, _, err := re.encodeRow(oldRow)
	if err != nil {
		return err
	}

	newKey, newVal, err := re.encodeRow(newRow)
	if err != nil {
		return err
	}

	if bytes.Equal(oldKey, newKey) {
		if re.txn != nil {
			return re.txn.Set(newKey, newVal)
		}
		return re.table.db.Update(func(txn *badger.Txn) error {
			return txn.Set(newKey, newVal)
		})
	}

	// PK changed
	if re.txn != nil {
		if err := re.txn.Delete(oldKey); err != nil {
			return err
		}
		return re.txn.Set(newKey, newVal)
	}

	return re.table.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(oldKey); err != nil {
			return err
		}
		return txn.Set(newKey, newVal)
	})
}

// Delete deletes a row.
func (re *rowEditor) Delete(ctx *sql.Context, row sql.Row) error {
	key, _, err := re.encodeRow(row)
	if err != nil {
		return err
	}

	if re.txn != nil {
		return re.txn.Delete(key)
	}

	return re.table.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (re *rowEditor) encodeRow(row sql.Row) ([]byte, []byte, error) {
	// Simple assumption: First column is PK.
	if len(row) == 0 {
		return nil, nil, nil
	}

	pkVal := row[0]
	var pkBytes []byte

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(pkVal); err != nil {
		return nil, nil, err
	}
	pkBytes = buf.Bytes()

	key := EncodeRowKey(re.table.dbName, re.table.name, pkBytes)

	var valBuf bytes.Buffer
	valEnc := gob.NewEncoder(&valBuf)
	if err := valEnc.Encode(row); err != nil {
		return nil, nil, err
	}

	return key, valBuf.Bytes(), nil
}

// tableRowIter implements sql.RowIter.
type tableRowIter struct {
	iter   *badger.Iterator
	txn    *badger.Txn
	schema sql.Schema
	prefix []byte
}

func (i *tableRowIter) Next() (sql.Row, error) {
	if !i.iter.ValidForPrefix(i.prefix) {
		return nil, io.EOF
	}

	item := i.iter.Item()
	var row sql.Row
	err := item.Value(func(val []byte) error {
		buf := bytes.NewBuffer(val)
		dec := gob.NewDecoder(buf)
		return dec.Decode(&row)
	})

	if err != nil {
		return nil, err
	}

	i.iter.Next()
	return row, nil
}

func (i *tableRowIter) Close() error {
	i.iter.Close()
	i.txn.Discard()
	return nil
}
