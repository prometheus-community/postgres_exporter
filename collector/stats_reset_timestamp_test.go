// Copyright 2026 The Prometheus Authors
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
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
)

type resetTestCollector []prometheus.Metric

func (c resetTestCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c {
		ch <- metric.Desc()
	}
}

func (c resetTestCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range c {
		ch <- metric
	}
}

func resetTestExposition(t *testing.T, metrics <-chan prometheus.Metric) string {
	t.Helper()
	var collected resetTestCollector
	for metric := range metrics {
		collected = append(collected, metric)
	}
	registry := prometheus.NewRegistry()
	if err := registry.Register(collected); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
	if recorder.Code != 200 {
		t.Fatalf("metrics response: %d: %s", recorder.Code, recorder.Body.String())
	}
	return recorder.Body.String()
}

func assertResetFamily(t *testing.T, exposition, legacy, canonical, labels string, value float64) {
	t.Helper()
	for _, family := range []struct {
		name, metricType, helpFragment string
	}{
		{legacy, "counter", "Deprecated: use " + canonical},
		{canonical, "gauge", "0 means PostgreSQL returned NULL"},
	} {
		for _, fragment := range []string{
			"# HELP " + family.name + " ",
			"# TYPE " + family.name + " " + family.metricType,
			family.helpFragment,
			family.name + labels + " " + formatResetTestValue(value) + "\n",
		} {
			if !strings.Contains(exposition, fragment) {
				t.Errorf("missing %q in exposition:\n%s", fragment, exposition)
			}
		}
	}
}

func formatResetTestValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func TestStatsResetTimestampExposition(t *testing.T) {
	resetTime := time.Date(2023, 5, 26, 0, 10, 42, 0, time.UTC)
	for _, test := range []struct {
		name, version string
		reset         driver.Value
	}{
		{"pre17_timestamp", "16.0.0", resetTime},
		{"pre17_null", "16.0.0", nil},
		{"post17_timestamp", "17.0.0", resetTime},
		{"post17_null", "17.0.0", nil},
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
				value = 1685059842
			}
			assertResetFamily(t, resetTestExposition(t, ch), "pg_stat_bgwriter_stats_reset_total", "pg_stat_bgwriter_stats_reset_timestamp_seconds", "", value)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}

	for _, test := range []struct {
		name  string
		reset driver.Value
	}{
		{"timestamp", resetTime},
		{"null", nil},
	} {
		t.Run("checkpointer_"+test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			inst := &instance{db: db, version: semver.MustParse("17.0.0")}
			rows := sqlmock.NewRows([]string{"num_timed", "num_requested", "restartpoints_timed", "restartpoints_req", "restartpoints_done", "write_time", "sync_time", "buffers_written", "stats_reset"}).AddRow(0, 0, 0, 0, 0, 0, 0, 0, test.reset)
			mock.ExpectQuery(sanitizeQuery(statCheckpointerQuery)).WillReturnRows(rows)
			ch := make(chan prometheus.Metric, 20)
			if err := (PGStatCheckpointerCollector{}).Update(context.Background(), inst, ch); err != nil {
				t.Fatal(err)
			}
			close(ch)
			value := float64(0)
			if test.reset != nil {
				value = 1685059842
			}
			assertResetFamily(t, resetTestExposition(t, ch), "pg_stat_checkpointer_stats_reset_total", "pg_stat_checkpointer_stats_reset_timestamp_seconds", "", value)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}

	t.Run("checkpointer_pre17", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 1)
		inst := &instance{version: semver.MustParse("16.0.0")}
		c := PGStatCheckpointerCollector{log: promslog.NewNopLogger()}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Fatal(err)
		}
		if len(ch) != 0 {
			t.Fatal("checkpointer emitted metrics on PostgreSQL 16")
		}
	})

	t.Run("database_labels_and_null", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		inst := &instance{db: db, version: semver.MustParse("14.0.0")}
		values := func(id, name string, reset driver.Value) []driver.Value {
			row := make([]driver.Value, len(pgStatCollectorTestColumns))
			for i := range row {
				row[i] = int64(0)
			}
			row[0], row[1], row[18] = id, name, reset
			return row
		}
		rows := sqlmock.NewRows(pgStatCollectorTestColumns).
			AddRow(values("1", "first", resetTime)...).
			AddRow(values("2", "second", nil)...)
		mock.ExpectQuery(sanitizeQuery(statDatabaseQuery(pgStatCollectorTestColumns))).WillReturnRows(rows)
		ch := make(chan prometheus.Metric, 100)
		c := PGStatDatabaseCollector{log: promslog.NewNopLogger()}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Fatal(err)
		}
		close(ch)
		exposition := resetTestExposition(t, ch)
		assertResetFamily(t, exposition, "pg_stat_database_stats_reset", "pg_stat_database_stats_reset_timestamp_seconds", `{datid="1",datname="first"}`, 1685059842)
		assertResetFamily(t, exposition, "pg_stat_database_stats_reset", "pg_stat_database_stats_reset_timestamp_seconds", `{datid="2",datname="second"}`, 0)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}
