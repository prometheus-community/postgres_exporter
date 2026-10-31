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

//go:build integration

package collector

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPGStatIOCollectorIntegration(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	version, err := queryVersion(ctx, db)
	if err != nil {
		t.Fatalf("queryVersion() error = %v", err)
	}
	if version.LT(semver.MustParse("16.0.0")) {
		t.Skipf("pg_stat_io is available starting with PostgreSQL 16, server is %s", version)
	}

	ch := make(chan prometheus.Metric)
	errCh := make(chan error, 1)
	go func() {
		errCh <- (PGStatIOCollector{}).Update(ctx, &instance{db: db, version: version}, ch)
		close(ch)
	}()

	wantNames := []string{
		"pg_stat_io_reads_total", "pg_stat_io_read_time_seconds_total",
		"pg_stat_io_writes_total", "pg_stat_io_write_time_seconds_total",
		"pg_stat_io_writebacks_total", "pg_stat_io_writeback_time_seconds_total",
		"pg_stat_io_extends_total", "pg_stat_io_extend_time_seconds_total",
		"pg_stat_io_hits_total", "pg_stat_io_evictions_total", "pg_stat_io_reuses_total",
		"pg_stat_io_fsyncs_total", "pg_stat_io_fsync_time_seconds_total", "pg_stat_io_reset_time_seconds",
	}
	if version.GE(semver.MustParse("18.0.0")) {
		wantNames = append(wantNames, "pg_stat_io_read_bytes_total", "pg_stat_io_write_bytes_total", "pg_stat_io_extend_bytes_total")
	}

	found := make(map[string]bool)
	metricCount := 0
	foundClientBackend := false
	for metric := range ch {
		metricCount++
		name := metricFamilyName(t, metric)
		found[name] = true
		result := readMetric(metric)
		wantType := dto.MetricType_COUNTER
		if name == "pg_stat_io_reset_time_seconds" {
			wantType = dto.MetricType_GAUGE
			if len(result.labels) != 0 {
				t.Errorf("%s labels = %v, want no labels", name, result.labels)
			}
		} else {
			for _, label := range []string{"backend_type", "object", "context"} {
				if _, ok := result.labels[label]; !ok {
					t.Errorf("%s missing %q label: %v", name, label, result.labels)
				}
			}
			if result.labels["backend_type"] == "client backend" && result.labels["object"] == "relation" && result.labels["context"] == "normal" {
				foundClientBackend = true
			}
		}
		if result.metricType != wantType {
			t.Errorf("%s type = %s, want %s", name, result.metricType, wantType)
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if metricCount == 0 {
		t.Fatal("collector returned no pg_stat_io metrics")
	}
	if !foundClientBackend {
		t.Error("collector did not emit a client backend/relation/normal series")
	}
	for _, name := range wantNames {
		if !found[name] {
			t.Errorf("collector did not emit %s", name)
		}
	}
}
