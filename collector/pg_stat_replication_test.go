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
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
		"pid",
		"pg_current_wal_lsn_bytes",
		"pg_wal_lsn_diff",
	}).AddRow("standby", "10.0.0.1", "streaming", "slot_a", 123, 1024.0, 64.0)
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
		"pid":              "123",
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
		"pid",
		"pg_xlog_location_diff",
	}).AddRow("standby", "10.0.0.1", "streaming", "slot_a", 123, 32.0)
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
				"pid":              "123",
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

func TestPGStatReplicationCollectorBefore95(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.4.0")}

	rows := sqlmock.NewRows([]string{
		"application_name",
		"client_addr",
		"state",
		"pid",
		"pg_xlog_location_diff",
	}).AddRow("standby", "10.0.0.1", "streaming", 123, 32.0)
	mock.ExpectQuery(sanitizeQuery(statReplicationQueryBefore95)).WillReturnRows(rows)

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
				"slot_name":        "",
				"pid":              "123",
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
		"pid",
		"pg_current_wal_lsn_bytes",
		"pg_wal_lsn_diff",
	}).AddRow(nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statReplicationQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL stat_replication value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatReplicationCollectorBefore10NullValues(t *testing.T) {
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
		"pid",
		"pg_xlog_location_diff",
	}).AddRow(nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statReplicationQueryBefore10)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL stat_replication value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

// collectFunc adapts a plain function to prometheus.Collector so it can be
// exercised through testutil.CollectAndCompare.
type collectFunc func(ch chan<- prometheus.Metric)

func (f collectFunc) Describe(chan<- *prometheus.Desc)    {}
func (f collectFunc) Collect(ch chan<- prometheus.Metric) { f(ch) }

// TestPGStatReplicationCollectorDuplicateConnections is a regression test
// for https://github.com/prometheus-community/postgres_exporter/issues/1352.
//
// application_name, client_addr, state and slot_name are not a unique key
// for pg_stat_replication. Only pid actually distinguishes the rows.
func TestPGStatReplicationCollectorDuplicateConnections(t *testing.T) {
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
		"pid",
		"pg_current_wal_lsn_bytes",
		"pg_wal_lsn_diff",
	}).
		AddRow("walreceiver", "172.18.0.254", "streaming", "", 111, 1024.0, 64.0).
		AddRow("walreceiver", "172.18.0.254", "streaming", "", 222, 2048.0, 128.0)
	mock.ExpectQuery(sanitizeQuery(statReplicationQuery)).WillReturnRows(rows)

	collector := collectFunc(func(ch chan<- prometheus.Metric) {
		c := PGStatReplicationCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatReplicationCollector.Update: %s", err)
		}
	})

	expected := strings.NewReader(`
# HELP pg_stat_replication_pg_current_wal_lsn_bytes WAL position in bytes
# TYPE pg_stat_replication_pg_current_wal_lsn_bytes gauge
pg_stat_replication_pg_current_wal_lsn_bytes{application_name="walreceiver",client_addr="172.18.0.254",pid="111",slot_name="",state="streaming"} 1024
pg_stat_replication_pg_current_wal_lsn_bytes{application_name="walreceiver",client_addr="172.18.0.254",pid="222",slot_name="",state="streaming"} 2048
# HELP pg_stat_replication_pg_wal_lsn_diff Lag in bytes between master and slave
# TYPE pg_stat_replication_pg_wal_lsn_diff gauge
pg_stat_replication_pg_wal_lsn_diff{application_name="walreceiver",client_addr="172.18.0.254",pid="111",slot_name="",state="streaming"} 64
pg_stat_replication_pg_wal_lsn_diff{application_name="walreceiver",client_addr="172.18.0.254",pid="222",slot_name="",state="streaming"} 128
`)
	if err := testutil.CollectAndCompare(collector, expected,
		"pg_stat_replication_pg_current_wal_lsn_bytes",
		"pg_stat_replication_pg_wal_lsn_diff",
	); err != nil {
		t.Errorf("unexpected collecting result for two connections with identical application_name/client_addr/state/slot_name:\n%s", err)
	}

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
