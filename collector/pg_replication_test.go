// Copyright 2023 The Prometheus Authors
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
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgReplicationCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"lag", "is_replica", "last_replay"}
	rows := sqlmock.NewRows(columns).
		AddRow(1000, 1, 3)
	mock.ExpectQuery(sanitizeQuery(pgReplicationQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGReplicationCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGReplicationCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 1000, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 1, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 3, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPgReplicationCollectorAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	// Aurora rejects the main query because pg_last_xact_replay_timestamp() is
	// not supported. The collector should fall back to the simpler is_replica
	// query and emit NaN for the time-based metrics.
	auroraErr := &pq.Error{
		Code:    "0A000", // feature_not_supported
		Message: "pg_last_xact_replay_timestamp() is currently not supported for Aurora",
	}
	mock.ExpectQuery(sanitizeQuery(pgReplicationQuery)).WillReturnError(auroraErr)

	fallbackColumns := []string{"is_replica"}
	fallbackRows := sqlmock.NewRows(fallbackColumns).AddRow(1)
	mock.ExpectQuery(sanitizeQuery(pgReplicationIsReplicaQuery)).WillReturnRows(fallbackRows)

	ch := make(chan prometheus.Metric, 3)
	c := PGReplicationCollector{}
	if err := c.Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("Unexpected error from Update on Aurora: %s", err)
	}
	close(ch)

	metrics := make([]MetricResult, 0, 3)
	for m := range ch {
		metrics = append(metrics, readMetric(m))
	}

	convey.Convey("Aurora fallback metrics", t, func() {
		convey.So(len(metrics), convey.ShouldEqual, 3)
		// lag should be NaN
		convey.So(math.IsNaN(metrics[0].value), convey.ShouldBeTrue)
		// is_replica should be 1
		convey.So(metrics[1].value, convey.ShouldEqual, 1)
		// last_replay should be NaN
		convey.So(math.IsNaN(metrics[2].value), convey.ShouldBeTrue)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
