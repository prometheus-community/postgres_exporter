// Copyright The Prometheus Authors
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

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(statWALSubsystem, serverScope, NewPGStatWALCollector)
}

type PGStatWALCollector struct {
	log *slog.Logger
}

func NewPGStatWALCollector(config collectorConfig) (Collector, error) {
	return &PGStatWALCollector{log: config.logger}, nil
}

var (
	statWALRecordsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "records_total"),
		"Total number of WAL records generated.",
		[]string{},
		prometheus.Labels{},
	)
	statWALFPIDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "fpi_total"),
		"Total number of WAL full page images generated.",
		[]string{},
		prometheus.Labels{},
	)
	statWALBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "bytes_total"),
		"Total amount of WAL generated in bytes.",
		[]string{},
		prometheus.Labels{},
	)
	statWALBuffersFullDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "buffers_full_total"),
		"Number of times WAL data was written to disk because WAL buffers became full.",
		[]string{},
		prometheus.Labels{},
	)
	statWALWriteDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "write_total"),
		"Number of times WAL buffers were written out to disk via XLogWrite request.",
		[]string{},
		prometheus.Labels{},
	)
	statWALSyncDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "sync_total"),
		"Number of times WAL files were synced to disk via issue_xlog_fsync request.",
		[]string{},
		prometheus.Labels{},
	)
	statWALWriteTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "write_time_seconds_total"),
		"Total time spent writing WAL buffers to disk in seconds.",
		[]string{},
		prometheus.Labels{},
	)
	statWALSyncTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "sync_time_seconds_total"),
		"Total time spent syncing WAL files to disk in seconds.",
		[]string{},
		prometheus.Labels{},
	)
	statWALResetTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWALSubsystem, "reset_time_seconds"),
		"Unix timestamp of when WAL statistics were last reset.",
		[]string{},
		prometheus.Labels{},
	)

	statWALQueryBefore18 = `SELECT
		wal_records AS records,
		wal_fpi AS fpi,
		wal_bytes AS bytes,
		wal_buffers_full AS buffers_full,
		wal_write AS write,
		wal_sync AS sync,
		wal_write_time AS write_time,
		wal_sync_time AS sync_time,
		extract(EPOCH FROM stats_reset) AS reset_time
	FROM
		pg_stat_wal;`

	statWALQuery18AndLater = `SELECT
		wal_records AS records,
		wal_fpi AS fpi,
		wal_bytes AS bytes,
		wal_buffers_full AS buffers_full,
		extract(EPOCH FROM stats_reset) AS reset_time
	FROM
		pg_stat_wal;`
)

func (c PGStatWALCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if instance.version.LT(semver.MustParse("14.0.0")) {
		if c.log != nil {
			c.log.Debug("stat_wal collector is not available on PostgreSQL < 14.0.0, skipping")
		}
		return nil
	}

	pg18 := instance.version.GE(semver.MustParse("18.0.0"))

	var records, fpi, buffersFull sql.NullInt64
	var bytes, resetTime sql.NullFloat64
	var write, sync sql.NullInt64
	var writeTimeMS, syncTimeMS sql.NullFloat64
	if pg18 {
		row := instance.getDB().QueryRowContext(ctx, statWALQuery18AndLater)
		if err := row.Scan(&records, &fpi, &bytes, &buffersFull, &resetTime); err != nil {
			return err
		}
	} else {
		row := instance.getDB().QueryRowContext(ctx, statWALQueryBefore18)
		if err := row.Scan(&records, &fpi, &bytes, &buffersFull, &write, &sync, &writeTimeMS, &syncTimeMS, &resetTime); err != nil {
			return err
		}
	}

	ch <- prometheus.MustNewConstMetric(statWALRecordsDesc, prometheus.CounterValue, int64CounterValue(records, instance.wrapLargeCounters))
	ch <- prometheus.MustNewConstMetric(statWALFPIDesc, prometheus.CounterValue, int64CounterValue(fpi, instance.wrapLargeCounters))
	ch <- prometheus.MustNewConstMetric(statWALBytesDesc, prometheus.CounterValue, nullFloat64Value(bytes))
	ch <- prometheus.MustNewConstMetric(statWALBuffersFullDesc, prometheus.CounterValue, int64CounterValue(buffersFull, instance.wrapLargeCounters))
	if !pg18 {
		ch <- prometheus.MustNewConstMetric(statWALWriteDesc, prometheus.CounterValue, int64CounterValue(write, instance.wrapLargeCounters))
		ch <- prometheus.MustNewConstMetric(statWALSyncDesc, prometheus.CounterValue, int64CounterValue(sync, instance.wrapLargeCounters))
		ch <- prometheus.MustNewConstMetric(statWALWriteTimeDesc, prometheus.CounterValue, nullMillisecondsToSeconds(writeTimeMS))
		ch <- prometheus.MustNewConstMetric(statWALSyncTimeDesc, prometheus.CounterValue, nullMillisecondsToSeconds(syncTimeMS))
	}
	if resetTime.Valid {
		ch <- prometheus.MustNewConstMetric(statWALResetTimeDesc, prometheus.GaugeValue, resetTime.Float64)
	}

	return nil
}
