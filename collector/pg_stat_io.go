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
	registerCollector(statIOSubsystem, serverScope, NewPGStatIOCollector)
}

type PGStatIOCollector struct {
	log *slog.Logger
}

func NewPGStatIOCollector(config collectorConfig) (Collector, error) {
	return &PGStatIOCollector{log: config.logger}, nil
}

var (
	statIOLabelNames = []string{"backend_type", "object", "context"}
	statIOReadsDesc  = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "reads_total"),
		"Number of read operations.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOReadBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "read_bytes_total"),
		"Number of read bytes.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOReadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "read_time_seconds_total"),
		"Time spent in read operations in seconds.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOWritesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "writes_total"),
		"Number of write operations.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOWriteBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "write_bytes_total"),
		"Number of write bytes.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOWriteTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "write_time_seconds_total"),
		"Time spent in write operations in seconds.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOWritebacksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "writebacks_total"),
		"Number of writeback requests the process issued to the kernel to write data out to permanent storage.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOWritebackTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "writeback_time_seconds_total"),
		"Time spent in writeback operations in seconds.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOExtendsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "extends_total"),
		"Number of relation extend operations.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOExtendBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "extend_bytes_total"),
		"Number of extend bytes.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOExtendTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "extend_time_seconds_total"),
		"Time spent in extend operations in seconds.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOHitsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "hits_total"),
		"The number of times a desired block was found in a shared buffer.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOEvictionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "evictions_total"),
		"Number of times a block has been written out from a shared or local buffer.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOReusesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "reuses_total"),
		"The number of times an existing buffer is reused.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOFsyncsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "fsyncs_total"),
		"Number of fsync calls. These are only tracked in context normal.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOFsyncTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "fsync_time_seconds_total"),
		"Time spent in fsync operations in seconds.",
		statIOLabelNames,
		prometheus.Labels{},
	)
	statIOResetTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "reset_time_seconds"),
		"Timestamp at which these statistics were last reset, in seconds since the Unix epoch.",
		[]string{},
		prometheus.Labels{},
	)

	statIOQueryBefore18 = `SELECT
		backend_type,
		object,
		context,
		reads,
		read_time,
		writes,
		write_time,
		writebacks,
		writeback_time,
		extends,
		extend_time,
		hits,
		evictions,
		reuses,
		fsyncs,
		fsync_time,
		extract(EPOCH FROM stats_reset) AS reset_time
	FROM
		pg_stat_io;`

	statIOQuery18AndLater = `SELECT
		backend_type,
		object,
		context,
		reads,
		read_bytes,
		read_time,
		writes,
		write_bytes,
		write_time,
		writebacks,
		writeback_time,
		extends,
		extend_bytes,
		extend_time,
		hits,
		evictions,
		reuses,
		fsyncs,
		fsync_time,
		extract(EPOCH FROM stats_reset) AS reset_time
	FROM
		pg_stat_io;`
)

type statIORow struct {
	backendType, object, context                                         sql.NullString
	reads, writes, writebacks, extends                                   sql.NullInt64
	readBytes, writeBytes, extendBytes                                   sql.NullFloat64
	readTime, writeTime, writebackTime, extendTime, fsyncTime, resetTime sql.NullFloat64
	hits, evictions, reuses, fsyncs                                      sql.NullInt64
}

func (c PGStatIOCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if instance.version.LT(semver.MustParse("16.0.0")) {
		if c.log != nil {
			c.log.Debug("stat_io collector is not available on PostgreSQL < 16.0.0, skipping")
		}
		return nil
	}

	pg18 := instance.version.GE(semver.MustParse("18.0.0"))
	query := statIOQueryBefore18
	if pg18 {
		query = statIOQuery18AndLater
	}

	rows, err := instance.getDB().QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// pg_stat_io reports NULL for operations that never happen for a given
	// backend_type/object/context combination. Those are omitted rather than
	// exported as zero.
	var metrics []prometheus.Metric
	var resetTime sql.NullFloat64
	for rows.Next() {
		var row statIORow
		if pg18 {
			err = rows.Scan(&row.backendType, &row.object, &row.context,
				&row.reads, &row.readBytes, &row.readTime,
				&row.writes, &row.writeBytes, &row.writeTime,
				&row.writebacks, &row.writebackTime,
				&row.extends, &row.extendBytes, &row.extendTime,
				&row.hits, &row.evictions, &row.reuses, &row.fsyncs, &row.fsyncTime, &row.resetTime)
		} else {
			err = rows.Scan(&row.backendType, &row.object, &row.context,
				&row.reads, &row.readTime, &row.writes, &row.writeTime,
				&row.writebacks, &row.writebackTime, &row.extends, &row.extendTime,
				&row.hits, &row.evictions, &row.reuses, &row.fsyncs, &row.fsyncTime, &row.resetTime)
		}
		if err != nil {
			return err
		}
		if !resetTime.Valid {
			resetTime = row.resetTime
		}

		labels := []string{row.backendType.String, row.object.String, row.context.String}
		addCounter := func(desc *prometheus.Desc, value sql.NullInt64) {
			if value.Valid {
				metrics = append(metrics, prometheus.MustNewConstMetric(desc, prometheus.CounterValue, int64CounterValue(value, instance.wrapLargeCounters), labels...))
			}
		}
		addFloatCounter := func(desc *prometheus.Desc, value sql.NullFloat64) {
			if value.Valid {
				metrics = append(metrics, prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value.Float64, labels...))
			}
		}
		addMillisecondsCounter := func(desc *prometheus.Desc, value sql.NullFloat64) {
			if value.Valid {
				metrics = append(metrics, prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value.Float64/1000, labels...))
			}
		}

		addCounter(statIOReadsDesc, row.reads)
		addMillisecondsCounter(statIOReadTimeDesc, row.readTime)
		addCounter(statIOWritesDesc, row.writes)
		addMillisecondsCounter(statIOWriteTimeDesc, row.writeTime)
		addCounter(statIOWritebacksDesc, row.writebacks)
		addMillisecondsCounter(statIOWritebackTimeDesc, row.writebackTime)
		addCounter(statIOExtendsDesc, row.extends)
		addMillisecondsCounter(statIOExtendTimeDesc, row.extendTime)
		addCounter(statIOHitsDesc, row.hits)
		addCounter(statIOEvictionsDesc, row.evictions)
		addCounter(statIOReusesDesc, row.reuses)
		addCounter(statIOFsyncsDesc, row.fsyncs)
		addMillisecondsCounter(statIOFsyncTimeDesc, row.fsyncTime)
		if pg18 {
			addFloatCounter(statIOReadBytesDesc, row.readBytes)
			addFloatCounter(statIOWriteBytesDesc, row.writeBytes)
			addFloatCounter(statIOExtendBytesDesc, row.extendBytes)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, metric := range metrics {
		ch <- metric
	}
	// stats_reset is cluster-wide and identical on every row.
	if resetTime.Valid {
		ch <- prometheus.MustNewConstMetric(statIOResetTimeDesc, prometheus.GaugeValue, resetTime.Float64)
	}
	return nil
}
