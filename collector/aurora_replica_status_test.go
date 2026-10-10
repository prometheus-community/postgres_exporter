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

var auroraReplicaStatusColumns = []string{
	"server_id",
	"replica_lag_in_msec",
	"cur_replay_latency_in_usec",
	"pending_read_ios",
}

func TestAuroraReplicaStatusCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	// The writer reports NULL lag and replay latency, so only
	// pending_read_ios is emitted for it.
	rows := sqlmock.NewRows(auroraReplicaStatusColumns).
		AddRow("writer-instance", nil, nil, 0).
		AddRow("reader-instance-1", 1500.0, 250000.0, 3)
	mock.ExpectQuery(sanitizeQuery(auroraReplicaStatusQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraReplicaStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraReplicaStatusCollector.Update: %s", err)
		}
	}()

	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraReplicaStatusPendingReadIOs, MetricResult{labels: labelMap{"server_id": "writer-instance"}, value: 0, metricType: dto.MetricType_GAUGE}},
		{auroraReplicaStatusLag, MetricResult{labels: labelMap{"server_id": "reader-instance-1"}, value: 1.5, metricType: dto.MetricType_GAUGE}},
		{auroraReplicaStatusReplayLatency, MetricResult{labels: labelMap{"server_id": "reader-instance-1"}, value: 0.25, metricType: dto.MetricType_GAUGE}},
		{auroraReplicaStatusPendingReadIOs, MetricResult{labels: labelMap{"server_id": "reader-instance-1"}, value: 3, metricType: dto.MetricType_GAUGE}},
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

func TestAuroraReplicaStatusCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraReplicaStatusCollector{}
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

func TestAuroraReplicaStatusCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	mock.ExpectQuery(sanitizeQuery(auroraReplicaStatusQuery)).WillReturnRows(sqlmock.NewRows(auroraReplicaStatusColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraReplicaStatusCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
