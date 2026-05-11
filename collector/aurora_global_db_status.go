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

// auroraGlobalDBSubsystem exposes metrics from Aurora PostgreSQL's
// aurora_global_db_status() function. Disabled by default; only useful for
// Aurora Global Database deployments.
const auroraGlobalDBSubsystem = "aurora_global_db"

func init() {
	registerCollector("aurora_global_db_status", defaultDisabled, NewAuroraGlobalDBStatusCollector)
}

type AuroraGlobalDBStatusCollector struct{}

func NewAuroraGlobalDBStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraGlobalDBStatusCollector{}, nil
}

var (
	auroraGlobalDurabilityLagDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBSubsystem, "durability_lag_milliseconds"),
		"Storage lag vs primary cluster in milliseconds (-1 for the primary cluster).",
		[]string{"aws_region"}, nil,
	)
	auroraGlobalRPOLagDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBSubsystem, "rpo_lag_milliseconds"),
		"Recovery point objective lag in milliseconds (-1 for the primary cluster).",
		[]string{"aws_region"}, nil,
	)

	auroraGlobalDBStatusQuery = `SELECT
		aws_region,
		durability_lag_in_msec,
		rpo_lag_in_msec
	FROM aurora_global_db_status()`
)

func (c AuroraGlobalDBStatusCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraGlobalDBStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var awsRegion string
		var durabilityLag, rpoLag sql.NullFloat64

		if err := rows.Scan(&awsRegion, &durabilityLag, &rpoLag); err != nil {
			return err
		}

		if durabilityLag.Valid {
			ch <- prometheus.MustNewConstMetric(auroraGlobalDurabilityLagDesc, prometheus.GaugeValue, durabilityLag.Float64, awsRegion)
		}
		if rpoLag.Valid {
			ch <- prometheus.MustNewConstMetric(auroraGlobalRPOLagDesc, prometheus.GaugeValue, rpoLag.Float64, awsRegion)
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
