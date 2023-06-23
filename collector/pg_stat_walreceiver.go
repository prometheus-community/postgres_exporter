// Copyright 2023 The Prometheus Authors
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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(statWalReceiverSubsystem, defaultDisabled, NewPGStatWalReceiverCollector)
}

type PGStatWalReceiverCollector struct {
	log log.Logger
}

const statWalReceiverSubsystem = "stat_wal_receiver"

func NewPGStatWalReceiverCollector(config collectorConfig) (Collector, error) {
	return &PGStatWalReceiverCollector{log: config.logger}, nil
}

var (
	statWalReceiverStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "status"),
		"Activity status of the WAL receiver process",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverReceiveStartLsn = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "receive_start_lsn"),
		"First write-ahead log location used when WAL receiver is started represented as a decimal",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverReceiveStartTli = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "receive_start_tli"),
		"First timeline number used when WAL receiver is started",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverFlushedLSN = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "flushed_lsn"),
		"Last write-ahead log location already received and flushed to disk, the initial value of this field being the first log location used when WAL receiver is started represented as a decimal",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverReceivedTli = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "received_tli"),
		"Timeline number of last write-ahead log location received and flushed to disk",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverLastMsgSendTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "last_msg_send_time"),
		"Send time of last message received from origin WAL sender",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverLastMsgReceiptTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "last_msg_receipt_time"),
		"Send time of last message received from origin WAL sender",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverLatestEndLsn = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "latest_end_lsn"),
		"Last write-ahead log location reported to origin WAL sender as integer",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverLatestEndTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "latest_end_time"),
		"Time of last write-ahead log location reported to origin WAL sender",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)
	statWalReceiverUpstreamNode = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "upstream_node"),
		"Node ID of the upstream node",
		[]string{"upstream_host", "slot_name"},
		prometheus.Labels{},
	)

	pgStatWalColumnQuery = `
	SELECT
		column_name
	FROM information_schema.columns
	WHERE
		table_name = 'pg_stat_wal_receiver' and
		column_name = 'flushed_lsn'
	`

	pgStatWalReceiverQueryWithNoFlushedLSN = `
	SELECT
		trim(both '''' from substring(conninfo from 'host=([^ ]*)')) as upstream_host,
		slot_name,
		case status
			when 'stopped' then 0
			when 'starting' then 1
			when 'streaming' then 2
			when 'waiting' then 3
			when 'restarting' then 4
			when 'stopping' then 5 else -1
		end as status,
		(receive_start_lsn- '0/0') % (2^52)::bigint as receive_start_lsn,
		receive_start_tli,
		received_tli,
		extract(epoch from last_msg_send_time) as last_msg_send_time,
		extract(epoch from last_msg_receipt_time) as last_msg_receipt_time,
		(latest_end_lsn - '0/0') % (2^52)::bigint as latest_end_lsn,
		extract(epoch from latest_end_time) as latest_end_time,
		substring(slot_name from 'repmgr_slot_([0-9]*)') as upstream_node
	FROM pg_catalog.pg_stat_wal_receiver
	`

	pgStatWalReceiverQueryWithFlushedLSN = `
	SELECT
		trim(both '''' from substring(conninfo from 'host=([^ ]*)')) as upstream_host,
		slot_name,
		case status
			when 'stopped' then 0
			when 'starting' then 1
			when 'streaming' then 2
			when 'waiting' then 3
			when 'restarting' then 4
			when 'stopping' then 5 else -1
		end as status,
		(receive_start_lsn- '0/0') % (2^52)::bigint as receive_start_lsn,
		receive_start_tli,
		(flushed_lsn- '0/0') % (2^52)::bigint as flushed_lsn,
		received_tli,
		extract(epoch from last_msg_send_time) as last_msg_send_time,
		extract(epoch from last_msg_receipt_time) as last_msg_receipt_time,
		(latest_end_lsn - '0/0') % (2^52)::bigint as latest_end_lsn,
		extract(epoch from latest_end_time) as latest_end_time,
		substring(slot_name from 'repmgr_slot_([0-9]*)') as upstream_node
	FROM pg_catalog.pg_stat_wal_receiver
	`
)

func (c *PGStatWalReceiverCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	hasFlushedLSNRows, err := db.QueryContext(ctx, pgStatWalColumnQuery)
	if err != nil {
		return err
	}

	defer hasFlushedLSNRows.Close()
	hasFlushedLSN := hasFlushedLSNRows.Next()
	var query string
	if hasFlushedLSN {
		query = pgStatWalReceiverQueryWithFlushedLSN
	} else {
		query = pgStatWalReceiverQueryWithNoFlushedLSN
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var upstreamHost string
		var slotName string
		var status int
		var receiveStartLsn int64
		var receiveStartTli int64
		var flushedLsn int64
		var receivedTli int64
		var lastMsgSendTime float64
		var lastMsgReceiptTime float64
		var latestEndLsn int64
		var latestEndTime float64
		var upstreamNode int

		if hasFlushedLSN {
			if err := rows.Scan(&upstreamHost, &slotName, &status, &receiveStartLsn, &receiveStartTli, &flushedLsn, &receivedTli, &lastMsgSendTime, &lastMsgReceiptTime, &latestEndLsn, &latestEndTime, &upstreamNode); err != nil {
				return err
			}
		} else {
			if err := rows.Scan(&upstreamHost, &slotName, &status, &receiveStartLsn, &receiveStartTli, &receivedTli, &lastMsgSendTime, &lastMsgReceiptTime, &latestEndLsn, &latestEndTime, &upstreamNode); err != nil {
				return err
			}
		}

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverStatus,
			prometheus.GaugeValue,
			float64(status),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceiveStartLsn,
			prometheus.CounterValue,
			float64(receiveStartLsn),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceiveStartTli,
			prometheus.GaugeValue,
			float64(receiveStartTli),
			upstreamHost, slotName)

		if hasFlushedLSN {
			ch <- prometheus.MustNewConstMetric(
				statWalReceiverFlushedLSN,
				prometheus.CounterValue,
				float64(flushedLsn),
				upstreamHost, slotName)
		}

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceivedTli,
			prometheus.GaugeValue,
			float64(receivedTli),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLastMsgSendTime,
			prometheus.CounterValue,
			float64(lastMsgSendTime),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLastMsgReceiptTime,
			prometheus.CounterValue,
			float64(lastMsgReceiptTime),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLatestEndLsn,
			prometheus.CounterValue,
			float64(latestEndLsn),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLatestEndTime,
			prometheus.CounterValue,
			float64(latestEndTime),
			upstreamHost, slotName)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverUpstreamNode,
			prometheus.GaugeValue,
			float64(upstreamNode),
			upstreamHost, slotName)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
