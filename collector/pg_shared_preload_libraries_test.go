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

func TestPGSharedPreloadLibrariesCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub database connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"setting"}
	rows := sqlmock.NewRows(columns).
		AddRow("pg_stat_statements, auto_explain, pg_hint_plan")

	mock.ExpectQuery(sanitizeQuery(pgSharedPreloadLibrariesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSharedPreloadLibrariesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGSharedPreloadLibrariesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		// Emitted in sorted order: auto_explain, pg_hint_plan, pg_stat_statements
		{labels: labelMap{"library": "auto_explain"}, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"library": "pg_hint_plan"}, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"library": "pg_stat_statements"}, value: 1, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGSharedPreloadLibrariesCollectorEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub database connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"setting"}
	rows := sqlmock.NewRows(columns).
		AddRow("")

	mock.ExpectQuery(sanitizeQuery(pgSharedPreloadLibrariesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSharedPreloadLibrariesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGSharedPreloadLibrariesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGSharedPreloadLibrariesCollectorSingle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub database connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"setting"}
	rows := sqlmock.NewRows(columns).
		AddRow("pg_stat_statements")

	mock.ExpectQuery(sanitizeQuery(pgSharedPreloadLibrariesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSharedPreloadLibrariesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGSharedPreloadLibrariesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"library": "pg_stat_statements"}, value: 1, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGSharedPreloadLibrariesCollectorWhitespaceAndDuplicates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub database connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"setting"}
	rows := sqlmock.NewRows(columns).
		AddRow("pg_stat_statements,  auto_explain,  pg_hint_plan , auto_explain , pg_stat_statements ")

	mock.ExpectQuery(sanitizeQuery(pgSharedPreloadLibrariesQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSharedPreloadLibrariesCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGSharedPreloadLibrariesCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"library": "auto_explain"}, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"library": "pg_hint_plan"}, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"library": "pg_stat_statements"}, value: 1, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
