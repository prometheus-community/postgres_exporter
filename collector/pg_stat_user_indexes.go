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
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(statUserIndexesSubsystem, defaultDisabled, NewPGStatUserIndexesCollector)
}

type PGStatUserIndexesCollector struct {
	log log.Logger
}

const statUserIndexesSubsystem = "stat_user_indexes"

func NewPGStatUserIndexesCollector(config collectorConfig) (Collector, error) {
	return &PGStatUserIndexesCollector{log: config.logger}, nil
}

var (
	statUserIndexesIdxScan = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_scan_total"),
		"Number of scans for this index",
		[]string{"datname", "schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	statUserIndexesLastIdxScan = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "last_idx_scan_time"),
		"Last timestamp of scan for this index",
		[]string{"datname", "schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	statUserIndexesIdxTupRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_tup_read"),
		"Number of tuples read for this index",
		[]string{"datname", "schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)

	statUserIndexesIdxTupFetch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statUserIndexesSubsystem, "idx_tup_fetch"),
		"Number of tuples fetch for this index",
		[]string{"datname", "schemaname", "relname", "indexrelname"},
		prometheus.Labels{},
	)
)

func statUserIndexesQuery(columns []string) string {
	return fmt.Sprintf("SELECT %s FROM pg_stat_user_indexes;", strings.Join(columns, ","))
}

func (c *PGStatUserIndexesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	columns := []string{
		"current_database() datname",
		"schemaname",
		"relname",
		"indexrelname",
		"idx_scan",
		"idx_tup_read",
		"idx_tup_fetch",
	}

	lastIdxScanAvail := instance.version.GTE(semver.MustParse("16.0.0"))
	if lastIdxScanAvail {
		columns = append(columns, "date_part('epoch', last_idx_scan) as last_idx_scan")
	}

	rows, err := db.QueryContext(ctx,
		statUserIndexesQuery(columns),
	)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var datname, schemaname, relname, indexrelname sql.NullString
		var idxScan, lastIdxScan, idxTupRead, idxTupFetch sql.NullFloat64

		r := []any{
			&datname,
			&schemaname,
			&relname,
			&indexrelname,
			&idxScan,
			&idxTupRead,
			&idxTupFetch,
		}

		if lastIdxScanAvail {
			r = append(r, &lastIdxScan)
		}

		if err := rows.Scan(r...); err != nil {
			return err
		}
		datnameLabel := "unknown"
		if datname.Valid {
			datnameLabel = datname.String
		}
		schemanameLabel := "unknown"
		if schemaname.Valid {
			schemanameLabel = schemaname.String
		}
		relnameLabel := "unknown"
		if relname.Valid {
			relnameLabel = relname.String
		}
		indexrelnameLabel := "unknown"
		if indexrelname.Valid {
			indexrelnameLabel = indexrelname.String
		}

		if lastIdxScanAvail && !lastIdxScan.Valid {
			level.Debug(c.log).Log("msg", "Skipping collecting metric because it has no active_time")
			continue
		}

		labels := []string{datnameLabel, schemanameLabel, relnameLabel, indexrelnameLabel}

		idxScanMetric := 0.0
		if idxScan.Valid {
			idxScanMetric = idxScan.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxScan,
			prometheus.CounterValue,
			idxScanMetric,
			labels...,
		)

		idxTupReadMetric := 0.0
		if idxTupRead.Valid {
			idxTupReadMetric = idxTupRead.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxTupRead,
			prometheus.CounterValue,
			idxTupReadMetric,
			labels...,
		)

		idxTupFetchMetric := 0.0
		if idxTupFetch.Valid {
			idxTupFetchMetric = idxTupFetch.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statUserIndexesIdxTupFetch,
			prometheus.CounterValue,
			idxTupFetchMetric,
			labels...,
		)

		if lastIdxScanAvail {
			ch <- prometheus.MustNewConstMetric(
				statUserIndexesLastIdxScan,
				prometheus.CounterValue,
				lastIdxScan.Float64,
				labels...,
			)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
