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

const locksSubsystem = "locks"

func init() {
	registerCollector(locksSubsystem, defaultEnabled, NewPGLocksCollector)
}

type PGLocksCollector struct {
	log log.Logger
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
		[]string{"datname", "mode"}, nil,
	)

	pgLocksQuery = `
		SELECT 
		  pg_database.datname as datname,
		  tmp.mode as mode,
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
		  ) AS tmp(mode)
		  CROSS JOIN pg_database 
		  LEFT JOIN (
		    SELECT 
		      database, 
		      lower(mode) AS mode, 
		      count(*) AS count 
		    FROM 
		      pg_locks 
		    WHERE 
		      database IS NOT NULL 
		    GROUP BY 
		      database, 
		      lower(mode)
		  ) AS tmp2 ON tmp.mode = tmp2.mode 
		  and pg_database.oid = tmp2.database 
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

	var datname, mode sql.NullString
	var count sql.NullInt64

	for rows.Next() {
		if err := rows.Scan(&datname, &mode, &count); err != nil {
			return err
		}

		if !datname.Valid || !mode.Valid {
			continue
		}

		countMetric := 0.0
		if count.Valid {
			countMetric = float64(count.Int64)
		}

		ch <- prometheus.MustNewConstMetric(
			pgLocksDesc,
			prometheus.GaugeValue, countMetric,
			datname.String, mode.String,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
