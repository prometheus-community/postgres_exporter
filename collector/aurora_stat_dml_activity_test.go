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
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

var auroraStatDMLActivityColumns = []string{
	"datid",
	"datname",
	"select_count",
	"select_latency_microsecs",
	"insert_count",
	"insert_latency_microsecs",
	"update_count",
	"update_latency_microsecs",
	"delete_count",
	"delete_latency_microsecs",
}

func TestAuroraStatDMLActivityCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	// template1 has no DML activity and reports NULLs, so nothing is
	// emitted for it.
	rows := sqlmock.NewRows(auroraStatDMLActivityColumns).
		AddRow("1", "template1", nil, nil, nil, nil, nil, nil, nil, nil).
		AddRow("16401", "mydb", 200246, 64302436, 200036, 107101855, 600000, 83659417514, 0, 0)
	mock.ExpectQuery(sanitizeQuery(auroraStatDMLActivityQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatDMLActivityCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraStatDMLActivityCollector.Update: %s", err)
		}
	}()

	labels := func(operation string) labelMap {
		return labelMap{"datid": "16401", "datname": "mydb", "operation": operation}
	}
	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraStatDMLActivityOperations, MetricResult{labels: labels("select"), value: 200246, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityLatency, MetricResult{labels: labels("select"), value: 64.302436, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityOperations, MetricResult{labels: labels("insert"), value: 200036, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityLatency, MetricResult{labels: labels("insert"), value: 107.101855, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityOperations, MetricResult{labels: labels("update"), value: 600000, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityLatency, MetricResult{labels: labels("update"), value: 83659.417514, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityOperations, MetricResult{labels: labels("delete"), value: 0, metricType: dto.MetricType_COUNTER}},
		{auroraStatDMLActivityLatency, MetricResult{labels: labels("delete"), value: 0, metricType: dto.MetricType_COUNTER}},
	}
	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := <-ch
			convey.So(m.Desc(), convey.ShouldEqual, expect.desc)
			convey.So(readMetric(m), convey.ShouldResemble, expect.result)
		}
		_, more := <-ch
		convey.So(more, convey.ShouldBeFalse)
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestAuroraStatDMLActivityCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatDMLActivityCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData on non-Aurora server, got: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("Expected no metrics on non-Aurora server, got %d", len(ch))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestAuroraStatDMLActivityCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	mock.ExpectQuery(sanitizeQuery(auroraStatDMLActivityQuery)).WillReturnRows(sqlmock.NewRows(auroraStatDMLActivityColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatDMLActivityCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
