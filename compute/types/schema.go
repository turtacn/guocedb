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
	"strings"
)

var (
	ErrColumnCountMismatch = fmt.Errorf("column count mismatch")
)

// Column is a column in a schema.
type Column struct {
	Name       string
	Source     string // table.column
	Type       Type
	Nullable   bool
	PrimaryKey bool
	Default    interface{}
}

// Schema is a collection of columns.
type Schema []*Column

// IndexOf returns the index of the column with the given name.
// It uses a map for O(1) lookups after the first call.
func (s Schema) IndexOf(columnName string) int {
	// Simple linear scan for now. A map can be added for performance later.
	for i, c := range s {
		if strings.EqualFold(c.Name, columnName) {
			return i
		}
	}
	return -1
}

// CheckRow checks if a row is valid for the schema.
func (s Schema) CheckRow(row Row) error {
	if len(row) != len(s) {
		return ErrColumnCountMismatch
	}

	for i, col := range s {
		val := row[i]

		if val == nil && !col.Nullable {
			return fmt.Errorf("column %s cannot be NULL", col.Name)
		}

		if val != nil {
			if _, err := col.Type.Convert(val); err != nil {
				return fmt.Errorf("column %s: %w", col.Name, err)
			}
		}
	}

	return nil
}

// Row is a row of data.
type Row []interface{}

// Copy returns a copy of the row.
func (r Row) Copy() Row {
	newRow := make(Row, len(r))
	copy(newRow, r)
	return newRow
}

// Equals checks if two rows are equal.
func (r Row) Equals(other Row, schema Schema) (bool, error) {
	if len(r) != len(other) {
		return false, nil
	}
	if len(r) != len(schema) {
		return false, ErrColumnCountMismatch
	}

	for i, val := range r {
		cmp, err := schema[i].Type.Compare(val, other[i])
		if err != nil {
			// For NULL comparison, Equals should return false, not an error.
			if err == ErrNullComparison {
				return false, nil
			}
			return false, err
		}
		if cmp != 0 {
			return false, nil
		}
	}

	return true, nil
}
