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
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const databaseWraparoundSubsystem = "database_wraparound"

func init() {
	registerCollector(databaseWraparoundSubsystem, defaultDisabled, NewPGDatabaseWraparoundCollector)
}

type PGDatabaseWraparoundCollector struct {
	log log.Logger
}

func NewPGDatabaseWraparoundCollector(config collectorConfig) (Collector, error) {
	return &PGDatabaseWraparoundCollector{log: config.logger}, nil
}

var (
	databaseWraparoundAgeDatfrozenxid = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, databaseWraparoundSubsystem, "age_datfrozenxid_seconds"),
		"Age of the oldest transaction ID that has not been frozen.",
		[]string{"datname"},
		prometheus.Labels{},
	)
	databaseWraparoundAgeDatminmxid = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, databaseWraparoundSubsystem, "age_datminmxid_seconds"),
		"Age of the oldest multi-transaction ID that has been replaced with a transaction ID.",
		[]string{"datname"},
		prometheus.Labels{},
	)

	databaseWraparoundQuery = `
	SELECT
		datname,
		age(d.datfrozenxid) as age_datfrozenxid,
		mxid_age(d.datminmxid) as age_datminmxid
	FROM
		pg_catalog.pg_database d
	WHERE
		d.datallowconn
	`
)

func (c *PGDatabaseWraparoundCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		databaseWraparoundQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname sql.NullString
		var ageDatfrozenxid, ageDatminmxid sql.NullFloat64

		if err := rows.Scan(&datname, &ageDatfrozenxid, &ageDatminmxid); err != nil {
			return err
		}

		if !datname.Valid {
			level.Debug(c.log).Log("msg", "Skipping database with NULL name")
			continue
		}
		if !ageDatfrozenxid.Valid {
			level.Debug(c.log).Log("msg", "Skipping stat emission with NULL age_datfrozenxid")
			continue
		}
		if !ageDatminmxid.Valid {
			level.Debug(c.log).Log("msg", "Skipping stat emission with NULL age_datminmxid")
			continue
		}

		ageDatfrozenxidMetric := ageDatfrozenxid.Float64

		ch <- prometheus.MustNewConstMetric(
			databaseWraparoundAgeDatfrozenxid,
			prometheus.GaugeValue,
			ageDatfrozenxidMetric, datname.String,
		)

		ageDatminmxidMetric := ageDatminmxid.Float64
		ch <- prometheus.MustNewConstMetric(
			databaseWraparoundAgeDatminmxid,
			prometheus.GaugeValue,
			ageDatminmxidMetric, datname.String,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
