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

func TestPGStatIOCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	columns := []string{
		"backend_type",
		"object",
		"context",
		"reads",
		"read_time",
		"writes",
		"write_time",
		"writebacks",
		"writeback_time",
		"extends",
		"extend_time",
		"hits",
		"evictions",
		"reuses",
		"fsyncs",
		"fsync_time"}

	rows := sqlmock.NewRows(columns).
		AddRow("vacuum", "relation", "vacuum",
			45, 3466.5,
			12, 3467.67,
			2, 4.5,
			1, 1.2,
			1234, 3, 56,
			1235, 12.0)

	mock.ExpectQuery(sanitizeQuery(statIOQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := StatIOCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 45},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 3466.5},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 12},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 3467.67},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 2},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 4.5},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 1},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 1.2},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 1234},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 3},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 56},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 1235},
		{labels: labelMap{"backend_type": "vacuum", "object": "relation", "context": "vacuum"}, metricType: dto.MetricType_COUNTER, value: 12.0},
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

func TestPGStatIOCollectorNull(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	columns := []string{
		"backend_type",
		"object",
		"context",
		"reads",
		"read_time",
		"writes",
		"write_time",
		"writebacks",
		"writeback_time",
		"extends",
		"extend_time",
		"hits",
		"evictions",
		"reuses",
		"fsyncs",
		"fsync_time"}

	rows := sqlmock.NewRows(columns).AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(sanitizeQuery(statIOQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := StatIOCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	// since we have no expected metrics, wait for the channel to close and then `Update` will have run.
	<-ch

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
