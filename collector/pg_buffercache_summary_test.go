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

func TestBuffercacheSummaryCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("16.0.0")}

	columns := []string{
		"buffers_used",
		"buffers_unused",
		"buffers_dirty",
		"buffers_pinned",
		"usagecount_avg"}

	rows := sqlmock.NewRows(columns).AddRow(123, 456, 789, 234, 56.6778)

	mock.ExpectQuery(sanitizeQuery(buffercacheQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := BuffercacheSummaryCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatStatementsCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 56.6778},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 123},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 456},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 789},
		{labels: labelMap{}, metricType: dto.MetricType_GAUGE, value: 234},
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
