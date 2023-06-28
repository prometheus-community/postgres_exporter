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
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgXidCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}
	columns := []string{
		"current",
		"xmin",
		"xmin_age",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(22, 25, 30)

	mock.ExpectQuery(sanitizeQuery(xidQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGXidCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGXidCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{}, value: 22, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 25, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 30, metricType: dto.MetricType_GAUGE},
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

func TestPgNanCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}
	columns := []string{
		"current",
		"xmin",
		"xmin_age",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(math.NaN(), math.NaN(), math.NaN())

	mock.ExpectQuery(sanitizeQuery(xidQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGXidCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGXidCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{}, value: math.NaN(), metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: math.NaN(), metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: math.NaN(), metricType: dto.MetricType_GAUGE},
	}
	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)

			convey.So(expect.labels, convey.ShouldResemble, m.labels)
			convey.So(math.IsNaN(m.value), convey.ShouldResemble, math.IsNaN(expect.value))
			convey.So(expect.metricType, convey.ShouldEqual, m.metricType)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
