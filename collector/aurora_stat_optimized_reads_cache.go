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

// aurora_stat_optimized_reads_cache() reports Aurora Optimized Reads (tiered
// SSD-backed) cache sizing. Available in Aurora PostgreSQL 14.9+ / 15.4+.
const auroraStatOptimizedReadsCacheSubsystem = "aurora_stat_optimized_reads_cache"

func init() {
	registerCollector("aurora_stat_optimized_reads_cache", defaultDisabled, NewAuroraStatOptimizedReadsCacheCollector)
}

type AuroraStatOptimizedReadsCacheCollector struct{}

func NewAuroraStatOptimizedReadsCacheCollector(collectorConfig) (Collector, error) {
	return &AuroraStatOptimizedReadsCacheCollector{}, nil
}

var (
	auroraOrcTotalSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatOptimizedReadsCacheSubsystem, "total_size_bytes"),
		"Total optimized reads cache size in bytes.",
		nil, nil,
	)
	auroraOrcUsedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatOptimizedReadsCacheSubsystem, "used_size_bytes"),
		"Used page size in optimized reads cache in bytes.",
		nil, nil,
	)

	auroraStatOptimizedReadsCacheQuery = `SELECT total_size, used_size FROM aurora_stat_optimized_reads_cache()`
)

func (c AuroraStatOptimizedReadsCacheCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	row := instance.getDB().QueryRowContext(ctx, auroraStatOptimizedReadsCacheQuery)

	var totalSize, usedSize sql.NullInt64
	if err := row.Scan(&totalSize, &usedSize); err != nil {
		return err
	}

	if totalSize.Valid {
		ch <- prometheus.MustNewConstMetric(auroraOrcTotalSizeDesc, prometheus.GaugeValue, float64(totalSize.Int64))
	}
	if usedSize.Valid {
		ch <- prometheus.MustNewConstMetric(auroraOrcUsedSizeDesc, prometheus.GaugeValue, float64(usedSize.Int64))
	}
	return nil
}
