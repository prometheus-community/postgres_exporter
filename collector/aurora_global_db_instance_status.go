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

const auroraGlobalDBInstanceStatusSubsystem = "aurora_global_db_instance_status"

func init() {
	registerCollector(auroraGlobalDBInstanceStatusSubsystem, serverScope, NewAuroraGlobalDBInstanceStatusCollector)
}

// AuroraGlobalDBInstanceStatusCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_global_db_instance_status() function: how far each
// instance, including the cross-Region replicas of an Aurora global
// database, lags behind the writer. It reports no data on servers that are
// not Aurora.
type AuroraGlobalDBInstanceStatusCollector struct{}

func NewAuroraGlobalDBInstanceStatusCollector(collectorConfig) (Collector, error) {
	return &AuroraGlobalDBInstanceStatusCollector{}, nil
}

var (
	auroraGlobalDBInstanceStatusVisibilityLag = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraGlobalDBInstanceStatusSubsystem, "visibility_lag_seconds"),
		"How far the instance lags behind the writer instance in seconds.",
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
		var visibilityLagMsec sql.NullFloat64

		if err := rows.Scan(&serverID, &awsRegion, &visibilityLagMsec); err != nil {
			return err
		}

		// The writer reports NULL: it doesn't lag behind itself.
		if visibilityLagMsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraGlobalDBInstanceStatusVisibilityLag, prometheus.GaugeValue, visibilityLagMsec.Float64/1e3, serverID, awsRegion)
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
