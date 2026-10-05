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

var auroraGlobalDBInstanceStatusColumns = []string{
	"server_id",
	"aws_region",
	"visibility_lag_in_msec",
}

func TestAuroraGlobalDBInstanceStatusCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	// The writer reports NULL visibility lag, so nothing is emitted for it.
	// The same server_id can appear in several Regions.
	rows := sqlmock.NewRows(auroraGlobalDBInstanceStatusColumns).
		AddRow("instance-01", "eu-west-1", nil).
		AddRow("instance-02", "eu-west-1", 6).
		AddRow("global-instance-1", "eu-central-1", 996).
		AddRow("global-instance-1", "eu-west-2", 14)
	mock.ExpectQuery(sanitizeQuery(auroraGlobalDBInstanceStatusQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := AuroraGlobalDBInstanceStatusCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling AuroraGlobalDBInstanceStatusCollector.Update: %s", err)
		}
	}()

	expected := []struct {
		desc   *prometheus.Desc
		result MetricResult
	}{
		{auroraGlobalDBInstanceStatusVisibilityLag, MetricResult{labels: labelMap{"server_id": "instance-02", "aws_region": "eu-west-1"}, value: 0.006, metricType: dto.MetricType_GAUGE}},
		{auroraGlobalDBInstanceStatusVisibilityLag, MetricResult{labels: labelMap{"server_id": "global-instance-1", "aws_region": "eu-central-1"}, value: 0.996, metricType: dto.MetricType_GAUGE}},
		{auroraGlobalDBInstanceStatusVisibilityLag, MetricResult{labels: labelMap{"server_id": "global-instance-1", "aws_region": "eu-west-2"}, value: 0.014, metricType: dto.MetricType_GAUGE}},
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

func TestAuroraGlobalDBInstanceStatusCollectorNotAurora(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	// No query is expected: the collector must not touch the database when
	// the server is not Aurora.
	inst := &instance{db: db, isAurora: false}

	ch := make(chan prometheus.Metric, 1)
	c := AuroraGlobalDBInstanceStatusCollector{}
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

func TestAuroraGlobalDBInstanceStatusCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, isAurora: true}

	mock.ExpectQuery(sanitizeQuery(auroraGlobalDBInstanceStatusQuery)).WillReturnRows(sqlmock.NewRows(auroraGlobalDBInstanceStatusColumns))

	ch := make(chan prometheus.Metric, 1)
	c := AuroraGlobalDBInstanceStatusCollector{}
	if err := c.Update(context.Background(), inst, ch); err != ErrNoData {
		t.Errorf("Expected ErrNoData, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
