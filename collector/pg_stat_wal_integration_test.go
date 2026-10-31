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

func TestPGStatWALCollectorIntegration(t *testing.T) {
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
	if version.LT(semver.MustParse("14.0.0")) {
		t.Skipf("pg_stat_wal is available starting with PostgreSQL 14, server is %s", version)
	}

	ch := make(chan prometheus.Metric, 9)
	if err := (PGStatWALCollector{}).Update(ctx, &instance{db: db, version: version}, ch); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	close(ch)

	wantNames := []string{
		"pg_stat_wal_records_total", "pg_stat_wal_fpi_total", "pg_stat_wal_bytes_total",
		"pg_stat_wal_buffers_full_total", "pg_stat_wal_reset_time_seconds",
	}
	if version.LT(semver.MustParse("18.0.0")) {
		wantNames = append(wantNames, "pg_stat_wal_write_total", "pg_stat_wal_sync_total", "pg_stat_wal_write_time_seconds_total", "pg_stat_wal_sync_time_seconds_total")
	}

	found := make(map[string]bool)
	for metric := range ch {
		name := metricFamilyName(t, metric)
		found[name] = true
		got := readMetric(metric)
		wantType := dto.MetricType_COUNTER
		if name == "pg_stat_wal_reset_time_seconds" {
			wantType = dto.MetricType_GAUGE
		}
		if got.metricType != wantType {
			t.Errorf("%s type = %s, want %s", name, got.metricType, wantType)
		}
		if len(got.labels) != 0 {
			t.Errorf("%s labels = %v, want no labels", name, got.labels)
		}
	}
	for _, name := range wantNames {
		if !found[name] {
			t.Errorf("collector did not emit %s", name)
		}
	}
}
