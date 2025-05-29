// Package badger implements the BadgerDB specific data iterator for Guocedb.
// This file provides an efficient way to traverse data stored in BadgerDB,
// implementing the iterator interface defined in interfaces/storage.go.
// It allows the compute layer (especially the executor) to uniformly iterate
// over data in tables or indexes. It relies on storage/engines/badger/encoding.go
// to decode the key-value data read from BadgerDB.
package badger

import (
	"fmt"
	"io"

	"github.com/dgraph-io/badger/v4" // Import BadgerDB client library

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/interfaces" // Import the defined interfaces
)

// ensure that BadgerRowIterator implements the interfaces.RowIterator interface.
var _ interfaces.RowIterator = (*BadgerRowIterator)(nil)

// BadgerRowIterator implements the interfaces.RowIterator for scanning rows in BadgerDB.
type BadgerRowIterator struct {
	txn         *badger.Txn             // The BadgerDB transaction
	it          *badger.Iterator        // The underlying BadgerDB iterator
	prefix      []byte                  // The key prefix for the table (e.g., RowDataPrefix + dbID + tableID)
	tableSchema *interfaces.TableSchema // The schema of the table being iterated
	scanOptions *interfaces.ScanOptions // Options for the scan (e.g., projections)

	currentRowID interfaces.RowID // The ID of the current row
	currentValue []value.Value    // The decoded values of the current row
	err          error            // Stores any error encountered during iteration

	initialized bool // Flag to ensure iterator is Rewind'd once
}

// NewBadgerRowIterator creates a new BadgerRowIterator.
// `txn` is the BadgerDB transaction, `tableID` is the ID of the table,
// `tableSchema` is the schema of the table, and `opts` are the scan options.
func NewBadgerRowIterator(txn *badger.Txn, dbID, tableID interfaces.ID, tableSchema *interfaces.TableSchema, opts *interfaces.ScanOptions) (*BadgerRowIterator, error) {
	badgerOpts := badger.DefaultIteratorOptions
	badgerOpts.PrefetchSize = 100                      // Optimize for reads
	badgerOpts.Prefix = EncodeRowKey(dbID, tableID, 0) // Start from the beginning of this table's rows

	it := txn.NewIterator(badgerOpts)

	iterator := &BadgerRowIterator{
		txn:         txn,
		it:          it,
		prefix:      badgerOpts.Prefix,
		tableSchema: tableSchema,
		scanOptions: opts,
		initialized: false,
	}

	return iterator, nil
}

// Next advances the iterator to the next row. Returns false if no more rows or an error occurred.
func (bri *BadgerRowIterator) Next() bool {
	if bri.err != nil {
		return false
	}

	// Initialize the iterator on the first Next() call
	if !bri.initialized {
		bri.it.Rewind()
		bri.initialized = true
	} else {
		bri.it.Next()
	}

	if !bri.it.ValidForPrefix(bri.prefix) {
		bri.currentRowID = 0
		bri.currentValue = nil
		bri.err = nil // No more valid items, not an error
		return false
	}

	item := bri.it.Item()
	key := item.KeyCopy(nil)

	// Decode RowID from the key
	_, _, rowID, err := DecodeRowKey(key)
	if err != nil {
		bri.err = errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode row key: %v", err), err)
		return false
	}
	bri.currentRowID = rowID

	// Read and decode the value
	valueBytes, err := item.ValueCopy(nil)
	if err != nil {
		bri.err = errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageReadFailed,
			fmt.Sprintf("failed to read value for key %x: %v", key, err), err)
		return false
	}

	decodedValues, err := DecodeRowData(valueBytes, bri.tableSchema)
	if err != nil {
		bri.err = errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode row data for row %d: %v", rowID, err), err)
		return false
	}

	// Apply projections if specified
	if bri.scanOptions != nil && len(bri.scanOptions.ProjectedColumns) > 0 {
		projectedValues := make([]value.Value, len(bri.scanOptions.ProjectedColumns))
		for i, colID := range bri.scanOptions.ProjectedColumns {
			// Find the corresponding column index in the schema
			colIndex := -1
			for idx, colDef := range bri.tableSchema.Columns {
				// Assuming column IDs are assigned based on schema order, or we need a map
				// For now, let's assume ColumnID directly maps to schema index for simplicity
				// In a real system, you'd likely use a map from ColumnID to index or a richer schema.
				if interfaces.ColumnID(idx) == colID { // This is a simplification; ColumnID should be unique across tables perhaps
					colIndex = idx
					break
				}
			}

			if colIndex == -1 || colIndex >= len(decodedValues) {
				bri.err = errors.NewGuocedbError(enum.ErrExecution, errors.CodeInvalidInput,
					fmt.Sprintf("projected column ID %d not found or out of bounds for table %s", colID, bri.tableSchema.TableName), nil)
				return false
			}
			projectedValues[i] = decodedValues[colIndex]
		}
		bri.currentValue = projectedValues
	} else {
		bri.currentValue = decodedValues
	}

	// TODO: Implement predicate filtering (WHERE clause) here if `ScanOptions.Predicate` is a function/expression.
	// This would involve evaluating `bri.scanOptions.Predicate.Eval(bri.currentValue)` and skipping if it returns false.
	// For simplicity, it's omitted in this direct iterator implementation,
	// as predicates are typically handled by the compute layer (executor).

	// TODO: Implement Limit and Offset logic here by tracking counts and skipping.
	// This would require additional state in the iterator and potentially a loop in Next().

	return true
}

// Current returns the current RowID and the row's values.
// Returns nil for values if Next() returned false or an error occurred.
func (bri *BadgerRowIterator) Current() (interfaces.RowID, []value.Value, error) {
	return bri.currentRowID, bri.currentValue, bri.err
}

// Close releases any resources held by the iterator.
func (bri *BadgerRowIterator) Close() error {
	if bri.it != nil {
		bri.it.Close()
	}
	// Note: The BadgerDB transaction itself is typically managed externally (by BadgerTransaction).
	// We do not commit/discard the transaction here.
	return nil
}

//Personal.AI order the ending
