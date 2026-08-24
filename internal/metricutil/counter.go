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

const float64ExactIntegerLimit uint64 = 1 << 53

// Int64CounterValue converts a PostgreSQL counter while preserving single-unit
// precision above float64's exact integer range when wrapping is enabled.
func Int64CounterValue(value int64, wrap bool) float64 {
	if wrap && value >= 0 {
		return float64(uint64(value) % float64ExactIntegerLimit)
	}

	return float64(value)
}
