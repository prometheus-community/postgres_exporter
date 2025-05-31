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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatWALCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"wal_records",      // bigint
		"wal_fpi",          // bigint
		"wal_bytes",        // numeric
		"wal_buffers_full", // bigint
		"wal_write",        // bigint
		"wal_sync",         // bigint
		"wal_write_time",   // double precision
		"wal_sync_time",    // double precision
		"stats_reset",      // timestamp with time zone
	}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(354, 4945, 289097744, 1242257, int64(3275602074), 89320867, 450.123439, 1234.5678, srT)
	mock.ExpectQuery(sanitizeQuery(statWALQuery(columns))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatWALCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatWALCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 354},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 450.123439},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1234.5678},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1685059842},
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

func TestPGStatWALCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}
	columns := []string{
		"wal_records",      // bigint
		"wal_fpi",          // bigint
		"wal_bytes",        // numeric
		"wal_buffers_full", // bigint
		"wal_write",        // bigint
		"wal_sync",         // bigint
		"wal_write_time",   // double precision
		"wal_sync_time",    // double precision
		"stats_reset",      // timestamp with time zone
	}

	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statWALQuery(columns))).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatWALCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatWALCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
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
