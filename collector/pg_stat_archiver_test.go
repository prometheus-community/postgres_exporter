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
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatArchiverCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{"archived_count", "failed_count", "last_archive_age"}).
		AddRow(42, 7, 13.5)
	mock.ExpectQuery(sanitizeQuery(statArchiverQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatArchiverCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatArchiverCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 42},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 7},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 13.5},
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

func TestPGStatArchiverCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	rows := sqlmock.NewRows([]string{"archived_count", "failed_count", "last_archive_age"}).
		AddRow(nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statArchiverQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatArchiverCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatArchiverCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 0},
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

func TestPGStatArchiverCollectorBeforePostgres94(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("9.3.0")}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatArchiverCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatArchiverCollector.Update: %s", err)
		}
	}()

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted for PostgreSQL 9.3: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
