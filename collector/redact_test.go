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

package collector

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestCollectLogsOmitPasswords collects from connection strings the pq
// driver cannot parse, and so echoes back in its error, and checks that
// nothing logged contains the password.
func TestCollectLogsOmitPasswords(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		secrets []string
	}{
		{
			name:    "url with invalid escape in password",
			dsn:     "postgresql://demo:bad%zzpw@127.0.0.1:1/demo?sslmode=disable",
			secrets: []string{"bad%zzpw"},
		},
		{
			// pq reports the part after the space as a keyword missing its
			// "=", quoting it.
			name:    "key value with unquoted password containing a space",
			dsn:     "host=127.0.0.1 port=1 user=demo password=dummy secret-PW sslmode=disable",
			secrets: []string{"dummy", "secret-PW"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			p, err := NewPostgresCollector(logger, nil, tt.dsn, nil)
			if err != nil {
				t.Fatalf("NewPostgresCollector() error = %v", err)
			}

			ch := make(chan prometheus.Metric)
			done := make(chan struct{})
			go func() {
				for range ch {
				}
				close(done)
			}()
			p.Collect(ch)
			close(ch)
			<-done

			logs := buf.String()
			if !strings.Contains(logs, "Error opening connection to database") {
				t.Fatalf("Collect() did not log the connection failure; logs:\n%s", logs)
			}
			for _, secret := range tt.secrets {
				if strings.Contains(logs, secret) {
					t.Errorf("Collect() logs contain %q:\n%s", secret, logs)
				}
			}
		})
	}
}
