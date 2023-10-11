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
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

var queryWithFlushedLSN = fmt.Sprintf(pgStatWalReceiverQueryTemplate, "(flushed_lsn - '0/0') % (2^52)::bigint as flushed_lsn,\n")
var queryWithNoFlushedLSN = fmt.Sprintf(pgStatWalReceiverQueryTemplate, "")

func TestPGStatWalReceiverCollectorWithFlushedLSN(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}
	infoSchemaColumns := []string{
		"column_name",
	}

	infoSchemaRows := sqlmock.NewRows(infoSchemaColumns).
		AddRow(
			"flushed_lsn",
		)

	mock.ExpectQuery(sanitizeQuery(pgStatWalColumnQuery)).WillReturnRows(infoSchemaRows)

	columns := []string{
		"upstream_host",
		"slot_name",
		"status",
		"receive_start_lsn",
		"receive_start_tli",
		"flushed_lsn",
		"received_tli",
		"last_msg_send_time",
		"last_msg_receipt_time",
		"latest_end_lsn",
		"latest_end_time",
		"upstream_node",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(
			"foo",
			"bar",
			"stopping",
			int64(1200668684563608),
			1687321285,
			int64(1200668684563609),
			1687321280,
			1687321275,
			1687321276,
			int64(1200668684563610),
			1687321277,
			5,
		)

	mock.ExpectQuery(sanitizeQuery(queryWithFlushedLSN)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatWalReceiverCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PgStatWalReceiverCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1200668684563608, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1687321285, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1200668684563609, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1687321280, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1687321275, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1687321276, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1200668684563610, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 1687321277, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "stopping"}, value: 5, metricType: dto.MetricType_GAUGE},
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

func TestPGStatWalReceiverCollectorWithNoFlushedLSN(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}
	infoSchemaColumns := []string{
		"column_name",
	}

	infoSchemaRows := sqlmock.NewRows(infoSchemaColumns)

	mock.ExpectQuery(sanitizeQuery(pgStatWalColumnQuery)).WillReturnRows(infoSchemaRows)

	columns := []string{
		"upstream_host",
		"slot_name",
		"status",
		"receive_start_lsn",
		"receive_start_tli",
		"received_tli",
		"last_msg_send_time",
		"last_msg_receipt_time",
		"latest_end_lsn",
		"latest_end_time",
		"upstream_node",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(
			"foo",
			"bar",
			"starting",
			int64(1200668684563608),
			1687321285,
			1687321280,
			1687321275,
			1687321276,
			int64(1200668684563610),
			1687321277,
			5,
		)
	mock.ExpectQuery(sanitizeQuery(queryWithNoFlushedLSN)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatWalReceiverCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PgStatWalReceiverCollector.Update: %s", err)
		}
	}()
	expected := []MetricResult{
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1200668684563608, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1687321285, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1687321280, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1687321275, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1687321276, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1200668684563610, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 1687321277, metricType: dto.MetricType_COUNTER},
		{labels: labelMap{"upstream_host": "foo", "slot_name": "bar", "status": "starting"}, value: 5, metricType: dto.MetricType_GAUGE},
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
