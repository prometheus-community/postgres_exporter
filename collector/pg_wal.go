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
	"github.com/prometheus/client_golang/prometheus"
)

const walSubsystem = "wal"

func init() {
	registerCollector(walSubsystem, defaultEnabled, NewPGWALCollector)
}

type PGWALCollector struct {
}

func NewPGWALCollector(config collectorConfig) (Collector, error) {
	return &PGWALCollector{}, nil
}

var (
	pgWALSegments = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			walSubsystem,
			"segments",
		),
		"Number of WAL segments",
		[]string{}, nil,
	)
	pgWALSize = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			walSubsystem,
			"size_bytes",
		),
		"Total size of WAL segments",
		[]string{}, nil,
	)

	pgWALQuery = `
		SELECT
			COUNT(*) AS segments,
			SUM(size) AS size
		FROM pg_ls_waldir()
		WHERE name ~ '^[0-9A-F]{24}$'`
)

func (c PGWALCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx,
		pgWALQuery,
	)

	var segments sql.NullInt64
	var size sql.NullInt64
	err := row.Scan(&segments, &size)
	if err != nil {
		return err
	}

	var segmentsValue float64
    if segments.Valid {
        segmentsValue = float64(segments.Int64)
    } else {
        segmentsValue = 0
    }

    var sizeValue float64
    if size.Valid {
        sizeValue = float64(size.Int64)
    } else {
        sizeValue = 0
    }

	ch <- prometheus.MustNewConstMetric(
		pgWALSegments,
		prometheus.GaugeValue, segmentsValue,
	)
	ch <- prometheus.MustNewConstMetric(
		pgWALSize,
		prometheus.GaugeValue, sizeValue,
	)
	return nil
}
