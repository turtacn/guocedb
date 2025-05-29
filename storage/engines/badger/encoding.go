// Package badger implements encoding mechanisms for Badger storage engine
// badger包，实现Badger存储引擎的编码机制
package badger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/turtacn/guocedb/common/types"
	"github.com/turtacn/guocedb/common/types/value"
	"github.com/turtacn/guocedb/internal/encoding"
)

// KeyType 键类型 Key type
type KeyType byte

const (
	KeyTypeTableMeta   KeyType = 0x01 // 表元数据 Table metadata
	KeyTypeRow         KeyType = 0x02 // 行数据 Row data
	KeyTypeIndex       KeyType = 0x03 // 索引数据 Index data
	KeyTypeSequence    KeyType = 0x04 // 序列 Sequence
	KeyTypeSchema      KeyType = 0x05 // 模式 Schema
	KeyTypeConstraint  KeyType = 0x06 // 约束 Constraint
	KeyTypeStatistics  KeyType = 0x07 // 统计信息 Statistics
	KeyTypeTransaction KeyType = 0x08 // 事务日志 Transaction log
	KeyTypeBackup      KeyType = 0x09 // 备份信息 Backup info
	KeyTypeConfig      KeyType = 0x0A // 配置信息 Configuration
)

// String 返回键类型的字符串表示 Returns string representation of key type
func (kt KeyType) String() string {
	switch kt {
	case KeyTypeTableMeta:
		return "TABLE_META"
	case KeyTypeRow:
		return "ROW"
	case KeyTypeIndex:
		return "INDEX"
	case KeyTypeSequence:
		return "SEQUENCE"
	case KeyTypeSchema:
		return "SCHEMA"
	case KeyTypeConstraint:
		return "CONSTRAINT"
	case KeyTypeStatistics:
		return "STATISTICS"
	case KeyTypeTransaction:
		return "TRANSACTION"
	case KeyTypeBackup:
		return "BACKUP"
	case KeyTypeConfig:
		return "CONFIG"
	default:
		return fmt.Sprintf("UNKNOWN_%d", byte(kt))
	}
}

// KeyEncoder 键编码器 Key encoder
type KeyEncoder struct {
	database string            // 数据库名 Database name
	encoder  *encoding.Encoder // 基础编码器 Base encoder
}

// NewKeyEncoder 创建键编码器 Create key encoder
func NewKeyEncoder(database string) *KeyEncoder {
	return &KeyEncoder{
		database: database,
		encoder:  encoding.NewEncoder(),
	}
}

// EncodeTableMetaKey 编码表元数据键 Encode table metadata key
func (ke *KeyEncoder) EncodeTableMetaKey(table string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeTableMeta))

	// 数据库名长度和内容 Database name length and content
	ke.writeString(&buf, ke.database)

	// 表名长度和内容 Table name length and content
	ke.writeString(&buf, table)

	return buf.Bytes()
}

// EncodeRowKey 编码行键 Encode row key
func (ke *KeyEncoder) EncodeRowKey(table string, rowID string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeRow))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 行ID Row ID
	ke.writeString(&buf, rowID)

	return buf.Bytes()
}

// EncodeRowKeyWithValues 根据列值编码行键 Encode row key with column values
func (ke *KeyEncoder) EncodeRowKeyWithValues(table string, keyColumns []string, values []*types.Value) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeRow))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 编码键值 Encode key values
	for i, colName := range keyColumns {
		// 列名 Column name
		ke.writeString(&buf, colName)

		// 列值 Column value
		if i < len(values) && values[i] != nil {
			ke.writeValue(&buf, values[i])
		} else {
			// 空值标记 Null value marker
			buf.WriteByte(0x00)
		}
	}

	return buf.Bytes()
}

// EncodeRowPrefix 编码行前缀 Encode row prefix
func (ke *KeyEncoder) EncodeRowPrefix(table string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeRow))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	return buf.Bytes()
}

// EncodeIndexKey 编码索引键 Encode index key
func (ke *KeyEncoder) EncodeIndexKey(table, index string, values []*types.Value, rowID string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeIndex))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 索引名 Index name
	ke.writeString(&buf, index)

	// 索引值 Index values
	for _, value := range values {
		ke.writeValue(&buf, value)
	}

	// 行ID（用于唯一性） Row ID (for uniqueness)
	ke.writeString(&buf, rowID)

	return buf.Bytes()
}

// EncodeIndexPrefix 编码索引前缀 Encode index prefix
func (ke *KeyEncoder) EncodeIndexPrefix(table, index string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeIndex))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 索引名 Index name
	ke.writeString(&buf, index)

	return buf.Bytes()
}

// EncodeSequenceKey 编码序列键 Encode sequence key
func (ke *KeyEncoder) EncodeSequenceKey(table, column string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeSequence))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 列名 Column name
	ke.writeString(&buf, column)

	return buf.Bytes()
}

// EncodeSchemaKey 编码模式键 Encode schema key
func (ke *KeyEncoder) EncodeSchemaKey(table string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeSchema))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	return buf.Bytes()
}

// EncodeConstraintKey 编码约束键 Encode constraint key
func (ke *KeyEncoder) EncodeConstraintKey(table, constraint string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeConstraint))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	// 约束名 Constraint name
	ke.writeString(&buf, constraint)

	return buf.Bytes()
}

// EncodeStatisticsKey 编码统计信息键 Encode statistics key
func (ke *KeyEncoder) EncodeStatisticsKey(table string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeStatistics))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 表名 Table name
	ke.writeString(&buf, table)

	return buf.Bytes()
}

// EncodeTransactionKey 编码事务键 Encode transaction key
func (ke *KeyEncoder) EncodeTransactionKey(txnID string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeTransaction))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 事务ID Transaction ID
	ke.writeString(&buf, txnID)

	return buf.Bytes()
}

// EncodeConfigKey 编码配置键 Encode configuration key
func (ke *KeyEncoder) EncodeConfigKey(key string) []byte {
	var buf bytes.Buffer

	// 键类型 Key type
	buf.WriteByte(byte(KeyTypeConfig))

	// 数据库名 Database name
	ke.writeString(&buf, ke.database)

	// 配置键 Config key
	ke.writeString(&buf, key)

	return buf.Bytes()
}

// writeString 写入字符串 Write string
func (ke *KeyEncoder) writeString(buf *bytes.Buffer, s string) {
	// 长度前缀编码 Length-prefixed encoding
	length := len(s)
	if length < 255 {
		buf.WriteByte(byte(length))
	} else {
		buf.WriteByte(255)
		binary.Write(buf, binary.BigEndian, uint32(length))
	}
	buf.WriteString(s)
}

// writeValue 写入值 Write value
func (ke *KeyEncoder) writeValue(buf *bytes.Buffer, val *types.Value) {
	if val == nil || val.IsNull() {
		// 空值标记 Null marker
		buf.WriteByte(0x00)
		return
	}

	// 类型标记 Type marker
	buf.WriteByte(byte(val.Type) + 1) // +1 to avoid null marker

	// 编码值数据 Encode value data
	switch val.Type {
	case types.DataTypeBool:
		if val.GetBool() {
			buf.WriteByte(0x01)
		} else {
			buf.WriteByte(0x00)
		}

	case types.DataTypeInt8:
		buf.WriteByte(byte(val.GetInt8() + 128)) // 转换为无符号 Convert to unsigned

	case types.DataTypeInt16:
		binary.Write(buf, binary.BigEndian, uint16(val.GetInt16()+32768)) // 转换为无符号 Convert to unsigned

	case types.DataTypeInt32:
		binary.Write(buf, binary.BigEndian, uint32(val.GetInt32()+2147483648)) // 转换为无符号 Convert to unsigned

	case types.DataTypeInt64:
		binary.Write(buf, binary.BigEndian, uint64(val.GetInt64()+9223372036854775808)) // 转换为无符号 Convert to unsigned

	case types.DataTypeUint8:
		buf.WriteByte(val.GetUint8())

	case types.DataTypeUint16:
		binary.Write(buf, binary.BigEndian, val.GetUint16())

	case types.DataTypeUint32:
		binary.Write(buf, binary.BigEndian, val.GetUint32())

	case types.DataTypeUint64:
		binary.Write(buf, binary.BigEndian, val.GetUint64())

	case types.DataTypeFloat32:
		bits := math.Float32bits(val.GetFloat32())
		// IEEE 754 浮点数排序 IEEE 754 floating point ordering
		if bits&0x80000000 != 0 {
			bits = ^bits
		} else {
			bits |= 0x80000000
		}
		binary.Write(buf, binary.BigEndian, bits)

	case types.DataTypeFloat64:
		bits := math.Float64bits(val.GetFloat64())
		// IEEE 754 浮点数排序 IEEE 754 floating point ordering
		if bits&0x8000000000000000 != 0 {
			bits = ^bits
		} else {
			bits |= 0x8000000000000000
		}
		binary.Write(buf, binary.BigEndian, bits)

	case types.DataTypeString, types.DataTypeVector, types.DataTypeText:
		str := val.GetString()
		ke.writeString(buf, str)

	case types.DataTypeBytes, types.DataTypeBlob:
		data := val.GetBytes()
		binary.Write(buf, binary.BigEndian, uint32(len(data)))
		buf.Write(data)

	case types.DataTypeDate:
		date := val.GetDate()
		days := int32(date.Unix() / 86400)                           // 转换为天数 Convert to days
		binary.Write(buf, binary.BigEndian, uint32(days+2147483648)) // 转换为无符号 Convert to unsigned

	case types.DataTypeTime:
		t := val.GetTime()
		// 转换为纳秒 Convert to nanoseconds
		nanos := int64(t.Hour())*3600000000000 + int64(t.Minute())*60000000000 + int64(t.Second())*1000000000 + int64(t.Nanosecond())
		binary.Write(buf, binary.BigEndian, uint64(nanos))

	case types.DataTypeTimestamp:
		ts := val.GetTimestamp()
		binary.Write(buf, binary.BigEndian, uint64(ts.UnixNano()))

	case types.DataTypeUUID:
		uuid := val.GetUUID()
		buf.Write(uuid[:])

	case types.DataTypeJSON:
		// JSON作为字符串处理 Treat JSON as string
		jsonStr := val.GetString()
		ke.writeString(buf, jsonStr)

	case types.DataTypeArray:
		// 数组编码 Array encoding
		array := val.GetArray()
		binary.Write(buf, binary.BigEndian, uint32(len(array)))
		for _, item := range array {
			ke.writeValue(buf, item)
		}

	default:
		// 未知类型，编码为字节 Unknown type, encode as bytes
		data, _ := val.Bytes()
		binary.Write(buf, binary.BigEndian, uint32(len(data)))
		buf.Write(data)
	}
}

// KeyDecoder 键解码器 Key decoder
type KeyDecoder struct {
	decoder *encoding.Decoder // 基础解码器 Base decoder
}

// NewKeyDecoder 创建键解码器 Create key decoder
func NewKeyDecoder() *KeyDecoder {
	return &KeyDecoder{
		decoder: encoding.NewDecoder(),
	}
}

// DecodeKeyType 解码键类型 Decode key type
func (kd *KeyDecoder) DecodeKeyType(key []byte) (KeyType, error) {
	if len(key) == 0 {
		return 0, fmt.Errorf("empty key")
	}
	return KeyType(key[0]), nil
}

// DecodeTableMetaKey 解码表元数据键 Decode table metadata key
func (kd *KeyDecoder) DecodeTableMetaKey(key []byte) (database, table string, err error) {
	if len(key) == 0 || KeyType(key[0]) != KeyTypeTableMeta {
		return "", "", fmt.Errorf("invalid table meta key")
	}

	buf := bytes.NewReader(key[1:])

	database, err = kd.readString(buf)
	if err != nil {
		return "", "", fmt.Errorf("failed to read database name: %w", err)
	}

	table, err = kd.readString(buf)
	if err != nil {
		return "", "", fmt.Errorf("failed to read table name: %w", err)
	}

	return database, table, nil
}

// DecodeRowKey 解码行键 Decode row key
func (kd *KeyDecoder) DecodeRowKey(key []byte) (database, table, rowID string, err error) {
	if len(key) == 0 || KeyType(key[0]) != KeyTypeRow {
		return "", "", "", fmt.Errorf("invalid row key")
	}

	buf := bytes.NewReader(key[1:])

	database, err = kd.readString(buf)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read database name: %w", err)
	}

	table, err = kd.readString(buf)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read table name: %w", err)
	}

	rowID, err = kd.readString(buf)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read row ID: %w", err)
	}

	return database, table, rowID, nil
}

// DecodeIndexKey 解码索引键 Decode index key
func (kd *KeyDecoder) DecodeIndexKey(key []byte) (database, table, index string, values []*types.Value, rowID string, err error) {
	if len(key) == 0 || KeyType(key[0]) != KeyTypeIndex {
		return "", "", "", nil, "", fmt.Errorf("invalid index key")
	}

	buf := bytes.NewReader(key[1:])

	database, err = kd.readString(buf)
	if err != nil {
		return "", "", "", nil, "", fmt.Errorf("failed to read database name: %w", err)
	}

	table, err = kd.readString(buf)
	if err != nil {
		return "", "", "", nil, "", fmt.Errorf("failed to read table name: %w", err)
	}

	index, err = kd.readString(buf)
	if err != nil {
		return "", "", "", nil, "", fmt.Errorf("failed to read index name: %w", err)
	}

	// 读取索引值（直到遇到行ID） Read index values (until row ID)
	values = make([]*types.Value, 0)
	for buf.Len() > 0 {
		// 尝试读取值 Try to read value
		val, readErr := kd.readValue(buf)
		if readErr != nil {
			// 可能是行ID Row ID maybe
			break
		}
		values = append(values, val)
	}

	// 最后一个字符串应该是行ID The last string should be row ID
	if buf.Len() > 0 {
		buf = bytes.NewReader(key[len(key)-buf.Len():])
		rowID, err = kd.readString(buf)
		if err != nil {
			return "", "", "", nil, "", fmt.Errorf("failed to read row ID: %w", err)
		}
	}

	return database, table, index, values, rowID, nil
}

// readString 读取字符串 Read string
func (kd *KeyDecoder) readString(buf *bytes.Reader) (string, error) {
	// 读取长度 Read length
	lengthByte, err := buf.ReadByte()
	if err != nil {
		return "", err
	}

	var length int
	if lengthByte < 255 {
		length = int(lengthByte)
	} else {
		var lengthUint32 uint32
		if err := binary.Read(buf, binary.BigEndian, &lengthUint32); err != nil {
			return "", err
		}
		length = int(lengthUint32)
	}

	// 读取字符串内容 Read string content
	data := make([]byte, length)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}

	return string(data), nil
}

// readValue 读取值 Read value
func (kd *KeyDecoder) readValue(buf *bytes.Reader) (*types.Value, error) {
	// 读取类型标记 Read type marker
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	if typeByte == 0x00 {
		// 空值 Null value
		return types.NewNullValue(), nil
	}

	dataType := types.DataType(typeByte - 1) // -1 to get original type

	switch dataType {
	case types.DataTypeBoolean:
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return types.NewBoolValue(b != 0), nil

	case types.DataTypeInt8:
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return types.NewInt8Value(int8(b) - 128), nil

	case types.DataTypeInt16:
		var val uint16
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewInt16Value(int16(val) - 32768), nil

	case types.DataTypeInt32:
		var val uint32
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewInt32Value(int32(val) - 2147483648), nil

	case types.DataTypeInt64:
		var val uint64
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewInt64Value(int64(val) - 9223372036854775808), nil

	case types.DataTypeUint8:
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return types.NewUint8Value(b), nil

	case types.DataTypeUint16:
		var val uint16
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewUint16Value(val), nil

	case types.DataTypeUint32:
		var val uint32
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewUint32Value(val), nil

	case types.DataTypeUint64:
		var val uint64
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return types.NewUint64Value(val), nil

	case types.DataTypeFloat32:
		var bits uint32
		if err := binary.Read(buf, binary.BigEndian, &bits); err != nil {
			return nil, err
		}
		// 还原IEEE 754浮点数 Restore IEEE 754 floating point
		if bits&0x80000000 != 0 {
			bits &= 0x7FFFFFFF
		} else {
			bits = ^bits
		}
		return types.NewFloat32Value(math.Float32frombits(bits)), nil

	case types.DataTypeFloat64:
		var bits uint64
		if err := binary.Read(buf, binary.BigEndian, &bits); err != nil {
			return nil, err
		}
		// 还原IEEE 754浮点数 Restore IEEE 754 floating point
		if bits&0x8000000000000000 != 0 {
			bits &= 0x7FFFFFFFFFFFFFFF
		} else {
			bits = ^bits
		}
		return types.NewFloat64Value(math.Float64frombits(bits)), nil

	case types.DataTypeString, types.DataTypeVarchar, types.DataTypeText:
		str, err := kd.readString(buf)
		if err != nil {
			return nil, err
		}
		return types.NewStringValue(str), nil

	case types.DataTypeBytes, types.DataTypeBlob:
		var length uint32
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		data := make([]byte, length)
		if _, err := buf.Read(data); err != nil {
			return nil, err
		}
		return types.NewBytesValue(data), nil

	case types.DataTypeDate:
		var days uint32
		if err := binary.Read(buf, binary.BigEndian, &days); err != nil {
			return nil, err
		}
		// 转换回日期 Convert back to date
		daysSigned := int32(days) - 2147483648
		timestamp := int64(daysSigned) * 86400
		date := time.Unix(timestamp, 0)
		return types.NewDateValue(date), nil

	case types.DataTypeTime:
		var nanos uint64
		if err := binary.Read(buf, binary.BigEndian, &nanos); err != nil {
			return nil, err
		}
		// 转换回时间 Convert back to time
		hours := nanos / 3600000000000
		minutes := (nanos % 3600000000000) / 60000000000
		seconds := (nanos % 60000000000) / 1000000000
		nanoseconds := nanos % 1000000000
		t := time.Date(0, 1, 1, int(hours), int(minutes), int(seconds), int(nanoseconds), time.UTC)
		return types.NewTimeValue(t), nil

	case types.DataTypeTimestamp:
		var nanos uint64
		if err := binary.Read(buf, binary.BigEndian, &nanos); err != nil {
			return nil, err
		}
		timestamp := time.Unix(0, int64(nanos))
		return types.NewTimestampValue(timestamp), nil

	case types.DataTypeUUID:
		data := make([]byte, 16)
		if _, err := buf.Read(data); err != nil {
			return nil, err
		}
		var uuid [16]byte
		copy(uuid[:], data)
		return types.NewUUIDValue(uuid), nil

	case types.DataTypeJSON:
		str, err := kd.readString(buf)
		if err != nil {
			return nil, err
		}
		return types.NewJSONValue(str), nil

	case types.DataTypeArray:
		var length uint32
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		array := make([]*types.Value, length)
		for i := uint32(0); i < length; i++ {
			val, err := kd.readValue(buf)
			if err != nil {
				return nil, err
			}
			array[i] = val
		}
		return types.NewArrayValue(array), nil

	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}

// ValueEncoder 值编码器 Value encoder
type ValueEncoder struct {
	encoder *encoding.Encoder // 基础编码器 Base encoder
}

// NewValueEncoder 创建值编码器 Create value encoder
func NewValueEncoder() *ValueEncoder {
	return &ValueEncoder{
		encoder: encoding.NewEncoder(),
	}
}

// EncodeRow 编码行数据 Encode row data
func (ve *ValueEncoder) EncodeRow(row *types.Row) ([]byte, error) {
	if row == nil {
		return nil, fmt.Errorf("row is nil")
	}

	var buf bytes.Buffer

	// 版本号 Version number
	binary.Write(&buf, binary.BigEndian, row.Version)

	// 创建时间 Creation time
	binary.Write(&buf, binary.BigEndian, row.CreatedAt.UnixNano())

	// 更新时间 Update time
	binary.Write(&buf, binary.BigEndian, row.UpdatedAt.UnixNano())

	// 删除标记 Deletion flag
	if row.Deleted {
		buf.WriteByte(0x01)
	} else {
		buf.WriteByte(0x00)
	}

	// 值数量 Value count
	binary.Write(&buf, binary.BigEndian, uint32(len(row.Values)))

	// 编码每个值 Encode each value
	for _, val := range row.Values {
		data, err := ve.encodeValue(val)
		if err != nil {
			return nil, fmt.Errorf("failed to encode value: %w", err)
		}
		// 值长度 Value length
		binary.Write(&buf, binary.BigEndian, uint32(len(data)))
		// 值数据 Value data
		buf.Write(data)
	}

	// 元数据 Metadata
	metadataData, err := ve.encoder.EncodeMap(row.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to encode metadata: %w", err)
	}
	binary.Write(&buf, binary.BigEndian, uint32(len(metadataData)))
	buf.Write(metadataData)

	return buf.Bytes(), nil
}

// DecodeRow 解码行数据 Decode row data
func (ve *ValueEncoder) DecodeRow(data []byte) (*types.Row, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	buf := bytes.NewReader(data)
	row := &types.Row{}

	// 版本号 Version number
	if err := binary.Read(buf, binary.BigEndian, &row.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// 创建时间 Creation time
	var createdNanos int64
	if err := binary.Read(buf, binary.BigEndian, &createdNanos); err != nil {
		return nil, fmt.Errorf("failed to read created time: %w", err)
	}
	row.CreatedAt = time.Unix(0, createdNanos)

	// 更新时间 Update time
	var updatedNanos int64
	if err := binary.Read(buf, binary.BigEndian, &updatedNanos); err != nil {
		return nil, fmt.Errorf("failed to read updated time: %w", err)
	}
	row.UpdatedAt = time.Unix(0, updatedNanos)

	// 删除标记 Deletion flag
	deletedByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read deleted flag: %w", err)
	}
	row.Deleted = deletedByte != 0x00

	// 值数量 Value count
	var valueCount uint32
	if err := binary.Read(buf, binary.BigEndian, &valueCount); err != nil {
		return nil, fmt.Errorf("failed to read value count: %w", err)
	}

	// 解码每个值 Decode each value
	row.Values = make([]*types.Value, valueCount)
	for i := uint32(0); i < valueCount; i++ {
		// 值长度 Value length
		var valueLength uint32
		if err := binary.Read(buf, binary.BigEndian, &valueLength); err != nil {
			return nil, fmt.Errorf("failed to read value length: %w", err)
		}

		// 值数据 Value data
		valueData := make([]byte, valueLength)
		if _, err := buf.Read(valueData); err != nil {
			return nil, fmt.Errorf("failed to read value data: %w", err)
		}

		val, err := ve.decodeValue(valueData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode value: %w", err)
		}
		row.Values[i] = val
	}

	// 元数据 Metadata
	var metadataLength uint32
	if err := binary.Read(buf, binary.BigEndian, &metadataLength); err != nil {
		return nil, fmt.Errorf("failed to read metadata length: %w", err)
	}

	metadataData := make([]byte, metadataLength)
	if _, err := buf.Read(metadataData); err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	metadata, err := ve.encoder.DecodeMap(metadataData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}
	row.Metadata = metadata

	return row, nil
}

// encodeValue 编码值 Encode value
func (ve *ValueEncoder) encodeValue(val *types.Value) ([]byte, error) {
	if val == nil {
		return []byte{0x00}, nil // Null marker
	}

	data, err := val.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get value bytes: %w", err)
	}

	var buf bytes.Buffer
	// 类型标记 Type marker
	buf.WriteByte(byte(val.Type))
	// 数据 Data
	buf.Write(data)

	return buf.Bytes(), nil
}

// decodeValue 解码值 Decode value
func (ve *ValueEncoder) decodeValue(data []byte) (*types.Value, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty value data")
	}

	if data[0] == 0x00 {
		return types.NewNullValue(), nil
	}

	dataType := types.DataType(data[0])
	valueData := data[1:]

	return value.FromBytes(dataType, valueData)
}

// EncodeSchema 编码模式 Encode schema
func (ve *ValueEncoder) EncodeSchema(schema *types.Schema) ([]byte, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	return ve.encoder.EncodeStruct(schema)
}

// DecodeSchema 解码模式 Decode schema
func (ve *ValueEncoder) DecodeSchema(data []byte) (*types.Schema, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty schema data")
	}

	var schema types.Schema
	err := ve.encoder.DecodeStruct(data, &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}

	return &schema, nil
}

// EncodeIndex 编码索引 Encode index
func (ve *ValueEncoder) EncodeIndex(index *types.Index) ([]byte, error) {
	if index == nil {
		return nil, fmt.Errorf("index is nil")
	}

	return ve.encoder.EncodeStruct(index)
}

// DecodeIndex 解码索引 Decode index
func (ve *ValueEncoder) DecodeIndex(data []byte) (*types.Index, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty index data")
	}

	var index types.Index
	err := ve.encoder.DecodeStruct(data, &index)
	if err != nil {
		return nil, fmt.Errorf("failed to decode index: %w", err)
	}

	return &index, nil
}

// KeyRange 键范围 Key range
type KeyRange struct {
	Start []byte // 开始键 Start key
	End   []byte // 结束键 End key
}

// NewKeyRange 创建键范围 Create key range
func NewKeyRange(start, end []byte) *KeyRange {
	return &KeyRange{
		Start: start,
		End:   end,
	}
}

// Contains 检查键是否在范围内 Check if key is in range
func (kr *KeyRange) Contains(key []byte) bool {
	return bytes.Compare(key, kr.Start) >= 0 &&
		(kr.End == nil || bytes.Compare(key, kr.End) < 0)
}

// String 返回字符串表示 Return string representation
func (kr *KeyRange) String() string {
	return fmt.Sprintf("[%x, %x)", kr.Start, kr.End)
}

// GenerateRowID 生成行ID Generate row ID
func GenerateRowID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(),
		time.Now().Nanosecond())
}

// ValidateKey 验证键格式 Validate key format
func ValidateKey(key []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("empty key")
	}

	keyType := KeyType(key[0])
	switch keyType {
	case KeyTypeTableMeta, KeyTypeRow, KeyTypeIndex, KeyTypeSequence,
		KeyTypeSchema, KeyTypeConstraint, KeyTypeStatistics,
		KeyTypeTransaction, KeyTypeBackup, KeyTypeConfig:
		// 有效的键类型 Valid key type
		return nil
	default:
		return fmt.Errorf("invalid key type: %d", keyType)
	}
}

// CompareKeys 比较键 Compare keys
func CompareKeys(key1, key2 []byte) int {
	return bytes.Compare(key1, key2)
}

// ExtractTableFromKey 从键中提取表名 Extract table name from key
func ExtractTableFromKey(key []byte) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("empty key")
	}

	keyType := KeyType(key[0])
	decoder := NewKeyDecoder()

	switch keyType {
	case KeyTypeTableMeta:
		_, table, err := decoder.DecodeTableMetaKey(key)
		return table, err
	case KeyTypeRow:
		_, table, _, err := decoder.DecodeRowKey(key)
		return table, err
	case KeyTypeIndex:
		_, table, _, _, _, err := decoder.DecodeIndexKey(key)
		return table, err
	case KeyTypeSchema:
		_, table, err := decoder.DecodeTableMetaKey(key) // Schema key has same format
		return table, err
	default:
		return "", fmt.Errorf("unsupported key type for table extraction: %v", keyType)
	}
}
