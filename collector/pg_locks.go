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
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const locksSubsystem = "locks"

func init() {
	registerCollector(locksSubsystem, defaultEnabled, NewPGLocksCollector)
}

type PGLocksCollector struct {
	log *slog.Logger
}

func NewPGLocksCollector(config collectorConfig) (Collector, error) {
	return &PGLocksCollector{
		log: config.logger,
	}, nil
}

var (
	pgLocksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			locksSubsystem,
			"count",
		),
		"Number of locks",
		[]string{"datname", "mode", "granted"}, nil,
	)

	pgLocksQuery = `
		SELECT
		  pg_database.datname as datname,
		  tmp_mode.mode as mode,
		  tmp_granted.granted as granted,
		  COALESCE(count, 0) as count
		FROM
		  (
		    VALUES
		      ('accesssharelock'),
		      ('rowsharelock'),
		      ('rowexclusivelock'),
		      ('shareupdateexclusivelock'),
		      ('sharelock'),
		      ('sharerowexclusivelock'),
		      ('exclusivelock'),
		      ('accessexclusivelock'),
		      ('sireadlock')
		  ) AS tmp_mode(mode)
		  CROSS JOIN (
		    VALUES
		      ('true'),
		      ('false')
		  ) AS tmp_granted(granted)
		  CROSS JOIN pg_database
		  LEFT JOIN (
		    SELECT
		      database,
		      lower(mode) AS mode,
		      granted::text AS granted,
		      count(*) AS count
		    FROM
		      pg_locks
		    WHERE
		      database IS NOT NULL
		    GROUP BY
		      database,
		      lower(mode),
		      granted
		  ) AS tmp2 ON tmp_mode.mode = tmp2.mode
		  AND tmp_granted.granted = tmp2.granted
		  AND pg_database.oid = tmp2.database
		ORDER BY
		  1
	`
)

// Update implements Collector and exposes database locks.
// It is called by the Prometheus registry when collecting metrics.
func (c PGLocksCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx,
		pgLocksQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var datname, mode, granted sql.NullString
	var count sql.NullInt64

	for rows.Next() {
		if err := rows.Scan(&datname, &mode, &granted, &count); err != nil {
			return err
		}

		if !datname.Valid || !mode.Valid || !granted.Valid {
			continue
		}

		countMetric := 0.0
		if count.Valid {
			countMetric = float64(count.Int64)
		}

		ch <- prometheus.MustNewConstMetric(
			pgLocksDesc,
			prometheus.GaugeValue, countMetric,
			datname.String, mode.String, granted.String,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
