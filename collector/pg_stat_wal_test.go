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
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
	"github.com/smartystreets/goconvey/convey"
)

type statsWALExpectedMetric struct {
	desc       *prometheus.Desc
	metricType dto.MetricType
	value      float64
}

func TestPGStatWALCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

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
	mock.ExpectQuery(sanitizeQuery(statWALQuery)).WillReturnRows(rows)

	metrics := collectStatWALMetrics(t, inst)
	expected := []statsWALExpectedMetric{
		{desc: statsWALRecordsDesc, metricType: dto.MetricType_COUNTER, value: 354},
		{desc: statsWALFPIDesc, metricType: dto.MetricType_COUNTER, value: 4945},
		{desc: statsWALBytesDesc, metricType: dto.MetricType_COUNTER, value: 289097744},
		{desc: statsWALBuffersFullDesc, metricType: dto.MetricType_COUNTER, value: 1242257},
		{desc: statsWALWriteDesc, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{desc: statsWALSyncDesc, metricType: dto.MetricType_COUNTER, value: 89320867},
		{desc: statsWALWriteTimeDesc, metricType: dto.MetricType_COUNTER, value: 450.123439 / 1000},
		{desc: statsWALSyncTimeDesc, metricType: dto.MetricType_COUNTER, value: 1234.5678 / 1000},
		{desc: statsWALStatsResetDesc, metricType: dto.MetricType_GAUGE, value: 1685059842},
	}

	convey.Convey("Metrics comparison", t, func() {
		convey.So(metrics, convey.ShouldHaveLength, len(expected))
		for i, expect := range expected {
			convey.So(metrics[i].Desc().String(), convey.ShouldEqual, expect.desc.String())
			convey.So(readMetric(metrics[i]), convey.ShouldResemble, MetricResult{
				labels:     labelMap{},
				metricType: expect.metricType,
				value:      expect.value,
			})
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

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

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
	mock.ExpectQuery(sanitizeQuery(statWALQuery)).WillReturnRows(rows)

	convey.Convey("No metrics are exposed for NULL values", t, func() {
		convey.So(collectStatWALMetrics(t, inst), convey.ShouldBeEmpty)
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatWALCollectorPostgres18(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("18.0.0")}

	columns := []string{
		"wal_records",
		"wal_fpi",
		"wal_bytes",
		"wal_buffers_full",
		"stats_reset",
	}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(354, 4945, 289097744, 1242257, srT)
	mock.ExpectQuery(sanitizeQuery(statWALQuery18)).WillReturnRows(rows)

	metrics := collectStatWALMetrics(t, inst)
	expected := []statsWALExpectedMetric{
		{desc: statsWALRecordsDesc, metricType: dto.MetricType_COUNTER, value: 354},
		{desc: statsWALFPIDesc, metricType: dto.MetricType_COUNTER, value: 4945},
		{desc: statsWALBytesDesc, metricType: dto.MetricType_COUNTER, value: 289097744},
		{desc: statsWALBuffersFullDesc, metricType: dto.MetricType_COUNTER, value: 1242257},
		{desc: statsWALStatsResetDesc, metricType: dto.MetricType_GAUGE, value: 1685059842},
	}

	convey.Convey("Metrics comparison", t, func() {
		convey.So(metrics, convey.ShouldHaveLength, len(expected))
		for i, expect := range expected {
			convey.So(metrics[i].Desc().String(), convey.ShouldEqual, expect.desc.String())
			convey.So(readMetric(metrics[i]), convey.ShouldResemble, MetricResult{
				labels:     labelMap{},
				metricType: expect.metricType,
				value:      expect.value,
			})
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatWALCollectorBeforePostgres14(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.0.0")}

	convey.Convey("No query or metrics are emitted before PostgreSQL 14", t, func() {
		convey.So(collectStatWALMetrics(t, inst), convey.ShouldBeEmpty)
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func collectStatWALMetrics(t *testing.T, inst *instance) []prometheus.Metric {
	t.Helper()

	ch := make(chan prometheus.Metric)
	errCh := make(chan error, 1)
	go func() {
		defer close(ch)
		c := PGStatWALCollector{log: promslog.NewNopLogger()}
		errCh <- c.Update(context.Background(), inst, ch)
	}()

	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Error calling PGStatWALCollector.Update: %s", err)
	}
	return metrics
}
