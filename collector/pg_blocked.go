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

const blockedSubsystem = "blocked"

func init() {
	registerCollector(blockedSubsystem, defaultDisabled, NewPGBlockedCollector)
}

type PGBlockedCollector struct {
	log log.Logger
}

func NewPGBlockedCollector(config collectorConfig) (Collector, error) {
	return &PGBlockedCollector{log: config.logger}, nil
}

var (
	blockedQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, blockedSubsystem, "queries"),
		"The current number of blocked queries",
		[]string{"table"},
		prometheus.Labels{},
	)

	blockedQuery = `
	SELECT
		count(blocked.transactionid) AS queries,
		'__transaction__' AS table
	FROM pg_catalog.pg_locks blocked
	WHERE NOT blocked.granted AND locktype = 'transactionid'
	GROUP BY locktype
	UNION
	SELECT
		count(blocked.relation) AS queries,
		blocked.relation::regclass::text AS table
	FROM pg_catalog.pg_locks blocked
	WHERE NOT blocked.granted AND locktype != 'transactionid'
	GROUP BY relation
	`
)

func (PGBlockedCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		blockedQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var table sql.NullString
		var queries sql.NullFloat64

		if err := rows.Scan(&queries, &table); err != nil {
			return err
		}

		tableLabel := "unknown"
		if table.Valid {
			tableLabel = table.String
		}

		queriesMetric := 0.0
		if queries.Valid {
			queriesMetric = queries.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			blockedQueries,
			prometheus.GaugeValue,
			queriesMetric,
			tableLabel,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
