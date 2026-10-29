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
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
	for _, tc := range []struct {
		name      string
		isReplica int64
	}{
		{name: "writer", isReplica: 0},
		{name: "reader", isReplica: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error opening a stub db connection: %s", err)
			}
			defer db.Close()

			// Aurora rejects the replay functions of pgReplicationQuery, so it
			// must not run there and only is_replica is reported.
			inst := &instance{db: db, isAurora: true}
			rows := sqlmock.NewRows([]string{"is_replica"}).AddRow(tc.isReplica)
			mock.ExpectQuery(sanitizeQuery(pgReplicationAuroraQuery)).WillReturnRows(rows)

			ch := make(chan prometheus.Metric, 3)
			c := PGReplicationCollector{}
			if err := c.Update(context.Background(), inst, ch); err != nil {
				t.Fatalf("Error calling PGReplicationCollector.Update: %s", err)
			}
			close(ch)

			var metrics []prometheus.Metric
			for m := range ch {
				metrics = append(metrics, m)
			}
			if len(metrics) != 1 {
				t.Fatalf("Expected 1 metric, got %d", len(metrics))
			}
			if metrics[0].Desc() != pgReplicationIsReplica {
				t.Fatalf("Expected pg_replication_is_replica, got %s", metrics[0].Desc())
			}
			expected := MetricResult{labels: labelMap{}, value: float64(tc.isReplica), metricType: dto.MetricType_GAUGE}
			convey.Convey("Metrics comparison", t, func() {
				convey.So(readMetric(metrics[0]), convey.ShouldResemble, expected)
			})
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled exceptions: %s", err)
			}
		})
	}
}

func TestPgReplicationCollectorAuroraError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}
	queryErr := errors.New("connection reset by peer")
	mock.ExpectQuery(sanitizeQuery(pgReplicationAuroraQuery)).WillReturnError(queryErr)

	ch := make(chan prometheus.Metric, 3)
	c := PGReplicationCollector{}
	if err := c.Update(context.Background(), inst, ch); !errors.Is(err, queryErr) {
		t.Errorf("PGReplicationCollector.Update returned %v, want %v", err, queryErr)
	}
	if len(ch) != 0 {
		t.Errorf("Expected no metrics, got %d", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
