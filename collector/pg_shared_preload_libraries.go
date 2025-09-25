// Copyright 2025 The Prometheus Authors
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
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const sharedPreloadLibrariesSubsystem = "settings"

func init() {
	registerCollector(sharedPreloadLibrariesSubsystem, defaultEnabled, NewPGSharedPreloadLibrariesCollector)
}

type PGSharedPreloadLibrariesCollector struct{}

func NewPGSharedPreloadLibrariesCollector(collectorConfig) (Collector, error) {
	return &PGSharedPreloadLibrariesCollector{}, nil
}

var (
	pgSharedPreloadLibrariesLibraryEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			sharedPreloadLibrariesSubsystem,
			"shared_preload_library_enabled",
		),
		"Whether a library is listed in shared_preload_libraries (1=yes).",
		[]string{"library"}, nil,
	)

	pgSharedPreloadLibrariesQuery = "SELECT setting FROM pg_settings WHERE name = 'shared_preload_libraries'"
)

func (c *PGSharedPreloadLibrariesCollector) Update(ctx context.Context, instance *Instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx, pgSharedPreloadLibrariesQuery)

	var setting sql.NullString
	err := row.Scan(&setting)
	if err != nil {
		return err
	}

	// Parse, trim, dedupe and sort libraries for stable series emission.
	libsSet := map[string]struct{}{}
	if setting.Valid && setting.String != "" {
		for _, raw := range strings.Split(setting.String, ",") {
			lib := strings.TrimSpace(raw)
			if lib == "" {
				continue
			}
			libsSet[lib] = struct{}{}
		}
	}
	libs := make([]string, 0, len(libsSet))
	for lib := range libsSet {
		libs = append(libs, lib)
	}
	sort.Strings(libs)

	for _, lib := range libs {
		ch <- prometheus.MustNewConstMetric(
			pgSharedPreloadLibrariesLibraryEnabled,
			prometheus.GaugeValue,
			1,
			lib,
		)
	}
	return nil
}
