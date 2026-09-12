// Copyright 2021 The Prometheus Authors
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

package connector

import (
	"database/sql/driver"
	"testing"
)

var _ driver.Connector = (*StaticConnector)(nil)

func TestNewStaticConnector(t *testing.T) {
	t.Run("valid dsn", func(t *testing.T) {
		conn, err := NewStaticConnector("postgres://user:pass@localhost:5432/mydb?sslmode=disable")
		if err != nil {
			t.Fatalf("NewStaticConnector() error = %v", err)
		}
		if conn == nil {
			t.Fatal("NewStaticConnector() = nil, want non-nil")
		}
	})

	t.Run("invalid dsn", func(t *testing.T) {
		if _, err := NewStaticConnector("postgres://localhost:5432/mydb?sslmode=bogus"); err == nil {
			t.Fatal("NewStaticConnector() error = nil, want error for an unsupported sslmode")
		}
	})
}

func TestStaticConnectorDriver(t *testing.T) {
	conn, err := NewStaticConnector("postgres://localhost:5432/mydb")
	if err != nil {
		t.Fatalf("NewStaticConnector() error = %v", err)
	}
	if conn.Driver() == nil {
		t.Fatal("Driver() = nil, want non-nil")
	}
}
