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

func TestAuroraStatCommitLatencyCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	mock.ExpectQuery(sanitizeQuery(auroraStatCommitLatencyQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"datid", "datname", "latency"}).
			AddRow("16401", "mydb", 229556).
			AddRow("69768", "postgres", 22011).
			AddRow("16384", "rdsadmin", 654387789))

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatCommitLatencyCollector{excludeDatabases: []string{"rdsadmin"}}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling Update: %s", err)
		}
	}()

	got := 0
	for range ch {
		got++
	}
	convey.Convey("rdsadmin excluded → 2 metrics", t, func() {
		convey.So(got, convey.ShouldEqual, 2)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestAuroraStatCommitLatencyCollectorNotAurora(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatCommitLatencyCollector{}
		if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
			t.Errorf("Expected ErrNoData, got: %v", err)
		}
	}()
	for range ch {
	}
}
