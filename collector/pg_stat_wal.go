// Copyright 2026 The Prometheus Authors
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
	"fmt"
	"log/slog"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Disabled by default because pg_stat_wal is only available from PostgreSQL 14.
	registerCollector(statWALSubsystem, serverScope, NewPGStatWALCollector)
}

type PGStatWALCollector struct {
	log *slog.Logger
}

func NewPGStatWALCollector(config collectorConfig) (Collector, error) {
	return &PGStatWALCollector{log: config.logger}, nil
}

var statsWALRecordsDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "records_total"),
	"Total number of WAL records generated",
	[]string{},
	prometheus.Labels{},
)

var statsWALFPIDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "fpi_total"),
	"Total number of WAL full page images generated",
	[]string{},
	prometheus.Labels{},
)

var statsWALBytesDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "bytes_total"),
	"Total amount of WAL generated in bytes",
	[]string{},
	prometheus.Labels{},
)

var statsWALBuffersFullDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "buffers_full_total"),
	"Number of times WAL data was written to disk because WAL buffers became full",
	[]string{},
	prometheus.Labels{},
)

var statsWALWriteDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "write_total"),
	"Number of times WAL buffers were written out to disk via XLogWrite request. See Section 30.5 for more information about the internal WAL function XLogWrite.",
	[]string{},
	prometheus.Labels{},
)

var statsWALSyncDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "sync_total"),
	"Number of times WAL files were synced to disk via issue_xlog_fsync request (if fsync is on and wal_sync_method is either fdatasync, fsync or fsync_writethrough, otherwise zero). See Section 30.5 for more information about the internal WAL function issue_xlog_fsync.",
	[]string{},
	prometheus.Labels{},
)

var statsWALWriteTimeDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "write_time_seconds_total"),
	"Total amount of time spent writing WAL buffers to disk via XLogWrite request, in seconds (if track_wal_io_timing is enabled, otherwise zero). This includes the sync time when wal_sync_method is either open_datasync or open_sync.",
	[]string{},
	prometheus.Labels{},
)

var statsWALSyncTimeDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "sync_time_seconds_total"),
	"Total amount of time spent syncing WAL files to disk via issue_xlog_fsync request, in seconds (if track_wal_io_timing is enabled, fsync is on, and wal_sync_method is either fdatasync, fsync or fsync_writethrough, otherwise zero).",
	[]string{},
	prometheus.Labels{},
)

var statsWALStatsResetDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "stats_reset_timestamp_seconds"),
	"Time at which these statistics were last reset",
	[]string{},
	prometheus.Labels{},
)

// statWALQuery is used on PostgreSQL 14 through 17.
const statWALQuery = `SELECT
	wal_records
	,wal_fpi
	,wal_bytes
	,wal_buffers_full
	,wal_write
	,wal_sync
	,wal_write_time
	,wal_sync_time
	,stats_reset
FROM pg_stat_wal;`

// statWALQuery18 is used on PostgreSQL 18+, which moved wal_write, wal_sync,
// wal_write_time and wal_sync_time to pg_stat_io.
const statWALQuery18 = `SELECT
	wal_records
	,wal_fpi
	,wal_bytes
	,wal_buffers_full
	,stats_reset
FROM pg_stat_wal;`

// statWALIOQuery collects the WAL write and sync counters that PostgreSQL 18
// moved from pg_stat_wal to pg_stat_io. Only context 'normal' is counted: it
// holds the XLogWrite and issue_xlog_fsync activity the pre-18 columns
// reported, whereas context 'init' tracks WAL segment creation.
const statWALIOQuery = `SELECT
	sum(writes)
	,sum(write_time)
	,sum(fsyncs)
	,sum(fsync_time)
FROM pg_stat_io
WHERE object = 'wal'
	AND context = 'normal';`

func (c *PGStatWALCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if instance.version.LT(semver.MustParse("14.0.0")) {
		c.log.Debug("pg_stat_wal collector requires PostgreSQL 14 or newer, skipping")
		return nil
	}

	db := instance.getDB()

	var walRecords, walFPI, walBytes, walBuffersFull, walWrite, walSync sql.NullInt64
	var walWriteTime, walSyncTime sql.NullFloat64
	var statsReset sql.NullTime

	if instance.version.GE(semver.MustParse("18.0.0")) {
		if err := db.QueryRowContext(ctx, statWALQuery18).Scan(
			&walRecords,
			&walFPI,
			&walBytes,
			&walBuffersFull,
			&statsReset,
		); err != nil {
			return fmt.Errorf("failed to query pg_stat_wal on PostgreSQL 18 or newer: %w", err)
		}
		if err := db.QueryRowContext(ctx, statWALIOQuery).Scan(
			&walWrite,
			&walWriteTime,
			&walSync,
			&walSyncTime,
		); err != nil {
			return fmt.Errorf("failed to query pg_stat_io WAL statistics: %w", err)
		}
	} else {
		if err := db.QueryRowContext(ctx, statWALQuery).Scan(
			&walRecords,
			&walFPI,
			&walBytes,
			&walBuffersFull,
			&walWrite,
			&walSync,
			&walWriteTime,
			&walSyncTime,
			&statsReset,
		); err != nil {
			return fmt.Errorf("failed to query pg_stat_wal: %w", err)
		}
	}

	if walRecords.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALRecordsDesc,
			prometheus.CounterValue,
			int64CounterValue(walRecords, instance.wrapLargeCounters),
		)
	}

	if walFPI.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALFPIDesc,
			prometheus.CounterValue,
			int64CounterValue(walFPI, instance.wrapLargeCounters),
		)
	}

	if walBytes.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALBytesDesc,
			prometheus.CounterValue,
			int64CounterValue(walBytes, instance.wrapLargeCounters),
		)
	}

	if walBuffersFull.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALBuffersFullDesc,
			prometheus.CounterValue,
			int64CounterValue(walBuffersFull, instance.wrapLargeCounters),
		)
	}

	if walWrite.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALWriteDesc,
			prometheus.CounterValue,
			int64CounterValue(walWrite, instance.wrapLargeCounters),
		)
	}

	if walSync.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALSyncDesc,
			prometheus.CounterValue,
			int64CounterValue(walSync, instance.wrapLargeCounters),
		)
	}

	if walWriteTime.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALWriteTimeDesc,
			prometheus.CounterValue,
			walWriteTime.Float64/1000,
		)
	}

	if walSyncTime.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALSyncTimeDesc,
			prometheus.CounterValue,
			walSyncTime.Float64/1000,
		)
	}

	if statsReset.Valid {
		ch <- prometheus.MustNewConstMetric(
			statsWALStatsResetDesc,
			prometheus.GaugeValue,
			float64(statsReset.Time.Unix()),
		)
	}

	return nil
}
