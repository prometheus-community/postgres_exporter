// Copyright 2021 The Prometheus Authors
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

	"github.com/prometheus/client_golang/prometheus"
)

type PGDatabaseCollector struct{}

func NewPGDatabaseCollector() *PGDatabaseCollector {
	return &PGDatabaseCollector{}
}

var pgDatabase = map[string]*prometheus.Desc{
	"size_bytes": prometheus.NewDesc(
		"pg_database_size_bytes",
		"Disk space used by the database",
		[]string{"datname"}, nil,
	),
}

func (PGDatabaseCollector) Update(ctx context.Context, db *sql.DB, server string) ([]prometheus.Metric, error) {
	metrics := []prometheus.Metric{}
	rows, err := db.QueryContext(ctx,
		`SELECT pg_database.datname
		,pg_database_size(pg_database.datname)
		FROM pg_database;`)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		var datname string
		var size int64
		if err := rows.Scan(&datname, &size); err != nil {
			return metrics, err
		}
		metrics = append(metrics, prometheus.MustNewConstMetric(
			pgDatabase["size_bytes"],
			prometheus.GaugeValue, float64(size), datname,
		))
	}
	if err := rows.Err(); err != nil {
		return metrics, err
	}
	return metrics, nil
}
