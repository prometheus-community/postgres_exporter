// Copyright 2025 The Prometheus Authors
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

// aurora_global_db_instance_status() exposes per-instance lag across regions
// in an Aurora Global Database. Disabled by default; only useful on Aurora.
const auroraGlobalDBInstanceSubsystem = "aurora_global_db_instance"

func init() {
	registerCollector("aurora_global_db_instance_status", defaultDisabled, NewAuroraGlobalDBInstanceStatusCollector)
}

type AuroraGlobalDBInstanceStatusCollector struct{}

func NewAuroraGlobalDBInstanceStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraGlobalDBInstanceStatusCollector{}, nil
}

var (
	auroraGlobalDBInstanceVisibilityLagDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBInstanceSubsystem, "visibility_lag_milliseconds"),
		"How far this DB instance is lagging behind the writer DB instance in milliseconds. NULL for the writer.",
		[]string{"server_id", "aws_region"}, nil,
	)

	auroraGlobalDBInstanceStatusQuery = `SELECT
		server_id,
		aws_region,
		visibility_lag_in_msec
	FROM aurora_global_db_instance_status()`
)

func (c AuroraGlobalDBInstanceStatusCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraGlobalDBInstanceStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var serverID, awsRegion string
		var visibilityLag sql.NullFloat64

		if err := rows.Scan(&serverID, &awsRegion, &visibilityLag); err != nil {
			return err
		}

		if visibilityLag.Valid {
			ch <- prometheus.MustNewConstMetric(auroraGlobalDBInstanceVisibilityLagDesc, prometheus.GaugeValue, visibilityLag.Float64, serverID, awsRegion)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if !found {
		return ErrNoData
	}
	return nil
}
