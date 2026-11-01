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

package rdsiam

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/lib/pq"
)

func TestOpenConnectorRejects(t *testing.T) {
	for name, dsn := range map[string]string{
		"password":      "postgresql://user:secret@db.example.com/postgres",
		"multiple host": "host=a.example.com,b.example.com user=user dbname=postgres",
		"sslmode":       "postgresql://user@db.example.com/postgres?sslmode=prefer",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := (&Driver{}).OpenConnector(dsn); err == nil {
				t.Fatalf("OpenConnector(%q) succeeded, want error", dsn)
			}
		})
	}
}

func TestConnectMintsTokenForEachConnection(t *testing.T) {
	origLoad, origBuild, origConnect := loadAWSConfig, buildToken, connect
	t.Cleanup(func() { loadAWSConfig, buildToken, connect = origLoad, origBuild, origConnect })
	var loads int
	loadAWSConfig = func(_ context.Context, region, roleARN string) (aws.Config, error) {
		loads++
		if region != "us-east-1" || roleARN != "arn:aws:iam::123456789012:role/exporter" {
			t.Errorf("loadAWSConfig(%q, %q), want the driver's region and role", region, roleARN)
		}
		return aws.Config{Region: region}, nil
	}
	var calls int
	buildToken = func(_ context.Context, endpoint, region, user string, _ aws.CredentialsProvider, _ ...func(*auth.BuildAuthTokenOptions)) (string, error) {
		calls++
		if endpoint != "db.example.com:6543" || region != "us-east-1" || user != "exporter" {
			t.Errorf("buildToken(%q, %q, %q), want db.example.com:6543, us-east-1, exporter", endpoint, region, user)
		}
		return "token", nil
	}
	var got pq.Config
	connect = func(_ context.Context, cfg pq.Config) (driver.Conn, error) {
		got = cfg
		return nil, errors.New("not connecting in tests")
	}

	// Discovery rewrites the database; the token only depends on host, port
	// and user, so both DSN forms work unchanged.
	d := &Driver{region: "us-east-1", roleARN: "arn:aws:iam::123456789012:role/exporter"}
	for _, dsn := range []string{
		"postgresql://exporter@db.example.com:6543/postgres",
		"host=db.example.com port=6543 user=exporter dbname=other",
	} {
		c, err := d.OpenConnector(dsn)
		if err != nil {
			t.Fatalf("OpenConnector(%q): %v", dsn, err)
		}
		_, _ = c.Connect(context.Background())
		if got.Password != "token" {
			t.Errorf("connected with password %q, want token", got.Password)
		}
	}
	if calls != 2 {
		t.Errorf("built %d tokens, want 2", calls)
	}
	if loads != 1 {
		t.Errorf("loaded AWS config %d times, want once per driver", loads)
	}
}

func TestDriverNamePerRegionAndRole(t *testing.T) {
	a := DriverName("us-east-1", "arn:aws:iam::123456789012:role/a")
	if again := DriverName("us-east-1", "arn:aws:iam::123456789012:role/a"); again != a {
		t.Errorf("DriverName returned %q then %q for the same region and role", a, again)
	}
	if b := DriverName("us-east-1", "arn:aws:iam::123456789012:role/b"); b == a {
		t.Errorf("DriverName returned %q for two different roles", a)
	}
}

func TestAWSConfigWaitStopsWhenContextIsDone(t *testing.T) {
	origLoad := loadAWSConfig
	t.Cleanup(func() { loadAWSConfig = origLoad })
	started, release := make(chan struct{}), make(chan struct{})
	loadAWSConfig = func(context.Context, string, string) (aws.Config, error) {
		close(started)
		<-release
		return aws.Config{Region: "us-east-1"}, nil
	}
	d := &Driver{region: "us-east-1"}

	firstDone := make(chan error, 1)
	go func() {
		_, err := d.aws(context.Background())
		firstDone <- err
	}()
	<-started

	// Another connection, e.g. a second /probe target, whose scrape times out
	// while the first one is still loading the AWS config.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	waited := make(chan error, 1)
	go func() {
		_, err := d.aws(ctx)
		waited <- err
	}()
	select {
	case err := <-waited:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("aws() error = %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("aws() kept waiting for the other load after its context was done")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Errorf("first aws() error = %v", err)
	}
}

// isolateAWSEnv keeps the developer's own AWS setup out of the test.
func isolateAWSEnv(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{
		"AWS_REGION":                  "",
		"AWS_DEFAULT_REGION":          "",
		"AWS_PROFILE":                 "",
		"AWS_CONFIG_FILE":             empty,
		"AWS_SHARED_CREDENTIALS_FILE": empty,
		"AWS_EC2_METADATA_DISABLED":   "",
	} {
		t.Setenv(k, v)
	}
}

// fakeIMDS serves the EC2 instance metadata endpoints that the region lookup
// uses and returns how many times the region was asked for. The SDK reads the
// region from the instance identity document.
func fakeIMDS(t *testing.T, documentStatus int, region string) *atomic.Int32 {
	var regionCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/latest/api/token":
			w.Header().Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")
			fmt.Fprint(w, "imds-token")
		case r.Method == http.MethodGet && r.URL.Path == "/latest/dynamic/instance-identity/document":
			regionCalls.Add(1)
			w.WriteHeader(documentStatus)
			fmt.Fprintf(w, `{
  "accountId": "123456789012",
  "architecture": "arm64",
  "availabilityZone": "%[1]sa",
  "billingProducts": null,
  "devpayProductCodes": null,
  "marketplaceProductCodes": null,
  "imageId": "ami-0123456789abcdef0",
  "instanceId": "i-0123456789abcdef0",
  "instanceType": "t4g.small",
  "kernelId": null,
  "pendingTime": "2026-10-01T12:00:00Z",
  "privateIp": "10.0.1.23",
  "ramdiskId": null,
  "region": "%[1]s",
  "version": "2017-09-30"
}`, region)
		default:
			t.Logf("fake IMDS: unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AWS_EC2_METADATA_SERVICE_ENDPOINT", srv.URL)
	return &regionCalls
}

func TestLoadAWSConfigWithoutRegionAsksInstanceMetadataOnce(t *testing.T) {
	isolateAWSEnv(t)
	regionCalls := fakeIMDS(t, http.StatusInternalServerError, "")

	_, err := loadDefaultAWSConfig(context.Background(), "", "")
	if err == nil || !strings.Contains(err.Error(), "AWS_REGION") {
		t.Errorf("loadDefaultAWSConfig() error = %v, want one that says to set AWS_REGION", err)
	}
	// Outside EC2 this fails on every new connection, so it mustn't retry.
	if n := regionCalls.Load(); n != 1 {
		t.Errorf("asked instance metadata for the region %d times, want 1", n)
	}
}

func TestLoadAWSConfigFallsBackToInstanceRegion(t *testing.T) {
	isolateAWSEnv(t)
	fakeIMDS(t, http.StatusOK, "eu-north-1")

	cfg, err := loadDefaultAWSConfig(context.Background(), "", "")
	if err != nil {
		t.Fatalf("loadDefaultAWSConfig() error = %v", err)
	}
	if cfg.Region != "eu-north-1" {
		t.Errorf("region = %q, want eu-north-1 from instance metadata", cfg.Region)
	}
}
