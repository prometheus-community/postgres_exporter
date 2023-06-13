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
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const userTableSubsystem = "stat_user_tables"

func init() {
	registerCollector(userTableSubsystem, defaultEnabled, NewPGStatUserTablesCollector)
}

type PGStatUserTablesCollector struct {
	log log.Logger
}

func NewPGStatUserTablesCollector(config collectorConfig) (Collector, error) {
	return &PGStatUserTablesCollector{log: config.logger}, nil
}

var (
	statUserTablesSeqScan = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "seq_scan"),
		"Number of sequential scans initiated on this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesSeqTupRead = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "seq_tup_read"),
		"Number of live rows fetched by sequential scans",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesIdxScan = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "idx_scan"),
		"Number of index scans initiated on this table",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesIdxTupFetch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "idx_tup_fetch"),
		"Number of live rows fetched by index scans",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNTupIns = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_tup_ins"),
		"Number of rows inserted",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNTupUpd = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_tup_upd"),
		"Number of rows updated",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNTupDel = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_tup_del"),
		"Number of rows deleted",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNTupHotUpd = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_tup_hot_upd"),
		"Number of rows HOT updated (i.e., with no separate index update required)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNLiveTup = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_live_tup"),
		"Estimated number of live rows",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNDeadTup = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_dead_tup"),
		"Estimated number of dead rows",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesNModSinceAnalyze = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "n_mod_since_analyze"),
		"Estimated number of rows changed since last analyze",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesLastVacuum = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "last_vacuum"),
		"Last time at which this table was manually vacuumed (not counting VACUUM FULL)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesLastAutovacuum = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "last_autovacuum"),
		"Last time at which this table was vacuumed by the autovacuum daemon",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesLastAnalyze = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "last_analyze"),
		"Last time at which this table was manually analyzed",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesLastAutoanalyze = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "last_autoanalyze"),
		"Last time at which this table was analyzed by the autovacuum daemon",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesVacuumCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "vacuum_count"),
		"Number of times this table has been manually vacuumed (not counting VACUUM FULL)",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesAutovacuumCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "autovacuum_count"),
		"Number of times this table has been vacuumed by the autovacuum daemon",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesAnalyzeCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "analyze_count"),
		"Number of times this table has been manually analyzed",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTablesAutoanalyzeCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "autoanalyze_count"),
		"Number of times this table has been analyzed by the autovacuum daemon",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)

	statUserTablesQuery = `SELECT
		current_database() datname,
		schemaname,
		relname,
		seq_scan,
		seq_tup_read,
		idx_scan,
		idx_tup_fetch,
		n_tup_ins,
		n_tup_upd,
		n_tup_del,
		n_tup_hot_upd,
		n_live_tup,
		n_dead_tup,
		n_mod_since_analyze,
		COALESCE(last_vacuum, '1970-01-01Z') as last_vacuum,
		COALESCE(last_autovacuum, '1970-01-01Z') as last_autovacuum,
		COALESCE(last_analyze, '1970-01-01Z') as last_analyze,
		COALESCE(last_autoanalyze, '1970-01-01Z') as last_autoanalyze,
		vacuum_count,
		autovacuum_count,
		analyze_count,
		autoanalyze_count
	FROM
		pg_stat_user_tables`
)

func (c *PGStatUserTablesCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		statUserTablesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname string
		var schemaname string
		var relname string
		var seqScan int64
		var seqTupRead int64
		var idxScan int64
		var idxTupFetch int64
		var nTupIns int64
		var nTupUpd int64
		var nTupDel int64
		var nTupHotUpd int64
		var nLiveTup int64
		var nDeadTup int64
		var nModSinceAnalyze int64
		var lastVacuum time.Time
		var lastAutovacuum time.Time
		var lastAnalyze time.Time
		var lastAutoanalyze time.Time
		var vacuumCount int64
		var autovacuumCount int64
		var analyzeCount int64
		var autoanalyzeCount int64

		if err := rows.Scan(&datname, &schemaname, &relname, &seqScan, &seqTupRead, &idxScan, &idxTupFetch, &nTupIns, &nTupUpd, &nTupDel, &nTupHotUpd, &nLiveTup, &nDeadTup, &nModSinceAnalyze, &lastVacuum, &lastAutovacuum, &lastAnalyze, &lastAutoanalyze, &vacuumCount, &autovacuumCount, &analyzeCount, &autoanalyzeCount); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statUserTablesSeqScan,
			prometheus.CounterValue,
			float64(seqScan),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesSeqTupRead,
			prometheus.CounterValue,
			float64(seqTupRead),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesIdxScan,
			prometheus.CounterValue,
			float64(idxScan),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesIdxTupFetch,
			prometheus.CounterValue,
			float64(idxTupFetch),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupIns,
			prometheus.CounterValue,
			float64(nTupIns),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupUpd,
			prometheus.CounterValue,
			float64(nTupUpd),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupDel,
			prometheus.CounterValue,
			float64(nTupDel),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupHotUpd,
			prometheus.CounterValue,
			float64(nTupHotUpd),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNLiveTup,
			prometheus.GaugeValue,
			float64(nLiveTup),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNDeadTup,
			prometheus.GaugeValue,
			float64(nDeadTup),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNModSinceAnalyze,
			prometheus.GaugeValue,
			float64(nModSinceAnalyze),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastVacuum,
			prometheus.GaugeValue,
			float64(lastVacuum.Unix()),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAutovacuum,
			prometheus.GaugeValue,
			float64(lastAutovacuum.Unix()),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAnalyze,
			prometheus.GaugeValue,
			float64(lastAnalyze.Unix()),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAutoanalyze,
			prometheus.GaugeValue,
			float64(lastAutoanalyze.Unix()),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesVacuumCount,
			prometheus.CounterValue,
			float64(vacuumCount),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAutovacuumCount,
			prometheus.CounterValue,
			float64(autovacuumCount),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAnalyzeCount,
			prometheus.CounterValue,
			float64(analyzeCount),
			datname, schemaname, relname,
		)
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAutoanalyzeCount,
			prometheus.CounterValue,
			float64(autoanalyzeCount),
			datname, schemaname, relname,
		)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
