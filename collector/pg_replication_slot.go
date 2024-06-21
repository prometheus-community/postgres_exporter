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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const replicationSlotSubsystem = "replication_slot"

func init() {
	registerCollector(replicationSlotSubsystem, defaultEnabled, NewPGReplicationSlotCollector)
}

type PGReplicationSlotCollector struct {
	log log.Logger
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
		active,
		wal_status
	FROM pg_replication_slots;`
)

func (PGReplicationSlotCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		pgReplicationSlotQuery)
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
		var walStatus sql.NullString
		if err := rows.Scan(&slotName, &slotType, &walLSN, &flushLSN, &isActive, &walStatus); err != nil {
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

		if walStatus.Valid {
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotWalStatus,
				prometheus.GaugeValue, 1, slotNameLabel, slotTypeLabel, walStatus.String,
			)
		}
	}
	return rows.Err()
}
