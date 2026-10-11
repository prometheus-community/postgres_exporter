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

package exporter

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestScrapeLogsOmitPasswords scrapes connection strings that cannot be
// connected to and checks that nothing logged along the way contains the
// password. Each case waits out one GetServer retry, so they run in parallel.
func TestScrapeLogsOmitPasswords(t *testing.T) {
	tests := []struct {
		name         string
		dsn          string
		autoDiscover bool
		secrets      []string
	}{
		{
			// The form GetDataSources builds from DATA_SOURCE_USER and
			// DATA_SOURCE_PASS (issue #1408). Port 1 refuses the connection.
			name:    "url with encoded space in password",
			dsn:     "postgresql://demo:dummy%20secret-PW@127.0.0.1:1/demo?sslmode=disable",
			secrets: []string{"dummy secret-PW", "dummy%20secret-PW", "secret-PW"},
		},
		{
			name:    "url with invalid escape in password",
			dsn:     "postgresql://demo:bad%zzpw@127.0.0.1:1/demo?sslmode=disable",
			secrets: []string{"bad%zzpw"},
		},
		{
			name:         "url with invalid escape in password, auto discovery",
			dsn:          "postgresql://demo:bad%zzpw@127.0.0.1:1/demo?sslmode=disable",
			autoDiscover: true,
			secrets:      []string{"bad%zzpw"},
		},
		{
			name:    "key value with unquoted password containing a space",
			dsn:     "host=127.0.0.1 port=1 user=demo password=dummy secret-PW sslmode=disable",
			secrets: []string{"dummy", "secret-PW"},
		},
		{
			name:         "key value with unquoted password containing a space, auto discovery",
			dsn:          "host=127.0.0.1 port=1 user=demo password=dummy secret-PW sslmode=disable",
			autoDiscover: true,
			secrets:      []string{"dummy", "secret-PW"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			e := NewExporter([]string{tt.dsn}, logger, AutoDiscoverDatabases(tt.autoDiscover))

			ch := make(chan prometheus.Metric)
			done := make(chan struct{})
			go func() {
				for range ch {
				}
				close(done)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			e.scrape(ctx, ch)
			close(ch)
			<-done

			logs := buf.String()
			if !strings.Contains(logs, "level=ERROR") {
				t.Fatalf("scrape() logged no error, want the failed connection to be logged; logs:\n%s", logs)
			}
			for _, secret := range tt.secrets {
				if strings.Contains(logs, secret) {
					t.Errorf("scrape() logs contain %q:\n%s", secret, logs)
				}
			}
		})
	}
}
