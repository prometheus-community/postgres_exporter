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

func init() {
	registerCollector("replication_slot", defaultEnabled, NewPGReplicationSlotCollector)
}

type PGReplicationSlotCollector struct {
	log log.Logger
}

func NewPGReplicationSlotCollector(config collectorConfig) (Collector, error) {
	return &PGReplicationSlotCollector{log: config.logger}, nil
}

var (
	pgReplicationSlotCurrentWalDesc = prometheus.NewDesc(
		"pg_replication_slot_current_wal_lsn",
		"current wal lsn value",
		[]string{"slot_name"}, nil,
	)

	pgReplicationSlotCurrentFlushDesc = prometheus.NewDesc(
		"pg_replication_slot_confirmed_flush_lsn",
		"last lsn confirmed flushed to the replication slot",
		[]string{"slot_name"}, nil,
	)

	pgReplicationSlotIsActiveDesc = prometheus.NewDesc(
		"pg_replication_slot_is_active",
		"whether the replication slot is active or not",
		[]string{"slot_name"}, nil,
	)

	pgReplicationSlotQuery = `SELECT
		slot_name,
		pg_current_wal_lsn() - '0/0' AS current_wal_lsn,
		coalesce(confirmed_flush_lsn, '0/0') - '0/0',
		active
	FROM
		pg_replication_slots;`
)

func (PGReplicationSlotCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		pgReplicationSlotQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slotName string
		var walLsn int64
		var flushLsn int64
		var isActive bool

		if err := rows.Scan(&slotName, &walLsn, &flushLsn, &isActive); err != nil {
			return err
		}

		isActiveValue := 0
		if isActive {
			isActiveValue = 1
		}

		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotCurrentWalDesc,
			prometheus.GaugeValue, float64(walLsn), slotName,
		)
		if isActive {
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotCurrentFlushDesc,
				prometheus.GaugeValue, float64(flushLsn), slotName,
			)
		}
		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotIsActiveDesc,
			prometheus.GaugeValue, float64(isActiveValue), slotName,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
