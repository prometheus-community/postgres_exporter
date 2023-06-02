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
	// WARNING:
	//   Disabled by default because this set of metrics can be quite expensive on a busy server
	//   Every unique query will cause a new timeseries to be created
	registerCollector("statements", defaultDisabled, NewPGStatStatementsCollector)
}

type PGStatStatementsCollector struct {
	log log.Logger
}

var statStatementsSubsystem = "stat_statements"

func NewPGStatStatementsCollector(config collectorConfig) (Collector, error) {
	return &PGStatStatementsCollector{log: config.logger}, nil
}

var statSTatementsCallsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statStatementsSubsystem, "calls_total"),
	"Number of times executed",
	[]string{"user", "datname", "queryid"},
	prometheus.Labels{},
)

var statStatementsSecondsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statStatementsSubsystem, "seconds_total"),
	"Total time spent in the statement, in seconds",
	[]string{"user", "datname", "queryid"},
	prometheus.Labels{},
)

var statStatementsRowsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statStatementsSubsystem, "rows_total"),
	"Total number of rows retrieved or affected by the statement",
	[]string{"user", "datname", "queryid"},
	prometheus.Labels{},
)

var statStatementsBlockReadSecondsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statStatementsSubsystem, "block_read_seconds_total"),
	"Total time the statement spent reading blocks, in seconds",
	[]string{"user", "datname", "queryid"},
	prometheus.Labels{},
)

var statStatementsBlockWriteSecondsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, statStatementsSubsystem, "block_write_seconds_total"),
	"Total time the statement spent writing blocks, in seconds",
	[]string{"user", "datname", "queryid"},
	prometheus.Labels{},
)

func (PGStatStatementsCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx,
		`SELECT
			pg_get_userbyid(userid) as user,
			pg_database.datname,
			pg_stat_statements.queryid,
			pg_stat_statements.calls as calls_total,
			pg_stat_statements.total_time / 1000.0 as seconds_total,
			pg_stat_statements.rows as rows_total,
			pg_stat_statements.blk_read_time / 1000.0 as block_read_seconds_total,
			pg_stat_statements.blk_write_time / 1000.0 as block_write_seconds_total
			FROM pg_stat_statements
			JOIN pg_database
				ON pg_database.oid = pg_stat_statements.dbid
			WHERE
				total_time > (
				SELECT percentile_cont(0.1)
					WITHIN GROUP (ORDER BY total_time)
					FROM pg_stat_statements
				)
			ORDER BY seconds_total DESC
			LIMIT 100`)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var user string
		var datname string
		var queryid string
		var callsTotal int64
		var secondsTotal float64
		var rowsTotal int64
		var blockReadSecondsTotal float64
		var blockWriteSecondsTotal float64

		if err := rows.Scan(&user, &datname, &queryid, &callsTotal, &secondsTotal, &rowsTotal, &blockReadSecondsTotal, &blockWriteSecondsTotal); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			statSTatementsCallsTotal,
			prometheus.CounterValue,
			float64(callsTotal),
			user, datname, queryid,
		)
		ch <- prometheus.MustNewConstMetric(
			statStatementsSecondsTotal,
			prometheus.CounterValue,
			secondsTotal,
			user, datname, queryid,
		)
		ch <- prometheus.MustNewConstMetric(
			statStatementsRowsTotal,
			prometheus.CounterValue,
			float64(rowsTotal),
			user, datname, queryid,
		)
		ch <- prometheus.MustNewConstMetric(
			statStatementsBlockReadSecondsTotal,
			prometheus.CounterValue,
			blockReadSecondsTotal,
			user, datname, queryid,
		)
		ch <- prometheus.MustNewConstMetric(
			statStatementsBlockWriteSecondsTotal,
			prometheus.CounterValue,
			blockWriteSecondsTotal,
			user, datname, queryid,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
