// Package badger implements the BadgerDB specific table-level operations for Guocedb.
// This file is responsible for mapping SQL table operations (CRUD, scanning, index management)
// to BadgerDB's key-value operations. It relies on storage/engines/badger/encoding.go
// for encoding and decoding key-value data, storage/engines/badger/iterator.go for data scanning,
// and storage/engines/badger/transaction.go for transaction management.
// storage/engines/badger/badger.go will use this file to manage table data.
package badger

import (
	"encoding/binary"
	"fmt"
	"sync/atomic" // For atomic operations on sequence counters

	"github.com/dgraph-io/badger/v4" // Import BadgerDB client library

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces
)

// ensure that BadgerTable implements the interfaces.Table interface.
var _ interfaces.Table = (*BadgerTable)(nil)

// BadgerTable represents a logical table within a BadgerDB database.
// It manages the storage and retrieval of rows for a specific table.
type BadgerTable struct {
	db           *badger.DB              // The underlying BadgerDB instance
	dbID         interfaces.ID           // The ID of the parent database
	tableID      interfaces.ID           // The unique ID of this table
	tableSchema  *interfaces.TableSchema // The schema of this table
	rowIDCounter uint64                  // Atomic counter for generating unique RowIDs for this table
}

// NewBadgerTable creates a new BadgerTable instance.
// This function is typically called by BadgerDatabase.
func NewBadgerTable(db *badger.DB, dbID, tableID interfaces.ID, schema *interfaces.TableSchema) *BadgerTable {
	return &BadgerTable{
		db:           db,
		dbID:         dbID,
		tableID:      tableID,
		tableSchema:  schema,
		rowIDCounter: 0, // Initialize, will be loaded or started from 1
	}
}

// Name returns the name of the table.
func (bt *BadgerTable) Name() string {
	return bt.tableSchema.TableName
}

// Schema returns the schema of the table.
func (bt *BadgerTable) Schema() *interfaces.TableSchema {
	return bt.tableSchema
}

// initRowIDCounter initializes the row ID counter for this table.
// It tries to load the last used RowID from BadgerDB. If not found, it starts from 0.
func (bt *BadgerTable) initRowIDCounter(txn *BadgerTransaction) error {
	seqKey := EncodeSequenceKey(bt.dbID, bt.tableID)
	item, err := txn.GetBadgerTxn().Get(seqKey)
	if err == badger.ErrKeyNotFound {
		bt.rowIDCounter = 0 // Start from 0 if no sequence exists
		log.Infof(enum.ComponentStorage, "Row ID counter for table '%s' initialized to 0.", bt.Name())
		return nil
	}
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read row ID sequence for table '%s'", bt.Name()), err)
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to copy row ID sequence value for table '%s'", bt.Name()), err)
	}

	if len(val) != 8 { // Expecting 8 bytes for uint64
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("invalid length for row ID sequence value for table '%s'", bt.Name()), nil)
	}
	bt.rowIDCounter = atomic.LoadUint64((*uint64)(val)) // Load current counter
	log.Infof(enum.ComponentStorage, "Row ID counter for table '%s' loaded as %d.", bt.Name(), bt.rowIDCounter)
	return nil
}

// getNextRowID generates a new unique RowID for the table.
// It atomically increments the internal counter and persists it.
func (bt *BadgerTable) getNextRowID(txn *BadgerTransaction) (interfaces.RowID, error) {
	// Atomically increment the counter
	newID := atomic.AddUint64(&bt.rowIDCounter, 1)

	// Persist the new counter value
	seqKey := EncodeSequenceKey(bt.dbID, bt.tableID)
	seqValue := make([]byte, 8)
	binary.BigEndian.PutUint64(seqValue, newID)

	err := txn.GetBadgerTxn().Set(seqKey, seqValue)
	if err != nil {
		// Decrement counter if persistence fails to avoid gaps on retry
		atomic.AddUint64(&bt.rowIDCounter, ^uint64(0)) // Subtract 1
		return 0, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to persist new row ID for table '%s'", bt.Name()), err)
	}
	return interfaces.RowID(newID), nil
}

// InsertRow inserts a new row into the table.
// The order of values must match the column order in the schema.
func (bt *BadgerTable) InsertRow(txn interfaces.Transaction, values []value.Value) (interfaces.RowID, error) {
	badgerTxn, ok := txn.(*BadgerTransaction)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrTransaction, errors.CodeInvalidTransaction,
			"invalid transaction type provided to BadgerTable", nil)
	}

	if len(values) != len(bt.tableSchema.Columns) {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeColumnCountMismatch,
			fmt.Sprintf("number of values (%d) does not match table schema columns (%d)",
				len(values), len(bt.tableSchema.Columns)), nil)
	}

	// First, initialize the row ID counter if not already done.
	// This should ideally be done once when the table is opened,
	// but can be lazy-loaded here for simplicity in this example.
	if bt.rowIDCounter == 0 {
		err := bt.initRowIDCounter(badgerTxn)
		if err != nil {
			return 0, err
		}
	}

	rowID, err := bt.getNextRowID(badgerTxn)
	if err != nil {
		return 0, err
	}

	rowKey := EncodeRowKey(bt.dbID, bt.tableID, rowID)
	rowData, err := EncodeRowData(values)
	if err != nil {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode row data for table '%s'", bt.Name()), err)
	}

	err = badgerTxn.GetBadgerTxn().Set(rowKey, rowData)
	if err != nil {
		return 0, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to insert row into table '%s'", bt.Name()), err)
	}

	log.Debugf(enum.ComponentStorage, "Inserted row with ID %d into table '%s'.", rowID, bt.Name())
	return rowID, nil
}

// ReadRow reads a row from the table given its RowID.
// Returns the values in the order of the table schema's columns.
func (bt *BadgerTable) ReadRow(txn interfaces.Transaction, rowID interfaces.RowID) ([]value.Value, error) {
	badgerTxn, ok := txn.(*BadgerTransaction)
	if !ok {
		return nil, errors.NewGuocedbError(enum.ErrTransaction, errors.CodeInvalidTransaction,
			"invalid transaction type provided to BadgerTable", nil)
	}

	rowKey := EncodeRowKey(bt.dbID, bt.tableID, rowID)
	item, err := badgerTxn.GetBadgerTxn().Get(rowKey)
	if err == badger.ErrKeyNotFound {
		return nil, errors.NewGuocedbError(enum.ErrNotFound, errors.CodeRowNotFound,
			fmt.Sprintf("row with ID %d not found in table '%s'", rowID, bt.Name()), nil)
	}
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read row with ID %d from table '%s'", rowID, bt.Name()), err)
	}

	rowData, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to get value for row key %x in table '%s'", rowKey, bt.Name()), err)
	}

	decodedValues, err := DecodeRowData(rowData, bt.tableSchema)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode row data for row ID %d in table '%s'", rowID, bt.Name()), err)
	}

	log.Debugf(enum.ComponentStorage, "Read row with ID %d from table '%s'.", rowID, bt.Name())
	return decodedValues, nil
}

// UpdateRow updates an existing row.
// 'updates' is a map of ColumnID to new Value.
func (bt *BadgerTable) UpdateRow(txn interfaces.Transaction, rowID interfaces.RowID, updates map[interfaces.ColumnID]value.Value) error {
	badgerTxn, ok := txn.(*BadgerTransaction)
	if !ok {
		return errors.NewGuocedbError(enum.ErrTransaction, errors.CodeInvalidTransaction,
			"invalid transaction type provided to BadgerTable", nil)
	}

	// Read the existing row
	existingValues, err := bt.ReadRow(badgerTxn, rowID)
	if err != nil {
		return err // ReadRow already wraps errors
	}

	// Apply updates to a mutable copy of the row
	newValues := make([]value.Value, len(existingValues))
	copy(newValues, existingValues)

	for colID, newVal := range updates {
		colIndex := -1
		for idx, colDef := range bt.tableSchema.Columns {
			// This assumes ColumnID directly maps to index or needs a proper mapping
			// For simplicity, let's assume it's the 0-based index of the column in schema
			if interfaces.ColumnID(idx) == colID {
				colIndex = idx
				break
			}
		}

		if colIndex == -1 || colIndex >= len(newValues) {
			return errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidColumn,
				fmt.Sprintf("column ID %d not found or out of bounds for table '%s'", colID, bt.Name()), nil)
		}
		newValues[colIndex] = newVal
	}

	// Encode the updated row
	rowKey := EncodeRowKey(bt.dbID, bt.tableID, rowID)
	updatedRowData, err := EncodeRowData(newValues)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to encode updated row data for table '%s'", bt.Name()), err)
	}

	// Write the updated row back to BadgerDB
	err = badgerTxn.GetBadgerTxn().Set(rowKey, updatedRowData)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageWriteFailed,
			fmt.Sprintf("failed to update row with ID %d in table '%s'", rowID, bt.Name()), err)
	}

	log.Debugf(enum.ComponentStorage, "Updated row with ID %d in table '%s'.", rowID, bt.Name())
	return nil
}

// DeleteRow deletes a row from the table.
func (bt *BadgerTable) DeleteRow(txn interfaces.Transaction, rowID interfaces.RowID) error {
	badgerTxn, ok := txn.(*BadgerTransaction)
	if !ok {
		return errors.NewGuocedbError(enum.ErrTransaction, errors.CodeInvalidTransaction,
			"invalid transaction type provided to BadgerTable", nil)
	}

	rowKey := EncodeRowKey(bt.dbID, bt.tableID, rowID)
	err := badgerTxn.GetBadgerTxn().Delete(rowKey)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageDeleteFailed,
			fmt.Sprintf("failed to delete row with ID %d from table '%s'", rowID, bt.Name()), err)
	}

	log.Debugf(enum.ComponentStorage, "Deleted row with ID %d from table '%s'.", rowID, bt.Name())
	return nil
}

// GetRowIterator returns an iterator for scanning rows.
func (bt *BadgerTable) GetRowIterator(txn interfaces.Transaction, opts *interfaces.ScanOptions) (interfaces.RowIterator, error) {
	badgerTxn, ok := txn.(*BadgerTransaction)
	if !ok {
		return nil, errors.NewGuocedbError(enum.ErrTransaction, errors.CodeInvalidTransaction,
			"invalid transaction type provided to BadgerTable", nil)
	}

	iterator, err := NewBadgerRowIterator(badgerTxn.GetBadgerTxn(), bt.dbID, bt.tableID, bt.tableSchema, opts)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIteratorCreationFailed,
			fmt.Sprintf("failed to create row iterator for table '%s'", bt.Name()), err)
	}
	return iterator, nil
}

// GetApproxRowCount returns an approximate count of rows in the table.
// This is an estimation and might not be perfectly accurate, especially with concurrent writes/deletes.
func (bt *BadgerTable) GetApproxRowCount() (int64, error) {
	// BadgerDB doesn't expose a direct row count for a specific prefix easily.
	// A full scan would be too slow. We can rely on the sequence counter as a proxy
	// for the highest RowID ever assigned, which gives an upper bound.
	// If rows are deleted, this will overestimate. For an accurate count, a dedicated
	// counter or index scan would be needed (or a full scan).
	// For now, we'll return the highest assigned RowID as an approximation.
	// Alternatively, one could periodically count and store it, or use Badger's metrics.
	return int64(atomic.LoadUint64(&bt.rowIDCounter)), nil
}

// GetApproxTableSize returns an approximate size of the table on disk in bytes.
// This is also an estimation and might not be perfectly accurate.
func (bt *BadgerTable) GetApproxTableSize() (int64, error) {
	// BadgerDB provides total size, but not easily per-table size without custom logic.
	// This would involve iterating over keys with the table's prefix and summing item sizes.
	// For simplicity, returning 0 for now or a general engine size.
	// In a real system, you might implement a background task to calculate and cache this.
	// Or, if using value log, it's harder to isolate per table.
	// For now, let's return a dummy value.
	log.Warnf(enum.ComponentStorage, "Approximate table size for '%s' is not directly available per table in BadgerDB without detailed scan. Returning 0.", bt.Name())
	return 0, nil
}

//Personal.AI order the ending
