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

func TestPGStatReplicationCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{
		"application_name",
		"client_addr",
		"state",
		"slot_name",
		"pg_current_wal_lsn_bytes",
		"pg_wal_lsn_diff",
	}).AddRow("standby", "10.0.0.1", "streaming", "slot_a", 1024.0, 64.0)
	mock.ExpectQuery(sanitizeQuery(statReplicationQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{
		"application_name": "standby",
		"client_addr":      "10.0.0.1",
		"state":            "streaming",
		"slot_name":        "slot_a",
	}
	expected := []MetricResult{
		{labels: expectedLabels, value: 1024, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 64, metricType: dto.MetricType_GAUGE},
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

func TestPGStatReplicationCollectorBefore10(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.6.0")}

	rows := sqlmock.NewRows([]string{
		"application_name",
		"client_addr",
		"state",
		"slot_name",
		"pg_xlog_location_diff",
	}).AddRow("standby", "10.0.0.1", "streaming", "slot_a", 32.0)
	mock.ExpectQuery(sanitizeQuery(statReplicationQueryBefore10)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{
			labels: labelMap{
				"application_name": "standby",
				"client_addr":      "10.0.0.1",
				"state":            "streaming",
				"slot_name":        "slot_a",
			},
			value:      32,
			metricType: dto.MetricType_GAUGE,
		},
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

func TestPGStatReplicationCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{
		"application_name",
		"client_addr",
		"state",
		"slot_name",
		"pg_current_wal_lsn_bytes",
		"pg_wal_lsn_diff",
	}).AddRow(nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statReplicationQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{
		"application_name": "",
		"client_addr":      "",
		"state":            "",
		"slot_name":        "",
	}
	expected := []MetricResult{
		{labels: expectedLabels, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 0, metricType: dto.MetricType_GAUGE},
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

func TestPGStatReplicationCollectorBefore92(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.1.0")}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for PostgreSQL 9.1: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
