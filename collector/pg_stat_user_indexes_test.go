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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgStatUserIndexesCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}
	columns := []string{
		"schemaname",
		"relname",
		"indexrelname",
		"idx_scan",
		"idx_tup_read",
		"idx_tup_fetch",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("public", "pgbench_accounts", "pgbench_accounts_pkey", 5, 6, 7)

	mock.ExpectQuery(sanitizeQuery(statUserIndexesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatUserIndexesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatUserIndexesCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"schemaname": "public", "relname": "pgbench_accounts", "indexrelname": "pgbench_accounts_pkey"}, value: 5, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"schemaname": "public", "relname": "pgbench_accounts", "indexrelname": "pgbench_accounts_pkey"}, value: 6, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"schemaname": "public", "relname": "pgbench_accounts", "indexrelname": "pgbench_accounts_pkey"}, value: 7, metricType: dto.MetricType_COUNTER},
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

func TestPgStatUserIndexesCollectorNull(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}
	columns := []string{
		"schemaname",
		"relname",
		"indexrelname",
		"idx_scan",
		"idx_tup_read",
		"idx_tup_fetch",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(sanitizeQuery(statUserIndexesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatUserIndexesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatUserIndexesCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"schemaname": "unknown", "relname": "unknown", "indexrelname": "unknown"}, value: 0, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"schemaname": "unknown", "relname": "unknown", "indexrelname": "unknown"}, value: 0, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"schemaname": "unknown", "relname": "unknown", "indexrelname": "unknown"}, value: 0, metricType: dto.MetricType_COUNTER},
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
