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

// QueryType is an enum that represents a type of a SQL expression.
type QueryType int

const (
	// NULL_TYPE is a null type.
	NULL_TYPE QueryType = iota
	// INT8 is an 8-bit signed integer.
	INT8
	// INT16 is a 16-bit signed integer.
	INT16
	// INT24 is a 24-bit signed integer.
	INT24
	// INT32 is a 32-bit signed integer.
	INT32
	// INT64 is a 64-bit signed integer.
	INT64
	// UINT8 is an 8-bit unsigned integer.
	UINT8
	// UINT16 is a 16-bit unsigned integer.
	UINT16
	// UINT24 is a 24-bit unsigned integer.
	UINT24
	// UINT32 is a 32-bit unsigned integer.
	UINT32
	// UINT64 is a 64-bit unsigned integer.
	UINT64
	// FLOAT32 is a 32-bit floating point number.
	FLOAT32
	// FLOAT64 is a 64-bit floating point number.
	FLOAT64
	// TIMESTAMP is a timestamp.
	TIMESTAMP
	// DATE is a date.
	DATE
	// TIME is a time.
	TIME
	// DATETIME is a datetime.
	DATETIME
	// YEAR is a year.
	YEAR
	// DECIMAL is a decimal.
	DECIMAL
	// TEXT is a text.
	TEXT
	// BLOB is a blob.
	BLOB
	// VARCHAR is a varchar.
	VARCHAR
	// VARBINARY is a varbinary.
	VARBINARY
	// CHAR is a char.
	CHAR
	// BINARY is a binary.
	BINARY
	// JSON is a json.
	JSON
	// ENUM is an enum.
	ENUM
	// SET is a set.
	SET
)
