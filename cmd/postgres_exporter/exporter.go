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

package postgres_exporter

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
)

var logger = log.NewNopLogger()

// Exporter collects Postgres metrics. It implements prometheus.Collector.
type Exporter struct {
	ClusterID    string
	TotalScrapes int64

	RdsMetrics RDSMetricsAPI
	Server     ServerAPI
}

// ExporterOpt configures Exporter.
type ExporterOpt func(*Exporter)

// Cluster ID.
func ClusterID(s string) ExporterOpt {
	return func(e *Exporter) {
		e.ClusterID = s
	}
}

// RDS Metrics.
func RdsMetrics(s RDSMetricsAPI) ExporterOpt {
	return func(e *Exporter) {
		e.RdsMetrics = s
	}
}

// Server Instance.
func ServerInstance(s ServerAPI) ExporterOpt {
	return func(e *Exporter) {
		e.Server = s
	}
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
func NewExporter(opts ...ExporterOpt) *Exporter {
	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)

	//    fmtlogger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	//    logger :=
	//    })

	e := &Exporter{}

	for _, opt := range opts {
		opt(e)
	}

	e.TotalScrapes = 0

	return e
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	err := e.Scrape(ch)
	if err != nil {
		level.Error(logger).Log("msg", err)
		panic(e)
	}
}

func newDesc(subsystem, name, help string, labels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, subsystem, name),
		help, nil, labels,
	)
}

func (e *Exporter) Scrape(ch chan<- prometheus.Metric) error {
	rdsCurrentCapacity, err := e.RdsMetrics.RdsCurrentCapacity(e.ClusterID)
	if err != nil {
		return fmt.Errorf("error check rds status: %v", err)
	}

	rdsDatabaseConnections, err := e.RdsMetrics.RdsCurrentConnections(e.ClusterID)
	if err != nil {
		return fmt.Errorf("error check rds status: %v", err)
	}

	if rdsCurrentCapacity == 0 || rdsDatabaseConnections == 0 {
		if err = level.Info(logger).Log("msg", fmt.Sprintf("database is not available, nothing to do - rdsCapacity: %d rdsConnections: %d", rdsCurrentCapacity, rdsDatabaseConnections)); err != nil {
			return err
		}
		return nil
	}
	level.Info(logger).Log("msg", fmt.Sprintf("database is up (capacity %d) and with connections(%d), collecting data", rdsCurrentCapacity, rdsDatabaseConnections))

	e.TotalScrapes++

	err = e.Server.Open()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("error to open database connection: %v", err))
		return err
	}

	err = e.Server.Scrape(ch, float64(e.TotalScrapes), float64(rdsDatabaseConnections), float64(rdsCurrentCapacity))
	if err != nil {
		return err
	}

	return e.Server.Close()
}
