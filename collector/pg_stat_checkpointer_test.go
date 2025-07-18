// Copyright 2024 The Prometheus Authors
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
	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatCheckpointerCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	columns := []string{
		"num_timed",
		"num_requested",
		"restartpoints_timed",
		"restartpoints_req",
		"restartpoints_done",
		"write_time",
		"sync_time",
		"buffers_written",
		"stats_reset"}

	srT, err := time.Parse("2006-01-02 15:04:05.00000-07", "2023-05-25 17:10:42.81132-07")
	if err != nil {
		t.Fatalf("Error parsing time: %s", err)
	}

	rows := sqlmock.NewRows(columns).
		AddRow(354, 4945, 289097744, 1242257, int64(3275602074), 89320867, 450139, 2034563757, srT)
	mock.ExpectQuery(sanitizeQuery(statCheckpointerQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatCheckpointerCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatCheckpointerCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 354},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 4945},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 289097744},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1242257},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 3275602074},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 89320867},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 450139},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 2034563757},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 1685059842},
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

func TestPGStatCheckpointerCollectorNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db, version: semver.MustParse("17.0.0")}

	columns := []string{
		"num_timed",
		"num_requested",
		"restartpoints_timed",
		"restartpoints_req",
		"restartpoints_done",
		"write_time",
		"sync_time",
		"buffers_written",
		"stats_reset"}

	rows := sqlmock.NewRows(columns).
		AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mock.ExpectQuery(sanitizeQuery(statCheckpointerQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatCheckpointerCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatCheckpointerCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: labelMap{}, metricType: dto.MetricType_COUNTER, value: 0},
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
