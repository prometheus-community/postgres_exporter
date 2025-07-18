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

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const statCheckpointerSubsystem = "stat_checkpointer"

func init() {
	// WARNING:
	//   Disabled by default because this set of metrics is only available from Postgres 17
	registerCollector(statCheckpointerSubsystem, defaultDisabled, NewPGStatCheckpointerCollector)
}

type PGStatCheckpointerCollector struct {
	log *slog.Logger
}

func NewPGStatCheckpointerCollector(config collectorConfig) (Collector, error) {
	return &PGStatCheckpointerCollector{log: config.logger}, nil
}

var (
	statCheckpointerNumTimedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "num_timed_total"),
		"Number of scheduled checkpoints due to timeout",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerNumRequestedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "num_requested_total"),
		"Number of requested checkpoints that have been performed",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerRestartpointsTimedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "restartpoints_timed_total"),
		"Number of scheduled restartpoints due to timeout or after a failed attempt to perform it",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerRestartpointsReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "restartpoints_req_total"),
		"Number of requested restartpoints",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerRestartpointsDoneDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "restartpoints_done_total"),
		"Number of restartpoints that have been performed",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerWriteTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "write_time_total"),
		"Total amount of time that has been spent in the portion of processing checkpoints and restartpoints where files are written to disk, in milliseconds",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerSyncTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "sync_time_total"),
		"Total amount of time that has been spent in the portion of processing checkpoints and restartpoints where files are synchronized to disk, in milliseconds",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerBuffersWrittenDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "buffers_written_total"),
		"Number of buffers written during checkpoints and restartpoints",
		[]string{},
		prometheus.Labels{},
	)
	statCheckpointerStatsResetDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statCheckpointerSubsystem, "stats_reset_total"),
		"Time at which these statistics were last reset",
		[]string{},
		prometheus.Labels{},
	)

	statCheckpointerQuery = `SELECT
		num_timed
		,num_requested
		,restartpoints_timed
		,restartpoints_req
		,restartpoints_done
		,write_time
		,sync_time
		,buffers_written
		,stats_reset
	FROM pg_stat_checkpointer;`
)

func (c PGStatCheckpointerCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	before17 := instance.version.LT(semver.MustParse("17.0.0"))
	if before17 {
		c.log.Warn("pg_stat_checkpointer collector is not available on PostgreSQL < 17.0.0, skipping")
		return nil
	}

	row := db.QueryRowContext(ctx, statCheckpointerQuery)

	// num_timed           = nt  = bigint
	// num_requested       = nr  = bigint
	// restartpoints_timed = rpt = bigint
	// restartpoints_req   = rpr = bigint
	// restartpoints_done  = rpd = bigint
	// write_time          = wt  = double precision
	// sync_time           = st  = double precision
	// buffers_written     = bw  = bigint
	// stats_reset         = sr  = timestamp

	var nt, nr, rpt, rpr, rpd, bw sql.NullInt64
	var wt, st sql.NullFloat64
	var sr sql.NullTime

	err := row.Scan(&nt, &nr, &rpt, &rpr, &rpd, &wt, &st, &bw, &sr)
	if err != nil {
		return err
	}

	ntMetric := 0.0
	if nt.Valid {
		ntMetric = float64(nt.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerNumTimedDesc,
		prometheus.CounterValue,
		ntMetric,
	)

	nrMetric := 0.0
	if nr.Valid {
		nrMetric = float64(nr.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerNumRequestedDesc,
		prometheus.CounterValue,
		nrMetric,
	)

	rptMetric := 0.0
	if rpt.Valid {
		rptMetric = float64(rpt.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerRestartpointsTimedDesc,
		prometheus.CounterValue,
		rptMetric,
	)

	rprMetric := 0.0
	if rpr.Valid {
		rprMetric = float64(rpr.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerRestartpointsReqDesc,
		prometheus.CounterValue,
		rprMetric,
	)

	rpdMetric := 0.0
	if rpd.Valid {
		rpdMetric = float64(rpd.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerRestartpointsDoneDesc,
		prometheus.CounterValue,
		rpdMetric,
	)

	wtMetric := 0.0
	if wt.Valid {
		wtMetric = float64(wt.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerWriteTimeDesc,
		prometheus.CounterValue,
		wtMetric,
	)

	stMetric := 0.0
	if st.Valid {
		stMetric = float64(st.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerSyncTimeDesc,
		prometheus.CounterValue,
		stMetric,
	)

	bwMetric := 0.0
	if bw.Valid {
		bwMetric = float64(bw.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerBuffersWrittenDesc,
		prometheus.CounterValue,
		bwMetric,
	)

	srMetric := 0.0
	if sr.Valid {
		srMetric = float64(sr.Time.Unix())
	}
	ch <- prometheus.MustNewConstMetric(
		statCheckpointerStatsResetDesc,
		prometheus.CounterValue,
		srMetric,
	)

	return nil
}
