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
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestHandleProbeRequiresTarget(t *testing.T) {
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	handler := handleProbe(
		promslog.NewNopLogger(),
		authHandler,
		config.NewConfigWithDefaults(),
	)
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
}

// syncBuffer is a bytes.Buffer that collectors can log to concurrently.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestHandleProbeConnectsThroughIAMModule(t *testing.T) {
	// Keep the developer's own AWS setup out. Without credentials the IAM
	// driver fails fast, which is enough to see that it handled the probe.
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{
		"AWS_CONFIG_FILE":                        empty,
		"AWS_SHARED_CREDENTIALS_FILE":            empty,
		"AWS_EC2_METADATA_DISABLED":              "true",
		"AWS_PROFILE":                            "",
		"AWS_ACCESS_KEY_ID":                      "",
		"AWS_SECRET_ACCESS_KEY":                  "",
		"AWS_SESSION_TOKEN":                      "",
		"AWS_WEB_IDENTITY_TOKEN_FILE":            "",
		"AWS_CONTAINER_CREDENTIALS_FULL_URI":     "",
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "",
	} {
		t.Setenv(k, v)
	}

	configFile := filepath.Join(dir, "postgres_exporter.yml")
	if err := os.WriteFile(configFile, []byte(`auth_modules:
  rds:
    type: iam
    iam:
      region: us-east-1
    options:
      user: exporter
      sslmode: require
`), 0o600); err != nil {
		t.Fatal(err)
	}
	var logs syncBuffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	if err := authHandler.ReloadAuthConfig(configFile, logger); err != nil {
		t.Fatalf("ReloadAuthConfig() error = %v", err)
	}
	baseConfig := config.NewConfigWithDefaults()
	baseConfig.CollectionTimeout = 500 * time.Millisecond

	handler := handleProbe(logger, authHandler, baseConfig)
	request := httptest.NewRequest(http.MethodGet, "/probe?target=127.0.0.1:1/postgres&auth_module=rds", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	// The module has no username or password, and that's fine for IAM.
	if got, want := response.Code, http.StatusOK; got != want {
		t.Fatalf("status code = %d, want %d: %s", got, want, response.Body)
	}
	if !strings.Contains(logs.String(), "rds iam:") {
		t.Errorf("probe didn't connect through the IAM driver; logs:\n%s", logs.String())
	}
}
