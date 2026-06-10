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

// aurora_stat_database() returns all pg_stat_database columns plus
// Aurora-specific ones for storage/local/orcache block reads. Only the
// Aurora-specific columns are exported here — standard pg_stat_database
// metrics are provided by the stat_database collector. Available in
// Aurora PostgreSQL 14.9+ / 15.4+.
const auroraStatDatabaseSubsystem = "aurora_stat_database"

func init() {
	registerCollector("aurora_stat_database", defaultDisabled, NewAuroraStatDatabaseCollector)
}

type AuroraStatDatabaseCollector struct {
	excludeDatabases []string
}

func NewAuroraStatDatabaseCollector(config collectorConfig) (Collector, error) {
	return &AuroraStatDatabaseCollector{excludeDatabases: config.excludeDatabases}, nil
}

var (
	auroraStatDBStorageBlksReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "storage_blocks_read_total"),
		"Total number of shared blocks read from Aurora storage in this database.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDBOrcacheBlksHitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "orcache_blocks_hit_total"),
		"Total number of optimized reads cache hits in this database.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDBLocalBlksReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "local_blocks_read_total"),
		"Total number of local blocks read in this database.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDBStorageBlkReadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "storage_block_read_time_seconds_total"),
		"Total time spent reading data file blocks from Aurora storage in seconds.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDBOrcacheBlkReadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "orcache_block_read_time_seconds_total"),
		"Total time spent reading data file blocks from optimized reads cache in seconds.",
		[]string{"datid", "datname"}, nil,
	)
	auroraStatDBLocalBlkReadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDatabaseSubsystem, "local_block_read_time_seconds_total"),
		"Total time spent reading local data file blocks in seconds.",
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
	FROM aurora_stat_database()
	WHERE datname IS NOT NULL`
)

func (c AuroraStatDatabaseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatDatabaseQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	excluded := make(map[string]struct{}, len(c.excludeDatabases))
	for _, d := range c.excludeDatabases {
		excluded[d] = struct{}{}
	}

	var found bool
	for rows.Next() {
		found = true

		var datid sql.NullString
		var datname sql.NullString
		var storageBlksRead, orcacheBlksHit, localBlksRead sql.NullInt64
		var storageBlkReadTimeMs, orcacheBlkReadTimeMs, localBlkReadTimeMs sql.NullFloat64

		if err := rows.Scan(
			&datid,
			&datname,
			&storageBlksRead,
			&orcacheBlksHit,
			&localBlksRead,
			&storageBlkReadTimeMs,
			&orcacheBlkReadTimeMs,
			&localBlkReadTimeMs,
		); err != nil {
			return err
		}

		if !datname.Valid {
			continue
		}
		if _, skip := excluded[datname.String]; skip {
			continue
		}

		labels := []string{datid.String, datname.String}

		if storageBlksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBStorageBlksReadDesc, prometheus.CounterValue, float64(storageBlksRead.Int64), labels...)
		}
		if orcacheBlksHit.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBOrcacheBlksHitDesc, prometheus.CounterValue, float64(orcacheBlksHit.Int64), labels...)
		}
		if localBlksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBLocalBlksReadDesc, prometheus.CounterValue, float64(localBlksRead.Int64), labels...)
		}
		if storageBlkReadTimeMs.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBStorageBlkReadTimeDesc, prometheus.CounterValue, storageBlkReadTimeMs.Float64/1000, labels...)
		}
		if orcacheBlkReadTimeMs.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBOrcacheBlkReadTimeDesc, prometheus.CounterValue, orcacheBlkReadTimeMs.Float64/1000, labels...)
		}
		if localBlkReadTimeMs.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatDBLocalBlkReadTimeDesc, prometheus.CounterValue, localBlkReadTimeMs.Float64/1000, labels...)
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
