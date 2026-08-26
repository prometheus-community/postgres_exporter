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
	"log/slog"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(longRunningTransactionsSubsystem, NewPGLongRunningTransactionsCollector)
}

type PGLongRunningTransactionsCollector struct {
	log       *slog.Logger
	threshold time.Duration
}

func defaultLongRunningTransactionsConfig() config.LongRunningTransactionsConfig {
	return config.LongRunningTransactionsConfig{
		Threshold: config.DefaultLongRunningTransactionsThreshold,
	}
}

func NewPGLongRunningTransactionsCollector(cfg collectorConfig) (Collector, error) {
	return &PGLongRunningTransactionsCollector{
		log:       cfg.logger,
		threshold: cfg.longRunningTransactionsConfig.Threshold,
	}, nil
}

var (
	longRunningTransactionsCount = prometheus.NewDesc(
		"pg_long_running_transactions",
		"Current number of long running transactions",
		[]string{},
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
AND (now() - pg_stat_activity.xact_start) > make_interval(secs => $1)
AND query NOT LIKE 'autovacuum:%'
AND pg_stat_activity.xact_start IS NOT NULL
AND pid <> pg_backend_pid();
	`
)

func (c PGLongRunningTransactionsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		longRunningTransactionsQuery,
		c.threshold.Seconds())

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var transactions float64
		var ageInSeconds sql.NullFloat64

		if err := rows.Scan(&transactions, &ageInSeconds); err != nil {
			return err
		}

		// If there are no long running transactions, ageInSeconds will be NULL
		// so we set it to 0
		age := 0.0
		if ageInSeconds.Valid {
			age = ageInSeconds.Float64
		}

		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsCount,
			prometheus.GaugeValue,
			transactions,
		)
		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsAgeInSeconds,
			prometheus.GaugeValue,
			age,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
