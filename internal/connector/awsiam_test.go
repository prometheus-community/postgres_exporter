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
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/lib/pq"
)

var _ driver.Connector = (*AWSIAMConnector)(nil)

func noopDSN(token string) string {
	return "postgresql://iamuser:" + token + "@db.example.com:5432/mydb"
}

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

func TestGetCredentialsWaiterHonoursContext(t *testing.T) {
	c := &AWSIAMConnector{loading: make(chan struct{})} // a load is in flight and never finishes
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, _, err := c.getCredentials(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("getCredentials() error = %v, want context.DeadlineExceeded", err)
	}
}

// fakeCreds hands out a different access key on every Retrieve, like
// temporary credentials that were refreshed between connection attempts.
type fakeCreds struct {
	mu sync.Mutex
	n  int
}

func (f *fakeCreds) Retrieve(context.Context) (aws.Credentials, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return aws.Credentials{AccessKeyID: fmt.Sprintf("AKID%d", f.n), SecretAccessKey: "secret", Source: "fake"}, nil
}

// newTestConnector returns a connector whose credential loading and
// connection opening are faked, plus a slice of the pq.Configs it opened.
func newTestConnector(t *testing.T, load func(context.Context, string) (aws.CredentialsProvider, string, error)) (*AWSIAMConnector, *[]pq.Config) {
	t.Helper()
	c, err := NewAWSIAMConnector("db.example.com:5433", "iamuser", "us-east-1", "", func(token string) string {
		return (&url.URL{Scheme: "postgresql", User: url.UserPassword("iamuser", token), Host: "db.example.com:5433", Path: "/mydb"}).String()
	})
	if err != nil {
		t.Fatal(err)
	}
	var opened []pq.Config
	c.loadCreds = load
	c.open = func(_ context.Context, cfg pq.Config) (driver.Conn, error) {
		opened = append(opened, cfg)
		return nil, nil
	}
	return c, &opened
}

func TestConnectMintsFreshTokenPerConnection(t *testing.T) {
	creds := &fakeCreds{}
	loads := 0
	c, opened := newTestConnector(t, func(context.Context, string) (aws.CredentialsProvider, string, error) {
		loads++
		return creds, "us-east-1", nil
	})

	for i := 0; i < 2; i++ {
		if _, err := c.Connect(context.Background()); err != nil {
			t.Fatalf("Connect() #%d error = %v", i+1, err)
		}
	}

	if loads != 1 {
		t.Errorf("credentials loaded %d times, want 1 (the provider is cached)", loads)
	}
	if len(*opened) != 2 {
		t.Fatalf("opened %d connections, want 2", len(*opened))
	}
	for i, cfg := range *opened {
		if cfg.User != "iamuser" || cfg.Host != "db.example.com" || cfg.Port != 5433 {
			t.Errorf("connection #%d = %s@%s:%d, want iamuser@db.example.com:5433", i+1, cfg.User, cfg.Host, cfg.Port)
		}
		// The password is a presigned URL for the signed endpoint and user, made
		// with the credentials retrieved for this attempt (AKID1, then AKID2).
		for _, want := range []string{"db.example.com:5433", "DBUser=iamuser", fmt.Sprintf("AKID%d%%2F", i+1), "us-east-1"} {
			if !strings.Contains(cfg.Password, want) {
				t.Errorf("connection #%d token missing %q: %s", i+1, want, cfg.Password)
			}
		}
	}
	if (*opened)[0].Password == (*opened)[1].Password {
		t.Error("two connections used the same token, want a fresh one each time")
	}
}

func TestConnectRetriesFailedCredentialLoad(t *testing.T) {
	loads := 0
	c, opened := newTestConnector(t, func(context.Context, string) (aws.CredentialsProvider, string, error) {
		loads++
		if loads == 1 {
			return nil, "", errors.New("imds unavailable")
		}
		return &fakeCreds{}, "us-east-1", nil
	})

	if _, err := c.Connect(context.Background()); err == nil {
		t.Fatal("first Connect() error = nil, want the credential load failure")
	}
	if _, err := c.Connect(context.Background()); err != nil {
		t.Fatalf("second Connect() error = %v, want the load to be retried and succeed", err)
	}
	if loads != 2 || len(*opened) != 1 {
		t.Errorf("loads = %d, opened = %d, want 2 and 1", loads, len(*opened))
	}
}

func TestConnectCancelledDuringCredentialLoad(t *testing.T) {
	started := make(chan struct{})
	c, opened := newTestConnector(t, func(ctx context.Context, _ string) (aws.CredentialsProvider, string, error) {
		close(started)
		<-ctx.Done()
		return nil, "", ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { _, err := c.Connect(ctx); errc <- err }()
	<-started
	cancel()

	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Connect() error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Connect() did not return after its context was cancelled")
	}
	if len(*opened) != 0 {
		t.Errorf("opened %d connections after cancellation, want 0", len(*opened))
	}
}

func TestConnectRejectsConnectionService(t *testing.T) {
	serviceFile := filepath.Join(t.TempDir(), "pg_service.conf")
	if err := os.WriteFile(serviceFile, []byte("[plain]\nsslmode=disable\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PGSERVICEFILE", serviceFile)
	t.Setenv("PGSERVICE", "plain")

	c, opened := newTestConnector(t, func(context.Context, string) (aws.CredentialsProvider, string, error) {
		return &fakeCreds{}, "us-east-1", nil
	})
	if _, err := c.Connect(context.Background()); err == nil {
		t.Fatal("Connect() error = nil, want a service to be rejected")
	}
	if len(*opened) != 0 {
		t.Errorf("opened %d connections, want 0: the token must not be sent when a service could override it", len(*opened))
	}
}
