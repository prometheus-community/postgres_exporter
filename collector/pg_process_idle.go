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
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Making this default disabled because we have no tests for it
	registerCollector(processIdleSubsystem, defaultDisabled, NewPGProcessIdleCollector)
}

type PGProcessIdleCollector struct {
	log log.Logger
}

const processIdleSubsystem = "process_idle"

func NewPGProcessIdleCollector(config collectorConfig) (Collector, error) {
	return &PGProcessIdleCollector{log: config.logger}, nil
}

var pgProcessIdleSeconds = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, processIdleSubsystem, "seconds"),
	"Idle time of server processes",
	[]string{"state", "application_name"},
	prometheus.Labels{},
)

func (PGProcessIdleCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx,
		`WITH
			metrics AS (
				SELECT
				state,
				application_name,
				SUM(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - state_change))::bigint)::float AS process_idle_seconds_sum,
				COUNT(*) AS process_idle_seconds_count
				FROM pg_stat_activity
				WHERE state ~ '^idle'
				GROUP BY state, application_name
			),
			buckets AS (
				SELECT
				state,
				application_name,
				le,
				SUM(
					CASE WHEN EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - state_change)) <= le
					THEN 1
					ELSE 0
					END
				)::bigint AS bucket
				FROM
				pg_stat_activity,
				UNNEST(ARRAY[1, 2, 5, 15, 30, 60, 90, 120, 300]) AS le
				GROUP BY state, application_name, le
				ORDER BY state, application_name, le
			)
			SELECT
			state,
			application_name,
			process_idle_seconds_sum as seconds_sum,
			process_idle_seconds_count as seconds_count,
			ARRAY_AGG(le) AS seconds,
			ARRAY_AGG(bucket) AS seconds_bucket
			FROM metrics JOIN buckets USING (state, application_name)
			GROUP BY 1, 2, 3, 4;`)

	var state sql.NullString
	var applicationName sql.NullString
	var secondsSum sql.NullFloat64
	var secondsCount sql.NullInt64
	var seconds []float64
	var secondsBucket []int64

	err := row.Scan(&state, &applicationName, &secondsSum, &secondsCount, pq.Array(&seconds), pq.Array(&secondsBucket))
	if err != nil {
		return err
	}

	var buckets = make(map[float64]uint64, len(seconds))
	for i, second := range seconds {
		if i >= len(secondsBucket) {
			break
		}
		buckets[second] = uint64(secondsBucket[i])
	}

	stateLabel := "unknown"
	if state.Valid {
		stateLabel = state.String
	}

	applicationNameLabel := "unknown"
	if applicationName.Valid {
		applicationNameLabel = applicationName.String
	}

	var secondsCountMetric uint64
	if secondsCount.Valid {
		secondsCountMetric = uint64(secondsCount.Int64)
	}
	secondsSumMetric := 0.0
	if secondsSum.Valid {
		secondsSumMetric = secondsSum.Float64
	}
	ch <- prometheus.MustNewConstHistogram(
		pgProcessIdleSeconds,
		secondsCountMetric, secondsSumMetric, buckets,
		stateLabel, applicationNameLabel,
	)
	return nil
}
