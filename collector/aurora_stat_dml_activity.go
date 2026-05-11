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

// aurora_stat_dml_activity(oid) returns cumulative per-operation
// counts and latency (in microseconds) for SELECT/INSERT/UPDATE/DELETE.
// Available on Aurora PostgreSQL 11.6+.
const auroraStatDMLActivitySubsystem = "aurora_stat_dml_activity"

func init() {
	registerCollector("aurora_stat_dml_activity", defaultDisabled, NewAuroraStatDMLActivityCollector)
}

type AuroraStatDMLActivityCollector struct {
	excludeDatabases []string
}

func NewAuroraStatDMLActivityCollector(config collectorConfig) (Collector, error) {
	return &AuroraStatDMLActivityCollector{excludeDatabases: config.excludeDatabases}, nil
}

var (
	auroraDMLCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDMLActivitySubsystem, "operations_total"),
		"Cumulative count of DML operations per database and operation type (aurora_stat_dml_activity).",
		[]string{"datid", "datname", "operation"}, nil,
	)
	auroraDMLLatencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDMLActivitySubsystem, "latency_microseconds_total"),
		"Cumulative DML latency in microseconds per database and operation type (aurora_stat_dml_activity).",
		[]string{"datid", "datname", "operation"}, nil,
	)

	auroraStatDMLActivityQuery = `SELECT
		pg_database.oid::text AS datid,
		pg_database.datname,
		(aurora_stat_dml_activity(pg_database.oid)).*
	FROM pg_database
	WHERE datallowconn`
)

func (c AuroraStatDMLActivityCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatDMLActivityQuery)
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
		var selectCount, selectLat, insertCount, insertLat sql.NullInt64
		var updateCount, updateLat, deleteCount, deleteLat sql.NullInt64
		if err := rows.Scan(
			&datid, &datname,
			&selectCount, &selectLat,
			&insertCount, &insertLat,
			&updateCount, &updateLat,
			&deleteCount, &deleteLat,
		); err != nil {
			return err
		}
		if !datname.Valid {
			continue
		}
		if _, skip := excluded[datname.String]; skip {
			continue
		}

		emit := func(op string, count, latency sql.NullInt64) {
			labels := []string{datid.String, datname.String, op}
			if count.Valid {
				ch <- prometheus.MustNewConstMetric(auroraDMLCountDesc, prometheus.CounterValue, float64(count.Int64), labels...)
			}
			if latency.Valid {
				ch <- prometheus.MustNewConstMetric(auroraDMLLatencyDesc, prometheus.CounterValue, float64(latency.Int64), labels...)
			}
		}
		emit("select", selectCount, selectLat)
		emit("insert", insertCount, insertLat)
		emit("update", updateCount, updateLat)
		emit("delete", deleteCount, deleteLat)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if !found {
		return ErrNoData
	}
	return nil
}
