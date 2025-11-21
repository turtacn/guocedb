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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaValidation(t *testing.T) {
	schema := Schema{
		{Name: "c1", Type: Int64, Nullable: false},
		{Name: "c2", Type: Text, Nullable: true},
	}

	testCases := []struct {
		name string
		row  Row
		err  error
	}{
		{"valid row", Row{int64(1), "a"}, nil},
		{"valid row with null", Row{int64(1), nil}, nil},
		{"non-nullable column is null", Row{nil, "a"}, fmt.Errorf("column c1 cannot be NULL")},
		{"wrong column count (less)", Row{int64(1)}, ErrColumnCountMismatch},
		{"wrong column count (more)", Row{int64(1), "a", "b"}, ErrColumnCountMismatch},
		{"type mismatch", Row{"a", "b"}, ErrInvalidConversion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.CheckRow(tc.row)
			if tc.err != nil {
				if tc.name == "non-nullable column is null" {
					require.EqualError(t, err, tc.err.Error())
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRowEquals(t *testing.T) {
	schema := Schema{
		{Name: "c1", Type: Int64},
		{Name: "c2", Type: Text},
	}

	testCases := []struct {
		name     string
		r1       Row
		r2       Row
		expected bool
	}{
		{"equal rows", Row{int64(1), "a"}, Row{int64(1), "a"}, true},
		{"different int", Row{int64(1), "a"}, Row{int64(2), "a"}, false},
		{"different string", Row{int64(1), "a"}, Row{int64(1), "b"}, false},
		{"nulls are not equal", Row{nil, "a"}, Row{nil, "a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eq, err := tc.r1.Equals(tc.r2, schema)
			require.NoError(t, err)
			require.Equal(t, tc.expected, eq)
		})
	}
}
