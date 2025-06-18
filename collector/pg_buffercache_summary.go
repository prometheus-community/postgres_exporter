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

const buffercacheSummarySubsystem = "buffercache_summary"

func init() {
	registerCollector(buffercacheSummarySubsystem, defaultDisabled, NewBuffercacheSummaryCollector)
}

// BuffercacheSummaryCollector collects stats from pg_buffercache: https://www.postgresql.org/docs/current/pgbuffercache.html.
//
// It depends on the extension being loaded with
//
//	create extension pg_buffercache;
//
// It does not take locks, see the PG docs above.
type BuffercacheSummaryCollector struct {
	log *slog.Logger
}

func NewBuffercacheSummaryCollector(config collectorConfig) (Collector, error) {
	return &BuffercacheSummaryCollector{
		log: config.logger,
	}, nil
}

var (
	buffersUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buffercacheSummarySubsystem, "buffers_used"),
		"Number of used shared buffers",
		[]string{},
		prometheus.Labels{},
	)
	buffersUnusedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buffercacheSummarySubsystem, "buffers_unused"),
		"Number of unused shared buffers",
		[]string{},
		prometheus.Labels{},
	)
	buffersDirtyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buffercacheSummarySubsystem, "buffers_dirty"),
		"Number of dirty shared buffers",
		[]string{},
		prometheus.Labels{},
	)
	buffersPinnedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buffercacheSummarySubsystem, "buffers_pinned"),
		"Number of pinned shared buffers",
		[]string{},
		prometheus.Labels{},
	)
	usageCountAvgDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buffercacheSummarySubsystem, "usagecount_avg"),
		"Average usage count of used shared buffers",
		[]string{},
		prometheus.Labels{},
	)

	buffercacheQuery = `
		SELECT
		  buffers_used,
			buffers_unused,
			buffers_dirty,
			buffers_pinned,
			usagecount_avg
		FROM
		  pg_buffercache_summary()
		`
)

func gaugeInt32(ch chan<- prometheus.Metric, desc *prometheus.Desc, m sql.NullInt32) {
	mM := 0.0
	if m.Valid {
		mM = float64(m.Int32)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, mM)
}

// Update implements Collector
// It is called by the Prometheus registry when collecting metrics.
func (c BuffercacheSummaryCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	// pg_buffercache_summary is only in v16, and we don't need support for earlier currently.
	if !instance.version.GE(semver.MustParse("16.0.0")) {
		return nil
	}
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, buffercacheQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var used, unused, dirty, pinned sql.NullInt32
	var usagecountAvg sql.NullFloat64

	for rows.Next() {
		if err := rows.Scan(
			&used,
			&unused,
			&dirty,
			&pinned,
			&usagecountAvg,
		); err != nil {
			return err
		}

		usagecountAvgMetric := 0.0
		if usagecountAvg.Valid {
			usagecountAvgMetric = usagecountAvg.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			usageCountAvgDesc,
			prometheus.GaugeValue,
			usagecountAvgMetric)
		gaugeInt32(used, buffersUsedDesc, ch)
		gaugeInt32(unused, buffersUnusedDesc, ch)
		gaugeInt32(dirty, buffersDirtyDesc, ch)
		gaugeInt32(pinned, buffersPinnedDesc, ch)
	}

	return rows.Err()
}
