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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const xidSubsystem = "xid"

func init() {
	registerCollector(xidSubsystem, defaultEnabled, NewPGXidCollector)
}

type PGXidCollector struct {
	log log.Logger
}

func NewPGXidCollector(config collectorConfig) (Collector, error) {
	return &PGXidCollector{log: config.logger}, nil
}

var (
	xidCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, xidSubsystem, "current"),
		"Current 64-bit transaction id of the query used to collect this metric (truncated to low 52 bits)",
		[]string{}, prometheus.Labels{},
	)
	xidXmin = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, xidSubsystem, "xmin"),
		"Oldest transaction id of a transaction still in progress, i.e. not known committed or aborted (truncated to low 52 bits)",
		[]string{}, prometheus.Labels{},
	)
	xidXminAge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, xidSubsystem, "xmin_age"),
		"Age of oldest transaction still not committed or aborted measured in transaction ids",
		[]string{}, prometheus.Labels{},
	)

	xidQuery = `
	SELECT
		CASE WHEN pg_is_in_recovery() THEN 'NaN'::float ELSE txid_current() % (2^52)::bigint END AS current,
		CASE WHEN pg_is_in_recovery() THEN 'NaN'::float ELSE txid_snapshot_xmin(txid_current_snapshot()) % (2^52)::bigint END AS xmin,
		CASE WHEN pg_is_in_recovery() THEN 'NaN'::float ELSE txid_current() - txid_snapshot_xmin(txid_current_snapshot()) END AS xmin_age
	`
)

func (PGXidCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		xidQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var current, xmin, xminAge float64

		if err := rows.Scan(&current, &xmin, &xminAge); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			xidCurrent,
			prometheus.CounterValue,
			current,
		)
		ch <- prometheus.MustNewConstMetric(
			xidXmin,
			prometheus.CounterValue,
			xmin,
		)
		ch <- prometheus.MustNewConstMetric(
			xidXminAge,
			prometheus.GaugeValue,
			xminAge,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
