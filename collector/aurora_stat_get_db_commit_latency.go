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

const auroraStatGetDBCommitLatencySubsystem = "aurora_stat_get_db_commit_latency"

func init() {
	registerCollector(auroraStatGetDBCommitLatencySubsystem, serverScope, NewAuroraStatGetDBCommitLatencyCollector)
}

// AuroraStatGetDBCommitLatencyCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_get_db_commit_latency() function: the total time
// spent committing transactions per database. It reports no data on servers
// that are not Aurora.
type AuroraStatGetDBCommitLatencyCollector struct{}

func NewAuroraStatGetDBCommitLatencyCollector(collectorConfig) (Collector, error) {
	return &AuroraStatGetDBCommitLatencyCollector{}, nil
}

var (
	auroraStatGetDBCommitLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatGetDBCommitLatencySubsystem, "seconds_total"),
		"Total time spent committing transactions in seconds, from the commit request until the client receives the acknowledgement.",
		[]string{"datid", "datname"}, nil,
	)

	auroraStatGetDBCommitLatencyQuery = `SELECT
		oid::text AS datid,
		datname,
		aurora_stat_get_db_commit_latency(oid)
	FROM pg_database
	WHERE datallowconn`
)

func (c AuroraStatGetDBCommitLatencyCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatGetDBCommitLatencyQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var datid, datname string
		var latencyUsec sql.NullInt64

		if err := rows.Scan(&datid, &datname, &latencyUsec); err != nil {
			return err
		}

		if latencyUsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatGetDBCommitLatency, prometheus.CounterValue, float64(latencyUsec.Int64)/1e6, datid, datname)
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
