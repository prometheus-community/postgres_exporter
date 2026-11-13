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
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPGStatIOCollectorVersionedMetrics(t *testing.T) {
	labels := labelMap{"backend_type": "client backend", "object": "relation", "context": "normal"}
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
			name:    "PostgreSQL 16 and 17",
			version: "17.0.0",
			query:   statIOQueryBefore18,
			columns: []string{"backend_type", "object", "context", "reads", "read_time", "writes", "write_time", "writebacks", "writeback_time", "extends", "extend_time", "hits", "evictions", "reuses", "fsyncs", "fsync_time", "reset_time"},
			values:  []driver.Value{"client backend", "relation", "normal", int64(10), 1500.0, int64(12), 2500.0, int64(14), 3500.0, int64(16), 4500.0, int64(18), int64(20), int64(22), int64(24), 5500.0, 1700000000.0},
			want: map[string]float64{
				"pg_stat_io_reads_total": 10, "pg_stat_io_read_time_seconds_total": 1.5,
				"pg_stat_io_writes_total": 12, "pg_stat_io_write_time_seconds_total": 2.5,
				"pg_stat_io_writebacks_total": 14, "pg_stat_io_writeback_time_seconds_total": 3.5,
				"pg_stat_io_extends_total": 16, "pg_stat_io_extend_time_seconds_total": 4.5,
				"pg_stat_io_hits_total": 18, "pg_stat_io_evictions_total": 20, "pg_stat_io_reuses_total": 22,
				"pg_stat_io_fsyncs_total": 24, "pg_stat_io_fsync_time_seconds_total": 5.5,
				"pg_stat_io_reset_time_seconds": 1700000000,
			},
			wantMetric: map[string]dto.MetricType{
				"pg_stat_io_reads_total": dto.MetricType_COUNTER, "pg_stat_io_read_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_writes_total": dto.MetricType_COUNTER, "pg_stat_io_write_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_writebacks_total": dto.MetricType_COUNTER, "pg_stat_io_writeback_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_extends_total": dto.MetricType_COUNTER, "pg_stat_io_extend_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_hits_total": dto.MetricType_COUNTER, "pg_stat_io_evictions_total": dto.MetricType_COUNTER,
				"pg_stat_io_reuses_total": dto.MetricType_COUNTER, "pg_stat_io_fsyncs_total": dto.MetricType_COUNTER,
				"pg_stat_io_fsync_time_seconds_total": dto.MetricType_COUNTER, "pg_stat_io_reset_time_seconds": dto.MetricType_GAUGE,
			},
		},
		{
			name:    "PostgreSQL 18 and later",
			version: "18.0.0",
			query:   statIOQuery18AndLater,
			columns: []string{"backend_type", "object", "context", "reads", "read_bytes", "read_time", "writes", "write_bytes", "write_time", "writebacks", "writeback_time", "extends", "extend_bytes", "extend_time", "hits", "evictions", "reuses", "fsyncs", "fsync_time", "reset_time"},
			values:  []driver.Value{"client backend", "relation", "normal", int64(10), []byte("110.0"), 1500.0, int64(12), []byte("130.0"), 2500.0, int64(14), 3500.0, int64(16), []byte("170.0"), 4500.0, int64(18), int64(20), int64(22), int64(24), 5500.0, 1700000000.0},
			want: map[string]float64{
				"pg_stat_io_reads_total": 10, "pg_stat_io_read_bytes_total": 110, "pg_stat_io_read_time_seconds_total": 1.5,
				"pg_stat_io_writes_total": 12, "pg_stat_io_write_bytes_total": 130, "pg_stat_io_write_time_seconds_total": 2.5,
				"pg_stat_io_writebacks_total": 14, "pg_stat_io_writeback_time_seconds_total": 3.5,
				"pg_stat_io_extends_total": 16, "pg_stat_io_extend_bytes_total": 170, "pg_stat_io_extend_time_seconds_total": 4.5,
				"pg_stat_io_hits_total": 18, "pg_stat_io_evictions_total": 20, "pg_stat_io_reuses_total": 22,
				"pg_stat_io_fsyncs_total": 24, "pg_stat_io_fsync_time_seconds_total": 5.5,
				"pg_stat_io_reset_time_seconds": 1700000000,
			},
			wantMetric: map[string]dto.MetricType{
				"pg_stat_io_reads_total": dto.MetricType_COUNTER, "pg_stat_io_read_bytes_total": dto.MetricType_COUNTER,
				"pg_stat_io_read_time_seconds_total": dto.MetricType_COUNTER, "pg_stat_io_writes_total": dto.MetricType_COUNTER,
				"pg_stat_io_write_bytes_total": dto.MetricType_COUNTER, "pg_stat_io_write_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_writebacks_total": dto.MetricType_COUNTER, "pg_stat_io_writeback_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_extends_total": dto.MetricType_COUNTER, "pg_stat_io_extend_bytes_total": dto.MetricType_COUNTER,
				"pg_stat_io_extend_time_seconds_total": dto.MetricType_COUNTER, "pg_stat_io_hits_total": dto.MetricType_COUNTER,
				"pg_stat_io_evictions_total": dto.MetricType_COUNTER, "pg_stat_io_reuses_total": dto.MetricType_COUNTER,
				"pg_stat_io_fsyncs_total": dto.MetricType_COUNTER, "pg_stat_io_fsync_time_seconds_total": dto.MetricType_COUNTER,
				"pg_stat_io_reset_time_seconds": dto.MetricType_GAUGE,
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
			if err := (PGStatIOCollector{}).Update(context.Background(), inst, ch); err != nil {
				t.Fatalf("updating collector: %v", err)
			}
			close(ch)

			got := make(map[string]MetricResult, len(test.want))
			for metric := range ch {
				name := metricFamilyName(t, metric)
				got[name] = readMetric(metric)
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
				if result.metricType != test.wantMetric[name] {
					t.Errorf("%s type = %s, want %s", name, result.metricType, test.wantMetric[name])
				}
				wantLabels := labels
				if name == "pg_stat_io_reset_time_seconds" {
					wantLabels = labelMap{}
				}
				if !mapsEqual(result.labels, wantLabels) {
					t.Errorf("%s labels = %v, want %v", name, result.labels, wantLabels)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled sqlmock expectations: %v", err)
			}
		})
	}
}

func TestPGStatIOCollectorOmitsNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening stub database: %v", err)
	}
	defer db.Close()

	columns := []string{"backend_type", "object", "context", "reads", "read_time", "writes", "write_time", "writebacks", "writeback_time", "extends", "extend_time", "hits", "evictions", "reuses", "fsyncs", "fsync_time", "reset_time"}
	mock.ExpectQuery(sanitizeQuery(statIOQueryBefore18)).WillReturnRows(sqlmock.NewRows(columns).
		AddRow("client backend", "temp relation", "normal", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).
		AddRow("checkpointer", "relation", "normal", nil, nil, int64(7), 700.0, int64(8), 800.0, nil, nil, nil, nil, nil, int64(9), 900.0, nil))
	inst := &instance{db: db, version: semver.MustParse("16.0.0")}
	ch := make(chan prometheus.Metric, 32)
	if err := (PGStatIOCollector{}).Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("updating collector: %v", err)
	}
	close(ch)

	want := map[string]float64{
		"pg_stat_io_writes_total": 7, "pg_stat_io_write_time_seconds_total": 0.7,
		"pg_stat_io_writebacks_total": 8, "pg_stat_io_writeback_time_seconds_total": 0.8,
		"pg_stat_io_fsyncs_total": 9, "pg_stat_io_fsync_time_seconds_total": 0.9,
	}
	wantLabels := labelMap{"backend_type": "checkpointer", "object": "relation", "context": "normal"}
	count := 0
	for metric := range ch {
		count++
		name := metricFamilyName(t, metric)
		result := readMetric(metric)
		wantValue, ok := want[name]
		if !ok {
			t.Errorf("unexpected metric %s %v for NULL value", name, result.labels)
			continue
		}
		if result.value != wantValue {
			t.Errorf("%s value = %v, want %v", name, result.value, wantValue)
		}
		if !mapsEqual(result.labels, wantLabels) {
			t.Errorf("%s labels = %v, want %v", name, result.labels, wantLabels)
		}
	}
	if count != len(want) {
		t.Errorf("got %d metrics, want %d", count, len(want))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

func TestPGStatIOCollectorEmitsResetTimeOnce(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening stub database: %v", err)
	}
	defer db.Close()

	columns := []string{"backend_type", "object", "context", "reads", "read_time", "writes", "write_time", "writebacks", "writeback_time", "extends", "extend_time", "hits", "evictions", "reuses", "fsyncs", "fsync_time", "reset_time"}
	mock.ExpectQuery(sanitizeQuery(statIOQueryBefore18)).WillReturnRows(sqlmock.NewRows(columns).
		AddRow("client backend", "relation", "normal", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 1700000000.0).
		AddRow("checkpointer", "relation", "normal", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 1700000000.0))
	inst := &instance{db: db, version: semver.MustParse("16.0.0")}
	ch := make(chan prometheus.Metric, 4)
	if err := (PGStatIOCollector{}).Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("updating collector: %v", err)
	}
	close(ch)

	count := 0
	for metric := range ch {
		count++
		result := readMetric(metric)
		if name := metricFamilyName(t, metric); name != "pg_stat_io_reset_time_seconds" {
			t.Errorf("unexpected metric %s", name)
		}
		if result.value != 1700000000 || len(result.labels) != 0 {
			t.Errorf("reset_time = %v %v, want 1700000000 with no labels", result.value, result.labels)
		}
	}
	if count != 1 {
		t.Errorf("got %d reset_time metrics, want 1", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

func TestPGStatIOCollectorSkipsBeforePostgreSQL16(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening stub database: %v", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("15.0.0")}
	ch := make(chan prometheus.Metric, 1)
	if err := (PGStatIOCollector{}).Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("updating collector: %v", err)
	}
	if len(ch) != 0 {
		t.Fatalf("collector emitted %d metrics on PostgreSQL 15, want none", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database query: %v", err)
	}
}

func metricFamilyName(t *testing.T, metric prometheus.Metric) string {
	t.Helper()
	const prefix = `fqName: "`
	desc := metric.Desc().String()
	start := strings.Index(desc, prefix)
	if start < 0 {
		t.Fatalf("metric descriptor %q has no fqName", desc)
	}
	desc = desc[start+len(prefix):]
	end := strings.IndexByte(desc, '"')
	if end < 0 {
		t.Fatalf("metric descriptor %q has an unterminated fqName", metric.Desc())
	}
	return desc[:end]
}

func mapsEqual(got, want labelMap) bool {
	if len(got) != len(want) {
		return false
	}
	for key, value := range want {
		if got[key] != value {
			return false
		}
	}
	return true
}
