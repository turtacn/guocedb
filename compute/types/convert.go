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
	"math"
	"strconv"
	"time"
)

var (
	ErrOutOfRange        = fmt.Errorf("value out of range")
	ErrInvalidConversion = fmt.Errorf("invalid conversion")
)

// ConvertToInt64 converts a value to an int64.
func ConvertToInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		if uint64(val) > math.MaxInt64 {
			return 0, ErrOutOfRange
		}
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		if val > math.MaxInt64 {
			return 0, ErrOutOfRange
		}
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		if val == "" {
			return 0, nil
		}
		// MySQL compatible conversion: "123abc" -> 123
		for i, c := range val {
			if c < '0' || c > '9' {
				val = val[:i]
				break
			}
		}
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrInvalidConversion, err)
		}
		return i, nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: cannot convert %T to int64", ErrInvalidConversion, v)
	}
}

// ConvertToString converts a value to a string.
func ConvertToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case []byte:
		return string(val), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ConvertToTimestamp converts a value to a time.Time.
func ConvertToTimestamp(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02",
			time.RFC3339,
			time.RFC3339Nano,
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, val); err == nil {
				return t, nil
			}
		}
		// Try parsing as Unix timestamp
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return time.Unix(i, 0), nil
		}
		return time.Time{}, fmt.Errorf("%w: cannot convert %q to timestamp", ErrInvalidConversion, val)
	case int64:
		return time.Unix(val, 0), nil
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("%w: cannot convert %T to timestamp", ErrInvalidConversion, v)
	}
}

// MustConvert converts a value to a given type, panicking on error.
func MustConvert(typ QueryType, v interface{}) interface{} {
	val, err := Convert(typ, v)
	if err != nil {
		panic(err)
	}
	return val
}

// Convert converts a value to a given type.
func Convert(typ QueryType, v interface{}) (interface{}, error) {
	switch typ {
	case INT8, INT16, INT24, INT32, INT64:
		return ConvertToInt64(v)
	case UINT8, UINT16, UINT24, UINT32, UINT64:
		// For simplicity, we'll use int64 for unsigned as well for now.
		// This can be changed later if needed.
		return ConvertToInt64(v)
	case TEXT, VARCHAR, CHAR:
		return ConvertToString(v)
	case TIMESTAMP, DATETIME, DATE:
		return ConvertToTimestamp(v)
	case NULL_TYPE:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported type for conversion: %v", typ)
	}
}
