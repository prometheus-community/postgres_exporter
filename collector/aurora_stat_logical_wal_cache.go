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

const auroraStatLogicalWalCacheSubsystem = "aurora_stat_logical_wal_cache"

func init() {
	registerCollector(auroraStatLogicalWalCacheSubsystem, serverScope, NewAuroraStatLogicalWalCacheCollector)
}

// AuroraStatLogicalWalCacheCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_logical_wal_cache() function: WAL cache usage of
// each logical replication slot. It reports no data on servers that are not
// Aurora, or on Aurora versions without the function.
type AuroraStatLogicalWalCacheCollector struct{}

func NewAuroraStatLogicalWalCacheCollector(collectorConfig) (Collector, error) {
	return &AuroraStatLogicalWalCacheCollector{}, nil
}

var (
	auroraStatLogicalWalCacheHits = prometheus.V2.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "cache_hits_total"),
		"Number of WAL cache hits for the replication slot.",
		prometheus.UnconstrainedLabels([]string{"slot_name"}), nil,
		prometheus.WithUnit("{hit}"),
	)
	auroraStatLogicalWalCacheMisses = prometheus.V2.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "cache_misses_total"),
		"Number of WAL cache misses for the replication slot.",
		prometheus.UnconstrainedLabels([]string{"slot_name"}), nil,
		prometheus.WithUnit("{miss}"),
	)
	auroraStatLogicalWalCacheBlocksRead = prometheus.V2.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatLogicalWalCacheSubsystem, "blocks_read_total"),
		"Number of WAL cache read requests for the replication slot.",
		prometheus.UnconstrainedLabels([]string{"slot_name"}), nil,
		prometheus.WithUnit("{block}"),
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
	db := instance.getDB()

	// aurora_stat_logical_wal_cache() is only in Aurora PostgreSQL 11.17+,
	// 12.12+, 13.8+, 14.7+ and 15.2+.
	exists, err := functionExists(ctx, db, "aurora_stat_logical_wal_cache")
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoData
	}

	rows, err := db.QueryContext(ctx, auroraStatLogicalWalCacheQuery)
	// Readers (Aurora Replicas) raise object_not_in_prerequisite_state.
	if isObjectNotInPrerequisiteState(err) {
		return ErrNoData
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var slotName string
		var cacheHit, cacheMiss, blksRead sql.NullInt64

		if err := rows.Scan(&slotName, &cacheHit, &cacheMiss, &blksRead); err != nil {
			return err
		}

		if cacheHit.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatLogicalWalCacheHits, prometheus.CounterValue, int64CounterValue(cacheHit, instance.wrapLargeCounters), slotName)
		}
		if cacheMiss.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatLogicalWalCacheMisses, prometheus.CounterValue, int64CounterValue(cacheMiss, instance.wrapLargeCounters), slotName)
		}
		if blksRead.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatLogicalWalCacheBlocksRead, prometheus.CounterValue, int64CounterValue(blksRead, instance.wrapLargeCounters), slotName)
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
