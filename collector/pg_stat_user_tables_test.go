// Copyright 2023 The Prometheus Authors
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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatUserTablesCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	lastVacuumTime, err := time.Parse("2006-01-02Z", "2023-06-02Z")
	if err != nil {
		t.Fatalf("Error parsing vacuum time: %s", err)
	}
	lastAutoVacuumTime, err := time.Parse("2006-01-02Z", "2023-06-03Z")
	if err != nil {
		t.Fatalf("Error parsing vacuum time: %s", err)
	}
	lastAnalyzeTime, err := time.Parse("2006-01-02Z", "2023-06-04Z")
	if err != nil {
		t.Fatalf("Error parsing vacuum time: %s", err)
	}
	lastAutoAnalyzeTime, err := time.Parse("2006-01-02Z", "2023-06-05Z")
	if err != nil {
		t.Fatalf("Error parsing vacuum time: %s", err)
	}

	columns := []string{
		"datname",
		"schemaname",
		"relname",
		"seq_scan",
		"seq_tup_read",
		"idx_scan",
		"idx_tup_fetch",
		"n_tup_ins",
		"n_tup_upd",
		"n_tup_del",
		"n_tup_hot_upd",
		"n_live_tup",
		"n_dead_tup",
		"n_mod_since_analyze",
		"last_vacuum",
		"last_autovacuum",
		"last_analyze",
		"last_autoanalyze",
		"vacuum_count",
		"autovacuum_count",
		"analyze_count",
		"autoanalyze_count",
		"index_size",
		"table_size"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres",
			"public",
			"a_table",
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			0,
			lastVacuumTime,
			lastAutoVacuumTime,
			lastAnalyzeTime,
			lastAutoAnalyzeTime,
			11,
			12,
			13,
			14,
			15,
			16)
	mock.ExpectQuery(sanitizeQuery(statUserTablesQuery)).WillReturnRows(rows)
	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatUserTablesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatUserTablesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 1},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 2},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 3},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 4},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 6},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 7},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 8},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 9},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 10},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 1685664000},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 1685750400},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 1685836800},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 1685923200},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 11},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 12},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 13},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_COUNTER, value: 14},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 15},
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "a_table"}, metricType: dto.MetricType_GAUGE, value: 16},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatUserTablesCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datname",
		"schemaname",
		"relname",
		"seq_scan",
		"seq_tup_read",
		"idx_scan",
		"idx_tup_fetch",
		"n_tup_ins",
		"n_tup_upd",
		"n_tup_del",
		"n_tup_hot_upd",
		"n_live_tup",
		"n_dead_tup",
		"n_mod_since_analyze",
		"last_vacuum",
		"last_autovacuum",
		"last_analyze",
		"last_autoanalyze",
		"vacuum_count",
		"autovacuum_count",
		"analyze_count",
		"autoanalyze_count",
		"index_size",
		"table_size"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil)
	mock.ExpectQuery(sanitizeQuery(statUserTablesQuery)).WillReturnRows(rows)
	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatUserTablesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatUserTablesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
		{labels: labelMap{"datname": "postgres", "schemaname": "unknown", "relname": "unknown"}, metricType: dto.MetricType_GAUGE, value: 0},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
