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

func TestAuroraGlobalDBInstanceStatusCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	columns := []string{"server_id", "aws_region", "visibility_lag_in_msec"}
	rows := sqlmock.NewRows(columns).
		AddRow("writer", "eu-west-1", nil). // writer: NULL lag, skipped
		AddRow("reader-1", "eu-west-1", 6.0).
		AddRow("reader-2", "eu-central-1", 996.0)
	mock.ExpectQuery(sanitizeQuery(auroraGlobalDBInstanceStatusQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraGlobalDBInstanceStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling Update: %s", err)
		}
	}()

	got := 0
	for range ch {
		got++
	}
	convey.Convey("Metrics count", t, func() {
		convey.So(got, convey.ShouldEqual, 2) // writer skipped (NULL), 2 readers emitted
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestAuroraGlobalDBInstanceStatusCollectorNotAurora(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraGlobalDBInstanceStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
			t.Errorf("Expected ErrNoData, got: %v", err)
		}
	}()
	for range ch {
	}
}
