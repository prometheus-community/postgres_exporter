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
	"maps"
	"os"
	"slices"
	"testing"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestPGStatActivityAutovacuumCollectorSplitsSchemaAndRelation(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	var datname string
	if err := db.QueryRow(`SELECT current_database()`).Scan(&datname); err != nil {
		t.Fatalf("querying current_database(): %v", err)
	}

	// pg_stat_activity keeps the text of the last statement of a session even
	// when it fails to parse, so a transaction with a failed statement can
	// stand in for a running autovacuum worker. The savepoint keeps xact_start
	// set after the error.
	for _, query := range []string{
		"autovacuum: VACUUM ANALYZE mob.datacdr",
		"autovacuum: VACUUM postpaid.datacdr",
	} {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("Conn() error = %v", err)
		}
		defer conn.Close()
		for _, stmt := range []string{"BEGIN", "SAVEPOINT s"} {
			if _, err := conn.ExecContext(ctx, stmt); err != nil {
				t.Fatalf("%s error = %v", stmt, err)
			}
		}
		conn.ExecContext(ctx, query) // fails to parse, only its text is recorded
	}

	ch := make(chan prometheus.Metric, 100)
	c := PGStatActivityAutovacuumCollector{log: promslog.NewNopLogger()}
	if err := c.Update(ctx, &instance{db: db}, ch); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	close(ch)

	var got []labelMap
	for m := range ch {
		got = append(got, readMetric(m).labels)
	}
	for _, schemaname := range []string{"mob", "postpaid"} {
		want := labelMap{"datname": datname, "schemaname": schemaname, "relname": "datacdr"}
		if !slices.ContainsFunc(got, func(l labelMap) bool { return maps.Equal(l, want) }) {
			t.Errorf("no series with labels %v in %v", want, got)
		}
	}
}
