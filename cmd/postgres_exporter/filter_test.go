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

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	prometheuscollectors "github.com/prometheus/client_golang/prometheus/collectors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promslog"
)

const unreachableDSN = "postgresql://postgres@127.0.0.1:1/postgres?sslmode=disable&connect_timeout=1"

func newTestMetricsHandler(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	runtime, err := collector.NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	t.Cleanup(func() { runtime.Close() })

	exporterMetrics := prometheus.NewRegistry()
	exporterMetrics.MustRegister(prometheuscollectors.NewGoCollector())
	if _, err := config.NewHandler(exporterMetrics); err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handleMetrics(exporterMetrics, runtime)
}

func scrapeFamilies(t *testing.T, handler http.Handler, target string) (int, string, map[string]*dto.MetricFamily) {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	body := response.Body.String()
	if response.Code != http.StatusOK {
		return response.Code, body, nil
	}
	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(body))
	if err != nil {
		t.Fatalf("%s: parse metrics: %v", target, err)
	}
	return response.Code, body, families
}

func TestHandleMetrics(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.Collectors[config.CollectorStatStatements] = true
	cfg.Collectors[config.CollectorLocks] = false
	handler := newTestMetricsHandler(t, cfg)

	tests := []struct {
		name       string
		query      string
		wantStatus int
		want       []string
		wantAbsent []string
		wantBody   string
	}{
		{
			name:       "no filter",
			query:      "",
			wantStatus: http.StatusOK,
			want:       []string{"go_goroutines", "postgres_exporter_config_last_reload_successful", "pg_exporter_scrapes_total"},
		},
		{
			name:       "collect keeps exporter metrics and skips the legacy exporter",
			query:      "collect[]=stat_statements",
			wantStatus: http.StatusOK,
			want:       []string{"go_goroutines", "postgres_exporter_config_last_reload_successful"},
			wantAbsent: []string{"pg_exporter_scrapes_total", "pg_up"},
		},
		{
			name:       "exclude keeps exporter metrics and the legacy exporter",
			query:      "exclude[]=stat_statements",
			wantStatus: http.StatusOK,
			want:       []string{"go_goroutines", "postgres_exporter_config_last_reload_successful", "pg_exporter_scrapes_total"},
		},
		{
			name:       "collect and exclude together",
			query:      "collect[]=database&exclude[]=stat_statements",
			wantStatus: http.StatusBadRequest,
			wantBody:   "combined collect[] and exclude[]",
		},
		{
			name:       "unknown collector",
			query:      "collect[]=no_such_collector",
			wantStatus: http.StatusBadRequest,
			wantBody:   `unknown collector "no_such_collector"`,
		},
		{
			name:       "collector disabled by flags",
			query:      "collect[]=locks",
			wantStatus: http.StatusBadRequest,
			wantBody:   `collector "locks" is disabled`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, body, families := scrapeFamilies(t, handler, "/metrics?"+test.query)

			if status != test.wantStatus {
				t.Fatalf("status code = %d, want %d, body: %s", status, test.wantStatus, body)
			}
			if test.wantBody != "" && !strings.Contains(body, test.wantBody) {
				t.Fatalf("body does not contain %q: %s", test.wantBody, body)
			}
			for _, name := range test.want {
				if _, ok := families[name]; !ok {
					t.Errorf("metric %s missing", name)
				}
			}
			for _, name := range test.wantAbsent {
				if _, ok := families[name]; ok {
					t.Errorf("metric %s present, want absent", name)
				}
			}
		})
	}
}

func TestHandleMetricsExcludeReusesLegacyExporter(t *testing.T) {
	handler := newTestMetricsHandler(t, config.NewConfigWithDefaults())

	for _, want := range []float64{1, 2, 3} {
		target := "/metrics?exclude[]=settings"
		if want == 2 {
			target = "/metrics"
		}
		status, body, families := scrapeFamilies(t, handler, target)
		if status != http.StatusOK {
			t.Fatalf("status code = %d, body: %s", status, body)
		}
		family, ok := families["pg_exporter_scrapes_total"]
		if !ok {
			t.Fatal("pg_exporter_scrapes_total missing")
		}
		if got := family.GetMetric()[0].GetCounter().GetValue(); got != want {
			t.Fatalf("%s: pg_exporter_scrapes_total = %v, want %v", target, got, want)
		}
	}
}

func TestHandleProbeRejectsInvalidCollectorFilter(t *testing.T) {
	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	handler := handleProbe(promslog.NewNopLogger(), authHandler, config.NewConfigWithDefaults())

	for _, query := range []string{
		"collect[]=database&exclude[]=locks",
		"collect[]=no_such_collector",
		"collect[]=stat_statements",
	} {
		t.Run(query, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/probe?target=127.0.0.1:1&"+query, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if got, want := response.Code, http.StatusBadRequest; got != want {
				t.Fatalf("status code = %d, want %d, body: %s", got, want, response.Body.String())
			}
		})
	}
}

type barrierCollector struct {
	started chan struct{}
	other   chan struct{}
	desc    *prometheus.Desc
}

func (c barrierCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c barrierCollector) Collect(ch chan<- prometheus.Metric) {
	close(c.started)
	value := 1.0
	select {
	case <-c.other:
	case <-time.After(5 * time.Second):
		value = 0
	}
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value)
}

func TestScrapeHandlerCollectsConcurrently(t *testing.T) {
	first, second := make(chan struct{}), make(chan struct{})
	handler := newScrapeHandler(
		barrierCollector{started: first, other: second, desc: prometheus.NewDesc("test_first_overlapped", "test", nil, nil)},
		barrierCollector{started: second, other: first, desc: prometheus.NewDesc("test_second_overlapped", "test", nil, nil)},
	)

	status, body, families := scrapeFamilies(t, handler, "/metrics")
	if status != http.StatusOK {
		t.Fatalf("status code = %d, body: %s", status, body)
	}
	for _, name := range []string{"test_first_overlapped", "test_second_overlapped"} {
		family, ok := families[name]
		if !ok {
			t.Fatalf("metric %s missing", name)
		}
		if got := family.GetMetric()[0].GetGauge().GetValue(); got != 1 {
			t.Fatalf("%s = %v, want 1: collectors did not run concurrently", name, got)
		}
	}
}

func TestFilteredScrapesShareRuntimeConcurrently(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{unreachableDSN}
	handler := newTestMetricsHandler(t, cfg)

	var wg sync.WaitGroup
	for _, target := range []string{
		"/metrics",
		"/metrics?exclude[]=settings",
		"/metrics?collect[]=settings",
		"/metrics?collect[]=database",
		"/metrics?collect[]=settings&collect[]=database",
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusOK {
				t.Errorf("%s: status code = %d, want %d", target, response.Code, http.StatusOK)
			}
		}()
	}
	wg.Wait()
}
