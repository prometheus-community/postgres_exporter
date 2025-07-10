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
	"log/slog"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
)

const databaseSubsystem = "database"

func init() {
	registerCollector(databaseSubsystem, defaultEnabled, NewPGDatabaseCollector)
}

type PGDatabaseCollector struct {
	log               *slog.Logger
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

var (
	pgDatabaseSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			databaseSubsystem,
			"size_bytes",
		),
		"Disk space used by the database",
		[]string{"datname"}, nil,
	)
	pgDatabaseConnectionLimitsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			databaseSubsystem,
			"connection_limit",
		),
		"Connection limit set for the database",
		[]string{"datname"}, nil,
	)

	pgDatabaseQuery     = "SELECT pg_database.datname, pg_database.datconnlimit FROM pg_database;"
	pgDatabaseSizeQuery = "SELECT pg_database_size($1)"
)

// Update implements Collector and exposes database size and connection limits.
// It is called by the Prometheus registry when collecting metrics.
// The list of databases is retrieved from pg_database and filtered
// by the excludeDatabase config parameter. The tradeoff here is that
// we have to query the list of databases and then query the size of
// each database individually. This is because we can't filter the
// list of databases in the query because the list of excluded
// databases is dynamic.
func (c PGDatabaseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
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
		var datname sql.NullString
		var connLimit sql.NullInt64
		if err := rows.Scan(&datname, &connLimit); err != nil {
			return err
		}

		if !datname.Valid {
			continue
		}
		database := datname.String
		// Ignore excluded databases
		// Filtering is done here instead of in the query to avoid
		// a complicated NOT IN query with a variable number of parameters
		if slices.Contains(c.excludedDatabases, database) {
			continue
		}

		databases = append(databases, database)

		connLimitMetric := 0.0
		if connLimit.Valid {
			connLimitMetric = float64(connLimit.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			pgDatabaseConnectionLimitsDesc,
			prometheus.GaugeValue, connLimitMetric, database,
		)
	}

	// Query the size of the databases
	for _, datname := range databases {
		var size sql.NullFloat64
		err = db.QueryRowContext(ctx, pgDatabaseSizeQuery, datname).Scan(&size)
		if err != nil {
			return err
		}

		sizeMetric := 0.0
		if size.Valid {
			sizeMetric = size.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			pgDatabaseSizeDesc,
			prometheus.GaugeValue, sizeMetric, datname,
		)

	}
	return rows.Err()
}
