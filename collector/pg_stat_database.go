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
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const statDatabaseSubsystem = "stat_database"

func init() {
	registerCollector(statDatabaseSubsystem, defaultEnabled, NewPGStatDatabaseCollector)
}

type PGStatDatabaseCollector struct {
	log log.Logger
}

func NewPGStatDatabaseCollector(config collectorConfig) (Collector, error) {
	return &PGStatDatabaseCollector{log: config.logger}, nil
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

func (c *PGStatDatabaseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
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

		if !datid.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no datid")
			continue
		}
		if !datname.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no datname")
			continue
		}
		if !numBackends.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no numbackends")
			continue
		}
		if !xactCommit.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no xact_commit")
			continue
		}
		if !xactRollback.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no xact_rollback")
			continue
		}
		if !blksRead.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no blks_read")
			continue
		}
		if !blksHit.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no blks_hit")
			continue
		}
		if !tupReturned.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no tup_returned")
			continue
		}
		if !tupFetched.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no tup_fetched")
			continue
		}
		if !tupInserted.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no tup_inserted")
			continue
		}
		if !tupUpdated.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no tup_updated")
			continue
		}
		if !tupDeleted.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no tup_deleted")
			continue
		}
		if !conflicts.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no conflicts")
			continue
		}
		if !tempFiles.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no temp_files")
			continue
		}
		if !tempBytes.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no temp_bytes")
			continue
		}
		if !deadlocks.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no deadlocks")
			continue
		}
		if !blkReadTime.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no blk_read_time")
			continue
		}
		if !blkWriteTime.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no blk_write_time")
			continue
		}

		statsResetMetric := 0.0
		if !statsReset.Valid {
			level.Debug(c.log).Log("msg", "No metric for stats_reset, will collect 0 instead")
		}
		if statsReset.Valid {
			statsResetMetric = float64(statsReset.Time.Unix())
		}

		labels := []string{datid.String, datname.String}

		ch <- prometheus.MustNewConstMetric(
			statDatabaseNumbackends,
			prometheus.GaugeValue,
			numBackends.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactCommit,
			prometheus.CounterValue,
			xactCommit.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseXactRollback,
			prometheus.CounterValue,
			xactRollback.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksRead,
			prometheus.CounterValue,
			blksRead.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlksHit,
			prometheus.CounterValue,
			blksHit.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupReturned,
			prometheus.CounterValue,
			tupReturned.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupFetched,
			prometheus.CounterValue,
			tupFetched.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupInserted,
			prometheus.CounterValue,
			tupInserted.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupUpdated,
			prometheus.CounterValue,
			tupUpdated.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTupDeleted,
			prometheus.CounterValue,
			tupDeleted.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseConflicts,
			prometheus.CounterValue,
			conflicts.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempFiles,
			prometheus.CounterValue,
			tempFiles.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseTempBytes,
			prometheus.CounterValue,
			tempBytes.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseDeadlocks,
			prometheus.CounterValue,
			deadlocks.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkReadTime,
			prometheus.CounterValue,
			blkReadTime.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseBlkWriteTime,
			prometheus.CounterValue,
			blkWriteTime.Float64,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			statDatabaseStatsReset,
			prometheus.CounterValue,
			statsResetMetric,
			labels...,
		)
	}
	return nil
}
