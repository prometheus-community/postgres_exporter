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

const auroraStatBgwriterSubsystem = "aurora_stat_bgwriter"

func init() {
	registerCollector(auroraStatBgwriterSubsystem, serverScope, NewAuroraStatBgwriterCollector)
}

// AuroraStatBgwriterCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_bgwriter() function: writes to the Optimized
// Reads cache. aurora_stat_bgwriter() returns all pg_stat_bgwriter columns
// plus Aurora's own; only Aurora's are exported here, the rest come from the
// stat_bgwriter collector. It reports no data on servers that are not
// Aurora, or on Aurora versions without the function.
type AuroraStatBgwriterCollector struct{}

func NewAuroraStatBgwriterCollector(collectorConfig) (Collector, error) {
	return &AuroraStatBgwriterCollector{}, nil
}

var (
	auroraStatBgwriterOrcacheBlocksWritten = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatBgwriterSubsystem, "orcache_blocks_written_total"),
		"Number of data blocks written to the Optimized Reads cache.",
		nil, nil,
	)
	auroraStatBgwriterOrcacheBlockWriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatBgwriterSubsystem, "orcache_block_write_time_seconds_total"),
		"Time spent writing data file blocks to the Optimized Reads cache in seconds, if track_io_timing is enabled.",
		nil, nil,
	)

	auroraStatBgwriterQuery = `SELECT
		orcache_blks_written,
		orcache_blk_write_time
	FROM aurora_stat_bgwriter()`
)

func (c AuroraStatBgwriterCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	db := instance.getDB()

	// aurora_stat_bgwriter() is only in Aurora PostgreSQL 14.9+ and 15.4+.
	exists, err := auroraFunctionExists(ctx, db, "aurora_stat_bgwriter")
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoData
	}

	var blksWritten sql.NullInt64
	var blkWriteTimeMsec sql.NullFloat64
	if err := db.QueryRowContext(ctx, auroraStatBgwriterQuery).Scan(&blksWritten, &blkWriteTimeMsec); err != nil {
		return err
	}

	if blksWritten.Valid {
		ch <- prometheus.MustNewConstMetric(auroraStatBgwriterOrcacheBlocksWritten, prometheus.CounterValue, int64CounterValue(blksWritten, instance.wrapLargeCounters))
	}
	if blkWriteTimeMsec.Valid {
		ch <- prometheus.MustNewConstMetric(auroraStatBgwriterOrcacheBlockWriteTime, prometheus.CounterValue, blkWriteTimeMsec.Float64/1e3)
	}
	return nil
}
