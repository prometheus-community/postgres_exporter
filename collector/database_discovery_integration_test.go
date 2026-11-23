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
	"bytes"
	"database/sql"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// TestPostgresCollectorRunsDatabaseScopedCollectorsAgainstDiscoveredDatabases
// is an end-to-end regression test for the bug this package's
// WithDatabaseDiscovery option fixes: database-scoped collectors, such as
// stat_user_tables, must run against every database on the server, not just
// the one DataSourceName points at.
func TestPostgresCollectorRunsDatabaseScopedCollectorsAgainstDiscoveredDatabases(t *testing.T) {
	primaryDSN := os.Getenv("DATA_SOURCE_NAME")
	if primaryDSN == "" {
		t.Skip("DATA_SOURCE_NAME not set")
	}

	admin, err := sql.Open("postgres", primaryDSN)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer admin.Close()

	const discoveredDB = "pe_discovery_test"
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + discoveredDB); err != nil {
		t.Fatalf("DROP DATABASE error = %v", err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + discoveredDB); err != nil {
		t.Fatalf("CREATE DATABASE error = %v", err)
	}
	defer admin.Exec("DROP DATABASE IF EXISTS " + discoveredDB)

	discoveredConnStr, err := (&instance{dsn: primaryDSN}).withDatabase(discoveredDB)
	if err != nil {
		t.Fatalf("withDatabase() error = %v", err)
	}
	discovered, err := sql.Open("postgres", discoveredConnStr.dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer discovered.Close()
	// pg_stat_user_tables/pg_statio_user_indexes rows exist as soon as the
	// relation does, initialized to zero, so no DML is needed to populate them.
	if _, err := discovered.Exec("CREATE TABLE t (id int primary key)"); err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	c, err := NewPostgresCollector(
		promslog.NewNopLogger(),
		nil,
		primaryDSN,
		[]string{userTableSubsystem},
		WithDatabaseDiscovery(nil, nil, 0),
	)
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}
	defer c.Close()

	// Registering with a real Registry and calling Gather (rather than just
	// draining the Collect channel) is what actually exercises Prometheus's
	// duplicate-metric detection: running the same collector name against two
	// database connections in one scrape must not collide, for either the
	// stat_user_tables data itself or the scrape_collector_duration/success
	// meta-metrics execute() emits for every collector run.
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	sawDiscoveredDatabase := false
	for _, mf := range mfs {
		if mf.GetName() != "pg_stat_user_tables_seq_scan" {
			continue
		}
		for _, m := range mf.GetMetric() {
			labels := make(map[string]string, len(m.GetLabel()))
			for _, l := range m.GetLabel() {
				labels[l.GetName()] = l.GetValue()
			}
			if labels["datname"] == discoveredDB && labels["relname"] == "t" {
				sawDiscoveredDatabase = true
			}
		}
	}
	if !sawDiscoveredDatabase {
		t.Fatalf("no stat_user_tables metric observed for table %q in discovered database %q", "t", discoveredDB)
	}
}

func TestPostgresCollectorSkipsDiscoveryWithoutDatabaseScopedCollectors(t *testing.T) {
	primaryDSN := os.Getenv("DATA_SOURCE_NAME")
	if primaryDSN == "" {
		t.Skip("DATA_SOURCE_NAME not set")
	}
	primaryURL, err := url.Parse(primaryDSN)
	if err != nil || primaryURL.Scheme == "" {
		t.Skip("DATA_SOURCE_NAME is not a URL")
	}

	admin, err := sql.Open("postgres", primaryDSN)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer admin.Close()

	var lockedDB string
	err = admin.QueryRow(`SELECT datname FROM pg_database WHERE datallowconn AND NOT datistemplate AND datname <> current_database() ORDER BY datname LIMIT 1`).Scan(&lockedDB)
	if err == sql.ErrNoRows {
		t.Skip("no other database to discover")
	}
	if err != nil {
		t.Fatalf("find database to discover: %v", err)
	}

	const (
		role     = "pe_discovery_user"
		password = "pe_discovery_password"
	)
	quotedDB := pq.QuoteIdentifier(lockedDB)
	admin.Exec("DROP ROLE IF EXISTS " + role)
	defer func() {
		admin.Exec("GRANT CONNECT ON DATABASE " + quotedDB + " TO PUBLIC")
		admin.Exec("DROP ROLE IF EXISTS " + role)
	}()
	for _, stmt := range []string{
		"CREATE ROLE " + role + " LOGIN PASSWORD '" + password + "'",
		"REVOKE CONNECT ON DATABASE " + quotedDB + " FROM PUBLIC",
	} {
		if _, err := admin.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}

	roleURL := *primaryURL
	roleURL.User = url.UserPassword(role, password)

	tests := []struct {
		name          string
		collectors    []string
		wantDiscovery bool
	}{
		{name: "server-scoped only", collectors: []string{settingsSubsystem}, wantDiscovery: false},
		{name: "database-scoped", collectors: []string{settingsSubsystem, userTableSubsystem}, wantDiscovery: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))

			c, err := NewPostgresCollector(logger, nil, roleURL.String(), test.collectors, WithDatabaseDiscovery(nil, nil, 0))
			if err != nil {
				t.Fatalf("NewPostgresCollector() error = %v", err)
			}
			defer c.Close()

			reg := prometheus.NewRegistry()
			if err := reg.Register(c); err != nil {
				t.Fatalf("Register() error = %v", err)
			}
			if _, err := reg.Gather(); err != nil {
				t.Fatalf("Gather() error = %v", err)
			}

			tried := strings.Contains(logs.String(), "failed to connect to discovered database") &&
				strings.Contains(logs.String(), lockedDB)
			if tried != test.wantDiscovery {
				t.Fatalf("connected to discovered database %q = %t, want %t; logs:\n%s", lockedDB, tried, test.wantDiscovery, logs.String())
			}
		})
	}
}
