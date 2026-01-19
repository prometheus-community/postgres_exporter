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

const postgresBinariesSubsystem = "postgres_binaries"

func init() {
	registerCollector(postgresBinariesSubsystem, defaultEnabled, NewPostgresBinariesCollector)
}

type PostgresBinariesCollector struct {
}

func NewPostgresBinariesCollector(collectorConfig) (Collector, error) {
	return &PostgresBinariesCollector{}, nil
}

var (
	pgPscaleUtilsBuildUnixTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			postgresBinariesSubsystem,
			"pg_pscale_utils_build_unix_timestamp",
		),
		"Unix timestamp when pg_pscale_utils was built",
		[]string{}, nil,
	)

	pgReadonlyBuildUnixTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			postgresBinariesSubsystem,
			"pg_readonly_build_unix_timestamp",
		),
		"Unix timestamp when pg_readonly was built",
		[]string{}, nil,
	)

	pginsightsBuildUnixTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			postgresBinariesSubsystem,
			"pginsights_build_unix_timestamp",
		),
		"Unix timestamp when pginsights was built",
		[]string{}, nil,
	)

	pgPscaleUtilsBuildTimestampQuery = "SELECT pg_pscale_utils_build_unix_timestamp();"
	pgReadonlyBuildTimestampQuery    = "SELECT pg_readonly_build_unix_timestamp();"
	pginsightsBuildTimestampQuery    = "SELECT pginsights_build_unix_timestamp();"
)

func (c *PostgresBinariesCollector) Update(ctx context.Context, instance *Instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	hasData := false

	// pg_pscale_utils build timestamp
	if ts, err := queryBuildTimestamp(ctx, db, pgPscaleUtilsBuildTimestampQuery); err == nil {
		ch <- prometheus.MustNewConstMetric(
			pgPscaleUtilsBuildUnixTimestamp,
			prometheus.GaugeValue, ts,
		)
		hasData = true
	}

	// pg_readonly build timestamp
	if ts, err := queryBuildTimestamp(ctx, db, pgReadonlyBuildTimestampQuery); err == nil {
		ch <- prometheus.MustNewConstMetric(
			pgReadonlyBuildUnixTimestamp,
			prometheus.GaugeValue, ts,
		)
		hasData = true
	}

	// pginsights build timestamp
	if ts, err := queryBuildTimestamp(ctx, db, pginsightsBuildTimestampQuery); err == nil {
		ch <- prometheus.MustNewConstMetric(
			pginsightsBuildUnixTimestamp,
			prometheus.GaugeValue, ts,
		)
		hasData = true
	}

	if !hasData {
		return ErrNoData
	}
	return nil
}

func queryBuildTimestamp(ctx context.Context, db *sql.DB, query string) (float64, error) {
	var ts sql.NullInt64
	err := db.QueryRowContext(ctx, query).Scan(&ts)
	if err != nil {
		return 0, err
	}
	if !ts.Valid {
		return 0, nil
	}
	return float64(ts.Int64), nil
}
