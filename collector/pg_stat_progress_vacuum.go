// Copyright 2025 The Prometheus Authors
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

	"github.com/prometheus/client_golang/prometheus"
)

const progressVacuumSubsystem = "stat_progress_vacuum"

func init() {
	registerCollector(progressVacuumSubsystem, defaultEnabled, NewPGStatProgressVacuumCollector)
}

type PGStatProgressVacuumCollector struct {
	log *slog.Logger
}

func NewPGStatProgressVacuumCollector(config collectorConfig) (Collector, error) {
	return &PGStatProgressVacuumCollector{log: config.logger}, nil
}

var vacuumPhases = []string{
	"initializing",
	"scanning heap",
	"vacuuming indexes",
	"vacuuming heap",
	"cleaning up indexes",
	"truncating heap",
	"performing final cleanup",
}

var (
	statProgressVacuumPhase = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "phase"),
		"Current vacuum phase (1 = active, 0 = inactive). Label 'phase' is human-readable.",
		[]string{"datname", "relname", "phase"},
		nil,
	)

	statProgressVacuumHeapBlksTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "heap_blks"),
		"Total number of heap blocks in the table being vacuumed.",
		[]string{"datname", "relname"},
		nil,
	)

	statProgressVacuumHeapBlksScanned = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "heap_blks_scanned"),
		"Number of heap blocks scanned so far.",
		[]string{"datname", "relname"},
		nil,
	)

	statProgressVacuumHeapBlksVacuumed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "heap_blks_vacuumed"),
		"Number of heap blocks vacuumed so far.",
		[]string{"datname", "relname"},
		nil,
	)

	statProgressVacuumIndexVacuumCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "index_vacuums"),
		"Number of completed index vacuum cycles.",
		[]string{"datname", "relname"},
		nil,
	)

	statProgressVacuumMaxDeadTuples = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "max_dead_tuples"),
		"Maximum number of dead tuples that can be stored before cleanup is performed.",
		[]string{"datname", "relname"},
		nil,
	)

	statProgressVacuumNumDeadTuples = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, progressVacuumSubsystem, "num_dead_tuples"),
		"Current number of dead tuples found so far.",
		[]string{"datname", "relname"},
		nil,
	)

	// This is the view definition of pg_stat_progress_vacuum, albeit without the conversion
	// of "phase" to a human-readable string. We will prefer the numeric representation.
	statProgressVacuumQuery = `SELECT
		d.datname,
		s.relid::regclass::text AS relname,
		s.param1 AS phase,
		s.param2 AS heap_blks_total,
		s.param3 AS heap_blks_scanned,
		s.param4 AS heap_blks_vacuumed,
		s.param5 AS index_vacuum_count,
		s.param6 AS max_dead_tuples,
		s.param7 AS num_dead_tuples
	FROM
		pg_stat_get_progress_info('VACUUM'::text)
		s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
	LEFT JOIN
		pg_database d ON s.datid = d.oid`
)

func (c *PGStatProgressVacuumCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statProgressVacuumQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			datname          sql.NullString
			relname          sql.NullString
			phase            sql.NullInt64
			heapBlksTotal    sql.NullInt64
			heapBlksScanned  sql.NullInt64
			heapBlksVacuumed sql.NullInt64
			indexVacuumCount sql.NullInt64
			maxDeadTuples    sql.NullInt64
			numDeadTuples    sql.NullInt64
		)

		if err := rows.Scan(
			&datname,
			&relname,
			&phase,
			&heapBlksTotal,
			&heapBlksScanned,
			&heapBlksVacuumed,
			&indexVacuumCount,
			&maxDeadTuples,
			&numDeadTuples,
		); err != nil {
			return err
		}

		datnameLabel := "unknown"
		if datname.Valid {
			datnameLabel = datname.String
		}
		relnameLabel := "unknown"
		if relname.Valid {
			relnameLabel = relname.String
		}

		labels := []string{datnameLabel, relnameLabel}

		var phaseMetric *float64
		if phase.Valid {
			v := float64(phase.Int64)
			phaseMetric = &v
		}

		for i, label := range vacuumPhases {
			v := 0.0
			// Only the current phase should be 1.0.
			if phaseMetric != nil && float64(i) == *phaseMetric {
				v = 1.0
			}
			labelsCopy := append(labels, label)
			ch <- prometheus.MustNewConstMetric(statProgressVacuumPhase, prometheus.GaugeValue, v, labelsCopy...)
		}

		heapTotal := 0.0
		if heapBlksTotal.Valid {
			heapTotal = float64(heapBlksTotal.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumHeapBlksTotal, prometheus.GaugeValue, heapTotal, labels...)

		heapScanned := 0.0
		if heapBlksScanned.Valid {
			heapScanned = float64(heapBlksScanned.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumHeapBlksScanned, prometheus.GaugeValue, heapScanned, labels...)

		heapVacuumed := 0.0
		if heapBlksVacuumed.Valid {
			heapVacuumed = float64(heapBlksVacuumed.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumHeapBlksVacuumed, prometheus.GaugeValue, heapVacuumed, labels...)

		indexCount := 0.0
		if indexVacuumCount.Valid {
			indexCount = float64(indexVacuumCount.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumIndexVacuumCount, prometheus.GaugeValue, indexCount, labels...)

		maxDead := 0.0
		if maxDeadTuples.Valid {
			maxDead = float64(maxDeadTuples.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumMaxDeadTuples, prometheus.GaugeValue, maxDead, labels...)

		numDead := 0.0
		if numDeadTuples.Valid {
			numDead = float64(numDeadTuples.Int64)
		}
		ch <- prometheus.MustNewConstMetric(statProgressVacuumNumDeadTuples, prometheus.GaugeValue, numDead, labels...)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
