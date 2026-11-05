// Copyright 2025 The Prometheus Authors
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
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatProgressVacuumCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datname", "relname", "phase", "heap_blks_total", "heap_blks_scanned",
		"heap_blks_vacuumed", "index_vacuum_count", "param6", "param7", "param8",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		"postgres", "a_table", 3, 3000, 400, 200, 2, 500, 123, 0)

	mock.ExpectQuery(sanitizeQuery(statProgressVacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatProgressVacuumCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatProgressVacuumCollector.Update; %+v", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "initializing"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "scanning heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "vacuuming indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "vacuuming heap"}, metricType: dto.MetricType_GAUGE, value: 1},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "cleaning up indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "truncating heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table", "phase": "performing final cleanup"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 3000},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 400},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 200},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 2},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 500},
		{labels: labelMap{"datname": "postgres", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 123},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(m, convey.ShouldResemble, expect)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled exceptions: %+v", err)
	}
}

func TestPGStatProgressVacuumCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datname", "relname", "phase", "heap_blks_total", "heap_blks_scanned",
		"heap_blks_vacuumed", "index_vacuum_count", "param6", "param7", "param8",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		"postgres", nil, nil, nil, nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(sanitizeQuery(statProgressVacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatProgressVacuumCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatProgressVacuumCollector.Update; %+v", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "initializing"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "scanning heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "vacuuming indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "vacuuming heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "cleaning up indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "truncating heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown", "phase": "performing final cleanup"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled exceptions: %+v", err)
	}
}

func TestPGStatProgressVacuumCollectorPG17(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.6.0")}

	columns := []string{
		"datname", "relname", "phase", "heap_blks_total", "heap_blks_scanned",
		"heap_blks_vacuumed", "index_vacuum_count", "param6", "param7", "param8",
	}

	// On PostgreSQL 17 param6 is max_dead_tuple_bytes, param7 is
	// dead_tuple_bytes and param8 is num_dead_item_ids.
	rows := sqlmock.NewRows(columns).AddRow(
		"postgres", "a_table", 1, 3000, 400, 0, 0, 67108864, 8192, 123)

	mock.ExpectQuery(sanitizeQuery(statProgressVacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatProgressVacuumCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatProgressVacuumCollector.Update; %+v", err)
		}
	}()

	got := map[string]float64{}
	for m := range ch {
		got[m.Desc().String()] = readMetric(m).value
	}

	if _, ok := got[statProgressVacuumMaxDeadTuples.String()]; ok {
		t.Errorf("max_dead_tuples must not be reported on PostgreSQL 17")
	}
	for desc, want := range map[*prometheus.Desc]float64{
		statProgressVacuumMaxDeadTupleBytes: 67108864,
		statProgressVacuumDeadTupleBytes:    8192,
		statProgressVacuumNumDeadTuples:     123,
	} {
		if v, ok := got[desc.String()]; !ok || v != want {
			t.Errorf("%s = %v (reported: %v), want %v", desc, v, ok, want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled exceptions: %+v", err)
	}
}

// TestPGStatProgressVacuumQueryCrossDatabaseFallback verifies that the query uses
// current_database() to gate schema-qualified name resolution, ensuring cross-database
// vacuums fall back to raw OID text. The actual SQL logic cannot be exercised in unit
// tests without a real Postgres connection.
func TestPGStatProgressVacuumQueryCrossDatabaseFallback(t *testing.T) {
	if !strings.Contains(statProgressVacuumQuery, "current_database()") {
		t.Error("statProgressVacuumQuery must use current_database() to avoid resolving cross-database OIDs via pg_class")
	}
}

func TestPGStatProgressVacuumCollectorOIDRelname(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datname", "relname", "phase", "heap_blks_total", "heap_blks_scanned",
		"heap_blks_vacuumed", "index_vacuum_count", "param6", "param7", "param8",
	}

	// Two concurrent vacuums returning OID-based relnames (as the query produces for
	// cross-database vacuums). Validates that the collector emits unique label sets.
	rows := sqlmock.NewRows(columns).
		AddRow("mydb", "16385", 1, 5000, 200, 100, 0, 1000, 50, 0).
		AddRow("mydb", "16390", 3, 8000, 600, 400, 1, 2000, 300, 0)

	mock.ExpectQuery(sanitizeQuery(statProgressVacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatProgressVacuumCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatProgressVacuumCollector.Update; %+v", err)
		}
	}()

	expected := []MetricResult{
		// First vacuum (relid 16385, phase 1 = scanning heap)
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "initializing"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "scanning heap"}, metricType: dto.MetricType_GAUGE, value: 1},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "vacuuming indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "vacuuming heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "cleaning up indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "truncating heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385", "phase": "performing final cleanup"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 5000},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 200},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 100},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 1000},
		{labels: labelMap{"datname": "mydb", "relname": "16385"}, metricType: dto.MetricType_GAUGE, value: 50},
		// Second vacuum (relid 16390, phase 3 = vacuuming heap)
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "initializing"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "scanning heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "vacuuming indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "vacuuming heap"}, metricType: dto.MetricType_GAUGE, value: 1},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "cleaning up indexes"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "truncating heap"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390", "phase": "performing final cleanup"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 8000},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 600},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 400},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 1},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 2000},
		{labels: labelMap{"datname": "mydb", "relname": "16390"}, metricType: dto.MetricType_GAUGE, value: 300},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(m, convey.ShouldResemble, expect)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled exceptions: %+v", err)
	}
}
