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

// auroraReplicaStatusCollector exposes metrics from Aurora PostgreSQL's
// aurora_replica_status() function. Disabled by default because the function
// only exists on Aurora.
const auroraReplicaSubsystem = "aurora_replica"

func init() {
	registerCollector("aurora_replica_status", defaultDisabled, NewAuroraReplicaStatusCollector)
}

type AuroraReplicaStatusCollector struct{}

func NewAuroraReplicaStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraReplicaStatusCollector{}, nil
}

var (
	auroraReplicaLagDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaSubsystem, "lag_milliseconds"),
		"Replica lag behind writer in milliseconds (aurora_replica_status.replica_lag_in_msec).",
		[]string{"server_id"}, nil,
	)
	auroraReplayLatencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaSubsystem, "replay_latency_microseconds"),
		"Expected log replay latency in microseconds (aurora_replica_status.cur_replay_latency_in_usec).",
		[]string{"server_id"}, nil,
	)
	auroraPendingReadIOsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaSubsystem, "pending_read_ios"),
		"Outstanding page reads pending on instance.",
		[]string{"server_id"}, nil,
	)

	auroraReplicaStatusQuery = `SELECT
		server_id,
		replica_lag_in_msec,
		cur_replay_latency_in_usec,
		pending_read_ios
	FROM aurora_replica_status()`
)

func (c AuroraReplicaStatusCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	rows, err := instance.getDB().QueryContext(ctx, auroraReplicaStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var serverID string
		var replicaLag, replayLatency sql.NullFloat64
		var pendingReadIOs sql.NullInt64

		if err := rows.Scan(&serverID, &replicaLag, &replayLatency, &pendingReadIOs); err != nil {
			return err
		}

		if replicaLag.Valid {
			ch <- prometheus.MustNewConstMetric(auroraReplicaLagDesc, prometheus.GaugeValue, replicaLag.Float64, serverID)
		}
		if replayLatency.Valid {
			ch <- prometheus.MustNewConstMetric(auroraReplayLatencyDesc, prometheus.GaugeValue, replayLatency.Float64, serverID)
		}
		if pendingReadIOs.Valid {
			ch <- prometheus.MustNewConstMetric(auroraPendingReadIOsDesc, prometheus.GaugeValue, float64(pendingReadIOs.Int64), serverID)
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
