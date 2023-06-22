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

const statioUserTableSubsystem = "statio_user_tables"

func init() {
	registerCollector(statioUserTableSubsystem, defaultEnabled, NewPGStatIOUserTablesCollector)
}

type PGStatIOUserTablesCollector struct {
	log log.Logger
}

func NewPGStatIOUserTablesCollector(config collectorConfig) (Collector, error) {
	return &PGStatIOUserTablesCollector{log: config.logger}, nil
}

var (
	statioUserTablesHeapBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "heap_blocks_read"),
		"Number of disk blocks read from this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesHeapBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "heap_blocks_hit"),
		"Number of buffer hits in this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesIdxBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "idx_blocks_read"),
		"Number of disk blocks read from all indexes on this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesIdxBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "idx_blocks_hit"),
		"Number of buffer hits in all indexes on this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesToastBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "toast_blocks_read"),
		"Number of disk blocks read from this table's TOAST table (if any)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesToastBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "toast_blocks_hit"),
		"Number of buffer hits in this table's TOAST table (if any)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesTidxBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "tidx_blocks_read"),
		"Number of disk blocks read from this table's TOAST table indexes (if any)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statioUserTablesTidxBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statioUserTableSubsystem, "tidx_blocks_hit"),
		"Number of buffer hits in this table's TOAST table indexes (if any)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)

	statioUserTablesQuery = `SELECT
		current_database() datname,
		schemaname,
		relname,
		heap_blks_read,
		heap_blks_hit,
		idx_blks_read,
		idx_blks_hit,
		toast_blks_read,
		toast_blks_hit,
		tidx_blks_read,
		tidx_blks_hit
	FROM pg_statio_user_tables`
)

func (PGStatIOUserTablesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statioUserTablesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname string
		var schemaname string
		var relname string
		var heapBlksRead int64
		var heapBlksHit int64
		var idxBlksRead int64
		var idxBlksHit int64
		var toastBlksRead int64
		var toastBlksHit int64
		var tidxBlksRead int64
		var tidxBlksHit int64

		if err := rows.Scan(&datname, &schemaname, &relname, &heapBlksRead, &heapBlksHit, &idxBlksRead, &idxBlksHit, &toastBlksRead, &toastBlksHit, &tidxBlksRead, &tidxBlksHit); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statioUserTablesHeapBlksRead,
			prometheus.CounterValue,
			float64(heapBlksRead),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesHeapBlksHit,
			prometheus.CounterValue,
			float64(heapBlksHit),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesIdxBlksRead,
			prometheus.CounterValue,
			float64(idxBlksRead),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesIdxBlksHit,
			prometheus.CounterValue,
			float64(idxBlksHit),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesToastBlksRead,
			prometheus.CounterValue,
			float64(toastBlksRead),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesToastBlksHit,
			prometheus.CounterValue,
			float64(toastBlksHit),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesTidxBlksRead,
			prometheus.CounterValue,
			float64(tidxBlksRead),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statioUserTablesTidxBlksHit,
			prometheus.CounterValue,
			float64(tidxBlksHit),
			datname, schemaname, relname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
