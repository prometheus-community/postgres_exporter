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
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPgWALCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"segments", "size", "lsn_location_bytes"}
	rows := sqlmock.NewRows(columns).
		AddRow(47, 788529152, 123456789)
	mock.ExpectQuery(sanitizeQuery(pgWALQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGWALCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGWALCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 47, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 788529152, metricType: dto.MetricType_GAUGE},
		{labels: labelMap{}, value: 123456789, metricType: dto.MetricType_GAUGE},
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

func TestPgWALCollectorZeroSegments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"segments", "size", "lsn_location_bytes"}
	rows := sqlmock.NewRows(columns).
		AddRow(0, nil, 0)
	mock.ExpectQuery(sanitizeQuery(pgWALQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGWALCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGWALCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 0, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison (zero segments)", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPgWALCollectorErrors(t *testing.T) {
	notSupported := &pq.Error{Code: "0A000", Message: "pg_ls_waldir() is currently not supported for Aurora"}
	permissionDenied := &pq.Error{Code: "42501", Message: "permission denied for function pg_ls_waldir"}

	for _, tc := range []struct {
		name     string
		isAurora bool
		err      error
		wantErr  error
	}{
		{name: "Aurora rejects pg_ls_waldir", isAurora: true, err: notSupported, wantErr: ErrNoData},
		{name: "not supported outside Aurora", isAurora: false, err: notSupported, wantErr: notSupported},
		{name: "other errors on Aurora", isAurora: true, err: permissionDenied, wantErr: permissionDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error opening a stub db connection: %s", err)
			}
			defer db.Close()

			inst := &instance{db: db, isAurora: tc.isAurora}
			mock.ExpectQuery(sanitizeQuery(pgWALQuery)).WillReturnError(tc.err)

			ch := make(chan prometheus.Metric, 2)
			err = PGWALCollector{}.Update(context.Background(), inst, ch)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("PGWALCollector.Update returned %v, want %v", err, tc.wantErr)
			}
			if len(ch) != 0 {
				t.Errorf("Expected no metrics, got %d", len(ch))
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled exceptions: %s", err)
			}
		})
	}
}
