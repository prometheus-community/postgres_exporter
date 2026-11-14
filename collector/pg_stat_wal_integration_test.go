// Copyright 2024 The Prometheus Authors
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
	dto "github.com/prometheus/client_model/go"
)

func TestPGStatWALCollectorVersionedQuery(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	version, err := queryVersion(ctx, db)
	if err != nil {
		t.Fatalf("queryVersion() error = %v", err)
	}

	metrics := collectStatWALMetrics(t, &instance{db: db, version: version})

	if version.LT(semver.MustParse("14.0.0")) {
		if len(metrics) != 0 {
			t.Fatalf("got %d metrics before PostgreSQL 14, want 0", len(metrics))
		}
		return
	}

	want := map[string]dto.MetricType{
		statsWALRecordsDesc.String():     dto.MetricType_COUNTER,
		statsWALFPIDesc.String():         dto.MetricType_COUNTER,
		statsWALBytesDesc.String():       dto.MetricType_COUNTER,
		statsWALBuffersFullDesc.String(): dto.MetricType_COUNTER,
		statsWALWriteDesc.String():       dto.MetricType_COUNTER,
		statsWALSyncDesc.String():        dto.MetricType_COUNTER,
		statsWALWriteTimeDesc.String():   dto.MetricType_COUNTER,
		statsWALSyncTimeDesc.String():    dto.MetricType_COUNTER,
		statsWALStatsResetDesc.String():  dto.MetricType_GAUGE,
	}

	if len(metrics) != len(want) {
		t.Fatalf("got %d metrics on PostgreSQL %s, want %d", len(metrics), version, len(want))
	}
	for _, metric := range metrics {
		desc := metric.Desc().String()
		metricType, ok := want[desc]
		if !ok {
			t.Errorf("unexpected metric %s on PostgreSQL %s", desc, version)
			continue
		}
		if got := readMetric(metric).metricType; got != metricType {
			t.Errorf("metric %s has type %s, want %s", desc, got, metricType)
		}
		delete(want, desc)
	}
	for desc := range want {
		t.Errorf("missing metric %s on PostgreSQL %s", desc, version)
	}
}
