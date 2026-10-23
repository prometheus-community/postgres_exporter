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

var auroraStatSystemWaitsColumns = []string{
	"wait_event_type",
	"wait_event",
	"waits",
	"wait_time",
}

func TestAuroraStatSystemWaitsCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	// wait_time is in microseconds.
	rows := sqlmock.NewRows(auroraStatSystemWaitsColumns).
		AddRow("LWLock", "XidGenLock", 11, 12864).
		AddRow("IO", "DataFileRead", 53171, 23641847)
	mock.ExpectQuery(sanitizeQuery(auroraStatSystemWaitsQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraStatSystemWaitsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraStatSystemWaitsCollector.Update: %s", err)
		}
	}()

	lwlock := labelMap{"wait_event_type": "LWLock", "wait_event": "XidGenLock"}
	io := labelMap{"wait_event_type": "IO", "wait_event": "DataFileRead"}
	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraStatSystemWaitsWaits, MetricResult{labels: lwlock, value: 11, metricType: dto.MetricType_COUNTER}},
		{auroraStatSystemWaitsWaitTime, MetricResult{labels: lwlock, value: 0.012864, metricType: dto.MetricType_COUNTER}},
		{auroraStatSystemWaitsWaits, MetricResult{labels: io, value: 53171, metricType: dto.MetricType_COUNTER}},
		{auroraStatSystemWaitsWaitTime, MetricResult{labels: io, value: 23.641847, metricType: dto.MetricType_COUNTER}},
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

func TestAuroraStatSystemWaitsCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatSystemWaitsCollector{}
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

func TestAuroraStatSystemWaitsCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}
	mock.ExpectQuery(sanitizeQuery(auroraStatSystemWaitsQuery)).WillReturnRows(sqlmock.NewRows(auroraStatSystemWaitsColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraStatSystemWaitsCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
