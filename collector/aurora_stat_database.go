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

	"github.com/prometheus/client_golang/prometheus"
)

const auroraStatDatabaseSubsystem = "aurora_stat_database"

func init() {
	registerCollector(auroraStatDatabaseSubsystem, serverScope, NewAuroraStatDatabaseCollector)
}

// AuroraStatDatabaseCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_database() function: where each database's block
// reads come from. aurora_stat_database() returns all pg_stat_database
// columns plus Aurora's own; only Aurora's are exported here, the rest come
// from the stat_database collector. It reports no data on servers that are
// not Aurora, or on Aurora versions without the function.
type AuroraStatDatabaseCollector struct{}

func NewAuroraStatDatabaseCollector(collectorConfig) (Collector, error) {
	return &AuroraStatDatabaseCollector{}, nil
}

var (
	auroraStatDatabaseStorageBlocksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "storage_blocks_read_total"),
		"Number of shared blocks read from Aurora storage.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDatabaseOrcacheBlocksHit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "orcache_blocks_hit_total"),
		"Number of blocks read from the Optimized Reads cache.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDatabaseLocalBlocksRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "local_blocks_read_total"),
		"Number of local blocks read.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDatabaseStorageBlockReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "storage_block_read_time_seconds_total"),
		"Time spent reading data file blocks from Aurora storage in seconds, if track_io_timing is enabled.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDatabaseOrcacheBlockReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "orcache_block_read_time_seconds_total"),
		"Time spent reading data file blocks from the Optimized Reads cache in seconds, if track_io_timing is enabled.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDatabaseLocalBlockReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "local_block_read_time_seconds_total"),
		"Time spent reading local data file blocks in seconds, if track_io_timing is enabled.",
		[]string{"datid", "datname"}, nil,
	)

	auroraStatDatabaseQuery = `SELECT
		datid,
		datname,
		storage_blks_read,
		orcache_blks_hit,
		local_blks_read,
		storage_blk_read_time,
		orcache_blk_read_time,
		local_blk_read_time
	FROM aurora_stat_database()`
)

func (c AuroraStatDatabaseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	db := instance.getDB()

	// aurora_stat_database() is only in Aurora PostgreSQL 14.9+ and 15.4+.
	exists, err := auroraFunctionExists(ctx, db, "aurora_stat_database")
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoData
	}

	rows, err := db.QueryContext(ctx, auroraStatDatabaseQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var datid, datname sql.NullString
		var storageBlksRead, orcacheBlksHit, localBlksRead sql.NullInt64
		var storageBlkReadTimeMsec, orcacheBlkReadTimeMsec, localBlkReadTimeMsec sql.NullFloat64

		if err := rows.Scan(
			&datid,
			&datname,
			&storageBlksRead,
			&orcacheBlksHit,
			&localBlksRead,
			&storageBlkReadTimeMsec,
			&orcacheBlkReadTimeMsec,
			&localBlkReadTimeMsec,
		); err != nil {
			return err
		}

		// Like in pg_stat_database, the row for shared objects has no
		// database name.
		if !datid.Valid || !datname.Valid {
			continue
		}
		labels := []string{datid.String, datname.String}

		if storageBlksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseStorageBlocksRead, prometheus.CounterValue, int64CounterValue(storageBlksRead, instance.wrapLargeCounters), labels...)
		}
		if orcacheBlksHit.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseOrcacheBlocksHit, prometheus.CounterValue, int64CounterValue(orcacheBlksHit, instance.wrapLargeCounters), labels...)
		}
		if localBlksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseLocalBlocksRead, prometheus.CounterValue, int64CounterValue(localBlksRead, instance.wrapLargeCounters), labels...)
		}
		if storageBlkReadTimeMsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseStorageBlockReadTime, prometheus.CounterValue, storageBlkReadTimeMsec.Float64/1e3, labels...)
		}
		if orcacheBlkReadTimeMsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseOrcacheBlockReadTime, prometheus.CounterValue, orcacheBlkReadTimeMsec.Float64/1e3, labels...)
		}
		if localBlkReadTimeMsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDatabaseLocalBlockReadTime, prometheus.CounterValue, localBlkReadTimeMsec.Float64/1e3, labels...)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !found {
		return ErrNoData
	}
	return nil
}
