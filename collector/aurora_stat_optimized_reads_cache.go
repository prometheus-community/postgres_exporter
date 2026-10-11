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
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

const auroraStatOptimizedReadsCacheSubsystem = "aurora_stat_optimized_reads_cache"

func init() {
	registerCollector(auroraStatOptimizedReadsCacheSubsystem, serverScope, NewAuroraStatOptimizedReadsCacheCollector)
}

// AuroraStatOptimizedReadsCacheCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_optimized_reads_cache() function: the size and
// usage of the Optimized Reads cache on local NVMe storage. It reports no
// data on servers that are not Aurora, or on Aurora versions without the
// function.
type AuroraStatOptimizedReadsCacheCollector struct{}

func NewAuroraStatOptimizedReadsCacheCollector(collectorConfig) (Collector, error) {
	return &AuroraStatOptimizedReadsCacheCollector{}, nil
}

var (
	auroraStatOptimizedReadsCacheTotalSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatOptimizedReadsCacheSubsystem, "total_size_bytes"),
		"Total size of the Optimized Reads cache in bytes.",
		nil, nil,
	)
	auroraStatOptimizedReadsCacheUsedSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatOptimizedReadsCacheSubsystem, "used_size_bytes"),
		"Used size of the Optimized Reads cache in bytes.",
		nil, nil,
	)

	auroraStatOptimizedReadsCacheQuery = `SELECT
		total_size,
		used_size
	FROM aurora_stat_optimized_reads_cache()`
)

func (c AuroraStatOptimizedReadsCacheCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	db := instance.getDB()

	// aurora_stat_optimized_reads_cache() is only in Aurora PostgreSQL 14.9+
	// and 15.4+.
	exists, err := auroraFunctionExists(ctx, db, "aurora_stat_optimized_reads_cache")
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoData
	}

	var totalSize, usedSize sql.NullInt64
	err = db.QueryRowContext(ctx, auroraStatOptimizedReadsCacheQuery).Scan(&totalSize, &usedSize)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoData
	}
	if err != nil {
		return err
	}

	if totalSize.Valid {
		ch <- prometheus.MustNewConstMetric(auroraStatOptimizedReadsCacheTotalSize, prometheus.GaugeValue, float64(totalSize.Int64))
	}
	if usedSize.Valid {
		ch <- prometheus.MustNewConstMetric(auroraStatOptimizedReadsCacheUsedSize, prometheus.GaugeValue, float64(usedSize.Int64))
	}
	return nil
}
