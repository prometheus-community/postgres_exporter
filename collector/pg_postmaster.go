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
	registerCollector("postmaster", defaultEnabled, NewPGPostmasterCollector)
}

type PGPostmasterCollector struct {
}

func NewPGPostmasterCollector(collectorConfig) (Collector, error) {
	return &PGPostmasterCollector{}, nil
}

var pgPostMasterStartTimeSeconds = prometheus.NewDesc(
	"pg_postmaster_start_time_seconds",
	"Time at which postmaster started",
	[]string{}, nil,
)

func (c *PGPostmasterCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	row := db.QueryRowContext(ctx,
		`SELECT
			pg_postmaster_start_time
		from pg_postmaster_start_time();`)

	var startTimeSeconds float64
	err := row.Scan(&startTimeSeconds)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		pgPostMasterStartTimeSeconds,
		prometheus.GaugeValue, startTimeSeconds,
	)
	return nil
}
