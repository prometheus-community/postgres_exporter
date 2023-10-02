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

const statActivityAutovacuumSubsystem = "stat_activity_autovacuum"

func init() {
	registerCollector(statActivityAutovacuumSubsystem, defaultDisabled, NewPGStatActivityAutovacuumCollector)
}

type PGStatActivityAutovacuumCollector struct {
	log log.Logger
}

func NewPGStatActivityAutovacuumCollector(config collectorConfig) (Collector, error) {
	return &PGStatActivityAutovacuumCollector{log: config.logger}, nil
}

var (
	statActivityAutovacuumAgeInSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivityAutovacuumSubsystem, "timestamp_seconds"),
		"Start timestamp of the vacuum process in seconds",
		[]string{"relname"},
		prometheus.Labels{},
	)

	statActivityAutovacuumQuery = `
    SELECT
		SPLIT_PART(query, '.', 2) AS relname,
		EXTRACT(EPOCH FROM xact_start) AS timestamp_seconds
    FROM
    	pg_catalog.pg_stat_activity
    WHERE
		query LIKE 'autovacuum:%'
	`
)

func (PGStatActivityAutovacuumCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statActivityAutovacuumQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var relname string
		var ageInSeconds float64

		if err := rows.Scan(&relname, &ageInSeconds); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statActivityAutovacuumAgeInSeconds,
			prometheus.GaugeValue,
			ageInSeconds, relname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
