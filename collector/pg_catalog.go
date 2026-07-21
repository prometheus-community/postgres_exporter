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

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const catalogSubsystem = "catalog"

func init() {
	registerCollector(catalogSubsystem, defaultEnabled, NewPGCatalogCollector)
}

type PGCatalogCollector struct {
	log *slog.Logger
}

func NewPGCatalogCollector(config collectorConfig) (Collector, error) {
	exclude := config.excludeDatabases
	if exclude == nil {
		exclude = []string{}
	}
	return &PGCatalogCollector{
		log: config.logger,
	}, nil
}

var (
	pgCatalogRestartPendingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			catalogSubsystem,
			"pending_restart_info",
		),
		"PostgreSQL configuration parameters that are pending a server restart to take effect.",
		[]string{"setting", "pending_value"}, nil,
	)

	pgCatalogQuery = "SELECT name, setting from pg_catalog.pg_settings WHERE pending_restart;"
)

// Update implements Collector and exposes whether changes to the
// configuration that require a restart were made. It is called by
// the Prometheus registry when collecting metrics. The list of
// settings is retrieved from pg_catalog and filtered for settings
// a restart to be applied.
func (c PGCatalogCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx,
		pgCatalogQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var setting string
		var value string
		if err := rows.Scan(&setting, &value); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgCatalogRestartPendingDesc,
			prometheus.GaugeValue, 1,
			setting, value,
		)
	}

	return rows.Err()
}
