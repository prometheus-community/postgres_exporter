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
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPGStatWALCollectorVersionedMetrics(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		query      string
		columns    []string
		values     []driver.Value
		want       map[string]float64
		wantMetric map[string]dto.MetricType
	}{
		{
			name:    "PostgreSQL 14 through 17",
			version: "17.0.0",
			query:   statWALQueryBefore18,
			columns: []string{"records", "fpi", "bytes", "buffers_full", "write", "sync", "write_time", "sync_time", "reset_time"},
			values:  []driver.Value{int64(10), int64(2), []byte("1024.00"), int64(3), int64(4), int64(5), 1500.0, 2500.0, 1700000000.0},
			want: map[string]float64{
				"pg_stat_wal_records_total": 10, "pg_stat_wal_fpi_total": 2, "pg_stat_wal_bytes_total": 1024,
				"pg_stat_wal_buffers_full_total": 3, "pg_stat_wal_write_total": 4, "pg_stat_wal_sync_total": 5,
				"pg_stat_wal_write_time_seconds_total": 1.5, "pg_stat_wal_sync_time_seconds_total": 2.5,
				"pg_stat_wal_reset_time_seconds": 1700000000,
			},
		},
		{
			name:    "PostgreSQL 18 and later",
			version: "18.0.0",
			query:   statWALQuery18AndLater,
			columns: []string{"records", "fpi", "bytes", "buffers_full", "reset_time"},
			values:  []driver.Value{int64(10), int64(2), []byte("1024.00"), int64(3), 1700000000.0},
			want: map[string]float64{
				"pg_stat_wal_records_total": 10, "pg_stat_wal_fpi_total": 2, "pg_stat_wal_bytes_total": 1024,
				"pg_stat_wal_buffers_full_total": 3, "pg_stat_wal_reset_time_seconds": 1700000000,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("opening stub database: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery(sanitizeQuery(test.query)).WillReturnRows(
				sqlmock.NewRows(test.columns).AddRow(test.values...),
			)
			inst := &instance{db: db, version: semver.MustParse(test.version), wrapLargeCounters: true}
			ch := make(chan prometheus.Metric, len(test.want))
			if err := (PGStatWALCollector{}).Update(context.Background(), inst, ch); err != nil {
				t.Fatalf("updating collector: %v", err)
			}
			close(ch)

			got := make(map[string]MetricResult, len(test.want))
			for metric := range ch {
				got[metricFamilyName(t, metric)] = readMetric(metric)
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %d metrics, want %d", len(got), len(test.want))
			}
			for name, wantValue := range test.want {
				result, ok := got[name]
				if !ok {
					t.Errorf("missing metric %s", name)
					continue
				}
				if result.value != wantValue {
					t.Errorf("%s value = %v, want %v", name, result.value, wantValue)
				}
				wantType := dto.MetricType_COUNTER
				if name == "pg_stat_wal_reset_time_seconds" {
					wantType = dto.MetricType_GAUGE
				}
				if result.metricType != wantType {
					t.Errorf("%s type = %s, want %s", name, result.metricType, wantType)
				}
				if len(result.labels) != 0 {
					t.Errorf("%s labels = %v, want no labels", name, result.labels)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled sqlmock expectations: %v", err)
			}
		})
	}
}

func TestPGStatWALCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening stub database: %v", err)
	}
	defer db.Close()

	columns := []string{"records", "fpi", "bytes", "buffers_full", "write", "sync", "write_time", "sync_time", "reset_time"}
	values := []driver.Value{nil, nil, nil, nil, nil, nil, nil, nil, nil}
	mock.ExpectQuery(sanitizeQuery(statWALQueryBefore18)).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...))
	inst := &instance{db: db, version: semver.MustParse("14.0.0")}
	ch := make(chan prometheus.Metric, 9)
	if err := (PGStatWALCollector{}).Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("updating collector: %v", err)
	}
	close(ch)
	count := 0
	for metric := range ch {
		count++
		name := metricFamilyName(t, metric)
		if name == "pg_stat_wal_reset_time_seconds" {
			t.Errorf("%s emitted for NULL stats_reset, want omitted", name)
		}
		if got := readMetric(metric).value; got != 0 {
			t.Errorf("%s value = %v, want 0", name, got)
		}
	}
	if count != 8 {
		t.Errorf("got %d metrics for nullable row, want 8", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

func TestPGStatWALCollectorSkipsBeforePostgreSQL14(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening stub database: %v", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("13.0.0")}
	ch := make(chan prometheus.Metric, 1)
	if err := (PGStatWALCollector{}).Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("updating collector: %v", err)
	}
	if len(ch) != 0 {
		t.Fatalf("collector emitted %d metrics on PostgreSQL 13, want none", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database query: %v", err)
	}
}
