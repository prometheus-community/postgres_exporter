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

package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestPQErrorHelpers(t *testing.T) {
	noDataFound := &pq.Error{Code: "P0002"}
	featureNotSupported := &pq.Error{Code: "0A000"}

	for _, tc := range []struct {
		name                      string
		err                       error
		noDataFound, notSupported bool
	}{
		{"no_data_found", noDataFound, true, false},
		{"wrapped no_data_found", fmt.Errorf("query: %w", noDataFound), true, false},
		{"feature_not_supported", featureNotSupported, false, true},
		{"wrapped feature_not_supported", fmt.Errorf("query: %w", featureNotSupported), false, true},
		{"other PostgreSQL error", &pq.Error{Code: "42883"}, false, false},
		{"not a PostgreSQL error", errors.New("connection refused"), false, false},
		{"nil", nil, false, false},
	} {
		if got := isNoDataFound(tc.err); got != tc.noDataFound {
			t.Errorf("%s: isNoDataFound = %v, want %v", tc.name, got, tc.noDataFound)
		}
		if got := isFeatureNotSupported(tc.err); got != tc.notSupported {
			t.Errorf("%s: isFeatureNotSupported = %v, want %v", tc.name, got, tc.notSupported)
		}
	}
}

func TestIsObjectNotInPrerequisiteState(t *testing.T) {
	notInPrerequisiteState := &pq.Error{Code: "55000"}

	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"object_not_in_prerequisite_state", notInPrerequisiteState, true},
		{"wrapped object_not_in_prerequisite_state", fmt.Errorf("query: %w", notInPrerequisiteState), true},
		{"other PostgreSQL error", &pq.Error{Code: "0A000"}, false},
		{"not a PostgreSQL error", errors.New("connection refused"), false},
		{"nil", nil, false},
	} {
		if got := isObjectNotInPrerequisiteState(tc.err); got != tc.want {
			t.Errorf("%s: isObjectNotInPrerequisiteState = %v, want %v", tc.name, got, tc.want)
		}
	}
}
