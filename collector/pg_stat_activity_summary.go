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

const statActivitySummarySubsystem = "stat_activity_summary"

func init() {
	registerCollector(statActivitySummarySubsystem, defaultEnabled, NewPGStatActivitySummaryCollector)
}

type PGStatActivitySummaryCollector struct {
	log log.Logger
}

func NewPGStatActivitySummaryCollector(config collectorConfig) (Collector, error) {
	return &PGStatActivitySummaryCollector{log: config.logger}, nil
}

var (
	statActivitySummaryActiveCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivitySummarySubsystem, "active_count"),
		"Number of active queries at time of sample",
		[]string{"usename", "application", "endpoint", "command", "state", "wait_event", "wait_event_type"},
		prometheus.Labels{},
	)
	statActivitySummaryMaxTxAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivitySummarySubsystem, "max_tx_age_in_seconds"),
		"Number of active queries at time of sample",
		[]string{"usename", "application", "endpoint", "command", "state", "wait_event", "wait_event_type"},
		prometheus.Labels{},
	)

	statActivitySummaryQuery = `
	SELECT
		usename AS usename,
		a.matches[1] AS application,
		a.matches[2] AS endpoint,
		a.matches[3] AS command,
		a.state AS state,
		a.wait_event AS wait_event,
		a.wait_event_type AS wait_event_type,
		COUNT(*) active_count,
		MAX(age_in_seconds) AS max_tx_age_in_seconds
	FROM (
		SELECT
			usename,
			regexp_matches(query, '^\s*(?:\/\*(?:application:(\w+),?)?(?:correlation_id:\w+,?)?(?:jid:\w+,?)?(?:endpoint_id:([\w/\-\.:\#\s]+),?)?.*?\*\/)?\s*(\w+)') AS matches,
			state,
			wait_event,
			wait_event_type,
			EXTRACT(EPOCH FROM (clock_timestamp() - xact_start)) AS age_in_seconds
			FROM
			pg_catalog.pg_stat_activity
		) a
	GROUP BY usename, application, endpoint, command, state, wait_event, wait_event_type
	ORDER BY active_count DESC
	`
)

func (PGStatActivitySummaryCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statActivitySummaryQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var usename, application, endpoint, command, state, waitEvent, waitEventType sql.NullString
		var count, maxTxAge sql.NullFloat64

		if err := rows.Scan(&usename, &application, &endpoint, &command, &state, &waitEvent, &waitEventType, &count, &maxTxAge); err != nil {
			return err
		}
		usenameLabel := "unknown"
		if usename.Valid {
			usenameLabel = usename.String
		}
		applicationLabel := "unknown"
		if application.Valid {
			applicationLabel = application.String
		}
		endpointLabel := "unknown"
		if endpoint.Valid {
			endpointLabel = endpoint.String
		}
		commandLabel := "unknown"
		if command.Valid {
			commandLabel = command.String
		}
		stateLabel := "unknown"
		if state.Valid {
			stateLabel = state.String
		}
		waitEventLabel := "unknown"
		if waitEvent.Valid {
			waitEventLabel = waitEvent.String
		}
		waitEventTypeLabel := "unknown"
		if waitEventType.Valid {
			waitEventTypeLabel = waitEventType.String
		}
		labels := []string{usenameLabel, applicationLabel, endpointLabel, commandLabel, stateLabel, waitEventLabel, waitEventTypeLabel}

		countMetric := 0.0
		if count.Valid {
			countMetric = count.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statActivitySummaryActiveCount,
			prometheus.GaugeValue,
			countMetric,
			labels...,
		)

		maxTxAgeMetric := 0.0
		if maxTxAge.Valid {
			maxTxAgeMetric = maxTxAge.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statActivitySummaryMaxTxAgeInSeconds,
			prometheus.GaugeValue,
			maxTxAgeMetric,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
