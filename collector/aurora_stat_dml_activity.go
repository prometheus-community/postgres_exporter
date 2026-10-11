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

const auroraStatDMLActivitySubsystem = "aurora_stat_dml_activity"

func init() {
	registerCollector(auroraStatDMLActivitySubsystem, serverScope, NewAuroraStatDMLActivityCollector)
}

// AuroraStatDMLActivityCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_dml_activity() function: the number and total
// time of successful SELECT, INSERT, UPDATE and DELETE operations per
// database. It reports no data on servers that are not Aurora.
type AuroraStatDMLActivityCollector struct{}

func NewAuroraStatDMLActivityCollector(collectorConfig) (Collector, error) {
	return &AuroraStatDMLActivityCollector{}, nil
}

var (
	auroraStatDMLActivityOperations = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDMLActivitySubsystem, "operations_total"),
		"Number of successful DML operations, by operation type.",
		[]string{"datid", "datname", "operation"}, nil,
	)
	auroraStatDMLActivityLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatDMLActivitySubsystem, "latency_seconds_total"),
		"Total time spent in successful DML operations in seconds, by operation type.",
		[]string{"datid", "datname", "operation"}, nil,
	)

	// LATERAL calls the function once per database; (f(oid)).* in the
	// select list would call it once per column.
	auroraStatDMLActivityQuery = `SELECT
		d.oid::text AS datid,
		d.datname,
		a.select_count,
		a.select_latency_microsecs,
		a.insert_count,
		a.insert_latency_microsecs,
		a.update_count,
		a.update_latency_microsecs,
		a.delete_count,
		a.delete_latency_microsecs
	FROM pg_database d
	CROSS JOIN LATERAL aurora_stat_dml_activity(d.oid) a
	WHERE d.datallowconn`
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

	var found bool
	for rows.Next() {
		found = true

		var datid, datname string
		var selectCount, selectLatencyUsec, insertCount, insertLatencyUsec sql.NullInt64
		var updateCount, updateLatencyUsec, deleteCount, deleteLatencyUsec sql.NullInt64

		if err := rows.Scan(
			&datid,
			&datname,
			&selectCount,
			&selectLatencyUsec,
			&insertCount,
			&insertLatencyUsec,
			&updateCount,
			&updateLatencyUsec,
			&deleteCount,
			&deleteLatencyUsec,
		); err != nil {
			return err
		}

		// Databases without DML activity, such as template1, report NULLs.
		for _, op := range []struct {
			name           string
			count, latency sql.NullInt64
		}{
			{"select", selectCount, selectLatencyUsec},
			{"insert", insertCount, insertLatencyUsec},
			{"update", updateCount, updateLatencyUsec},
			{"delete", deleteCount, deleteLatencyUsec},
		} {
			if op.count.Valid {
				ch <- prometheus.MustNewConstMetric(auroraStatDMLActivityOperations, prometheus.CounterValue, int64CounterValue(op.count, instance.wrapLargeCounters), datid, datname, op.name)
			}
			if op.latency.Valid {
				ch <- prometheus.MustNewConstMetric(auroraStatDMLActivityLatency, prometheus.CounterValue, float64(op.latency.Int64)/1e6, datid, datname, op.name)
			}
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
