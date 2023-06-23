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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const stuckIdleInTransactionSubsystem = "stuck_in_transaction"

func init() {
	registerCollector(stuckIdleInTransactionSubsystem, defaultDisabled, NewPGStuckIdleInTransactionCollector)
}

type PGStuckIdleInTransactionCollector struct {
	log log.Logger
}

func NewPGStuckIdleInTransactionCollector(config collectorConfig) (Collector, error) {
	return &PGStuckIdleInTransactionCollector{log: config.logger}, nil
}

var (
	stuckIdleInTransactionQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, longRunningTransactionsSubsystem, "queries"),
		"Current number of queries that are stuck being idle in transactions",
		[]string{},
		prometheus.Labels{},
	)

	stuckIdleInTransactionQuery = `
		SELECT
			COUNT(*) AS queries
		FROM pg_catalog.pg_stat_activity
		WHERE
			state = 'idle in transaction' AND (now() - query_start) > '10 minutes'::interval
	`
)

func (PGStuckIdleInTransactionCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		stuckIdleInTransactionQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var queries float64

		if err := rows.Scan(&queries); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			stuckIdleInTransactionQueries,
			prometheus.GaugeValue,
			queries,
		)
		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsAgeInSeconds,
			prometheus.GaugeValue,
			queries,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
