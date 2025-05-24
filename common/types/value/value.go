// Package value 提供了 GuoceDB 的值系统实现
// Package value provides value system implementation for GuoceDB
package value

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/guocedb/guocedb/common/errors"
	"github.com/guocedb/guocedb/common/types/enum"
)

// ===== 值结构体定义 Value Structure Definition =====

// Value 表示数据库中的一个值
// Value represents a value in database
type Value struct {
	Type      enum.DataType // 数据类型 Data type
	IsNull    bool          // 是否为NULL Is NULL
	Val       interface{}   // 实际值 Actual value
	Collation string        // 字符集排序规则 Collation
}

// ===== 构造函数 Constructors =====

// NewNull 创建NULL值
// NewNull creates NULL value
func NewNull() *Value {
	return &Value{
		Type:   enum.TypeUnknown,
		IsNull: true,
		Val:    nil,
	}
}

// NewBool 创建布尔值
// NewBool creates boolean value
func NewBool(v bool) *Value {
	return &Value{
		Type:   enum.TypeBool,
		IsNull: false,
		Val:    v,
	}
}

// NewInt8 创建TINYINT值
// NewInt8 creates TINYINT value
func NewInt8(v int8) *Value {
	return &Value{
		Type:   enum.TypeTinyInt,
		IsNull: false,
		Val:    v,
	}
}

// NewInt16 创建SMALLINT值
// NewInt16 creates SMALLINT value
func NewInt16(v int16) *Value {
	return &Value{
		Type:   enum.TypeSmallInt,
		IsNull: false,
		Val:    v,
	}
}

// NewInt32 创建INT值
// NewInt32 creates INT value
func NewInt32(v int32) *Value {
	return &Value{
		Type:   enum.TypeInt,
		IsNull: false,
		Val:    v,
	}
}

// NewInt64 创建BIGINT值
// NewInt64 creates BIGINT value
func NewInt64(v int64) *Value {
	return &Value{
		Type:   enum.TypeBigInt,
		IsNull: false,
		Val:    v,
	}
}

// NewFloat32 创建FLOAT值
// NewFloat32 creates FLOAT value
func NewFloat32(v float32) *Value {
	return &Value{
		Type:   enum.TypeFloat,
		IsNull: false,
		Val:    v,
	}
}

// NewFloat64 创建DOUBLE值
// NewFloat64 creates DOUBLE value
func NewFloat64(v float64) *Value {
	return &Value{
		Type:   enum.TypeDouble,
		IsNull: false,
		Val:    v,
	}
}

// NewDecimal 创建DECIMAL值
// NewDecimal creates DECIMAL value
func NewDecimal(v *big.Float) *Value {
	return &Value{
		Type:   enum.TypeDecimal,
		IsNull: false,
		Val:    v,
	}
}

// NewString 创建字符串值
// NewString creates string value
func NewString(v string) *Value {
	return &Value{
		Type:      enum.TypeVarchar,
		IsNull:    false,
		Val:       v,
		Collation: "utf8mb4_general_ci",
	}
}

// NewStringWithType 创建指定类型的字符串值
// NewStringWithType creates string value with specified type
func NewStringWithType(v string, t enum.DataType) *Value {
	return &Value{
		Type:      t,
		IsNull:    false,
		Val:       v,
		Collation: "utf8mb4_general_ci",
	}
}

// NewBytes 创建二进制值
// NewBytes creates binary value
func NewBytes(v []byte) *Value {
	return &Value{
		Type:   enum.TypeVarBinary,
		IsNull: false,
		Val:    v,
	}
}

// NewBytesWithType 创建指定类型的二进制值
// NewBytesWithType creates binary value with specified type
func NewBytesWithType(v []byte, t enum.DataType) *Value {
	return &Value{
		Type:   t,
		IsNull: false,
		Val:    v,
	}
}

// NewDate 创建DATE值
// NewDate creates DATE value
func NewDate(v time.Time) *Value {
	return &Value{
		Type:   enum.TypeDate,
		IsNull: false,
		Val:    v.Truncate(24 * time.Hour),
	}
}

// NewTime 创建TIME值
// NewTime creates TIME value
func NewTime(v time.Duration) *Value {
	return &Value{
		Type:   enum.TypeTime,
		IsNull: false,
		Val:    v,
	}
}

// NewDateTime 创建DATETIME值
// NewDateTime creates DATETIME value
func NewDateTime(v time.Time) *Value {
	return &Value{
		Type:   enum.TypeDateTime,
		IsNull: false,
		Val:    v,
	}
}

// NewTimestamp 创建TIMESTAMP值
// NewTimestamp creates TIMESTAMP value
func NewTimestamp(v time.Time) *Value {
	return &Value{
		Type:   enum.TypeTimestamp,
		IsNull: false,
		Val:    v.UTC(),
	}
}

// NewJSON 创建JSON值
// NewJSON creates JSON value
func NewJSON(v interface{}) *Value {
	return &Value{
		Type:   enum.TypeJSON,
		IsNull: false,
		Val:    v,
	}
}

// ===== 类型转换 Type Conversion =====

// AsBool 转换为布尔值
// AsBool converts to boolean
func (v *Value) AsBool() (bool, error) {
	if v.IsNull {
		return false, errors.New(errors.ErrInvalidDataType, "cannot convert NULL to bool")
	}

	switch v.Type {
	case enum.TypeBool:
		return v.Val.(bool), nil
	case enum.TypeTinyInt:
		return v.Val.(int8) != 0, nil
	case enum.TypeSmallInt:
		return v.Val.(int16) != 0, nil
	case enum.TypeInt:
		return v.Val.(int32) != 0, nil
	case enum.TypeBigInt:
		return v.Val.(int64) != 0, nil
	case enum.TypeFloat:
		return v.Val.(float32) != 0, nil
	case enum.TypeDouble:
		return v.Val.(float64) != 0, nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		s := strings.ToLower(v.Val.(string))
		return s == "true" || s == "1" || s == "yes" || s == "on", nil
	default:
		return false, errors.New(errors.ErrInvalidDataType,
			"cannot convert %s to bool", v.Type)
	}
}

// AsInt64 转换为64位整数
// AsInt64 converts to 64-bit integer
func (v *Value) AsInt64() (int64, error) {
	if v.IsNull {
		return 0, errors.New(errors.ErrInvalidDataType, "cannot convert NULL to int64")
	}

	switch v.Type {
	case enum.TypeTinyInt:
		return int64(v.Val.(int8)), nil
	case enum.TypeSmallInt:
		return int64(v.Val.(int16)), nil
	case enum.TypeInt:
		return int64(v.Val.(int32)), nil
	case enum.TypeBigInt:
		return v.Val.(int64), nil
	case enum.TypeFloat:
		return int64(v.Val.(float32)), nil
	case enum.TypeDouble:
		return int64(v.Val.(float64)), nil
	case enum.TypeDecimal:
		i, _ := v.Val.(*big.Float).Int64()
		return i, nil
	case enum.TypeBool:
		if v.Val.(bool) {
			return 1, nil
		}
		return 0, nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		return strconv.ParseInt(v.Val.(string), 10, 64)
	default:
		return 0, errors.New(errors.ErrInvalidDataType,
			"cannot convert %s to int64", v.Type)
	}
}

// AsFloat64 转换为64位浮点数
// AsFloat64 converts to 64-bit float
func (v *Value) AsFloat64() (float64, error) {
	if v.IsNull {
		return 0, errors.New(errors.ErrInvalidDataType, "cannot convert NULL to float64")
	}

	switch v.Type {
	case enum.TypeTinyInt:
		return float64(v.Val.(int8)), nil
	case enum.TypeSmallInt:
		return float64(v.Val.(int16)), nil
	case enum.TypeInt:
		return float64(v.Val.(int32)), nil
	case enum.TypeBigInt:
		return float64(v.Val.(int64)), nil
	case enum.TypeFloat:
		return float64(v.Val.(float32)), nil
	case enum.TypeDouble:
		return v.Val.(float64), nil
	case enum.TypeDecimal:
		f, _ := v.Val.(*big.Float).Float64()
		return f, nil
	case enum.TypeBool:
		if v.Val.(bool) {
			return 1.0, nil
		}
		return 0.0, nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		return strconv.ParseFloat(v.Val.(string), 64)
	default:
		return 0, errors.New(errors.ErrInvalidDataType,
			"cannot convert %s to float64", v.Type)
	}
}

// AsString 转换为字符串
// AsString converts to string
func (v *Value) AsString() (string, error) {
	if v.IsNull {
		return "", errors.New(errors.ErrInvalidDataType, "cannot convert NULL to string")
	}

	switch v.Type {
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		return v.Val.(string), nil
	case enum.TypeBool:
		if v.Val.(bool) {
			return "true", nil
		}
		return "false", nil
	case enum.TypeTinyInt:
		return strconv.FormatInt(int64(v.Val.(int8)), 10), nil
	case enum.TypeSmallInt:
		return strconv.FormatInt(int64(v.Val.(int16)), 10), nil
	case enum.TypeInt:
		return strconv.FormatInt(int64(v.Val.(int32)), 10), nil
	case enum.TypeBigInt:
		return strconv.FormatInt(v.Val.(int64), 10), nil
	case enum.TypeFloat:
		return strconv.FormatFloat(float64(v.Val.(float32)), 'g', -1, 32), nil
	case enum.TypeDouble:
		return strconv.FormatFloat(v.Val.(float64), 'g', -1, 64), nil
	case enum.TypeDecimal:
		return v.Val.(*big.Float).String(), nil
	case enum.TypeDate:
		return v.Val.(time.Time).Format("2006-01-02"), nil
	case enum.TypeTime:
		return formatDuration(v.Val.(time.Duration)), nil
	case enum.TypeDateTime:
		return v.Val.(time.Time).Format("2006-01-02 15:04:05"), nil
	case enum.TypeTimestamp:
		return v.Val.(time.Time).Format("2006-01-02 15:04:05"), nil
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		return string(v.Val.([]byte)), nil
	case enum.TypeJSON:
		b, err := json.Marshal(v.Val)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", errors.New(errors.ErrInvalidDataType,
			"cannot convert %s to string", v.Type)
	}
}

// AsBytes 转换为字节数组
// AsBytes converts to byte array
func (v *Value) AsBytes() ([]byte, error) {
	if v.IsNull {
		return nil, errors.New(errors.ErrInvalidDataType, "cannot convert NULL to bytes")
	}

	switch v.Type {
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		return v.Val.([]byte), nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		return []byte(v.Val.(string)), nil
	default:
		// 其他类型先转换为字符串再转换为字节
		// Other types convert to string first then to bytes
		s, err := v.AsString()
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	}
}

// AsTime 转换为时间
// AsTime converts to time
func (v *Value) AsTime() (time.Time, error) {
	if v.IsNull {
		return time.Time{}, errors.New(errors.ErrInvalidDataType, "cannot convert NULL to time")
	}

	switch v.Type {
	case enum.TypeDate, enum.TypeDateTime, enum.TypeTimestamp:
		return v.Val.(time.Time), nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		// 尝试多种时间格式
		// Try multiple time formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02",
			time.RFC3339,
			time.RFC3339Nano,
		}
		s := v.Val.(string)
		for _, format := range formats {
			if t, err := time.Parse(format, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, errors.New(errors.ErrInvalidDatetime, "invalid datetime: %s", s)
	case enum.TypeInt, enum.TypeBigInt:
		// Unix时间戳
		// Unix timestamp
		i, _ := v.AsInt64()
		return time.Unix(i, 0), nil
	default:
		return time.Time{}, errors.New(errors.ErrInvalidDataType,
			"cannot convert %s to time", v.Type)
	}
}

// ===== 比较操作 Comparison Operations =====

// Compare 比较两个值
// Compare compares two values
func (v *Value) Compare(other *Value) (int, error) {
	// NULL值的比较
	// NULL value comparison
	if v.IsNull || other.IsNull {
		if v.IsNull && other.IsNull {
			return 0, nil
		}
		if v.IsNull {
			return -1, nil
		}
		return 1, nil
	}

	// 相同类型直接比较
	// Direct comparison for same type
	if v.Type == other.Type {
		return v.comparesSameType(other)
	}

	// 不同类型需要转换
	// Different types need conversion
	return v.compareDifferentType(other)
}

// compareSameType 比较相同类型的值
// compareSameType compares values of same type
func (v *Value) comparesSameType(other *Value) (int, error) {
	switch v.Type {
	case enum.TypeBool:
		v1, v2 := v.Val.(bool), other.Val.(bool)
		if v1 == v2 {
			return 0, nil
		}
		if !v1 && v2 {
			return -1, nil
		}
		return 1, nil

	case enum.TypeTinyInt:
		v1, v2 := v.Val.(int8), other.Val.(int8)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeSmallInt:
		v1, v2 := v.Val.(int16), other.Val.(int16)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeInt:
		v1, v2 := v.Val.(int32), other.Val.(int32)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeBigInt:
		v1, v2 := v.Val.(int64), other.Val.(int64)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeFloat:
		v1, v2 := v.Val.(float32), other.Val.(float32)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeDouble:
		v1, v2 := v.Val.(float64), other.Val.(float64)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case enum.TypeDecimal:
		v1, v2 := v.Val.(*big.Float), other.Val.(*big.Float)
		return v1.Cmp(v2), nil

	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		v1, v2 := v.Val.(string), other.Val.(string)
		// TODO: 使用collation进行比较
		// TODO: use collation for comparison
		return strings.Compare(v1, v2), nil

	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		v1, v2 := v.Val.([]byte), other.Val.([]byte)
		return bytes.Compare(v1, v2), nil

	case enum.TypeDate, enum.TypeDateTime, enum.TypeTimestamp:
		v1, v2 := v.Val.(time.Time), other.Val.(time.Time)
		if v1.Before(v2) {
			return -1, nil
		} else if v1.After(v2) {
			return 1, nil
		}
		return 0, nil

	case enum.TypeTime:
		v1, v2 := v.Val.(time.Duration), other.Val.(time.Duration)
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, errors.New(errors.ErrInvalidOperation,
			"cannot compare values of type %s", v.Type)
	}
}

// compareDifferentType 比较不同类型的值
// compareDifferentType compares values of different types
func (v *Value) compareDifferentType(other *Value) (int, error) {
	// 尝试将两个值都转换为相同类型
	// Try to convert both values to same type

	// 如果都是数值类型，转换为float64比较
	// If both are numeric, convert to float64 for comparison
	if v.Type.IsNumeric() && other.Type.IsNumeric() {
		f1, err1 := v.AsFloat64()
		f2, err2 := other.AsFloat64()
		if err1 != nil || err2 != nil {
			return 0, errors.New(errors.ErrInvalidOperation,
				"cannot compare %s with %s", v.Type, other.Type)
		}
		if f1 < f2 {
			return -1, nil
		} else if f1 > f2 {
			return 1, nil
		}
		return 0, nil
	}

	// 如果都是字符串类型，直接比较
	// If both are string types, compare directly
	if v.Type.IsString() && other.Type.IsString() {
		s1, _ := v.AsString()
		s2, _ := other.AsString()
		return strings.Compare(s1, s2), nil
	}

	// 如果都是时间类型，转换为time.Time比较
	// If both are datetime types, convert to time.Time for comparison
	if v.Type.IsDateTime() && other.Type.IsDateTime() {
		t1, err1 := v.AsTime()
		t2, err2 := other.AsTime()
		if err1 != nil || err2 != nil {
			return 0, errors.New(errors.ErrInvalidOperation,
				"cannot compare %s with %s", v.Type, other.Type)
		}
		if t1.Before(t2) {
			return -1, nil
		} else if t1.After(t2) {
			return 1, nil
		}
		return 0, nil
	}

	// 其他情况尝试转换为字符串比较
	// Other cases try to convert to string for comparison
	s1, err1 := v.AsString()
	s2, err2 := other.AsString()
	if err1 != nil || err2 != nil {
		return 0, errors.New(errors.ErrInvalidOperation,
			"cannot compare %s with %s", v.Type, other.Type)
	}
	return strings.Compare(s1, s2), nil
}

// Equals 判断是否相等
// Equals checks if equal
func (v *Value) Equals(other *Value) bool {
	cmp, err := v.Compare(other)
	return err == nil && cmp == 0
}

// LessThan 判断是否小于
// LessThan checks if less than
func (v *Value) LessThan(other *Value) bool {
	cmp, err := v.Compare(other)
	return err == nil && cmp < 0
}

// GreaterThan 判断是否大于
// GreaterThan checks if greater than
func (v *Value) GreaterThan(other *Value) bool {
	cmp, err := v.Compare(other)
	return err == nil && cmp > 0
}

// ===== 算术操作 Arithmetic Operations =====

// Add 加法
// Add addition
func (v *Value) Add(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	// 字符串连接
	// String concatenation
	if v.Type.IsString() || other.Type.IsString() {
		s1, err1 := v.AsString()
		s2, err2 := other.AsString()
		if err1 != nil || err2 != nil {
			return nil, errors.New(errors.ErrInvalidOperation,
				"cannot add %s and %s", v.Type, other.Type)
		}
		return NewString(s1 + s2), nil
	}

	// 数值加法
	// Numeric addition
	if v.Type.IsNumeric() && other.Type.IsNumeric() {
		// 根据类型选择合适的运算
		// Choose appropriate operation based on type
		if v.Type == enum.TypeDecimal || other.Type == enum.TypeDecimal {
			// DECIMAL运算
			// DECIMAL operation
			d1 := v.toDecimal()
			d2 := other.toDecimal()
			result := new(big.Float).Add(d1, d2)
			return NewDecimal(result), nil
		} else if v.Type == enum.TypeDouble || other.Type == enum.TypeDouble ||
			v.Type == enum.TypeFloat || other.Type == enum.TypeFloat {
			// 浮点运算
			// Float operation
			f1, _ := v.AsFloat64()
			f2, _ := other.AsFloat64()
			return NewFloat64(f1 + f2), nil
		} else {
			// 整数运算
			// Integer operation
			i1, _ := v.AsInt64()
			i2, _ := other.AsInt64()
			return NewInt64(i1 + i2), nil
		}
	}

	// 时间运算
	// Time operations
	if v.Type.IsDateTime() && other.Type == enum.TypeTime {
		t, _ := v.AsTime()
		d := other.Val.(time.Duration)
		return NewDateTime(t.Add(d)), nil
	}

	return nil, errors.New(errors.ErrInvalidOperation,
		"cannot add %s and %s", v.Type, other.Type)
}

// Subtract 减法
// Subtract subtraction
func (v *Value) Subtract(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	// 数值减法
	// Numeric subtraction
	if v.Type.IsNumeric() && other.Type.IsNumeric() {
		if v.Type == enum.TypeDecimal || other.Type == enum.TypeDecimal {
			d1 := v.toDecimal()
			d2 := other.toDecimal()
			result := new(big.Float).Sub(d1, d2)
			return NewDecimal(result), nil
		} else if v.Type == enum.TypeDouble || other.Type == enum.TypeDouble ||
			v.Type == enum.TypeFloat || other.Type == enum.TypeFloat {
			f1, _ := v.AsFloat64()
			f2, _ := other.AsFloat64()
			return NewFloat64(f1 - f2), nil
		} else {
			i1, _ := v.AsInt64()
			i2, _ := other.AsInt64()
			return NewInt64(i1 - i2), nil
		}
	}

	// 时间减法
	// Time subtraction
	if v.Type.IsDateTime() && other.Type.IsDateTime() {
		t1, _ := v.AsTime()
		t2, _ := other.AsTime()
		duration := t1.Sub(t2)
		return NewTime(duration), nil
	}

	return nil, errors.New(errors.ErrInvalidOperation,
		"cannot subtract %s from %s", other.Type, v.Type)
}

// Multiply 乘法
// Multiply multiplication
func (v *Value) Multiply(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot multiply %s and %s", v.Type, other.Type)
	}

	if v.Type == enum.TypeDecimal || other.Type == enum.TypeDecimal {
		d1 := v.toDecimal()
		d2 := other.toDecimal()
		result := new(big.Float).Mul(d1, d2)
		return NewDecimal(result), nil
	} else if v.Type == enum.TypeDouble || other.Type == enum.TypeDouble ||
		v.Type == enum.TypeFloat || other.Type == enum.TypeFloat {
		f1, _ := v.AsFloat64()
		f2, _ := other.AsFloat64()
		return NewFloat64(f1 * f2), nil
	} else {
		i1, _ := v.AsInt64()
		i2, _ := other.AsInt64()
		return NewInt64(i1 * i2), nil
	}
}

// Divide 除法
// Divide division
func (v *Value) Divide(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot divide %s by %s", v.Type, other.Type)
	}

	// 检查除零
	// Check division by zero
	if isZero(other) {
		return nil, errors.New(errors.ErrDivisionByZero, "division by zero")
	}

	// 除法总是返回浮点数
	// Division always returns float
	f1, _ := v.AsFloat64()
	f2, _ := other.AsFloat64()
	return NewFloat64(f1 / f2), nil
}

// Modulo 取模
// Modulo modulo operation
func (v *Value) Modulo(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot modulo %s by %s", v.Type, other.Type)
	}

	// 检查除零
	// Check division by zero
	if isZero(other) {
		return nil, errors.New(errors.ErrDivisionByZero, "modulo by zero")
	}

	// 转换为整数进行取模
	// Convert to integer for modulo
	i1, _ := v.AsInt64()
	i2, _ := other.AsInt64()
	return NewInt64(i1 % i2), nil
}

// ===== 逻辑操作 Logical Operations =====

// And 逻辑与
// And logical AND
func (v *Value) And(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		// SQL三值逻辑
		// SQL three-valued logic
		b1, err1 := v.AsBool()
		b2, err2 := other.AsBool()
		if err1 == nil && !b1 {
			return NewBool(false), nil
		}
		if err2 == nil && !b2 {
			return NewBool(false), nil
		}
		return NewNull(), nil
	}

	b1, err1 := v.AsBool()
	b2, err2 := other.AsBool()
	if err1 != nil || err2 != nil {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform AND on %s and %s", v.Type, other.Type)
	}

	return NewBool(b1 && b2), nil
}

// Or 逻辑或
// Or logical OR
func (v *Value) Or(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		// SQL三值逻辑
		// SQL three-valued logic
		b1, err1 := v.AsBool()
		b2, err2 := other.AsBool()
		if err1 == nil && b1 {
			return NewBool(true), nil
		}
		if err2 == nil && b2 {
			return NewBool(true), nil
		}
		return NewNull(), nil
	}

	b1, err1 := v.AsBool()
	b2, err2 := other.AsBool()
	if err1 != nil || err2 != nil {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform OR on %s and %s", v.Type, other.Type)
	}

	return NewBool(b1 || b2), nil
}

// Not 逻辑非
// Not logical NOT
func (v *Value) Not() (*Value, error) {
	if v.IsNull {
		return NewNull(), nil
	}

	b, err := v.AsBool()
	if err != nil {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform NOT on %s", v.Type)
	}

	return NewBool(!b), nil
}

// ===== 位操作 Bitwise Operations =====

// BitAnd 位与
// BitAnd bitwise AND
func (v *Value) BitAnd(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform bitwise AND on %s and %s", v.Type, other.Type)
	}

	i1, _ := v.AsInt64()
	i2, _ := other.AsInt64()
	return NewInt64(i1 & i2), nil
}

// BitOr 位或
// BitOr bitwise OR
func (v *Value) BitOr(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform bitwise OR on %s and %s", v.Type, other.Type)
	}

	i1, _ := v.AsInt64()
	i2, _ := other.AsInt64()
	return NewInt64(i1 | i2), nil
}

// BitXor 位异或
// BitXor bitwise XOR
func (v *Value) BitXor(other *Value) (*Value, error) {
	if v.IsNull || other.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() || !other.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform bitwise XOR on %s and %s", v.Type, other.Type)
	}

	i1, _ := v.AsInt64()
	i2, _ := other.AsInt64()
	return NewInt64(i1 ^ i2), nil
}

// BitNot 位非
// BitNot bitwise NOT
func (v *Value) BitNot() (*Value, error) {
	if v.IsNull {
		return NewNull(), nil
	}

	if !v.Type.IsNumeric() {
		return nil, errors.New(errors.ErrInvalidOperation,
			"cannot perform bitwise NOT on %s", v.Type)
	}

	i, _ := v.AsInt64()
	return NewInt64(^i), nil
}

// ===== 序列化与反序列化 Serialization and Deserialization =====

// Serialize 序列化值
// Serialize serializes value
func (v *Value) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入类型
	// Write type
	if err := binary.Write(buf, binary.LittleEndian, int32(v.Type)); err != nil {
		return nil, err
	}

	// 写入NULL标志
	// Write NULL flag
	if err := binary.Write(buf, binary.LittleEndian, v.IsNull); err != nil {
		return nil, err
	}

	// 如果是NULL，直接返回
	// If NULL, return directly
	if v.IsNull {
		return buf.Bytes(), nil
	}

	// 根据类型序列化值
	// Serialize value based on type
	switch v.Type {
	case enum.TypeBool:
		return serializeBool(buf, v.Val.(bool))
	case enum.TypeTinyInt:
		return serializeInt8(buf, v.Val.(int8))
	case enum.TypeSmallInt:
		return serializeInt16(buf, v.Val.(int16))
	case enum.TypeInt:
		return serializeInt32(buf, v.Val.(int32))
	case enum.TypeBigInt:
		return serializeInt64(buf, v.Val.(int64))
	case enum.TypeFloat:
		return serializeFloat32(buf, v.Val.(float32))
	case enum.TypeDouble:
		return serializeFloat64(buf, v.Val.(float64))
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		return serializeString(buf, v.Val.(string))
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		return serializeBytes(buf, v.Val.([]byte))
	case enum.TypeDate, enum.TypeDateTime, enum.TypeTimestamp:
		return serializeTime(buf, v.Val.(time.Time))
	case enum.TypeTime:
		return serializeDuration(buf, v.Val.(time.Duration))
	case enum.TypeJSON:
		return serializeJSON(buf, v.Val)
	default:
		return nil, errors.New(errors.ErrUnsupported,
			"cannot serialize type %s", v.Type)
	}
}

// Deserialize 反序列化值
// Deserialize deserializes value
func Deserialize(data []byte) (*Value, error) {
	buf := bytes.NewReader(data)

	// 读取类型
	// Read type
	var typeInt int32
	if err := binary.Read(buf, binary.LittleEndian, &typeInt); err != nil {
		return nil, err
	}
	dataType := enum.DataType(typeInt)

	// 读取NULL标志
	// Read NULL flag
	var isNull bool
	if err := binary.Read(buf, binary.LittleEndian, &isNull); err != nil {
		return nil, err
	}

	// 如果是NULL，直接返回
	// If NULL, return directly
	if isNull {
		return NewNull(), nil
	}

	// 根据类型反序列化值
	// Deserialize value based on type
	switch dataType {
	case enum.TypeBool:
		val, err := deserializeBool(buf)
		if err != nil {
			return nil, err
		}
		return NewBool(val), nil
	case enum.TypeTinyInt:
		val, err := deserializeInt8(buf)
		if err != nil {
			return nil, err
		}
		return NewInt8(val), nil
	case enum.TypeSmallInt:
		val, err := deserializeInt16(buf)
		if err != nil {
			return nil, err
		}
		return NewInt16(val), nil
	case enum.TypeInt:
		val, err := deserializeInt32(buf)
		if err != nil {
			return nil, err
		}
		return NewInt32(val), nil
	case enum.TypeBigInt:
		val, err := deserializeInt64(buf)
		if err != nil {
			return nil, err
		}
		return NewInt64(val), nil
	case enum.TypeFloat:
		val, err := deserializeFloat32(buf)
		if err != nil {
			return nil, err
		}
		return NewFloat32(val), nil
	case enum.TypeDouble:
		val, err := deserializeFloat64(buf)
		if err != nil {
			return nil, err
		}
		return NewFloat64(val), nil
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		val, err := deserializeString(buf)
		if err != nil {
			return nil, err
		}
		return NewStringWithType(val, dataType), nil
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		val, err := deserializeBytes(buf)
		if err != nil {
			return nil, err
		}
		return NewBytesWithType(val, dataType), nil
	case enum.TypeDate:
		val, err := deserializeTime(buf)
		if err != nil {
			return nil, err
		}
		return NewDate(val), nil
	case enum.TypeDateTime:
		val, err := deserializeTime(buf)
		if err != nil {
			return nil, err
		}
		return NewDateTime(val), nil
	case enum.TypeTimestamp:
		val, err := deserializeTime(buf)
		if err != nil {
			return nil, err
		}
		return NewTimestamp(val), nil
	case enum.TypeTime:
		val, err := deserializeDuration(buf)
		if err != nil {
			return nil, err
		}
		return NewTime(val), nil
	case enum.TypeJSON:
		val, err := deserializeJSON(buf)
		if err != nil {
			return nil, err
		}
		return NewJSON(val), nil
	default:
		return nil, errors.New(errors.ErrUnsupported,
			"cannot deserialize type %s", dataType)
	}
}

// ===== 辅助函数 Helper Functions =====

// toDecimal 转换为DECIMAL
// toDecimal converts to DECIMAL
func (v *Value) toDecimal() *big.Float {
	if v.Type == enum.TypeDecimal {
		return v.Val.(*big.Float)
	}
	f, _ := v.AsFloat64()
	return big.NewFloat(f)
}

// isZero 判断是否为零
// isZero checks if zero
func isZero(v *Value) bool {
	if v.IsNull {
		return false
	}

	switch v.Type {
	case enum.TypeTinyInt:
		return v.Val.(int8) == 0
	case enum.TypeSmallInt:
		return v.Val.(int16) == 0
	case enum.TypeInt:
		return v.Val.(int32) == 0
	case enum.TypeBigInt:
		return v.Val.(int64) == 0
	case enum.TypeFloat:
		return v.Val.(float32) == 0
	case enum.TypeDouble:
		return v.Val.(float64) == 0
	case enum.TypeDecimal:
		return v.Val.(*big.Float).Sign() == 0
	default:
		return false
	}
}

// formatDuration 格式化时间间隔
// formatDuration formats duration
func formatDuration(d time.Duration) string {
	negative := d < 0
	if negative {
		d = -d
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	result := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	if negative {
		result = "-" + result
	}
	return result
}

// ===== 序列化辅助函数 Serialization Helper Functions =====

func serializeBool(buf *bytes.Buffer, v bool) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeBool(buf *bytes.Reader) (bool, error) {
	var v bool
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeInt8(buf *bytes.Buffer, v int8) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeInt8(buf *bytes.Reader) (int8, error) {
	var v int8
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeInt16(buf *bytes.Buffer, v int16) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeInt16(buf *bytes.Reader) (int16, error) {
	var v int16
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeInt32(buf *bytes.Buffer, v int32) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeInt32(buf *bytes.Reader) (int32, error) {
	var v int32
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeInt64(buf *bytes.Buffer, v int64) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeInt64(buf *bytes.Reader) (int64, error) {
	var v int64
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeFloat32(buf *bytes.Buffer, v float32) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeFloat32(buf *bytes.Reader) (float32, error) {
	var v float32
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeFloat64(buf *bytes.Buffer, v float64) ([]byte, error) {
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeFloat64(buf *bytes.Reader) (float64, error) {
	var v float64
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func serializeString(buf *bytes.Buffer, v string) ([]byte, error) {
	data := []byte(v)
	// 写入长度
	// Write length
	if err := binary.Write(buf, binary.LittleEndian, int32(len(data))); err != nil {
		return nil, err
	}
	// 写入数据
	// Write data
	if _, err := buf.Write(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeString(buf *bytes.Reader) (string, error) {
	// 读取长度
	// Read length
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	// 读取数据
	// Read data
	data := make([]byte, length)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}
	return string(data), nil
}

func serializeBytes(buf *bytes.Buffer, v []byte) ([]byte, error) {
	// 写入长度
	// Write length
	if err := binary.Write(buf, binary.LittleEndian, int32(len(v))); err != nil {
		return nil, err
	}
	// 写入数据
	// Write data
	if _, err := buf.Write(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeBytes(buf *bytes.Reader) ([]byte, error) {
	// 读取长度
	// Read length
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	// 读取数据
	// Read data
	data := make([]byte, length)
	if _, err := buf.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func serializeTime(buf *bytes.Buffer, v time.Time) ([]byte, error) {
	// 序列化为Unix纳秒时间戳
	// Serialize as Unix nanosecond timestamp
	if err := binary.Write(buf, binary.LittleEndian, v.UnixNano()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeTime(buf *bytes.Reader) (time.Time, error) {
	// 反序列化Unix纳秒时间戳
	// Deserialize Unix nanosecond timestamp
	var nanos int64
	if err := binary.Read(buf, binary.LittleEndian, &nanos); err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, nanos), nil
}

func serializeDuration(buf *bytes.Buffer, v time.Duration) ([]byte, error) {
	// 序列化为纳秒
	// Serialize as nanoseconds
	if err := binary.Write(buf, binary.LittleEndian, int64(v)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeDuration(buf *bytes.Reader) (time.Duration, error) {
	// 反序列化纳秒
	// Deserialize nanoseconds
	var nanos int64
	if err := binary.Read(buf, binary.LittleEndian, &nanos); err != nil {
		return 0, err
	}
	return time.Duration(nanos), nil
}

func serializeJSON(buf *bytes.Buffer, v interface{}) ([]byte, error) {
	// 序列化为JSON字符串
	// Serialize as JSON string
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return serializeBytes(buf, data)
}

func deserializeJSON(buf *bytes.Reader) (interface{}, error) {
	// 反序列化JSON字符串
	// Deserialize JSON string
	data, err := deserializeBytes(buf)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// ===== 类型推断 Type Inference =====

// InferType 从Go值推断数据类型
// InferType infers data type from Go value
func InferType(v interface{}) enum.DataType {
	if v == nil {
		return enum.TypeUnknown
	}

	switch v.(type) {
	case bool:
		return enum.TypeBool
	case int8:
		return enum.TypeTinyInt
	case int16:
		return enum.TypeSmallInt
	case int32, int:
		return enum.TypeInt
	case int64:
		return enum.TypeBigInt
	case float32:
		return enum.TypeFloat
	case float64:
		return enum.TypeDouble
	case *big.Float:
		return enum.TypeDecimal
	case string:
		return enum.TypeVarchar
	case []byte:
		return enum.TypeVarBinary
	case time.Time:
		return enum.TypeDateTime
	case time.Duration:
		return enum.TypeTime
	default:
		return enum.TypeUnknown
	}
}

// NewFromInterface 从接口创建值
// NewFromInterface creates value from interface
func NewFromInterface(v interface{}) *Value {
	if v == nil {
		return NewNull()
	}

	switch val := v.(type) {
	case bool:
		return NewBool(val)
	case int8:
		return NewInt8(val)
	case int16:
		return NewInt16(val)
	case int32:
		return NewInt32(val)
	case int:
		return NewInt32(int32(val))
	case int64:
		return NewInt64(val)
	case float32:
		return NewFloat32(val)
	case float64:
		return NewFloat64(val)
	case *big.Float:
		return NewDecimal(val)
	case string:
		return NewString(val)
	case []byte:
		return NewBytes(val)
	case time.Time:
		return NewDateTime(val)
	case time.Duration:
		return NewTime(val)
	default:
		// 尝试JSON序列化
		// Try JSON serialization
		return NewJSON(v)
	}
}

// ===== 聚合操作 Aggregate Operations =====

// Sum 求和多个值
// Sum sums multiple values
func Sum(values []*Value) (*Value, error) {
	if len(values) == 0 {
		return NewNull(), nil
	}

	// 找到第一个非NULL值
	// Find first non-NULL value
	var result *Value
	for _, v := range values {
		if !v.IsNull {
			result = v
			break
		}
	}

	if result == nil {
		return NewNull(), nil
	}

	// 累加所有非NULL值
	// Accumulate all non-NULL values
	for i := 1; i < len(values); i++ {
		if !values[i].IsNull {
			var err error
			result, err = result.Add(values[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// Avg 求平均值
// Avg calculates average
func Avg(values []*Value) (*Value, error) {
	if len(values) == 0 {
		return NewNull(), nil
	}

	sum, err := Sum(values)
	if err != nil {
		return nil, err
	}

	if sum.IsNull {
		return NewNull(), nil
	}

	// 计算非NULL值的数量
	// Count non-NULL values
	count := 0
	for _, v := range values {
		if !v.IsNull {
			count++
		}
	}

	if count == 0 {
		return NewNull(), nil
	}

	return sum.Divide(NewInt64(int64(count)))
}

// Min 求最小值
// Min finds minimum value
func Min(values []*Value) (*Value, error) {
	if len(values) == 0 {
		return NewNull(), nil
	}

	var min *Value
	for _, v := range values {
		if v.IsNull {
			continue
		}
		if min == nil {
			min = v
		} else {
			cmp, err := v.Compare(min)
			if err != nil {
				return nil, err
			}
			if cmp < 0 {
				min = v
			}
		}
	}

	if min == nil {
		return NewNull(), nil
	}
	return min, nil
}

// Max 求最大值
// Max finds maximum value
func Max(values []*Value) (*Value, error) {
	if len(values) == 0 {
		return NewNull(), nil
	}

	var max *Value
	for _, v := range values {
		if v.IsNull {
			continue
		}
		if max == nil {
			max = v
		} else {
			cmp, err := v.Compare(max)
			if err != nil {
				return nil, err
			}
			if cmp > 0 {
				max = v
			}
		}
	}

	if max == nil {
		return NewNull(), nil
	}
	return max, nil
}

// Count 计数非NULL值
// Count counts non-NULL values
func Count(values []*Value) *Value {
	count := 0
	for _, v := range values {
		if !v.IsNull {
			count++
		}
	}
	return NewInt64(int64(count))
}

// ===== 类型验证 Type Validation =====

// ValidateType 验证值是否符合指定类型
// ValidateType validates if value conforms to specified type
func (v *Value) ValidateType(targetType enum.DataType) error {
	if v.IsNull {
		return nil // NULL值对任何类型都有效
		// NULL value is valid for any type
	}

	// 相同类型直接返回
	// Same type returns directly
	if v.Type == targetType {
		return nil
	}

	// 检查是否可以转换
	// Check if conversion is possible
	switch targetType {
	case enum.TypeBool:
		_, err := v.AsBool()
		return err
	case enum.TypeTinyInt, enum.TypeSmallInt, enum.TypeInt, enum.TypeBigInt:
		i, err := v.AsInt64()
		if err != nil {
			return err
		}
		// 检查范围
		// Check range
		return checkIntRange(i, targetType)
	case enum.TypeFloat, enum.TypeDouble:
		_, err := v.AsFloat64()
		return err
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		_, err := v.AsString()
		return err
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob:
		_, err := v.AsBytes()
		return err
	case enum.TypeDate, enum.TypeDateTime, enum.TypeTimestamp:
		_, err := v.AsTime()
		return err
	default:
		return errors.New(errors.ErrInvalidDataType,
			"cannot validate type %s", targetType)
	}
}

// checkIntRange 检查整数范围
// checkIntRange checks integer range
func checkIntRange(v int64, t enum.DataType) error {
	switch t {
	case enum.TypeTinyInt:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return errors.New(errors.ErrDataOutOfRange,
				"value %d out of range for TINYINT", v)
		}
	case enum.TypeSmallInt:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return errors.New(errors.ErrDataOutOfRange,
				"value %d out of range for SMALLINT", v)
		}
	case enum.TypeInt:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return errors.New(errors.ErrDataOutOfRange,
				"value %d out of range for INT", v)
		}
	}
	return nil
}

// ===== 字符串表示 String Representation =====

// String 返回值的字符串表示
// String returns string representation of value
func (v *Value) String() string {
	if v.IsNull {
		return "NULL"
	}

	s, err := v.AsString()
	if err != nil {
		return fmt.Sprintf("<%s: %v>", v.Type, v.Val)
	}
	return s
}

// SQLString 返回SQL字符串表示
// SQLString returns SQL string representation
func (v *Value) SQLString() string {
	if v.IsNull {
		return "NULL"
	}

	switch v.Type {
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText,
		enum.TypeTinyText, enum.TypeMediumText, enum.TypeLongText:
		// 转义单引号
		// Escape single quotes
		s := strings.ReplaceAll(v.Val.(string), "'", "''")
		return fmt.Sprintf("'%s'", s)
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
		enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
		// 十六进制表示
		// Hexadecimal representation
		return fmt.Sprintf("X'%X'", v.Val.([]byte))
	case enum.TypeDate:
		return fmt.Sprintf("'%s'", v.Val.(time.Time).Format("2006-01-02"))
	case enum.TypeTime:
		return fmt.Sprintf("'%s'", formatDuration(v.Val.(time.Duration)))
	case enum.TypeDateTime, enum.TypeTimestamp:
		return fmt.Sprintf("'%s'", v.Val.(time.Time).Format("2006-01-02 15:04:05"))
	default:
		s, _ := v.AsString()
		return s
	}
}

// ===== 哈希计算 Hash Calculation =====

// Hash 计算值的哈希
// Hash calculates hash of value
func (v *Value) Hash() uint64 {
	if v.IsNull {
		return 0
	}

	// 使用FNV-1a哈希算法
	// Use FNV-1a hash algorithm
	const fnvPrime = 1099511628211
	hash := uint64(14695981039346656037)

	// 哈希类型
	// Hash type
	hash ^= uint64(v.Type)
	hash *= fnvPrime

	// 哈希值
	// Hash value
	switch v.Type {
	case enum.TypeBool:
		if v.Val.(bool) {
			hash ^= 1
		}
		hash *= fnvPrime
	case enum.TypeTinyInt:
		hash ^= uint64(v.Val.(int8))
		hash *= fnvPrime
	case enum.TypeSmallInt:
		hash ^= uint64(v.Val.(int16))
		hash *= fnvPrime
	case enum.TypeInt:
		hash ^= uint64(v.Val.(int32))
		hash *= fnvPrime
	case enum.TypeBigInt:
		hash ^= uint64(v.Val.(int64))
		hash *= fnvPrime
	case enum.TypeFloat:
		bits := math.Float32bits(v.Val.(float32))
		hash ^= uint64(bits)
		hash *= fnvPrime
	case enum.TypeDouble:
		bits := math.Float64bits(v.Val.(float64))
		hash ^= bits
		hash *= fnvPrime
	case enum.TypeChar, enum.TypeVarchar, enum.TypeText:
		s := v.Val.(string)
		for i := 0; i < len(s); i++ {
			hash ^= uint64(s[i])
			hash *= fnvPrime
		}
	case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob:
		b := v.Val.([]byte)
		for i := 0; i < len(b); i++ {
			hash ^= uint64(b[i])
			hash *= fnvPrime
		}
	case enum.TypeDate, enum.TypeDateTime, enum.TypeTimestamp:
		t := v.Val.(time.Time)
		hash ^= uint64(t.Unix())
		hash *= fnvPrime
	case enum.TypeTime:
		d := v.Val.(time.Duration)
		hash ^= uint64(d)
		hash *= fnvPrime
	}

	return hash
}

// Clone 克隆值
// Clone clones value
func (v *Value) Clone() *Value {
	if v == nil {
		return nil
	}

	clone := &Value{
		Type:      v.Type,
		IsNull:    v.IsNull,
		Collation: v.Collation,
	}

	if !v.IsNull {
		switch v.Type {
		case enum.TypeBinary, enum.TypeVarBinary, enum.TypeBlob,
			enum.TypeTinyBlob, enum.TypeMediumBlob, enum.TypeLongBlob:
			// 复制字节数组
			// Copy byte array
			src := v.Val.([]byte)
			dst := make([]byte, len(src))
			copy(dst, src)
			clone.Val = dst
		case enum.TypeDecimal:
			// 复制big.Float
			// Copy big.Float
			src := v.Val.(*big.Float)
			dst := new(big.Float).Set(src)
			clone.Val = dst
		default:
			// 其他类型直接赋值
			// Other types assign directly
			clone.Val = v.Val
		}
	}

	return clone
}
