// Copyright The Prometheus Authors
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
	"github.com/prometheus/client_golang/prometheus"
)

const statActivitySubsystem = "stat_activity"

func init() {
	registerCollector(statActivitySubsystem, defaultEnabled, NewPGStatActivityCollector)
}

type PGStatActivityCollector struct{}

func NewPGStatActivityCollector(collectorConfig) (Collector, error) {
	return &PGStatActivityCollector{}, nil
}

var (
	statActivityLabels = []string{
		"datname",
		"state",
		"usename",
		"application_name",
		"backend_type",
		"wait_event_type",
		"wait_event",
	}
	statActivityCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivitySubsystem, "count"),
		"number of connections in this state",
		statActivityLabels,
		prometheus.Labels{},
	)
	statActivityMaxTxDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statActivitySubsystem, "max_tx_duration"),
		"max duration in seconds any active transaction has been running",
		statActivityLabels,
		prometheus.Labels{},
	)

	statActivityQuery = `
			SELECT
				pg_database.datname,
				tmp.state,
				tmp2.usename,
				tmp2.application_name,
				tmp2.backend_type,
				tmp2.wait_event_type,
				tmp2.wait_event,
				COALESCE(count,0) as count,
				COALESCE(max_tx_duration,0) as max_tx_duration
			FROM
				(
				  VALUES ('active'),
				  		 ('idle'),
				  		 ('idle in transaction'),
				  		 ('idle in transaction (aborted)'),
				  		 ('fastpath function call'),
				  		 ('disabled')
				) AS tmp(state) CROSS JOIN pg_database
			LEFT JOIN
			(
				SELECT
					datname,
					state,
					usename,
					application_name,
					backend_type,
					wait_event_type,
					wait_event,
					count(*) AS count,
					MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
				FROM pg_stat_activity
				WHERE pid <> pg_backend_pid()
				GROUP BY datname,state,usename,application_name,backend_type,wait_event_type,wait_event) AS tmp2
				ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname
			`

	statActivityQueryBefore92 = `
			SELECT
				datname,
				'unknown' AS state,
				usename,
				application_name,
				'' AS backend_type,
				'' AS wait_event_type,
				'' AS wait_event,
				COALESCE(count(*),0) AS count,
				COALESCE(MAX(EXTRACT(EPOCH FROM now() - xact_start))::float,0) AS max_tx_duration
			FROM pg_stat_activity
			WHERE procpid <> pg_backend_pid()
			GROUP BY datname,usename,application_name
			`
)

func (PGStatActivityCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	query := statActivityQuery
	if instance.version.LT(semver.MustParse("9.2.0")) {
		query = statActivityQueryBefore92
	}

	db := instance.getDB()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname, state, usename, applicationName, backendType, waitEventType, waitEvent sql.NullString
		var count, maxTxDuration sql.NullFloat64
		if err := rows.Scan(
			&datname,
			&state,
			&usename,
			&applicationName,
			&backendType,
			&waitEventType,
			&waitEvent,
			&count,
			&maxTxDuration,
		); err != nil {
			return err
		}

		labels := []string{
			stringValue(datname),
			stringValue(state),
			stringValue(usename),
			stringValue(applicationName),
			stringValue(backendType),
			stringValue(waitEventType),
			stringValue(waitEvent),
		}

		countMetric := 0.0
		if count.Valid {
			countMetric = count.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statActivityCountDesc,
			prometheus.GaugeValue,
			countMetric,
			labels...,
		)

		maxTxDurationMetric := 0.0
		if maxTxDuration.Valid {
			maxTxDurationMetric = maxTxDuration.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			statActivityMaxTxDurationDesc,
			prometheus.GaugeValue,
			maxTxDurationMetric,
			labels...,
		)
	}

	return rows.Err()
}

func stringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}
