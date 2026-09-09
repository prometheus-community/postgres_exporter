// Copyright 2022 The Prometheus Authors
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

package config

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// Test_dsn_String is designed to test different dsn combinations for their string representation.
// dsn.String() is designed to be safe to print, redacting any password information and these test
// cases are intended to cover known cases.
func Test_dsn_String(t *testing.T) {
	type fields struct {
		scheme   string
		username string
		password string
		host     string
		path     string
		query    url.Values
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Without Password",
			fields: fields{
				scheme:   "postgresql",
				username: "test",
				host:     "localhost:5432",
				query:    url.Values{},
			},
			want: "postgresql://test@localhost:5432?",
		},
		{
			name: "With Password",
			fields: fields{
				scheme:   "postgresql",
				username: "test",
				password: "supersecret",
				host:     "localhost:5432",
				query:    url.Values{},
			},
			want: "postgresql://test:******@localhost:5432?",
		},
		{
			name: "With Password and Query String",
			fields: fields{
				scheme:   "postgresql",
				username: "test",
				password: "supersecret",
				host:     "localhost:5432",
				query: url.Values{
					"ssldisable": []string{"true"},
				},
			},
			want: "postgresql://test:******@localhost:5432?ssldisable=true",
		},
		{
			name: "With Password, Path, and Query String",
			fields: fields{
				scheme:   "postgresql",
				username: "test",
				password: "supersecret",
				host:     "localhost:5432",
				path:     "/somevalue",
				query: url.Values{
					"ssldisable": []string{"true"},
				},
			},
			want: "postgresql://test:******@localhost:5432/somevalue?ssldisable=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DSN{
				scheme:   tt.fields.scheme,
				username: tt.fields.username,
				password: tt.fields.password,
				host:     tt.fields.host,
				path:     tt.fields.path,
				query:    tt.fields.query,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("dsn.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_dsn_String_Redacted tests the parse-then-print round trip that log lines rely on:
// no matter which form the password arrives in, it must not survive into dsn.String().
// https://github.com/prometheus-community/postgres_exporter/issues/643
func TestDSNStringRedacted(t *testing.T) {
	const password = "S3CRET"

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "password in userinfo",
			input: "postgresql://user:" + password + "@localhost:5432/postgres?sslmode=disable",
		},
		{
			name:  "password as a query parameter",
			input: "postgresql://localhost:5432/postgres?user=test&password=" + password + "&sslmode=disable",
		},
		{
			name:  "password in a key=value connection string",
			input: "host=localhost port=5432 user=test password=" + password + " sslmode=disable",
		},
		{
			name:  "quoted password in a key=value connection string",
			input: "host=localhost user=test password='" + password + " with spaces' sslmode=disable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := dsnFromString(tt.input)
			if err != nil {
				t.Fatalf("dsnFromString(%q) error = %v", tt.input, err)
			}
			got := d.String()
			if strings.Contains(got, password) {
				t.Errorf("dsn.String() leaked the password: %v", got)
			}
			if !strings.Contains(got, "******") {
				t.Errorf("dsn.String() did not redact the password: %v", got)
			}
		})
	}
}

// Test_dsnFromString tests the dsnFromString function with known variations
// of connection string inputs to ensure that it properly parses the input into
// a dsn.
func Test_dsnFromString(t *testing.T) {

	tests := []struct {
		name    string
		input   string
		want    DSN
		wantErr bool
	}{
		{
			name:  "Key value with password",
			input: "host=host.example.com user=postgres port=5432 password=s3cr3t",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				username: "postgres",
				password: "s3cr3t",
				query:    url.Values{},
			},
			wantErr: false,
		},
		{
			name:  "Key value with quoted password and space",
			input: "host=host.example.com user=postgres port=5432 password=\"s3cr 3t\"",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				username: "postgres",
				password: "s3cr 3t",
				query:    url.Values{},
			},
			wantErr: false,
		},
		{
			name:  "Key value with different order",
			input: "password=abcde host=host.example.com user=postgres port=5432",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				username: "postgres",
				password: "abcde",
				query:    url.Values{},
			},
			wantErr: false,
		},
		{
			name:  "Key value with different order, quoted password, duplicate password",
			input: "password=abcde host=host.example.com user=postgres port=5432 password=\"s3cr 3t\"",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				username: "postgres",
				password: "s3cr 3t",
				query:    url.Values{},
			},
			wantErr: false,
		},
		{
			name:  "URL with user in query string",
			input: "postgresql://host.example.com:5432/tsdb?user=postgres",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				path:     "/tsdb",
				query:    url.Values{},
				username: "postgres",
			},
			wantErr: false,
		},
		{
			name:  "URL with user and password",
			input: "postgresql://user:s3cret@host.example.com:5432/tsdb?user=postgres",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				path:     "/tsdb",
				query:    url.Values{},
				username: "user",
				password: "s3cret",
			},
			wantErr: false,
		},
		{
			name:  "Alternative URL prefix",
			input: "postgres://user:s3cret@host.example.com:5432/tsdb?user=postgres",
			want: DSN{
				scheme:   "postgres",
				host:     "host.example.com:5432",
				path:     "/tsdb",
				query:    url.Values{},
				username: "user",
				password: "s3cret",
			},
			wantErr: false,
		},
		{
			name:  "URL with user and password in query string",
			input: "postgresql://host.example.com:5432/tsdb?user=postgres&password=s3cr3t",
			want: DSN{
				scheme:   "postgresql",
				host:     "host.example.com:5432",
				path:     "/tsdb",
				query:    url.Values{},
				username: "postgres",
				password: "s3cr3t",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dsnFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dsnFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dsnFromString() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test_DSN_WithDatabase tests that WithDatabase overrides whichever database
// a dsn originally targeted, regardless of how that database was specified.
func Test_DSN_WithDatabase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "URL with path-based database",
			input: "postgresql://user:pass@host.example.com:5432/original?sslmode=disable",
			want:  "postgresql://user:pass@host.example.com:5432?dbname=other&sslmode=disable",
		},
		{
			name:  "key-value with dbname",
			input: "host=host.example.com dbname=original sslmode=disable",
			want:  "postgresql://host.example.com?dbname=other&sslmode=disable",
		},
		{
			name:  "key-value without dbname",
			input: "host=host.example.com sslmode=disable",
			want:  "postgresql://host.example.com?dbname=other&sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := dsnFromString(tt.input)
			if err != nil {
				t.Fatalf("dsnFromString() error = %v", err)
			}
			got := d.WithDatabase("other").GetConnectionString()
			if got != tt.want {
				t.Fatalf("WithDatabase(\"other\").GetConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
