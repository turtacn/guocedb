package badger

import (
	"encoding/binary"
	"fmt"
)

// This package contains key encoding and decoding logic specific to the BadgerDB
// storage engine implementation. It builds upon the general prefixes defined in
// the internal/encoding package.

const (
	// Prefixes for different types of metadata stored within the engine.
	dbMetaPrefix    = 'd' // Key for database metadata
	tableMetaPrefix = 't' // Key for table metadata (schema, etc.)
)

// encodeDBMetaKey creates a key for storing database metadata.
// Format: MetaPrefix | dbMetaPrefix | dbName
func encodeDBMetaKey(dbName string) []byte {
	key := make([]byte, 0, 1+1+len(dbName))
	key = append(key, byte(dbMetaPrefix))
	key = append(key, []byte(dbName)...)
	return key
}

// encodeTableMetaKey creates a key for storing table metadata (e.g., schema).
// Format: MetaPrefix | tableMetaPrefix | dbName | 0x00 | tableName
func encodeTableMetaKey(dbName, tableName string) []byte {
	key := make([]byte, 0, 1+1+len(dbName)+1+len(tableName))
	key = append(key, byte(tableMetaPrefix))
	key = append(key, []byte(dbName)...)
	key = append(key, 0x00) // Null separator
	key = append(key, []byte(tableName)...)
	return key
}

// encodeRowKey creates a key for a table row.
// Format: TablePrefix | tableID (uint64) | rowPK
func encodeRowKey(tableID uint64, pk []byte) []byte {
	key := make([]byte, 8+len(pk))
	binary.BigEndian.PutUint64(key, tableID)
	copy(key[8:], pk)
	return key
}

// decodeRowKey extracts the primary key part from a full row key.
func decodeRowKey(key []byte) (pk []byte, err error) {
	if len(key) <= 8 {
		return nil, fmt.Errorf("invalid row key: too short")
	}
	return key[8:], nil
}

// TODO: Implement encoding/decoding for secondary index keys.
// This will involve encoding the indexed column values into the key to allow for lookups.
