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
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGReplicationSlotsCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"}).
		AddRow("slot_a", "postgres", true, 64.0)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
	slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active", "safe_wal_size", "wal_status"}).
		AddRow("slot_a", "physical", 5, 3, true, 323906992, "reserved")
	mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotNewQuery)).WillReturnRows(slotRows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{"slot_name": "slot_a", "database": "postgres"}
	expectedSlotLabels := labelMap{"slot_name": "slot_a", "slot_type": "physical"}
	expected := []MetricResult{
		{labels: expectedLabels, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 64, metricType: dto.MetricType_GAUGE},
		{labels: expectedSlotLabels, value: 5, metricType: dto.MetricType_GAUGE},
		{labels: expectedSlotLabels, value: 3, metricType: dto.MetricType_GAUGE},
		{labels: expectedSlotLabels, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: expectedSlotLabels, value: 323906992, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "slot_a", "slot_type": "physical", "wal_status": "reserved"}, value: 1, metricType: dto.MetricType_GAUGE},
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

func TestPGReplicationSlotsCollectorBefore10(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.6.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_xlog_location_diff"}).
		AddRow("slot_a", "postgres", false, 32.0)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQueryBefore10)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	expectedLabels := labelMap{"slot_name": "slot_a", "database": "postgres"}
	expected := []MetricResult{
		{labels: expectedLabels, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: expectedLabels, value: 32, metricType: dto.MetricType_UNTYPED},
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

func TestPGReplicationSlotsCollectorSlotInactive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.3.7")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"})
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
	slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active", "safe_wal_size", "wal_status"}).
		AddRow("test_slot", "physical", 6, 12, false, -4000, "extended")
	mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotNewQuery)).WillReturnRows(slotRows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 6, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: -4000, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical", "wal_status": "extended"}, value: 1, metricType: dto.MetricType_GAUGE},
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

func TestPGReplicationSlotsCollectorSlotBefore13(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("12.0.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"})
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
	slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active"}).
		AddRow("test_slot", "logical", 10, 8, true)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotQuery)).WillReturnRows(slotRows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "logical"}, value: 10, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "logical"}, value: 8, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "logical"}, value: 1, metricType: dto.MetricType_GAUGE},
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

func TestPGReplicationSlotsCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"}).
		AddRow(nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
	slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active", "safe_wal_size", "wal_status"}).
		AddRow(nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotNewQuery)).WillReturnRows(slotRows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL replication_slots value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGReplicationSlotsCollectorBefore10NullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.6.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_xlog_location_diff"}).
		AddRow(nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQueryBefore10)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL replication_slots value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGReplicationSlotsCollectorBefore94(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.3.0")}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for PostgreSQL 9.3: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGReplicationSlotsCollectorPG17(t *testing.T) {
	for _, version := range []string{"17.0.0", "18.1.0"} {
		t.Run(version, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error opening a stub db connection: %s", err)
			}
			defer db.Close()

			inst := &instance{db: db, version: semver.MustParse(version)}

			rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"})
			mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
			slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active", "safe_wal_size", "wal_status", "conflicting", "inactive_since", "failover", "synced"}).
				AddRow("failover_slot", "logical", 200, 150, false, nil, "reserved", false, 1760697600.25, true, false).
				AddRow("synced_slot", "logical", 200, 120, false, nil, "reserved", false, nil, true, true).
				AddRow("conflicting_slot", "logical", 200, 110, false, nil, "lost", true, nil, true, true).
				AddRow("physical_slot", "physical", 200, 0, true, 4096, "reserved", nil, nil, false, false)
			mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotQueryPG17)).WillReturnRows(slotRows)

			ch := make(chan prometheus.Metric)
			go func() {
				defer close(ch)
				c := PGReplicationSlotsCollector{}
				if err := c.Update(context.Background(), inst, ch); err != nil {
					t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
				}
			}()

			failoverLabels := labelMap{"slot_name": "failover_slot", "slot_type": "logical"}
			syncedLabels := labelMap{"slot_name": "synced_slot", "slot_type": "logical"}
			conflictingLabels := labelMap{"slot_name": "conflicting_slot", "slot_type": "logical"}
			physicalLabels := labelMap{"slot_name": "physical_slot", "slot_type": "physical"}
			assertReplicationSlotsMetrics(t, ch, []replicationSlotsExpectedMetric{
				{replicationSlotsCurrentWalDesc, MetricResult{labels: failoverLabels, value: 200, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsIsActiveDesc, MetricResult{labels: failoverLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsWalStatus, MetricResult{labels: labelMap{"slot_name": "failover_slot", "slot_type": "logical", "wal_status": "reserved"}, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsConflictingDesc, MetricResult{labels: failoverLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsInactiveSinceDesc, MetricResult{labels: failoverLabels, value: 1760697600.25, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsFailoverDesc, MetricResult{labels: failoverLabels, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsSyncedDesc, MetricResult{labels: failoverLabels, value: 0, metricType: dto.MetricType_GAUGE}},

				{replicationSlotsCurrentWalDesc, MetricResult{labels: syncedLabels, value: 200, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsIsActiveDesc, MetricResult{labels: syncedLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsWalStatus, MetricResult{labels: labelMap{"slot_name": "synced_slot", "slot_type": "logical", "wal_status": "reserved"}, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsConflictingDesc, MetricResult{labels: syncedLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsFailoverDesc, MetricResult{labels: syncedLabels, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsSyncedDesc, MetricResult{labels: syncedLabels, value: 1, metricType: dto.MetricType_GAUGE}},

				{replicationSlotsCurrentWalDesc, MetricResult{labels: conflictingLabels, value: 200, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsIsActiveDesc, MetricResult{labels: conflictingLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsWalStatus, MetricResult{labels: labelMap{"slot_name": "conflicting_slot", "slot_type": "logical", "wal_status": "lost"}, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsConflictingDesc, MetricResult{labels: conflictingLabels, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsFailoverDesc, MetricResult{labels: conflictingLabels, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsSyncedDesc, MetricResult{labels: conflictingLabels, value: 1, metricType: dto.MetricType_GAUGE}},

				{replicationSlotsCurrentWalDesc, MetricResult{labels: physicalLabels, value: 200, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsCurrentFlushDesc, MetricResult{labels: physicalLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsIsActiveDesc, MetricResult{labels: physicalLabels, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsSafeWal, MetricResult{labels: physicalLabels, value: 4096, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsWalStatus, MetricResult{labels: labelMap{"slot_name": "physical_slot", "slot_type": "physical", "wal_status": "reserved"}, value: 1, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsFailoverDesc, MetricResult{labels: physicalLabels, value: 0, metricType: dto.MetricType_GAUGE}},
				{replicationSlotsSyncedDesc, MetricResult{labels: physicalLabels, value: 0, metricType: dto.MetricType_GAUGE}},
			})
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled exceptions: %s", err)
			}
		})
	}
}

func TestPGReplicationSlotsCollectorPG17NullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	rows := sqlmock.NewRows([]string{"slot_name", "database", "active", "pg_wal_lsn_diff"})
	mock.ExpectQuery(sanitizeQuery(replicationSlotsQuery)).WillReturnRows(rows)
	slotRows := sqlmock.NewRows([]string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active", "safe_wal_size", "wal_status", "conflicting", "inactive_since", "failover", "synced"}).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(replicationSlotsSlotQueryPG17)).WillReturnRows(slotRows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotsCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for NULL replication_slots value: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

type replicationSlotsExpectedMetric struct {
	desc   *prometheus.Desc
	result MetricResult
}

// assertReplicationSlotsMetrics reads metrics from ch in order and fails if
// any of them differs from expected or if more metrics follow.
func assertReplicationSlotsMetrics(t *testing.T, ch <-chan prometheus.Metric, expected []replicationSlotsExpectedMetric) {
	t.Helper()
	for i, want := range expected {
		metric, ok := <-ch
		if !ok {
			t.Fatalf("metric %d: channel closed, want %s", i, want.desc)
		}
		if got := metric.Desc().String(); got != want.desc.String() {
			t.Fatalf("metric %d: desc = %s, want %s", i, got, want.desc)
		}
		if got := readMetric(metric); !reflect.DeepEqual(got, want.result) {
			t.Fatalf("metric %d (%s): got %+v, want %+v", i, want.desc, got, want.result)
		}
	}
	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted: %s", metric.Desc())
	}
}
