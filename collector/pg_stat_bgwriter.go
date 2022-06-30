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
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("bgwriter", defaultEnabled, NewPGStatBGWriterCollector)
}

type PGStatBGWriterCollector struct {
}

func NewPGStatBGWriterCollector(logger log.Logger) (Collector, error) {
	return &PGStatBGWriterCollector{}, nil
}

const bgWriterSubsystem = "stat_bgwriter"

var statBGWriter = map[string]*prometheus.Desc{
	"checkpoints_timed": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_timed_total"),
		"Number of scheduled checkpoints that have been performed",
		[]string{"server"},
		prometheus.Labels{},
	),
	"checkpoints_req": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoints_req_total"),
		"Number of requested checkpoints that have been performed",
		[]string{"server"},
		prometheus.Labels{},
	),
	"checkpoint_write_time": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_write_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds",
		[]string{"server"},
		prometheus.Labels{},
	),
	"checkpoint_sync_time": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "checkpoint_sync_time_total"),
		"Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds",
		[]string{"server"},
		prometheus.Labels{},
	),
	"buffers_checkpoint": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_checkpoint_total"),
		"Number of buffers written during checkpoints",
		[]string{"server"},
		prometheus.Labels{},
	),
	"buffers_clean": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_clean_total"),
		"Number of buffers written by the background writer",
		[]string{"server"},
		prometheus.Labels{},
	),
	"maxwritten_clean": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "maxwritten_clean_total"),
		"Number of times the background writer stopped a cleaning scan because it had written too many buffers",
		[]string{"server"},
		prometheus.Labels{},
	),
	"buffers_backend": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_total"),
		"Number of buffers written directly by a backend",
		[]string{"server"},
		prometheus.Labels{},
	),
	"buffers_backend_fsync": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_backend_fsync_total"),
		"Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)",
		[]string{"server"},
		prometheus.Labels{},
	),
	"buffers_alloc": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "buffers_alloc_total"),
		"Number of buffers allocated",
		[]string{"server"},
		prometheus.Labels{},
	),
	"stats_reset": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, bgWriterSubsystem, "stats_reset_total"),
		"Time at which these statistics were last reset",
		[]string{"server"},
		prometheus.Labels{},
	),
}

func (PGStatBGWriterCollector) Update(ctx context.Context, server *server, ch chan<- prometheus.Metric) error {
	db, err := server.GetDB()
	if err != nil {
		return err
	}

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
	var cpwt int
	var cpst int
	var bcp int
	var bc int
	var mwc int
	var bb int
	var bbf int
	var ba int
	var sr time.Time

	err = row.Scan(&cpt, &cpr, &cpwt, &cpst, &bcp, &bc, &mwc, &bb, &bbf, &ba, &sr)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		statBGWriter["checkpoints_timed"],
		prometheus.CounterValue,
		float64(cpt),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["checkpoints_req"],
		prometheus.CounterValue,
		float64(cpr),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["checkpoint_write_time"],
		prometheus.CounterValue,
		float64(cpwt),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["checkpoint_sync_time"],
		prometheus.CounterValue,
		float64(cpst),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["buffers_checkpoint"],
		prometheus.CounterValue,
		float64(bcp),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["buffers_clean"],
		prometheus.CounterValue,
		float64(bc),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["maxwritten_clean"],
		prometheus.CounterValue,
		float64(mwc),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["buffers_backend"],
		prometheus.CounterValue,
		float64(bb),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["buffers_backend_fsync"],
		prometheus.CounterValue,
		float64(bbf),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["buffers_alloc"],
		prometheus.CounterValue,
		float64(ba),
		server.GetName(),
	)
	ch <- prometheus.MustNewConstMetric(
		statBGWriter["stats_reset"],
		prometheus.CounterValue,
		float64(sr.Unix()),
		server.GetName(),
	)

	return nil
}
