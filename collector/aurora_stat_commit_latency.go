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

// aurora_stat_get_db_commit_latency(oid) returns the cumulative commit
// latency in microseconds for an Aurora PostgreSQL database. We join with
// pg_database in a single query so each scrape executes one round-trip.
const auroraStatCommitLatencySubsystem = "aurora_stat_commit_latency"

func init() {
	registerCollector("aurora_stat_commit_latency", defaultDisabled, NewAuroraStatCommitLatencyCollector)
}

type AuroraStatCommitLatencyCollector struct {
	excludeDatabases []string
}

func NewAuroraStatCommitLatencyCollector(config collectorConfig) (Collector, error) {
	return &AuroraStatCommitLatencyCollector{excludeDatabases: config.excludeDatabases}, nil
}

var (
	auroraStatCommitLatencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatCommitLatencySubsystem, "microseconds_total"),
		"Cumulative commit latency in microseconds per database (aurora_stat_get_db_commit_latency).",
		[]string{"datid", "datname"}, nil,
	)

	auroraStatCommitLatencyQuery = `SELECT
		oid::text AS datid,
		datname,
		aurora_stat_get_db_commit_latency(oid) AS latency
	FROM pg_database
	WHERE datallowconn`
)

func (c AuroraStatCommitLatencyCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatCommitLatencyQuery)
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

		var datid, datname sql.NullString
		var latency sql.NullInt64
		if err := rows.Scan(&datid, &datname, &latency); err != nil {
			return err
		}
		if !datname.Valid {
			continue
		}
		if _, skip := excluded[datname.String]; skip {
			continue
		}
		if latency.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatCommitLatencyDesc, prometheus.CounterValue, float64(latency.Int64), datid.String, datname.String)
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
