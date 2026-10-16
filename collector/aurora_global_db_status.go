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

	"github.com/prometheus/client_golang/prometheus"
)

const auroraGlobalDBStatusSubsystem = "aurora_global_db_status"

func init() {
	registerCollector(auroraGlobalDBStatusSubsystem, serverScope, NewAuroraGlobalDBStatusCollector)
}

// AuroraGlobalDBStatusCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_global_db_status() function: how far each secondary
// cluster of an Aurora global database lags behind the primary cluster. It
// reports no data on servers that are not Aurora.
type AuroraGlobalDBStatusCollector struct{}

func NewAuroraGlobalDBStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraGlobalDBStatusCollector{}, nil
}

var (
	auroraGlobalDBStatusDurabilityLag = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBStatusSubsystem, "durability_lag_seconds"),
		"Storage lag of the secondary cluster behind the primary cluster in seconds.",
		[]string{"aws_region"}, nil,
	)
	auroraGlobalDBStatusRPOLag = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBStatusSubsystem, "rpo_lag_seconds"),
		"Recovery point objective lag of the secondary cluster in seconds: how long a commit on the primary cluster takes to be stored on it.",
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
		var durabilityLagMsec, rpoLagMsec sql.NullFloat64

		if err := rows.Scan(&awsRegion, &durabilityLagMsec, &rpoLagMsec); err != nil {
			return err
		}

		// The primary cluster reports -1: lag doesn't apply to it.
		if durabilityLagMsec.Valid && durabilityLagMsec.Float64 >= 0 {
			ch <- prometheus.MustNewConstMetric(auroraGlobalDBStatusDurabilityLag, prometheus.GaugeValue, durabilityLagMsec.Float64/1e3, awsRegion)
		}
		if rpoLagMsec.Valid && rpoLagMsec.Float64 >= 0 {
			ch <- prometheus.MustNewConstMetric(auroraGlobalDBStatusRPOLag, prometheus.GaugeValue, rpoLagMsec.Float64/1e3, awsRegion)
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
