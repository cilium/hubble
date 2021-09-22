// Copyright 2021 Authors of Hubble
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

package printer

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint64Grouping(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{
			n:    0,
			want: "0",
		}, {
			n:    1,
			want: "1",
		}, {
			n:    10,
			want: "10",
		}, {
			n:    100,
			want: "100",
		}, {
			n:    1_000,
			want: "1,000",
		}, {
			n:    10_000,
			want: "10,000",
		}, {
			n:    100_000,
			want: "100,000",
		}, {
			n:    1_000_000,
			want: "1,000,000",
		}, {
			n:    math.MaxUint64,
			want: "18,446,744,073,709,551,615",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d => %s", tt.n, tt.want), func(t *testing.T) {
			got := uint64Grouping(tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
}
