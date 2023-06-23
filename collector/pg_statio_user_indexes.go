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

func init() {
	registerCollector(statioUserIndexesSubsystem, defaultEnabled, NewPGStatioUserIndexesCollector)
}

type PGStatioUserIndexesCollector struct {
	log log.Logger
}

const statioUserIndexesSubsystem = "statio_user_indexes"

func NewPGStatioUserIndexesCollector(collectorConfig) (Collector, error) {
	return &PGStatioUserIndexesCollector{}, nil
}

var (
	statioUserIndexesIdxBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_read"),
		"Number of disk blocks read from this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
	statioUserIndexesIdxBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_hit"),
		"Number of buffer hits in this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	statioUserIndexesQuery = `
	SELECT
		schemaname,
		relname,
		indexrelname,
		idx_blks_read,
		idx_blks_hit
	FROM pg_statio_user_indexes
	`
)

func (c *PGStatioUserIndexesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statioUserIndexesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaname, relname, indexrelname string
		var idxBlksRead, idxBlksHit float64

		if err := rows.Scan(&schemaname, &relname, &indexrelname, &idxBlksRead, &idxBlksHit); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statioUserIndexesIdxBlksRead,
			prometheus.CounterValue,
			idxBlksRead,
			schemaname, relname, indexrelname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserIndexesIdxBlksHit,
			prometheus.CounterValue,
			idxBlksHit,
			schemaname, relname, indexrelname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
