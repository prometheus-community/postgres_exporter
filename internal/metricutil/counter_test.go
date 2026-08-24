// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metricutil

import "testing"

func TestInt64CounterValue(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		wrap  bool
		want  float64
	}{
		{
			name:  "largest exactly represented integer",
			value: int64(1<<53) - 1,
			wrap:  true,
			want:  float64(int64(1<<53) - 1),
		},
		{
			name:  "wrap boundary",
			value: int64(1 << 53),
			wrap:  true,
			want:  0,
		},
		{
			name:  "above wrap boundary",
			value: int64(1<<53) + 1,
			wrap:  true,
			want:  1,
		},
		{
			name:  "upgrade warning example",
			value: int64(1<<53) + 1_234_567,
			wrap:  true,
			want:  1_234_567,
		},
		{
			name:  "wrapping disabled",
			value: int64(1<<53) + 1,
			wrap:  false,
			want:  float64(int64(1<<53) + 1),
		},
		{
			name:  "negative value",
			value: -1,
			wrap:  true,
			want:  -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Int64CounterValue(test.value, test.wrap); got != test.want {
				t.Fatalf("Int64CounterValue(%d, %t) = %v, want %v", test.value, test.wrap, got, test.want)
			}
		})
	}
}
