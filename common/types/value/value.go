package value

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types/enum"
)

// Value represents a single SQL value with its type.
// It aims to be the primary internal representation for data manipulation and storage encoding.
// Value represents a single SQL value with its type.
// It aims to be the primary internal representation for data manipulation and storage encoding.
type Value struct {
	typ enum.SQLDataType
	val interface{} // Holds the actual Go value (e.g., int64, string, time.Time, []byte, nil)
}

// nullValue is a singleton representation for SQL NULL.
// We use the type enum.SQLDataTypeNull and a nil val field primarily.
// var nullValue = Value{typ: enum.SQLDataTypeNull, val: nil}

// NewNull creates a new NULL Value.
// NewNull creates a new NULL Value.
func NewNull() Value {
	// return nullValue // Return the singleton
	return Value{typ: enum.SQLDataTypeNull, val: nil}
}

// NewInt8 creates a new Int8 (TINYINT) Value.
// NewInt8 creates a new Int8 (TINYINT) Value.
func NewInt8(v int8) Value {
	return Value{typ: enum.SQLDataTypeInt8, val: v}
}

// NewInt16 creates a new Int16 (SMALLINT) Value.
// NewInt16 creates a new Int16 (SMALLINT) Value.
func NewInt16(v int16) Value {
	return Value{typ: enum.SQLDataTypeInt16, val: v}
}

// NewInt32 creates a new Int32 (INT) Value.
// NewInt32 creates a new Int32 (INT) Value.
func NewInt32(v int32) Value {
	return Value{typ: enum.SQLDataTypeInt32, val: v}
}

// NewInt64 creates a new Int64 (BIGINT) Value.
// NewInt64 creates a new Int64 (BIGINT) Value.
func NewInt64(v int64) Value {
	return Value{typ: enum.SQLDataTypeInt64, val: v}
}

// NewUint8 creates a new Uint8 (UNSIGNED TINYINT) Value.
// NewUint8 creates a new Uint8 (UNSIGNED TINYINT) Value.
func NewUint8(v uint8) Value {
	return Value{typ: enum.SQLDataTypeUint8, val: v}
}

// NewUint16 creates a new Uint16 (UNSIGNED SMALLINT) Value.
// NewUint16 creates a new Uint16 (UNSIGNED SMALLINT) Value.
func NewUint16(v uint16) Value {
	return Value{typ: enum.SQLDataTypeUint16, val: v}
}

// NewUint32 creates a new Uint32 (UNSIGNED INT) Value.
// NewUint32 creates a new Uint32 (UNSIGNED INT) Value.
func NewUint32(v uint32) Value {
	return Value{typ: enum.SQLDataTypeUint32, val: v}
}

// NewUint64 creates a new Uint64 (UNSIGNED BIGINT) Value.
// NewUint64 creates a new Uint64 (UNSIGNED BIGINT) Value.
func NewUint64(v uint64) Value {
	return Value{typ: enum.SQLDataTypeUint64, val: v}
}

// NewFloat32 creates a new Float32 (FLOAT) Value.
// NewFloat32 creates a new Float32 (FLOAT) Value.
func NewFloat32(v float32) Value {
	return Value{typ: enum.SQLDataTypeFloat32, val: v}
}

// NewFloat64 creates a new Float64 (DOUBLE) Value.
// NewFloat64 creates a new Float64 (DOUBLE) Value.
func NewFloat64(v float64) Value {
	return Value{typ: enum.SQLDataTypeFloat64, val: v}
}

// NewBoolean creates a new Boolean Value.
// NewBoolean creates a new Boolean Value.
func NewBoolean(v bool) Value {
	// Internally store as bool, can map to TINYINT(1) if needed for MySQL compat
	// Internally store as bool, can map to TINYINT(1) if needed for MySQL compat
	return Value{typ: enum.SQLDataTypeBoolean, val: v}
}

// NewString creates a new Varchar/Text Value.
// NewString creates a new Varchar/Text Value.
func NewString(v string) Value {
	// We might differentiate between VARCHAR, CHAR, TEXT later if needed
	// for specific padding or length constraints, but internally string is fine.
	// We might differentiate between VARCHAR, CHAR, TEXT later if needed
	// for specific padding or length constraints, but internally string is fine.
	return Value{typ: enum.SQLDataTypeVarchar, val: v}
}

// NewBlob creates a new Blob Value.
// NewBlob creates a new Blob Value.
func NewBlob(v []byte) Value {
	// Store a copy to prevent external modification
	// Store a copy to prevent external modification
	c := make([]byte, len(v))
	copy(c, v)
	return Value{typ: enum.SQLDataTypeBlob, val: c}
}

// NewDate creates a new Date Value.
// NewDate creates a new Date Value.
func NewDate(v time.Time) Value {
	// Store only year, month, day
	// Store only year, month, day
	year, month, day := v.Date()
	dateOnly := time.Date(year, month, day, 0, 0, 0, 0, time.UTC) // Store in UTC
	return Value{typ: enum.SQLDataTypeDate, val: dateOnly}
}

// NewTime creates a new Time Value.
// NewTime creates a new Time Value.
func NewTime(v time.Time) Value {
	// Store duration since midnight or a specific representation
	// Storing as time.Time for now, but only hour, min, sec, nsec matter.
	// TODO: Define precise internal representation for TIME type (e.g., nanos since midnight)
	// Storing as time.Time for now, but only hour, min, sec, nsec matter.
	// TODO: Define precise internal representation for TIME type (e.g., nanos since midnight)
	return Value{typ: enum.SQLDataTypeTime, val: v}
}

// NewTimestamp creates a new Timestamp Value.
// NewTimestamp creates a new Timestamp Value.
func NewTimestamp(v time.Time) Value {
	// Timestamps often depend on session timezone in MySQL, store as UTC internally
	// Timestamps often depend on session timezone in MySQL, store as UTC internally
	return Value{typ: enum.SQLDataTypeTimestamp, val: v.UTC()}
}

// NewDateTime creates a new DateTime Value.
// NewDateTime creates a new DateTime Value.
func NewDateTime(v time.Time) Value {
	// DATETIME is timezone-agnostic, store as is (but recommend UTC for consistency)
	// DATETIME is timezone-agnostic, store as is (but recommend UTC for consistency)
	return Value{typ: enum.SQLDataTypeDateTime, val: v}
}

// NewJSON creates a new JSON Value.
// NewJSON creates a new JSON Value.
func NewJSON(v []byte) Value {
	// Assume input is valid JSON bytes, store as []byte
	// Assume input is valid JSON bytes, store as []byte
	c := make([]byte, len(v))
	copy(c, v)
	return Value{typ: enum.SQLDataTypeJSON, val: c}
}

// TODO: Add constructors for Decimal, Enum, Set when their internal representations are defined.
// TODO: Add constructors for Decimal, Enum, Set when their internal representations are defined.

// Type returns the SQL data type of the Value.
// Type returns the SQL data type of the Value.
func (v Value) Type() enum.SQLDataType {
	return v.typ
}

// IsNull checks if the Value represents SQL NULL.
// IsNull checks if the Value represents SQL NULL.
func (v Value) IsNull() bool {
	return v.typ == enum.SQLDataTypeNull
}

// Get returns the underlying Go value. Use with caution.
// Get returns the underlying Go value. Use with caution.
func (v Value) Get() interface{} {
	return v.val
}

// Compare compares this Value with another Value.
// It returns:
// -1 if this value is less than other
//
//	0 if this value is equal to other
//
// +1 if this value is greater than other
// An error if the types are incompatible for comparison.
// SQL NULL comparison rules: NULL is considered less than any non-NULL value. NULL == NULL is false in SQL,
// but for sorting/indexing purposes, we treat NULLs as equal and smaller than others.
// Compare compares this Value with another Value.
// It returns:
// -1 if this value is less than other
//
//	0 if this value is equal to other
//
// +1 if this value is greater than other
// An error if the types are incompatible for comparison.
// SQL NULL comparison rules: NULL is considered less than any non-NULL value. NULL == NULL is false in SQL,
// but for sorting/indexing purposes, we treat NULLs as equal and smaller than others.
func (v Value) Compare(other Value) (int, error) {
	// Handle NULLs first: NULLs sort before non-NULLs
	// Handle NULLs first: NULLs sort before non-NULLs
	if v.IsNull() {
		if other.IsNull() {
			return 0, nil // NULL == NULL for sorting
		}
		return -1, nil // NULL < non-NULL
	}
	if other.IsNull() {
		return 1, nil // non-NULL > NULL
	}

	// TODO: Implement comprehensive comparison logic across compatible types.
	// This requires handling type promotions (e.g., INT vs FLOAT).
	// For now, implement basic same-type comparisons.
	// TODO: Implement comprehensive comparison logic across compatible types.
	// This requires handling type promotions (e.g., INT vs FLOAT).
	// For now, implement basic same-type comparisons.

	if v.typ != other.typ {
		// Attempt numeric comparison if both are numeric
		// Attempt numeric comparison if both are numeric
		if v.isNumeric() && other.isNumeric() {
			// Promote to float64 for comparison (simplistic approach)
			// Promote to float64 for comparison (simplistic approach)
			vf, errV := v.toFloat64()
			if errV != nil {
				return 0, errors.New(errors.ErrCodeComparison, fmt.Sprintf("cannot convert left value for comparison: %v", errV))
			}
			of, errO := other.toFloat64()
			if errO != nil {
				return 0, errors.New(errors.ErrCodeComparison, fmt.Sprintf("cannot convert right value for comparison: %v", errO))
			}
			if vf < of {
				return -1, nil
			}
			if vf > of {
				return 1, nil
			}
			return 0, nil
		}

		// Attempt string-like comparison
		// Attempt string-like comparison
		if v.isStringLike() && other.isStringLike() {
			vs, errV := v.ToString()
			if errV != nil {
				return 0, errors.New(errors.ErrCodeComparison, fmt.Sprintf("cannot convert left value to string for comparison: %v", errV))
			}
			os, errO := other.ToString()
			if errO != nil {
				return 0, errors.New(errors.ErrCodeComparison, fmt.Sprintf("cannot convert right value to string for comparison: %v", errO))
			}
			if vs < os {
				return -1, nil
			}
			if vs > os {
				return 1, nil
			}
			return 0, nil
		}

		// Add other compatible type comparisons (e.g., DATE vs TIMESTAMP) if needed
		// Add other compatible type comparisons (e.g., DATE vs TIMESTAMP) if needed

		return 0, errors.New(errors.ErrCodeComparison, fmt.Sprintf("incompatible types for comparison: %s vs %s", v.typ, other.typ))
	}

	// Same type comparison
	// Same type comparison
	switch v.typ {
	case enum.SQLDataTypeInt8:
		vv, _ := v.val.(int8)
		ov, _ := other.val.(int8)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeInt16:
		vv, _ := v.val.(int16)
		ov, _ := other.val.(int16)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeInt32:
		vv, _ := v.val.(int32)
		ov, _ := other.val.(int32)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeInt64:
		vv, _ := v.val.(int64)
		ov, _ := other.val.(int64)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeUint8:
		vv, _ := v.val.(uint8)
		ov, _ := other.val.(uint8)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeUint16:
		vv, _ := v.val.(uint16)
		ov, _ := other.val.(uint16)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeUint32:
		vv, _ := v.val.(uint32)
		ov, _ := other.val.(uint32)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeUint64:
		vv, _ := v.val.(uint64)
		ov, _ := other.val.(uint64)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeFloat32:
		vv, _ := v.val.(float32)
		ov, _ := other.val.(float32)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeFloat64:
		vv, _ := v.val.(float64)
		ov, _ := other.val.(float64)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeBoolean:
		vv, _ := v.val.(bool)
		ov, _ := other.val.(bool)
		// false < true
		// false < true
		if !vv && ov {
			return -1, nil
		}
		if vv && !ov {
			return 1, nil
		}
	case enum.SQLDataTypeVarchar, enum.SQLDataTypeChar, enum.SQLDataTypeText, enum.SQLDataTypeEnum, enum.SQLDataTypeSet:
		vv, _ := v.val.(string)
		ov, _ := other.val.(string)
		if vv < ov {
			return -1, nil
		}
		if vv > ov {
			return 1, nil
		}
	case enum.SQLDataTypeBlob:
		vv, _ := v.val.([]byte)
		ov, _ := other.val.([]byte)
		// Byte-wise comparison
		// Byte-wise comparison
		return compareBytes(vv, ov), nil
	case enum.SQLDataTypeDate, enum.SQLDataTypeTime, enum.SQLDataTypeTimestamp, enum.SQLDataTypeDateTime:
		vv, okV := v.val.(time.Time)
		ov, okO := other.val.(time.Time)
		if !okV || !okO {
			return 0, errors.New(errors.ErrCodeInternal, "internal error: time value is not time.Time")
		}
		if vv.Before(ov) {
			return -1, nil
		}
		if vv.After(ov) {
			return 1, nil
		}
	case enum.SQLDataTypeJSON:
		// JSON comparison is complex. For simple equality check, compare bytes.
		// For ordering, SQL typically doesn't define cross-JSON value order.
		// TODO: Define JSON comparison semantics if needed beyond equality.
		// JSON comparison is complex. For simple equality check, compare bytes.
		// For ordering, SQL typically doesn't define cross-JSON value order.
		// TODO: Define JSON comparison semantics if needed beyond equality.
		vv, _ := v.val.([]byte)
		ov, _ := other.val.([]byte)
		return compareBytes(vv, ov), nil
	// case enum.SQLDataTypeDecimal:
	// TODO: Implement Decimal comparison
	// TODO: Implement Decimal comparison
	default:
		return 0, errors.New(errors.ErrCodeInternal, fmt.Sprintf("comparison not implemented for type %s", v.typ))
	}

	return 0, nil // Values are equal
}

// ToString attempts to convert the Value to its string representation.
// ToString attempts to convert the Value to its string representation.
func (v Value) ToString() (string, error) {
	if v.IsNull() {
		return "NULL", nil // Or maybe "" depending on desired behavior? SQL usually shows NULL.
		// Or maybe "" depending on desired behavior? SQL usually shows NULL.
	}

	switch v.typ {
	case enum.SQLDataTypeVarchar, enum.SQLDataTypeChar, enum.SQLDataTypeText, enum.SQLDataTypeEnum, enum.SQLDataTypeSet:
		return v.val.(string), nil
	case enum.SQLDataTypeInt8:
		return strconv.FormatInt(int64(v.val.(int8)), 10), nil
	case enum.SQLDataTypeInt16:
		return strconv.FormatInt(int64(v.val.(int16)), 10), nil
	case enum.SQLDataTypeInt32:
		return strconv.FormatInt(int64(v.val.(int32)), 10), nil
	case enum.SQLDataTypeInt64:
		return strconv.FormatInt(v.val.(int64), 10), nil
	case enum.SQLDataTypeUint8:
		return strconv.FormatUint(uint64(v.val.(uint8)), 10), nil
	case enum.SQLDataTypeUint16:
		return strconv.FormatUint(uint64(v.val.(uint16)), 10), nil
	case enum.SQLDataTypeUint32:
		return strconv.FormatUint(uint64(v.val.(uint32)), 10), nil
	case enum.SQLDataTypeUint64:
		return strconv.FormatUint(v.val.(uint64), 10), nil
	case enum.SQLDataTypeFloat32:
		return strconv.FormatFloat(float64(v.val.(float32)), 'f', -1, 32), nil
	case enum.SQLDataTypeFloat64:
		return strconv.FormatFloat(v.val.(float64), 'f', -1, 64), nil
	case enum.SQLDataTypeBoolean:
		if v.val.(bool) {
			return "1", nil // Or "true"
		}
		return "0", nil // Or "false"
	case enum.SQLDataTypeBlob:
		// Represent blob as hex string? Or base64? Or just indicate <blob>?
		// Represent blob as hex string? Or base64? Or just indicate <blob>?
		return fmt.Sprintf("0x%X", v.val.([]byte)), nil // Hex representation
	case enum.SQLDataTypeDate:
		return v.val.(time.Time).Format("2006-01-02"), nil
	case enum.SQLDataTypeTime:
		// TODO: Format TIME correctly (HH:MM:SS.ffffff)
		// TODO: Format TIME correctly (HH:MM:SS.ffffff)
		return v.val.(time.Time).Format("15:04:05.999999"), nil
	case enum.SQLDataTypeTimestamp:
		// Format timestamp, potentially considering session timezone later
		// Format timestamp, potentially considering session timezone later
		return v.val.(time.Time).Format("2006-01-02 15:04:05.999999"), nil // Assumes UTC for now
	case enum.SQLDataTypeDateTime:
		return v.val.(time.Time).Format("2006-01-02 15:04:05.999999"), nil
	case enum.SQLDataTypeJSON:
		// Return JSON as string
		// Return JSON as string
		return string(v.val.([]byte)), nil
	// case enum.SQLDataTypeDecimal:
	// TODO: Implement Decimal string conversion
	// TODO: Implement Decimal string conversion
	default:
		return "", errors.New(errors.ErrCodeInternal, fmt.Sprintf("string conversion not implemented for type %s", v.typ))
	}
}

// ToInt64 attempts to convert the Value to an int64.
// ToInt64 attempts to convert the Value to an int64.
func (v Value) ToInt64() (int64, error) {
	if v.IsNull() {
		return 0, errors.New(errors.ErrCodeConversion, "cannot convert NULL to int64")
	}

	switch v.typ {
	case enum.SQLDataTypeInt8:
		return int64(v.val.(int8)), nil
	case enum.SQLDataTypeInt16:
		return int64(v.val.(int16)), nil
	case enum.SQLDataTypeInt32:
		return int64(v.val.(int32)), nil
	case enum.SQLDataTypeInt64:
		return v.val.(int64), nil
	case enum.SQLDataTypeUint8:
		// Check for overflow? For now, allow direct cast.
		// Check for overflow? For now, allow direct cast.
		return int64(v.val.(uint8)), nil
	case enum.SQLDataTypeUint16:
		return int64(v.val.(uint16)), nil
	case enum.SQLDataTypeUint32:
		// Check for overflow
		// Check for overflow
		uval := v.val.(uint32)
		if uval > uint32(^uint64(0)>>1) { // Check if > max int64
			return 0, errors.New(errors.ErrCodeConversion, fmt.Sprintf("uint32 value %d overflows int64", uval))
		}
		return int64(uval), nil
	case enum.SQLDataTypeUint64:
		// Check for overflow
		// Check for overflow
		uval := v.val.(uint64)
		if uval > uint64(^uint64(0)>>1) { // Check if > max int64
			return 0, errors.New(errors.ErrCodeConversion, fmt.Sprintf("uint64 value %d overflows int64", uval))
		}
		return int64(uval), nil
	case enum.SQLDataTypeFloat32:
		// Truncate decimal part
		// Truncate decimal part
		return int64(v.val.(float32)), nil
	case enum.SQLDataTypeFloat64:
		// Truncate decimal part
		// Truncate decimal part
		return int64(v.val.(float64)), nil
	case enum.SQLDataTypeBoolean:
		if v.val.(bool) {
			return 1, nil
		}
		return 0, nil
	case enum.SQLDataTypeVarchar, enum.SQLDataTypeChar, enum.SQLDataTypeText:
		// Attempt to parse string as integer
		// Attempt to parse string as integer
		i, err := strconv.ParseInt(v.val.(string), 10, 64)
		if err != nil {
			// Try parsing as float first for cases like "1.0"
			// Try parsing as float first for cases like "1.0"
			f, ferr := strconv.ParseFloat(v.val.(string), 64)
			if ferr == nil {
				return int64(f), nil
			}
			return 0, errors.New(errors.ErrCodeConversion, fmt.Sprintf("cannot convert string '%s' to int64: %v", v.val.(string), err))
		}
		return i, nil
	// case enum.SQLDataTypeDecimal:
	// TODO: Implement Decimal to Int64 conversion
	// TODO: Implement Decimal to Int64 conversion
	default:
		return 0, errors.New(errors.ErrCodeConversion, fmt.Sprintf("cannot convert type %s to int64", v.typ))
	}
}

// TODO: Implement other conversion methods as needed:
// ToUint64(), ToFloat64(), ToBool(), ToBytes(), ToTime(), ToDecimal() etc.
// TODO: Implement other conversion methods as needed:
// ToUint64(), ToFloat64(), ToBool(), ToBytes(), ToTime(), ToDecimal() etc.

// isNumeric checks if the value type is considered numeric.
// isNumeric checks if the value type is considered numeric.
func (v Value) isNumeric() bool {
	switch v.typ {
	case enum.SQLDataTypeInt8, enum.SQLDataTypeInt16, enum.SQLDataTypeInt32, enum.SQLDataTypeInt64,
		enum.SQLDataTypeUint8, enum.SQLDataTypeUint16, enum.SQLDataTypeUint32, enum.SQLDataTypeUint64,
		enum.SQLDataTypeFloat32, enum.SQLDataTypeFloat64, enum.SQLDataTypeDecimal, enum.SQLDataTypeBoolean: // Bool often treated as 0/1
		return true
	default:
		return false
	}
}

// isStringLike checks if the value type can be reasonably treated as a string.
// isStringLike checks if the value type can be reasonably treated as a string.
func (v Value) isStringLike() bool {
	switch v.typ {
	case enum.SQLDataTypeVarchar, enum.SQLDataTypeChar, enum.SQLDataTypeText, enum.SQLDataTypeEnum, enum.SQLDataTypeSet:
		return true
	default:
		return false
	}
}

// toFloat64 is a helper for comparisons, converting numeric types to float64.
// toFloat64 is a helper for comparisons, converting numeric types to float64.
func (v Value) toFloat64() (float64, error) {
	if v.IsNull() {
		// Should have been handled by caller, but return error for safety
		// Should have been handled by caller, but return error for safety
		return 0, errors.New(errors.ErrCodeConversion, "cannot convert NULL to float64")
	}

	switch v.typ {
	case enum.SQLDataTypeInt8:
		return float64(v.val.(int8)), nil
	case enum.SQLDataTypeInt16:
		return float64(v.val.(int16)), nil
	case enum.SQLDataTypeInt32:
		return float64(v.val.(int32)), nil
	case enum.SQLDataTypeInt64:
		return float64(v.val.(int64)), nil
	case enum.SQLDataTypeUint8:
		return float64(v.val.(uint8)), nil
	case enum.SQLDataTypeUint16:
		return float64(v.val.(uint16)), nil
	case enum.SQLDataTypeUint32:
		return float64(v.val.(uint32)), nil
	case enum.SQLDataTypeUint64:
		// Potential precision loss for very large uint64
		// Potential precision loss for very large uint64
		return float64(v.val.(uint64)), nil
	case enum.SQLDataTypeFloat32:
		return float64(v.val.(float32)), nil
	case enum.SQLDataTypeFloat64:
		return v.val.(float64), nil
	case enum.SQLDataTypeBoolean:
		if v.val.(bool) {
			return 1.0, nil
		}
		return 0.0, nil
	// case enum.SQLDataTypeDecimal:
	// TODO: Implement Decimal to Float64 conversion
	// TODO: Implement Decimal to Float64 conversion
	default:
		return 0, errors.New(errors.ErrCodeConversion, fmt.Sprintf("cannot convert type %s to float64", v.typ))
	}
}

// compareBytes performs lexicographical comparison of two byte slices.
// compareBytes performs lexicographical comparison of two byte slices.
func compareBytes(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// Uses gob encoding for simplicity. This is NOT suitable for sortable keys.
// MarshalBinary implements the encoding.BinaryMarshaler interface.
// Uses gob encoding for simplicity. This is NOT suitable for sortable keys.
func (v Value) MarshalBinary() ([]byte, error) {
	// Need to register concrete types used in interface{} with gob
	// This should ideally happen once at init time
	// Need to register concrete types used in interface{} with gob
	// This should ideally happen once at init time
	registerGobTypes()

	// Use a custom struct for gob encoding to handle the interface{} field
	// Use a custom struct for gob encoding to handle the interface{} field
	type gobValue struct {
		Typ enum.SQLDataType
		Val interface{}
	}
	gv := gobValue{Typ: v.typ, Val: v.val}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(gv); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeSerialization, "failed to gob encode Value")
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (v *Value) UnmarshalBinary(data []byte) error {
	registerGobTypes()

	type gobValue struct {
		Typ enum.SQLDataType
		Val interface{}
	}
	var gv gobValue

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&gv); err != nil {
		return errors.Wrap(err, errors.ErrCodeSerialization, "failed to gob decode Value")
	}

	v.typ = gv.Typ
	v.val = gv.Val
	return nil
}

// --- Gob Registration ---
// --- Gob Registration ---

var gobTypesRegistered = false

func registerGobTypes() {
	if gobTypesRegistered {
		return
	}
	// Register all concrete types that might be stored in Value.val interface{}
	// Register all concrete types that might be stored in Value.val interface{}
	gob.Register(int8(0))
	gob.Register(int16(0))
	gob.Register(int32(0))
	gob.Register(int64(0))
	gob.Register(uint8(0))
	gob.Register(uint16(0))
	gob.Register(uint32(0))
	gob.Register(uint64(0))
	gob.Register(float32(0))
	gob.Register(float64(0))
	gob.Register(false)
	gob.Register("")
	gob.Register([]byte{})
	gob.Register(time.Time{})
	// Register nil explicitly? Gob handles nil interfaces, but explicit registration might be safer.
	// gob.Register(nil)
	// TODO: Register Decimal, Enum, Set types when defined.
	// TODO: Register Decimal, Enum, Set types when defined.

	gobTypesRegistered = true
}

// String provides a human-readable representation of the Value, mainly for debugging.
// Use ToString() for SQL-like string conversion.
// String provides a human-readable representation of the Value, mainly for debugging.
// Use ToString() for SQL-like string conversion.
func (v Value) String() string {
	if v.IsNull() {
		return "NULL"
	}
	// Use ToString for consistency in simple cases, fallback for others
	// Use ToString for consistency in simple cases, fallback for others
	str, err := v.ToString()
	if err == nil {
		// Quote string-like types for clarity in debugging
		// Quote string-like types for clarity in debugging
		if v.isStringLike() || v.typ == enum.SQLDataTypeJSON {
			return strconv.Quote(str)
		}
		if v.typ == enum.SQLDataTypeBlob {
			// Use the hex representation from ToString
			// Use the hex representation from ToString
			return str
		}
		return str
	}
	// Fallback for types without simple ToString or on error
	// Fallback for types without simple ToString or on error
	return fmt.Sprintf("%s(%v)", v.typ, v.val)
}

// Note: EncodeSortable and DecodeSortable methods are intentionally omitted here.
// They belong in the storage encoding layer (e.g., storage/engines/badger/encoding.go)
// because the sortable format is highly specific to the storage requirements (like Badger keys)
// and often needs context (like whether it's ascending/descending, part of a composite key).
// Keeping Value itself focused on representing the typed value and basic operations.
// Note: EncodeSortable and DecodeSortable methods are intentionally omitted here.
// They belong in the storage encoding layer (e.g., storage/engines/badger/encoding.go)
// because the sortable format is highly specific to the storage requirements (like Badger keys)
// and often needs context (like whether it's ascending/descending, part of a composite key).
// Keeping Value itself focused on representing the typed value and basic operations.
