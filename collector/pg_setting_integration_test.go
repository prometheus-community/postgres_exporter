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

//go:build integration

package collector

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestPGSettingsCollectorPendingRestart(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	version, err := queryVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("queryVersion() error = %v", err)
	}

	// max_connections requires a server restart to take effect.
	var current int
	if err := db.QueryRow(`SELECT setting::int FROM pg_settings WHERE name = 'max_connections'`).Scan(&current); err != nil {
		t.Fatalf("querying max_connections: %v", err)
	}

	// ALTER SYSTEM does not support parameter placeholders.
	if _, err := db.Exec(fmt.Sprintf("ALTER SYSTEM SET max_connections = %d", current+1)); err != nil {
		t.Fatalf("ALTER SYSTEM SET max_connections: %v", err)
	}

	// Reset the setting and reload the configuration after the test.
	defer func() {
		if _, err := db.Exec(`ALTER SYSTEM RESET max_connections`); err != nil {
			t.Logf("ALTER SYSTEM RESET max_connections: %v", err)
		}
		var reloaded bool
		if err := db.QueryRow(`SELECT pg_reload_conf()`).Scan(&reloaded); err != nil {
			t.Logf("pg_reload_conf() after reset: %v", err)
		}
	}()

	// pending_restart is updated after a configuration reload.
	var reloaded bool
	if err := db.QueryRow(`SELECT pg_reload_conf()`).Scan(&reloaded); err != nil {
		t.Fatalf("pg_reload_conf(): %v", err)
	}
	if !waitForPendingRestart(t, db) {
		t.Fatalf("max_connections was not marked as pending a restart")
	}

	inst := &instance{db: db, version: version}
	collector := PGSettingsCollector{log: promslog.NewNopLogger()}

	ch := make(chan prometheus.Metric)
	errCh := make(chan error, 1)
	go func() {
		errCh <- collector.Update(context.Background(), inst, ch)
		close(ch)
	}()

	foundPending, foundValue := false, false
	for metric := range ch {
		got := readMetric(metric)
		switch metric.Desc().String() {
		case pgSettingsPendingRestartDesc.String():
			if got.labels["name"] != "max_connections" {
				continue
			}
			if got.value != 1 {
				t.Fatalf("pending restart value = %v, want 1", got.value)
			}
			foundPending = true
		case `Desc{fqName: "pg_settings_max_connections", help: "Server Parameter: max_connections", unit: "", constLabels: {}, variableLabels: {}}`:
			// The value metric for the same setting must come from the very
			// same query, since pending_restart is part of it.
			foundValue = true
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !foundPending {
		t.Errorf(`pg_settings_pending_restart{name="max_connections"} was not reported`)
	}
	if !foundValue {
		t.Errorf("pg_settings_max_connections was not reported in the same scrape")
	}
}

// waitForPendingRestart polls pg_settings until max_connections reports a
// pending restart, returning false if it does not within the timeout.
func waitForPendingRestart(t *testing.T, db *sql.DB) bool {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var pending bool
		if err := db.QueryRow(`SELECT pending_restart FROM pg_settings WHERE name = 'max_connections'`).Scan(&pending); err != nil {
			t.Fatalf("querying pending_restart: %v", err)
		}
		if pending {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
