// Copyright 2022 The Prometheus Authors
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
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleProbe(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := r.URL.Query()
		target := params.Get("target")
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}

		// TODO: Timeout
		// TODO: Auth Module

		probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "probe_success",
			Help: "Displays whether or not the probe was a success",
		})
		probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "probe_duration_seconds",
			Help: "Returns how long the probe took to complete in seconds",
		})

		tl := log.With(logger, "target", target)
		_ = tl

		start := time.Now()
		registry := prometheus.NewRegistry()
		registry.MustRegister(probeSuccessGauge)
		registry.MustRegister(probeDurationGauge)

		// TODO(@sysadmind): this is a temp hack until we have a proper auth module
		target = "postgres://postgres:test@localhost:5432/circle_test?sslmode=disable"

		// Run the probe
		pc, err := collector.NewProbeCollector(tl, registry, target)
		if err != nil {
			probeSuccessGauge.Set(0)
			probeDurationGauge.Set(time.Since(start).Seconds())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = ctx

		// TODO: Which way should this be? Register or handle the collection manually?
		// Also, what about the context?

		// Option 1: Register the collector
		registry.MustRegister(pc)

		// Option 2: Handle the collection manually. This allows us to collect duration metrics.
		// The collectors themselves already support their own duration metrics.
		// err = pc.Update(ctx)
		// if err != nil {
		// 	probeSuccessGauge.Set(0)
		// } else {
		// 	probeSuccessGauge.Set(1)
		// }

		duration := time.Since(start).Seconds()
		probeDurationGauge.Set(duration)

		// TODO check success, etc
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
