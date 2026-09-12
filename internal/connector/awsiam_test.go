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

package connector

import (
	"database/sql/driver"
	"testing"
)

var _ driver.Connector = (*AWSIAMConnector)(nil)

func noopDSN(token string) string { return "postgresql://iamuser:" + token + "@db.example.com:5432/mydb" }

func TestNewAWSIAMConnector(t *testing.T) {
	c, err := NewAWSIAMConnector("db.example.com:5432", "iamuser", "us-east-1", "arn:aws:iam::123456789012:role/rds-connect", noopDSN)
	if err != nil {
		t.Fatalf("NewAWSIAMConnector() error = %v", err)
	}

	if got, want := c.DBEndpoint, "db.example.com:5432"; got != want {
		t.Errorf("DBEndpoint = %q, want %q", got, want)
	}
	if got, want := c.DBUser, "iamuser"; got != want {
		t.Errorf("DBUser = %q, want %q", got, want)
	}
	if got, want := c.Region, "us-east-1"; got != want {
		t.Errorf("Region = %q, want %q", got, want)
	}
	if got, want := c.RoleARN, "arn:aws:iam::123456789012:role/rds-connect"; got != want {
		t.Errorf("RoleARN = %q, want %q", got, want)
	}
	if got, want := c.DSN("token"), "postgresql://iamuser:token@db.example.com:5432/mydb"; got != want {
		t.Errorf("DSN(token) = %q, want %q", got, want)
	}
}

func TestAWSIAMConnectorDriver(t *testing.T) {
	c, err := NewAWSIAMConnector("db.example.com:5432", "iamuser", "us-east-1", "", noopDSN)
	if err != nil {
		t.Fatalf("NewAWSIAMConnector() error = %v", err)
	}
	if c.Driver() == nil {
		t.Fatal("Driver() = nil, want non-nil")
	}
}

func TestAWSIAMConnectorPort(t *testing.T) {
	t.Run("missing port defaults to 5432, matching lib/pq", func(t *testing.T) {
		c, err := NewAWSIAMConnector("db.example.com", "iamuser", "us-east-1", "", noopDSN)
		if err != nil {
			t.Fatalf("NewAWSIAMConnector() error = %v", err)
		}
		if got, want := c.DBEndpoint, "db.example.com:5432"; got != want {
			t.Errorf("DBEndpoint = %q, want %q", got, want)
		}
	})

	t.Run("explicit port is left untouched", func(t *testing.T) {
		c, err := NewAWSIAMConnector("db.example.com:1234", "iamuser", "us-east-1", "", noopDSN)
		if err != nil {
			t.Fatalf("NewAWSIAMConnector() error = %v", err)
		}
		if got, want := c.DBEndpoint, "db.example.com:1234"; got != want {
			t.Errorf("DBEndpoint = %q, want %q", got, want)
		}
	})

	t.Run("malformed endpoint returns a friendly error", func(t *testing.T) {
		_, err := NewAWSIAMConnector("db.example.com:1234:5678", "iamuser", "us-east-1", "", noopDSN)
		if err == nil {
			t.Fatal("NewAWSIAMConnector() error = nil, want error")
		}
	})
}
