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

package exporter

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
)

func TestDBToFloat64Counter(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "largest exactly represented integer",
			input: int64(1<<53) - 1,
			want:  float64(int64(1<<53) - 1),
		},
		{
			name:  "wrap boundary",
			input: int64(1 << 53),
			want:  0,
		},
		{
			name:  "above wrap boundary",
			input: int64(1<<53) + 1,
			want:  1,
		},
		{
			name:  "negative value remains unchanged",
			input: int64(-1),
			want:  -1,
		},
		{
			name:  "non-integer conversion remains unchanged",
			input: 42.5,
			want:  42.5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := dbToFloat64Counter(test.input, promslog.NewNopLogger())
			if !ok {
				t.Fatal("conversion failed")
			}
			if got != test.want {
				t.Errorf("counter value = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNewExporterWrapLargeCounters(t *testing.T) {
	if exporter := NewExporter(nil, promslog.NewNopLogger()); !exporter.wrapLargeCounters {
		t.Error("large counter wrapping should be enabled by default")
	}

	if exporter := NewExporter(nil, promslog.NewNopLogger(), WrapLargeCounters(false)); exporter.wrapLargeCounters {
		t.Error("large counter wrapping should be disabled by option")
	}
}

func TestQueryNamespaceMappingWrapsLargeCounters(t *testing.T) {
	const largeInteger = int64(1<<53) + 1

	tests := []struct {
		name                 string
		wrapLargeCounters    bool
		expectedCounterValue float64
	}{
		{
			name:                 "enabled",
			wrapLargeCounters:    true,
			expectedCounterValue: 1,
		},
		{
			name:                 "disabled",
			wrapLargeCounters:    false,
			expectedCounterValue: float64(largeInteger),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("creating mock database: %v", err)
			}
			t.Cleanup(func() {
				_ = db.Close()
			})

			mock.ExpectQuery(`SELECT \* FROM test_namespace;`).WillReturnRows(
				sqlmock.NewRows([]string{"counter", "gauge"}).AddRow(largeInteger, largeInteger),
			)

			server := &Server{
				db:                db,
				logger:            promslog.NewNopLogger(),
				wrapLargeCounters: test.wrapLargeCounters,
			}
			mapping := MetricMapNamespace{
				columnMappings: map[string]MetricMap{
					"counter": {
						vtype: prometheus.CounterValue,
						desc:  prometheus.NewDesc("test_counter", "Test counter.", nil, nil),
					},
					"gauge": {
						vtype: prometheus.GaugeValue,
						desc:  prometheus.NewDesc("test_gauge", "Test gauge.", nil, nil),
					},
				},
			}

			metrics, nonfatalErrors, err := queryNamespaceMapping(server, "test_namespace", mapping)
			if err != nil {
				t.Fatalf("querying namespace: %v", err)
			}
			if len(nonfatalErrors) != 0 {
				t.Fatalf("unexpected nonfatal errors: %v", nonfatalErrors)
			}
			if len(metrics) != 2 {
				t.Fatalf("got %d metrics, want 2", len(metrics))
			}

			counter := &dto.Metric{}
			if err := metrics[0].Write(counter); err != nil {
				t.Fatalf("writing counter metric: %v", err)
			}
			if got := counter.GetCounter().GetValue(); got != test.expectedCounterValue {
				t.Errorf("counter value = %v, want %v", got, test.expectedCounterValue)
			}

			gauge := &dto.Metric{}
			if err := metrics[1].Write(gauge); err != nil {
				t.Fatalf("writing gauge metric: %v", err)
			}
			if got, want := gauge.GetGauge().GetValue(), float64(largeInteger); got != want {
				t.Errorf("gauge value = %v, want %v", got, want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet database expectations: %v", err)
			}
		})
	}
}
