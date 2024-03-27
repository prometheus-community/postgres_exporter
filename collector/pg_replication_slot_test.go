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

func TestPgReplicationSlotCollectorActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow("test_slot", "physical", 5, 3, true)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGPostmasterCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 5, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 3, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 1, metricType: dto.MetricType_GAUGE},
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

func TestPgReplicationSlotCollectorInActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow("test_slot", "physical", 6, 12, false)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 6, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 0, metricType: dto.MetricType_GAUGE},
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

func TestPgReplicationSlotCollectorActiveNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow("test_slot", "physical", 6, 12, nil)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 6, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "test_slot", "slot_type": "physical"}, value: 0, metricType: dto.MetricType_GAUGE},
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

func TestPgReplicationSlotCollectorTestNilValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"slot_name", "slot_type", "current_wal_lsn", "confirmed_flush_lsn", "active"}
	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, true)
	mock.ExpectQuery(sanitizeQuery(pgReplicationSlotQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationSlotCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationSlotCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"slot_name": "unknown", "slot_type": "unknown"}, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "unknown", "slot_type": "unknown"}, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"slot_name": "unknown", "slot_type": "unknown"}, value: 1, metricType: dto.MetricType_GAUGE},
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
