// Copyright 2024 The Prometheus Authors
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

const tableSizeSubsystem = "table_size"

func init() {
	registerCollector(tableSizeSubsystem, defaultEnabled, NewPGTableSizeCollector)
}

type PGTableSizeCollector struct {
	log log.Logger
}

func NewPGTableSizeCollector(config collectorConfig) (Collector, error) {
	return &PGTableSizeCollector{
		log: config.logger,
	}, nil
}

var (
	pgTableTotalRelationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			tableSizeSubsystem,
			"total_relation",
		),
		"Total Relation Size of the table",
		[]string{"schemaname", "relname", "datname"}, nil,
	)
	pgTableIndexSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			tableSizeSubsystem,
			"index",
		),
		"Indexes Size of the Table",
		[]string{"schemaname", "relname", "datname"}, nil,
	)
	pgRelationSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			tableSizeSubsystem,
			"relation",
		),
		"Relation Size of the table",
		[]string{"schemaname", "relname", "datname"}, nil,
	)
	pgTableSizeQuery = `SELECT
        table_catalog relname,
        table_name datname,
        table_schema schemaname,
        pg_total_relation_size('"'||table_schema||'"."'||table_name||'"') total_relation_size,
        pg_relation_size('"'||table_schema||'"."'||table_name||'"') relation_size,
        pg_indexes_size('"'||table_schema||'"."'||table_name||'"')  indexes_size
    FROM information_schema.tables`
)

// Update implements Collector and exposes database locks.
// It is called by the Prometheus registry when collecting metrics.
func (c PGTableSizeCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx, pgTableSizeQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaName, relName, databaseName sql.NullString
		var totalRelationSize, relationSize, indexesSize sql.NullInt64

		if err := rows.Scan(&databaseName, &relName, &schemaName, &totalRelationSize, &relationSize, &indexesSize); err != nil {
			return err
		}

		if !schemaName.Valid || !relName.Valid || !databaseName.Valid {
			continue
		}

		totalRelationsSizeMetric := 0.0
		relationSizeMetric := 0.0
		indexesSizeMetric := 0.0
		if totalRelationSize.Valid {
			totalRelationsSizeMetric = float64(totalRelationSize.Int64)
		}

		if relationSize.Valid {
			relationSizeMetric = float64(relationSize.Int64)
		}

		if indexesSize.Valid {
			indexesSizeMetric = float64(indexesSize.Int64)
		}

		ch <- prometheus.MustNewConstMetric(
			pgTableTotalRelationDesc,
			prometheus.CounterValue, totalRelationsSizeMetric,
			schemaName.String, relName.String, databaseName.String,
		)

		ch <- prometheus.MustNewConstMetric(
			pgRelationSizeDesc,
			prometheus.CounterValue, relationSizeMetric,
			schemaName.String, relName.String, databaseName.String,
		)

		ch <- prometheus.MustNewConstMetric(
			pgTableIndexSizeDesc,
			prometheus.CounterValue, indexesSizeMetric,
			schemaName.String, relName.String, databaseName.String,
		)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
