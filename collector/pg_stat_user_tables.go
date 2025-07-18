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

	"github.com/prometheus/client_golang/prometheus"
)

const userTableSubsystem = "stat_user_tables"

func init() {
	registerCollector(userTableSubsystem, defaultEnabled, NewPGStatUserTablesCollector)
}

type PGStatUserTablesCollector struct {
	log *slog.Logger
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
	statUserIndexSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "index_size_bytes"),
		"Total disk space used by this index, in bytes",
		[]string{"datname", "schemaname", "relname"},
		prometheus.Labels{},
	)
	statUserTableSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, userTableSubsystem, "table_size_bytes"),
		"Total disk space used by this table, in bytes",
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
		autoanalyze_count,
		pg_indexes_size(relid) as indexes_size,
		pg_table_size(relid) as table_size
	FROM
		pg_stat_user_tables`
)

func (c *PGStatUserTablesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statUserTablesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname, schemaname, relname sql.NullString
		var seqScan, seqTupRead, idxScan, idxTupFetch, nTupIns, nTupUpd, nTupDel, nTupHotUpd, nLiveTup, nDeadTup,
			nModSinceAnalyze, vacuumCount, autovacuumCount, analyzeCount, autoanalyzeCount, indexSize, tableSize sql.NullInt64
		var lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze sql.NullTime

		if err := rows.Scan(&datname, &schemaname, &relname, &seqScan, &seqTupRead, &idxScan, &idxTupFetch, &nTupIns, &nTupUpd, &nTupDel, &nTupHotUpd, &nLiveTup, &nDeadTup, &nModSinceAnalyze, &lastVacuum, &lastAutovacuum, &lastAnalyze, &lastAutoanalyze, &vacuumCount, &autovacuumCount, &analyzeCount, &autoanalyzeCount, &indexSize, &tableSize); err != nil {
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

		seqScanMetric := 0.0
		if seqScan.Valid {
			seqScanMetric = float64(seqScan.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesSeqScan,
			prometheus.CounterValue,
			seqScanMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		seqTupReadMetric := 0.0
		if seqTupRead.Valid {
			seqTupReadMetric = float64(seqTupRead.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesSeqTupRead,
			prometheus.CounterValue,
			seqTupReadMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		idxScanMetric := 0.0
		if idxScan.Valid {
			idxScanMetric = float64(idxScan.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesIdxScan,
			prometheus.CounterValue,
			idxScanMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		idxTupFetchMetric := 0.0
		if idxTupFetch.Valid {
			idxTupFetchMetric = float64(idxTupFetch.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesIdxTupFetch,
			prometheus.CounterValue,
			idxTupFetchMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nTupInsMetric := 0.0
		if nTupIns.Valid {
			nTupInsMetric = float64(nTupIns.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupIns,
			prometheus.CounterValue,
			nTupInsMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nTupUpdMetric := 0.0
		if nTupUpd.Valid {
			nTupUpdMetric = float64(nTupUpd.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupUpd,
			prometheus.CounterValue,
			nTupUpdMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nTupDelMetric := 0.0
		if nTupDel.Valid {
			nTupDelMetric = float64(nTupDel.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupDel,
			prometheus.CounterValue,
			nTupDelMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nTupHotUpdMetric := 0.0
		if nTupHotUpd.Valid {
			nTupHotUpdMetric = float64(nTupHotUpd.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNTupHotUpd,
			prometheus.CounterValue,
			nTupHotUpdMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nLiveTupMetric := 0.0
		if nLiveTup.Valid {
			nLiveTupMetric = float64(nLiveTup.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNLiveTup,
			prometheus.GaugeValue,
			nLiveTupMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nDeadTupMetric := 0.0
		if nDeadTup.Valid {
			nDeadTupMetric = float64(nDeadTup.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNDeadTup,
			prometheus.GaugeValue,
			nDeadTupMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		nModSinceAnalyzeMetric := 0.0
		if nModSinceAnalyze.Valid {
			nModSinceAnalyzeMetric = float64(nModSinceAnalyze.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesNModSinceAnalyze,
			prometheus.GaugeValue,
			nModSinceAnalyzeMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		lastVacuumMetric := 0.0
		if lastVacuum.Valid {
			lastVacuumMetric = float64(lastVacuum.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastVacuum,
			prometheus.GaugeValue,
			lastVacuumMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		lastAutovacuumMetric := 0.0
		if lastAutovacuum.Valid {
			lastAutovacuumMetric = float64(lastAutovacuum.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAutovacuum,
			prometheus.GaugeValue,
			lastAutovacuumMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		lastAnalyzeMetric := 0.0
		if lastAnalyze.Valid {
			lastAnalyzeMetric = float64(lastAnalyze.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAnalyze,
			prometheus.GaugeValue,
			lastAnalyzeMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		lastAutoanalyzeMetric := 0.0
		if lastAutoanalyze.Valid {
			lastAutoanalyzeMetric = float64(lastAutoanalyze.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesLastAutoanalyze,
			prometheus.GaugeValue,
			lastAutoanalyzeMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		vacuumCountMetric := 0.0
		if vacuumCount.Valid {
			vacuumCountMetric = float64(vacuumCount.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesVacuumCount,
			prometheus.CounterValue,
			vacuumCountMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		autovacuumCountMetric := 0.0
		if autovacuumCount.Valid {
			autovacuumCountMetric = float64(autovacuumCount.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAutovacuumCount,
			prometheus.CounterValue,
			autovacuumCountMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		analyzeCountMetric := 0.0
		if analyzeCount.Valid {
			analyzeCountMetric = float64(analyzeCount.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAnalyzeCount,
			prometheus.CounterValue,
			analyzeCountMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		autoanalyzeCountMetric := 0.0
		if autoanalyzeCount.Valid {
			autoanalyzeCountMetric = float64(autoanalyzeCount.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTablesAutoanalyzeCount,
			prometheus.CounterValue,
			autoanalyzeCountMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		indexSizeMetric := 0.0
		if indexSize.Valid {
			indexSizeMetric = float64(indexSize.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserIndexSize,
			prometheus.GaugeValue,
			indexSizeMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)

		tableSizeMetric := 0.0
		if tableSize.Valid {
			tableSizeMetric = float64(tableSize.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statUserTableSize,
			prometheus.GaugeValue,
			tableSizeMetric,
			datnameLabel, schemanameLabel, relnameLabel,
		)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
