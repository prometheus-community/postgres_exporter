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

const statActivityMarginaliaSubsystem = "stat_activity_marginalia"

func init() {
	registerCollector(statActivityMarginaliaSubsystem, defaultEnabled, NewPGStatActivityMarginaliaCollector)
}

type PGStatActivityMarginaliaCollector struct {
	log log.Logger
}

func NewPGStatActivityMarginaliaCollector(config collectorConfig) (Collector, error) {
	return &PGStatActivityMarginaliaCollector{log: config.logger}, nil
}

var (
	statActivityMarginaliaActiveCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivityMarginaliaSubsystem, "active_count"),
		"Number of active queries at time of sample",
		[]string{"usename", "application", "endpoint", "command", "state", "wait_event", "wait_event_type"},
		prometheus.Labels{},
	)
	statActivityMarginaliaMaxTxAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivityMarginaliaSubsystem, "max_tx_age_in_seconds"),
		"Number of active queries at time of sample",
		[]string{"usename", "application", "endpoint", "command", "state", "wait_event", "wait_event_type"},
		prometheus.Labels{},
	)

	statActivityMarginaliaQuery = `
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

func (PGStatActivityMarginaliaCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statActivityMarginaliaQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var usename, application, endpoint, command, state, waitEvent, waitEventType string
		var count, maxTxAge float64

		if err := rows.Scan(&usename, &application, &endpoint, &command, &state, &waitEvent, &waitEventType, &count, &maxTxAge); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statActivityMarginaliaActiveCount,
			prometheus.GaugeValue,
			count,
			usename, application, endpoint, command, state, waitEvent, waitEventType,
		)
		ch <- prometheus.MustNewConstMetric(
			statActivityMarginaliaMaxTxAgeInSeconds,
			prometheus.GaugeValue,
			maxTxAge,
			usename, application, endpoint, command, state, waitEvent, waitEventType,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
