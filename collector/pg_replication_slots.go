// Copyright 2022 The Prometheus Authors
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

func NewPGReplicationSlotCollector(logger log.Logger) (Collector, error) {
	return &PGReplicationSlotCollector{log: logger}, nil
}

var pgReplicationSlot = map[string]*prometheus.Desc{
	"lsn_distance": prometheus.NewDesc(
		"pg_replication_slot_lsn_distance",
		"Disk space used by the database",
		[]string{"slot_name"}, nil,
	),
}

func (PGReplicationSlotCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		`SELECT
			slot_name,
			(pg_current_wal_lsn() - confirmed_flush_lsn) AS lsn_distance
			FROM
		pg_replication_slots;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slot_name string
		var size int64
		if err := rows.Scan(&slot_name, &size); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlot["size_bytes"],
			prometheus.GaugeValue, float64(size), slot_name,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
