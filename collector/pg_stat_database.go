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
	registerCollector("stat_database", defaultEnabled, NewPGStatDatabaseCollector)
}

type PGStatDatabaseCollector struct{}

func NewPGStatDatabaseCollector(logger log.Logger) (Collector, error) {
	return &PGStatDatabaseCollector{}, nil
}

const statDatabaseSubsystem = "stat_database"

var statDatabase = map[string]*prometheus.Desc{
	"numbackends": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"numbackends",
		),
		"Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"xact_commit": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"xact_commit",
		),
		"Number of transactions in this database that have been committed",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"xact_rollback": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"xact_rollback",
		),
		"Number of transactions in this database that have been rolled back",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"blks_read": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blks_read",
		),
		"Number of disk blocks read in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"blks_hit": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blks_hit",
		),
		"Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"tup_returned": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_returned",
		),
		"Number of rows returned by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"tup_fetched": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_fetched",
		),
		"Number of rows fetched by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"tup_inserted": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_inserted",
		),
		"Number of rows inserted by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"tup_updated": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_updated",
		),
		"Number of rows updated by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"tup_deleted": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"tup_deleted",
		),
		"Number of rows deleted by queries in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"conflicts": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"conflicts",
		),
		"Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"temp_files": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"temp_files",
		),
		"Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"temp_bytes": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"temp_bytes",
		),
		"Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"deadlocks": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"deadlocks",
		),
		"Number of deadlocks detected in this database",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"blk_read_time": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blk_read_time",
		),
		"Time spent reading data file blocks by backends in this database, in milliseconds",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"blk_write_time": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"blk_write_time",
		),
		"Time spent writing data file blocks by backends in this database, in milliseconds",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
	"stats_reset": prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statDatabaseSubsystem,
			"stats_reset",
		),
		"Time at which these statistics were last reset",
		[]string{"datid", "datname"},
		prometheus.Labels{},
	),
}

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
			statDatabase["numbackends"],
			prometheus.GaugeValue,
			numBackends,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["xact_commit"],
			prometheus.CounterValue,
			xactCommit,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["xact_rollback"],
			prometheus.CounterValue,
			xactRollback,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["blks_read"],
			prometheus.CounterValue,
			blksRead,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["blks_hit"],
			prometheus.CounterValue,
			blksHit,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["tup_returned"],
			prometheus.CounterValue,
			tupReturned,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["tup_fetched"],
			prometheus.CounterValue,
			tupFetched,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["tup_inserted"],
			prometheus.CounterValue,
			tupInserted,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["tup_updated"],
			prometheus.CounterValue,
			tupUpdated,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["tup_deleted"],
			prometheus.CounterValue,
			tupDeleted,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["conflicts"],
			prometheus.CounterValue,
			conflicts,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["temp_files"],
			prometheus.CounterValue,
			tempFiles,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["temp_bytes"],
			prometheus.CounterValue,
			tempBytes,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["deadlocks"],
			prometheus.CounterValue,
			deadlocks,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["blk_read_time"],
			prometheus.CounterValue,
			blkReadTime,
			datid,
			datname,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabase["blk_write_time"],
			prometheus.CounterValue,
			blkWriteTime,
			datid,
			datname,
		)

		if statsReset.Valid {
			ch <- prometheus.MustNewConstMetric(
				statDatabase["stats_reset"],
				prometheus.CounterValue,
				float64(statsReset.Time.Unix()),
				datid,
				datname,
			)
		} else {
			ch <- prometheus.MustNewConstMetric(
				statDatabase["stats_reset"],
				prometheus.CounterValue,
				0,
				datid,
				datname,
			)
		}
	}
	return nil
}
