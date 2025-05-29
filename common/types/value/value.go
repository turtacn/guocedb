// Package value defines the base data types for representing SQL values (e.g., integers, strings, dates, booleans)
// within the Guocedb project. These types encapsulate Go native types and provide necessary type conversion,
// comparison, and serialization/deserialization methods. This file is a core dependency for the compute layer
// (parser, analyzer, executor, plan) and the storage layer (badger/encoding.go), ensuring consistent data
// representation and manipulation internally.
package value

import (
	"bytes"           // For byte buffer operations, especially for serialization.
	"encoding/binary" // For binary encoding of primitive types.
	"fmt"             // For string formatting, especially for error messages.
	"math"            // For numeric type range checks and conversions.
	"strconv"         // For string conversions to/from numeric types.
	"time"            // For handling date and time types.

	"guocedb/common/errors"     // For unified error handling.
	"guocedb/common/types/enum" // For SQL type enumerations.
)

// Value is the interface that all SQL value types must implement.
// It provides methods for type checking, comparison, and serialization.
type Value interface {
	Type() enum.SQLType               // Type returns the SQL type of the value.
	String() string                   // String returns the string representation of the value.
	Bytes() ([]byte, error)           // Bytes returns the binary representation of the value for storage/network.
	Compare(other Value) (int, error) // Compare compares this value with another. Returns -1, 0, or 1.
	IsNil() bool                      // IsNil returns true if the value is SQL NULL.
	// Add other common methods like Equal, Add, Subtract etc. as needed for operations.
}

// typedValue is a base struct for all concrete Value implementations.
type typedValue struct {
	sqlType enum.SQLType // The SQL type corresponding to this value.
	isNull  bool         // Indicates if the value is SQL NULL.
}

// Type implements the Value interface for typedValue.
func (tv *typedValue) Type() enum.SQLType {
	return tv.sqlType
}

// IsNil implements the Value interface for typedValue.
func (tv *typedValue) IsNil() bool {
	return tv.isNull
}

// NewNullValue creates a new SQL NULL value of a specific type.
func NewNullValue(sqlType enum.SQLType) Value {
	return &typedValue{
		sqlType: sqlType,
		isNull:  true,
	}
}

// -----------------------------------------------------------------------------
// Boolean Type
// -----------------------------------------------------------------------------

// Boolean represents a SQL BOOLEAN value.
type Boolean struct {
	typedValue
	Val bool
}

// NewBoolean creates a new Boolean value.
func NewBoolean(b bool) *Boolean {
	return &Boolean{typedValue: typedValue{sqlType: enum.SQLTypeBoolean}, Val: b}
}

// NewBooleanFromNil creates a new SQL NULL Boolean value.
func NewBooleanFromNil() *Boolean {
	return &Boolean{typedValue: typedValue{sqlType: enum.SQLTypeBoolean, isNull: true}}
}

// String implements the Value interface for Boolean.
func (b *Boolean) String() string {
	if b.IsNil() {
		return "NULL"
	}
	return strconv.FormatBool(b.Val)
}

// Bytes implements the Value interface for Boolean.
func (b *Boolean) Bytes() ([]byte, error) {
	if b.IsNil() {
		// Represent NULL boolean as a single byte: 0xFF
		return []byte{0xFF}, nil
	}
	if b.Val {
		return []byte{0x01}, nil
	}
	return []byte{0x00}, nil
}

// Compare implements the Value interface for Boolean.
func (b *Boolean) Compare(other Value) (int, error) {
	if b.IsNil() || other.IsNil() {
		// NULLs are handled based on SQL semantics, typically unordered.
		// For comparison, we might consider NULL to be equal to NULL,
		// and NULL to be less than/greater than non-NULL based on specific SQL dialect.
		// For now, return error or handle as per strictness. Here, we define NULLs as "equal" for sorting purposes,
		// but actual SQL comparison rules for NULLs are often more nuanced.
		if b.IsNil() && other.IsNil() {
			return 0, nil
		}
		// Non-null is always greater than null (arbitrary, for consistent sorting)
		if b.IsNil() {
			return -1, nil // This value is NULL, other is not.
		}
		return 1, nil // This value is NOT NULL, other is NULL.
	}

	otherBoolean, ok := other.(*Boolean)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare Boolean with %s", other.Type().String()), nil)
	}

	if b.Val == otherBoolean.Val {
		return 0, nil
	} else if b.Val { // true > false
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// Integer Type
// -----------------------------------------------------------------------------

// Integer represents a SQL INTEGER (INT64) value.
type Integer struct {
	typedValue
	Val int64
}

// NewInteger creates a new Integer value.
func NewInteger(i int64) *Integer {
	return &Integer{typedValue: typedValue{sqlType: enum.SQLTypeInteger}, Val: i}
}

// NewIntegerFromNil creates a new SQL NULL Integer value.
func NewIntegerFromNil() *Integer {
	return &Integer{typedValue: typedValue{sqlType: enum.SQLTypeInteger, isNull: true}}
}

// String implements the Value interface for Integer.
func (i *Integer) String() string {
	if i.IsNil() {
		return "NULL"
	}
	return strconv.FormatInt(i.Val, 10)
}

// Bytes implements the Value interface for Integer.
func (i *Integer) Bytes() ([]byte, error) {
	if i.IsNil() {
		// Represent NULL integer as a special byte sequence (e.g., specific marker + type)
		// Or a simple common NULL marker byte. For simplicity, let's use a convention:
		// For integers, a single byte 0xFE could mean NULL to differentiate from boolean 0xFF.
		// A more robust approach might be to encode a type byte + NULL marker.
		// Let's use a convention from the encoding package: prefix with 0x00 for NULL, 0x01 for non-NULL
		var buf bytes.Buffer
		buf.WriteByte(0x00)                      // NULL marker
		buf.WriteByte(byte(enum.SQLTypeInteger)) // Type indicator
		return buf.Bytes(), nil
	}
	buf := new(bytes.Buffer)
	// Write a non-NULL marker
	buf.WriteByte(0x01)
	if err := binary.Write(buf, binary.BigEndian, i.Val); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to serialize integer", err)
	}
	return buf.Bytes(), nil
}

// Compare implements the Value interface for Integer.
func (i *Integer) Compare(other Value) (int, error) {
	if i.IsNil() || other.IsNil() {
		if i.IsNil() && other.IsNil() {
			return 0, nil
		}
		if i.IsNil() {
			return -1, nil
		}
		return 1, nil
	}

	otherInteger, ok := other.(*Integer)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare Integer with %s", other.Type().String()), nil)
	}

	if i.Val == otherInteger.Val {
		return 0, nil
	} else if i.Val > otherInteger.Val {
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// Float Type
// -----------------------------------------------------------------------------

// Float represents a SQL FLOAT (FLOAT64) value.
type Float struct {
	typedValue
	Val float64
}

// NewFloat creates a new Float value.
func NewFloat(f float64) *Float {
	return &Float{typedValue: typedValue{sqlType: enum.SQLTypeFloat}, Val: f}
}

// NewFloatFromNil creates a new SQL NULL Float value.
func NewFloatFromNil() *Float {
	return &Float{typedValue: typedValue{sqlType: enum.SQLTypeFloat, isNull: true}}
}

// String implements the Value interface for Float.
func (f *Float) String() string {
	if f.IsNil() {
		return "NULL"
	}
	return strconv.FormatFloat(f.Val, 'f', -1, 64)
}

// Bytes implements the Value interface for Float.
func (f *Float) Bytes() ([]byte, error) {
	if f.IsNil() {
		var buf bytes.Buffer
		buf.WriteByte(0x00)                    // NULL marker
		buf.WriteByte(byte(enum.SQLTypeFloat)) // Type indicator
		return buf.Bytes(), nil
	}
	buf := new(bytes.Buffer)
	buf.WriteByte(0x01) // Non-NULL marker
	if err := binary.Write(buf, binary.BigEndian, f.Val); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to serialize float", err)
	}
	return buf.Bytes(), nil
}

// Compare implements the Value interface for Float.
func (f *Float) Compare(other Value) (int, error) {
	if f.IsNil() || other.IsNil() {
		if f.IsNil() && other.IsNil() {
			return 0, nil
		}
		if f.IsNil() {
			return -1, nil
		}
		return 1, nil
	}

	otherFloat, ok := other.(*Float)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare Float with %s", other.Type().String()), nil)
	}

	if f.Val == otherFloat.Val {
		return 0, nil
	} else if f.Val > otherFloat.Val {
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// String Type
// -----------------------------------------------------------------------------

// String represents a SQL VARCHAR/TEXT value.
type String struct {
	typedValue
	Val string
}

// NewString creates a new String value.
func NewString(s string) *String {
	return &String{typedValue: typedValue{sqlType: enum.SQLTypeString}, Val: s}
}

// NewStringFromNil creates a new SQL NULL String value.
func NewStringFromNil() *String {
	return &String{typedValue: typedValue{sqlType: enum.SQLTypeString, isNull: true}}
}

// String implements the Value interface for String.
func (s *String) String() string {
	if s.IsNil() {
		return "NULL"
	}
	return "'" + s.Val + "'" // Quote string for SQL-like representation
}

// Bytes implements the Value interface for String.
func (s *String) Bytes() ([]byte, error) {
	if s.IsNil() {
		var buf bytes.Buffer
		buf.WriteByte(0x00)                     // NULL marker
		buf.WriteByte(byte(enum.SQLTypeString)) // Type indicator
		return buf.Bytes(), nil
	}
	var buf bytes.Buffer
	buf.WriteByte(0x01) // Non-NULL marker
	// For strings, we can simply write the length prefix then the bytes.
	// Using binary.Write for length for consistency.
	lenBytes := make([]byte, 4) // Assuming max string length fits in int32
	binary.BigEndian.PutUint32(lenBytes, uint32(len(s.Val)))
	buf.Write(lenBytes)
	buf.WriteString(s.Val)
	return buf.Bytes(), nil
}

// Compare implements the Value interface for String.
func (s *String) Compare(other Value) (int, error) {
	if s.IsNil() || other.IsNil() {
		if s.IsNil() && other.IsNil() {
			return 0, nil
		}
		if s.IsNil() {
			return -1, nil
		}
		return 1, nil
	}

	otherString, ok := other.(*String)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare String with %s", other.Type().String()), nil)
	}

	if s.Val == otherString.Val {
		return 0, nil
	} else if s.Val > otherString.Val {
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// Date Type
// -----------------------------------------------------------------------------

// Date represents a SQL DATE value (only date part).
type Date struct {
	typedValue
	Val time.Time // Stored as UTC time.Time, only YYYY-MM-DD matters.
}

// NewDate creates a new Date value.
func NewDate(t time.Time) *Date {
	return &Date{typedValue: typedValue{sqlType: enum.SQLTypeDate}, Val: t.UTC().Truncate(24 * time.Hour)}
}

// NewDateFromNil creates a new SQL NULL Date value.
func NewDateFromNil() *Date {
	return &Date{typedValue: typedValue{sqlType: enum.SQLTypeDate, isNull: true}}
}

// String implements the Value interface for Date.
func (d *Date) String() string {
	if d.IsNil() {
		return "NULL"
	}
	return d.Val.Format("2006-01-02")
}

// Bytes implements the Value interface for Date.
func (d *Date) Bytes() ([]byte, error) {
	if d.IsNil() {
		var buf bytes.Buffer
		buf.WriteByte(0x00)                   // NULL marker
		buf.WriteByte(byte(enum.SQLTypeDate)) // Type indicator
		return buf.Bytes(), nil
	}
	var buf bytes.Buffer
	buf.WriteByte(0x01) // Non-NULL marker
	// Store date as YYYYMMDD integer or similar compact representation.
	// For simplicity, store as string bytes for now.
	dateStr := d.Val.Format("2006-01-02")
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(dateStr)))
	buf.Write(lenBytes)
	buf.WriteString(dateStr)
	return buf.Bytes(), nil
}

// Compare implements the Value interface for Date.
func (d *Date) Compare(other Value) (int, error) {
	if d.IsNil() || other.IsNil() {
		if d.IsNil() && other.IsNil() {
			return 0, nil
		}
		if d.IsNil() {
			return -1, nil
		}
		return 1, nil
	}

	otherDate, ok := other.(*Date)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare Date with %s", other.Type().String()), nil)
	}

	if d.Val.Equal(otherDate.Val) {
		return 0, nil
	} else if d.Val.After(otherDate.Val) {
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// Datetime Type
// -----------------------------------------------------------------------------

// Datetime represents a SQL DATETIME value.
type Datetime struct {
	typedValue
	Val time.Time // Stored as UTC time.Time.
}

// NewDatetime creates a new Datetime value.
func NewDatetime(t time.Time) *Datetime {
	return &Datetime{typedValue: typedValue{sqlType: enum.SQLTypeDatetime}, Val: t.UTC()}
}

// NewDatetimeFromNil creates a new SQL NULL Datetime value.
func NewDatetimeFromNil() *Datetime {
	return &Datetime{typedValue: typedValue{sqlType: enum.SQLTypeDatetime, isNull: true}}
}

// String implements the Value interface for Datetime.
func (dt *Datetime) String() string {
	if dt.IsNil() {
		return "NULL"
	}
	return dt.Val.Format("2006-01-02 15:04:05")
}

// Bytes implements the Value interface for Datetime.
func (dt *Datetime) Bytes() ([]byte, error) {
	if dt.IsNil() {
		var buf bytes.Buffer
		buf.WriteByte(0x00)                       // NULL marker
		buf.WriteByte(byte(enum.SQLTypeDatetime)) // Type indicator
		return buf.Bytes(), nil
	}
	var buf bytes.Buffer
	buf.WriteByte(0x01) // Non-NULL marker
	// Store time.Time as its UnixNano representation for precise timestamp.
	if err := binary.Write(buf, binary.BigEndian, dt.Val.UnixNano()); err != nil {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to serialize datetime", err)
	}
	return buf.Bytes(), nil
}

// Compare implements the Value interface for Datetime.
func (dt *Datetime) Compare(other Value) (int, error) {
	if dt.IsNil() || other.IsNil() {
		if dt.IsNil() && other.IsNil() {
			return 0, nil
		}
		if dt.IsNil() {
			return -1, nil
		}
		return 1, nil
	}

	otherDatetime, ok := other.(*Datetime)
	if !ok {
		return 0, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput,
			fmt.Sprintf("cannot compare Datetime with %s", other.Type().String()), nil)
	}

	if dt.Val.Equal(otherDatetime.Val) {
		return 0, nil
	} else if dt.Val.After(otherDatetime.Val) {
		return 1, nil
	}
	return -1, nil
}

// -----------------------------------------------------------------------------
// Helper functions for type conversions and deserialization from bytes
// (These would typically reside in an encoding or deserialization package)
// -----------------------------------------------------------------------------

// ValueFromBytes reconstructs a Value from its byte representation.
// This function needs to know the expected SQLType to correctly deserialize.
// This is a placeholder; a more robust solution would embed type information in bytes.
func ValueFromBytes(data []byte, expectedType enum.SQLType) (Value, error) {
	if len(data) == 0 {
		return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeInvalidInput, "empty byte slice for value deserialization", nil)
	}

	// Check for NULL marker (our convention: 0x00 for NULL, 0x01 for non-NULL)
	isNonNull := data[0] == 0x01
	if !isNonNull {
		// Expect type byte to follow NULL marker
		if len(data) < 2 {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeInvalidInput, "malformed NULL value bytes: missing type indicator", nil)
		}
		// TODO: Could also check if data[1] matches expectedType
		return NewNullValue(expectedType), nil
	}

	// For non-NULL values, skip the non-NULL marker and deserialize
	data = data[1:] // Advance past the non-NULL marker

	buf := bytes.NewReader(data)

	switch expectedType {
	case enum.SQLTypeBoolean:
		var b byte
		if err := binary.Read(buf, binary.BigEndian, &b); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize boolean", err)
		}
		return NewBoolean(b == 0x01), nil
	case enum.SQLTypeInteger:
		var i int64
		if err := binary.Read(buf, binary.BigEndian, &i); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize integer", err)
		}
		return NewInteger(i), nil
	case enum.SQLTypeFloat:
		var f float64
		if err := binary.Read(buf, binary.BigEndian, &f); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize float", err)
		}
		return NewFloat(f), nil
	case enum.SQLTypeString:
		var strLen uint32
		if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize string length", err)
		}
		strBytes := make([]byte, strLen)
		if _, err := buf.Read(strBytes); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize string bytes", err)
		}
		return NewString(string(strBytes)), nil
	case enum.SQLTypeDate:
		var strLen uint32
		if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize date string length", err)
		}
		dateStrBytes := make([]byte, strLen)
		if _, err := buf.Read(dateStrBytes); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize date string bytes", err)
		}
		t, err := time.Parse("2006-01-02", string(dateStrBytes))
		if err != nil {
			return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, "failed to parse date string", err)
		}
		return NewDate(t), nil
	case enum.SQLTypeDatetime:
		var unixNano int64
		if err := binary.Read(buf, binary.BigEndian, &unixNano); err != nil {
			return nil, errors.NewGuocedbError(enum.ErrStorage, errors.CodeStorageIOError, "failed to deserialize datetime unix nano", err)
		}
		return NewDatetime(time.Unix(0, unixNano).UTC()), nil
	default:
		return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported,
			fmt.Sprintf("deserialization for SQL type %s not yet supported", expectedType.String()), nil)
	}
}

// ConvertValue converts a Value to another SQLType if possible.
// This is a simplified example and needs to handle all possible type conversions.
func ConvertValue(val Value, targetType enum.SQLType) (Value, error) {
	if val.Type() == targetType {
		return val, nil // No conversion needed.
	}
	if val.IsNil() {
		return NewNullValue(targetType), nil // Null values can be converted to other null types.
	}

	switch targetType {
	case enum.SQLTypeBoolean:
		switch v := val.(type) {
		case *Integer:
			return NewBoolean(v.Val != 0), nil
		case *Float:
			return NewBoolean(v.Val != 0.0), nil
		case *String:
			b, err := strconv.ParseBool(v.Val)
			if err != nil {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("cannot convert string '%s' to boolean", v.Val), err)
			}
			return NewBoolean(b), nil
		default:
			return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("cannot convert %s to Boolean", val.Type().String()), nil)
		}
	case enum.SQLTypeInteger:
		switch v := val.(type) {
		case *Boolean:
			if v.Val {
				return NewInteger(1), nil
			} else {
				return NewInteger(0), nil
			}
		case *Float:
			if v.Val > float64(math.MaxInt64) || v.Val < float64(math.MinInt64) {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("float '%f' out of integer range", v.Val), nil)
			}
			return NewInteger(int64(v.Val)), nil
		case *String:
			i, err := strconv.ParseInt(v.Val, 10, 64)
			if err != nil {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("cannot convert string '%s' to integer", v.Val), err)
			}
			return NewInteger(i), nil
		default:
			return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("cannot convert %s to Integer", val.Type().String()), nil)
		}
	case enum.SQLTypeFloat:
		switch v := val.(type) {
		case *Boolean:
			if v.Val {
				return NewFloat(1.0), nil
			} else {
				return NewFloat(0.0), nil
			}
		case *Integer:
			return NewFloat(float64(v.Val)), nil
		case *String:
			f, err := strconv.ParseFloat(v.Val, 64)
			if err != nil {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("cannot convert string '%s' to float", v.Val), err)
			}
			return NewFloat(f), nil
		default:
			return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("cannot convert %s to Float", val.Type().String()), nil)
		}
	case enum.SQLTypeString:
		return NewString(val.String()), nil // All types can be converted to string representation.
	case enum.SQLTypeDate:
		switch v := val.(type) {
		case *String:
			t, err := time.Parse("2006-01-02", v.Val)
			if err != nil {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("cannot convert string '%s' to date (expected YYYY-MM-DD)", v.Val), err)
			}
			return NewDate(t), nil
		case *Datetime:
			return NewDate(v.Val), nil // Extract date part from datetime
		default:
			return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("cannot convert %s to Date", val.Type().String()), nil)
		}
	case enum.SQLTypeDatetime:
		switch v := val.(type) {
		case *String:
			t, err := time.Parse("2006-01-02 15:04:05", v.Val)
			if err != nil {
				return nil, errors.NewGuocedbError(enum.ErrInvalidArgument, errors.CodeInvalidInput, fmt.Sprintf("cannot convert string '%s' to datetime (expected YYYY-MM-DD HH:MM:SS)", v.Val), err)
			}
			return NewDatetime(t), nil
		case *Date:
			// Convert date to datetime, time part will be 00:00:00
			return NewDatetime(v.Val), nil
		default:
			return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("cannot convert %s to Datetime", val.Type().String()), nil)
		}
	default:
		return nil, errors.NewGuocedbError(enum.ErrNotSupported, errors.CodeFeatureNotSupported, fmt.Sprintf("conversion to SQL type %s not yet supported", targetType.String()), nil)
	}
}

//Personal.AI order the ending
