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

func init() {
	registerCollector(statioUserIndexesSubsystem, databaseScope, NewPGStatioUserIndexesCollector)
}

type PGStatioUserIndexesCollector struct {
	log *slog.Logger
	// includeDatname is set when this collector may run concurrently against
	// more than one database in the same scrape (WithDatabaseDiscovery), the
	// only case where two rows can otherwise collide in the registry with
	// identical schemaname/relname/indexrelname. Gating the label on that
	// keeps installs that never opted into discovery from seeing a new label
	// appear on this metric after an upgrade.
	includeDatname bool

	idxBlksRead *prometheus.Desc
	idxBlksHit  *prometheus.Desc
}

func NewPGStatioUserIndexesCollector(config collectorConfig) (Collector, error) {
	labels := []string{"schemaname", "relname", "indexrelname"}
	if config.databaseDiscoveryEnabled {
		labels = []string{"datname", "schemaname", "relname", "indexrelname"}
	}
	return &PGStatioUserIndexesCollector{
		log:            config.logger,
		includeDatname: config.databaseDiscoveryEnabled,
		idxBlksRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_read_total"),
			"Number of disk blocks read from this index",
			labels,
			prometheus.Labels{},
		),
		idxBlksHit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statioUserIndexesSubsystem, "idx_blks_hit_total"),
			"Number of buffer hits in this index",
			labels,
			prometheus.Labels{},
		),
	}, nil
}

const statioUserIndexesQuery = `
	SELECT
		current_database() datname,
		schemaname,
		relname,
		indexrelname,
		idx_blks_read,
		idx_blks_hit
	FROM pg_statio_user_indexes
	`

func (c *PGStatioUserIndexesCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		statioUserIndexesQuery)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		// datname comes from current_database(), which Postgres guarantees is
		// never NULL. Unlike the other columns below, it must not fall back to
		// a shared sentinel on an invalid value: this label is what keeps rows
		// from different, concurrently-scraped databases from colliding in the
		// registry, so it needs to always be the real, distinct database name.
		var datname string
		var schemaname, relname, indexrelname sql.NullString
		var idxBlksRead, idxBlksHit sql.NullInt64

		if err := rows.Scan(&datname, &schemaname, &relname, &indexrelname, &idxBlksRead, &idxBlksHit); err != nil {
			return err
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
		labels := []string{schemanameLabel, relnameLabel, indexrelnameLabel}
		if c.includeDatname {
			labels = append([]string{datname}, labels...)
		}

		idxBlksReadMetric := int64CounterValue(idxBlksRead, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			c.idxBlksRead,
			prometheus.CounterValue,
			idxBlksReadMetric,
			labels...,
		)

		idxBlksHitMetric := int64CounterValue(idxBlksHit, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			c.idxBlksHit,
			prometheus.CounterValue,
			idxBlksHitMetric,
			labels...,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
