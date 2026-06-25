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
	"errors"
	"strings"

	"github.com/lib/pq"
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

	var segments uint64
	var size sql.NullInt64
	err := row.Scan(&segments, &size)
	if err != nil {
		// Amazon Aurora PostgreSQL does not support pg_ls_waldir(). Skip
		// emitting WAL metrics on Aurora instead of failing the scrape.
		if isAuroraUnsupportedFunction(err) {
			return nil
		}
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		pgWALSegments,
		prometheus.GaugeValue, float64(segments),
	)
	if size.Valid {
		ch <- prometheus.MustNewConstMetric(
			pgWALSize,
			prometheus.GaugeValue, float64(size.Int64),
		)
	}
	return nil
}

// isAuroraUnsupportedFunction reports whether the error is Aurora
// PostgreSQL rejecting a query because it calls a function unsupported
// on Aurora (here, pg_ls_waldir). Aurora surfaces these as Postgres
// error class "0A" (feature_not_supported) with a message that
// contains "Aurora".
func isAuroraUnsupportedFunction(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code.Class() == "0A" && strings.Contains(pqErr.Message, "Aurora")
	}
	return false
}
