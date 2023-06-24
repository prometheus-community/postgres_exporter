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

	statDatabaseQuery = `
		SELECT
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
	`
)

func (PGStatDatabaseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statDatabaseQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datid, datname sql.NullString
		var numBackends, xactCommit, xactRollback, blksRead, blksHit, tupReturned, tupFetched, tupInserted, tupUpdated, tupDeleted, conflicts, tempFiles, tempBytes, deadlocks, blkReadTime, blkWriteTime sql.NullFloat64
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
		datidLabel := "unknown"
		if datid.Valid {
			datidLabel = datid.String
		}
		datnameLabel := "unknown"
		if datname.Valid {
			datnameLabel = datname.String
		}

		var numBackendsMetric float64
		if numBackends.Valid {
			numBackendsMetric = numBackends.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseNumbackends,
			prometheus.GaugeValue,
			numBackendsMetric,
			datidLabel,
			datnameLabel,
		)

		var xactCommitMetric float64
		if xactCommit.Valid {
			xactCommitMetric = xactCommit.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactCommit,
			prometheus.CounterValue,
			xactCommitMetric,
			datidLabel,
			datnameLabel,
		)

		var xactRollbackMetric float64
		if xactRollback.Valid {
			xactRollbackMetric = xactRollback.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactRollback,
			prometheus.CounterValue,
			xactRollbackMetric,
			datidLabel,
			datnameLabel,
		)

		var blksReadMetric float64
		if blksRead.Valid {
			blksReadMetric = blksRead.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksRead,
			prometheus.CounterValue,
			blksReadMetric,
			datidLabel,
			datnameLabel,
		)

		var blksHitMetric float64
		if blksHit.Valid {
			blksHitMetric = blksHit.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksHit,
			prometheus.CounterValue,
			blksHitMetric,
			datidLabel,
			datnameLabel,
		)

		var tupReturnedMetric float64
		if tupReturned.Valid {
			tupReturnedMetric = tupReturned.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupReturned,
			prometheus.CounterValue,
			tupReturnedMetric,
			datidLabel,
			datnameLabel,
		)

		var tupFetchedMetric float64
		if tupFetched.Valid {
			tupFetchedMetric = tupFetched.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupFetched,
			prometheus.CounterValue,
			tupFetchedMetric,
			datidLabel,
			datnameLabel,
		)

		var tupInsertedMetric float64
		if tupInserted.Valid {
			tupInsertedMetric = tupInserted.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupInserted,
			prometheus.CounterValue,
			tupInsertedMetric,
			datidLabel,
			datnameLabel,
		)

		var tupUpdatedMetric float64
		if tupUpdated.Valid {
			tupUpdatedMetric = tupUpdated.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupUpdated,
			prometheus.CounterValue,
			tupUpdatedMetric,
			datidLabel,
			datnameLabel,
		)

		var tupDeletedMetric float64
		if tupDeleted.Valid {
			tupDeletedMetric = tupDeleted.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupDeleted,
			prometheus.CounterValue,
			tupDeletedMetric,
			datidLabel,
			datnameLabel,
		)

		var conflictsMetric float64
		if conflicts.Valid {
			conflictsMetric = conflicts.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseConflicts,
			prometheus.CounterValue,
			conflictsMetric,
			datidLabel,
			datnameLabel,
		)

		var tempFilesMetric float64
		if tempFiles.Valid {
			tempFilesMetric = tempFiles.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempFiles,
			prometheus.CounterValue,
			tempFilesMetric,
			datidLabel,
			datnameLabel,
		)

		var tempBytesMetric float64
		if tempBytes.Valid {
			tempBytesMetric = tempBytes.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempBytes,
			prometheus.CounterValue,
			tempBytesMetric,
			datidLabel,
			datnameLabel,
		)

		var deadlocksMetric float64
		if deadlocks.Valid {
			deadlocksMetric = deadlocks.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseDeadlocks,
			prometheus.CounterValue,
			deadlocksMetric,
			datidLabel,
			datnameLabel,
		)

		var blkReadTimeMetric float64
		if blkReadTime.Valid {
			blkReadTimeMetric = blkReadTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkReadTime,
			prometheus.CounterValue,
			blkReadTimeMetric,
			datidLabel,
			datnameLabel,
		)

		var blkWriteTimeMetric float64
		if blkWriteTime.Valid {
			blkWriteTimeMetric = blkWriteTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkWriteTime,
			prometheus.CounterValue,
			blkWriteTimeMetric,
			datidLabel,
			datnameLabel,
		)

		var statsResetMetric float64
		if statsReset.Valid {
			statsResetMetric = float64(statsReset.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statDatabaseStatsReset,
			prometheus.CounterValue,
			statsResetMetric,
			datidLabel,
			datnameLabel,
		)
	}
	return nil
}
