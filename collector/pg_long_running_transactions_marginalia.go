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

const longRunningTransactionsMarginaliaSubsystem = "long_running_transactions_marginalia"

func init() {
	registerCollector(longRunningTransactionsMarginaliaSubsystem, defaultEnabled, NewPGLongRunningTransactionsMarginaliaCollector)
}

type PGLongRunningTransactionsMarginaliaCollector struct {
	log log.Logger
}

func NewPGLongRunningTransactionsMarginaliaCollector(config collectorConfig) (Collector, error) {
	return &PGLongRunningTransactionsMarginaliaCollector{log: config.logger}, nil
}

var (
	longRunningTransactionsMarginaliaMaxAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, longRunningTransactionsMarginaliaSubsystem, "max_age_in_seconds"),
		"The current maximum transaction age in seconds",
		[]string{"application", "endpoint"},
		prometheus.Labels{},
	)

	longRunningTransactionsMarginaliaQuery = `
	SELECT
		activity.matches[1] AS application,
		activity.matches[2] AS endpoint,
		MAX(age_in_seconds) AS max_age_in_seconds
	FROM (
		SELECT
		regexp_matches(query, '^\s*(?:\/\*(?:application:(\w+),?)?(?:correlation_id:\w+,?)?(?:jid:\w+,?)?(?:endpoint_id:([\w/\-\.:\#\s]+),?)?.*?\*\/)?\s*(\w+)') AS matches,
		EXTRACT(EPOCH FROM (clock_timestamp() - xact_start)) AS age_in_seconds
		FROM
		pg_catalog.pg_stat_activity
		WHERE state <> 'idle'
		AND (clock_timestamp() - xact_start) > '30 seconds'::interval
		AND query NOT LIKE 'autovacuum:%'
		) activity
	GROUP BY application, endpoint
	ORDER BY max_age_in_seconds DESC
	`
)

func (PGLongRunningTransactionsMarginaliaCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		longRunningTransactionsMarginaliaQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var application, endpoint string
		var max_age_in_seconds float64

		if err := rows.Scan(&application, &endpoint, &max_age_in_seconds); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsAgeInSeconds,
			prometheus.GaugeValue,
			max_age_in_seconds,
			application, endpoint,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
