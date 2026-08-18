// Copyright 2026 The Prometheus Authors
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

func TestPGSequenceOverflowCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	seqColumns := []string{
		"sequence",
		"sequence_datatype",
		"owned_by",
		"column_datatype",
		"schemaname",
		"last_sequence_value",
		"sequence_ratio",
		"column_ratio",
	}

	// Row 1: integer sequence on an integer column, 50% consumed.
	// Row 2: bigint sequence on an integer column — sequence type has far more
	// headroom than the column type, so sequence_ratio is near 0 while column_ratio is 0.5.
	rows := sqlmock.NewRows(seqColumns).
		AddRow("my_table_id_seq", "integer", "my_table.id", "integer", "public", 1073741823, 0.5, 0.5).
		AddRow("other_table_id_seq", "bigint", "other_table.id", "integer", "public", 1073741823, 0.25, 0.5)

	mock.ExpectQuery(sanitizeQuery(sequenceOverflowQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSequenceOverflowCollector{}
		if err := c.updateDatabase(context.Background(), db, "mydb", ch); err != nil {
			t.Errorf("Error calling updateDatabase: %s", err)
		}
	}()

	labels1 := labelMap{
		"datname":           "mydb",
		"schemaname":        "public",
		"sequence":          "my_table_id_seq",
		"sequence_datatype": "integer",
		"owned_by":          "my_table.id",
		"column_datatype":   "integer",
	}
	labels2 := labelMap{
		"datname":           "mydb",
		"schemaname":        "public",
		"sequence":          "other_table_id_seq",
		"sequence_datatype": "bigint",
		"owned_by":          "other_table.id",
		"column_datatype":   "integer",
	}

	expected := []MetricResult{
		{labels: labels1, value: 1073741823, metricType: dto.MetricType_GAUGE},
		{labels: labels1, value: 0.5, metricType: dto.MetricType_GAUGE},
		{labels: labels1, value: 0.5, metricType: dto.MetricType_GAUGE},
		{labels: labels2, value: 1073741823, metricType: dto.MetricType_GAUGE},
		{labels: labels2, value: 0.25, metricType: dto.MetricType_GAUGE},
		{labels: labels2, value: 0.5, metricType: dto.MetricType_GAUGE},
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

func TestPGSequenceOverflowCollectorNullValues(t *testing.T) {
	// Sequences that have never been used: COALESCE in the query converts NULL
	// from pg_sequence_last_value to 0, so all numeric columns return 0.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	seqColumns := []string{
		"sequence",
		"sequence_datatype",
		"owned_by",
		"column_datatype",
		"schemaname",
		"last_sequence_value",
		"sequence_ratio",
		"column_ratio",
	}

	rows := sqlmock.NewRows(seqColumns).
		AddRow("unused_seq", "integer", "some_table.id", "integer", "public", 0, 0, 0)

	mock.ExpectQuery(sanitizeQuery(sequenceOverflowQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGSequenceOverflowCollector{}
		if err := c.updateDatabase(context.Background(), db, "mydb", ch); err != nil {
			t.Errorf("Error calling updateDatabase: %s", err)
		}
	}()

	labels := labelMap{
		"datname":           "mydb",
		"schemaname":        "public",
		"sequence":          "unused_seq",
		"sequence_datatype": "integer",
		"owned_by":          "some_table.id",
		"column_datatype":   "integer",
	}

	expected := []MetricResult{
		{labels: labels, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labels, value: 0, metricType: dto.MetricType_GAUGE},
		{labels: labels, value: 0, metricType: dto.MetricType_GAUGE},
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
