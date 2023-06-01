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

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("replication", defaultEnabled, NewPGReplicationCollector)
}

type PGReplicationCollector struct {
}

func NewPGReplicationCollector(collectorConfig) (Collector, error) {
	return &PGPostmasterCollector{}, nil
}

var pgReplication = map[string]*prometheus.Desc{
	"replication_lag": prometheus.NewDesc(
		"pg_replication_lag",
		"Replication lag behind master in seconds",
		[]string{"process_name"}, nil,
	),
}

func (c *PGReplicationCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	row := db.QueryRowContext(ctx,
		`SELECT
			CASE
				WHEN NOT pg_is_in_recovery() THEN 0
				ELSE GREATEST (0, EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())))
				END AS lag`)

	var lag float64
	err := row.Scan(&lag)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		pgReplication["replication_lag"],
		prometheus.GaugeValue, lag, "replication",
	)
	return nil
}
