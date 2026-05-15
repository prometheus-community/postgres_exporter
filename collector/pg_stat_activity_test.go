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

package collector

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatActivityCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{
		"datname",
		"state",
		"usename",
		"application_name",
		"backend_type",
		"wait_event_type",
		"wait_event",
		"count",
		"max_tx_duration",
	}).AddRow("postgres", "active", "postgres", "psql", "client backend", "Lock", "relation", 3, 12.5)
	mock.ExpectQuery(sanitizeQuery(statActivityQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatActivityCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatActivityCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{
		"datname":          "postgres",
		"state":            "active",
		"usename":          "postgres",
		"application_name": "psql",
		"backend_type":     "client backend",
		"wait_event_type":  "Lock",
		"wait_event":       "relation",
	}
	expected := []MetricResult{
		{labels: expectedLabels, value: 3, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 12.5, metricType: dto.MetricType_GAUGE},
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

func TestPGStatActivityCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{
		"datname",
		"state",
		"usename",
		"application_name",
		"backend_type",
		"wait_event_type",
		"wait_event",
		"count",
		"max_tx_duration",
	}).AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statActivityQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatActivityCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatActivityCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL stat_activity value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatActivityCollectorBefore92(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.1.0")}

	rows := sqlmock.NewRows([]string{
		"datname",
		"state",
		"usename",
		"application_name",
		"backend_type",
		"wait_event_type",
		"wait_event",
		"count",
		"max_tx_duration",
	}).AddRow("postgres", "unknown", "postgres", "psql", "", "", "", 2, 7.25)
	mock.ExpectQuery(sanitizeQuery(statActivityQueryBefore92)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatActivityCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatActivityCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{
		"datname":          "postgres",
		"state":            "unknown",
		"usename":          "postgres",
		"application_name": "psql",
		"backend_type":     "",
		"wait_event_type":  "",
		"wait_event":       "",
	}
	expected := []MetricResult{
		{labels: expectedLabels, value: 2, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 7.25, metricType: dto.MetricType_GAUGE},
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
