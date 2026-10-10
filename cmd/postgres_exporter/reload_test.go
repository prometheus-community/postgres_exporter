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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHandleReload(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		configFile string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "successful reload",
			method:     http.MethodPost,
			configFile: "../../config/testdata/config-good.yaml",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid config",
			method:     http.MethodPost,
			configFile: "../../config/testdata/config-bad-auth-module.yaml",
			wantStatus: http.StatusInternalServerError,
			wantBody:   "failed to reload configuration:",
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			configFile: "../../config/testdata/config-good.yaml",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHandler, err := config.NewHandler(prometheus.NewRegistry())
			if err != nil {
				t.Fatalf("NewHandler() error = %v", err)
			}

			handler := handleReload(authHandler, tt.configFile)
			request := httptest.NewRequest(tt.method, "/-/reload", nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if got := response.Code; got != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", got, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Fatalf("response body = %q, want it to contain %q", response.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandleReloadConcurrentRequests(t *testing.T) {
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	handler := handleReload(authHandler, "../../config/testdata/config-good.yaml")

	const requests = 10
	errs := make(chan error, requests)

	var wg sync.WaitGroup
	for range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			request := httptest.NewRequest(http.MethodPost, "/-/reload", nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				errs <- fmt.Errorf("status code = %d, want %d", response.Code, http.StatusOK)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestHandleReloadSignal(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "postgres_exporter.yml")

	initial := []byte(`auth_modules:
  first:
    type: userpass
    userpass:
      username: before
      password: firstpass
`)
	if err := os.WriteFile(configFile, initial, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	if err := authHandler.ReloadAuthConfig(configFile, logger); err != nil {
		t.Fatalf("ReloadAuthConfig() initial load error = %v", err)
	}

	updated := []byte(`auth_modules:
  first:
    type: userpass
    userpass:
      username: after
      password: firstpass
`)
	if err := os.WriteFile(configFile, updated, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	go func() {
		handleReloadSignal(signals, authHandler, configFile, logger)
		close(done)
	}()

	signals <- syscall.SIGHUP
	close(signals)
	<-done

	if got, want := authHandler.GetAuthConfig().AuthModules["first"].UserPass.Username, "after"; got != want {
		t.Fatalf("username after SIGHUP = %q, want %q", got, want)
	}
}
