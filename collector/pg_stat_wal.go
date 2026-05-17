// Copyright 2024 The Prometheus Authors
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
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const statWALSubsystem = "stat_wal"

func init() {
	registerCollector(statWALSubsystem, defaultDisabled, NewPGStatWALCollector)
}

type PGStatWALCollector struct {
	log *slog.Logger
}

func NewPGStatWALCollector(config collectorConfig) (Collector, error) {
	return &PGStatWALCollector{log: config.logger}, nil
}

var statsWALRecordsDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_records_total"),
	"Total number of WAL records generated",
	[]string{},
	prometheus.Labels{},
)

var statsWALFPIDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_fpi"),
	"Total number of WAL full page images generated",
	[]string{},
	prometheus.Labels{},
)

var statsWALBytesDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_bytes"),
	"Total amount of WAL generated in bytes",
	[]string{},
	prometheus.Labels{},
)

var statsWALBuffersFullDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_buffers_full"),
	"Number of times WAL data was written to disk because WAL buffers became full",
	[]string{},
	prometheus.Labels{},
)

var statsWALWriteDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_write"),
	"Number of times WAL buffers were written out to disk via XLogWrite request. See Section 30.5 for more information about the internal WAL function XLogWrite.",
	[]string{},
	prometheus.Labels{},
)

var statsWALSyncDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_sync"),
	"Number of times WAL files were synced to disk via issue_xlog_fsync request (if fsync is on and wal_sync_method is either fdatasync, fsync or fsync_writethrough, otherwise zero). See Section 30.5 for more information about the internal WAL function issue_xlog_fsync.",
	[]string{},
	prometheus.Labels{},
)

var statsWALWriteTimeDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_write_time"),
	"Total amount of time spent writing WAL buffers to disk via XLogWrite request, in milliseconds (if track_wal_io_timing is enabled, otherwise zero). This includes the sync time when wal_sync_method is either open_datasync or open_sync.",
	[]string{},
	prometheus.Labels{},
)

var statsWALSyncTimeDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "wal_sync_time"),
	"Total amount of time spent syncing WAL files to disk via issue_xlog_fsync request, in milliseconds (if track_wal_io_timing is enabled, fsync is on, and wal_sync_method is either fdatasync, fsync or fsync_writethrough, otherwise zero).",
	[]string{},
	prometheus.Labels{},
)

var statsWALStatsResetDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statWALSubsystem, "stats_reset"),
	"Time at which these statistics were last reset",
	[]string{},
	prometheus.Labels{},
)

func statWALQuery(columns []string) string {
	return fmt.Sprintf("SELECT %s FROM pg_stat_wal;", strings.Join(columns, ","))
}

func (c *PGStatWALCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	columns := []string{
		"wal_records",      // bigint
		"wal_fpi",          // bigint
		"wal_bytes",        // numeric
		"wal_buffers_full", // bigint
		"wal_write",        // bigint
		"wal_sync",         // bigint
		"wal_write_time",   // double precision
		"wal_sync_time",    // double precision
		"stats_reset",      // timestamp with time zone
	}

	rows, err := db.QueryContext(ctx,
		statWALQuery(columns),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var walRecords, walFPI, walBytes, walBuffersFull, walWrite, walSync sql.NullInt64
		var walWriteTime, walSyncTime sql.NullFloat64
		var statsReset sql.NullTime

		err := rows.Scan(
			&walRecords,
			&walFPI,
			&walBytes,
			&walBuffersFull,
			&walWrite,
			&walSync,
			&walWriteTime,
			&walSyncTime,
			&statsReset,
		)
		if err != nil {
			return err
		}

		walRecordsMetric := 0.0
		if walRecords.Valid {
			walRecordsMetric = float64(walRecords.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALRecordsDesc,
			prometheus.CounterValue,
			walRecordsMetric,
		)

		walFPIMetric := 0.0
		if walFPI.Valid {
			walFPIMetric = float64(walFPI.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALFPIDesc,
			prometheus.CounterValue,
			walFPIMetric,
		)

		walBytesMetric := 0.0
		if walBytes.Valid {
			walBytesMetric = float64(walBytes.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALBytesDesc,
			prometheus.CounterValue,
			walBytesMetric,
		)

		walBuffersFullMetric := 0.0
		if walBuffersFull.Valid {
			walBuffersFullMetric = float64(walBuffersFull.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALBuffersFullDesc,
			prometheus.CounterValue,
			walBuffersFullMetric,
		)

		walWriteMetric := 0.0
		if walWrite.Valid {
			walWriteMetric = float64(walWrite.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALWriteDesc,
			prometheus.CounterValue,
			walWriteMetric,
		)

		walSyncMetric := 0.0
		if walSync.Valid {
			walSyncMetric = float64(walSync.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALSyncDesc,
			prometheus.CounterValue,
			walSyncMetric,
		)

		walWriteTimeMetric := 0.0
		if walWriteTime.Valid {
			walWriteTimeMetric = float64(walWriteTime.Float64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALWriteTimeDesc,
			prometheus.CounterValue,
			walWriteTimeMetric,
		)

		walSyncTimeMetric := 0.0
		if walSyncTime.Valid {
			walSyncTimeMetric = float64(walSyncTime.Float64)
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALSyncTimeDesc,
			prometheus.CounterValue,
			walSyncTimeMetric,
		)

		resetMetric := 0.0
		if statsReset.Valid {
			resetMetric = float64(statsReset.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statsWALStatsResetDesc,
			prometheus.CounterValue,
			resetMetric,
		)
	}
	return nil
}
