// Copyright 2025 The Prometheus Authors
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

// aurora_stat_bgwriter() exposes Aurora's Optimized Reads cache writes.
// Only the Aurora-specific columns are exported; standard pg_stat_bgwriter
// metrics are covered by the pg_stat_bgwriter collector. Available in
// Aurora PostgreSQL 14.9+ / 15.4+.
const auroraStatBgwriterSubsystem = "aurora_stat_bgwriter"

func init() {
	registerCollector("aurora_stat_bgwriter", defaultDisabled, NewAuroraStatBgwriterCollector)
}

type AuroraStatBgwriterCollector struct{}

func NewAuroraStatBgwriterCollector(collectorConfig) (Collector, error) {
	return &AuroraStatBgwriterCollector{}, nil
}

var (
	auroraBgwriterOrcacheBlksWrittenDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatBgwriterSubsystem, "orcache_blocks_written_total"),
		"Total number of optimized reads cache data blocks written.",
		nil, nil,
	)
	auroraBgwriterOrcacheBlkWriteTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatBgwriterSubsystem, "orcache_block_write_time_seconds_total"),
		"Total time spent writing optimized reads cache data file blocks in seconds (track_io_timing must be enabled).",
		nil, nil,
	)

	auroraStatBgwriterQuery = `SELECT orcache_blks_written, orcache_blk_write_time FROM aurora_stat_bgwriter()`
)

func (c AuroraStatBgwriterCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	row := instance.getDB().QueryRowContext(ctx, auroraStatBgwriterQuery)

	var blksWritten sql.NullInt64
	var blkWriteTimeMs sql.NullFloat64
	if err := row.Scan(&blksWritten, &blkWriteTimeMs); err != nil {
		return err
	}

	if blksWritten.Valid {
		ch <- prometheus.MustNewConstMetric(auroraBgwriterOrcacheBlksWrittenDesc, prometheus.CounterValue, float64(blksWritten.Int64))
	}
	if blkWriteTimeMs.Valid {
		// AWS returns milliseconds; Prometheus convention is seconds.
		ch <- prometheus.MustNewConstMetric(auroraBgwriterOrcacheBlkWriteTimeDesc, prometheus.CounterValue, blkWriteTimeMs.Float64/1000)
	}
	return nil
}
