package badger

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/interfaces"
)

// BadgerTable represents a SQL table stored in BadgerDB.
// It implements the sql.Table interface.
type BadgerTable struct {
	db       *BadgerStorage
	name     string
	schema   sql.Schema
	tableID  uint64 // A unique ID for this table, used for key prefixes
}

var _ sql.Table = (*BadgerTable)(nil)

// NewBadgerTable creates a new BadgerTable instance.
func NewBadgerTable(db *BadgerStorage, name string, schema sql.Schema, tableID uint64) *BadgerTable {
	return &BadgerTable{
		db:      db,
		name:    name,
		schema:  schema,
		tableID: tableID,
	}
}

// Name returns the name of the table.
func (t *BadgerTable) Name() string {
	return t.name
}

// String returns a human-readable string representation of the table.
func (t *BadgerTable) String() string {
	return t.name
}

// Schema returns the schema of the table.
func (t *BadgerTable) Schema() sql.Schema {
	return t.schema
}

// Collation returns the collation of the table.
func (t *BadgerTable) Collation() sql.CollationID {
	return sql.Collation_Default
}

// Partitions returns a sql.RowIter for the table's data.
func (t *BadgerTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	// For a simple key-value store, we can treat the whole table as a single partition.
	return &badgerPartitionIter{
		table: t,
	}, nil
}

// PartitionRows returns a sql.RowIter for a given partition.
func (t *BadgerTable) PartitionRows(ctx *sql.Context, part sql.Partition) (sql.RowIter, error) {
	txn, err := t.db.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}

	prefix := encodeRowKey(t.tableID, nil)
	iter, err := txn.NewIterator(ctx, prefix)
	if err != nil {
		return nil, err
	}

	return &badgerRowIter{
		iter:   iter,
		txn:    txn,
		schema: t.schema,
	}, nil
}

// badgerPartitionIter is a simple partition iterator that returns a single partition.
type badgerPartitionIter struct {
	table *BadgerTable
	done  bool
}

func (it *badgerPartitionIter) Next(*sql.Context) (sql.Partition, error) {
	if it.done {
		return nil, io.EOF
	}
	it.done = true
	// The key is just the table ID, as we scan all rows.
	return &badgerPartition{key: encodeRowKey(it.table.tableID, nil)}, nil
}

func (it *badgerPartitionIter) Close(*sql.Context) error { return nil }

// badgerPartition represents a single partition of a table.
type badgerPartition struct {
	key []byte
}

func (p *badgerPartition) Key() []byte { return p.key }

// badgerRowIter iterates over the rows of a table.
type badgerRowIter struct {
	iter   interfaces.Iterator
	txn    interfaces.Transaction
	schema sql.Schema
}

func (it *badgerRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if !it.iter.Next() {
		return nil, io.EOF
	}
	// TODO: Implement proper row decoding from the value.
	// The value would contain the serialized row data.
	// For now, returning an empty row of the correct length.
	return make(sql.Row, len(it.schema)), nil
}

func (it *badgerRowIter) Close(ctx *sql.Context) error {
	it.iter.Close()
	return it.txn.Rollback(ctx) // Ensure transaction is always closed
}
