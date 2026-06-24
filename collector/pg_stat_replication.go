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

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const statReplicationSubsystem = "stat_replication"

func init() {
	registerCollector(statReplicationSubsystem, defaultEnabled, NewPGStatReplicationCollector)
}

type PGStatReplicationCollector struct{}

func NewPGStatReplicationCollector(collectorConfig) (Collector, error) {
	return &PGStatReplicationCollector{}, nil
}

var (
	statReplicationLabels = []string{"application_name", "client_addr", "state", "slot_name"}

	statReplicationCurrentWalLSNBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statReplicationSubsystem, "pg_current_wal_lsn_bytes"),
		"WAL position in bytes",
		statReplicationLabels,
		prometheus.Labels{},
	)
	statReplicationWalLSNDiffDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statReplicationSubsystem, "pg_wal_lsn_diff"),
		"Lag in bytes between master and slave",
		statReplicationLabels,
		prometheus.Labels{},
	)
	statReplicationXlogLocationDiffDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statReplicationSubsystem, "pg_xlog_location_diff"),
		"Lag in bytes between master and slave",
		statReplicationLabels,
		prometheus.Labels{},
	)

	// slot_name is not a column of pg_stat_replication; it lives on
	// pg_replication_slots and is linked via the standby's WAL sender
	// PID. LEFT JOIN so clients that do not use a named slot still get
	// a row with an empty slot_name label.
	statReplicationQuery = `
			SELECT
				application_name,
				client_addr::text,
				state,
				s.slot_name,
				(case pg_is_in_recovery() when 't' then pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_lsn('0/0'))::float else pg_wal_lsn_diff(pg_current_wal_lsn(), pg_lsn('0/0'))::float end) AS pg_current_wal_lsn_bytes,
				(case pg_is_in_recovery() when 't' then pg_wal_lsn_diff(pg_last_wal_receive_lsn(), replay_lsn)::float else pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float end) AS pg_wal_lsn_diff
			FROM pg_stat_replication
			LEFT JOIN pg_replication_slots s ON s.active_pid = pg_stat_replication.pid
			`

	statReplicationQueryBefore10 = `
			SELECT
				application_name,
				client_addr::text,
				state,
				s.slot_name,
				(case pg_is_in_recovery() when 't' then pg_xlog_location_diff(pg_last_xlog_receive_location(), replay_location)::float else pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float end) AS pg_xlog_location_diff
			FROM pg_stat_replication
			LEFT JOIN pg_replication_slots s ON s.active_pid = pg_stat_replication.pid
			`
)

func (PGStatReplicationCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	switch {
	case instance.version.GTE(semver.MustParse("10.0.0")):
		return updateStatReplication(ctx, instance, ch)
	case instance.version.GTE(semver.MustParse("9.2.0")):
		return updateStatReplicationBefore10(ctx, instance, ch)
	default:
		return nil
	}
}

func updateStatReplication(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, statReplicationQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var applicationName, clientAddr, state, slotName sql.NullString
		var currentWalLSNBytes, walLSNDiff sql.NullFloat64
		if err := rows.Scan(&applicationName, &clientAddr, &state, &slotName, &currentWalLSNBytes, &walLSNDiff); err != nil {
			return err
		}
		labels := statReplicationLabelValues(applicationName, clientAddr, state, slotName)

		if currentWalLSNBytes.Valid {
			ch <- prometheus.MustNewConstMetric(
				statReplicationCurrentWalLSNBytesDesc,
				prometheus.GaugeValue,
				currentWalLSNBytes.Float64,
				labels...,
			)
		}
		if walLSNDiff.Valid {
			ch <- prometheus.MustNewConstMetric(
				statReplicationWalLSNDiffDesc,
				prometheus.GaugeValue,
				walLSNDiff.Float64,
				labels...,
			)
		}
	}

	return rows.Err()
}

func updateStatReplicationBefore10(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, statReplicationQueryBefore10)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var applicationName, clientAddr, state, slotName sql.NullString
		var xlogLocationDiff sql.NullFloat64
		if err := rows.Scan(&applicationName, &clientAddr, &state, &slotName, &xlogLocationDiff); err != nil {
			return err
		}

		if xlogLocationDiff.Valid {
			ch <- prometheus.MustNewConstMetric(
				statReplicationXlogLocationDiffDesc,
				prometheus.GaugeValue,
				xlogLocationDiff.Float64,
				statReplicationLabelValues(applicationName, clientAddr, state, slotName)...,
			)
		}
	}

	return rows.Err()
}

func statReplicationLabelValues(applicationName, clientAddr, state, slotName sql.NullString) []string {
	return []string{
		nullStringValue(applicationName),
		nullStringValue(clientAddr),
		nullStringValue(state),
		nullStringValue(slotName),
	}
}

func nullStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}
