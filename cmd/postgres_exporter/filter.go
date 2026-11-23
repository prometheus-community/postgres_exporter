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

	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newScrapeHandler(collectors ...prometheus.Collector) http.Handler {
	registry := prometheus.NewRegistry()
	for _, c := range collectors {
		registry.MustRegister(c)
	}
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func handleMetrics(exporterMetrics prometheus.Collector, runtime *collector.Runtime) http.Handler {
	unfiltered := newScrapeHandler(append([]prometheus.Collector{exporterMetrics}, runtime.Collectors()...)...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		include, exclude := params["collect[]"], params["exclude[]"]
		if len(include) == 0 && len(exclude) == 0 {
			unfiltered.ServeHTTP(w, r)
			return
		}
		collectors, err := runtime.FilteredCollectors(include, exclude)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newScrapeHandler(append([]prometheus.Collector{exporterMetrics}, collectors...)...).ServeHTTP(w, r)
	})
}
