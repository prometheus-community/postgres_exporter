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

	"github.com/prometheus/client_golang/prometheus"
)

const bgWriterSubsystem = "stat_bgwriter"

func init() {
	registerCollector(bgWriterSubsystem, defaultEnabled, NewPGStatBGWriterCollector)
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
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_req_total"),
		"Number of requested checkpoints that have been performed",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsReqTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_write_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterCheckpointsSyncTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_sync_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterBuffersCheckpointDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_checkpoint_total"),
		"Number of buffers written during checkpoints",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterBuffersCleanDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_clean_total"),
		"Number of buffers written by the background writer",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterMaxwrittenCleanDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "maxwritten_clean_total"),
		"Number of times the background writer stopped a cleaning scan because it had written too many buffers",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterBuffersBackendDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_total"),
		"Number of buffers written directly by a backend",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterBuffersBackendFsyncDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_fsync_total"),
		"Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterBuffersAllocDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_alloc_total"),
		"Number of buffers allocated",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
	statBGWriterStatsResetDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "stats_reset_total"),
		"Time at which these statistics were last reset",
		[]string{"collector", "server"},
		prometheus.Labels{},
	)
)
var statBGWriter = map[string]*prometheus.Desc{
	"percona_checkpoints_timed": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_timed"),
		"Number of scheduled checkpoints that have been performed",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_checkpoints_req": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_req"),
		"Number of requested checkpoints that have been performed",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_checkpoint_write_time": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_write_time"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_checkpoint_sync_time": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_sync_time"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_buffers_checkpoint": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_checkpoint"),
		"Number of buffers written during checkpoints",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_buffers_clean": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_clean"),
		"Number of buffers written by the background writer",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_maxwritten_clean": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "maxwritten_clean"),
		"Number of times the background writer stopped a cleaning scan because it had written too many buffers",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_buffers_backend": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend"),
		"Number of buffers written directly by a backend",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_buffers_backend_fsync": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_fsync"),
		"Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_buffers_alloc": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_alloc"),
		"Number of buffers allocated",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
	"percona_stats_reset": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "stats_reset"),
		"Time at which these statistics were last reset",
		[]string{"collector", "server"},
		prometheus.Labels{},
	),
}

const statBGWriterQuery = `SELECT
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

func (PGStatBGWriterCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {

	db := instance.getDB()
	row := db.QueryRowContext(ctx,
		statBGWriterQuery)

	var cpt, cpr, bcp, bc, mwc, bb, bbf, ba sql.NullInt64
	var cpwt, cpst sql.NullFloat64
	var sr sql.NullTime

	err := row.Scan(&cpt, &cpr, &cpwt, &cpst, &bcp, &bc, &mwc, &bb, &bbf, &ba, &sr)
	if err != nil {
		return err
	}

	cptMetric := 0.0
	if cpt.Valid {
		cptMetric = float64(cpt.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsTimedDesc,
		prometheus.CounterValue,
		cptMetric,
		"exporter",
		instance.name,
	)
	cprMetric := 0.0
	if cpr.Valid {
		cprMetric = float64(cpr.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsReqDesc,
		prometheus.CounterValue,
		cprMetric,
		"exporter",
		instance.name,
	)
	cpwtMetric := 0.0
	if cpwt.Valid {
		cpwtMetric = float64(cpwt.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsReqTimeDesc,
		prometheus.CounterValue,
		cpwtMetric,
		"exporter",
		instance.name,
	)
	cpstMetric := 0.0
	if cpst.Valid {
		cpstMetric = float64(cpst.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsSyncTimeDesc,
		prometheus.CounterValue,
		cpstMetric,
		"exporter",
		instance.name,
	)
	bcpMetric := 0.0
	if bcp.Valid {
		bcpMetric = float64(bcp.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersCheckpointDesc,
		prometheus.CounterValue,
		bcpMetric,
		"exporter",
		instance.name,
	)
	bcMetric := 0.0
	if bc.Valid {
		bcMetric = float64(bc.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersCleanDesc,
		prometheus.CounterValue,
		bcMetric,
		"exporter",
		instance.name,
	)
	mwcMetric := 0.0
	if mwc.Valid {
		mwcMetric = float64(mwc.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterMaxwrittenCleanDesc,
		prometheus.CounterValue,
		mwcMetric,
		"exporter",
		instance.name,
	)
	bbMetric := 0.0
	if bb.Valid {
		bbMetric = float64(bb.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersBackendDesc,
		prometheus.CounterValue,
		bbMetric,
		"exporter",
		instance.name,
	)
	bbfMetric := 0.0
	if bbf.Valid {
		bbfMetric = float64(bbf.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersBackendFsyncDesc,
		prometheus.CounterValue,
		bbfMetric,
		"exporter",
		instance.name,
	)
	baMetric := 0.0
	if ba.Valid {
		baMetric = float64(ba.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersAllocDesc,
		prometheus.CounterValue,
		baMetric,
		"exporter",
		instance.name,
	)
	srMetric := 0.0
	if sr.Valid {
		srMetric = float64(sr.Time.Unix())
	}
	ch <- prometheus.MustNewConstMetric(
		statBGWriterStatsResetDesc,
		prometheus.CounterValue,
		srMetric,
		"exporter",
		instance.name,
	)

	// TODO: analyze metrics below, why do we duplicate them?

	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_checkpoints_timed"],
		prometheus.CounterValue,
		cptMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_checkpoints_req"],
		prometheus.CounterValue,
		cprMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_checkpoint_write_time"],
		prometheus.CounterValue,
		cpwtMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_checkpoint_sync_time"],
		prometheus.CounterValue,
		cpstMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_buffers_checkpoint"],
		prometheus.CounterValue,
		bcpMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_buffers_clean"],
		prometheus.CounterValue,
		bcMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_maxwritten_clean"],
		prometheus.CounterValue,
		mwcMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_buffers_backend"],
		prometheus.CounterValue,
		bbMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_buffers_backend_fsync"],
		prometheus.CounterValue,
		bbfMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_buffers_alloc"],
		prometheus.CounterValue,
		baMetric,
		"exporter",
		instance.name,
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["percona_stats_reset"],
		prometheus.CounterValue,
		srMetric,
		"exporter",
		instance.name,
	)

	return nil
}
