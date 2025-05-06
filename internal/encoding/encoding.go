// Package encoding provides internal utilities for data encoding and decoding.
// This is used for serializing/deserializing data values and keys for storage.
//
// encoding 包提供用于数据编码和解码的内部工具。
// 它用于存储的序列化/反序列化数据值和 key。
package encoding

import (
	"context"
	"fmt"
	"time"

	"github.com/dolthub/go-mysql-server/sql" // Need GMS SQL types for encoding/decoding
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
)

// Data Encoding Strategy:
// How SQL data types (like INT, VARCHAR, BOOL, FLOAT, DATETIME) are converted into bytes
// for storage in the underlying key-value store (Badger).
//
// 数据编码策略：
// 如何将 SQL 数据类型（如 INT, VARCHAR, BOOL, FLOAT, DATETIME）转换为字节
// 用于存储在底层的键值存储（Badger）中。
// Considerations:
// - Fixed-width vs variable-width types.
// - Endianness for numeric types.
// - Null values representation.
// - Handling complex types (e.g., JSON, BLOBs).
// - Compatibility with potential indexing requirements (prefix scans, range queries).
//
// 考虑因素：
// - 固定宽度与可变宽度类型。
// - 数字类型的字节序。
// - Null 值表示。
// - 处理复杂类型（例如，JSON, BLOBs）。
// - 与潜在索引要求（前缀扫描、范围查询）的兼容性。
// A common approach is to use a serialization library or implement custom logic
// based on type codes and binary representations.

// Key Encoding Strategy:
// How database, table, row, and index information is encoded into keys
// for the underlying key-value store (Badger).
//
// Key 编码策略：
// 如何将数据库、表、行和索引信息编码到 Key 中
// 用于底层的键值存储（Badger）。
// Considerations:
// - Structure: Namespace:Database:Table:Row/IndexSuffix.
// - Separators between parts.
// - Order of parts to enable efficient range scans (e.g., all rows in a table).
// - Encoding of row primary keys or index keys.
// - Handling composite keys and different data types within keys.
//
// 考虑因素：
// - 结构：Namespace:Database:Table:Row/IndexSuffix。
// - 各部分之间的分隔符。
// - 各部分的顺序，以实现高效的范围扫描（例如，表中的所有行）。
// - 行主键或索引 Key 的编码。
// - 处理复合 Key 和 Key 中不同数据类型。
// A common approach is to use a fixed prefix for namespaces, followed by database name, table name,
// and then data-specific or index-specific encoding.

// Placeholder encoding functions.
// These should be implemented to handle actual data types and key structures.
//
// 占位符编码函数。
// 应实现这些函数来处理实际的数据类型和 Key 结构。

// EncodeValue encodes a go-mysql-server sql.Row value into bytes for storage.
// It needs the value and its SQL type definition.
//
// EncodeValue 将 go-mysql-server sql.Row 值编码为字节进行存储。
// 它需要值及其 SQL 类型定义。
func EncodeValue(ctx context.Context, val interface{}, sqlType sql.Type) ([]byte, error) {
	log.Warn("Encoding EncodeValue called (Placeholder)") // 调用 EncodeValue（占位符）。
	// TODO: Implement actual value encoding based on sqlType.
	// Example: For INT, use binary encoding. For VARCHAR, use string bytes. Handle NULLs.
	// 示例：对于 INT，使用二进制编码。对于 VARCHAR，使用字符串字节。处理 NULL。
	// This is complex and type-dependent.
	// 这很复杂，取决于类型。

	// Dummy implementation: attempt string conversion or return error
	// 虚拟实现：尝试字符串转换或返回错误
	switch v := val.(type) {
	case nil:
		// How to represent NULL? Maybe a special byte prefix?
		// 如何表示 NULL？也许是一个特殊的字节前缀？
		return []byte{0x00}, nil // Dummy NULL representation
	case bool:
		if v { return []byte{0x01}, nil } else { return []byte{0x00}, nil } // Dummy bool representation
	case int, int8, int16, int32, int64:
		// Needs binary encoding based on size
		// 需要基于大小的二进制编码
		log.Warn("Encoding integer value as string (Placeholder)") // 将整数值编码为字符串（占位符）。
		return []byte(fmt.Sprintf("%v", v)), nil // Insecure/incorrect for real storage
	case uint, uint8, uint16, uint32, uint64:
		log.Warn("Encoding unsigned integer value as string (Placeholder)") // 将无符号整数值编码为字符串（占位符）。
		return []byte(fmt.Sprintf("%v", v)), nil // Insecure/incorrect for real storage
	case float32, float64:
		// Needs binary encoding (IEEE 754)
		// 需要二进制编码 (IEEE 754)
		log.Warn("Encoding float value as string (Placeholder)") // 将浮点值编码为字符串（占位符）。
		return []byte(fmt.Sprintf("%v", v)), nil // Insecure/incorrect for real storage
	case string:
		return []byte(v), nil // Simple string encoding
	case []byte:
		return v, nil // Simple byte slice encoding
	case time.Time:
		// Needs specific date/time encoding format
		// 需要特定的日期/时间编码格式
		log.Warn("Encoding time.Time value as string (Placeholder)") // 将 time.Time 值编码为字符串（占位符）。
		return []byte(v.Format(time.RFC3339Nano)), nil // Example format
	default:
		log.Error("Unsupported value type for encoding: %T", v) // 编码不支持的值类型。
		return nil, errors.ErrUnsupportedType.New(fmt.Sprintf("encoding value of type %T", v)) // Unsupported type
	}
}

// DecodeValue decodes bytes from storage back into a go-mysql-server sql.Row value.
// It needs the bytes and the expected SQL type definition.
//
// DecodeValue 将存储中的字节解码回 go-mysql-server sql.Row 值。
// 它需要字节和预期的 SQL 类型定义。
func DecodeValue(ctx context.Context, encoded []byte, sqlType sql.Type) (interface{}, error) {
	log.Warn("Encoding DecodeValue called (Placeholder)") // 调用 DecodeValue（占位符）。
	// TODO: Implement actual value decoding based on sqlType.
	// Example: For INT, read binary bytes. For VARCHAR, convert bytes to string. Handle NULLs.
	// 示例：对于 INT，读取二进制字节。对于 VARCHAR，将字节转换为字符串。处理 NULL。
	// This is complex and type-dependent, and must match the encoding logic.
	// 这很复杂，取决于类型，并且必须与编码逻辑匹配。

	// Dummy implementation: attempt string conversion or return error
	// 虚拟实现：尝试字符串转换或返回错误
	if len(encoded) == 1 && encoded[0] == 0x00 {
		// Dummy NULL representation
		// 虚拟 NULL 表示
		return nil, nil
	}

	switch sqlType.Type() { // Use sqlType.Type() to get the underlying type constant
	case sql.Int8, sql.Int16, sql.Int32, sql.Int64:
		// Needs binary decoding
		// 需要二进制解码
		log.Warn("Decoding bytes as string for integer (Placeholder)") // 将字节解码为字符串表示的整数（占位符）。
		// Attempt to parse the string representation from the dummy encoder
		// 尝试从虚拟编码器解析字符串表示
		var i int64
		_, err := fmt.Sscan(string(encoded), &i)
		if err != nil {
			log.Error("Failed to parse string as integer during decoding: %v", err) // 解码时将字符串解析为整数失败。
			return nil, errors.ErrDecodingFailed.New(fmt.Sprintf("decoding integer: %v", err))
		}
		// Need to convert back to the specific integer type requested by sqlType
		// 需要转换回 sqlType 请求的特定整数类型
		switch sqlType.Type() {
		case sql.Int8: return int8(i), nil
		case sql.Int16: return int16(i), nil
		case sql.Int32: return int32(i), nil
		case sql.Int64: return i, nil
		default: // Should not happen
			return nil, errors.ErrInternal.New("unhandled integer type in decoding switch")
		}

	case sql.Uint8, sql.Uint16, sql.Uint32, sql.Uint64:
		log.Warn("Decoding bytes as string for unsigned integer (Placeholder)") // 将字节解码为字符串表示的无符号整数（占位符）。
		var ui uint64
		_, err := fmt.Sscan(string(encoded), &ui)
		if err != nil {
			log.Error("Failed to parse string as unsigned integer during decoding: %v", err) // 解码时将字符串解析为无符号整数失败。
			return nil, errors.ErrDecodingFailed.New(fmt.Sprintf("decoding unsigned integer: %v", err))
		}
		switch sqlType.Type() {
		case sql.Uint8: return uint8(ui), nil
		case sql.Uint16: return uint16(ui), nil
		case sql.Uint32: return uint32(ui), nil
		case sql.Uint64: return ui, nil
		default:
			return nil, errors.ErrInternal.New("unhandled unsigned integer type in decoding switch")
		}

	case sql.Float32, sql.Float64:
		log.Warn("Decoding bytes as string for float (Placeholder)") // 将字节解码为字符串表示的浮点数（占位符）。
		var f float64
		_, err := fmt.Sscan(string(encoded), &f)
		if err != nil {
			log.Error("Failed to parse string as float during decoding: %v", err) // 解码时将字符串解析为浮点数失败。
			return nil, errors.ErrDecodingFailed.New(fmt.Sprintf("decoding float: %v", err))
		}
		switch sqlType.Type() {
		case sql.Float32: return float32(f), nil
		case sql.Float64: return f, nil
		default:
			return nil, errors.ErrInternal.New("unhandled float type in decoding switch")
		}

	case sql.Text, sql.Blob, sql.JSON: // Assuming Text, Blob, JSON are stored as raw bytes or strings
		return string(encoded), nil // Convert bytes to string (assuming UTF-8) or return raw bytes for Blob/JSON

	case sql.Type.Timestamp, sql.Type.Datetime, sql.Type.Date: // Assuming Time types are stored in a specific format
		log.Warn("Decoding bytes as string for time.Time (Placeholder)") // 将字节解码为字符串表示的 time.Time（占位符）。
		t, err := time.Parse(time.RFC3339Nano, string(encoded)) // Example format matching encoder
		if err != nil {
			log.Error("Failed to parse string as time.Time during decoding: %v", err) // 解码时将字符串解析为 time.Time 失败。
			return nil, errors.ErrDecodingFailed.New(fmt.Sprintf("decoding time: %v", err))
		}
		// Need to return the correct time type based on sqlType.Type() (Timestamp, Datetime, Date)
		// 需要根据 sqlType.Type() 返回正确的时间类型（Timestamp, Datetime, Date）
		// sql.Date is often just a date part of time.Time.
		// sql.Date 通常只是 time.Time 的日期部分。
		return t, nil // Return as time.Time for now

	case sql.Boolean:
		// Needs decoding from byte(s)
		// 需要从字节解码
		log.Warn("Decoding bytes for boolean (Placeholder)") // 解码布尔值字节（占位符）。
		if len(encoded) == 1 {
			return encoded[0] != 0x00, nil // Dummy bool decoding
		}
		log.Error("Invalid byte length for boolean decoding: %d", len(encoded)) // 布尔值解码的字节长度无效。
		return false, errors.ErrDecodingFailed.New("invalid byte length for boolean")


	default:
		log.Warn("Unsupported SQL type for decoding: %v (%T)", sqlType, sqlType) // 解码不支持的 SQL 类型。
		return nil, errors.ErrUnsupportedType.New(fmt.Sprintf("decoding value of type %v", sqlType)) // Unsupported type
	}
}

// EncodeKey encodes components into a byte key for storage.
// It defines the key structure for namespaces, databases, tables, rows, and indexes.
//
// EncodeKey 将组件编码为字节 Key 进行存储。
// 它定义了命名空间、数据库、表、行和索引的 Key 结构。
// Key format: <Namespace><Separator><DBName><Separator><TableName><Separator>[Row/Index Specific Encoding]
// Key 格式：<命名空间><分隔符><数据库名><分隔符><表名><分隔符>[行/索引特定编码]
// Separators need to be chosen carefully not to appear in names.
// 分隔符需要谨慎选择，不能出现在名称中。
// Using null byte (0x00) as separator is common, assuming names don't contain null bytes.
// 使用 null 字节 (0x00) 作为分隔符很常见，假设名称不包含 null 字节。
// Namespace bytes are defined in storage/engines/badger/keys.go (or a common place).
// 命名空间字节在 storage/engines/badger/keys.go（或一个通用位置）中定义。

var (
	// KeySeparator is used to separate parts of the key. Null byte is a common choice.
	// KeySeparator 用于分隔 Key 的各部分。Null 字节是常见的选择。
	KeySeparator = []byte{0x00}
)

// EncodeNamespaceKey creates a key prefix for a given namespace.
// EncodeNamespaceKey 为给定的命名空间创建 Key 前缀。
func EncodeNamespaceKey(namespace []byte) []byte {
	return append(namespace, KeySeparator...)
}

// EncodeDatabaseKey creates a key for a database metadata entry.
// Assumes database names don't contain KeySeparator.
//
// EncodeDatabaseKey 为数据库元数据条目创建 Key。
// 假设数据库名称不包含 KeySeparator。
func EncodeDatabaseKey(namespace []byte, dbName string) []byte {
	// Format: <Namespace>:<DBName>
	// 格式：<命名空间>:<数据库名>
	key := append(EncodeNamespaceKey(namespace), []byte(dbName)...)
	return key
}

// EncodeTableKey creates a key for a table metadata entry.
// Assumes database and table names don't contain KeySeparator.
//
// EncodeTableKey 为表元数据条目创建 Key。
// 假设数据库名和表名不包含 KeySeparator。
func EncodeTableKey(namespace []byte, dbName, tableName string) []byte {
	// Format: <Namespace>:<DBName>:<TableName>
	// 格式：<命名空间>:<数据库名>:<表名>
	key := append(EncodeDatabaseKey(namespace, dbName), KeySeparator...)
	key = append(key, []byte(tableName)...)
	return key
}

// EncodeRowKeyPrefix creates a key prefix for all rows in a table.
// Format: <Namespace_Data>:<DBName>:<TableName>:
//
// EncodeRowKeyPrefix 为表中的所有行创建 Key 前缀。
// 格式：<数据命名空间>:<数据库名>:<表名>:
func EncodeRowKeyPrefix(dbName, tableName string) []byte {
	// Use the data namespace
	// 使用数据命名空间
	key := append(EncodeTableKey(NamespaceDataBytes, dbName, tableName), KeySeparator...)
	return key
}


// EncodeRowKey creates a key for a specific row in a table.
// It includes the primary key encoding.
// Format: <Namespace_Data>:<DBName>:<TableName>:<Encoded_PrimaryKey>
//
// EncodeRowKey 为表中的特定行创建 Key。
// 它包括主键编码。
// TODO: Implement encoding of primary key values. This is complex for composite keys.
// TODO: 实现主键值的编码。对于复合 Key，这很复杂。
func EncodeRowKey(dbName, tableName string, primaryKeyValues []interface{}, primaryKeySchema sql.Schema) ([]byte, error) {
	log.Warn("Encoding EncodeRowKey called (Placeholder)") // 调用 EncodeRowKey（占位符）。
	// Get the prefix for the table
	// 获取表的 Key 前缀
	key := EncodeRowKeyPrefix(dbName, tableName)

	// TODO: Implement encoding of primary key values based on their types and order in the schema.
	// Need to ensure primary key encoding is unique and preserves order for efficient range scans if possible.
	//
	// TODO: 根据主键值类型和模式中的顺序实现主键值的编码。
	// 需要确保主键编码是唯一的，并且如果可能，保留顺序以实现高效的范围扫描。
	// For a simple placeholder, just concatenate string representations (NOT for production).
	// 对于一个简单的占位符，只连接字符串表示（不用于生产）。
	pkEncoded := []byte{}
	for i, val := range primaryKeyValues {
		// Encode each primary key value
		// 编码每个主键值
		valBytes, err := EncodeValue(context.Background(), val, primaryKeySchema[i].Type) // Use dummy context/schema
		if err != nil {
			log.Error("Failed to encode primary key value %v for row key: %v", val, err) // 编码主键值失败。
			return nil, fmt.Errorf("failed to encode primary key value: %w", err)
		}
		pkEncoded = append(pkEncoded, valBytes...)
		if i < len(primaryKeyValues)-1 {
			pkEncoded = append(pkEncoded, KeySeparator...) // Use separator between primary key parts
		}
	}

	key = append(key, pkEncoded...)
	log.Debug("Encoded row key for %s.%s with PK %v: %v", dbName, tableName, primaryKeyValues, key) // 编码行 Key。
	return key, nil
}

// DecodeRowKey decodes a row key to extract database, table, and primary key values.
// DecodeRowKey 解码行 Key 以提取数据库、表和主键值。
// TODO: Implement decoding of primary key values, which matches EncodeRowKey logic.
// TODO: 实现主键值的解码，它与 EncodeRowKey 逻辑匹配。
func DecodeRowKey(key []byte) (dbName, tableName string, primaryKeyValues []interface{}, err error) {
	log.Warn("Encoding DecodeRowKey called (Placeholder)") // 调用 DecodeRowKey（占位符）。
	// Split the key by the separator
	// 按分隔符分割 Key
	parts := bytes.Split(key, KeySeparator)

	if len(parts) < 4 { // Expect at least Namespace_Data, DBName, TableName, PrimaryKey part(s)
		log.Error("Invalid row key format: %v", key) // 无效的行 Key 格式。
		return "", "", nil, errors.ErrDecodingFailed.New("invalid row key format")
	}

	// Check namespace
	// 检查命名空间
	if !bytes.Equal(parts[0], NamespaceDataBytes) {
		log.Error("Invalid namespace in row key: %v", parts[0]) // 行 Key 中的命名空间无效。
		return "", "", nil, errors.ErrDecodingFailed.New("invalid namespace in row key")
	}

	dbName = string(parts[1])
	tableName = string(parts[2])

	// The remaining parts are the encoded primary key values.
	// Need the table's primary key schema to decode correctly.
	//
	// 剩余部分是编码的主键值。
	// 需要表的 Primary Key 模式才能正确解码。
	// This function would typically be called with the table schema available.
	// 此函数通常在表模式可用的情况下被调用。
	// For this placeholder, we can't decode the primary key values without the schema.
	// For now, return the raw encoded PK bytes.
	//
	// 对于此占位符，在没有模式的情况下无法解码主键值。
	// 目前，返回原始编码的 PK 字节。

	// Dummy PK decoding (assuming simple case)
	// 虚拟 PK 解码（假设简单情况）
	encodedPKParts := parts[3:] // Get the parts after Namespace, DBName, TableName
	dummyPKValues := make([]interface{}, len(encodedPKParts))
	for i, part := range encodedPKParts {
		// Assuming each part is a simple string representation (from dummy encoder)
		// 假设每个部分是简单的字符串表示（来自虚拟编码器）
		dummyPKValues[i] = string(part) // Insecure/incorrect for real data types
	}


	log.Debug("Decoded row key %v: DB='%s', Table='%s', PK parts=%d", key, dbName, tableName, len(dummyPKValues)) // 解码行 Key。
	return dbName, tableName, dummyPKValues, nil // Return dummy values
}


// TODO: Implement encoding and decoding for index keys.
// Index keys typically include indexed column values and potentially the primary key to make it unique.
// Format: <Namespace_Index>:<DBName>:<TableName>:<IndexName>:<Encoded_IndexedColumnValues>[:<Encoded_PrimaryKey>]
//
// TODO: 实现索引 Key 的编码和解码。
// 索引 Key 通常包含索引列值，并可能包含主键以使其唯一。
// 格式：<索引命名空间>:<数据库名>:<表名>:<索引名称>:<编码的索引列值>[:<编码的主键>]
// Need to handle different index types (BTREE, spatial, full-text) and their specific encoding requirements.
// 需要处理不同的索引类型（BTREE、空间、全文）及其特定的编码要求。

// Note: Namespace bytes like NamespaceDataBytes, NamespaceCatalogBytes, NamespaceIndexBytes
// should be defined in a common place, possibly within this package or a dedicated 'keys' package
// in internal or storage. They are currently defined in storage/engines/badger/keys.go.
//
// 注意：命名空间字节，如 NamespaceDataBytes, NamespaceCatalogBytes, NamespaceIndexBytes，
// 应在通用位置定义，可能在此包中或 internal 或 storage 中的专用 'keys' 包中。
// 它们目前在 storage/engines/badger/keys.go 中定义。
// Let's redefine them here for independence, or rely on the Badger package for now.
// 暂时在此处重新定义它们以保持独立性，或依赖 Badger 包。
var (
	NamespaceDataBytes    = []byte("data")
	NamespaceCatalogBytes = []byte("catalog")
	NamespaceIndexBytes   = []byte("index")
)

// TODO: Refactor namespace definitions to a common package accessible by both internal/encoding and storage.
// TODO: 将命名空间定义重构到 internal/encoding 和 storage 都可以访问的通用包中。