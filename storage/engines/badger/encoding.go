package badger

import (
	"bytes"
	"fmt"
	"time"

	"github.com/dolthub/go-mysql-server/sql" // For sql.Type and potentially sql.Row structure

	"github.com/turtacn/guocedb/common/errors" // Core sortable encoding functions
	"github.com/turtacn/guocedb/interfaces"    // For Schema, ColumnDefinition, IndexDefinition
	"github.com/turtacn/guocedb/internal/encoding"
	// "guocedb/common/types/value" // If you have a custom value type system, import it here.
	// For now, we'll work with interface{} and sql.Type.
)

// Define Badger key prefixes to separate different data categories.
// These ensure that keys for different types of data don't collide and
// allow for efficient prefix scans.
const (
	// MetaPrefix is used for all metadata keys (DBs, Tables, Indexes, Sequences, Name mappings).
	MetaPrefix byte = 0x01
	// TableDataPrefix is used for keys storing actual table row data.
	TableDataPrefix byte = 0x02
	// IndexDataPrefix is used for keys storing index entries.
	IndexDataPrefix byte = 0x03

	// --- Meta Sub-types ---
	// Used within MetaPrefix keys to distinguish different kinds of metadata.
	metaSequenceTag byte = 's' // Sequence counters (e.g., for generating IDs)
	metaDbTag       byte = 'd' // Database metadata marker
	metaTableTag    byte = 't' // Table metadata marker (schema)
	metaIndexTag    byte = 'i' // Index metadata marker (definition)
	metaNameMapTag  byte = 'n' // Name-to-ID mapping marker

	// --- Sequence Types ---
	// Used with metaSequenceTag.
	seqTypeDatabase byte = 'D' // Sequence for Database IDs
	seqTypeTable    byte = 'T' // Sequence for Table IDs (per DB)
	seqTypeIndex    byte = 'I' // Sequence for Index IDs (per Table)

	// --- Name Mapping Types ---
	// Used with metaNameMapTag.
	nameMapTypeDatabase byte = 'D' // Maps DB Name -> DB ID
	nameMapTypeTable    byte = 'T' // Maps Table Name -> Table ID (per DB)
	nameMapTypeIndex    byte = 'I' // Maps Index Name -> Index ID (per Table)

	// --- Null Flag ---
	// Used in EncodeValue/DecodeValue to handle NULLs in a sortable way.
	nullFlag    byte = 0x00 // Represents SQL NULL
	nonNullFlag byte = 0x01 // Precedes a non-NULL encoded value
)

// DatabaseID, TableID, IndexID are assumed to be uint64 for encoding purposes.
type DatabaseID = uint64
type TableID = uint64
type IndexID = uint64

// ==============================================================================
// Metadata Key Encoding
// ==============================================================================

// --- Sequence Keys ---

// EncodeSequenceKey creates a key for fetching/updating sequence counters.
// Example: MetaPrefix | metaSequenceTag | seqTypeDatabase -> value=lastDbID
// Example: MetaPrefix | metaSequenceTag | dbID | seqTypeTable -> value=lastTableID
// Example: MetaPrefix | metaSequenceTag | dbID | tableID | seqTypeIndex -> value=lastIndexID
func EncodeSequenceKey(seqType byte, dbID DatabaseID, tableID TableID) []byte {
	var key []byte
	switch seqType {
	case seqTypeDatabase:
		key = make([]byte, 0, 1+1+1)
		key = append(key, MetaPrefix, metaSequenceTag, seqTypeDatabase)
	case seqTypeTable:
		key = make([]byte, 0, 1+1+8+1)
		key = append(key, MetaPrefix, metaSequenceTag)
		key = encoding.EncodeUint64(key, dbID)
		key = append(key, seqTypeTable)
	case seqTypeIndex:
		key = make([]byte, 0, 1+1+8+8+1)
		key = append(key, MetaPrefix, metaSequenceTag)
		key = encoding.EncodeUint64(key, dbID)
		key = encoding.EncodeUint64(key, tableID)
		key = append(key, seqTypeIndex)
	default:
		// Should not happen with controlled usage
		panic(fmt.Sprintf("unknown sequence type: %c", seqType))
	}
	return key
}

// --- Name Mapping Keys ---

// EncodeNameMapDbKey creates a key for mapping database name to database ID.
// Key: MetaPrefix | metaNameMapTag | nameMapTypeDatabase | dbName -> value=dbID
func EncodeNameMapDbKey(dbName string) []byte {
	key := make([]byte, 0, 1+1+1+len(dbName)+1) // Estimate size
	key = append(key, MetaPrefix, metaNameMapTag, nameMapTypeDatabase)
	key = encoding.EncodeString(key, dbName) // EncodeString includes terminator
	return key
}

// EncodeNameMapTableKey creates a key for mapping table name to table ID within a database.
// Key: MetaPrefix | metaNameMapTag | dbID | nameMapTypeTable | tableName -> value=tableID
func EncodeNameMapTableKey(dbID DatabaseID, tableName string) []byte {
	key := make([]byte, 0, 1+1+8+1+len(tableName)+1) // Estimate size
	key = append(key, MetaPrefix, metaNameMapTag)
	key = encoding.EncodeUint64(key, dbID)
	key = append(key, nameMapTypeTable)
	key = encoding.EncodeString(key, tableName)
	return key
}

// EncodeNameMapIndexKey creates a key for mapping index name to index ID within a table.
// Key: MetaPrefix | metaNameMapTag | dbID | tableID | nameMapTypeIndex | indexName -> value=indexID
func EncodeNameMapIndexKey(dbID DatabaseID, tableID TableID, indexName string) []byte {
	key := make([]byte, 0, 1+1+8+8+1+len(indexName)+1) // Estimate size
	key = append(key, MetaPrefix, metaNameMapTag)
	key = encoding.EncodeUint64(key, dbID)
	key = encoding.EncodeUint64(key, tableID)
	key = append(key, nameMapTypeIndex)
	key = encoding.EncodeString(key, indexName)
	return key
}

// --- Schema/Definition Keys ---

// EncodeDbMetaKey creates a key for storing database-level metadata (currently unused, placeholder).
// Key: MetaPrefix | metaDbTag | dbID -> value=encodedDbMetadata
func EncodeDbMetaKey(dbID DatabaseID) []byte {
	key := make([]byte, 0, 1+1+8)
	key = append(key, MetaPrefix, metaDbTag)
	key = encoding.EncodeUint64(key, dbID)
	return key
}

// EncodeTableSchemaKey creates the key for storing a table's schema definition.
// Key: MetaPrefix | metaTableTag | dbID | tableID -> value=encodedSchemaData
func EncodeTableSchemaKey(dbID DatabaseID, tableID TableID) []byte {
	key := make([]byte, 0, 1+1+8+8)
	key = append(key, MetaPrefix, metaTableTag)
	key = encoding.EncodeUint64(key, dbID)
	key = encoding.EncodeUint64(key, tableID)
	return key
}

// EncodeIndexDefinitionKey creates the key for storing an index's definition.
// Key: MetaPrefix | metaIndexTag | dbID | tableID | indexID -> value=encodedIndexDefinitionData
func EncodeIndexDefinitionKey(dbID DatabaseID, tableID TableID, indexID IndexID) []byte {
	key := make([]byte, 0, 1+1+8+8+8)
	key = append(key, MetaPrefix, metaIndexTag)
	key = encoding.EncodeUint64(key, dbID)
	key = encoding.EncodeUint64(key, tableID)
	key = encoding.EncodeUint64(key, indexID)
	return key
}

// ==============================================================================
// Value Encoding (SQL Types)
// ==============================================================================

// EncodeValue encodes a single Go value (representing an SQL value) into a sortable byte slice.
// It handles NULL values by prefixing the encoded data with a nullFlag or nonNullFlag.
// Uses the functions from internal/encoding for the actual type-specific encoding.
func EncodeValue(buf []byte, value interface{}, typ sql.Type) ([]byte, error) {
	if value == nil {
		return append(buf, nullFlag), nil // Prepend null flag for NULL values
	}

	// Prepend non-null flag for actual values
	buf = append(buf, nonNullFlag)

	var err error
	switch typ.Type() { // Use Type() to get the base type kind
	case sql.Int8, sql.Int16, sql.Int32, sql.Int64, sql.Bit: // Treat Bit as int64 for simplicity
		v, ok := sql.Int64.Convert(value)
		if ok != nil || v == nil {
			return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to int64 for encoding", value)
		}
		buf = encoding.EncodeInt64(buf, v.(int64))
	case sql.Uint8, sql.Uint16, sql.Uint32, sql.Uint64:
		v, ok := sql.Uint64.Convert(value)
		if ok != nil || v == nil {
			return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to uint64 for encoding", value)
		}
		buf = encoding.EncodeUint64(buf, v.(uint64))
	case sql.Float32, sql.Float64:
		v, ok := sql.Float64.Convert(value)
		if ok != nil || v == nil {
			return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to float64 for encoding", value)
		}
		buf = encoding.EncodeFloat64(buf, v.(float64))
	case sql.Timestamp, sql.Datetime, sql.Date: // Encode all as timestamp (int64 nanoseconds)
		v, ok := sql.Timestamp.Convert(value)
		if ok != nil || v == nil {
			return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to time.Time for encoding", value)
		}
		buf = encoding.EncodeTime(buf, v.(time.Time))
	case sql.Char, sql.Varchar, sql.Text, sql.LongText, sql.MediumText, sql.TinyText,
		sql.Enum, sql.Set, sql.JSON, sql.Blob, sql.LongBlob, sql.MediumBlob, sql.TinyBlob, sql.Binary, sql.VarBinary:
		// Treat all string/byte types using EncodeString for sortability and null byte handling
		v, ok := sql.LongText.Convert(value) // Convert to string
		if ok != nil || v == nil {
			// Try converting to bytes if string fails (for BLOB types)
			bVal, bOk := sql.LongBlob.Convert(value)
			if bOk != nil || bVal == nil {
				return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to string or []byte for encoding", value)
			}
			// Encode bytes as string - assumes UTF-8 compatibility or careful handling
			buf = encoding.EncodeString(buf, string(bVal.([]byte)))
		} else {
			buf = encoding.EncodeString(buf, v.(string))
		}
	case sql.Boolean, sql.Bool: // sql.Bool is deprecated but handle it
		v, ok := sql.Boolean.Convert(value)
		if ok != nil || v == nil {
			return buf[:len(buf)-1], errors.Newf(errors.ErrCodeConversion, "failed to convert %v to bool for encoding", value)
		}
		// EncodeBool doesn't return error, just appends 0x00 or 0x01
		buf = encoding.EncodeBool(buf, v.(int8) == 1) // sql.Boolean uses int8(1) for true
	case sql.Null:
		// This case should be handled by the initial value == nil check
		buf = buf[:len(buf)-1] // Remove the nonNullFlag added earlier
		buf = append(buf, nullFlag)
	// Add cases for Decimal, Year, Time, Geometry etc. if needed
	default:
		err = errors.Newf(errors.ErrCodeUnsupported, "unsupported type for badger encoding: %v", typ.String())
		buf = buf[:len(buf)-1] // Remove the nonNullFlag
	}

	return buf, err
}

// DecodeValue decodes a single value from the beginning of the byte slice.
// It checks the null flag first. If not null, it uses the appropriate
// function from internal/encoding based on the expected sql.Type.
// Returns the decoded value, the remaining buffer, and any error.
func DecodeValue(buf []byte, typ sql.Type) (interface{}, []byte, error) {
	if len(buf) == 0 {
		return nil, buf, errors.New(errors.ErrCodeSerialization, "cannot decode value from empty buffer")
	}

	flag := buf[0]
	buf = buf[1:] // Consume the flag

	if flag == nullFlag {
		return nil, buf, nil // SQL NULL value
	}

	if flag != nonNullFlag {
		return nil, buf, errors.Newf(errors.ErrCodeSerialization, "invalid null flag encountered: expected %x or %x, got %x", nullFlag, nonNullFlag, flag)
	}

	// Value is not NULL, proceed with type-specific decoding
	var val interface{}
	var remainder []byte
	var err error

	switch typ.Type() {
	case sql.Int8, sql.Int16, sql.Int32, sql.Int64, sql.Bit:
		var i64 int64
		i64, remainder, err = encoding.DecodeInt64(buf)
		// Convert back to the specific target type if needed, though go-mysql-server often handles int64
		val, err = typ.Convert(i64)
		if err != nil { // Handle conversion error after successful decode
			err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded int64 %d back to %s", i64, typ.String())
		}
	case sql.Uint8, sql.Uint16, sql.Uint32, sql.Uint64:
		var u64 uint64
		u64, remainder, err = encoding.DecodeUint64(buf)
		val, err = typ.Convert(u64)
		if err != nil {
			err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded uint64 %d back to %s", u64, typ.String())
		}
	case sql.Float32, sql.Float64:
		var f64 float64
		f64, remainder, err = encoding.DecodeFloat64(buf)
		val, err = typ.Convert(f64)
		if err != nil {
			err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded float64 %f back to %s", f64, typ.String())
		}
	case sql.Timestamp, sql.Datetime, sql.Date:
		var t time.Time
		t, remainder, err = encoding.DecodeTime(buf)
		val, err = typ.Convert(t) // Convert to specific time type (Timestamp, Datetime, Date)
		if err != nil {
			err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded time %v back to %s", t, typ.String())
		}
	case sql.Char, sql.Varchar, sql.Text, sql.LongText, sql.MediumText, sql.TinyText,
		sql.Enum, sql.Set, sql.JSON:
		var s string
		s, remainder, err = encoding.DecodeString(buf)
		val, err = typ.Convert(s) // Convert to specific string type if necessary (rarely needed)
		if err != nil {
			err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded string back to %s", typ.String())
		}
	case sql.Blob, sql.LongBlob, sql.MediumBlob, sql.TinyBlob, sql.Binary, sql.VarBinary:
		// Decode as string first, then convert to []byte for blob types
		var s string
		s, remainder, err = encoding.DecodeString(buf)
		if err == nil {
			val, err = typ.Convert([]byte(s)) // Convert string -> []byte -> target blob type
			if err != nil {
				err = errors.Wrapf(err, errors.ErrCodeConversion, "failed to convert decoded bytes back to %s", typ.String())
			}
		}
	case sql.Boolean, sql.Bool:
		var b bool
		b, remainder, err = encoding.DecodeBool(buf)
		if err == nil {
			if b {
				val = int8(1) // Convert bool true to int8(1) for sql.Boolean
			} else {
				val = int8(0) // Convert bool false to int8(0)
			}
			// Optional: Convert back to exact type if needed, though int8 often works
			// val, err = typ.Convert(val)
		}
	case sql.Null:
		// Handled by the nullFlag check above
		val = nil
	// Add cases for Decimal, Year, Time, Geometry etc. if needed
	default:
		err = errors.Newf(errors.ErrCodeUnsupported, "unsupported type for badger decoding: %v", typ.String())
	}

	if err != nil {
		// Don't return remainder if decoding failed partway
		return nil, buf, err // Return original buffer slice before flag consumption
	}

	return val, remainder, nil
}

// ==============================================================================
// Row Encoding
// ==============================================================================

// EncodeRow encodes a sql.Row (slice of interface{}) into a single byte slice.
// It iterates through the row values, encoding each one using EncodeValue based on the schema.
// The resulting byte slice is a concatenation of the individually encoded values.
func EncodeRow(buf []byte, row sql.Row, schema interfaces.Schema) ([]byte, error) {
	cols := schema.Columns()
	if len(row) != len(cols) {
		return buf, errors.Newf(errors.ErrCodeInternal, "row length (%d) does not match schema column count (%d)", len(row), len(cols))
	}

	var err error
	for i, val := range row {
		colDef := cols[i]
		buf, err = EncodeValue(buf, val, colDef.Type())
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to encode value at column %d ('%s')", i, colDef.Name())
		}
	}
	return buf, nil
}

// DecodeRow decodes a byte slice back into a sql.Row based on the provided schema.
// It iteratively calls DecodeValue for each column defined in the schema.
func DecodeRow(data []byte, schema interfaces.Schema) (sql.Row, error) {
	cols := schema.Columns()
	row := make(sql.Row, len(cols))
	remainder := data
	var err error
	var val interface{}

	for i, colDef := range cols {
		if len(remainder) == 0 && i < len(cols) {
			// This can happen if the stored data is truncated or schema changed incompatibly
			return nil, errors.Newf(errors.ErrCodeSerialization, "unexpected end of row data while decoding column %d ('%s'), data length %d", i, colDef.Name(), len(data))
		}

		val, remainder, err = DecodeValue(remainder, colDef.Type())
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to decode value for column %d ('%s')", i, colDef.Name())
		}
		row[i] = val
	}

	// Optional: Check if there's unexpected trailing data
	// if len(remainder) > 0 {
	//     log.Warnf("Trailing data (%d bytes) remaining after decoding row for table '%s'", len(remainder), schema.Name())
	// }

	return row, nil
}

// ==============================================================================
// Table Data Key Encoding
// ==============================================================================

// EncodeTableKeyPrefix creates the common prefix for all rows within a specific table.
// Prefix: TableDataPrefix | dbID | tableID |
func EncodeTableKeyPrefix(dbID DatabaseID, tableID TableID) []byte {
	key := make([]byte, 0, 1+8+8)
	key = append(key, TableDataPrefix)
	key = encoding.EncodeUint64(key, dbID)
	key = encoding.EncodeUint64(key, tableID)
	return key
}

// EncodeTableKey creates the full key for a specific row in a table.
// Key: TableDataPrefix | dbID | tableID | encodedPKCol1 | encodedPKCol2 | ...
// The primary key values are extracted from the row based on the schema's PK definition.
func EncodeTableKey(keyBuf []byte, dbID DatabaseID, tableID TableID, row sql.Row, schema interfaces.Schema) ([]byte, error) {
	prefix := EncodeTableKeyPrefix(dbID, tableID)
	keyBuf = append(keyBuf[:0], prefix...) // Start with the prefix

	pkIndexes := schema.PrimaryKeyColumnIndexes()
	if len(pkIndexes) == 0 {
		return nil, errors.Newf(errors.ErrCodeInternal, "cannot encode table key for table '%s' without a primary key", schema.Name())
	}

	cols := schema.Columns()
	var err error
	for _, pkIndex := range pkIndexes {
		if pkIndex >= len(row) || pkIndex >= len(cols) {
			return nil, errors.Newf(errors.ErrCodeInternal, "primary key index %d out of bounds for row/schema length %d/%d", pkIndex, len(row), len(cols))
		}
		pkVal := row[pkIndex]
		pkCol := cols[pkIndex]
		keyBuf, err = EncodeValue(keyBuf, pkVal, pkCol.Type())
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to encode primary key column '%s'", pkCol.Name())
		}
	}

	return keyBuf, nil
}

// DecodeTableKeyPK extracts the primary key values from a full table data key.
// It assumes the key starts with the standard prefix and decodes values based on the PK schema.
func DecodeTableKeyPK(key []byte, schema interfaces.Schema) (sql.Row, error) {
	pkIndexes := schema.PrimaryKeyColumnIndexes()
	if len(pkIndexes) == 0 {
		return nil, errors.Newf(errors.ErrCodeInternal, "cannot decode primary key from table '%s' without PK definition", schema.Name())
	}

	// Calculate expected prefix length
	prefixLen := 1 + 8 + 8 // TableDataPrefix + dbID + tableID
	if len(key) <= prefixLen {
		return nil, errors.Newf(errors.ErrCodeSerialization, "key too short to contain primary key data (len %d, prefix %d)", len(key), prefixLen)
	}

	remainder := key[prefixLen:]
	pkRow := make(sql.Row, len(pkIndexes))
	cols := schema.Columns()
	var err error
	var pkVal interface{}

	for i, pkIndex := range pkIndexes {
		if pkIndex >= len(cols) {
			return nil, errors.Newf(errors.ErrCodeInternal, "schema PK index %d out of bounds for schema length %d", pkIndex, len(cols))
		}
		pkCol := cols[pkIndex]
		pkVal, remainder, err = DecodeValue(remainder, pkCol.Type())
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to decode primary key column '%s' from key", pkCol.Name())
		}
		pkRow[i] = pkVal
	}

	// We don't check for remaining bytes here, as the key structure is fixed.

	return pkRow, nil
}

// ==============================================================================
// Index Data Key Encoding
// ==============================================================================

// EncodeIndexKeyPrefix creates the common prefix for all entries within a specific index.
// Prefix: IndexDataPrefix | dbID | tableID | indexID |
func EncodeIndexKeyPrefix(dbID DatabaseID, tableID TableID, indexID IndexID) []byte {
	key := make([]byte, 0, 1+8+8+8)
	key = append(key, IndexDataPrefix)
	key = encoding.EncodeUint64(key, dbID)
	key = encoding.EncodeUint64(key, tableID)
	key = encoding.EncodeUint64(key, indexID)
	return key
}

// extractValuesByNames extracts values from a row corresponding to a list of column names.
func extractValuesByNames(row sql.Row, schema interfaces.Schema, colNames []string) ([]interface{}, []sql.Type, error) {
	values := make([]interface{}, len(colNames))
	types := make([]sql.Type, len(colNames))
	cols := schema.Columns()

	for i, name := range colNames {
		found := false
		for j, colDef := range cols {
			// Case-insensitive comparison might be needed depending on SQL standard / GMS behavior
			if colDef.Name() == name {
				if j >= len(row) {
					return nil, nil, errors.Newf(errors.ErrCodeInternal, "column index %d for name '%s' is out of bounds for row length %d", j, name, len(row))
				}
				values[i] = row[j]
				types[i] = colDef.Type()
				found = true
				break
			}
		}
		if !found {
			return nil, nil, errors.Newf(errors.ErrCodeNotFound, "column '%s' not found in schema for index key encoding", name)
		}
	}
	return values, types, nil
}

// EncodeIndexKey creates the full key for a specific index entry.
// Key (Unique): IndexDataPrefix | dbID | tableID | indexID | encodedIdxVal1 | ... -> value=encodedPK
// Key (NonUnique): IndexDataPrefix | dbID | tableID | indexID | encodedIdxVal1 | ... | encodedPKVal1 | ... -> value=empty
// The row is used to extract both index column values and primary key values.
func EncodeIndexKey(keyBuf []byte, dbID DatabaseID, tableID TableID, indexID IndexID, row sql.Row, schema interfaces.Schema, indexDef interfaces.IndexDefinition) ([]byte, []byte, error) {
	prefix := EncodeIndexKeyPrefix(dbID, tableID, indexID)
	keyBuf = append(keyBuf[:0], prefix...) // Start with the prefix

	// 1. Encode Index Values
	indexColNames := indexDef.ColumnNames()
	indexValues, indexTypes, err := extractValuesByNames(row, schema, indexColNames)
	if err != nil {
		return nil, nil, errors.Wrapf(err, errors.ErrCodeInternal, "failed to extract index values for index '%s'", indexDef.Name())
	}

	for i, idxVal := range indexValues {
		keyBuf, err = EncodeValue(keyBuf, idxVal, indexTypes[i])
		if err != nil {
			return nil, nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to encode index column '%s' for index '%s'", indexColNames[i], indexDef.Name())
		}
	}

	// 2. Encode Primary Key Values
	pkIndexes := schema.PrimaryKeyColumnIndexes()
	if len(pkIndexes) == 0 {
		return nil, nil, errors.Newf(errors.ErrCodeInternal, "cannot encode index key for table '%s' without a primary key", schema.Name())
	}
	pkValues := make([]interface{}, len(pkIndexes))
	pkTypes := make([]sql.Type, len(pkIndexes))
	cols := schema.Columns()
	for i, pkIdx := range pkIndexes {
		if pkIdx >= len(row) || pkIdx >= len(cols) {
			return nil, nil, errors.Newf(errors.ErrCodeInternal, "primary key index %d out of bounds for row/schema length %d/%d", pkIdx, len(row), len(cols))
		}
		pkValues[i] = row[pkIdx]
		pkTypes[i] = cols[pkIdx].Type()
	}

	// Decide where to put PK: in key for non-unique, in value for unique
	if indexDef.IsUnique() {
		// Encode PK into the value
		valueBuf := make([]byte, 0, 64) // Estimate PK size
		for i, pkVal := range pkValues {
			valueBuf, err = EncodeValue(valueBuf, pkVal, pkTypes[i])
			if err != nil {
				return nil, nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to encode primary key column (index value) '%s'", cols[pkIndexes[i]].Name())
			}
		}
		return keyBuf, valueBuf, nil
	} else {
		// Append PK to the key for non-unique indexes to ensure key uniqueness
		// and allow iteration over duplicates for a given index value set.
		for i, pkVal := range pkValues {
			keyBuf, err = EncodeValue(keyBuf, pkVal, pkTypes[i])
			if err != nil {
				return nil, nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to encode primary key column (index key suffix) '%s'", cols[pkIndexes[i]].Name())
			}
		}
		// Value can be empty for non-unique indexes where PK is in the key
		return keyBuf, []byte{}, nil
	}
}

// DecodeIndexKeyPK extracts the primary key values from an index key or value.
// For unique indexes, it decodes the PK from the `valueBytes`.
// For non-unique indexes, it decodes the PK from the suffix of the `keyBytes`
// (after the index values).
func DecodeIndexKeyPK(keyBytes, valueBytes []byte, schema interfaces.Schema, indexDef interfaces.IndexDefinition) (sql.Row, error) {
	pkIndexes := schema.PrimaryKeyColumnIndexes()
	if len(pkIndexes) == 0 {
		return nil, errors.Newf(errors.ErrCodeInternal, "cannot decode primary key from index '%s' for table '%s' without PK definition", indexDef.Name(), schema.Name())
	}

	pkRow := make(sql.Row, len(pkIndexes))
	cols := schema.Columns()
	var remainder []byte
	var err error

	if indexDef.IsUnique() {
		// PK is entirely in the value
		remainder = valueBytes
		if len(remainder) == 0 {
			return nil, errors.Newf(errors.ErrCodeSerialization, "expected primary key in value for unique index '%s', but value is empty", indexDef.Name())
		}
	} else {
		// PK is in the key suffix. Need to skip the prefix and index values.
		prefixLen := 1 + 8 + 8 + 8 // IndexDataPrefix + dbID + tableID + indexID
		if len(keyBytes) <= prefixLen {
			return nil, errors.Newf(errors.ErrCodeSerialization, "index key too short (len %d) to contain index/pk data for non-unique index '%s'", len(keyBytes), indexDef.Name())
		}
		keyRemainder := keyBytes[prefixLen:]

		// Decode and discard index values to find where PK starts
		indexColNames := indexDef.ColumnNames()
		tempSchemaCols := make([]interfaces.ColumnDefinition, len(indexColNames))
		foundCount := 0
		for i, name := range indexColNames {
			for _, colDef := range cols {
				if colDef.Name() == name { // Case-sensitive match assumed here
					tempSchemaCols[i] = colDef
					foundCount++
					break
				}
			}
		}
		if foundCount != len(indexColNames) {
			return nil, errors.Newf(errors.ErrCodeInternal, "could not find all index columns in schema to decode non-unique index key for '%s'", indexDef.Name())
		}

		var discardVal interface{}
		for _, colDef := range tempSchemaCols {
			if len(keyRemainder) == 0 {
				return nil, errors.Newf(errors.ErrCodeSerialization, "ran out of key bytes while skipping index values for non-unique index '%s'", indexDef.Name())
			}
			discardVal, keyRemainder, err = DecodeValue(keyRemainder, colDef.Type())
			if err != nil {
				return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "error skipping index value for column '%s' in non-unique index key for '%s'", colDef.Name(), indexDef.Name())
			}
		}
		// Now, keyRemainder should point to the start of the encoded PK
		remainder = keyRemainder
		if len(remainder) == 0 {
			return nil, errors.Newf(errors.ErrCodeSerialization, "expected primary key suffix in key for non-unique index '%s', but reached end of key", indexDef.Name())
		}
	}

	// Decode the PK values from the identified remainder (either valueBytes or key suffix)
	var pkVal interface{}
	for i, pkIdx := range pkIndexes {
		if pkIdx >= len(cols) {
			return nil, errors.Newf(errors.ErrCodeInternal, "schema PK index %d out of bounds for schema length %d", pkIdx, len(cols))
		}
		pkCol := cols[pkIdx]
		if len(remainder) == 0 {
			return nil, errors.Newf(errors.ErrCodeSerialization, "ran out of data while decoding primary key column '%s' for index '%s'", pkCol.Name(), indexDef.Name())
		}
		pkVal, remainder, err = DecodeValue(remainder, pkCol.Type())
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeSerialization, "failed to decode primary key column '%s' from index data for '%s'", pkCol.Name(), indexDef.Name())
		}
		pkRow[i] = pkVal
	}

	return pkRow, nil
}
