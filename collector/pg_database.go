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
	log               log.Logger
	excludedDatabases []string
}

func NewPGDatabaseCollector(config collectorConfig) (Collector, error) {
	exclude := config.excludeDatabases
	if exclude == nil {
		exclude = []string{}
	}
	return &PGDatabaseCollector{
		log:               config.logger,
		excludedDatabases: exclude,
	}, nil
}

var pgDatabaseSizeBytes = prometheus.NewDesc(
	"pg_database_size_bytes",
	"Disk space used by the database",
	[]string{"datname"}, nil,
)

var pgDatabaseQuery = "SELECT pg_database.datname FROM pg_database;"
var pgDatabaseSizeQuery = "SELECT pg_database_size($1)"

// Update implements Collector and exposes database size.
// It is called by the Prometheus registry when collecting metrics.
// The list of databases is retrieved from pg_database and filtered
// by the excludeDatabase config parameter. The tradeoff here is that
// we have to query the list of databases and then query the size of
// each database individually. This is because we can't filter the
// list of databases in the query because the list of excluded
// databases is dynamic.
func (c PGDatabaseCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	// Query the list of databases
	rows, err := db.QueryContext(ctx,
		pgDatabaseQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var databases []string

	for rows.Next() {
		var datname string
		if err := rows.Scan(&datname); err != nil {
			return err
		}

		// Ignore excluded databases
		// Filtering is done here instead of in the query to avoid
		// a complicated NOT IN query with a variable number of parameters
		if sliceContains(c.excludedDatabases, datname) {
			continue
		}

		databases = append(databases, datname)
	}

	// Query the size of the databases
	for _, datname := range databases {
		var size int64
		err = db.QueryRowContext(ctx, pgDatabaseSizeQuery, datname).Scan(&size)
		if err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgDatabaseSizeBytes,
			prometheus.GaugeValue, float64(size), datname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func sliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
