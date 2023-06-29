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

const statActivityAutovacuumActiveSubsystem = "stat_activity_autovacuum_active"

func init() {
	registerCollector(statActivityAutovacuumActiveSubsystem, defaultDisabled, NewPGStatActivityAutovacuumActiveCollector)
}

type PGStatActivityAutovacuumActiveCollector struct {
	log log.Logger
}

func NewPGStatActivityAutovacuumActiveCollector(config collectorConfig) (Collector, error) {
	return &PGStatActivityAutovacuumActiveCollector{log: config.logger}, nil
}

var (
	statActivityAutovacuumActiveWorkersCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivityAutovacuumActiveSubsystem, "workers"),
		"Current number of statActivityAutovacuumActive queries",
		[]string{"phase", "mode"},
		prometheus.Labels{},
	)

	statActivityAutovacuumActiveQuery = `
	SELECT
		v.phase,
		CASE
			when a.query ~ '^autovacuum.*to prevent wraparound' then 'wraparound'
			when a.query ~* '^vacuum' then 'user'
			when a.pid is null then 'idle'
		ELSE 'regular'
		END AS mode,
		count(1) AS workers_count
	FROM pg_stat_progress_vacuum v
	LEFT JOIN pg_catalog.pg_stat_activity a using (pid)
	GROUP BY 1,2
	`
)

func (PGStatActivityAutovacuumActiveCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statActivityAutovacuumActiveQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var phase, mode string
		var workersCount float64

		if err := rows.Scan(&phase, &mode, &workersCount); err != nil {
			return err
		}
		labels := []string{phase, mode}

		ch <- prometheus.MustNewConstMetric(
			statActivityAutovacuumActiveWorkersCount,
			prometheus.GaugeValue,
			workersCount,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
