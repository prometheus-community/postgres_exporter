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

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const statIOSubsystem = "stat_io"

func init() {
	registerCollector(statIOSubsystem, defaultDisabled, NewStatIOCollector)
}

type StatIOCollector struct {
	log *slog.Logger
}

func NewStatIOCollector(config collectorConfig) (Collector, error) {
	return &StatIOCollector{
		log: config.logger,
	}, nil
}

var (
	statIOReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "reads_total"),
		"Number of read operations, each of the size specified in op_bytes.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOReadTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "read_time_total"),
		"Time spent in read operations in milliseconds (if track_io_timing is enabled, otherwise zero)",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOWritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "writes_total"),
		"Number of write operations, each of the size specified in op_bytes.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOWriteTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "writes_time_total"),
		"Time spent in write operations in milliseconds (if track_io_timing is enabled, otherwise zero)",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOWriteBackTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "write_back_total"),
		"Number of units of size op_bytes which the process requested the kernel write out to permanent storage.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOWriteBackTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "write_back_time_total"),
		"Time spent in writeback operations in milliseconds (if track_io_timing is enabled, otherwise zero). This includes the time spent queueing write-out requests and, potentially, the time spent to write out the dirty data.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOExtendsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "extends_total"),
		"Number of relation extend operations, each of the size specified in op_bytes.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOExtendsTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "extends_time_total"),
		"Time spent in extend operations in milliseconds (if track_io_timing is enabled, otherwise zero)",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "hits_total"),
		"The number of times a desired block was found in a shared buffer.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOEvictionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "evictions_total"),
		"Number of times a block has been written out from a shared or local buffer in order to make it available for another use.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOReusesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "reuses_total"),
		"The number of times an existing buffer in a size-limited ring buffer outside of shared buffers was reused as part of an I/O operation in the bulkread, bulkwrite, or vacuum contexts.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)

	statIOFsyncsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "fsync_total"),
		"Number of fsync calls. These are only tracked in context normal.",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOFsyncTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statIOSubsystem, "fsync_time_total"),
		"Time spent in fsync operations in milliseconds (if track_io_timing is enabled, otherwise zero)",
		[]string{"backend_type", "object", "context"},
		prometheus.Labels{},
	)
	statIOQuery = `
		SELECT
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
		fsync_time

		FROM
		pg_stat_io
		`
)

// Update implements Collector and exposes database locks.
// It is called by the Prometheus registry when collecting metrics.
func (c StatIOCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	// pg_stat_io is only in v16, and we don't need support for earlier currently.
	if !instance.version.GE(semver.MustParse("16.0.0")) {
		return nil
	}
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx, statIOQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var backendType, object, PGContext sql.NullString
	var reads, writes, writeBacks, extends, hits, evictions, reuses, fsyncs sql.NullInt64
	var readTime, writeTime, writeBackTime, extendsTime, fsyncTime sql.NullFloat64

	for rows.Next() {
		if err := rows.Scan(
			&backendType, &object, &PGContext,
			&reads,
			&readTime,
			&writes,
			&writeTime,
			&writeBacks,
			&writeBackTime,
			&extends,
			&extendsTime,
			&hits,
			&evictions,
			&reuses,
			&fsyncs,
			&fsyncTime); err != nil {
			return err
		}

		if !backendType.Valid || !object.Valid || !PGContext.Valid {
			continue
		}

		readsMetric := 0.0
		if reads.Valid {
			readsMetric = float64(reads.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOReadsTotal,
			prometheus.CounterValue,
			readsMetric,
			backendType.String, object.String, PGContext.String)

		readTimeMetric := 0.0
		if readTime.Valid {
			readTimeMetric = readTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statIOReadTimeTotal,
			prometheus.CounterValue,
			readTimeMetric,
			backendType.String, object.String, PGContext.String)

		writesMetric := 0.0
		if writes.Valid {
			writesMetric = float64(writes.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOWritesTotal,
			prometheus.CounterValue,
			writesMetric,
			backendType.String, object.String, PGContext.String)

		writeTimeMetric := 0.0
		if writeTime.Valid {
			writeTimeMetric = writeTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statIOWriteTimeTotal,
			prometheus.CounterValue,
			writeTimeMetric,
			backendType.String, object.String, PGContext.String)

		writeBackMetric := 0.0
		if writeBacks.Valid {
			writeBackMetric = float64(writeBacks.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOWriteBackTotal,
			prometheus.CounterValue,
			writeBackMetric,
			backendType.String, object.String, PGContext.String)

		writeBackTimeMetric := 0.0
		if writeBackTime.Valid {
			writeBackTimeMetric = writeBackTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statIOWriteBackTimeTotal,
			prometheus.CounterValue,
			writeBackTimeMetric,
			backendType.String, object.String, PGContext.String)

		extendsMetric := 0.0
		if extends.Valid {
			extendsMetric = float64(extends.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOExtendsTotal,
			prometheus.CounterValue,
			extendsMetric,
			backendType.String, object.String, PGContext.String)

		extendsTimeMetric := 0.0
		if extendsTime.Valid {
			extendsTimeMetric = extendsTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statIOExtendsTimeTotal,
			prometheus.CounterValue,
			extendsTimeMetric,
			backendType.String, object.String, PGContext.String)

		hitsMetric := 0.0
		if hits.Valid {
			hitsMetric = float64(hits.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOHitsTotal,
			prometheus.CounterValue,
			hitsMetric,
			backendType.String, object.String, PGContext.String)

		evictionsMetric := 0.0
		if evictions.Valid {
			evictionsMetric = float64(evictions.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOEvictionsTotal,
			prometheus.CounterValue,
			evictionsMetric,
			backendType.String, object.String, PGContext.String)

		reusesMetric := 0.0
		if reuses.Valid {
			reusesMetric = float64(reuses.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOReusesTotal,
			prometheus.CounterValue,
			reusesMetric,
			backendType.String, object.String, PGContext.String)

		fsyncsMetric := 0.0
		if fsyncs.Valid {
			fsyncsMetric = float64(fsyncs.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statIOFsyncsTotal,
			prometheus.CounterValue,
			fsyncsMetric,
			backendType.String, object.String, PGContext.String)

		fsyncTimeMetric := 0.0
		if fsyncTime.Valid {
			fsyncTimeMetric = fsyncTime.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statIOFsyncTimeTotal,
			prometheus.CounterValue,
			fsyncTimeMetric,
			backendType.String, object.String, PGContext.String)

	}

	return rows.Err()
}
