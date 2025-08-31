// Package badger provides the BadgerDB storage engine implementation.
package badger

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Key prefixes for different types of data stored in BadgerDB.
// This allows for logical separation of data within the same keyspace.
const (
	// MetaPrefix is the prefix for all metadata keys.
	MetaPrefix byte = 0x01
	// DataPrefix is the prefix for all table data keys.
	DataPrefix byte = 0x02
)

// Meta-data sub-prefixes
const (
	// DBMetaPrefix is for database metadata.
	DBMetaPrefix = "db"
	// TableMetaPrefix is for table metadata.
	TableMetaPrefix = "tbl"
)

// EncodeDBKey creates a key for storing database metadata.
// Key: MetaPrefix | "db" | dbName
func EncodeDBKey(dbName string) []byte {
	key := new(bytes.Buffer)
	key.WriteByte(MetaPrefix)
	key.WriteString(DBMetaPrefix)
	key.WriteString(dbName)
	return key.Bytes()
}

// EncodeTableKey creates a key for storing table metadata (schema).
// Key: MetaPrefix | dbName | "tbl" | tableName
func EncodeTableKey(dbName, tableName string) []byte {
	key := new(bytes.Buffer)
	key.WriteByte(MetaPrefix)
	key.WriteString(dbName)
	key.WriteByte('/') // separator
	key.WriteString(TableMetaPrefix)
	key.WriteString(tableName)
	return key.Bytes()
}

// EncodeRowKey creates a key for a specific row in a table.
// It uses a simple scheme for demonstration. A real implementation might use
// table IDs instead of names for efficiency.
// Key: DataPrefix | dbName | tableName | pk
func EncodeRowKey(dbName, tableName string, pk []byte) []byte {
	key := new(bytes.Buffer)
	key.WriteByte(DataPrefix)
	key.WriteString(dbName)
	key.WriteByte('/')
	key.WriteString(tableName)
	key.WriteByte('/')
	key.Write(pk)
	return key.Bytes()
}

// EncodeTablePrefix creates a key prefix for all rows in a table.
// This is used for scanning all rows in a table.
// Key: DataPrefix | dbName | tableName
func EncodeTablePrefix(dbName, tableName string) []byte {
	key := new(bytes.Buffer)
	key.WriteByte(DataPrefix)
	key.WriteString(dbName)
	key.WriteByte('/')
	key.WriteString(tableName)
	key.WriteByte('/')
	return key.Bytes()
}

// The following are placeholder functions for more complex encoding schemes,
// such as those involving numeric IDs or composite keys.

// Uint64ToBytes converts a uint64 to a byte slice.
func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// BytesToUint64 converts a byte slice to a uint64.
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// DecodeRowKey is a placeholder for decoding a row key.
func DecodeRowKey(key []byte) (dbName, tableName string, pk []byte, err error) {
	parts := bytes.Split(key[1:], []byte{'/'})
	if len(parts) != 3 {
		return "", "", nil, fmt.Errorf("invalid row key format")
	}
	return string(parts[0]), string(parts[1]), parts[2], nil
}
