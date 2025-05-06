// Package badger contains the implementation of the Badger storage engine for guocedb.
// badger 包包含了 Guocedb 的 Badger 存储引擎实现。
package badger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/internal/encoding" // Assuming internal encoding helpers
)

// Key encoding format:
// Catalog:   NamespaceCatalog + Sep + DBName + Sep + TableName + Sep + MetadataType
// Data:      NamespaceData + Sep + DBName + Sep + TableName + Sep + PrimaryKeyEncoded
// Index:     NamespaceIndex + Sep + DBName + Sep + TableName + Sep + IndexName + Sep + IndexValuesEncoded + Sep + PrimaryKeyEncoded (if not part of indexValues)
//
// Key 编码格式：
// Catalog:   NamespaceCatalog + Sep + DBName + Sep + TableName + Sep + MetadataType
// Data:      NamespaceData + Sep + DBName + Sep + TableName + Sep + PrimaryKeyEncoded (主键编码)
// Index:     NamespaceIndex + Sep + DBName + Sep + TableName + Sep + IndexName + Sep + IndexValuesEncoded (索引值编码) + Sep + PrimaryKeyEncoded (如果主键不是索引值的一部分)

var (
	// Sep is the separator byte used in keys.
	// Sep 是 key 中使用的分隔符字节。
	Sep = constants.KeySeparator

	// NsSep is the namespace separator byte used in keys.
	// NsSep 是 key 中使用的命名空间分隔符字节。
	NsSep = constants.NamespaceSeparator

	// NamespaceCatalogBytes is the byte slice for the catalog namespace.
	// NamespaceCatalogBytes 是用于 catalog 命名空间的字节切片。
	NamespaceCatalogBytes = []byte(constants.NamespaceCatalog)

	// NamespaceDataBytes is the byte slice for the data namespace.
	// NamespaceDataBytes 是用于数据命名空间的字节切片。
	NamespaceDataBytes = []byte(constants.NamespaceData)

	// NamespaceIndexBytes is the byte slice for the index namespace.
	// NamespaceIndexBytes 是用于索引命名空间的字节切片。
	NamespaceIndexBytes = []byte(constants.NamespaceIndex)
)

// Metadata types for catalog keys.
// 用于 catalog key 的元数据类型。
const (
	MetadataTypeTable string = "table" // Table schema definition
	// Add other metadata types like index definitions, database properties, etc.
	// 添加其他元数据类型，如索引定义、数据库属性等。
)

// EncodeCatalogKey encodes components into a catalog key.
// EncodeCatalogKey 将组件编码为 catalog key。
func EncodeCatalogKey(dbName, tableName, metadataType string) []byte {
	// catalog:<db_name>:<table_name>:<metadata_type>
	// 使用 NsSep 分隔命名空间，使用 Sep 分隔其他组件。
	return bytes.Join([][]byte{
		NamespaceCatalogBytes,
		[]byte(dbName),
		[]byte(tableName),
		[]byte(metadataType),
	}, []byte{NsSep, Sep, Sep}) // Use NsSep for namespace, Sep for others
}

// DecodeCatalogKey decodes a catalog key back into its components.
// DecodeCatalogKey 将 catalog key 解码回其组件。
func DecodeCatalogKey(key []byte) (dbName, tableName, metadataType string, ok bool) {
	parts := bytes.Split(key, []byte{Sep})
	if len(parts) != 4 {
		return "", "", "", false
	}

	// Check namespace prefix
	nsAndDb := bytes.Split(parts[0], []byte{NsSep})
	if len(nsAndDb) != 2 || !bytes.Equal(nsAndDb[0], NamespaceCatalogBytes) {
		return "", "", "", false
	}

	dbName = string(nsAndDb[1])
	tableName = string(parts[1])
	metadataType = string(parts[2])
	return dbName, tableName, metadataType, true
}

// EncodeDataKey encodes database name, table name, and primary key into a data key.
// Primary key encoding needs to be order-preserving.
// EncodeDataKey 将数据库名、表名和主键编码为数据 key。
// 主键编码需要保证顺序。
func EncodeDataKey(dbName, tableName string, pk sql.Row) ([]byte, error) {
	// data:<db_name>:<table_name>:<PrimaryKeyEncoded>
	// For simplicity, let's assume primary key is a single column (e.g., INT or VARCHAR) for now.
	// A robust implementation needs to handle composite keys and different data types with order-preserving encoding.
	//
	// 为了简化，我们暂时假设主键是单列（例如 INT 或 VARCHAR）。
	// 一个健壮的实现需要处理复合主键和不同数据类型，并使用保证顺序的编码。

	if len(pk) == 0 {
		// This should not happen for tables with primary keys
		return nil, errors.ErrPrimaryKeyRequired.New(tableName)
	}

	// Placeholder for robust primary key encoding
	// This is a simplified example and will NOT work correctly for all types or composite keys.
	// A real implementation needs `internal/encoding` or similar for order-preserving encoding of sql.Row.
	//
	// 这是一个简化的示例，对于所有类型或复合主键来说是不正确的。
	// 实际实现需要 `internal/encoding` 或类似机制来实现 sql.Row 的顺序编码。
	pkEncoded, err := encoding.EncodeRowForKV(pk) // Assuming EncodeRowForKV handles order-preserving encoding
	if err != nil {
		return nil, fmt.Errorf("failed to encode primary key for table %s: %w", tableName, err)
	}

	return bytes.Join([][]byte{
		NamespaceDataBytes,
		[]byte(dbName),
		[]byte(tableName),
		pkEncoded,
	}, []byte{NsSep, Sep, Sep}), nil
}

// DecodeDataKey decodes a data key back into its components.
// Decodes the primary key part into a byte slice.
// DecodeDataKey 将数据 key 解码回其组件。
// 将主键部分解码为字节切片。
func DecodeDataKey(key []byte) (dbName, tableName string, pkEncoded []byte, ok bool) {
	parts := bytes.Split(key, []byte{Sep})
	if len(parts) != 3 {
		return "", "", nil, false
	}

	// Check namespace prefix
	nsAndDb := bytes.Split(parts[0], []byte{NsSep})
	if len(nsAndDb) != 2 || !bytes.Equal(nsAndDb[0], NamespaceDataBytes) {
		return "", "", nil, false
	}

	dbName = string(nsAndDb[1])
	tableName = string(parts[1])
	pkEncoded = parts[2] // Return encoded primary key
	return dbName, tableName, pkEncoded, true
}

// EncodeIndexKey encodes database name, table name, index name, index values, and primary key into an index key.
// Index values and primary key encoding must be order-preserving.
// EncodeIndexKey 将数据库名、表名、索引名、索引值和主键编码为索引 key。
// 索引值和主键编码必须保证顺序。
func EncodeIndexKey(dbName, tableName, indexName string, indexValues, pk sql.Row) ([]byte, error) {
	// index:<db_name>:<table_name>:<index_name>:<IndexValuesEncoded>:<PrimaryKeyEncoded>
	// Need to handle composite index values and ensure order-preserving encoding.
	// If the primary key is part of the index values, we might store less data.
	// For simplicity, we include both encoded index values and encoded primary key.
	//
	// 需要处理复合索引值，并确保顺序编码。
	// 如果主键是索引值的一部分，我们可以存储更少的数据。
	// 为了简化，我们同时包含编码后的索引值和编码后的主键。

	// Placeholder for robust index value and primary key encoding
	// A real implementation needs `internal/encoding` or similar for order-preserving encoding.
	//
	// 这是一个简化的示例，对于所有类型、复合索引或主键来说是不正确的。
	// 实际实现需要 `internal/encoding` 或类似机制来实现顺序编码。
	indexValuesEncoded, err := encoding.EncodeRowForKV(indexValues) // Assuming EncodeRowForKV handles order-preserving encoding
	if err != nil {
		return nil, fmt.Errorf("failed to encode index values for index %s on table %s: %w", indexName, tableName, err)
	}

	pkEncoded, err := encoding.EncodeRowForKV(pk) // Assuming EncodeRowForKV handles order-preserving encoding
	if err != nil {
		return nil, fmt.Errorf("failed to encode primary key for index %s on table %s: %w", indexName, tableName, err)
	}

	return bytes.Join([][]byte{
		NamespaceIndexBytes,
		[]byte(dbName),
		[]byte(tableName),
		[]byte(indexName),
		indexValuesEncoded,
		pkEncoded, // Include PK to handle duplicate index values
	}, []byte{NsSep, Sep, Sep, Sep, Sep}), nil
}

// DecodeIndexKey decodes an index key back into its components.
// Decodes index values and primary key parts into byte slices.
// DecodeIndexKey 将索引 key 解码回其组件。
// 将索引值和主键部分解码为字节切片。
func DecodeIndexKey(key []byte) (dbName, tableName, indexName string, indexValuesEncoded []byte, pkEncoded []byte, ok bool) {
	parts := bytes.Split(key, []byte{Sep})
	if len(parts) != 5 {
		return "", "", "", nil, nil, false
	}

	// Check namespace prefix
	nsAndDb := bytes.Split(parts[0], []byte{NsSep})
	if len(nsAndDb) != 2 || !bytes.Equal(nsAndDb[0], NamespaceIndexBytes) {
		return "", "", "", nil, nil, false
	}

	dbName = string(nsAndDb[1])
	tableName = string(parts[1])
	indexName = string(parts[2])
	indexValuesEncoded = parts[3] // Return encoded index values
	pkEncoded = parts[4]          // Return encoded primary key
	return dbName, tableName, indexValuesEncoded, pkEncoded, true
}

// EncodeRow serializes a sql.Row into a byte slice.
// This needs a robust serialization format (e.g., Gob, Protobuf, or custom binary).
// EncodeRow 将一个 sql.Row 序列化为字节切片。
// 这需要一个健壮的序列化格式（例如 Gob、Protobuf 或自定义二进制格式）。
func EncodeRow(row sql.Row) ([]byte, error) {
	// Placeholder implementation. Replace with actual serialization logic.
	// Consider using a library or a custom binary format that handles different SQL types.
	// The format should ideally include schema information or rely on the table schema for decoding.
	//
	// 占位符实现。替换为实际的序列化逻辑。
	// 考虑使用库或自定义二进制格式来处理不同的 SQL 类型。
	// 格式理想情况下应包含模式信息，或依赖表模式进行解码。
	var buf bytes.Buffer
	for i, val := range row {
		// Very basic example: convert each value to string and join
		// This is NOT suitable for production as it's not robust, efficient, or type-aware.
		// Example: INT(123) -> "123", VARCHAR("abc") -> "abc", NULL -> "" or specific marker
		// A proper implementation needs to handle nil, various types (int, float, string, bool, time, etc.)
		// and potentially variable length encoding.
		//
		// 这是一个非常基础的示例：将每个值转换为字符串并连接。
		// 这不适合生产使用，因为它不健壮、不高效且不感知类型。
		// 示例：INT(123) -> "123", VARCHAR("abc") -> "abc", NULL -> "" 或特定标记
		// 一个适当的实现需要处理 nil、各种类型（int、float、string、bool、time 等）
		// 以及可能的变长编码。

		var valBytes []byte
		if val == nil {
			valBytes = []byte("NULL") // Using a marker for NULL
		} else {
			switch v := val.(type) {
			case int:
				valBytes = []byte(strconv.Itoa(v))
			case int64:
				valBytes = []byte(strconv.FormatInt(v, 10))
			case uint64:
				valBytes = []byte(strconv.FormatUint(v, 10))
			case float64:
				valBytes = []byte(strconv.FormatFloat(v, 'g', -1, 64))
			case string:
				valBytes = []byte(v)
			case bool:
				valBytes = []byte(strconv.FormatBool(v))
			// Add other types as needed (time.Time, []byte, etc.)
			// 根据需要添加其他类型（time.Time, []byte 等）
			default:
				// Fallback using fmt.Sprintf, may not be ideal
				valBytes = []byte(fmt.Sprintf("%v", v))
			}
		}

		buf.Write(valBytes)
		if i < len(row)-1 {
			buf.WriteByte(Sep) // Use the same separator for values within a row (could be different)
		}
	}

	// IMPORTANT: This simple string join is BAD. Use a proper binary encoding.
	// 警告：这种简单的字符串连接是错误的。请使用适当的二进制编码。
	return buf.Bytes(), nil // This is just a placeholder
}

// DecodeRow deserializes a byte slice back into a sql.Row.
// Needs the table schema to interpret the byte slice correctly.
// DecodeRow 将字节切片反序列化回 sql.Row。
// 需要表模式才能正确解释字节切片。
func DecodeRow(data []byte, schema sql.Schema) (sql.Row, error) {
	// Placeholder implementation. Replace with actual deserialization logic
	// that matches the EncodeRow implementation.
	// This needs to read byte slices based on the schema and convert them back to Go types.
	//
	// 占位符实现。替换为与 EncodeRow 实现匹配的实际反序列化逻辑。
	// 这需要根据模式读取字节切片，并将其转换回 Go 类型。

	if len(data) == 0 {
		return sql.Row{}, nil // Handle empty data? Depends on encoding.
	}

	// This simple split logic only works for the placeholder EncodeRow above.
	// A real implementation needs to handle binary formats, different types, and nil values correctly.
	//
	// 这种简单的分割逻辑只适用于上面的占位符 EncodeRow。
	// 实际实现需要正确处理二进制格式、不同类型和 nil 值。
	parts := bytes.Split(data, []byte{Sep})
	if len(parts) != len(schema) {
		// This is a strong indicator of an encoding/decoding mismatch or corrupted data
		return nil, errors.ErrCorruptedData.New(fmt.Sprintf("expected %d columns, got %d parts", len(schema), len(parts)))
	}

	row := make(sql.Row, len(schema))
	for i, part := range parts {
		col := schema[i]
		// Basic type conversion based on schema column type
		// This is highly incomplete and error-prone.
		//
		// 根据模式列类型进行基本类型转换。
		// 这是高度不完整且容易出错的。
		valBytes := part
		if bytes.Equal(valBytes, []byte("NULL")) {
			row[i] = nil
			continue
		}

		switch col.Type.Type() {
		case sql.Int8, sql.Int16, sql.Int32, sql.Int64:
			v, err := strconv.ParseInt(string(valBytes), 10, 64)
			if err != nil {
				// Handle parsing error, maybe return an error or a specific value
				row[i] = nil // Or some error representation
			} else {
				// Need to cast to the correct int type based on col.Type
				switch col.Type.Type() {
				case sql.Int8:
					row[i] = int8(v)
				case sql.Int16:
					row[i] = int16(v)
				case sql.Int32:
					row[i] = int32(v)
				case sql.Int64:
					row[i] = v
				}
			}
		case sql.Uint8, sql.Uint16, sql.Uint32, sql.Uint64:
			v, err := strconv.ParseUint(string(valBytes), 10, 64)
			if err != nil {
				row[i] = nil
			} else {
				switch col.Type.Type() {
				case sql.Uint8:
					row[i] = uint8(v)
				case sql.Uint16:
					row[i] = uint16(v)
				case sql.Uint32:
					row[i] = uint32(v)
				case sql.Uint64:
					row[i] = v
				}
			}
		case sql.Float32, sql.Float64:
			v, err := strconv.ParseFloat(string(valBytes), 64)
			if err != nil {
				row[i] = nil
			} else {
				if col.Type.Type() == sql.Float32 {
					row[i] = float32(v)
				} else {
					row[i] = v
				}
			}
		case sql.Text, sql.Blob, sql.LongBlob, sql.MediumBlob, sql.VarChar, sql.VarcharMax:
			row[i] = string(valBytes) // Assuming string for text/varchar, need to handle bytes for blob
		case sql.Boolean:
			v, err := strconv.ParseBool(string(valBytes))
			if err != nil {
				row[i] = nil
			} else {
				row[i] = v
			}
		// Add other types (sql.Timestamp, sql.Date, etc.)
		// 添加其他类型（sql.Timestamp, sql.Date 等）
		default:
			// Fallback: leave as byte slice or attempt generic conversion
			row[i] = valBytes // Or try fmt.Sscan
		}
	}

	// IMPORTANT: This simple split/convert is BAD. Use a proper binary decoding.
	// 警告：这种简单的分割/转换是错误的。请使用适当的二进制解码。
	return row, nil // This is just a placeholder
}

// --- Internal Helper Functions (Order Preserving Encoding Placeholders) ---
// --- 内部辅助函数（顺序保留编码占位符） ---

// encodeValueOrderPreserving attempts to encode a single value for use in a key, ensuring order.
// This is a placeholder. A real implementation needs to handle all SQL types correctly for order.
// encodeValueOrderPreserving 尝试对单个值进行编码以用于 key，确保顺序。
// 这是一个占位符。实际实现需要正确处理所有 SQL 类型以保证顺序。
func encodeValueOrderPreserving(val interface{}) ([]byte, error) {
	if val == nil {
		return []byte{encoding.NullValueByte}, nil // Use a specific byte for NULL
	}

	switch v := val.(type) {
	case int8:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v)) // Example: Encode as 64-bit integer
		return buf[:], nil
	case int16:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		return buf[:], nil
	case int32:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		return buf[:], nil
	case int64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		return buf[:], nil
	case uint8:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		return buf[:], nil
	case uint16:
		var buf [8]byte
		binary.BigEndian.PutUint66(buf[:], uint64(v))
		return buf[:], nil
	case uint32:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		return buf[:], nil
	case uint64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], v)
		return buf[:], nil
	case float32:
		// Floating point encoding for order preserving is complex.
		// Needs to handle sign bit correctly.
		// Consider using a library or custom logic.
		//
		// 浮点数编码要保证顺序是复杂的。
		// 需要正确处理符号位。
		// 考虑使用库或自定义逻辑。
		return nil, fmt.Errorf("float32 order-preserving encoding not implemented")
	case float64:
		// See comment for float32
		return nil, fmt.Errorf("float64 order-preserving encoding not implemented")
	case string:
		// Simple approach: prefix with length, then bytes.
		// Or null-terminate (risky if strings can contain nulls).
		// A common technique is to use a special termination byte not in the alphabet.
		//
		// 简单方法：前缀加长度，然后是字节。
		// 或者 null 终止（如果字符串包含 null 则有风险）。
		// 一个常用技术是使用不在字母表中的特殊终止字节。
		return []byte(v), nil // Placeholder: direct bytes
	case []byte:
		// Similar considerations as string
		return v, nil // Placeholder: direct bytes
	case bool:
		if v {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	// Add other types (time.Time, decimal, etc.)
	// 添加其他类型（time.Time, decimal 等）
	default:
		return nil, fmt.Errorf("unsupported type for order-preserving encoding: %T", v)
	}
}

// decodeValueOrderPreserving attempts to decode a byte slice back into a single value.
// Needs to know the expected type. This is a placeholder.
// decodeValueOrderPreserving 尝试将字节切片解码回单个值。
// 需要知道预期的类型。这是一个占位符。
func decodeValueOrderPreserving(data []byte, expectedType sql.Type) (interface{}, error) {
	if len(data) == 1 && data[0] == encoding.NullValueByte {
		return nil, nil // Decode NULL marker
	}

	// Placeholder implementation. Needs to match encodeValueOrderPreserving logic.
	//
	// 占位符实现。需要与 encodeValueOrderPreserving 逻辑匹配。
	switch expectedType.Type() {
	case sql.Int8, sql.Int16, sql.Int32, sql.Int64:
		if len(data) != 8 {
			return nil, fmt.Errorf("invalid encoded integer length: %d", len(data))
		}
		val := int64(binary.BigEndian.Uint64(data))
		// Need to cast based on expectedType
		switch expectedType.Type() {
		case sql.Int8:
			return int8(val), nil
		case sql.Int16:
			return int16(val), nil
		case sql.Int32:
			return int32(val), nil
		case sql.Int64:
			return val, nil
		}
	case sql.Uint8, sql.Uint16, sql.Uint32, sql.Uint64:
		if len(data) != 8 {
			return nil, fmt.Errorf("invalid encoded unsigned integer length: %d", len(data))
		}
		val := binary.BigEndian.Uint64(data)
		// Need to cast based on expectedType
		switch expectedType.Type() {
		case sql.Uint8:
			return uint8(val), nil
		case sql.Uint16:
			return uint16(val), nil
		case sql.Uint32:
			return uint32(val), nil
		case sql.Uint64:
			return val, nil
		}
	case sql.Float32, sql.Float64:
		v, err := strconv.ParseFloat(string(data), 64) // Placeholder: string conversion
		if err != nil {
			return nil, fmt.Errorf("failed to parse float: %w", err)
		}
		if expectedType.Type() == sql.Float32 {
			return float32(v), nil
		} else {
			return v, nil
		}
	case sql.Text, sql.Blob, sql.LongBlob, sql.MediumBlob, sql.VarChar, sql.VarcharMax:
		return string(data), nil // Placeholder: assuming string for text/varchar, need to handle bytes for blob
	case sql.Boolean:
		if len(data) != 1 {
			return nil, fmt.Errorf("invalid encoded boolean length: %d", len(data))
		}
		return data[0] != 0, nil
	// Add other types (sql.Timestamp, sql.Date, etc.)
	// 添加其他类型（sql.Timestamp, sql.Date 等）
	default:
		return nil, fmt.Errorf("unsupported type for order-preserving decoding: %s", expectedType.Type().String())
	}
	return nil, fmt.Errorf("unknown type or decoding error") // Should not reach here
}

// --- Internal Helper Functions (Used by the above, potentially in internal/encoding) ---
// These functions are assumed to exist in internal/encoding and handle the actual order-preserving encoding/decoding of rows for KV keys.
// --- 内部辅助函数（被上面函数使用，可能在 internal/encoding 中） ---
// 这些函数假定存在于 internal/encoding 中，处理 sql.Row 用于 KV key 的实际顺序编码/解码。
//
// func EncodeRowForKV(row sql.Row) ([]byte, error) { ... } // Needs order-preserving encoding of multiple values
// func DecodeRowFromKV(data []byte, schema sql.Schema) (sql.Row, error) { ... } // Needs order-preserving decoding

// For the purpose of compilation in this initial phase, let's add dummy implementations
// until internal/encoding is generated.
// 为了在此初始阶段保证编译通过，我们添加虚拟实现，直到 internal/encoding 生成。

// Assuming this function exists in internal/encoding
// 假设此函数存在于 internal/encoding 中
func init() {
	// Replace this dummy implementation with the real one from internal/encoding
	// 用来自 internal/encoding 的真实实现替换此虚拟实现
	encoding.EncodeRowForKV = func(row sql.Row) ([]byte, error) {
		// Dummy implementation: just join string representations. NOT order-preserving.
		// 虚拟实现：仅连接字符串表示。不保证顺序。
		var parts []string
		for _, val := range row {
			if val == nil {
				parts = append(parts, "NULL")
			} else {
				parts = append(parts, fmt.Sprintf("%v", val))
			}
		}
		// Using a comma as a dummy separator for the dummy implementation
		// 在虚拟实现中使用逗号作为虚拟分隔符
		return []byte(bytes.Join([][]byte{}, []byte(","))), nil // Dummy join
	}
	// Assuming this constant exists in internal/encoding
	// 假设此常量存在于 internal/encoding 中
	encoding.NullValueByte = 0xFF // Dummy NULL marker byte
}

// TODO: Implement robust order-preserving encoding and decoding in internal/encoding package.
// TODO: 在 internal/encoding 包中实现健壮的顺序保留编码和解码。