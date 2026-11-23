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
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatBGWriterCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"checkpoints_timed",
		"checkpoints_req",
		"checkpoint_write_time",
		"checkpoint_sync_time",
		"buffers_checkpoint",
		"buffers_clean",
		"maxwritten_clean",
		"buffers_backend",
		"buffers_backend_fsync",
		"buffers_alloc",
		"stats_reset"}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(354, 4945, 289097744, 1242257, int64(3275602074), 89320867, 450139, 2034563757, 0, int64(2725688749), srT)
	mock.ExpectQuery(sanitizeQuery(statBGWriterQueryBefore17)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatBGWriterCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatBGWriterCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 354},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 2725688749},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1685059842},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 1685059842},
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

func TestPGStatBGWriterCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"checkpoints_timed",
		"checkpoints_req",
		"checkpoint_write_time",
		"checkpoint_sync_time",
		"buffers_checkpoint",
		"buffers_clean",
		"maxwritten_clean",
		"buffers_backend",
		"buffers_backend_fsync",
		"buffers_alloc",
		"stats_reset"}

	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statBGWriterQueryBefore17)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatBGWriterCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatBGWriterCollector.Update: %s", err)
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

func TestPGStatBGWriterCollectorWrapsLargeCounters(t *testing.T) {
	const largeCounter = int64(1<<53) + 1

	tests := []struct {
		name  string
		wrap  bool
		value float64
	}{
		{name: "enabled", wrap: true, value: 1},
		{name: "disabled", wrap: false, value: float64(largeCounter)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("opening stub database: %s", err)
			}
			t.Cleanup(func() { _ = db.Close() })

			inst := &instance{
				db:                db,
				version:           semver.MustParse("17.0.0"),
				wrapLargeCounters: test.wrap,
			}
			columns := []string{"buffers_clean", "maxwritten_clean", "buffers_alloc", "stats_reset"}
			rows := sqlmock.NewRows(columns).AddRow(largeCounter, 0, 0, nil)
			mock.ExpectQuery(sanitizeQuery(statBGWriterQueryAfter17)).WillReturnRows(rows)

			ch := make(chan prometheus.Metric)
			go func() {
				defer close(ch)
				if err := (PGStatBGWriterCollector{}).Update(context.Background(), inst, ch); err != nil {
					t.Errorf("updating collector: %s", err)
				}
			}()

			metric := readMetric(<-ch)
			if metric.value != test.value {
				t.Fatalf("buffers_clean value = %v, want %v", metric.value, test.value)
			}
			for range ch {
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet database expectations: %v", err)
			}
		})
	}
}

func TestPGStatBGWriterCollectorStatsResetTimestampExposition(t *testing.T) {
	resetTime := time.Date(2023, 5, 26, 0, 10, 42, 0, time.UTC)
	for _, test := range []struct {
		name, version string
		reset         driver.Value
	}{
		{"pre17_timestamp", "16.0.0", resetTime},
		{"pre17_null", "16.0.0", nil},
		{"pre17_epoch", "16.0.0", time.Unix(0, 0)},
		{"post17_timestamp", "17.0.0", resetTime},
		{"post17_null", "17.0.0", nil},
		{"post17_epoch", "17.0.0", time.Unix(0, 0)},
	} {
		t.Run("bgwriter_"+test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			inst := &instance{db: db, version: semver.MustParse(test.version)}
			query := statBGWriterQueryBefore17
			rows := sqlmock.NewRows([]string{"checkpoints_timed", "checkpoints_req", "checkpoint_write_time", "checkpoint_sync_time", "buffers_checkpoint", "buffers_clean", "maxwritten_clean", "buffers_backend", "buffers_backend_fsync", "buffers_alloc", "stats_reset"}).AddRow(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, test.reset)
			if inst.version.GE(semver.MustParse("17.0.0")) {
				query = statBGWriterQueryAfter17
				rows = sqlmock.NewRows([]string{"buffers_clean", "maxwritten_clean", "buffers_alloc", "stats_reset"}).AddRow(0, 0, 0, test.reset)
			}
			mock.ExpectQuery(sanitizeQuery(query)).WillReturnRows(rows)
			ch := make(chan prometheus.Metric, 20)
			if err := (PGStatBGWriterCollector{}).Update(context.Background(), inst, ch); err != nil {
				t.Fatal(err)
			}
			close(ch)
			value := float64(0)
			if test.reset != nil {
				value = float64(test.reset.(time.Time).Unix())
			}
			assertResetFamily(t, resetTestExposition(t, ch), "pg_stat_bgwriter_stats_reset_total", "pg_stat_bgwriter_stats_reset_timestamp_seconds", "", value, test.reset != nil)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
