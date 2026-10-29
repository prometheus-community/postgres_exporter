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
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus-community/postgres_exporter/config"
)

func handleReload(authHandler *config.Handler, configFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := authHandler.ReloadAuthConfig(configFile, logger); err != nil {
			http.Error(w, fmt.Sprintf("failed to reload configuration: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleReloadSignal(signals <-chan os.Signal, authHandler *config.Handler, configFile string, logger *slog.Logger) {
	for range signals {
		if err := authHandler.ReloadAuthConfig(configFile, logger); err != nil {
			logger.Error("Error reloading config", "err", err)
		}
	}
}
