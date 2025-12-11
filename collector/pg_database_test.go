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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
	"github.com/smartystreets/goconvey/convey"
)

// We ensure that when the database respond after a long time
// The collection process still occurs in a predictable manner
// Will avoid accumulation of queries on a completely frozen DB
func TestPGDatabaseTimeout(t *testing.T) {

	timeoutForQuery := time.Duration(100 * time.Millisecond)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"pg_roles.rolname", "pg_roles.rolconnlimit"}
	rows := sqlmock.NewRows(columns).AddRow("role1", 2)
	mock.ExpectQuery(pgRolesConnectionLimitsQuery).
		WillDelayFor(30 * time.Second).
		WillReturnRows(rows)

	log_config := promslog.Config{}

	logger := promslog.New(&log_config)

	c, err := NewPostgresCollector(logger, []string{}, "postgresql://local", []string{}, CollectionTimeout(timeoutForQuery.String()))
	if err != nil {
		t.Fatalf("error creating NewPostgresCollector: %s", err)
	}
	collector_config := collectorConfig{
		logger:           logger,
		excludeDatabases: []string{},
	}

	collector, err := NewPGRolesCollector(collector_config)
	if err != nil {
		t.Fatalf("error creating collector: %s", err)
	}
	c.Collectors["test"] = collector
	c.instance = inst

	ch := make(chan prometheus.Metric)
	defer close(ch)

	go func() {
		for {
			<-ch
			time.Sleep(1 * time.Millisecond)
		}
	}()

	startTime := time.Now()
	c.collectFromConnection(inst, ch)
	elapsed := time.Since(startTime)

	if elapsed <= timeoutForQuery {
		t.Errorf("elapsed time was %v, should be bigger than timeout=%v", elapsed, timeoutForQuery)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGDatabaseCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	mock.ExpectQuery(sanitizeQuery(pgDatabaseQuery)).WillReturnRows(sqlmock.NewRows([]string{"datname", "datconnlimit"}).
		AddRow("postgres", 15))

	mock.ExpectQuery(sanitizeQuery(pgDatabaseSizeQuery)).WithArgs("postgres").WillReturnRows(sqlmock.NewRows([]string{"pg_database_size"}).
		AddRow(1024))

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGDatabaseCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres"}, value: 15, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"datname": "postgres"}, value: 1024, metricType: dto.MetricType_GAUGE},
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

// TODO add a null db test

func TestPGDatabaseCollectorNullMetric(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	mock.ExpectQuery(sanitizeQuery(pgDatabaseQuery)).WillReturnRows(sqlmock.NewRows([]string{"datname", "datconnlimit"}).
		AddRow("postgres", nil))

	mock.ExpectQuery(sanitizeQuery(pgDatabaseSizeQuery)).WithArgs("postgres").WillReturnRows(sqlmock.NewRows([]string{"pg_database_size"}).
		AddRow(nil))

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGDatabaseCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datname": "postgres"}, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"datname": "postgres"}, value: 0, metricType: dto.MetricType_GAUGE},
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
