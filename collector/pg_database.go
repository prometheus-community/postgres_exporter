// Copyright 2022 The Prometheus Authors
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
	registerCollector("database", defaultEnabled, NewPGDatabaseCollector)
}

type PGDatabaseCollector struct {
	log log.Logger
}

func NewPGDatabaseCollector(logger log.Logger) (Collector, error) {
	return &PGDatabaseCollector{log: logger}, nil
}

var pgDatabase = map[string]*prometheus.Desc{
	"size_bytes": prometheus.NewDesc(
		"pg_database_size_bytes",
		"Disk space used by the database",
		[]string{"datname"}, nil,
	),

	"xid_ages": prometheus.NewDesc(
		"pg_database_xid_ages",
		"Age by the database",
		[]string{"datname"}, nil,
	),
}

func (PGDatabaseCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		`SELECT pg_database.datname
		,pg_database_size(pg_database.datname)
        ,age(datfrozenxid)
		FROM pg_database;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname string
		var (
			size int64
			age  int64
		)
		if err := rows.Scan(&datname, &size, &age); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgDatabase["size_bytes"],
			prometheus.GaugeValue, float64(size), datname,
		)
		ch <- prometheus.MustNewConstMetric(
			pgDatabase["xid_ages"],
			prometheus.GaugeValue, float64(age), datname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
