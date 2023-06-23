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

const oldestBlockedSubsystem = "oldest_blocked"

func init() {
	registerCollector(statioUserTableSubsystem, defaultEnabled, NewPGStatIOUserTablesCollector)
}

type PGOldestBlockedCollector struct {
	log log.Logger
}

func NewPGOldestBlockedCollector(config collectorConfig) (Collector, error) {
	return &PGOldestBlockedCollector{log: config.logger}, nil
}

var (
	oldestBlockedAgeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, oldestBlockedSubsystem, "age_seconds"),
		"Largest number of seconds any transaction is currently waiting on a lock",
		[]string{},
		prometheus.Labels{},
	)

	oldestBlockedQuery = `
	SELECT
		coalesce(extract('epoch' from max(clock_timestamp() - state_change)), 0) age_seconds
	FROM
		pg_catalog.pg_stat_activity
	WHERE
		wait_event_type = 'Lock'
	 	AND state='active'
	`
)

func (PGOldestBlockedCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		oldestBlockedQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ageSeconds float64

		if err := rows.Scan(&ageSeconds); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			oldestBlockedAgeSeconds,
			prometheus.GaugeValue,
			ageSeconds,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
