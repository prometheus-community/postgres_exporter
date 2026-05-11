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

func TestAuroraStatDMLActivityCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	cols := []string{
		"datid", "datname",
		"select_count", "select_latency_microsecs",
		"insert_count", "insert_latency_microsecs",
		"update_count", "update_latency_microsecs",
		"delete_count", "delete_latency_microsecs",
	}
	mock.ExpectQuery(sanitizeQuery(auroraStatDMLActivityQuery)).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("14007", "postgres", 178961, 66716329, 171065, 28876649, 519538, 1454579206167, 1, 53027).
			AddRow("16384", "rdsadmin", 2346623, 1211703821, 4297518, 817184554, 0, 0, 0, 0))

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDMLActivityCollector{excludeDatabases: []string{"rdsadmin"}}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling Update: %s", err)
		}
	}()

	got := 0
	for range ch {
		got++
	}
	convey.Convey("2 metrics × 4 operations × 1 db = 8", t, func() {
		convey.So(got, convey.ShouldEqual, 8)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestAuroraStatDMLActivityCollectorNotAurora(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDMLActivityCollector{}
		if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
			t.Errorf("Expected ErrNoData, got: %v", err)
		}
	}()
	for range ch {
	}
}
