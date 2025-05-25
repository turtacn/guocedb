// Package value defines the types used to represent SQL values within Guocedb.
// This includes basic data types and their Go equivalents, as well as
// potentially more complex types needed for the query execution engine.
//
// 此包定义了 Guocedb 中用于表示 SQL 值的类型。
// 这包括基本数据类型及其 Go 语言的对应类型，以及
// 查询执行引擎可能需要的更复杂类型。
package value

import (
	"fmt"
	"strconv"
	"time"
)

// Type represents the abstract type of a SQL value.
// This is similar to MySQL data types but may have a more abstract representation internally.
//
// Type 表示 SQL 值的抽象类型。
// 这类似于 MySQL 数据类型，但在内部可能具有更抽象的表示。
type Type int

const (
	// TypeUnknown represents an unknown or unspecified SQL type.
	// 未知或未指定的 SQL 类型。
	TypeUnknown Type = iota
	// TypeNull represents the NULL SQL value.
	// NULL SQL 值。
	TypeNull
	// TypeBoolean represents the BOOLEAN SQL type.
	// BOOLEAN SQL 类型。
	TypeBoolean
	// TypeTinyInt represents the TINYINT SQL type (8-bit signed integer).
	// TINYINT SQL 类型（8 位有符号整数）。
	TypeTinyInt
	// TypeSmallInt represents the SMALLINT SQL type (16-bit signed integer).
	// SMALLINT SQL 类型（16 位有符号整数）。
	TypeSmallInt
	// TypeInt represents the INT or INTEGER SQL type (32-bit signed integer).
	// INT 或 INTEGER SQL 类型（32 位有符号整数）。
	TypeInt
	// TypeBigInt represents the BIGINT SQL type (64-bit signed integer).
	// BIGINT SQL 类型（64 位有符号整数）。
	TypeBigInt
	// TypeFloat represents the FLOAT SQL type (single-precision floating-point number).
	// FLOAT SQL 类型（单精度浮点数）。
	TypeFloat
	// TypeDouble represents the DOUBLE or REAL SQL type (double-precision floating-point number).
	// DOUBLE 或 REAL SQL 类型（双精度浮点数）。
	TypeDouble
	// TypeText represents variable-length string data.
	// 可变长度字符串数据。
	TypeText
	// TypeBlob represents binary large objects.
	// 二进制大对象。
	TypeBlob
	// TypeDateTime represents the DATETIME SQL type.
	// DATETIME SQL 类型。
	TypeDateTime
	// TypeDate represents the DATE SQL type.
	// DATE SQL 类型。
	TypeDate
	// TypeTime represents the TIME SQL type.
	// TIME SQL 类型。
	TypeTime
)

// String returns the string representation of a Type.
// String 方法返回 Type 的字符串表示。
func (t Type) String() string {
	switch t {
	case TypeUnknown:
		return "UNKNOWN"
	case TypeNull:
		return "NULL"
	case TypeBoolean:
		return "BOOLEAN"
	case TypeTinyInt:
		return "TINYINT"
	case TypeSmallInt:
		return "SMALLINT"
	case TypeInt:
		return "INT"
	case TypeBigInt:
		return "BIGINT"
	case TypeFloat:
		return "FLOAT"
	case TypeDouble:
		return "DOUBLE"
	case TypeText:
		return "TEXT"
	case TypeBlob:
		return "BLOB"
	case TypeDateTime:
		return "DATETIME"
	case TypeDate:
		return "DATE"
	case TypeTime:
		return "TIME"
	default:
		return fmt.Sprintf("UNKNOWN_TYPE(%d)", t)
	}
}

// Value represents a SQL value. It holds the Go representation of the value
// and its corresponding SQL type.
//
// Value 结构体表示一个 SQL 值。它持有该值的 Go 语言表示
// 及其对应的 SQL 类型。
type Value struct {
	Type  Type
	Value interface{}
}

// NewNull creates a new NULL Value.
// NewNull 函数创建一个新的 NULL Value。
func NewNull() Value {
	return Value{Type: TypeNull, Value: nil}
}

// NewBoolean creates a new BOOLEAN Value.
// NewBoolean 函数创建一个新的 BOOLEAN Value。
func NewBoolean(v bool) Value {
	return Value{Type: TypeBoolean, Value: v}
}

// NewTinyInt creates a new TINYINT Value.
// NewTinyInt 函数创建一个新的 TINYINT Value。
func NewTinyInt(v int8) Value {
	return Value{Type: TypeTinyInt, Value: v}
}

// NewSmallInt creates a new SMALLINT Value.
// NewSmallInt 函数创建一个新的 SMALLINT Value。
func NewSmallInt(v int16) Value {
	return Value{Type: TypeSmallInt, Value: v}
}

// NewInt creates a new INT Value.
// NewInt 函数创建一个新的 INT Value。
func NewInt(v int32) Value {
	return Value{Type: TypeInt, Value: v}
}

// NewBigInt creates a new BIGINT Value.
// NewBigInt 函数创建一个新的 BIGINT Value。
func NewBigInt(v int64) Value {
	return Value{Type: TypeBigInt, Value: v}
}

// NewFloat creates a new FLOAT Value.
// NewFloat 函数创建一个新的 FLOAT Value。
func NewFloat(v float32) Value {
	return Value{Type: TypeFloat, Value: v}
}

// NewDouble creates a new DOUBLE Value.
// NewDouble 函数创建一个新的 DOUBLE Value。
func NewDouble(v float64) Value {
	return Value{Type: TypeDouble, Value: v}
}

// NewText creates a new TEXT Value.
// NewText 函数创建一个新的 TEXT Value。
func NewText(v string) Value {
	return Value{Type: TypeText, Value: v}
}

// NewBlob creates a new BLOB Value.
// NewBlob 函数创建一个新的 BLOB Value。
func NewBlob(v []byte) Value {
	return Value{Type: TypeBlob, Value: v}
}

// NewDateTime creates a new DATETIME Value.
// It expects a time.Time value.
// NewDateTime 函数创建一个新的 DATETIME Value。
// 它期望一个 time.Time 类型的值。
func NewDateTime(v time.Time) Value {
	return Value{Type: TypeDateTime, Value: v}
}

// NewDate creates a new DATE Value.
// It expects a time.Time value (only the date part will be relevant).
// NewDate 函数创建一个新的 DATE Value。
// 它期望一个 time.Time 类型的值（只有日期部分是相关的）。
func NewDate(v time.Time) Value {
	return Value{Type: TypeDate, Value: v}
}

// NewTime creates a new TIME Value.
// It expects a time.Time value (only the time part will be relevant).
// NewTime 函数创建一个新的 TIME Value。
// 它期望一个 time.Time 类型的值（只有时间部分是相关的）。
func NewTime(v time.Time) Value {
	return Value{Type: TypeTime, Value: v}
}

// AsBool attempts to convert the Value to a boolean.
// Returns false and an error if the conversion is not possible.
// AsBool 尝试将 Value 转换为布尔值。
// 如果转换不可能，则返回 false 和一个错误。
func (v Value) AsBool() (bool, error) {
	if v.Type == TypeBoolean {
		return v.Value.(bool), nil
	}
	if v.Type == TypeTinyInt || v.Type == TypeSmallInt || v.Type == TypeInt || v.Type == TypeBigInt {
		return v.Value.(int64) != 0, nil
	}
	if v.Type == TypeFloat || v.Type == TypeDouble {
		return v.Value.(float64) != 0, nil
	}
	if v.Type == TypeText {
		s := strings.ToLower(v.Value.(string))
		if s == "true" || s == "1" {
			return true, nil
		}
		if s == "false" || s == "0" {
			return false, nil
		}
	}
	return false, fmt.Errorf("cannot convert %s to boolean", v.Type)
}

// AsInt64 attempts to convert the Value to an int64.
// Returns 0 and an error if the conversion is not possible.
// AsInt64 尝试将 Value 转换为 int64 类型。
// 如果转换不可能，则返回 0 和一个错误。
func (v Value) AsInt64() (int64, error) {
	switch v.Type {
	case TypeTinyInt:
		return int64(v.Value.(int8)), nil
	case TypeSmallInt:
		return int64(v.Value.(int16)), nil
	case TypeInt:
		return int64(v.Value.(int32)), nil
	case TypeBigInt:
		return v.Value.(int64), nil
	case TypeBoolean:
		if v.Value.(bool) {
			return 1, nil
		}
		return 0, nil
	case TypeFloat:
		return int64(v.Value.(float32)), nil
	case TypeDouble:
		return int64(v.Value.(float64)), nil
	case TypeText:
		i, err := strconv.ParseInt(v.Value.(string), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert text '%s' to int64: %w", v.Value.(string), err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %s to int64", v.Type)
	}
}

// AsFloat64 attempts to convert the Value to a float64.
// Returns 0 and an error if the conversion is not possible.
// AsFloat64 尝试将 Value 转换为 float64 类型。
// 如果转换不可能，则返回 0 和一个错误。
func (v Value) AsFloat64() (float64, error) {
	switch v.Type {
	case TypeTinyInt:
		return float64(v.Value.(int8)), nil
	case TypeSmallInt:
		return float64(v.Value.(int16)), nil
	case TypeInt:
		return float64(v.Value.(int32)), nil
	case TypeBigInt:
		return float64(v.Value.(int64)), nil
	case TypeFloat:
		return float64(v.Value.(float32)), nil
	case TypeDouble:
		return v.Value.(float64), nil
	case TypeBoolean:
		if v.Value.(bool) {
			return 1.0, nil
		}
		return 0.0, nil
	case TypeText:
		f, err := strconv.ParseFloat(v.Value.(string), 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert text '%s' to float64: %w", v.Value.(string), err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %s to float64", v.Type)
	}
}

// AsString attempts to convert the Value to a string.
// Returns an empty string and an error if the conversion is not possible.
// AsString 尝试将 Value 转换为字符串类型。
// 如果转换不可能，则返回空字符串和一个错误。
func (v Value) AsString() (string, error) {
	switch v.Type {
	case TypeNull:
		return "", nil
	case TypeBoolean:
		return strconv.FormatBool(v.Value.(bool)), nil
	case TypeTinyInt:
		return strconv.FormatInt(int64(v.Value.(int8)), 10), nil
	case TypeSmallInt:
		return strconv.FormatInt(int64(v.Value.(int16)), 10), nil
	case TypeInt:
		return strconv.FormatInt(int64(v.Value.(int32)), 10), nil
	case TypeBigInt:
		return strconv.FormatInt(v.Value.(int64), 10), nil
	case TypeFloat:
		return strconv.FormatFloat(float64(v.Value.(float32)), 'g', -1, 32), nil
	case TypeDouble:
		return strconv.FormatFloat(v.Value.(float64), 'g', -1, 64), nil
	case TypeText:
		return v.Value.(string), nil
	case TypeBlob:
		return string(v.Value.([]byte)), nil // Be cautious with large blobs as strings
	case TypeDateTime:
		return v.Value.(time.Time).Format(time.RFC3339), nil // Or a more MySQL-like format
	case TypeDate:
		return v.Value.(time.Time).Format("2006-01-02"), nil
	case TypeTime:
		return v.Value.(time.Time).Format("15:04:05"), nil
	default:
		return "", fmt.Errorf("cannot convert %s to string", v.Type)
	}
}

// AsTime attempts to convert the Value to a time.Time.
// The interpretation depends on the underlying SQL type.
// Returns the zero time and an error if the conversion is not possible.
// AsTime 尝试将 Value 转换为 time.Time 类型。
// 具体的解释取决于底层的 SQL 类型。
// 如果转换不可能，则返回零时间和错误。
func (v Value) AsTime() (time.Time, error) {
	switch v.Type {
	case TypeDateTime:
		return v.Value.(time.Time), nil
	case TypeDate:
		return v.Value.(time.Time), nil
	case TypeTime:
		return v.Value.(time.Time), nil
	case TypeText:
		// Attempt to parse from common formats, adjust as needed
		formats := []string{
			time.RFC3339,
			"2006-01-02",
			"15:04:05",
			"2006-01-02 15:04:05",
		}
		s := v.Value.(string)
		for _, format := range formats {
			t, err := time.Parse(format, s)
			if err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse text '%s' as time", s)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %s to time", v.Type)
	}
}

// Equals checks if this Value is equal to another Value.
// It performs type-aware comparison.
// Equals 方法检查此 Value 是否等于另一个 Value。
// 它执行类型感知的比较。
func (v Value) Equals(other Value) bool {
	if v.Type != other.Type {
		return false
	}
	switch v.Type {
	case TypeNull:
		return true
	case TypeBoolean:
		return v.Value.(bool) == other.Value.(bool)
	case TypeTinyInt:
		return v.Value.(int8) == other.Value.(int8)
	case TypeSmallInt:
		return v.Value.(int16) == other.Value.(int16)
	case TypeInt:
		return v.Value.(int32) == other.Value.(int32)
	case TypeBigInt:
		return v.Value.(int64) == other.Value.(int64)
	case TypeFloat:
		return v.Value.(float32) == other.Value.(float32)
	case TypeDouble:
		return v.Value.(float64) == other.Value.(float64)
	case TypeText:
		return v.Value.(string) == other.Value.(string)
	case TypeBlob:
		b1 := v.Value.([]byte)
		b2 := other.Value.([]byte)
		if len(b1) != len(b2) {
			return false
		}
		for i := range b1 {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	case TypeDateTime, TypeDate, TypeTime:
		t1 := v.Value.(time.Time)
		t2 := other.Value.(time.Time)
		return t1.Equal(t2)
	default:
		return false
	}
}
