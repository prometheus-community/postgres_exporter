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

const indexSizeSubsystem = "index_size"

func init() {
	registerCollector(indexSizeSubsystem, defaultDisabled, NewPGIndexSizeCollector)
}

type PGIndexSizeCollector struct {
	log log.Logger
}

func NewPGIndexSizeCollector(config collectorConfig) (Collector, error) {
	return &PGIndexSizeCollector{log: config.logger}, nil
}

var (
	indexSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, indexSizeSubsystem, "bytes"),
		"Size of the index as per pg_table_size function",
		[]string{"schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	indexSizeQuery = `
	SELECT
		schemaname,
		tablename as relname,
		indexname as indexrelname,
		pg_class.relpages * 8192::bigint as index_size
	FROM
		pg_indexes inner join pg_namespace on pg_indexes.schemaname = pg_namespace.nspname
		inner join pg_class on pg_class.relnamespace = pg_namespace.oid and pg_class.relname = pg_indexes.indexname
	WHERE
		pg_indexes.schemaname != 'pg_catalog'
	`
)

func (PGIndexSizeCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		indexSizeQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaname, relname, indexrelname sql.NullString
		var indexSize sql.NullFloat64

		if err := rows.Scan(&schemaname, &relname, &indexrelname, &indexSize); err != nil {
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

		indexSizeMetric := 0.0
		if indexSize.Valid {
			indexSizeMetric = indexSize.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			indexSizeDesc,
			prometheus.GaugeValue,
			indexSizeMetric,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
