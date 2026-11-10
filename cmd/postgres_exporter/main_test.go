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

	"github.com/prometheus-community/postgres_exporter/rdsiam"
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

func TestValidateIAMDataSources(t *testing.T) {
	driverName := rdsiam.DriverName("us-east-1", "")
	for _, tt := range []struct {
		name    string
		dsns    []string
		wantErr bool
	}{
		{"no password, TLS", []string{"postgresql://exporter@db.example.com/postgres?sslmode=require"}, false},
		{"password", []string{"postgresql://exporter:hunter2@db.example.com/postgres"}, true},
		{"second DSN without TLS", []string{
			"postgresql://exporter@a.example.com/postgres",
			"postgresql://exporter@b.example.com/postgres?sslmode=disable",
		}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIAMDataSources(driverName, tt.dsns)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIAMDataSources(%q) error = %v, want error: %v", tt.dsns, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIAMDataSourcesKeepsPasswordOutOfError(t *testing.T) {
	// A malformed URL makes the URL parser quote the whole input, password
	// included, and the exporter logs this error when it refuses to start.
	err := validateIAMDataSources(rdsiam.DriverName("us-east-1", ""),
		[]string{"postgresql://exporter:hunter2@db.example.com:badport/postgres"})
	if err == nil {
		t.Fatal("validateIAMDataSources() error = nil, want one for the malformed port")
	}
	if strings.Contains(err.Error(), "hunter2") {
		t.Errorf("validateIAMDataSources() error leaks the password: %v", err)
	}
}
