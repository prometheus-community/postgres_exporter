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

	pgPscaleUtilsBuildTimestampFunc = "pg_pscale_utils_build_unix_timestamp"
	pgReadonlyBuildTimestampFunc    = "pg_readonly_build_unix_timestamp"
	pginsightsBuildTimestampFunc    = "pginsights_build_unix_timestamp"
)

func (c *PostgresBinariesCollector) Update(ctx context.Context, instance *Instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	hasData := false

	// pg_pscale_utils build timestamp
	if ts, exists, err := queryBuildTimestampIfExists(ctx, db, pgPscaleUtilsBuildTimestampFunc); err != nil {
		return err
	} else if exists {
		ch <- prometheus.MustNewConstMetric(
			pgPscaleUtilsBuildUnixTimestamp,
			prometheus.GaugeValue, ts,
		)
		hasData = true
	}

	// pg_readonly build timestamp
	if ts, exists, err := queryBuildTimestampIfExists(ctx, db, pgReadonlyBuildTimestampFunc); err != nil {
		return err
	} else if exists {
		ch <- prometheus.MustNewConstMetric(
			pgReadonlyBuildUnixTimestamp,
			prometheus.GaugeValue, ts,
		)
		hasData = true
	}

	// pginsights build timestamp
	if ts, exists, err := queryBuildTimestampIfExists(ctx, db, pginsightsBuildTimestampFunc); err != nil {
		return err
	} else if exists {
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

// queryBuildTimestampIfExists checks if a function exists in pg_proc before calling it.
// This avoids logging errors in Postgres when the function doesn't exist.
// Returns (timestamp, exists, error). If the function doesn't exist, returns (0, false, nil).
// Real errors (connection issues, etc.) are returned as errors.
func queryBuildTimestampIfExists(ctx context.Context, db *sql.DB, funcName string) (float64, bool, error) {
	// Check if the function exists before calling it
	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = $1)", funcName).Scan(&exists)
	if err != nil {
		return 0, false, err
	}
	if !exists {
		return 0, false, nil
	}

	// Function exists, call it
	var ts sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT "+funcName+"()").Scan(&ts)
	if err != nil {
		return 0, false, err
	}
	if !ts.Valid {
		return 0, false, nil
	}
	return float64(ts.Int64), true, nil
}
