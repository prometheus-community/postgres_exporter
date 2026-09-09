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
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// TestPostgresCollectorRunsDatabaseScopedCollectorsAgainstDiscoveredDatabases
// is an end-to-end regression test for the bug this package's
// WithDatabaseDiscovery option fixes: database-scoped collectors, such as
// stat_user_tables, must run against every database on the server, not just
// the one DataSourceNames[0] points at.
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
		WithDatabaseDiscovery(nil, nil),
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
