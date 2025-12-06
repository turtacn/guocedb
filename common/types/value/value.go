// Package value defines the SQL value types used in guocedb.
package value

import (
	"fmt"
	"time"

	"github.com/turtacn/guocedb/compute/sql"
)

// Value is the interface that all value types must implement.
type Value interface {
	// Type returns the underlying SQL type.
	Type() sql.Type
	// IsNull returns true if the value is NULL.
	IsNull() bool
	// Compare compares this value with another value.
	Compare(other Value) (int, error)
	// ToString returns the string representation of the value.
	ToString() string
	// ToBytes returns the byte representation for serialization.
	ToBytes() ([]byte, error)
	// FromBytes populates the value from a byte slice.
	FromBytes(data []byte) error
}

// BaseValue provides a base implementation for common Value methods.
type BaseValue struct{}

// IsNull returns false by default.
func (v *BaseValue) IsNull() bool {
	return false
}

// Null represents a NULL value.
type Null struct {
	BaseValue
	sqlType sql.Type
}

// NewNull creates a new Null value for a given SQL type.
func NewNull(sqlType sql.Type) *Null {
	return &Null{sqlType: sqlType}
}

// Type returns the underlying SQL type.
func (n *Null) Type() sql.Type {
	return n.sqlType
}

// IsNull returns true for Null values.
func (n *Null) IsNull() bool {
	return true
}

// Compare compares this null value with another value.
func (n *Null) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 0, nil // NULL is equal to NULL
	}
	return -1, nil // NULL is considered smaller than any non-NULL value
}

// ToString returns "NULL".
func (n *Null) ToString() string {
	return "NULL"
}

// ToBytes returns an empty byte slice for NULL.
func (n *Null) ToBytes() ([]byte, error) {
	return []byte{}, nil
}

// FromBytes does nothing for NULL.
func (n *Null) FromBytes(data []byte) error {
	return nil
}

// Int64 represents a 64-bit integer value.
type Int64 struct {
	BaseValue
	Value int64
}

// Type returns the BIGINT SQL type.
func (i *Int64) Type() sql.Type {
	return sql.Int64
}

// ToString returns the string representation of the integer.
func (i *Int64) ToString() string {
	return fmt.Sprintf("%d", i.Value)
}

// Compare compares this int64 with another value.
func (i *Int64) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherInt, ok := other.(*Int64)
	if !ok {
		return 0, fmt.Errorf("cannot compare Int64 with %T", other)
	}
	if i.Value < otherInt.Value {
		return -1, nil
	}
if i.Value > otherInt.Value {
		return 1, nil
	}
	return 0, nil
}

// ToBytes is a placeholder for serialization.
func (i *Int64) ToBytes() ([]byte, error) {
	// Implementation would go here.
	return []byte(i.ToString()), nil
}

// FromBytes is a placeholder for deserialization.
func (i *Int64) FromBytes(data []byte) error {
	// Implementation would go here.
	_, err := fmt.Sscanf(string(data), "%d", &i.Value)
	return err
}


// String represents a VARCHAR value.
type String struct {
	BaseValue
	Value string
}

// Type returns the VARCHAR SQL type.
func (s *String) Type() sql.Type {
	return sql.Text
}

// ToString returns the string itself.
func (s *String) ToString() string {
	return s.Value
}

// Compare compares this string with another value.
func (s *String) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherStr, ok := other.(*String)
	if !ok {
		return 0, fmt.Errorf("cannot compare String with %T", other)
	}
	if s.Value < otherStr.Value {
		return -1, nil
	}
	if s.Value > otherStr.Value {
		return 1, nil
	}
	return 0, nil
}

// ToBytes returns the byte representation of the string.
func (s *String) ToBytes() ([]byte, error) {
	return []byte(s.Value), nil
}

// FromBytes populates the string from a byte slice.
func (s *String) FromBytes(data []byte) error {
	s.Value = string(data)
	return nil
}

// Timestamp represents a TIMESTAMP value.
type Timestamp struct {
	BaseValue
	Value time.Time
}

// Type returns the TIMESTAMP SQL type.
func (t *Timestamp) Type() sql.Type {
	return sql.Timestamp
}

// ToString returns the formatted timestamp string.
func (t *Timestamp) ToString() string {
	return t.Value.Format(time.RFC3339)
}

// Compare compares this timestamp with another value.
func (t *Timestamp) Compare(other Value) (int, error) {
	if other.IsNull() {
		return 1, nil
	}
	otherTs, ok := other.(*Timestamp)
	if !ok {
		return 0, fmt.Errorf("cannot compare Timestamp with %T", other)
	}
	if t.Value.Before(otherTs.Value) {
		return -1, nil
	}
	if t.Value.After(otherTs.Value) {
		return 1, nil
	}
	return 0, nil
}

// ToBytes serializes the timestamp.
func (t *Timestamp) ToBytes() ([]byte, error) {
	return t.Value.MarshalBinary()
}

// FromBytes deserializes the timestamp.
func (t *Timestamp) FromBytes(data []byte) error {
	return t.Value.UnmarshalBinary(data)
}
