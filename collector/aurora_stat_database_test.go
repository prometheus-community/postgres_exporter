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

var auroraStatDatabaseColumns = []string{
	"datid",
	"datname",
	"storage_blks_read",
	"orcache_blks_hit",
	"local_blks_read",
	"storage_blk_read_time",
	"orcache_blk_read_time",
	"local_blk_read_time",
}

func TestAuroraStatDatabaseCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	// The row for shared objects has no database name and is skipped. Read
	// times are exact in binary, so they convert to seconds without
	// rounding errors.
	expectAuroraFunctionExists(mock, "aurora_stat_database", true)
	rows := sqlmock.NewRows(auroraStatDatabaseColumns).
		AddRow("0", nil, 7, 0, 0, 1.5, 0, 0).
		AddRow("14717", "postgres", 623, 425, 0, 3254.5, 89.5, 0)
	mock.ExpectQuery(sanitizeQuery(auroraStatDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDatabaseCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraStatDatabaseCollector.Update: %s", err)
		}
	}()

	labels := labelMap{"datid": "14717", "datname": "postgres"}
	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraStatDatabaseStorageBlocksRead, MetricResult{labels: labels, value: 623, metricType: dto.MetricType_COUNTER}},
		{auroraStatDatabaseOrcacheBlocksHit, MetricResult{labels: labels, value: 425, metricType: dto.MetricType_COUNTER}},
		{auroraStatDatabaseLocalBlocksRead, MetricResult{labels: labels, value: 0, metricType: dto.MetricType_COUNTER}},
		{auroraStatDatabaseStorageBlockReadTime, MetricResult{labels: labels, value: 3.2545, metricType: dto.MetricType_COUNTER}},
		{auroraStatDatabaseOrcacheBlockReadTime, MetricResult{labels: labels, value: 0.0895, metricType: dto.MetricType_COUNTER}},
		{auroraStatDatabaseLocalBlockReadTime, MetricResult{labels: labels, value: 0, metricType: dto.MetricType_COUNTER}},
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

func TestAuroraStatDatabaseCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatDatabaseCollector{}
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

func TestAuroraStatDatabaseCollectorFunctionMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// Older Aurora versions don't have the function: only the existence
	// check runs.
	inst := &instance{db: db, isAurora: true}
	expectAuroraFunctionExists(mock, "aurora_stat_database", false)

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatDatabaseCollector{}
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

func TestAuroraStatDatabaseCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	expectAuroraFunctionExists(mock, "aurora_stat_database", true)
	mock.ExpectQuery(sanitizeQuery(auroraStatDatabaseQuery)).WillReturnRows(sqlmock.NewRows(auroraStatDatabaseColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatDatabaseCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
