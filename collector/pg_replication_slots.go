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
	replicationSlotsLabels     = []string{"slot_name", "database"}
	replicationSlotsSlotLabels = []string{"slot_name", "slot_type"}

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
	replicationSlotsCurrentWalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotsSubsystem,
			"slot_current_wal_lsn",
		),
		"current wal lsn value",
		replicationSlotsSlotLabels, nil,
	)
	replicationSlotsCurrentFlushDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotsSubsystem,
			"slot_confirmed_flush_lsn",
		),
		"last lsn confirmed flushed to the replication slot",
		replicationSlotsSlotLabels, nil,
	)
	replicationSlotsIsActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotsSubsystem,
			"slot_is_active",
		),
		"whether the replication slot is active or not",
		replicationSlotsSlotLabels, nil,
	)
	replicationSlotsSafeWal = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotsSubsystem,
			"safe_wal_size_bytes",
		),
		"number of bytes that can be written to WAL such that this slot is not in danger of getting in state lost",
		replicationSlotsSlotLabels, nil,
	)
	replicationSlotsWalStatus = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotsSubsystem,
			"wal_status",
		),
		"availability of WAL files claimed by this slot",
		[]string{"slot_name", "slot_type", "wal_status"}, nil,
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
	replicationSlotsSlotQuery = `SELECT
		slot_name,
		slot_type,
		CASE WHEN pg_is_in_recovery() THEN
		    pg_last_wal_receive_lsn() - '0/0'
		ELSE
		    pg_current_wal_lsn() - '0/0'
		END AS current_wal_lsn,
		COALESCE(confirmed_flush_lsn, '0/0') - '0/0' AS confirmed_flush_lsn,
		active
	FROM pg_replication_slots;`
	replicationSlotsSlotNewQuery = `SELECT
		slot_name,
		slot_type,
		CASE WHEN pg_is_in_recovery() THEN
		    pg_last_wal_receive_lsn() - '0/0'
		ELSE
		    pg_current_wal_lsn() - '0/0'
		END AS current_wal_lsn,
		COALESCE(confirmed_flush_lsn, '0/0') - '0/0' AS confirmed_flush_lsn,
		active,
		safe_wal_size,
		wal_status
	FROM pg_replication_slots;`
)

func (PGReplicationSlotsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	switch {
	case instance.version.GTE(semver.MustParse("10.0.0")):
		if err := updateReplicationSlots(ctx, instance, ch); err != nil {
			return err
		}
		return updateReplicationSlotsSlotMetrics(ctx, instance, ch)
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
		if walLSNDiff.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsWalLSNDiffDesc,
				prometheus.GaugeValue,
				walLSNDiff.Float64,
				labels...,
			)
		}
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
		if xlogLocationDiff.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsXlogLocationDiffDesc,
				prometheus.UntypedValue,
				xlogLocationDiff.Float64,
				labels...,
			)
		}
	}

	return rows.Err()
}

func updateReplicationSlotsSlotMetrics(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	query := replicationSlotsSlotQuery
	abovePG13 := instance.version.GTE(semver.MustParse("13.0.0"))
	if abovePG13 {
		query = replicationSlotsSlotNewQuery
	}

	db := instance.getDB()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slotName sql.NullString
		var slotType sql.NullString
		var walLSN sql.NullFloat64
		var flushLSN sql.NullFloat64
		var isActive sql.NullBool
		var safeWalSize sql.NullInt64
		var walStatus sql.NullString

		r := []any{
			&slotName,
			&slotType,
			&walLSN,
			&flushLSN,
			&isActive,
		}

		if abovePG13 {
			r = append(r, &safeWalSize)
			r = append(r, &walStatus)
		}

		if err := rows.Scan(r...); err != nil {
			return err
		}

		slotLabels := replicationSlotsSlotLabelValues(slotName, slotType)

		if walLSN.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsCurrentWalDesc,
				prometheus.GaugeValue, walLSN.Float64, slotLabels...,
			)
		}
		if isActive.Valid && isActive.Bool && flushLSN.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsCurrentFlushDesc,
				prometheus.GaugeValue, flushLSN.Float64, slotLabels...,
			)
		}
		emitReplicationSlotsSlotIsActive(ch, isActive, slotLabels)

		if safeWalSize.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsSafeWal,
				prometheus.GaugeValue, float64(safeWalSize.Int64), slotLabels...,
			)
		}

		if walStatus.Valid {
			ch <- prometheus.MustNewConstMetric(
				replicationSlotsWalStatus,
				prometheus.GaugeValue, 1, slotLabels[0], slotLabels[1], walStatus.String,
			)
		}
	}
	return rows.Err()
}

func emitReplicationSlotsActive(ch chan<- prometheus.Metric, active sql.NullBool, labels []string) {
	if !active.Valid {
		return
	}

	activeValue := 0.0
	if active.Bool {
		activeValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		replicationSlotsActiveDesc,
		prometheus.GaugeValue,
		activeValue,
		labels...,
	)
}

func emitReplicationSlotsSlotIsActive(ch chan<- prometheus.Metric, active sql.NullBool, labels []string) {
	if !active.Valid {
		return
	}

	activeValue := 0.0
	if active.Bool {
		activeValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		replicationSlotsIsActiveDesc,
		prometheus.GaugeValue,
		activeValue,
		labels...,
	)
}

func replicationSlotsLabelValues(slotName, database sql.NullString) []string {
	return []string{replicationSlotsNullStringValue(slotName), replicationSlotsNullStringValue(database)}
}

func replicationSlotsNullStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func replicationSlotsSlotLabelValues(slotName, slotType sql.NullString) []string {
	return []string{unknownStringValue(slotName), unknownStringValue(slotType)}
}

func unknownStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "unknown"
}
