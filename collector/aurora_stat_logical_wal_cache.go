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

// aurora_stat_logical_wal_cache() exposes per-replication-slot WAL cache usage.
// Available in Aurora PostgreSQL 11.17+, 12.12+, 13.8+, 14.7+, 15.2+.
const auroraStatLogicalWalCacheSubsystem = "aurora_stat_logical_wal_cache"

func init() {
	registerCollector("aurora_stat_logical_wal_cache", defaultDisabled, NewAuroraStatLogicalWalCacheCollector)
}

type AuroraStatLogicalWalCacheCollector struct{}

func NewAuroraStatLogicalWalCacheCollector(collectorConfig) (Collector, error) {
	return &AuroraStatLogicalWalCacheCollector{}, nil
}

var (
	auroraLogicalWalCacheHitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "cache_hits_total"),
		"Total number of WAL cache hits since last reset, per replication slot.",
		[]string{"slot_name"}, nil,
	)
	auroraLogicalWalCacheMissDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "cache_misses_total"),
		"Total number of WAL cache misses since last reset, per replication slot.",
		[]string{"slot_name"}, nil,
	)
	auroraLogicalWalCacheBlksReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "blocks_read_total"),
		"Total number of WAL cache read requests, per replication slot.",
		[]string{"slot_name"}, nil,
	)

	auroraStatLogicalWalCacheQuery = `SELECT
		name,
		cache_hit,
		cache_miss,
		blks_read
	FROM aurora_stat_logical_wal_cache()`
)

func (c AuroraStatLogicalWalCacheCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatLogicalWalCacheQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var name string
		var cacheHit, cacheMiss, blksRead sql.NullInt64
		if err := rows.Scan(&name, &cacheHit, &cacheMiss, &blksRead); err != nil {
			return err
		}

		if cacheHit.Valid {
			ch <- prometheus.MustNewConstMetric(auroraLogicalWalCacheHitDesc, prometheus.CounterValue, float64(cacheHit.Int64), name)
		}
		if cacheMiss.Valid {
			ch <- prometheus.MustNewConstMetric(auroraLogicalWalCacheMissDesc, prometheus.CounterValue, float64(cacheMiss.Int64), name)
		}
		if blksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraLogicalWalCacheBlksReadDesc, prometheus.CounterValue, float64(blksRead.Int64), name)
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
