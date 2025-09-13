package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/turtacn/guocedb/interfaces"
)

// badgerIterator implements the interfaces.Iterator for the BadgerDB engine.
// It wraps a badger.Iterator to provide a consistent iteration interface to the
// upper layers of the database.
type badgerIterator struct {
	it     *badger.Iterator
	prefix []byte
	err    error
}

// newBadgerIterator creates a new iterator.
// The iterator is initialized to the first key matching the prefix.
func newBadgerIterator(txn *badger.Txn, prefix []byte) interfaces.Iterator {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	it.Rewind() // Move to the first key
	return &badgerIterator{
		it:     it,
		prefix: prefix,
	}
}

// Next moves the iterator to the next key. It returns false if the iterator is exhausted.
func (i *badgerIterator) Next() bool {
	if !i.it.Valid() {
		return false
	}
	i.it.Next()
	return i.it.Valid()
}

// Key returns the current key.
func (i *badgerIterator) Key() []byte {
	return i.it.Item().Key()
}

// Value returns the current value.
func (i *badgerIterator) Value() []byte {
	val, err := i.it.Item().ValueCopy(nil)
	if err != nil {
		i.err = err
		return nil
	}
	return val
}

// Error returns any error that occurred during iteration.
func (i *badgerIterator) Error() error {
	return i.err
}

// Close closes the iterator.
func (i *badgerIterator) Close() error {
	i.it.Close()
	return nil
}
