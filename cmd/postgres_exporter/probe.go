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

	"github.com/prometheus-community/postgres_exporter/collectors"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleProbe(logger *slog.Logger, baseConfig config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		// Copy process-level config before setting the per-request target DSN.
		probeConfig := baseConfig
		probeConfig.DataSourceNames = []string{dsn.GetConnectionString()}
		if err := probeConfig.Validate(); err != nil {
			logger.Error("invalid probe config", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		runtime, err := collectors.NewRuntime(&probeConfig, tl)
		if err != nil {
			logger.Error("error creating probe runtime", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := runtime.Close(); err != nil {
				logger.Error("error closing probe runtime", "err", err)
			}
		}()

		registry := prometheus.NewRegistry()
		for _, collector := range runtime.Collectors() {
			registry.MustRegister(collector)
		}

		_ = ctx

		// TODO check success, etc
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
