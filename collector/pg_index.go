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
	"fmt"
	"strings"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(pgIndexSubsystem, defaultDisabled, NewPgIndexCollector)
}

type PGIndexCollector struct {
	log *slog.Logger
}

const pgIndexSubsystem = "index"

func NewPgIndexCollector(config collectorConfig) (Collector, error) {
	return &PGIndexCollector{log: config.logger}, nil
}

var (
	pgIndexProperties = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, pgIndexSubsystem, "properties"),
		"Postgresql index properties",
		[]string{"datname", "schemaname", "relname", "indexrelname", "is_unique", "is_primary", "is_valid", "is_ready"},
		prometheus.Labels{},
	)
	pgIndexSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, pgIndexSubsystem, "size_bytes"),
		"Postgresql index size in bytes",
		[]string{"datname", "schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
)

func pgIndexQuery(columns []string) string {
	return fmt.Sprintf("SELECT %s FROM pg_catalog.pg_stat_user_indexes s JOIN pg_catalog.pg_index i ON s.indexrelid = i.indexrelid WHERE i.indislive='1';", strings.Join(columns, ","))
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (c *PGIndexCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	columns := []string{
		"current_database() datname",
		"s.schemaname",
		"s.relname",
		"s.indexrelname",
		"i.indisunique",
		"i.indisprimary",
		"i.indisvalid",
		"i.indisready",
		"pg_relation_size(i.indexrelid) AS indexsize",
	}

	rows, err := db.QueryContext(ctx,
		pgIndexQuery(columns),
	)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var datname, schemaname, relname, indexrelname sql.NullString
		var idxIsUnique, idxIsPrimary, idxIsValid, idxIsReady sql.NullBool
		var idxSize sql.NullFloat64

		r := []any{
			&datname,
			&schemaname,
			&relname,
			&indexrelname,
			&idxIsUnique,
			&idxIsPrimary,
			&idxIsValid,
			&idxIsReady,
			&idxSize,
		}

		if err := rows.Scan(r...); err != nil {
			return err
		}
		datnameLabel := "unknown"
		if datname.Valid {
			datnameLabel = datname.String
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

		indexIsUniqueLabel := "unknown"
		if idxIsUnique.Valid {
			indexIsUniqueLabel = boolToString(idxIsUnique.Bool)
		}

		indexIsPrimaryLabel := "unknown"
		if idxIsPrimary.Valid {
			indexIsPrimaryLabel = boolToString(idxIsPrimary.Bool)
		}

		indexIsValidLabel := "unknown"
		if idxIsValid.Valid {
			indexIsValidLabel = boolToString(idxIsValid.Bool)
		}

		indexIsReadyLabel := "unknown"
		if idxIsReady.Valid {
			indexIsReadyLabel = boolToString(idxIsReady.Bool)
		}

		indexSizeMetric := -1.0
		if idxSize.Valid {
			indexSizeMetric = idxSize.Float64
		}

		propertiesLabels := []string{datnameLabel, schemanameLabel, relnameLabel, indexrelnameLabel, indexIsUniqueLabel, indexIsPrimaryLabel, indexIsValidLabel, indexIsReadyLabel}
		ch <- prometheus.MustNewConstMetric(
			pgIndexProperties,
			prometheus.CounterValue,
			1,
			propertiesLabels...,
		)

		sizeLabels := []string{datnameLabel, schemanameLabel, relnameLabel, indexrelnameLabel}
		ch <- prometheus.MustNewConstMetric(
			pgIndexSize,
			prometheus.GaugeValue,
			indexSizeMetric,
			sizeLabels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
