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
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPGLongRunningTransactionsCollectorFiltersByThreshold(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback()
	time.Sleep(10 * time.Millisecond)

	inst := &instance{db: db}
	if got := collectLongRunningTransactionsCount(t, inst, time.Millisecond); got < 1 {
		t.Fatalf("transaction count with 1ms threshold = %v, want at least 1", got)
	}
	if got := collectLongRunningTransactionsCount(t, inst, 24*time.Hour); got != 0 {
		t.Fatalf("transaction count with 24h threshold = %v, want 0", got)
	}
}

func collectLongRunningTransactionsCount(t *testing.T, inst *instance, threshold time.Duration) float64 {
	t.Helper()

	ch := make(chan prometheus.Metric, 2)
	collector := PGLongRunningTransactionsCollector{threshold: threshold}
	if err := collector.Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	return readMetric(<-ch).value
}
