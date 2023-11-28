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
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatDatabaseCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datid",
		"datname",
		"numbackends",
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"blk_read_time",
		"blk_write_time",
		"active_time",
		"stats_reset",
	}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			"pid",
			"postgres",
			354,
			4945,
			289097744,
			1242257,
			int64(3275602074),
			89320867,
			450139,
			2034563757,
			0,
			int64(2725688749),
			23,
			52,
			74,
			925,
			16,
			823,
			33,
			srT)

	mock.ExpectQuery(sanitizeQuery(statDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatDatabaseCollector{
			log: log.With(log.NewNopLogger(), "collector", "pg_stat_database"),
		}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_GAUGE, value: 354},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2725688749},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 23},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 52},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 74},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 925},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 16},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 823},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0.033},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1685059842},
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

func TestPGStatDatabaseCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}
	inst := &instance{db: db}

	columns := []string{
		"datid",
		"datname",
		"numbackends",
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"blk_read_time",
		"blk_write_time",
		"active_time",
		"stats_reset",
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			nil,
			"postgres",
			354,
			4945,
			289097744,
			1242257,
			int64(3275602074),
			89320867,
			450139,
			2034563757,
			0,
			int64(2725688749),
			23,
			52,
			74,
			925,
			16,
			823,
			32,
			srT).
		AddRow(
			"pid",
			"postgres",
			354,
			4945,
			289097744,
			1242257,
			int64(3275602074),
			89320867,
			450139,
			2034563757,
			0,
			int64(2725688749),
			23,
			52,
			74,
			925,
			16,
			823,
			32,
			srT)
	mock.ExpectQuery(sanitizeQuery(statDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatDatabaseCollector{
			log: log.With(log.NewNopLogger(), "collector", "pg_stat_database"),
		}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_GAUGE, value: 354},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2725688749},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 23},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 52},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 74},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 925},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 16},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 823},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0.032},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1685059842},
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
func TestPGStatDatabaseCollectorRowLeakTest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datid",
		"datname",
		"numbackends",
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"blk_read_time",
		"blk_write_time",
		"active_time",
		"stats_reset",
	}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			"pid",
			"postgres",
			354,
			4945,
			289097744,
			1242257,
			int64(3275602074),
			89320867,
			450139,
			2034563757,
			0,
			int64(2725688749),
			23,
			52,
			74,
			925,
			16,
			823,
			14,
			srT).
		AddRow(
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		).
		AddRow(
			"pid",
			"postgres",
			355,
			4946,
			289097745,
			1242258,
			int64(3275602075),
			89320868,
			450140,
			2034563758,
			1,
			int64(2725688750),
			24,
			53,
			75,
			926,
			17,
			824,
			15,
			srT)
	mock.ExpectQuery(sanitizeQuery(statDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatDatabaseCollector{
			log: log.With(log.NewNopLogger(), "collector", "pg_stat_database"),
		}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_GAUGE, value: 354},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2725688749},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 23},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 52},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 74},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 925},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 16},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 823},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0.014},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1685059842},

		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_GAUGE, value: 355},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 4946},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 289097745},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1242258},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 3275602075},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 89320868},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 450140},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2034563758},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2725688750},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 24},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 53},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 75},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 926},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 17},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 824},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0.015},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1685059842},
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

func TestPGStatDatabaseCollectorTestNilStatReset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{
		"datid",
		"datname",
		"numbackends",
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"blk_read_time",
		"blk_write_time",
		"active_time",
		"stats_reset",
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			"pid",
			"postgres",
			354,
			4945,
			289097744,
			1242257,
			int64(3275602074),
			89320867,
			450139,
			2034563757,
			0,
			int64(2725688749),
			23,
			52,
			74,
			925,
			16,
			823,
			7,
			nil)

	mock.ExpectQuery(sanitizeQuery(statDatabaseQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatDatabaseCollector{
			log: log.With(log.NewNopLogger(), "collector", "pg_stat_database"),
		}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatDatabaseCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_GAUGE, value: 354},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 2725688749},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 23},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 52},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 74},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 925},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 16},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 823},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0.007},
		{labels: labelMap{"datid": "pid", "datname": "postgres"}, metricType: dto.MetricType_COUNTER, value: 0},
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
