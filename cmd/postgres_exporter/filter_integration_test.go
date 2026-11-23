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

//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"testing"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promslog"
)

func scrapedCollectors(t *testing.T, handler http.Handler, target string) (map[string]float64, bool) {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("%s: status code = %d, body: %s", target, response.Code, response.Body.String())
	}

	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(response.Body)
	if err != nil {
		t.Fatalf("%s: parse metrics: %v", target, err)
	}

	got := make(map[string]float64)
	if family, ok := families["pg_scrape_collector_success"]; ok {
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "collector" {
					got[label.GetValue()] = metric.GetGauge().GetValue()
				}
			}
		}
	}
	_, hasLegacy := families["pg_up"]
	return got, hasLegacy
}

func TestCollectorFilterIntegration(t *testing.T) {
	dsn := os.Getenv("DATA_SOURCE_NAME")
	if dsn == "" {
		t.Skip("DATA_SOURCE_NAME is not set")
	}

	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{dsn}

	var enabled []string
	for name, on := range cfg.Collectors {
		if on && name != config.CollectorSettings {
			enabled = append(enabled, name)
		}
	}
	slices.Sort(enabled)

	metrics := newTestMetricsHandler(t, cfg)

	authHandler, err := config.NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	probe := handleProbe(promslog.NewNopLogger(), authHandler, cfg)

	probeTarget := url.QueryEscape(dsn)

	tests := []struct {
		name       string
		handler    http.Handler
		target     string
		want       []string
		wantLegacy bool
	}{
		{
			name:    "metrics collect",
			handler: metrics,
			target:  "/metrics?collect[]=settings&collect[]=database",
			want:    []string{config.CollectorDatabase, config.CollectorSettings},
		},
		{
			name:       "metrics exclude",
			handler:    metrics,
			target:     "/metrics?exclude[]=settings",
			want:       enabled,
			wantLegacy: true,
		},
		{
			name:    "probe collect",
			handler: probe,
			target:  "/probe?target=" + probeTarget + "&collect[]=settings",
			want:    []string{config.CollectorSettings},
		},
		{
			name:       "probe exclude",
			handler:    probe,
			target:     "/probe?target=" + probeTarget + "&exclude[]=settings",
			want:       enabled,
			wantLegacy: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, hasLegacy := scrapedCollectors(t, test.handler, test.target)

			var names []string
			for name := range got {
				names = append(names, name)
			}
			slices.Sort(names)
			if !slices.Equal(names, test.want) {
				t.Fatalf("pg_scrape_collector_success collectors = %v, want %v", names, test.want)
			}
			for _, name := range []string{config.CollectorSettings, config.CollectorDatabase} {
				if value, ok := got[name]; ok && value != 1 {
					t.Errorf("pg_scrape_collector_success{collector=%q} = %v, want 1", name, value)
				}
			}
			if hasLegacy != test.wantLegacy {
				t.Fatalf("pg_up present = %t, want %t", hasLegacy, test.wantLegacy)
			}
		})
	}
}
