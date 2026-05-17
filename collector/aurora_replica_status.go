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

const auroraReplicaStatusSubsystem = "aurora_replica_status"

func init() {
	registerCollector(auroraReplicaStatusSubsystem, serverScope, NewAuroraReplicaStatusCollector)
}

// AuroraReplicaStatusCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_replica_status() function. It reports no data on
// servers that are not Aurora.
type AuroraReplicaStatusCollector struct{}

func NewAuroraReplicaStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraReplicaStatusCollector{}, nil
}

var (
	auroraReplicaStatusLag = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaStatusSubsystem, "lag_seconds"),
		"Replica lag behind the writer instance in seconds.",
		[]string{"server_id"}, nil,
	)
	auroraReplicaStatusReplayLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaStatusSubsystem, "replay_latency_seconds"),
		"Expected log replay latency in seconds.",
		[]string{"server_id"}, nil,
	)
	auroraReplicaStatusPendingReadIOs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraReplicaStatusSubsystem, "pending_read_ios"),
		"Outstanding page reads pending on the instance.",
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
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraReplicaStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var serverID string
		var replicaLagMsec, replayLatencyUsec sql.NullFloat64
		var pendingReadIOs sql.NullInt64

		if err := rows.Scan(&serverID, &replicaLagMsec, &replayLatencyUsec, &pendingReadIOs); err != nil {
			return err
		}

		if replicaLagMsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraReplicaStatusLag, prometheus.GaugeValue, replicaLagMsec.Float64/1e3, serverID)
		}
		if replayLatencyUsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraReplicaStatusReplayLatency, prometheus.GaugeValue, replayLatencyUsec.Float64/1e6, serverID)
		}
		if pendingReadIOs.Valid {
			ch <- prometheus.MustNewConstMetric(auroraReplicaStatusPendingReadIOs, prometheus.GaugeValue, float64(pendingReadIOs.Int64), serverID)
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
