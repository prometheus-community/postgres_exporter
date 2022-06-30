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
		[]string{"datname", "server"}, nil,
	),
}

func (PGDatabaseCollector) Update(ctx context.Context, server *server, ch chan<- prometheus.Metric) error {
	db, err := server.GetDB()
	if err != nil {
		return err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT pg_database.datname
		,pg_database_size(pg_database.datname)
		FROM pg_database;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname string
		var size int64
		if err := rows.Scan(&datname, &size); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgDatabase["size_bytes"],
			prometheus.GaugeValue, float64(size), datname, server.GetName(),
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
