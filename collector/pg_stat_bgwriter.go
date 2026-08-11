// Copyright 2021 The Prometheus Authors
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

func init() {
	registerCollector(bgWriterSubsystem, NewPGStatBGWriterCollector)
}

type PGStatBGWriterCollector struct {
}

func NewPGStatBGWriterCollector(collectorConfig) (Collector, error) {
	return &PGStatBGWriterCollector{}, nil
}

var (
	statBGWriterCheckpointsTimedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_timed_total"),
		"Number of scheduled checkpoints that have been performed",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_req_total"),
		"Number of requested checkpoints that have been performed",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsReqTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_write_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsSyncTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_sync_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterBuffersCheckpointDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_checkpoint_total"),
		"Number of buffers written during checkpoints",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterBuffersCleanDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_clean_total"),
		"Number of buffers written by the background writer",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterMaxwrittenCleanDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "maxwritten_clean_total"),
		"Number of times the background writer stopped a cleaning scan because it had written too many buffers",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterBuffersBackendDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_total"),
		"Number of buffers written directly by a backend",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterBuffersBackendFsyncDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_fsync_total"),
		"Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterBuffersAllocDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_alloc_total"),
		"Number of buffers allocated",
		[]string{},
		prometheus.Labels{},
	)
	statBGWriterStatsResetDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "stats_reset_total"),
		"Time at which these statistics were last reset",
		[]string{},
		prometheus.Labels{},
	)

	statBGWriterQueryBefore17 = `SELECT
		checkpoints_timed
		,checkpoints_req
		,checkpoint_write_time
		,checkpoint_sync_time
		,buffers_checkpoint
		,buffers_clean
		,maxwritten_clean
		,buffers_backend
		,buffers_backend_fsync
		,buffers_alloc
		,stats_reset
	FROM pg_stat_bgwriter;`

	statBGWriterQueryAfter17 = `SELECT
		buffers_clean
		,maxwritten_clean
		,buffers_alloc
		,stats_reset
	FROM pg_stat_bgwriter;`
)

func (PGStatBGWriterCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if instance.version.GE(semver.MustParse("17.0.0")) {
		db := instance.getDB()
		row := db.QueryRowContext(ctx, statBGWriterQueryAfter17)

		var bc, mwc, ba sql.NullInt64
		var sr sql.NullTime

		err := row.Scan(&bc, &mwc, &ba, &sr)
		if err != nil {
			return err
		}

		bcMetric := int64CounterValue(bc, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersCleanDesc,
			prometheus.CounterValue,
			bcMetric,
		)
		mwcMetric := int64CounterValue(mwc, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterMaxwrittenCleanDesc,
			prometheus.CounterValue,
			mwcMetric,
		)
		baMetric := int64CounterValue(ba, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersAllocDesc,
			prometheus.CounterValue,
			baMetric,
		)
		srMetric := 0.0
		if sr.Valid {
			srMetric = float64(sr.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statBGWriterStatsResetDesc,
			prometheus.CounterValue,
			srMetric,
		)
	} else {
		db := instance.getDB()
		row := db.QueryRowContext(ctx, statBGWriterQueryBefore17)

		var cpt, cpr, bcp, bc, mwc, bb, bbf, ba sql.NullInt64
		var cpwt, cpst sql.NullFloat64
		var sr sql.NullTime

		err := row.Scan(&cpt, &cpr, &cpwt, &cpst, &bcp, &bc, &mwc, &bb, &bbf, &ba, &sr)
		if err != nil {
			return err
		}

		cptMetric := int64CounterValue(cpt, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterCheckpointsTimedDesc,
			prometheus.CounterValue,
			cptMetric,
		)
		cprMetric := int64CounterValue(cpr, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterCheckpointsReqDesc,
			prometheus.CounterValue,
			cprMetric,
		)
		cpwtMetric := 0.0
		if cpwt.Valid {
			cpwtMetric = float64(cpwt.Float64)
		}
		ch <- prometheus.MustNewConstMetric(
			statBGWriterCheckpointsReqTimeDesc,
			prometheus.CounterValue,
			cpwtMetric,
		)
		cpstMetric := 0.0
		if cpst.Valid {
			cpstMetric = float64(cpst.Float64)
		}
		ch <- prometheus.MustNewConstMetric(
			statBGWriterCheckpointsSyncTimeDesc,
			prometheus.CounterValue,
			cpstMetric,
		)
		bcpMetric := int64CounterValue(bcp, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersCheckpointDesc,
			prometheus.CounterValue,
			bcpMetric,
		)
		bcMetric := int64CounterValue(bc, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersCleanDesc,
			prometheus.CounterValue,
			bcMetric,
		)
		mwcMetric := int64CounterValue(mwc, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterMaxwrittenCleanDesc,
			prometheus.CounterValue,
			mwcMetric,
		)
		bbMetric := int64CounterValue(bb, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersBackendDesc,
			prometheus.CounterValue,
			bbMetric,
		)
		bbfMetric := int64CounterValue(bbf, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersBackendFsyncDesc,
			prometheus.CounterValue,
			bbfMetric,
		)
		baMetric := int64CounterValue(ba, instance.wrapLargeCounters)
		ch <- prometheus.MustNewConstMetric(
			statBGWriterBuffersAllocDesc,
			prometheus.CounterValue,
			baMetric,
		)
		srMetric := 0.0
		if sr.Valid {
			srMetric = float64(sr.Time.Unix())
		}
		ch <- prometheus.MustNewConstMetric(
			statBGWriterStatsResetDesc,
			prometheus.CounterValue,
			srMetric,
		)
	}

	return nil
}
