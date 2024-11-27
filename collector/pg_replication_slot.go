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
	"log/slog"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const replicationSlotSubsystem = "replication_slot"

func init() {
	registerCollector(replicationSlotSubsystem, defaultEnabled, NewPGReplicationSlotCollector)
}

type PGReplicationSlotCollector struct {
	log *slog.Logger
}

func NewPGReplicationSlotCollector(config collectorConfig) (Collector, error) {
	return &PGReplicationSlotCollector{log: config.logger}, nil
}

var (
	pgReplicationSlotCurrentWalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_current_wal_lsn",
		),
		"current wal lsn value",
		[]string{"slot_name", "slot_type"}, nil,
	)
	pgReplicationSlotCurrentFlushDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_confirmed_flush_lsn",
		),
		"last lsn confirmed flushed to the replication slot",
		[]string{"slot_name", "slot_type"}, nil,
	)
	pgReplicationSlotIsActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_is_active",
		),
		"whether the replication slot is active or not",
		[]string{"slot_name", "slot_type"}, nil,
	)
	pgReplicationSlotSafeWal = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"safe_wal_size_bytes",
		),
		"number of bytes that can be written to WAL such that this slot is not in danger of getting in state lost",
		[]string{"slot_name", "slot_type"}, nil,
	)
	pgReplicationSlotWalStatus = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"wal_status",
		),
		"availability of WAL files claimed by this slot",
		[]string{"slot_name", "slot_type", "wal_status"}, nil,
	)
	pgReplicationSlotQuery = `SELECT
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
	pgReplicationSlotNewQuery = `SELECT
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

func (PGReplicationSlotCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	query := pgReplicationSlotQuery
	abovePG13 := instance.version.GTE(semver.MustParse("13.0.0"))
	if abovePG13 {
		query = pgReplicationSlotNewQuery
	}

	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		query)
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

		err := rows.Scan(r...)
		if err != nil {
			return err
		}

		isActiveValue := 0.0
		if isActive.Valid && isActive.Bool {
			isActiveValue = 1.0
		}
		slotNameLabel := "unknown"
		if slotName.Valid {
			slotNameLabel = slotName.String
		}
		slotTypeLabel := "unknown"
		if slotType.Valid {
			slotTypeLabel = slotType.String
		}

		var walLSNMetric float64
		if walLSN.Valid {
			walLSNMetric = walLSN.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotCurrentWalDesc,
			prometheus.GaugeValue, walLSNMetric, slotNameLabel, slotTypeLabel,
		)
		if isActive.Valid && isActive.Bool {
			var flushLSNMetric float64
			if flushLSN.Valid {
				flushLSNMetric = flushLSN.Float64
			}
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotCurrentFlushDesc,
				prometheus.GaugeValue, flushLSNMetric, slotNameLabel, slotTypeLabel,
			)
		}
		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotIsActiveDesc,
			prometheus.GaugeValue, isActiveValue, slotNameLabel, slotTypeLabel,
		)

		if safeWalSize.Valid {
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotSafeWal,
				prometheus.GaugeValue, float64(safeWalSize.Int64), slotNameLabel, slotTypeLabel,
			)
		}

		if walStatus.Valid {
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotWalStatus,
				prometheus.GaugeValue, 1, slotNameLabel, slotTypeLabel, walStatus.String,
			)
		}
	}
	return rows.Err()
}
