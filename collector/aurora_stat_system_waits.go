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

const auroraStatSystemWaitsSubsystem = "aurora_stat_system_waits"

func init() {
	registerCollector(auroraStatSystemWaitsSubsystem, serverScope, NewAuroraStatSystemWaitsCollector)
}

// AuroraStatSystemWaitsCollector exposes metrics from Amazon Aurora
// PostgreSQL's aurora_stat_system_waits() function: how many times the
// instance waited for each wait event and how long, with the names of the
// wait events and their types from aurora_stat_wait_event() and
// aurora_stat_wait_type(). It reports no data on servers that are not Aurora.
type AuroraStatSystemWaitsCollector struct{}

func NewAuroraStatSystemWaitsCollector(collectorConfig) (Collector, error) {
	return &AuroraStatSystemWaitsCollector{}, nil
}

var (
	auroraStatSystemWaitsWaits = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatSystemWaitsSubsystem, "total"),
		"Number of times the instance waited for the wait event.",
		[]string{"wait_event_type", "wait_event", "wait_event_id"}, nil,
	)
	auroraStatSystemWaitsWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, auroraStatSystemWaitsSubsystem, "seconds_total"),
		"Time the instance spent waiting for the wait event in seconds.",
		[]string{"wait_event_type", "wait_event", "wait_event_id"}, nil,
	)

	// aurora_stat_system_waits() returns only events that have occurred.
	// Events without a name, if any, keep their IDs. The event ID is a label
	// of its own as well: Aurora PostgreSQL 16.15 gives two different IPC
	// events the same name, LogicalSyncStateChange.
	auroraStatSystemWaitsQuery = `SELECT
		COALESCE(t.type_name, w.type_id::text) AS wait_event_type,
		COALESCE(e.event_name, w.event_id::text) AS wait_event,
		w.event_id::text AS wait_event_id,
		w.waits,
		w.wait_time
	FROM aurora_stat_system_waits() w
	LEFT JOIN aurora_stat_wait_type() t ON t.type_id = w.type_id
	LEFT JOIN aurora_stat_wait_event() e ON e.type_id = w.type_id AND e.event_id = w.event_id`
)

func (c AuroraStatSystemWaitsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	if !instance.isAurora {
		return ErrNoData
	}
	// AWS doesn't document since which versions these functions exist.
	for _, name := range []string{"aurora_stat_system_waits", "aurora_stat_wait_type", "aurora_stat_wait_event"} {
		exists, err := functionExists(ctx, instance.getDB(), name)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNoData
		}
	}
	rows, err := instance.getDB().QueryContext(ctx, auroraStatSystemWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		found = true

		var eventType, event, eventID string
		var waits, waitTimeUsec sql.NullInt64
		if err := rows.Scan(&eventType, &event, &eventID, &waits, &waitTimeUsec); err != nil {
			return err
		}

		if waits.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatSystemWaitsWaits, prometheus.CounterValue, int64CounterValue(waits, instance.wrapLargeCounters), eventType, event, eventID)
		}
		if waitTimeUsec.Valid {
			ch <- prometheus.MustNewConstMetric(auroraStatSystemWaitsWaitTime, prometheus.CounterValue, float64(waitTimeUsec.Int64)/1e6, eventType, event, eventID)
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
