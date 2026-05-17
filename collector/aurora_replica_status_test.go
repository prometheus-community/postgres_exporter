// Copyright 2025 The Prometheus Authors
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
	"github.com/smartystreets/goconvey/convey"
)

func TestAuroraReplicaStatusCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	columns := []string{
		"server_id",
		"replica_lag_in_msec",
		"cur_replay_latency_in_usec",
		"pending_read_ios",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("writer-instance", nil, nil, 0).
		AddRow("reader-instance-1", 15.5, 200.0, 3)

	mock.ExpectQuery(sanitizeQuery(auroraReplicaStatusQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraReplicaStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling Update: %s", err)
		}
	}()

	// writer: 1 metric (pending_read_ios), reader: 3 metrics → 4 total
	expected := 4
	got := 0
	for range ch {
		got++
	}

	convey.Convey("Metrics count", t, func() {
		convey.So(got, convey.ShouldEqual, expected)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestAuroraReplicaStatusCollectorNotAurora(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// isAurora=false → collector must skip immediately and not even attempt the query.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraReplicaStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
			t.Errorf("Expected ErrNoData on non-Aurora, got: %v", err)
		}
	}()
	for range ch {
	}
}

func TestAuroraReplicaStatusCollectorNoData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	columns := []string{
		"server_id",
		"replica_lag_in_msec",
		"cur_replay_latency_in_usec",
		"pending_read_ios",
	}
	rows := sqlmock.NewRows(columns)
	mock.ExpectQuery(sanitizeQuery(auroraReplicaStatusQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraReplicaStatusCollector{}
		err := c.Update(context.Background(), inst, ch)
		if err != ErrNoData {
			t.Errorf("Expected ErrNoData, got: %v", err)
		}
	}()

	for range ch {
	}
}
