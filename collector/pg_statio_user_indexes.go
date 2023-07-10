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

func init() {
	registerCollector(statioUserIndexesSubsystem, defaultDisabled, NewPGStatioUserIndexesCollector)
}

type PGStatioUserIndexesCollector struct {
	log log.Logger
}

const statioUserIndexesSubsystem = "statio_user_indexes"

func NewPGStatioUserIndexesCollector(config collectorConfig) (Collector, error) {
	return &PGStatioUserIndexesCollector{log: config.logger}, nil
}

var (
	statioUserIndexesIdxBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_read_total"),
		"Number of disk blocks read from this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
	statioUserIndexesIdxBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_hit_total"),
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
		var schemaname, relname, indexrelname sql.NullString
		var idxBlksRead, idxBlksHit sql.NullFloat64

		if err := rows.Scan(&schemaname, &relname, &indexrelname, &idxBlksRead, &idxBlksHit); err != nil {
			return err
		}
		schemanameLabel := "unknown"
		if schemaname.Valid {
			schemanameLabel = schemaname.String
		}
		relnameLabel := "unknown"
		if relname.Valid {
			relnameLabel = relname.String
		}
		indexrelnameLabel := "unknown"
		if indexrelname.Valid {
			indexrelnameLabel = indexrelname.String
		}
		labels := []string{schemanameLabel, relnameLabel, indexrelnameLabel}

		idxBlksReadMetric := 0.0
		if idxBlksRead.Valid {
			idxBlksReadMetric = idxBlksRead.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statioUserIndexesIdxBlksRead,
			prometheus.CounterValue,
			idxBlksReadMetric,
			labels...,
		)

		idxBlksHitMetric := 0.0
		if idxBlksHit.Valid {
			idxBlksHitMetric = idxBlksHit.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statioUserIndexesIdxBlksHit,
			prometheus.CounterValue,
			idxBlksHitMetric,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
