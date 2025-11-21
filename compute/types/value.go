// Copyright 2024 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// Value is a SQL value.
type Value struct {
	typ  Type
	data interface{}
}

// NewValue creates a new Value.
func NewValue(typ Type, data interface{}) (*Value, error) {
	if data == nil {
		return &Value{typ: Null, data: nil}, nil
	}
	converted, err := typ.Convert(data)
	if err != nil {
		return nil, err
	}
	return &Value{typ: typ, data: converted}, nil
}

// IsNull returns true if the value is null.
func (v *Value) IsNull() bool {
	return v.data == nil
}

// Compare compares the value with another value.
func (v *Value) Compare(other *Value) (int, error) {
	return v.typ.Compare(v.data, other.data)
}

// ToBytes serializes the value to a byte slice.
func (v *Value) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v.typ.QueryType()); err != nil {
		return nil, err
	}
	if err := enc.Encode(&v.data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBytes deserializes the value from a byte slice.
func (v *Value) FromBytes(b []byte) error {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var qt QueryType
	if err := dec.Decode(&qt); err != nil {
		return err
	}
	typ, err := MysqlTypeToType(qt)
	if err != nil {
		return err
	}
	v.typ = typ
	return dec.Decode(&v.data)
}

// AsInt64 returns the value as an int64.
func (v *Value) AsInt64() (int64, error) {
	if v.IsNull() {
		return 0, nil
	}
	val, ok := v.data.(int64)
	if !ok {
		return 0, fmt.Errorf("value is not an int64")
	}
	return val, nil
}

// AsString returns the value as a string.
func (v *Value) AsString() (string, error) {
	if v.IsNull() {
		return "", nil
	}
	val, ok := v.data.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	return val, nil
}
