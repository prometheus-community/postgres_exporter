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
	registerCollector(statUserIndexesSubsystem, defaultEnabled, NewPGStatUserIndexesCollector)
}

type PGStatUserIndexesCollector struct {
	log log.Logger
}

const statUserIndexesSubsystem = "stat_user_indexes"

func NewPGStatUserIndexesCollector(config collectorConfig) (Collector, error) {
	return &PGStatUserIndexesCollector{log: config.logger}, nil
}

var (
	statUserIndexesIdxScan = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_scan"),
		"Number of index scans initiated on this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
	statUserIndexesIdxTupRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_tup_read"),
		"Number of index entries returned by scans on this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
	statUserIndexesIdxTupFetch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_tup_fetch"),
		"Number of live table rows fetched by simple index scans using this index",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	statUserIndexesQuery = `
	SELECT
		schemaname,
		relname,
		indexrelname,
		idx_scan,
		idx_tup_read,
		idx_tup_fetch
	FROM pg_stat_user_indexes
	`
)

func (c *PGStatUserIndexesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statUserIndexesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaname, relname, indexrelname string
		var idxScan, idxTupRead, idxTupFetch float64

		if err := rows.Scan(&schemaname, &relname, &indexrelname, &idxScan, &idxTupRead, &idxTupFetch); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxScan,
			prometheus.CounterValue,
			idxScan,
			schemaname, relname, indexrelname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxTupRead,
			prometheus.CounterValue,
			idxTupRead,
			schemaname, relname, indexrelname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxTupFetch,
			prometheus.CounterValue,
			idxTupFetch,
			schemaname, relname, indexrelname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
