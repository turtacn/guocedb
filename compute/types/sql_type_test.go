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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTypeComparison(t *testing.T) {
	testCases := []struct {
		typ      Type
		a, b     interface{}
		expected int
		err      error
	}{
		// Int64
		{Int64, int64(1), int64(2), -1, nil},
		{Int64, int64(2), int64(1), 1, nil},
		{Int64, int64(1), int64(1), 0, nil},
		{Int64, "1", int64(2), -1, nil},
		{Int64, "2", int64(1), 1, nil},
		{Int64, "1", int64(1), 0, nil},
		{Int64, nil, int64(1), 0, ErrNullComparison},
		{Int64, int64(1), nil, 0, ErrNullComparison},

		// Text
		{Text, "a", "b", -1, nil},
		{Text, "b", "a", 1, nil},
		{Text, "a", "a", 0, nil},
		{Text, nil, "a", 0, ErrNullComparison},
		{Text, "a", nil, 0, ErrNullComparison},

		// Timestamp
		{Timestamp, time.Unix(1, 0), time.Unix(2, 0), -1, nil},
		{Timestamp, time.Unix(2, 0), time.Unix(1, 0), 1, nil},
		{Timestamp, time.Unix(1, 0), time.Unix(1, 0), 0, nil},
		{Timestamp, "2024-01-01 00:00:00", time.Unix(1704067201, 0), -1, nil},
		{Timestamp, nil, time.Unix(1, 0), 0, ErrNullComparison},
	}

	for _, tc := range testCases {
		res, err := tc.typ.Compare(tc.a, tc.b)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.expected, res)
	}
}

func TestTypeConversion(t *testing.T) {
	testCases := []struct {
		typ      Type
		val      interface{}
		expected interface{}
		err      bool
	}{
		// Int64
		{Int64, "123", int64(123), false},
		{Int64, "123abc", int64(123), false},
		{Int64, "", int64(0), false},
		{Int64, "abc", nil, true},
		{Int64, 123.456, int64(123), false},
		{Int64, uint64(9223372036854775808), nil, true}, // MaxInt64 + 1

		// Text
		{Text, 123, "123", false},
		{Text, 123.456, "123.456", false},

		// Timestamp
		{Timestamp, "2024-01-01 10:20:30", time.Date(2024, 1, 1, 10, 20, 30, 0, time.UTC), false},
		{Timestamp, "invalid-time", nil, true},
	}

	for _, tc := range testCases {
		res, err := tc.typ.Convert(tc.val)
		if tc.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			// Special handling for time.Time due to location differences
			if expectedTime, ok := tc.expected.(time.Time); ok {
				actualTime, ok := res.(time.Time)
				require.True(t, ok)
				require.True(t, expectedTime.Equal(actualTime), "expected: %v, got: %v", expectedTime, actualTime)
			} else {
				require.Equal(t, tc.expected, res)
			}
		}
	}
}
