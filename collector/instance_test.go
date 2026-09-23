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

import "testing"

func TestInstanceWithDatabase(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    string
		wantErr bool
	}{
		{name: "postgresql URI", dsn: "postgresql://user:pass@localhost:5432/postgres?sslmode=disable", want: "postgresql://user:pass@localhost:5432?dbname=other&sslmode=disable"},
		{name: "postgres URI", dsn: "postgres://localhost/postgres", want: "postgres://localhost?dbname=other"},
		{name: "connstring", dsn: "host=localhost port=5432 dbname=postgres", want: "postgresql://localhost:5432?dbname=other"},
		{name: "unparsable", dsn: "not a dsn", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			primary := &instance{dsn: test.dsn, wrapLargeCounters: true}
			got, err := primary.withDatabase("other")
			if test.wantErr {
				if err == nil {
					t.Fatalf("withDatabase() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("withDatabase() error = %v", err)
			}
			if got.dsn != test.want {
				t.Fatalf("withDatabase() dsn = %q, want %q", got.dsn, test.want)
			}
			if got.wrapLargeCounters != primary.wrapLargeCounters {
				t.Fatalf("withDatabase() wrapLargeCounters = %v, want %v", got.wrapLargeCounters, primary.wrapLargeCounters)
			}
		})
	}
}
