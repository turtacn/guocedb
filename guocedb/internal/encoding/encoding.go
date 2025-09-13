package encoding

import (
	"encoding/binary"
	"fmt"
)

// This package provides utilities for encoding and decoding keys and values for the
// internal storage layout. The specific formats defined here are considered an
// internal implementation detail of the database.

// Key prefixes are used to logically partition the key space in the underlying KV store.
const (
	MetaPrefix  byte = 0x01 // Prefix for metadata keys (e.g., schema info)
	TablePrefix byte = 0x02 // Prefix for table row data keys
	IndexPrefix byte = 0x03 // Prefix for secondary index keys
)

// EncodeMetaKey creates a key for storing global or database-level metadata.
// Example usage: EncodeMetaKey([]byte("db"), []byte("mydb"))
func EncodeMetaKey(parts ...[]byte) []byte {
	// A more robust implementation might use length prefixes to avoid collisions.
	key := []byte{MetaPrefix}
	for _, p := range parts {
		key = append(key, p...)
		key = append(key, 0x00) // Null separator to prevent ambiguity
	}
	return key
}

// EncodeTableKey creates a key for a row in a table.
// The key format is: TablePrefix | tableID (uint64) | primaryKey
func EncodeTableKey(tableID uint64, primaryKey []byte) []byte {
	key := make([]byte, 1+8, 1+8+len(primaryKey))
	key[0] = TablePrefix
	binary.BigEndian.PutUint64(key[1:9], tableID)
	key = append(key, primaryKey...)
	return key
}

// DecodeTableKey extracts the tableID and primaryKey from a table row key.
func DecodeTableKey(key []byte) (tableID uint64, primaryKey []byte, err error) {
	if len(key) < 9 || key[0] != TablePrefix {
		return 0, nil, fmt.Errorf("invalid table key format")
	}
	tableID = binary.BigEndian.Uint64(key[1:9])
	primaryKey = key[9:]
	return
}

// TODO: Implement value serialization/deserialization (e.g., using gob, protobuf, or a custom format).
// TODO: Add support for data compression (e.g., snappy or zstd) for large values.
// TODO: Ensure the encoding format is versioned to allow for backward-compatible schema changes.
// TODO: Implement encoding/decoding functions for index keys.
