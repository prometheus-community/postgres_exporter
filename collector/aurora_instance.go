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

	"github.com/prometheus/client_golang/prometheus"
)

const auroraInstanceSubsystem = "aurora_instance"

func init() {
	registerCollector(auroraInstanceSubsystem, serverScope, NewAuroraInstanceCollector)
}

// AuroraInstanceCollector exposes the Amazon Aurora PostgreSQL instance,
// which the exporter is connected to, and its Aurora version, from
// aurora_db_instance_identifier() and aurora_version(). It reports no data on
// servers that are not Aurora, or on Aurora versions without
// aurora_db_instance_identifier().
type AuroraInstanceCollector struct{}

func NewAuroraInstanceCollector(collectorConfig) (Collector, error) {
	return &AuroraInstanceCollector{}, nil
}

var (
	auroraInstanceInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraInstanceSubsystem, "info"),
		"The Aurora instance the exporter is connected to and its Aurora version, with value 1.",
		[]string{"server_id", "version"}, nil,
	)

	auroraInstanceQuery = `SELECT aurora_db_instance_identifier(), aurora_version()`
)

func (c AuroraInstanceCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	db := instance.getDB()

	// aurora_db_instance_identifier() is only in Aurora PostgreSQL 13.7+,
	// 12.11+, 11.16+ and 10.21+.
	exists, err := functionExists(ctx, db, "aurora_db_instance_identifier")
	if err != nil {
		return err
	}
	if !exists {
		return ErrNoData
	}

	var serverID, version string
	if err := db.QueryRowContext(ctx, auroraInstanceQuery).Scan(&serverID, &version); err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(auroraInstanceInfo, prometheus.GaugeValue, 1, serverID, version)
	return nil
}
