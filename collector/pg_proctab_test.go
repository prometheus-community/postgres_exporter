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
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGProctabCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	rows := sqlmock.NewRows([]string{"memused", "memfree", "memshared", "membuffers", "memcached", "swapused"}).
		AddRow(123, 456, 789, 234, 567, 89)
	mock.ExpectQuery(sanitizeQuery(memoryQuery)).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"load1"}).AddRow(123.456)
	mock.ExpectQuery(sanitizeQuery(load1Query)).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"user", "nice", "system", "idle", "iowait"}).AddRow(
		345, 678, 9, 1234, 56)
	mock.ExpectQuery(sanitizeQuery(cpuQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGProctabCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGProctabCollector .Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 123 * 1024, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 456 * 1024, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 789 * 1024, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 234 * 1024, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 567 * 1024, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 89 * 1024, metricType: dto.MetricType_GAUGE},
		// load
		{labels: labelMap{}, value: 123.456, metricType: dto.MetricType_GAUGE},
		// cpu
		{labels: labelMap{}, value: 345, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 678, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 9, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 1234, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{}, value: 56, metricType: dto.MetricType_COUNTER},
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
