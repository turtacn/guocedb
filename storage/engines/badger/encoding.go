// Package badger implements the BadgerDB specific encoding and decoding strategies for Guocedb.
// This file is responsible for converting SQL value types handled by the compute layer
// (defined in common/types/value/value.go) into byte arrays suitable for BadgerDB storage, and vice-versa.
// It relies on common/types/value/value.go for type conversion and internal/encoding/encoding.go
// for some general encoding utilities. It is core to Badger engine's ability to correctly read and write data.
package badger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types/enum"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/interfaces" // For interfaces.ID, interfaces.RowID, interfaces.ColumnID
	"github.com/turtacn/guocedb/internal/encoding"
)

// KeyTypePrefix defines prefixes for different key types in BadgerDB.
// This helps in distinguishing between different kinds of data stored in BadgerDB.
type KeyTypePrefix byte

const (
	// Prefix for database metadata keys.
	DatabaseMetadataPrefix KeyTypePrefix = 0x01
	// Prefix for table metadata keys.
	TableMetadataPrefix KeyTypePrefix = 0x02
	// Prefix for row data keys.
	RowDataPrefix KeyTypePrefix = 0x03
	// Prefix for index keys.
	IndexPrefix KeyTypePrefix = 0x04
	// Prefix for sequence/counter keys (e.g., for generating RowIDs).
	SequencePrefix KeyTypePrefix = 0x05
	// Prefix for transaction metadata keys (e.g., for MVCC).
	TransactionMetadataPrefix KeyTypePrefix = 0x06
	// Prefix for column metadata (within a table metadata key).
	ColumnMetadataPrefix KeyTypePrefix = 0x07
)

// EncodeKey prefixes a key with its type and concatenates other parts.
// This forms the full key used in BadgerDB.
// Example: EncodeKey(RowDataPrefix, databaseID, tableID, rowID)
func EncodeKey(prefix KeyTypePrefix, parts ...[]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(prefix))
	for _, part := range parts {
		buf.Write(part)
	}
	return buf.Bytes()
}

// EncodeDatabaseMetadataKey encodes a key for database metadata.
// Format: DatabaseMetadataPrefix + databaseName
func EncodeDatabaseMetadataKey(dbName string) []byte {
	return EncodeKey(DatabaseMetadataPrefix, []byte(dbName))
}

// DecodeDatabaseMetadataKey decodes a database metadata key to get the database name.
func DecodeDatabaseMetadataKey(key []byte) (string, error) {
	if len(key) < 1 || KeyTypePrefix(key[0]) != DatabaseMetadataPrefix {
		return "", errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			"invalid database metadata key prefix", nil)
	}
	return string(key[1:]), nil
}

// EncodeTableMetadataKey encodes a key for table metadata.
// Format: TableMetadataPrefix + databaseIDBytes + tableName
func EncodeTableMetadataKey(databaseID interfaces.ID, tableName string) []byte {
	dbIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dbIDBytes, uint64(databaseID))
	return EncodeKey(TableMetadataPrefix, dbIDBytes, []byte(tableName))
}

// DecodeTableMetadataKey decodes a table metadata key.
func DecodeTableMetadataKey(key []byte) (interfaces.ID, string, error) {
	if len(key) < 9 || KeyTypePrefix(key[0]) != TableMetadataPrefix { // 1 byte prefix + 8 bytes ID
		return 0, "", errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			"invalid table metadata key prefix or length", nil)
	}
	dbID := interfaces.ID(binary.BigEndian.Uint64(key[1:9]))
	tableName := string(key[9:])
	return dbID, tableName, nil
}

// EncodeRowKey encodes a key for a specific row.
// Format: RowDataPrefix + databaseIDBytes + tableIDBytes + rowIDBytes
func EncodeRowKey(databaseID interfaces.ID, tableID interfaces.ID, rowID interfaces.RowID) []byte {
	dbIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dbIDBytes, uint64(databaseID))
	tblIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tblIDBytes, uint64(tableID))
	rIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rIDBytes, uint64(rowID))
	return EncodeKey(RowDataPrefix, dbIDBytes, tblIDBytes, rIDBytes)
}

// DecodeRowKey decodes a row key to extract database ID, table ID, and row ID.
func DecodeRowKey(key []byte) (interfaces.ID, interfaces.ID, interfaces.RowID, error) {
	// Expected length: 1 (prefix) + 8 (dbID) + 8 (tableID) + 8 (rowID) = 25 bytes
	if len(key) != 25 || KeyTypePrefix(key[0]) != RowDataPrefix {
		return 0, 0, 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			"invalid row key prefix or length", nil)
	}
	dbID := interfaces.ID(binary.BigEndian.Uint64(key[1:9]))
	tableID := interfaces.ID(binary.BigEndian.Uint64(key[9:17]))
	rowID := interfaces.RowID(binary.BigEndian.Uint64(key[17:25]))
	return dbID, tableID, rowID, nil
}

// EncodeColumnValue encodes a single SQL value into bytes for storage.
// This typically uses the common/types/value.Value's Bytes() method.
func EncodeColumnValue(val value.Value) ([]byte, error) {
	if val == nil {
		// Represent NULL values in a specific way, e.g., a special byte or empty slice
		// For now, let's represent NULL as a single null byte (0x00) for simple types
		// and ensure it's distinct from actual 0 values.
		// A more robust scheme might use a prefix for NULL or an explicit NULL Value type.
		return []byte{0x00}, nil // Placeholder for NULL
	}
	// Use the Value's own Bytes() method for its specific serialization
	// This relies on common/types/value.Value having robust binary serialization.
	return val.Bytes()
}

// DecodeColumnValue decodes bytes from storage into a SQL value.
// It requires the expected SQLType to correctly interpret the bytes.
func DecodeColumnValue(data []byte, sqlType enum.SQLType) (value.Value, error) {
	if len(data) == 1 && data[0] == 0x00 { // Placeholder for NULL
		return value.NewNullValue(sqlType), nil
	}

	// Use the value.ValueFromBytes method from common/types/value
	val, err := value.ValueFromBytes(data, sqlType)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to decode column value for type %s", sqlType.String()), err)
	}
	return val, nil
}

// EncodeRowData encodes an entire row of values into a single byte slice.
// This simple approach concatenates column values. A more advanced approach
// might use Protobuf or a custom binary format for efficiency and schema evolution.
// For now, we assume fixed order and rely on value.Value.Bytes() for individual column encoding.
func EncodeRowData(values []value.Value) ([]byte, error) {
	var buffer bytes.Buffer
	for i, val := range values {
		encodedVal, err := EncodeColumnValue(val)
		if err != nil {
			return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
				fmt.Sprintf("failed to encode value for column %d", i), err)
		}
		// Prefix each value with its length to allow robust deserialization.
		// Use a fixed-size integer for length prefix, e.g., 4 bytes (uint32).
		lenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBytes, uint32(len(encodedVal)))
		buffer.Write(lenBytes)
		buffer.Write(encodedVal)
	}
	return buffer.Bytes(), nil
}

// DecodeRowData decodes a byte slice back into a slice of SQL values.
// It requires the TableSchema to know the expected types of columns.
func DecodeRowData(data []byte, schema *interfaces.TableSchema) ([]value.Value, error) {
	reader := bytes.NewReader(data)
	decodedValues := make([]value.Value, len(schema.Columns))

	for i, colDef := range schema.Columns {
		// Read length prefix
		lenBytes := make([]byte, 4)
		if _, err := reader.Read(lenBytes); err != nil {
			if err == io.EOF && i < len(schema.Columns) {
				return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
					fmt.Sprintf("unexpected EOF while reading length prefix for column %d (expected %d columns)", i, len(schema.Columns)), err)
			}
			return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to read length prefix for column %d", i), err)
		}
		valLen := binary.BigEndian.Uint32(lenBytes)

		// Read the actual value bytes
		valBytes := make([]byte, valLen)
		if _, err := reader.Read(valBytes); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to read value bytes for column %d", i), err)
		}

		val, err := DecodeColumnValue(valBytes, colDef.SQLType)
		if err != nil {
			return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
				fmt.Sprintf("failed to decode value for column '%s' (index %d)", colDef.Name, i), err)
		}
		decodedValues[i] = val
	}

	if reader.Len() > 0 {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("extra bytes found after decoding all columns (%d bytes remaining)", reader.Len()), nil)
	}

	return decodedValues, nil
}

// EncodeSequenceKey encodes a key for a sequence (e.g., for next RowID).
// Format: SequencePrefix + databaseIDBytes + tableIDBytes
func EncodeSequenceKey(databaseID interfaces.ID, tableID interfaces.ID) []byte {
	dbIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dbIDBytes, uint64(databaseID))
	tblIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tblIDBytes, uint64(tableID))
	return EncodeKey(SequencePrefix, dbIDBytes, tblIDBytes)
}

// EncodeTransactionMetadataKey encodes a key for transaction metadata.
// Format: TransactionMetadataPrefix + transactionIDBytes
func EncodeTransactionMetadataKey(txnID interfaces.ID) []byte {
	txnIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(txnIDBytes, uint64(txnID))
	return EncodeKey(TransactionMetadataPrefix, txnIDBytes)
}

// DecodeTransactionMetadataKey decodes a transaction metadata key.
func DecodeTransactionMetadataKey(key []byte) (interfaces.ID, error) {
	if len(key) != 9 || KeyTypePrefix(key[0]) != TransactionMetadataPrefix {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			"invalid transaction metadata key prefix or length", nil)
	}
	txnID := interfaces.ID(binary.BigEndian.Uint64(key[1:9]))
	return txnID, nil
}

// EncodeTableSchemaValue encodes a TableSchema into bytes.
// This typically uses a structured serialization format like Protobuf or YAML.
// For now, let's use the YAML encoder for human-readability and simplicity in this example.
func EncodeTableSchemaValue(schema *interfaces.TableSchema) ([]byte, error) {
	yamlEncoder := encoding.NewYAMLEncoder()
	data, err := yamlEncoder.Encode(schema)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			"failed to encode table schema to YAML", err)
	}
	return data, nil
}

// DecodeTableSchemaValue decodes bytes into a TableSchema.
func DecodeTableSchemaValue(data []byte) (*interfaces.TableSchema, error) {
	yamlDecoder := encoding.NewYAMLDecoder()
	schema := &interfaces.TableSchema{}
	err := yamlDecoder.Decode(data, schema)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			"failed to decode table schema from YAML", err)
	}
	return schema, nil
}

// EncodeDatabaseID encodes an interfaces.ID (for database) into a byte slice.
func EncodeDatabaseID(id interfaces.ID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// DecodeDatabaseID decodes a byte slice into an interfaces.ID (for database).
func DecodeDatabaseID(b []byte) (interfaces.ID, error) {
	if len(b) != 8 {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			fmt.Sprintf("invalid byte slice length for database ID: expected 8, got %d", len(b)), nil)
	}
	return interfaces.ID(binary.BigEndian.Uint64(b)), nil
}

// EncodeTableID encodes an interfaces.ID (for table) into a byte slice.
func EncodeTableID(id interfaces.ID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// DecodeTableID decodes a byte slice into an interfaces.ID (for table).
func DecodeTableID(b []byte) (interfaces.ID, error) {
	if len(b) != 8 {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			fmt.Sprintf("invalid byte slice length for table ID: expected 8, got %d", len(b)), nil)
	}
	return interfaces.ID(binary.BigEndian.Uint64(b)), nil
}

// EncodeRowID encodes an interfaces.RowID into a byte slice.
func EncodeRowID(id interfaces.RowID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// DecodeRowID decodes a byte slice into an interfaces.RowID.
func DecodeRowID(b []byte) (interfaces.RowID, error) {
	if len(b) != 8 {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			fmt.Sprintf("invalid byte slice length for row ID: expected 8, got %d", len(b)), nil)
	}
	return interfaces.RowID(binary.BigEndian.Uint64(b)), nil
}

// EncodeColumnID encodes an interfaces.ColumnID into a byte slice.
func EncodeColumnID(id interfaces.ColumnID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

// DecodeColumnID decodes a byte slice into an interfaces.ColumnID.
func DecodeColumnID(b []byte) (interfaces.ColumnID, error) {
	if len(b) != 8 {
		return 0, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			fmt.Sprintf("invalid byte slice length for column ID: expected 8, got %d", len(b)), nil)
	}
	return interfaces.ColumnID(binary.BigEndian.Uint64(b)), nil
}

// EncodeTimestamp encodes a time.Time into a byte slice (UnixNano).
func EncodeTimestamp(t time.Time) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t.UnixNano()))
	return b
}

// DecodeTimestamp decodes a byte slice into a time.Time (UnixNano).
func DecodeTimestamp(b []byte) (time.Time, error) {
	if len(b) != 8 {
		return time.Time{}, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeInvalidKeyFormat,
			fmt.Sprintf("invalid byte slice length for timestamp: expected 8, got %d", len(b)), nil)
	}
	return time.Unix(0, int64(binary.BigEndian.Uint64(b))), nil
}

//Personal.AI order the ending
