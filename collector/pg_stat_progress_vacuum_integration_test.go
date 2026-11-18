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
	"os"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

// TestPGStatProgressVacuumParamMapping checks that the progress parameters the
// collector reads mean what it reports for the server version under test.
func TestPGStatProgressVacuumParamMapping(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	version, err := queryVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("queryVersion() error = %v", err)
	}

	var viewDef string
	if err := db.QueryRow(`SELECT pg_get_viewdef('pg_catalog.pg_stat_progress_vacuum'::regclass)`).Scan(&viewDef); err != nil {
		t.Fatalf("pg_get_viewdef() error = %v", err)
	}

	want := []string{"param6 AS max_dead_tuples", "param7 AS num_dead_tuples"}
	if version.GTE(semver.MustParse("17.0.0")) {
		want = []string{"param6 AS max_dead_tuple_bytes", "param7 AS dead_tuple_bytes", "param8 AS num_dead_item_ids"}
	}
	for _, w := range want {
		if !strings.Contains(viewDef, w) {
			t.Errorf("pg_stat_progress_vacuum on PostgreSQL %s does not contain %q:\n%s", version, w, viewDef)
		}
	}

	inst := &instance{db: db, version: version}
	ch := make(chan prometheus.Metric)
	errCh := make(chan error, 1)
	go func() {
		errCh <- (&PGStatProgressVacuumCollector{}).Update(context.Background(), inst, ch)
		close(ch)
	}()
	for range ch {
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}
