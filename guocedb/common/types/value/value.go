package value

import (
	"fmt"
	"time"
)

// Type represents a SQL data type.
type Type int

const (
	Null Type = iota
	Int64
	Float64
	String
	Timestamp
	// Other types can be added here
)

// Value is an interface for all SQL value types.
type Value interface {
	GetType() Type
	IsNull() bool
	Compare(other Value) (int, error)
	ToString() string
	// Serialize and Deserialize methods would go here
}

// NullValue represents a NULL value.
type NullValue struct{}

// NewNull creates a new NullValue.
func NewNull() Value {
	return NullValue{}
}

func (v NullValue) GetType() Type      { return Null }
func (v NullValue) IsNull() bool       { return true }
func (v NullValue) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 0, nil
	}
	return -1, nil
}
func (v NullValue) ToString() string { return "NULL" }

// Int64Value represents an INT64 value.
type Int64Value struct {
	Val int64
}

// NewInt64 creates a new Int64Value.
func NewInt64(val int64) Value {
	return Int64Value{Val: val}
}

func (v Int64Value) GetType() Type      { return Int64 }
func (v Int64Value) IsNull() bool       { return false }
func (v Int64Value) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherVal, ok := other.(Int64Value)
	if !ok {
		return 0, fmt.Errorf("cannot compare Int64 with %T", other)
	}
	if v.Val < otherVal.Val {
		return -1, nil
	}
	if v.Val > otherVal.Val {
		return 1, nil
	}
	return 0, nil
}
func (v Int64Value) ToString() string { return fmt.Sprintf("%d", v.Val) }

// StringValue represents a STRING/VARCHAR value.
type StringValue struct {
	Val string
}

// NewString creates a new StringValue.
func NewString(val string) Value {
	return StringValue{Val: val}
}

func (v StringValue) GetType() Type      { return String }
func (v StringValue) IsNull() bool       { return false }
func (v StringValue) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherVal, ok := other.(StringValue)
	if !ok {
		return 0, fmt.Errorf("cannot compare String with %T", other)
	}
	if v.Val < otherVal.Val {
		return -1, nil
	}
	if v.Val > otherVal.Val {
		return 1, nil
	}
	return 0, nil
}
func (v StringValue) ToString() string { return v.Val }

// TimestampValue represents a TIMESTAMP value.
type TimestampValue struct {
	Val time.Time
}

// NewTimestamp creates a new TimestampValue.
func NewTimestamp(val time.Time) Value {
	return TimestampValue{Val: val}
}

func (v TimestampValue) GetType() Type      { return Timestamp }
func (v TimestampValue) IsNull() bool       { return false }
func (v TimestampValue) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherVal, ok := other.(TimestampValue)
	if !ok {
		return 0, fmt.Errorf("cannot compare Timestamp with %T", other)
	}
	if v.Val.Before(otherVal.Val) {
		return -1, nil
	}
	if v.Val.After(otherVal.Val) {
		return 1, nil
	}
	return 0, nil
}
func (v TimestampValue) ToString() string { return v.Val.Format(time.RFC3339Nano) }
