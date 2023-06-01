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
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("bgwriter", defaultEnabled, NewPGStatBGWriterCollector)
}

type PGStatBGWriterCollector struct {
}

func NewPGStatBGWriterCollector(collectorConfig) (Collector, error) {
	return &PGStatBGWriterCollector{}, nil
}

const bgWriterSubsystem = "stat_bgwriter"

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
)

func (PGStatBGWriterCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	row := db.QueryRowContext(ctx,
		`SELECT
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
		 FROM pg_stat_bgwriter;`)

	var cpt int
	var cpr int
	var cpwt float64
	var cpst float64
	var bcp int
	var bc int
	var mwc int
	var bb int
	var bbf int
	var ba int
	var sr time.Time

	err := row.Scan(&cpt, &cpr, &cpwt, &cpst, &bcp, &bc, &mwc, &bb, &bbf, &ba, &sr)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsTimedDesc,
		prometheus.CounterValue,
		float64(cpt),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsReqDesc,
		prometheus.CounterValue,
		float64(cpr),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsReqTimeDesc,
		prometheus.CounterValue,
		float64(cpwt),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterCheckpointsSyncTimeDesc,
		prometheus.CounterValue,
		float64(cpst),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersCheckpointDesc,
		prometheus.CounterValue,
		float64(bcp),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersCleanDesc,
		prometheus.CounterValue,
		float64(bc),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterMaxwrittenCleanDesc,
		prometheus.CounterValue,
		float64(mwc),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersBackendDesc,
		prometheus.CounterValue,
		float64(bb),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersBackendFsyncDesc,
		prometheus.CounterValue,
		float64(bbf),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterBuffersAllocDesc,
		prometheus.CounterValue,
		float64(ba),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriterStatsResetDesc,
		prometheus.CounterValue,
		float64(sr.Unix()),
	)

	return nil
}
