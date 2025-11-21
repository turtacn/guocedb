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
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	gob.Register(int64(0))
	gob.Register("")
	gob.Register(time.Time{})
}

func TestValueSerialization(t *testing.T) {
	testCases := []struct {
		name string
		val  *Value
	}{
		{"int64", &Value{typ: Int64, data: int64(123)}},
		{"string", &Value{typ: Text, data: "hello"}},
		{"timestamp", &Value{typ: Timestamp, data: time.Now()}},
		{"null", &Value{typ: Null, data: nil}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := tc.val.ToBytes()
			require.NoError(t, err)

			newValue := &Value{}
			err = newValue.FromBytes(bytes)
			require.NoError(t, err)

			// Can't directly compare time.Time with require.Equal
			if tc.val.typ.QueryType() == TIMESTAMP {
				require.Equal(t, tc.val.typ, newValue.typ)
				require.True(t, tc.val.data.(time.Time).Equal(newValue.data.(time.Time)))
			} else {
				require.Equal(t, tc.val, newValue)
			}
		})
	}
}
