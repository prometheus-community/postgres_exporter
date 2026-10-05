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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

var auroraStatOptimizedReadsCacheColumns = []string{
	"total_size",
	"used_size",
}

func TestAuroraStatOptimizedReadsCacheCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	expectAuroraFunctionExists(mock, "aurora_stat_optimized_reads_cache", true)
	rows := sqlmock.NewRows(auroraStatOptimizedReadsCacheColumns).
		AddRow(1131723177984, 1046896472064)
	mock.ExpectQuery(sanitizeQuery(auroraStatOptimizedReadsCacheQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatOptimizedReadsCacheCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraStatOptimizedReadsCacheCollector.Update: %s", err)
		}
	}()

	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraStatOptimizedReadsCacheTotalSize, MetricResult{labels: labelMap{}, value: 1131723177984, metricType: dto.MetricType_GAUGE}},
		{auroraStatOptimizedReadsCacheUsedSize, MetricResult{labels: labelMap{}, value: 1046896472064, metricType: dto.MetricType_GAUGE}},
	}
	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := <-ch
			convey.So(m.Desc(), convey.ShouldEqual, expect.desc)
			convey.So(readMetric(m), convey.ShouldResemble, expect.result)
		}
		_, more := <-ch
		convey.So(more, convey.ShouldBeFalse)
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestAuroraStatOptimizedReadsCacheCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatOptimizedReadsCacheCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData on non-Aurora server, got: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("Expected no metrics on non-Aurora server, got %d", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestAuroraStatOptimizedReadsCacheCollectorFunctionMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// Older Aurora versions don't have the function: only the existence
	// check runs.
	inst := &instance{db: db, isAurora: true}
	expectAuroraFunctionExists(mock, "aurora_stat_optimized_reads_cache", false)

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatOptimizedReadsCacheCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData when the function is missing, got: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("Expected no metrics when the function is missing, got %d", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestAuroraStatOptimizedReadsCacheCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	expectAuroraFunctionExists(mock, "aurora_stat_optimized_reads_cache", true)
	mock.ExpectQuery(sanitizeQuery(auroraStatOptimizedReadsCacheQuery)).WillReturnRows(sqlmock.NewRows(auroraStatOptimizedReadsCacheColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatOptimizedReadsCacheCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
