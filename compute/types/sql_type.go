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
	"fmt"
	"time"
)

var (
	ErrNullComparison = fmt.Errorf("NULL values are not comparable")
)

// Type is the interface for all SQL types.
type Type interface {
	// QueryType returns the query type.
	QueryType() QueryType
	// SQL returns the SQL representation of the type.
	SQL() string
	// Compare compares two values of the same type.
	Compare(a interface{}, b interface{}) (int, error)
	// Convert converts a value to the type.
	Convert(v interface{}) (interface{}, error)
	// Zero returns the zero value for the type.
	Zero() interface{}
}

// baseType is a base implementation of the Type interface.
type baseType struct {
	typ QueryType
}

// QueryType implements the Type interface.
func (bt *baseType) QueryType() QueryType {
	return bt.typ
}

// int64Type is the implementation of the INT64 type.
type int64Type struct {
	baseType
}

// SQL implements the Type interface.
func (t *int64Type) SQL() string {
	return "BIGINT"
}

// Compare implements the Type interface.
func (t *int64Type) Compare(a interface{}, b interface{}) (int, error) {
	if a == nil || b == nil {
		return 0, ErrNullComparison
	}

	aVal, err := ConvertToInt64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := ConvertToInt64(b)
	if err != nil {
		return 0, err
	}

	if aVal < bVal {
		return -1, nil
	}
	if aVal > bVal {
		return 1, nil
	}
	return 0, nil
}

// Convert implements the Type interface.
func (t *int64Type) Convert(v interface{}) (interface{}, error) {
	return ConvertToInt64(v)
}

// Zero implements the Type interface.
func (t *int64Type) Zero() interface{} {
	return int64(0)
}

// stringType is the implementation of the TEXT type.
type stringType struct {
	baseType
}

// SQL implements the Type interface.
func (t *stringType) SQL() string {
	return "TEXT"
}

// Compare implements the Type interface.
func (t *stringType) Compare(a interface{}, b interface{}) (int, error) {
	if a == nil || b == nil {
		return 0, ErrNullComparison
	}

	aVal, err := ConvertToString(a)
	if err != nil {
		return 0, err
	}
	bVal, err := ConvertToString(b)
	if err != nil {
		return 0, err
	}

	if aVal < bVal {
		return -1, nil
	}
	if aVal > bVal {
		return 1, nil
	}
	return 0, nil
}

// Convert implements the Type interface.
func (t *stringType) Convert(v interface{}) (interface{}, error) {
	return ConvertToString(v)
}

// Zero implements the Type interface.
func (t *stringType) Zero() interface{} {
	return ""
}

// timestampType is the implementation of the TIMESTAMP type.
type timestampType struct {
	baseType
}

// SQL implements the Type interface.
func (t *timestampType) SQL() string {
	return "TIMESTAMP"
}

// Compare implements the Type interface.
func (t *timestampType) Compare(a interface{}, b interface{}) (int, error) {
	if a == nil || b == nil {
		return 0, ErrNullComparison
	}

	aVal, err := ConvertToTimestamp(a)
	if err != nil {
		return 0, err
	}
	bVal, err := ConvertToTimestamp(b)
	if err != nil {
		return 0, err
	}

	if aVal.Before(bVal) {
		return -1, nil
	}
	if aVal.After(bVal) {
		return 1, nil
	}
	return 0, nil
}

// Convert implements the Type interface.
func (t *timestampType) Convert(v interface{}) (interface{}, error) {
	return ConvertToTimestamp(v)
}

// Zero implements the Type interface.
func (t *timestampType) Zero() interface{} {
	return time.Time{}
}

// nullType is the implementation of the NULL_TYPE type.
type nullType struct {
	baseType
}

// SQL implements the Type interface.
func (t *nullType) SQL() string {
	return "NULL"
}

// Compare implements the Type interface.
func (t *nullType) Compare(a interface{}, b interface{}) (int, error) {
	return 0, ErrNullComparison
}

// Convert implements the Type interface.
func (t *nullType) Convert(v interface{}) (interface{}, error) {
	return nil, nil
}

// Zero implements the Type interface.
func (t *nullType) Zero() interface{} {
	return nil
}

var (
	// Int64 is the INT64 type.
	Int64 Type = &int64Type{baseType{typ: INT64}}
	// Text is the TEXT type.
	Text Type = &stringType{baseType{typ: TEXT}}
	// Timestamp is the TIMESTAMP type.
	Timestamp Type = &timestampType{baseType{typ: TIMESTAMP}}
	// Null is the NULL_TYPE type.
	Null Type = &nullType{baseType{typ: NULL_TYPE}}
)

// MysqlTypeToType converts a query.Type to a Type.
func MysqlTypeToType(qt QueryType) (Type, error) {
	switch qt {
	case INT8, INT16, INT24, INT32, INT64, UINT8, UINT16, UINT24, UINT32, UINT64:
		return Int64, nil
	case TEXT, VARCHAR, CHAR:
		return Text, nil
	case TIMESTAMP, DATETIME, DATE:
		return Timestamp, nil
	case NULL_TYPE:
		return Null, nil
	default:
		return nil, fmt.Errorf("unsupported query type: %v", qt)
	}
}
