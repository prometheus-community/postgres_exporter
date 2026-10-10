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

var auroraStatLogicalWalCacheColumns = []string{
	"name",
	"cache_hit",
	"cache_miss",
	"blks_read",
}

func TestAuroraStatLogicalWalCacheCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	expectAuroraFunctionExists(mock, "aurora_stat_logical_wal_cache", true)
	rows := sqlmock.NewRows(auroraStatLogicalWalCacheColumns).
		AddRow("test_slot1", 24, 0, 24).
		AddRow("test_slot2", 1, 2, 3)
	mock.ExpectQuery(sanitizeQuery(auroraStatLogicalWalCacheQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatLogicalWalCacheCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraStatLogicalWalCacheCollector.Update: %s", err)
		}
	}()

	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraStatLogicalWalCacheHits, MetricResult{labels: labelMap{"slot_name": "test_slot1"}, value: 24, metricType: dto.MetricType_COUNTER}},
		{auroraStatLogicalWalCacheMisses, MetricResult{labels: labelMap{"slot_name": "test_slot1"}, value: 0, metricType: dto.MetricType_COUNTER}},
		{auroraStatLogicalWalCacheBlocksRead, MetricResult{labels: labelMap{"slot_name": "test_slot1"}, value: 24, metricType: dto.MetricType_COUNTER}},
		{auroraStatLogicalWalCacheHits, MetricResult{labels: labelMap{"slot_name": "test_slot2"}, value: 1, metricType: dto.MetricType_COUNTER}},
		{auroraStatLogicalWalCacheMisses, MetricResult{labels: labelMap{"slot_name": "test_slot2"}, value: 2, metricType: dto.MetricType_COUNTER}},
		{auroraStatLogicalWalCacheBlocksRead, MetricResult{labels: labelMap{"slot_name": "test_slot2"}, value: 3, metricType: dto.MetricType_COUNTER}},
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

func TestAuroraStatLogicalWalCacheCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatLogicalWalCacheCollector{}
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

func TestAuroraStatLogicalWalCacheCollectorFunctionMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// Older Aurora versions don't have the function: only the existence
	// check runs.
	inst := &instance{db: db, isAurora: true}
	expectAuroraFunctionExists(mock, "aurora_stat_logical_wal_cache", false)

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatLogicalWalCacheCollector{}
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

func TestAuroraStatLogicalWalCacheCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	expectAuroraFunctionExists(mock, "aurora_stat_logical_wal_cache", true)
	mock.ExpectQuery(sanitizeQuery(auroraStatLogicalWalCacheQuery)).WillReturnRows(sqlmock.NewRows(auroraStatLogicalWalCacheColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatLogicalWalCacheCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
