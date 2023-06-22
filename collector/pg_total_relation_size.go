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

const totalRelationSizeSubsystem = "total_relation_size"

func init() {
	registerCollector(statioUserTableSubsystem, defaultEnabled, NewPGStatIOUserTablesCollector)
}

type PGTotalRelationSizeCollector struct {
	log log.Logger
}

func NewPGTotalRelationSizeCollector(config collectorConfig) (Collector, error) {
	return &PGTotalRelationSizeCollector{log: config.logger}, nil
}

var (
	totalRelationSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, totalRelationSizeSubsystem, "bytes"),
		"total disk space usage for the specified table and associated indexes",
		[]string{"schemaname", "relname"},
		prometheus.Labels{},
	)

	totalRelationSizeQuery = `
		SELECT
			relnamespace::regnamespace as schemaname,
			relname as relname,
			pg_total_relation_size(oid) bytes
		FROM pg_class
		WHERE relkind = 'r';
	`
)

func (PGTotalRelationSizeCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		totalRelationSizeQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaname, relname string
		var bytes float64

		if err := rows.Scan(&schemaname, &relname, &bytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			totalRelationSizeBytes,
			prometheus.GaugeValue,
			bytes,
			schemaname, relname,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
