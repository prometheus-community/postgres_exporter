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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const longRunningTransactionsSummarySubsystem = "long_running_transactions_summary"

func init() {
	registerCollector(longRunningTransactionsSummarySubsystem, defaultDisabled, NewPGLongRunningTransactionsSummaryCollector)
}

type PGLongRunningTransactionsSummaryCollector struct {
	log log.Logger
}

func NewPGLongRunningTransactionsSummaryCollector(config collectorConfig) (Collector, error) {
	return &PGLongRunningTransactionsSummaryCollector{log: config.logger}, nil
}

var (
	longRunningTransactionsSummaryMaxAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, longRunningTransactionsSummarySubsystem, "max_age_in_seconds"),
		"The current maximum transaction age in seconds",
		[]string{"application", "endpoint"},
		prometheus.Labels{},
	)

	longRunningTransactionsSummaryQuery = `
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

func (PGLongRunningTransactionsSummaryCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		longRunningTransactionsSummaryQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var application, endpoint sql.NullString
		var maxAgeInSeconds sql.NullFloat64

		if err := rows.Scan(&application, &endpoint, &maxAgeInSeconds); err != nil {
			return err
		}

		applicationLabel := "unknown"
		if application.Valid {
			applicationLabel = application.String
		}
		endpointLabel := "unknown"
		if endpoint.Valid {
			endpointLabel = endpoint.String
		}
		labels := []string{applicationLabel, endpointLabel}

		maxAgeInSecondsMetric := 0.0
		if maxAgeInSeconds.Valid {
			maxAgeInSecondsMetric = maxAgeInSeconds.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			longRunningTransactionsSummaryMaxAgeInSeconds,
			prometheus.GaugeValue,
			maxAgeInSecondsMetric,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
