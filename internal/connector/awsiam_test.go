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
	"net/url"
	"testing"
)

var _ driver.Connector = (*AWSIAMConnector)(nil)

func TestNewAWSIAMConnector(t *testing.T) {
	options := map[string]string{"sslmode": "verify-full"}
	c := NewAWSIAMConnector("db.example.com:5432", "iamuser", "mydb", "us-east-1", "arn:aws:iam::123456789012:role/rds-connect", options)

	if got, want := c.DBEndpoint, "db.example.com:5432"; got != want {
		t.Errorf("DBEndpoint = %q, want %q", got, want)
	}
	if got, want := c.DBUser, "iamuser"; got != want {
		t.Errorf("DBUser = %q, want %q", got, want)
	}
	if got, want := c.Database, "mydb"; got != want {
		t.Errorf("Database = %q, want %q", got, want)
	}
	if got, want := c.Region, "us-east-1"; got != want {
		t.Errorf("Region = %q, want %q", got, want)
	}
	if got, want := c.RoleARN, "arn:aws:iam::123456789012:role/rds-connect"; got != want {
		t.Errorf("RoleARN = %q, want %q", got, want)
	}
	if got, want := c.Options["sslmode"], "verify-full"; got != want {
		t.Errorf(`Options["sslmode"] = %q, want %q`, got, want)
	}
}

func TestAWSIAMConnectorDriver(t *testing.T) {
	c := NewAWSIAMConnector("db.example.com:5432", "iamuser", "mydb", "us-east-1", "", nil)
	if c.Driver() == nil {
		t.Fatal("Driver() = nil, want non-nil")
	}
}

func TestAWSIAMConnectorDSN(t *testing.T) {
	t.Run("no options means no extra query parameters", func(t *testing.T) {
		c := NewAWSIAMConnector("db.example.com:5432", "iamuser", "mydb", "us-east-1", "", nil)
		u, err := url.Parse(c.dsn("token"))
		if err != nil {
			t.Fatalf("dsn() produced an unparseable connection string: %v", err)
		}
		if got := u.Query().Get("sslmode"); got != "" {
			t.Errorf("sslmode = %q, want unset", got)
		}
	})

	t.Run("options are passed through to the query string, same as userpass", func(t *testing.T) {
		c := NewAWSIAMConnector("db.example.com:5432", "iamuser", "mydb", "us-east-1", "", map[string]string{"sslmode": "verify-full"})
		u, err := url.Parse(c.dsn("token"))
		if err != nil {
			t.Fatalf("dsn() produced an unparseable connection string: %v", err)
		}
		if got, want := u.Query().Get("sslmode"), "verify-full"; got != want {
			t.Errorf("sslmode = %q, want %q", got, want)
		}
	})

	t.Run("token round-trips through URL escaping", func(t *testing.T) {
		c := NewAWSIAMConnector("db.example.com:5432", "iamuser", "mydb", "us-east-1", "", nil)

		// A real token is a presigned URL containing '?', '&', '=' and '/',
		// which must round-trip through URL escaping intact for pq to parse
		// it back out as the literal password.
		const token = "db.example.com:5432/?Action=connect&X-Amz-Signature=abc/def"

		u, err := url.Parse(c.dsn(token))
		if err != nil {
			t.Fatalf("dsn() produced an unparseable connection string: %v", err)
		}
		if got, want := u.Host, "db.example.com:5432"; got != want {
			t.Errorf("host = %q, want %q", got, want)
		}
		if got, want := u.Path, "/mydb"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got, want := u.User.Username(), "iamuser"; got != want {
			t.Errorf("user = %q, want %q", got, want)
		}
		password, ok := u.User.Password()
		if !ok {
			t.Fatal("password not set")
		}
		if password != token {
			t.Errorf("password = %q, want %q (token must round-trip through URL escaping)", password, token)
		}
	})
}
