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

package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseCollectionTimeout(t *testing.T) {
	t.Run("valid duration", func(t *testing.T) {
		got, err := parseCollectionTimeout("30s")
		if err != nil {
			t.Fatalf("parseCollectionTimeout() error = %v", err)
		}
		if want := 30 * time.Second; got != want {
			t.Fatalf("parseCollectionTimeout() = %v, want %v", got, want)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		const input = "not-a-duration"

		got, err := parseCollectionTimeout(input)
		if err == nil {
			t.Fatal("parseCollectionTimeout() error = nil, want error")
		}
		if got != 0 {
			t.Fatalf("parseCollectionTimeout() = %v, want 0", got)
		}
		if want := `invalid collection timeout "not-a-duration"`; !strings.Contains(err.Error(), want) {
			t.Fatalf("parseCollectionTimeout() error = %q, want it to contain %q", err, want)
		}
	})
}
