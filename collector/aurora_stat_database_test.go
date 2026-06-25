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

func TestAuroraStatDatabaseCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	columns := []string{
		"datid", "datname",
		"storage_blks_read", "orcache_blks_hit", "local_blks_read",
		"storage_blk_read_time", "orcache_blk_read_time", "local_blk_read_time",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("14717", "postgres", 623, 425, 0, 3254.914, 89.934, 0.0).
		AddRow("16384", "rdsadmin", 100, 50, 0, 200.0, 30.0, 0.0)
	mock.ExpectQuery(sanitizeQuery(auroraStatDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDatabaseCollector{excludeDatabases: []string{"rdsadmin"}}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling Update: %s", err)
		}
	}()

	got := 0
	for range ch {
		got++
	}
	convey.Convey("6 metrics × 1 non-excluded database = 6", t, func() {
		convey.So(got, convey.ShouldEqual, 6)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestAuroraStatDatabaseCollectorNotAurora(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDatabaseCollector{}
		if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
			t.Errorf("Expected ErrNoData, got: %v", err)
		}
	}()
	for range ch {
	}
}
