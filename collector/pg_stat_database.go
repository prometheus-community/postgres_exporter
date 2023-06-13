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

	"github.com/prometheus/client_golang/prometheus"
)

const statDatabaseSubsystem = "stat_database"

func init() {
	registerCollector(statDatabaseSubsystem, defaultEnabled, NewPGStatDatabaseCollector)
}

type PGStatDatabaseCollector struct{}

func NewPGStatDatabaseCollector(config collectorConfig) (Collector, error) {
	return &PGStatDatabaseCollector{}, nil
}

var (
	statDatabaseNumbackends = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"numbackends",
		),
		"Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseXactCommit = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"xact_commit",
		),
		"Number of transactions in this database that have been committed",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseXactRollback = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"xact_rollback",
		),
		"Number of transactions in this database that have been rolled back",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseBlksRead = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blks_read",
		),
		"Number of disk blocks read in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseBlksHit = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blks_hit",
		),
		"Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTupReturned = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_returned",
		),
		"Number of rows returned by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTupFetched = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_fetched",
		),
		"Number of rows fetched by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTupInserted = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_inserted",
		),
		"Number of rows inserted by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTupUpdated = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_updated",
		),
		"Number of rows updated by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTupDeleted = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_deleted",
		),
		"Number of rows deleted by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseConflicts = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"conflicts",
		),
		"Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTempFiles = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"temp_files",
		),
		"Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseTempBytes = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"temp_bytes",
		),
		"Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseDeadlocks = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"deadlocks",
		),
		"Number of deadlocks detected in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseBlkReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blk_read_time",
		),
		"Time spent reading data file blocks by backends in this database, in milliseconds",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseBlkWriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blk_write_time",
		),
		"Time spent writing data file blocks by backends in this database, in milliseconds",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
	statDatabaseStatsReset = prometheus.NewDesc(prometheus.BuildFQName(
		namespace,
		statDatabaseSubsystem,
		"stats_reset",
	),
		"Time at which these statistics were last reset",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
)

func (PGStatDatabaseCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		`SELECT
			datid
			,datname
			,numbackends
			,xact_commit
			,xact_rollback
			,blks_read
			,blks_hit
			,tup_returned
			,tup_fetched
			,tup_inserted
			,tup_updated
			,tup_deleted
			,conflicts
			,temp_files
			,temp_bytes
			,deadlocks
			,blk_read_time
			,blk_write_time
			,stats_reset
		FROM pg_stat_database;
		`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datid string
		var datname string
		var numBackends float64
		var xactCommit float64
		var xactRollback float64
		var blksRead float64
		var blksHit float64
		var tupReturned float64
		var tupFetched float64
		var tupInserted float64
		var tupUpdated float64
		var tupDeleted float64
		var conflicts float64
		var tempFiles float64
		var tempBytes float64
		var deadlocks float64
		var blkReadTime float64
		var blkWriteTime float64
		var statsReset sql.NullTime

		err := rows.Scan(
			&datid,
			&datname,
			&numBackends,
			&xactCommit,
			&xactRollback,
			&blksRead,
			&blksHit,
			&tupReturned,
			&tupFetched,
			&tupInserted,
			&tupUpdated,
			&tupDeleted,
			&conflicts,
			&tempFiles,
			&tempBytes,
			&deadlocks,
			&blkReadTime,
			&blkWriteTime,
			&statsReset,
		)
		if err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statDatabaseNumbackends,
			prometheus.GaugeValue,
			numBackends,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactCommit,
			prometheus.CounterValue,
			xactCommit,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactRollback,
			prometheus.CounterValue,
			xactRollback,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksRead,
			prometheus.CounterValue,
			blksRead,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksHit,
			prometheus.CounterValue,
			blksHit,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupReturned,
			prometheus.CounterValue,
			tupReturned,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupFetched,
			prometheus.CounterValue,
			tupFetched,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupInserted,
			prometheus.CounterValue,
			tupInserted,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupUpdated,
			prometheus.CounterValue,
			tupUpdated,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupDeleted,
			prometheus.CounterValue,
			tupDeleted,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseConflicts,
			prometheus.CounterValue,
			conflicts,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempFiles,
			prometheus.CounterValue,
			tempFiles,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempBytes,
			prometheus.CounterValue,
			tempBytes,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseDeadlocks,
			prometheus.CounterValue,
			deadlocks,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkReadTime,
			prometheus.CounterValue,
			blkReadTime,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkWriteTime,
			prometheus.CounterValue,
			blkWriteTime,
			datid,
			datname,
		)

		if statsReset.Valid {
			ch <- prometheus.MustNewConstMetric(
				statDatabaseStatsReset,
				prometheus.CounterValue,
				float64(statsReset.Time.Unix()),
				datid,
				datname,
			)
		} else {
			ch <- prometheus.MustNewConstMetric(
				statDatabaseStatsReset,
				prometheus.CounterValue,
				0,
				datid,
				datname,
			)
		}
	}
	return nil
}
