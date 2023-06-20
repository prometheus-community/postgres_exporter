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
	registerCollector("replication", defaultEnabled, NewPGStatWalReceiverCollector)
}

type PGArchiverCollector struct {
	log log.Logger
}

const archiverSubsystem = "archiver"

func NewPGArchiverCollector(collectorConfig) (Collector, error) {
	return &PGArchiverCollector{}, nil
}

var (
	pgArchiverPendingWalCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, archiverSubsystem, "pending_wal_count"),
		"Number of WAL files waiting to be archived",
		[]string{}, prometheus.Labels{},
	)

	pgArchiverQuery = `
	WITH
      current_wal_file AS (
         SELECT CASE WHEN NOT pg_is_in_recovery() THEN pg_walfile_name(pg_current_wal_insert_lsn()) ELSE NULL END pg_walfile_name
      ),
      current_wal AS (
        SELECT
          ('x'||substring(pg_walfile_name,9,8))::bit(32)::int log,
          ('x'||substring(pg_walfile_name,17,8))::bit(32)::int seg,
          pg_walfile_name
        FROM current_wal_file
      ),
      archive_wal AS(
        SELECT
          ('x'||substring(last_archived_wal,9,8))::bit(32)::int log,
          ('x'||substring(last_archived_wal,17,8))::bit(32)::int seg,
          last_archived_wal
        FROM pg_stat_archiver
      )
    SELECT coalesce(((cw.log - aw.log) * 256) + (cw.seg-aw.seg),'NaN'::float) as pending_wal_count FROM current_wal cw, archive_wal aw
	`
)

func (c *PGArchiverCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	row := db.QueryRowContext(ctx,
		pgArchiverQuery)
	var pendingWalCount float64
	err := row.Scan(&pendingWalCount)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		pgArchiverPendingWalCount,
		prometheus.GaugeValue,
		pendingWalCount,
	)
	return nil
}
