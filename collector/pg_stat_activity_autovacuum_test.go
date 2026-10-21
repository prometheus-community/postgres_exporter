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
	"github.com/prometheus/common/promslog"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatActivityAutovacuumCollector(t *testing.T) {
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
		"timestamp_seconds",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("app", "mob", "datacdr", 3600).
		AddRow("app", "postpaid", "datacdr", 7200).
		AddRow("backoffice", "mob", "datacdr", 10800)

	mock.ExpectQuery(sanitizeQuery(statActivityAutovacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatActivityAutovacuumCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatActivityAutovacuumCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"datname": "app", "schemaname": "mob", "relname": "datacdr"}, value: 3600, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"datname": "app", "schemaname": "postpaid", "relname": "datacdr"}, value: 7200, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"datname": "backoffice", "schemaname": "mob", "relname": "datacdr"}, value: 10800, metricType: dto.MetricType_GAUGE},
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

func TestPGStatActivityAutovacuumCollectorNullTimestamp(t *testing.T) {
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
		"timestamp_seconds",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("postgres", "public", "no_xact_start", nil).
		AddRow("postgres", "public", "test", 3600)

	mock.ExpectQuery(sanitizeQuery(statActivityAutovacuumQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatActivityAutovacuumCollector{log: promslog.NewNopLogger()}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatActivityAutovacuumCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres", "schemaname": "public", "relname": "test"}, value: 3600, metricType: dto.MetricType_GAUGE},
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
