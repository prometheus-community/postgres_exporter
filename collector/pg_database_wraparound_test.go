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
	"database/sql"
	"io"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
)

type wraparoundTestCollector struct {
	metrics []prometheus.Metric
}

func (c wraparoundTestCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc()
	}
}

func (c wraparoundTestCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func TestPGDatabaseWraparoundCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}
	columns := []string{
		"datname",
		"age_datfrozenxid",
		"age_datminmxid",
	}
	rows := sqlmock.NewRows(columns).
		AddRow("newreddit", 87126426, 17)

	mock.ExpectQuery(sanitizeQuery(databaseWraparoundQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric, 4)
	c := PGDatabaseWraparoundCollector{}
	if err := c.Update(context.Background(), inst, ch); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	close(ch)
	metrics := make([]prometheus.Metric, 0, len(ch))
	for metric := range ch {
		metrics = append(metrics, metric)
	}
	registry := prometheus.NewRegistry()
	if err := registry.Register(wraparoundTestCollector{metrics}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	expected := map[string]struct {
		help  string
		value float64
	}{
		"pg_database_wraparound_age_datfrozenxid": {
			"Age of pg_database.datfrozenxid in transaction IDs.", 87126426,
		},
		"pg_database_wraparound_age_datfrozenxid_seconds": {
			"DEPRECATED: Age of pg_database.datfrozenxid in transaction IDs, not seconds. Use pg_database_wraparound_age_datfrozenxid instead.", 87126426,
		},
		"pg_database_wraparound_age_datminmxid": {
			"Age of pg_database.datminmxid in MultiXact IDs.", 17,
		},
		"pg_database_wraparound_age_datminmxid_seconds": {
			"DEPRECATED: Age of pg_database.datminmxid in MultiXact IDs, not seconds. Use pg_database_wraparound_age_datminmxid instead.", 17,
		},
	}
	if len(families) != len(expected) {
		t.Fatalf("got %d metric families, want %d", len(families), len(expected))
	}
	for _, family := range families {
		want, ok := expected[family.GetName()]
		if !ok {
			t.Errorf("unexpected metric family %q", family.GetName())
			continue
		}
		if family.GetHelp() != want.help {
			t.Errorf("%s help = %q, want %q", family.GetName(), family.GetHelp(), want.help)
		}
		if family.GetType().String() != "GAUGE" {
			t.Errorf("%s type = %s, want GAUGE", family.GetName(), family.GetType())
		}
		if len(family.Metric) != 1 {
			t.Errorf("%s has %d samples, want 1", family.GetName(), len(family.Metric))
			continue
		}
		sample := family.Metric[0]
		if sample.GetGauge().GetValue() != want.value {
			t.Errorf("%s value = %v, want %v", family.GetName(), sample.GetGauge().GetValue(), want.value)
		}
		if len(sample.Label) != 1 || sample.Label[0].GetName() != "datname" || sample.Label[0].GetValue() != "newreddit" {
			t.Errorf("%s labels = %v, want datname=newreddit", family.GetName(), sample.Label)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGDatabaseWraparoundCollectorNullAge(t *testing.T) {
	for _, tt := range []struct {
		name, datname string
		xid, mxid     any
	}{
		{"null database", "", 1, 2},
		{"null XID age", "postgres", nil, 2},
		{"null MultiXact age", "postgres", 1, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New() error = %v", err)
			}
			defer db.Close()
			var datname any = tt.datname
			if tt.name == "null database" {
				datname = sql.NullString{}
			}
			rows := sqlmock.NewRows([]string{"datname", "age_datfrozenxid", "age_datminmxid"}).AddRow(datname, tt.xid, tt.mxid)
			mock.ExpectQuery(sanitizeQuery(databaseWraparoundQuery)).WillReturnRows(rows)
			ch := make(chan prometheus.Metric, 4)
			c := PGDatabaseWraparoundCollector{log: slog.New(slog.NewTextHandler(io.Discard, nil))}
			if err := c.Update(context.Background(), &instance{db: db}, ch); err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if len(ch) != 0 {
				t.Errorf("got %d metrics for null row, want none", len(ch))
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled SQL expectation: %v", err)
			}
		})
	}
}
