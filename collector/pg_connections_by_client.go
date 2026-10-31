// Copyright 2024 The Prometheus Authors
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

const connectionsByClientSubsystem = "connections_by_client"

func init() {
	registerCollector(connectionsByClientSubsystem, defaultEnabled, NewPGConnectionsByClientCollector)
}

type PGConnectionByClientCollector struct {
	log *slog.Logger
}

func NewPGConnectionsByClientCollector(config collectorConfig) (Collector, error) {
	return &PGConnectionByClientCollector{
		log: config.logger,
	}, nil
}

var (
	pgClientCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			connectionsByClientSubsystem,
			"count",
		),
		"Number of clients",
		[]string{"name"}, nil,
	)

	pgConnectionsByClientQuery = `
	SELECT
		count(*) as count,
		client_hostname
	FROM pg_stat_activity
	WHERE client_hostname is not null
	GROUP BY client_hostname;
	`
)

func (c PGConnectionByClientCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	rows, err := db.QueryContext(ctx,
		pgConnectionsByClientQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	var count sql.NullInt64
	var name sql.NullString

	for rows.Next() {
		if err := rows.Scan(&count, &name); err != nil {
			return err
		}

		if !name.Valid {
			continue
		}

		countMetric := 0.0
		if count.Valid {
			countMetric = float64(count.Int64)
		}

		ch <- prometheus.MustNewConstMetric(
			pgClientCountDesc,
			prometheus.GaugeValue,
			countMetric,
			name.String,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
