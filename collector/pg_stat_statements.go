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

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const statStatementsSubsystem = "stat_statements"

func init() {
	// WARNING:
	//   Disabled by default because this set of metrics can be quite expensive on a busy server
	//   Every unique query will cause a new timeseries to be created
	registerCollector(statStatementsSubsystem, defaultDisabled, NewPGStatStatementsCollector)
}

type PGStatStatementsCollector struct {
	log log.Logger
}

func NewPGStatStatementsCollector(config collectorConfig) (Collector, error) {
	return &PGStatStatementsCollector{log: config.logger}, nil
}

var (
	statSTatementsCallsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statStatementsSubsystem, "calls_total"),
		"Number of times executed",
		[]string{"user", "datname", "queryid"},
		prometheus.Labels{},
	)
	statStatementsSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statStatementsSubsystem, "seconds_total"),
		"Total time spent in the statement, in seconds",
		[]string{"user", "datname", "queryid"},
		prometheus.Labels{},
	)
	statStatementsRowsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statStatementsSubsystem, "rows_total"),
		"Total number of rows retrieved or affected by the statement",
		[]string{"user", "datname", "queryid"},
		prometheus.Labels{},
	)
	statStatementsBlockReadSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statStatementsSubsystem, "block_read_seconds_total"),
		"Total time the statement spent reading blocks, in seconds",
		[]string{"user", "datname", "queryid"},
		prometheus.Labels{},
	)
	statStatementsBlockWriteSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statStatementsSubsystem, "block_write_seconds_total"),
		"Total time the statement spent writing blocks, in seconds",
		[]string{"user", "datname", "queryid"},
		prometheus.Labels{},
	)

	pgStatStatementsQuery = `WITH percentiles AS (
    		SELECT percentile_cont(0.1) WITHIN GROUP (ORDER BY total_time) AS percentile
    		FROM pg_stat_statements
		)
		SELECT DISTINCT ON (pss.queryid, pg_get_userbyid(pss.userid), pg_database.datname)
    		pg_get_userbyid(pss.userid) as user,
    		pg_database.datname,
    		pss.queryid,
    		pss.calls as calls_total,
    		pss.total_time / 1000.0 as seconds_total,
    		pss.rows as rows_total,
    		pss.blk_read_time / 1000.0 as block_read_seconds_total,
    		pss.blk_write_time / 1000.0 as block_write_seconds_total
		FROM pg_stat_statements pss
		JOIN pg_database ON pg_database.oid = pss.dbid
			CROSS JOIN percentiles
		WHERE pss.total_time > (SELECT percentile FROM percentiles)
		ORDER BY pss.queryid, pg_get_userbyid(pss.userid) DESC, pg_database.datname
		LIMIT 100;`

	pgStatStatementsNewQuery = `SELECT DISTINCT ON (pss.queryid, pg_get_userbyid(pss.userid), pg_database.datname)
    		pg_get_userbyid(pss.userid) AS user,
    		pg_database.datname AS database_name,
    		pss.queryid,
    		pss.calls AS calls_total,
    		pss.total_exec_time / 1000.0 AS seconds_total,
    		pss.rows AS rows_total,
    		pss.blk_read_time / 1000.0 AS block_read_seconds_total,
    		pss.blk_write_time / 1000.0 AS block_write_seconds_total
		FROM pg_stat_statements pss
		JOIN pg_database ON pg_database.oid = pss.dbid
		JOIN (
    			SELECT percentile_cont(0.1) WITHIN GROUP (ORDER BY total_exec_time) AS percentile_val
    			FROM pg_stat_statements
		) AS perc ON pss.total_exec_time > perc.percentile_val
		ORDER BY pss.queryid, pg_get_userbyid(pss.userid) DESC, pg_database.datname
		LIMIT 100;`
)

func (PGStatStatementsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	query := pgStatStatementsQuery
	if instance.version.GE(semver.MustParse("13.0.0")) {
		query = pgStatStatementsNewQuery
	}

	db := instance.getDB()
	rows, err := db.QueryContext(ctx, query)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var user, datname, queryid sql.NullString
		var callsTotal, rowsTotal sql.NullInt64
		var secondsTotal, blockReadSecondsTotal, blockWriteSecondsTotal sql.NullFloat64

		if err := rows.Scan(&user, &datname, &queryid, &callsTotal, &secondsTotal, &rowsTotal, &blockReadSecondsTotal, &blockWriteSecondsTotal); err != nil {
			return err
		}

		userLabel := "unknown"
		if user.Valid {
			userLabel = user.String
		}
		datnameLabel := "unknown"
		if datname.Valid {
			datnameLabel = datname.String
		}
		queryidLabel := "unknown"
		if queryid.Valid {
			queryidLabel = queryid.String
		}

		callsTotalMetric := 0.0
		if callsTotal.Valid {
			callsTotalMetric = float64(callsTotal.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statSTatementsCallsTotal,
			prometheus.CounterValue,
			callsTotalMetric,
			userLabel, datnameLabel, queryidLabel,
		)

		secondsTotalMetric := 0.0
		if secondsTotal.Valid {
			secondsTotalMetric = secondsTotal.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statStatementsSecondsTotal,
			prometheus.CounterValue,
			secondsTotalMetric,
			userLabel, datnameLabel, queryidLabel,
		)

		rowsTotalMetric := 0.0
		if rowsTotal.Valid {
			rowsTotalMetric = float64(rowsTotal.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			statStatementsRowsTotal,
			prometheus.CounterValue,
			rowsTotalMetric,
			userLabel, datnameLabel, queryidLabel,
		)

		blockReadSecondsTotalMetric := 0.0
		if blockReadSecondsTotal.Valid {
			blockReadSecondsTotalMetric = blockReadSecondsTotal.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statStatementsBlockReadSecondsTotal,
			prometheus.CounterValue,
			blockReadSecondsTotalMetric,
			userLabel, datnameLabel, queryidLabel,
		)

		blockWriteSecondsTotalMetric := 0.0
		if blockWriteSecondsTotal.Valid {
			blockWriteSecondsTotalMetric = blockWriteSecondsTotal.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statStatementsBlockWriteSecondsTotal,
			prometheus.CounterValue,
			blockWriteSecondsTotalMetric,
			userLabel, datnameLabel, queryidLabel,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
