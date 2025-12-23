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
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatStatementsCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery, "", "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
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

func TestPGStatStatementsCollectorWithStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 100) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery, fmt.Sprintf(pgStatStatementQuerySelect, 100), "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 100}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollectorWithStatementAndLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 100) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery, fmt.Sprintf(pgStatStatementQuerySelect, 100), "", "10"))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 100, statementLimit: 10}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollectorNull(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, "", "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
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

func TestPGStatStatementsCollectorNullWithStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 200) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, fmt.Sprintf(pgStatStatementQuerySelect, 200), "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 200}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"queryid": "unknown", "query": "unknown"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollectorNullWithStatementAndLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 200) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, fmt.Sprintf(pgStatStatementQuerySelect, 200), "", "10"))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 200, statementLimit: 10}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"user": "unknown", "datname": "unknown", "queryid": "unknown"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"queryid": "unknown", "query": "unknown"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollector_PG13(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, "", "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
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

func TestPGStatStatementsCollector_PG13_WithStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 300) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, fmt.Sprintf(pgStatStatementQuerySelect, 300), "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 300}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollector_PG13_WithStatementAndLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 300) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, fmt.Sprintf(pgStatStatementQuerySelect, 300), "", "10"))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 300, statementLimit: 10}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollector_PG17(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG17, "", "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
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

func TestPGStatStatementsCollector_PG17_WithStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 300) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG17, fmt.Sprintf(pgStatStatementQuerySelect, 300), "", defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 300}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollector_PG17_WithStatementAndLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	columns := []string{"user", "datname", "queryid", "LEFT(pg_stat_statements.query, 300) as query", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, "select 1 from foo", 5, 0.4, 100, 0.1, 0.2)
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG17, fmt.Sprintf(pgStatStatementQuerySelect, 300), "", "10"))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{includeQueryStatement: true, statementLength: 300, statementLimit: 10}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
		{labels: labelMap{"queryid": "1500", "query": "select 1 from foo"}, metricType: dto.MetricType_COUNTER, value: 1},
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

func TestPGStatStatementsCollectorWithExcludedDatabases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	excludedDatabases := []string{"rdsadmin", "cloudsqladmin"}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)

	// Expected query should include database exclusion filters
	expectedFilter := " AND pg_database.datname NOT IN ('rdsadmin', 'cloudsqladmin')"
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery, "", expectedFilter, defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{excludedDatabases: excludedDatabases}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
	}

	convey.Convey("Metrics comparison with excluded databases", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatStatementsCollectorWithExcludedDatabasesSQLEscaping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.0.0")}

	// exclude a database name with a single quote (SQL injection test)
	excludedDatabases := []string{"test'db"}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)

	expectedFilter := " AND pg_database.datname NOT IN ('test''db')"
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, "", expectedFilter, defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{excludedDatabases: excludedDatabases}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
	}

	convey.Convey("Metrics comparison with SQL escaping", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatStatementsCollectorWithExcludedUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	excludedUsers := []string{"monitoring", "readonly"}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)

	expectedFilter := " AND pg_get_userbyid(userid) NOT IN ('monitoring', 'readonly')"
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery, "", expectedFilter, defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{excludedUsers: excludedUsers}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
	}

	convey.Convey("Metrics comparison with excluded users", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatStatementsCollectorWithExcludedUsersSQLEscaping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.0.0")}

	// exclude a user name with a single quote (SQL injection test)
	excludedUsers := []string{"test'user"}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)

	expectedFilter := " AND pg_get_userbyid(userid) NOT IN ('test''user')"
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG13, "", expectedFilter, defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{excludedUsers: excludedUsers}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
	}

	convey.Convey("Metrics comparison with SQL escaping for users", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatStatementsCollectorWithExcludedDatabasesAndUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	// exclude both databases and users
	excludedDatabases := []string{"rdsadmin", "cloudsqladmin"}
	excludedUsers := []string{"monitoring", "readonly"}

	columns := []string{"user", "datname", "queryid", "calls_total", "seconds_total", "rows_total", "block_read_seconds_total", "block_write_seconds_total"}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "postgres", 1500, 5, 0.4, 100, 0.1, 0.2)

	expectedFilter := " AND pg_database.datname NOT IN ('rdsadmin', 'cloudsqladmin') AND pg_get_userbyid(userid) NOT IN ('monitoring', 'readonly')"
	mock.ExpectQuery(sanitizeQuery(fmt.Sprintf(pgStatStatementsQuery_PG17, "", expectedFilter, defaultStatementLimit))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatStatementsCollector{excludedDatabases: excludedDatabases, excludedUsers: excludedUsers}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.4},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 100},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.1},
		{labels: labelMap{"user": "postgres", "datname": "postgres", "queryid": "1500"}, metricType: dto.MetricType_COUNTER, value: 0.2},
	}

	convey.Convey("Metrics comparison with excluded databases and users", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
