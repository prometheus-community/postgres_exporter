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
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus-community/postgres_exporter/postgresmetrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleProbe(logger *slog.Logger, excludeDatabases []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := c.GetAuthConfig()
		params := r.URL.Query()
		target := params.Get("target")
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		var authModule config.AuthModule
		authModuleName := params.Get("auth_module")
		if authModuleName == "" {
			logger.Info("no auth_module specified, using default")
		} else {
			var ok bool
			authModule, ok = conf.AuthModules[authModuleName]
			if !ok {
				http.Error(w, fmt.Sprintf("auth_module %s not found", authModuleName), http.StatusBadRequest)
				return
			}
			if authModule.UserPass.Username == "" || authModule.UserPass.Password == "" {
				http.Error(w, fmt.Sprintf("auth_module %s has no username or password", authModuleName), http.StatusBadRequest)
				return
			}
		}

		dsn, err := authModule.ConfigureTarget(target)
		if err != nil {
			logger.Error("failed to configure target", "err", err)
			http.Error(w, fmt.Sprintf("could not configure dsn for target: %v", err), http.StatusBadRequest)
			return
		}

		// TODO(@sysadmind): Timeout

		tl := logger.With("target", target)

		registry := prometheus.NewRegistry()

		opts := []exporter.ExporterOpt{
			exporter.DisableDefaultMetrics(cfg.DisableDefaultMetrics),
			exporter.DisableSettingsMetrics(cfg.DisableSettingsMetrics),
			exporter.AutoDiscoverDatabases(legacyMetricsFlags.AutoDiscoverDatabases),
			exporter.WithUserQueriesPath(legacyMetricsFlags.UserQueriesPath),
			exporter.WithConstantLabels(legacyMetricsFlags.ConstantLabels),
			exporter.ExcludeDatabases(excludeDatabases),
			exporter.IncludeDatabases(legacyMetricsFlags.IncludeDatabases),
			exporter.WithMetricPrefix(cfg.MetricPrefix),
		}

		dsns := []string{dsn.GetConnectionString()}
		exporter := exporter.NewExporter(dsns, logger, opts...)
		defer func() {
			exporter.CloseServers()
		}()
		registry.MustRegister(exporter)

		// Run the probe
		registration, err := postgresmetrics.NewProbe(cfg, tl, dsn.GetConnectionString())
		if err != nil {
			logger.Error("Error creating probe collectors", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer registration.Close()
		if err := registration.Register(registry); err != nil {
			logger.Error("Error registering probe collectors", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO check success, etc
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
