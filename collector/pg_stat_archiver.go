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

const statArchiverSubsystem = "stat_archiver"

func init() {
	registerCollector(statArchiverSubsystem, defaultEnabled, NewPGStatArchiverCollector)
}

type PGStatArchiverCollector struct{}

func NewPGStatArchiverCollector(collectorConfig) (Collector, error) {
	return &PGStatArchiverCollector{}, nil
}

var (
	statArchiverArchivedCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statArchiverSubsystem, "archived_count"),
		"Number of WAL files that have been successfully archived",
		[]string{},
		prometheus.Labels{},
	)
	statArchiverFailedCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statArchiverSubsystem, "failed_count"),
		"Number of failed attempts for archiving WAL files",
		[]string{},
		prometheus.Labels{},
	)
	statArchiverLastArchiveAgeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statArchiverSubsystem, "last_archive_age"),
		"Time in seconds since last WAL segment was successfully archived",
		[]string{},
		prometheus.Labels{},
	)

	statArchiverQuery = `SELECT
		archived_count,
		failed_count,
		extract(epoch from now() - last_archived_time) AS last_archive_age
	FROM pg_stat_archiver;`
)

func (PGStatArchiverCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if instance.version.LT(semver.MustParse("9.4.0")) {
		return nil
	}

	db := instance.getDB()
	row := db.QueryRowContext(ctx, statArchiverQuery)

	var archivedCount, failedCount sql.NullInt64
	var lastArchiveAge sql.NullFloat64

	if err := row.Scan(&archivedCount, &failedCount, &lastArchiveAge); err != nil {
		return err
	}

	if archivedCount.Valid {
		ch <- prometheus.MustNewConstMetric(
			statArchiverArchivedCountDesc,
			prometheus.CounterValue,
			float64(archivedCount.Int64),
		)
	}

	if failedCount.Valid {
		ch <- prometheus.MustNewConstMetric(
			statArchiverFailedCountDesc,
			prometheus.CounterValue,
			float64(failedCount.Int64),
		)
	}

	if lastArchiveAge.Valid {
		ch <- prometheus.MustNewConstMetric(
			statArchiverLastArchiveAgeDesc,
			prometheus.GaugeValue,
			lastArchiveAge.Float64,
		)
	}

	return nil
}
