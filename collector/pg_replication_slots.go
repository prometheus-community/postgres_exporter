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

const replicationSlotsSubsystem = "replication_slots"

func init() {
	registerCollector(replicationSlotsSubsystem, defaultEnabled, NewPGReplicationSlotsCollector)
}

type PGReplicationSlotsCollector struct{}

func NewPGReplicationSlotsCollector(collectorConfig) (Collector, error) {
	return &PGReplicationSlotsCollector{}, nil
}

var (
	replicationSlotsLabels = []string{"slot_name", "database"}

	replicationSlotsActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, replicationSlotsSubsystem, "active"),
		"Flag indicating if the slot is active",
		replicationSlotsLabels,
		prometheus.Labels{},
	)
	replicationSlotsWalLSNDiffDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, replicationSlotsSubsystem, "pg_wal_lsn_diff"),
		"Replication lag in bytes",
		replicationSlotsLabels,
		prometheus.Labels{},
	)
	replicationSlotsXlogLocationDiffDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, replicationSlotsSubsystem, "pg_xlog_location_diff"),
		"Unknown metric from pg_replication_slots",
		replicationSlotsLabels,
		prometheus.Labels{},
	)

	replicationSlotsQuery = `
			SELECT slot_name, database, active,
				(case pg_is_in_recovery() when 't' then pg_wal_lsn_diff(pg_last_wal_receive_lsn(), restart_lsn) else pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) end) as pg_wal_lsn_diff
			FROM pg_replication_slots
			`

	replicationSlotsQueryBefore10 = `
			SELECT slot_name, database, active,
				(case pg_is_in_recovery() when 't' then pg_xlog_location_diff(pg_last_xlog_receive_location(), restart_lsn) else pg_xlog_location_diff(pg_current_xlog_location(), restart_lsn) end) as pg_xlog_location_diff
			FROM pg_replication_slots
			`
)

func (PGReplicationSlotsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	switch {
	case instance.version.GTE(semver.MustParse("10.0.0")):
		return updateReplicationSlots(ctx, instance, ch)
	case instance.version.GTE(semver.MustParse("9.4.0")):
		return updateReplicationSlotsBefore10(ctx, instance, ch)
	default:
		return nil
	}
}

func updateReplicationSlots(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, replicationSlotsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slotName, database sql.NullString
		var active sql.NullBool
		var walLSNDiff sql.NullFloat64
		if err := rows.Scan(&slotName, &database, &active, &walLSNDiff); err != nil {
			return err
		}
		labels := replicationSlotsLabelValues(slotName, database)
		emitReplicationSlotsActive(ch, active, labels)
		ch <- prometheus.MustNewConstMetric(
			replicationSlotsWalLSNDiffDesc,
			prometheus.GaugeValue,
			nullFloat64Value(walLSNDiff),
			labels...,
		)
	}

	return rows.Err()
}

func updateReplicationSlotsBefore10(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, replicationSlotsQueryBefore10)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slotName, database sql.NullString
		var active sql.NullBool
		var xlogLocationDiff sql.NullFloat64
		if err := rows.Scan(&slotName, &database, &active, &xlogLocationDiff); err != nil {
			return err
		}
		labels := replicationSlotsLabelValues(slotName, database)
		emitReplicationSlotsActive(ch, active, labels)
		ch <- prometheus.MustNewConstMetric(
			replicationSlotsXlogLocationDiffDesc,
			prometheus.UntypedValue,
			nullFloat64Value(xlogLocationDiff),
			labels...,
		)
	}

	return rows.Err()
}

func emitReplicationSlotsActive(ch chan<- prometheus.Metric, active sql.NullBool, labels []string) {
	activeValue := 0.0
	if active.Valid && active.Bool {
		activeValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		replicationSlotsActiveDesc,
		prometheus.GaugeValue,
		activeValue,
		labels...,
	)
}

func replicationSlotsLabelValues(slotName, database sql.NullString) []string {
	return []string{nullStringValue(slotName), nullStringValue(database)}
}

func nullStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullFloat64Value(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0
}
