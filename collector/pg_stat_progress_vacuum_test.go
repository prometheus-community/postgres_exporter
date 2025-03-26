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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
		"heap_blks_vacuumed", "index_vacuum_count", "max_dead_tuples", "num_dead_tuples",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		"postgres", "a_table", 3, 3000, 400, 200, 2, 500, 123)

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
		"heap_blks_vacuumed", "index_vacuum_count", "max_dead_tuples", "num_dead_tuples",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		"postgres", nil, nil, nil, nil, nil, nil, nil, nil)

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
