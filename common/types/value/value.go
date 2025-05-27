// Package value 定义GuoceDB的SQL值类型系统
// Package value defines SQL value type system for GuoceDB
package value

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
)

// ValueType 值类型枚举
// ValueType enumeration for value types
type ValueType int

const (
	// TypeNull NULL类型
	// TypeNull NULL type
	TypeNull ValueType = iota
	// TypeBool 布尔类型
	// TypeBool boolean type
	TypeBool
	// TypeInt8 8位整数类型
	// TypeInt8 8-bit integer type
	TypeInt8
	// TypeInt16 16位整数类型
	// TypeInt16 16-bit integer type
	TypeInt16
	// TypeInt32 32位整数类型
	// TypeInt32 32-bit integer type
	TypeInt32
	// TypeInt64 64位整数类型
	// TypeInt64 64-bit integer type
	TypeInt64
	// TypeFloat32 32位浮点类型
	// TypeFloat32 32-bit float type
	TypeFloat32
	// TypeFloat64 64位浮点类型
	// TypeFloat64 64-bit float type
	TypeFloat64
	// TypeString 字符串类型
	// TypeString string type
	TypeString
	// TypeBlob 二进制类型
	// TypeBlob binary type
	TypeBlob
	// TypeDate 日期类型
	// TypeDate date type
	TypeDate
	// TypeTime 时间类型
	// TypeTime time type
	TypeTime
	// TypeDateTime 日期时间类型
	// TypeDateTime datetime type
	TypeDateTime
	// TypeTimestamp 时间戳类型
	// TypeTimestamp timestamp type
	TypeTimestamp
	// TypeDecimal 精确小数类型
	// TypeDecimal decimal type
	TypeDecimal
)

// String 返回值类型的字符串表示
// String returns string representation of value type
func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "NULL"
	case TypeBool:
		return "BOOL"
	case TypeInt8:
		return "TINYINT"
	case TypeInt16:
		return "SMALLINT"
	case TypeInt32:
		return "INT"
	case TypeInt64:
		return "BIGINT"
	case TypeFloat32:
		return "FLOAT"
	case TypeFloat64:
		return "DOUBLE"
	case TypeString:
		return "VARCHAR"
	case TypeBlob:
		return "BLOB"
	case TypeDate:
		return "DATE"
	case TypeTime:
		return "TIME"
	case TypeDateTime:
		return "DATETIME"
	case TypeTimestamp:
		return "TIMESTAMP"
	case TypeDecimal:
		return "DECIMAL"
	default:
		return fmt.Sprintf("UNKNOWN_TYPE(%d)", int(t))
	}
}

// IsNumeric 检查类型是否为数值类型
// IsNumeric checks if type is numeric
func (t ValueType) IsNumeric() bool {
	return t >= TypeInt8 && t <= TypeDecimal && t != TypeString && t != TypeBlob
}

// IsInteger 检查类型是否为整数类型
// IsInteger checks if type is integer type
func (t ValueType) IsInteger() bool {
	return t >= TypeInt8 && t <= TypeInt64
}

// IsFloat 检查类型是否为浮点类型
// IsFloat checks if type is float type
func (t ValueType) IsFloat() bool {
	return t == TypeFloat32 || t == TypeFloat64 || t == TypeDecimal
}

// IsString 检查类型是否为字符串类型
// IsString checks if type is string type
func (t ValueType) IsString() bool {
	return t == TypeString || t == TypeBlob
}

// IsTemporal 检查类型是否为时间类型
// IsTemporal checks if type is temporal type
func (t ValueType) IsTemporal() bool {
	return t >= TypeDate && t <= TypeTimestamp
}

// CompareResult 比较结果枚举
// CompareResult enumeration for comparison results
type CompareResult int

const (
	// CompareLess 小于
	// CompareLess less than
	CompareLess CompareResult = -1
	// CompareEqual 等于
	// CompareEqual equal
	CompareEqual CompareResult = 0
	// CompareGreater 大于
	// CompareGreater greater than
	CompareGreater CompareResult = 1
	// CompareIncomparable 不可比较
	// CompareIncomparable incomparable
	CompareIncomparable CompareResult = 2
)

// Value SQL值接口
// Value interface for SQL values
type Value interface {
	// Type 返回值类型
	// Type returns value type
	Type() ValueType

	// IsNull 检查是否为NULL
	// IsNull checks if value is NULL
	IsNull() bool

	// String 返回字符串表示
	// String returns string representation
	String() string

	// Compare 与另一个值比较
	// Compare compares with another value
	Compare(other Value) CompareResult

	// Equals 检查是否相等
	// Equals checks if values are equal
	Equals(other Value) bool

	// Clone 克隆值
	// Clone clones the value
	Clone() Value

	// ToBool 转换为布尔值
	// ToBool converts to boolean
	ToBool() (bool, error)

	// ToInt64 转换为64位整数
	// ToInt64 converts to 64-bit integer
	ToInt64() (int64, error)

	// ToFloat64 转换为64位浮点数
	// ToFloat64 converts to 64-bit float
	ToFloat64() (float64, error)

	// ToString 转换为字符串
	// ToString converts to string
	ToString() (string, error)

	// ToBytes 转换为字节数组
	// ToBytes converts to byte array
	ToBytes() ([]byte, error)

	// ToTime 转换为时间
	// ToTime converts to time
	ToTime() (time.Time, error)

	// Serialize 序列化为字节数组
	// Serialize serializes to byte array
	Serialize() ([]byte, error)

	// Size 返回值占用的字节大小
	// Size returns size in bytes
	Size() int

	// Hash 返回哈希值
	// Hash returns hash value
	Hash() uint64
}

// NullValue NULL值实现
// NullValue NULL value implementation
type NullValue struct{}

// NewNullValue 创建NULL值
// NewNullValue creates NULL value
func NewNullValue() Value {
	return &NullValue{}
}

// Type 返回值类型
// Type returns value type
func (v *NullValue) Type() ValueType {
	return TypeNull
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *NullValue) IsNull() bool {
	return true
}

// String 返回字符串表示
// String returns string representation
func (v *NullValue) String() string {
	return "NULL"
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *NullValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		return CompareEqual
	}
	return CompareLess
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *NullValue) Equals(other Value) bool {
	return other.IsNull()
}

// Clone 克隆值
// Clone clones the value
func (v *NullValue) Clone() Value {
	return &NullValue{}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *NullValue) ToBool() (bool, error) {
	return false, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to boolean")
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *NullValue) ToInt64() (int64, error) {
	return 0, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to integer")
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *NullValue) ToFloat64() (float64, error) {
	return 0, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to float")
}

// ToString 转换为字符串
// ToString converts to string
func (v *NullValue) ToString() (string, error) {
	return "", errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to string")
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *NullValue) ToBytes() ([]byte, error) {
	return nil, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to bytes")
}

// ToTime 转换为时间
// ToTime converts to time
func (v *NullValue) ToTime() (time.Time, error) {
	return time.Time{}, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NULL to time")
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *NullValue) Serialize() ([]byte, error) {
	return []byte{byte(TypeNull)}, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *NullValue) Size() int {
	return 1
}

// Hash 返回哈希值
// Hash returns hash value
func (v *NullValue) Hash() uint64 {
	return 0
}

// BoolValue 布尔值实现
// BoolValue boolean value implementation
type BoolValue struct {
	value bool
}

// NewBoolValue 创建布尔值
// NewBoolValue creates boolean value
func NewBoolValue(value bool) Value {
	return &BoolValue{value: value}
}

// Type 返回值类型
// Type returns value type
func (v *BoolValue) Type() ValueType {
	return TypeBool
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *BoolValue) IsNull() bool {
	return false
}

// String 返回字符串表示
// String returns string representation
func (v *BoolValue) String() string {
	if v.value {
		return "TRUE"
	}
	return "FALSE"
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *BoolValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		return CompareGreater
	}

	otherBool, err := other.ToBool()
	if err != nil {
		return CompareIncomparable
	}

	if v.value == otherBool {
		return CompareEqual
	} else if v.value {
		return CompareGreater
	}
	return CompareLess
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *BoolValue) Equals(other Value) bool {
	return v.Compare(other) == CompareEqual
}

// Clone 克隆值
// Clone clones the value
func (v *BoolValue) Clone() Value {
	return &BoolValue{value: v.value}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *BoolValue) ToBool() (bool, error) {
	return v.value, nil
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *BoolValue) ToInt64() (int64, error) {
	if v.value {
		return 1, nil
	}
	return 0, nil
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *BoolValue) ToFloat64() (float64, error) {
	if v.value {
		return 1.0, nil
	}
	return 0.0, nil
}

// ToString 转换为字符串
// ToString converts to string
func (v *BoolValue) ToString() (string, error) {
	return v.String(), nil
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *BoolValue) ToBytes() ([]byte, error) {
	return []byte(v.String()), nil
}

// ToTime 转换为时间
// ToTime converts to time
func (v *BoolValue) ToTime() (time.Time, error) {
	return time.Time{}, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert boolean to time")
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *BoolValue) Serialize() ([]byte, error) {
	buf := make([]byte, 2)
	buf[0] = byte(TypeBool)
	if v.value {
		buf[1] = 1
	} else {
		buf[1] = 0
	}
	return buf, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *BoolValue) Size() int {
	return 2
}

// Hash 返回哈希值
// Hash returns hash value
func (v *BoolValue) Hash() uint64 {
	if v.value {
		return 1
	}
	return 0
}

// IntValue 整数值实现
// IntValue integer value implementation
type IntValue struct {
	value   int64
	intType ValueType
}

// NewInt8Value 创建8位整数值
// NewInt8Value creates 8-bit integer value
func NewInt8Value(value int8) Value {
	return &IntValue{value: int64(value), intType: TypeInt8}
}

// NewInt16Value 创建16位整数值
// NewInt16Value creates 16-bit integer value
func NewInt16Value(value int16) Value {
	return &IntValue{value: int64(value), intType: TypeInt16}
}

// NewInt32Value 创建32位整数值
// NewInt32Value creates 32-bit integer value
func NewInt32Value(value int32) Value {
	return &IntValue{value: int64(value), intType: TypeInt32}
}

// NewInt64Value 创建64位整数值
// NewInt64Value creates 64-bit integer value
func NewInt64Value(value int64) Value {
	return &IntValue{value: value, intType: TypeInt64}
}

// Type 返回值类型
// Type returns value type
func (v *IntValue) Type() ValueType {
	return v.intType
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *IntValue) IsNull() bool {
	return false
}

// String 返回字符串表示
// String returns string representation
func (v *IntValue) String() string {
	return strconv.FormatInt(v.value, 10)
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *IntValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		return CompareGreater
	}

	otherInt, err := other.ToInt64()
	if err != nil {
		otherFloat, err := other.ToFloat64()
		if err != nil {
			return CompareIncomparable
		}
		thisFloat := float64(v.value)
		if thisFloat < otherFloat {
			return CompareLess
		} else if thisFloat > otherFloat {
			return CompareGreater
		}
		return CompareEqual
	}

	if v.value < otherInt {
		return CompareLess
	} else if v.value > otherInt {
		return CompareGreater
	}
	return CompareEqual
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *IntValue) Equals(other Value) bool {
	return v.Compare(other) == CompareEqual
}

// Clone 克隆值
// Clone clones the value
func (v *IntValue) Clone() Value {
	return &IntValue{value: v.value, intType: v.intType}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *IntValue) ToBool() (bool, error) {
	return v.value != 0, nil
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *IntValue) ToInt64() (int64, error) {
	return v.value, nil
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *IntValue) ToFloat64() (float64, error) {
	return float64(v.value), nil
}

// ToString 转换为字符串
// ToString converts to string
func (v *IntValue) ToString() (string, error) {
	return v.String(), nil
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *IntValue) ToBytes() ([]byte, error) {
	return []byte(v.String()), nil
}

// ToTime 转换为时间
// ToTime converts to time
func (v *IntValue) ToTime() (time.Time, error) {
	if v.intType == TypeInt64 {
		return time.Unix(v.value, 0), nil
	}
	return time.Time{}, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert non-int64 integer to time")
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *IntValue) Serialize() ([]byte, error) {
	buf := make([]byte, 9)
	buf[0] = byte(v.intType)
	binary.LittleEndian.PutUint64(buf[1:], uint64(v.value))
	return buf, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *IntValue) Size() int {
	switch v.intType {
	case TypeInt8:
		return 2
	case TypeInt16:
		return 3
	case TypeInt32:
		return 5
	case TypeInt64:
		return 9
	default:
		return 9
	}
}

// Hash 返回哈希值
// Hash returns hash value
func (v *IntValue) Hash() uint64 {
	return uint64(v.value)
}

// FloatValue 浮点值实现
// FloatValue float value implementation
type FloatValue struct {
	value     float64
	floatType ValueType
}

// NewFloat32Value 创建32位浮点值
// NewFloat32Value creates 32-bit float value
func NewFloat32Value(value float32) Value {
	return &FloatValue{value: float64(value), floatType: TypeFloat32}
}

// NewFloat64Value 创建64位浮点值
// NewFloat64Value creates 64-bit float value
func NewFloat64Value(value float64) Value {
	return &FloatValue{value: value, floatType: TypeFloat64}
}

// Type 返回值类型
// Type returns value type
func (v *FloatValue) Type() ValueType {
	return v.floatType
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *FloatValue) IsNull() bool {
	return false
}

// String 返回字符串表示
// String returns string representation
func (v *FloatValue) String() string {
	if v.floatType == TypeFloat32 {
		return strconv.FormatFloat(v.value, 'g', 7, 32)
	}
	return strconv.FormatFloat(v.value, 'g', 15, 64)
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *FloatValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		return CompareGreater
	}

	otherFloat, err := other.ToFloat64()
	if err != nil {
		return CompareIncomparable
	}

	if math.IsNaN(v.value) || math.IsNaN(otherFloat) {
		return CompareIncomparable
	}

	if v.value < otherFloat {
		return CompareLess
	} else if v.value > otherFloat {
		return CompareGreater
	}
	return CompareEqual
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *FloatValue) Equals(other Value) bool {
	return v.Compare(other) == CompareEqual
}

// Clone 克隆值
// Clone clones the value
func (v *FloatValue) Clone() Value {
	return &FloatValue{value: v.value, floatType: v.floatType}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *FloatValue) ToBool() (bool, error) {
	return v.value != 0.0, nil
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *FloatValue) ToInt64() (int64, error) {
	if math.IsNaN(v.value) || math.IsInf(v.value, 0) {
		return 0, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NaN/Inf to integer")
	}
	return int64(v.value), nil
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *FloatValue) ToFloat64() (float64, error) {
	return v.value, nil
}

// ToString 转换为字符串
// ToString converts to string
func (v *FloatValue) ToString() (string, error) {
	return v.String(), nil
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *FloatValue) ToBytes() ([]byte, error) {
	return []byte(v.String()), nil
}

// ToTime 转换为时间
// ToTime converts to time
func (v *FloatValue) ToTime() (time.Time, error) {
	if math.IsNaN(v.value) || math.IsInf(v.value, 0) {
		return time.Time{}, errors.NewError(errors.ErrCodeInvalidOperation, "Cannot convert NaN/Inf to time")
	}
	return time.Unix(int64(v.value), 0), nil
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *FloatValue) Serialize() ([]byte, error) {
	buf := make([]byte, 9)
	buf[0] = byte(v.floatType)
	binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(v.value))
	return buf, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *FloatValue) Size() int {
	if v.floatType == TypeFloat32 {
		return 5
	}
	return 9
}

// Hash 返回哈希值
// Hash returns hash value
func (v *FloatValue) Hash() uint64 {
	return math.Float64bits(v.value)
}

// StringValue 字符串值实现
// StringValue string value implementation
type StringValue struct {
	value      string
	stringType ValueType
}

// NewStringValue 创建字符串值
// NewStringValue creates string value
func NewStringValue(value string) Value {
	return &StringValue{value: value, stringType: TypeString}
}

// NewBlobValue 创建二进制数据值
// NewBlobValue creates blob value
func NewBlobValue(value []byte) Value {
	return &StringValue{value: string(value), stringType: TypeBlob}
}

// Type 返回值类型
// Type returns value type
func (v *StringValue) Type() ValueType {
	return v.stringType
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *StringValue) IsNull() bool {
	return false
}

// String 返回字符串表示
// String returns string representation
func (v *StringValue) String() string {
	return v.value
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *StringValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		return CompareGreater
	}

	otherStr, err := other.ToString()
	if err != nil {
		return CompareIncomparable
	}

	result := strings.Compare(v.value, otherStr)
	if result < 0 {
		return CompareLess
	} else if result > 0 {
		return CompareGreater
	}
	return CompareEqual
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *StringValue) Equals(other Value) bool {
	return v.Compare(other) == CompareEqual
}

// Clone 克隆值
// Clone clones the value
func (v *StringValue) Clone() Value {
	return &StringValue{value: v.value, stringType: v.stringType}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *StringValue) ToBool() (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(v.value))
	switch lower {
	case "true", "t", "1", "yes", "y", "on":
		return true, nil
	case "false", "f", "0", "no", "n", "off", "":
		return false, nil
	default:
		return false, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot convert '%s' to boolean", v.value)
	}
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *StringValue) ToInt64() (int64, error) {
	trimmed := strings.TrimSpace(v.value)
	if trimmed == "" {
		return 0, nil
	}

	result, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot convert '%s' to integer: %v", v.value, err)
	}
	return result, nil
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *StringValue) ToFloat64() (float64, error) {
	trimmed := strings.TrimSpace(v.value)
	if trimmed == "" {
		return 0.0, nil
	}

	result, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot convert '%s' to float: %v", v.value, err)
	}
	return result, nil
}

// ToString 转换为字符串
// ToString converts to string
func (v *StringValue) ToString() (string, error) {
	return v.value, nil
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *StringValue) ToBytes() ([]byte, error) {
	return []byte(v.value), nil
}

// ToTime 转换为时间
// ToTime converts to time
func (v *StringValue) ToTime() (time.Time, error) {
	trimmed := strings.TrimSpace(v.value)
	if trimmed == "" {
		return time.Time{}, nil
	}

	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"15:04:05",
		time.RFC3339,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, trimmed); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot convert '%s' to time", v.value)
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *StringValue) Serialize() ([]byte, error) {
	data := []byte(v.value)
	buf := make([]byte, 1+4+len(data))
	buf[0] = byte(v.stringType)
	binary.LittleEndian.PutUint32(buf[1:5], uint32(len(data)))
	copy(buf[5:], data)
	return buf, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *StringValue) Size() int {
	return 1 + 4 + len(v.value)
}

// Hash 返回哈希值
// Hash returns hash value
func (v *StringValue) Hash() uint64 {
	var h uint64 = 5381
	for _, c := range []byte(v.value) {
		h = ((h << 5) + h) + uint64(c)
	}
	return h
}

// DateTimeValue 日期时间值实现
// DateTimeValue datetime value implementation
type DateTimeValue struct {
	value        time.Time
	temporalType ValueType
}

// NewDateValue 创建日期值
// NewDateValue creates date value
func NewDateValue(value time.Time) Value {
	return &DateTimeValue{value: value, temporalType: TypeDate}
}

// NewTimeValue 创建时间值
// NewTimeValue creates time value
func NewTimeValue(value time.Time) Value {
	return &DateTimeValue{value: value, temporalType: TypeTime}
}

// NewDateTimeValue 创建日期时间值
// NewDateTimeValue creates datetime value
func NewDateTimeValue(value time.Time) Value {
	return &DateTimeValue{value: value, temporalType: TypeDateTime}
}

// NewTimestampValue 创建时间戳值
// NewTimestampValue creates timestamp value
func NewTimestampValue(value time.Time) Value {
	return &DateTimeValue{value: value, temporalType: TypeTimestamp}
}

// Type 返回值类型
// Type returns value type
func (v *DateTimeValue) Type() ValueType {
	return v.temporalType
}

// IsNull 检查是否为NULL
// IsNull checks if value is NULL
func (v *DateTimeValue) IsNull() bool {
	return v.value.IsZero()
}

// String 返回字符串表示
// String returns string representation
func (v *DateTimeValue) String() string {
	if v.value.IsZero() {
		return "NULL"
	}

	switch v.temporalType {
	case TypeDate:
		return v.value.Format("2006-01-02")
	case TypeTime:
		return v.value.Format("15:04:05")
	case TypeDateTime:
		return v.value.Format("2006-01-02 15:04:05")
	case TypeTimestamp:
		return v.value.Format("2006-01-02 15:04:05")
	default:
		return v.value.String()
	}
}

// Compare 与另一个值比较
// Compare compares with another value
func (v *DateTimeValue) Compare(other Value) CompareResult {
	if other.IsNull() {
		if v.IsNull() {
			return CompareEqual
		}
		return CompareGreater
	}

	if v.IsNull() {
		return CompareLess
	}

	otherTime, err := other.ToTime()
	if err != nil {
		return CompareIncomparable
	}

	if v.value.Before(otherTime) {
		return CompareLess
	} else if v.value.After(otherTime) {
		return CompareGreater
	}
	return CompareEqual
}

// Equals 检查是否相等
// Equals checks if values are equal
func (v *DateTimeValue) Equals(other Value) bool {
	return v.Compare(other) == CompareEqual
}

// Clone 克隆值
// Clone clones the value
func (v *DateTimeValue) Clone() Value {
	return &DateTimeValue{value: v.value, temporalType: v.temporalType}
}

// ToBool 转换为布尔值
// ToBool converts to boolean
func (v *DateTimeValue) ToBool() (bool, error) {
	return !v.value.IsZero(), nil
}

// ToInt64 转换为64位整数
// ToInt64 converts to 64-bit integer
func (v *DateTimeValue) ToInt64() (int64, error) {
	if v.value.IsZero() {
		return 0, nil
	}
	return v.value.Unix(), nil
}

// ToFloat64 转换为64位浮点数
// ToFloat64 converts to 64-bit float
func (v *DateTimeValue) ToFloat64() (float64, error) {
	if v.value.IsZero() {
		return 0.0, nil
	}
	return float64(v.value.Unix()), nil
}

// ToString 转换为字符串
// ToString converts to string
func (v *DateTimeValue) ToString() (string, error) {
	return v.String(), nil
}

// ToBytes 转换为字节数组
// ToBytes converts to byte array
func (v *DateTimeValue) ToBytes() ([]byte, error) {
	return []byte(v.String()), nil
}

// ToTime 转换为时间
// ToTime converts to time
func (v *DateTimeValue) ToTime() (time.Time, error) {
	return v.value, nil
}

// Serialize 序列化为字节数组
// Serialize serializes to byte array
func (v *DateTimeValue) Serialize() ([]byte, error) {
	buf := make([]byte, 9)
	buf[0] = byte(v.temporalType)
	binary.LittleEndian.PutUint64(buf[1:], uint64(v.value.Unix()))
	return buf, nil
}

// Size 返回值占用的字节大小
// Size returns size in bytes
func (v *DateTimeValue) Size() int {
	return 9
}

// Hash 返回哈希值
// Hash returns hash value
func (v *DateTimeValue) Hash() uint64 {
	return uint64(v.value.Unix())
}

// 工具函数 Utility functions

// ParseValue 从字符串解析值
// ParseValue parses value from string
func ParseValue(valueType ValueType, str string) (Value, error) {
	str = strings.TrimSpace(str)

	// 处理NULL值
	if strings.ToUpper(str) == "NULL" || str == "" {
		return NewNullValue(), nil
	}

	switch valueType {
	case TypeNull:
		return NewNullValue(), nil

	case TypeBool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as boolean", str)
		}
		return NewBoolValue(val), nil

	case TypeInt8:
		val, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as int8", str)
		}
		return NewInt8Value(int8(val)), nil

	case TypeInt16:
		val, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as int16", str)
		}
		return NewInt16Value(int16(val)), nil

	case TypeInt32:
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as int32", str)
		}
		return NewInt32Value(int32(val)), nil

	case TypeInt64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as int64", str)
		}
		return NewInt64Value(val), nil

	case TypeFloat32:
		val, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as float32", str)
		}
		return NewFloat32Value(float32(val)), nil

	case TypeFloat64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as float64", str)
		}
		return NewFloat64Value(val), nil

	case TypeString:
		return NewStringValue(str), nil

	case TypeBlob:
		return NewBlobValue([]byte(str)), nil

	case TypeDate:
		val, err := time.Parse("2006-01-02", str)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as date", str)
		}
		return NewDateValue(val), nil

	case TypeTime:
		val, err := time.Parse("15:04:05", str)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as time", str)
		}
		return NewTimeValue(val), nil

	case TypeDateTime:
		val, err := time.Parse("2006-01-02 15:04:05", str)
		if err != nil {
			return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as datetime", str)
		}
		return NewDateTimeValue(val), nil

	case TypeTimestamp:
		val, err := time.Parse("2006-01-02 15:04:05", str)
		if err != nil {
			// 尝试解析Unix时间戳
			if timestamp, err2 := strconv.ParseInt(str, 10, 64); err2 == nil {
				val = time.Unix(timestamp, 0)
			} else {
				return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Cannot parse '%s' as timestamp", str)
			}
		}
		return NewTimestampValue(val), nil

	default:
		return nil, errors.NewErrorf(errors.ErrCodeInvalidParameter, "Unsupported value type: %v", valueType)
	}
}

// DeserializeValue 从字节数组反序列化值
// DeserializeValue deserializes value from byte array
func DeserializeValue(data []byte) (Value, error) {
	if len(data) < 1 {
		return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized value: empty data")
	}

	valueType := ValueType(data[0])

	switch valueType {
	case TypeNull:
		return NewNullValue(), nil

	case TypeBool:
		if len(data) < 2 {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized boolean value")
		}
		return NewBoolValue(data[1] != 0), nil

	case TypeInt8, TypeInt16, TypeInt32, TypeInt64:
		if len(data) < 9 {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized integer value")
		}
		value := int64(binary.LittleEndian.Uint64(data[1:9]))
		switch valueType {
		case TypeInt8:
			return NewInt8Value(int8(value)), nil
		case TypeInt16:
			return NewInt16Value(int16(value)), nil
		case TypeInt32:
			return NewInt32Value(int32(value)), nil
		case TypeInt64:
			return NewInt64Value(value), nil
		}

	case TypeFloat32, TypeFloat64:
		if len(data) < 9 {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized float value")
		}
		bits := binary.LittleEndian.Uint64(data[1:9])
		value := math.Float64frombits(bits)
		if valueType == TypeFloat32 {
			return NewFloat32Value(float32(value)), nil
		}
		return NewFloat64Value(value), nil

	case TypeString, TypeBlob:
		if len(data) < 5 {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized string value")
		}
		length := binary.LittleEndian.Uint32(data[1:5])
		if len(data) < int(5+length) {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized string value: insufficient data")
		}
		value := string(data[5 : 5+length])
		if valueType == TypeString {
			return NewStringValue(value), nil
		}
		return NewBlobValue([]byte(value)), nil

	case TypeDate, TypeTime, TypeDateTime, TypeTimestamp:
		if len(data) < 9 {
			return nil, errors.NewError(errors.ErrCodeInvalidFormat, "Invalid serialized time value")
		}
		timestamp := int64(binary.LittleEndian.Uint64(data[1:9]))
		value := time.Unix(timestamp, 0)
		switch valueType {
		case TypeDate:
			return NewDateValue(value), nil
		case TypeTime:
			return NewTimeValue(value), nil
		case TypeDateTime:
			return NewDateTimeValue(value), nil
		case TypeTimestamp:
			return NewTimestampValue(value), nil
		}

	default:
		return nil, errors.NewErrorf(errors.ErrCodeInvalidParameter, "Unsupported value type: %v", valueType)
	}

	return nil, errors.NewError(errors.ErrCodeInternalError, "Unexpected error in deserialization")
}

// ConvertValue 类型转换
// ConvertValue converts value to target type
func ConvertValue(value Value, targetType ValueType) (Value, error) {
	if value.IsNull() {
		return NewNullValue(), nil
	}

	if value.Type() == targetType {
		return value.Clone(), nil
	}

	switch targetType {
	case TypeNull:
		return NewNullValue(), nil

	case TypeBool:
		val, err := value.ToBool()
		if err != nil {
			return nil, err
		}
		return NewBoolValue(val), nil

	case TypeInt8:
		val, err := value.ToInt64()
		if err != nil {
			return nil, err
		}
		if val < math.MinInt8 || val > math.MaxInt8 {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %d out of range for int8", val)
		}
		return NewInt8Value(int8(val)), nil

	case TypeInt16:
		val, err := value.ToInt64()
		if err != nil {
			return nil, err
		}
		if val < math.MinInt16 || val > math.MaxInt16 {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %d out of range for int16", val)
		}
		return NewInt16Value(int16(val)), nil

	case TypeInt32:
		val, err := value.ToInt64()
		if err != nil {
			return nil, err
		}
		if val < math.MinInt32 || val > math.MaxInt32 {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %d out of range for int32", val)
		}
		return NewInt32Value(int32(val)), nil

	case TypeInt64:
		val, err := value.ToInt64()
		if err != nil {
			return nil, err
		}
		return NewInt64Value(val), nil

	case TypeFloat32:
		val, err := value.ToFloat64()
		if err != nil {
			return nil, err
		}
		if math.Abs(val) > math.MaxFloat32 {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %f out of range for float32", val)
		}
		return NewFloat32Value(float32(val)), nil

	case TypeFloat64:
		val, err := value.ToFloat64()
		if err != nil {
			return nil, err
		}
		return NewFloat64Value(val), nil

	case TypeString:
		val, err := value.ToString()
		if err != nil {
			return nil, err
		}
		if len(val) > constants.MaxValueLength {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "String length %d exceeds maximum %d", len(val), constants.MaxValueLength)
		}
		return NewStringValue(val), nil

	case TypeBlob:
		val, err := value.ToBytes()
		if err != nil {
			return nil, err
		}
		if len(val) > constants.MaxValueLength {
			return nil, errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Blob length %d exceeds maximum %d", len(val), constants.MaxValueLength)
		}
		return NewBlobValue(val), nil

	case TypeDate:
		val, err := value.ToTime()
		if err != nil {
			return nil, err
		}
		return NewDateValue(val), nil

	case TypeTime:
		val, err := value.ToTime()
		if err != nil {
			return nil, err
		}
		return NewTimeValue(val), nil

	case TypeDateTime:
		val, err := value.ToTime()
		if err != nil {
			return nil, err
		}
		return NewDateTimeValue(val), nil

	case TypeTimestamp:
		val, err := value.ToTime()
		if err != nil {
			return nil, err
		}
		return NewTimestampValue(val), nil

	default:
		return nil, errors.NewErrorf(errors.ErrCodeInvalidParameter, "Unsupported target type: %v", targetType)
	}
}

// ValidateValue 验证值的有效性
// ValidateValue validates value
func ValidateValue(value Value, constraints map[string]interface{}) error {
	if value.IsNull() {
		if required, ok := constraints["required"]; ok && required.(bool) {
			return errors.NewError(errors.ErrCodeRequiredField, "Value cannot be NULL")
		}
		return nil
	}

	// 检查长度限制
	if maxLen, ok := constraints["max_length"]; ok {
		str, err := value.ToString()
		if err == nil && utf8.RuneCountInString(str) > maxLen.(int) {
			return errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value length %d exceeds maximum %d", utf8.RuneCountInString(str), maxLen.(int))
		}
	}

	// 检查数值范围
	if value.Type().IsNumeric() {
		if minVal, ok := constraints["min_value"]; ok {
			if floatVal, err := value.ToFloat64(); err == nil && floatVal < minVal.(float64) {
				return errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %f is less than minimum %f", floatVal, minVal.(float64))
			}
		}

		if maxVal, ok := constraints["max_value"]; ok {
			if floatVal, err := value.ToFloat64(); err == nil && floatVal > maxVal.(float64) {
				return errors.NewErrorf(errors.ErrCodeValueOutOfRange, "Value %f is greater than maximum %f", floatVal, maxVal.(float64))
			}
		}
	}

	return nil
}

// CompareValues 比较两个值数组
// CompareValues compares two value arrays
func CompareValues(left, right []Value) CompareResult {
	minLen := len(left)
	if len(right) < minLen {
		minLen = len(right)
	}

	for i := 0; i < minLen; i++ {
		result := left[i].Compare(right[i])
		if result != CompareEqual {
			return result
		}
	}

	if len(left) < len(right) {
		return CompareLess
	} else if len(left) > len(right) {
		return CompareGreater
	}

	return CompareEqual
}

// HashValues 计算值数组的哈希值
// HashValues calculates hash value for value array
func HashValues(values []Value) uint64 {
	var h uint64 = 5381
	for _, v := range values {
		h = ((h << 5) + h) + v.Hash()
	}
	return h
}

// ValuesToJSON 将值数组转换为JSON
// ValuesToJSON converts value array to JSON
func ValuesToJSON(values []Value) ([]byte, error) {
	result := make([]interface{}, len(values))

	for i, v := range values {
		if v.IsNull() {
			result[i] = nil
		} else {
			switch v.Type() {
			case TypeBool:
				val, _ := v.ToBool()
				result[i] = val
			case TypeInt8, TypeInt16, TypeInt32, TypeInt64:
				val, _ := v.ToInt64()
				result[i] = val
			case TypeFloat32, TypeFloat64:
				val, _ := v.ToFloat64()
				result[i] = val
			case TypeString, TypeBlob:
				val, _ := v.ToString()
				result[i] = val
			case TypeDate, TypeTime, TypeDateTime, TypeTimestamp:
				val, _ := v.ToString()
				result[i] = val
			default:
				val, _ := v.ToString()
				result[i] = val
			}
		}
	}

	return json.Marshal(result)
}
