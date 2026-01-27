// Copyright 2023 The Prometheus Authors
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
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const longRunningTransactionsSubsystem = "long_running_transactions"

func init() {
	registerCollector(longRunningTransactionsSubsystem, defaultDisabled, NewPGLongRunningTransactionsCollector)
}

type PGLongRunningTransactionsCollector struct {
	log *slog.Logger
}

func NewPGLongRunningTransactionsCollector(config collectorConfig) (Collector, error) {
	return &PGLongRunningTransactionsCollector{log: config.logger}, nil
}

var longRunningTransactionThresholds = []int{60, 300, 600, 1800} // 1min, 5min, 10min, 30min

var (
	longRunningTransactionsCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, longRunningTransactionsSubsystem, "count"),
		"Number of transactions running longer than threshold",
		[]string{"threshold"},
		prometheus.Labels{},
	)

	longRunningTransactionsAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, longRunningTransactionsSubsystem, "oldest_timestamp_seconds"),
		"The current maximum transaction age in seconds",
		[]string{},
		prometheus.Labels{},
	)

	longRunningTransactionsQuery = `
	SELECT
      COUNT(*) as transactions,
      MAX(EXTRACT(EPOCH FROM clock_timestamp() - pg_stat_activity.xact_start)) AS oldest_timestamp_seconds
    FROM pg_catalog.pg_stat_activity
    WHERE state IS DISTINCT FROM 'idle'
    AND query NOT LIKE 'autovacuum:%'
    AND pg_stat_activity.xact_start IS NOT NULL
    AND EXTRACT(EPOCH FROM clock_timestamp() - pg_stat_activity.xact_start) >= $1;
    `

	longRunningTransactionsMaxAgeQuery = `
      SELECT
      	MAX(EXTRACT(EPOCH FROM clock_timestamp() - pg_stat_activity.xact_start)) AS oldest_timestamp_seconds
      FROM pg_catalog.pg_stat_activity
      WHERE state IS DISTINCT FROM 'idle'
      AND query NOT LIKE 'autovacuum:%'
      AND pg_stat_activity.xact_start IS NOT NULL;
      `
)

func (PGLongRunningTransactionsCollector) Update(ctx context.Context, instance *Instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	// Query for each threshold
	for _, threshold := range longRunningTransactionThresholds {
		rows, err := db.QueryContext(ctx, longRunningTransactionsQuery, threshold)
		if err != nil {
			return err
		}

		var count float64
		var maxAge sql.NullFloat64

		if rows.Next() {
			if err := rows.Scan(&count, &maxAge); err != nil {
				rows.Close()
				return err
			}
		}
		rows.Close()

		// Emit count metric with threshold label
		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsCount,
			prometheus.GaugeValue,
			count,
			fmt.Sprintf("%d", threshold),
		)
	}

	// Query for max age (no threshold filter)
	rows, err := db.QueryContext(ctx, longRunningTransactionsMaxAgeQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var maxAge sql.NullFloat64
		if err := rows.Scan(&maxAge); err != nil {
			return err
		}

		ageValue := 0.0
		if maxAge.Valid {
			ageValue = maxAge.Float64
		}

		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsAgeInSeconds,
			prometheus.GaugeValue,
			ageValue,
		)
	}

	return rows.Err()
}
