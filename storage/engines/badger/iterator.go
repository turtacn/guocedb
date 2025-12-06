// Package badger provides the BadgerDB storage engine implementation.
package badger

import (
	"github.com/dgraph-io/badger/v3"
)

// Iterator is a wrapper around a badger.Iterator that implements the interfaces.Iterator.
type Iterator struct {
	txn    *badger.Txn
	iter   *badger.Iterator
	prefix []byte
	item   *badger.Item
	err    error
}

// newIterator creates a new iterator for a given transaction and prefix.
func newIterator(txn *badger.Txn, prefix []byte) *Iterator {
	opts := badger.DefaultIteratorOptions
	// A real implementation might configure these options.
	// opts.PrefetchValues = true
	// opts.PrefetchSize = 100
	it := txn.NewIterator(opts)
	it.Seek(prefix)
	return &Iterator{
		txn:    txn,
		iter:   it,
		prefix: prefix,
	}
}

// Next moves the iterator to the next key/value pair.
func (it *Iterator) Next() bool {
	if !it.iter.ValidForPrefix(it.prefix) {
		return false
	}
	it.item = it.iter.Item()
	it.iter.Next()
	return true
}

// Key returns the current key.
func (it *Iterator) Key() []byte {
	if it.item == nil {
		return nil
	}
	// The key is only valid until the next call to Next().
	// We need to copy it to be safe.
	key := it.item.Key()
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	return keyCopy
}

// Value returns the current value.
func (it *Iterator) Value() []byte {
	if it.item == nil {
		return nil
	}
	// The value is only valid until the transaction is closed.
	// We need to copy it.
	var valCopy []byte
	err := it.item.Value(func(val []byte) error {
		valCopy = append([]byte{}, val...)
		return nil
	})
	if err != nil {
		it.err = err
		return nil
	}
	return valCopy
}

// Error returns any error that occurred during iteration.
func (it *Iterator) Error() error {
	return it.err
}

// Close closes the iterator and releases any resources.
func (it *Iterator) Close() error {
	it.iter.Close()
	return nil
}
