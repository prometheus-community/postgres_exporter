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
	"github.com/prometheus/common/promslog"
)

func TestPGSettingNormaliseUnit(t *testing.T) {
	tests := []struct {
		name      string
		setting   pgSetting
		wantValue float64
		wantUnit  string
		wantErr   string
	}{
		{
			name:      "seconds",
			setting:   pgSetting{name: "seconds_fixture_metric", setting: "5", unit: "s", vartype: "integer"},
			wantValue: 5,
			wantUnit:  "seconds",
		},
		{
			name:      "milliseconds",
			setting:   pgSetting{name: "milliseconds_fixture_metric", setting: "5000", unit: "ms", vartype: "integer"},
			wantValue: 5,
			wantUnit:  "seconds",
		},
		{
			name:      "8kB",
			setting:   pgSetting{name: "eight_kb_fixture_metric", setting: "17", unit: "8kB", vartype: "integer"},
			wantValue: 139264,
			wantUnit:  "bytes",
		},
		{
			name:      "special minus one",
			setting:   pgSetting{name: "special_minus_one_value", setting: "-1", unit: "d", vartype: "integer"},
			wantValue: -1,
			wantUnit:  "seconds",
		},
		{
			name:      "unknown unit",
			setting:   pgSetting{name: "unknown_unit", setting: "10", unit: "nonexistent", vartype: "integer"},
			wantValue: 10,
			wantErr:   `unknown unit for runtime variable: "nonexistent"`,
		},
		{
			name:      "value with unit suffix",
			setting:   pgSetting{name: "aurora_value", setting: "16MB", unit: "MB", vartype: "integer"},
			wantValue: 16777216,
			wantUnit:  "bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotUnit, err := tt.setting.normaliseUnit()
			if gotValue != tt.wantValue {
				t.Fatalf("normaliseUnit() value = %v, want %v", gotValue, tt.wantValue)
			}
			if gotUnit != tt.wantUnit {
				t.Fatalf("normaliseUnit() unit = %q, want %q", gotUnit, tt.wantUnit)
			}
			if tt.wantErr == "" && err != nil {
				t.Fatalf("normaliseUnit() unexpected error: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("normaliseUnit() expected error %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("normaliseUnit() error = %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestPGSettingMetric(t *testing.T) {
	tests := []struct {
		name       string
		setting    pgSetting
		wantDesc   string
		wantMetric float64
	}{
		{
			name:       "integer seconds",
			setting:    pgSetting{name: "seconds_fixture_metric", setting: "5", unit: "s", vartype: "integer"},
			wantDesc:   `Desc{fqName: "pg_settings_seconds_fixture_metric_seconds", help: "Server Parameter: seconds_fixture_metric [Units converted to seconds.]", constLabels: {}, variableLabels: {}}`,
			wantMetric: 5,
		},
		{
			name:       "bool on",
			setting:    pgSetting{name: "bool_on_fixture_metric", setting: "on", vartype: "bool"},
			wantDesc:   `Desc{fqName: "pg_settings_bool_on_fixture_metric", help: "Server Parameter: bool_on_fixture_metric", constLabels: {}, variableLabels: {}}`,
			wantMetric: 1,
		},
		{
			name:       "sanitized name",
			setting:    pgSetting{name: "rds.rds-superuser-reserved-connections", setting: "2", vartype: "integer"},
			wantDesc:   `Desc{fqName: "pg_settings_rds_rds_superuser_reserved_connections", help: "Server Parameter: rds.rds-superuser-reserved-connections", constLabels: {}, variableLabels: {}}`,
			wantMetric: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := tt.setting.metric()
			if err != nil {
				t.Fatalf("metric() unexpected error: %v", err)
			}
			got := &dto.Metric{}
			if err := metric.Write(got); err != nil {
				t.Fatalf("Write() unexpected error: %v", err)
			}
			if metric.Desc().String() != tt.wantDesc {
				t.Fatalf("metric() desc = %q, want %q", metric.Desc().String(), tt.wantDesc)
			}
			if got.GetGauge().GetValue() != tt.wantMetric {
				t.Fatalf("metric() value = %v, want %v", got.GetGauge().GetValue(), tt.wantMetric)
			}
		})
	}
}

func TestPGSettingsCollectorUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}
	rows := sqlmock.NewRows([]string{"name", "setting", "unit", "short_desc", "vartype"}).
		AddRow("shared_buffers", "128", "8kB", "Sets the number of shared memory buffers used by the server.", "integer").
		AddRow("track_counts", "on", "", "Collects statistics on database activity.", "bool").
		AddRow("bad_setting", "not-a-number", "", "Bad setting.", "integer")
	mock.ExpectQuery(sanitizeQuery(pgSettingsQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		collector := PGSettingsCollector{log: promslog.NewNopLogger()}
		if err := collector.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGSettingsCollector.Update: %s", err)
		}
	}()

	tests := []struct {
		wantDesc  string
		wantValue float64
	}{
		{
			wantDesc:  `Desc{fqName: "pg_settings_shared_buffers_bytes", help: "Server Parameter: shared_buffers [Units converted to bytes.]", constLabels: {}, variableLabels: {}}`,
			wantValue: 1048576,
		},
		{
			wantDesc:  `Desc{fqName: "pg_settings_track_counts", help: "Server Parameter: track_counts", constLabels: {}, variableLabels: {}}`,
			wantValue: 1,
		},
	}

	for _, tt := range tests {
		metric := <-ch
		got := &dto.Metric{}
		if err := metric.Write(got); err != nil {
			t.Fatalf("Write() unexpected error: %v", err)
		}
		if metric.Desc().String() != tt.wantDesc {
			t.Fatalf("metric desc = %q, want %q", metric.Desc().String(), tt.wantDesc)
		}
		if got.GetGauge().GetValue() != tt.wantValue {
			t.Fatalf("metric value = %v, want %v", got.GetGauge().GetValue(), tt.wantValue)
		}
	}

	if metric, ok := <-ch; ok {
		t.Fatalf("unexpected metric emitted after bad setting was skipped: %s", metric.Desc())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}
