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
	"database/sql"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	labelCats                      = []string{"upstream_host", "slot_name", "status"}
	statWalReceiverReceiveStartLsn = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "receive_start_lsn"),
		"First write-ahead log location used when WAL receiver is started represented as a decimal",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverReceiveStartTli = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "receive_start_tli"),
		"First timeline number used when WAL receiver is started",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverFlushedLSN = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "flushed_lsn"),
		"Last write-ahead log location already received and flushed to disk, the initial value of this field being the first log location used when WAL receiver is started represented as a decimal",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverReceivedTli = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "received_tli"),
		"Timeline number of last write-ahead log location received and flushed to disk",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverLastMsgSendTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "last_msg_send_time"),
		"Send time of last message received from origin WAL sender",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverLastMsgReceiptTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "last_msg_receipt_time"),
		"Send time of last message received from origin WAL sender",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverLatestEndLsn = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "latest_end_lsn"),
		"Last write-ahead log location reported to origin WAL sender as integer",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverLatestEndTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "latest_end_time"),
		"Time of last write-ahead log location reported to origin WAL sender",
		labelCats,
		prometheus.Labels{},
	)
	statWalReceiverUpstreamNode = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statWalReceiverSubsystem, "upstream_node"),
		"Node ID of the upstream node",
		labelCats,
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

	pgStatWalReceiverQueryTemplate = `
	SELECT
		trim(both '''' from substring(conninfo from 'host=([^ ]*)')) as upstream_host,
		slot_name,
		status,
		(receive_start_lsn- '0/0') %% (2^52)::bigint as receive_start_lsn,
		%s
receive_start_tli,
		received_tli,
		extract(epoch from last_msg_send_time) as last_msg_send_time,
		extract(epoch from last_msg_receipt_time) as last_msg_receipt_time,
		(latest_end_lsn - '0/0') %% (2^52)::bigint as latest_end_lsn,
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

	hasFlushedLSN := hasFlushedLSNRows.Next()
	var query string
	if hasFlushedLSN {
		query = fmt.Sprintf(pgStatWalReceiverQueryTemplate, "(flushed_lsn - '0/0') % (2^52)::bigint as flushed_lsn,\n")
	} else {
		query = fmt.Sprintf(pgStatWalReceiverQueryTemplate, "")
	}

	hasFlushedLSNRows.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var upstreamHost, slotName, status sql.NullString
		var receiveStartLsn, receiveStartTli, flushedLsn, receivedTli, latestEndLsn, upstreamNode sql.NullInt64
		var lastMsgSendTime, lastMsgReceiptTime, latestEndTime sql.NullFloat64

		if hasFlushedLSN {
			if err := rows.Scan(&upstreamHost, &slotName, &status, &receiveStartLsn, &receiveStartTli, &flushedLsn, &receivedTli, &lastMsgSendTime, &lastMsgReceiptTime, &latestEndLsn, &latestEndTime, &upstreamNode); err != nil {
				return err
			}
		} else {
			if err := rows.Scan(&upstreamHost, &slotName, &status, &receiveStartLsn, &receiveStartTli, &receivedTli, &lastMsgSendTime, &lastMsgReceiptTime, &latestEndLsn, &latestEndTime, &upstreamNode); err != nil {
				return err
			}
		}
		if !upstreamHost.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because upstream host is null")
			continue
		}

		if !slotName.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because slotname host is null")
			continue
		}

		if !status.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because status is null")
			continue
		}
		labels := []string{upstreamHost.String, slotName.String, status.String}

		if !receiveStartLsn.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because receive_start_lsn is null")
			continue
		}
		if !receiveStartTli.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because receive_start_tli is null")
			continue
		}
		if hasFlushedLSN && !flushedLsn.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because flushed_lsn is null")
			continue
		}
		if !receivedTli.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because received_tli is null")
			continue
		}
		if !lastMsgSendTime.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because last_msg_send_time is null")
			continue
		}
		if !lastMsgReceiptTime.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because last_msg_receipt_time is null")
			continue
		}
		if !latestEndLsn.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because latest_end_lsn is null")
			continue
		}
		if !latestEndTime.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because latest_end_time is null")
			continue
		}
		if !upstreamNode.Valid {
			level.Debug(c.log).Log("msg", "Skipping wal receiver stats because upstream_node is null")
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceiveStartLsn,
			prometheus.CounterValue,
			float64(receiveStartLsn.Int64),
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceiveStartTli,
			prometheus.GaugeValue,
			float64(receiveStartTli.Int64),
			labels...)

		if hasFlushedLSN {
			ch <- prometheus.MustNewConstMetric(
				statWalReceiverFlushedLSN,
				prometheus.CounterValue,
				float64(flushedLsn.Int64),
				labels...)
		}

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverReceivedTli,
			prometheus.GaugeValue,
			float64(receivedTli.Int64),
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLastMsgSendTime,
			prometheus.CounterValue,
			float64(lastMsgSendTime.Float64),
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLastMsgReceiptTime,
			prometheus.CounterValue,
			float64(lastMsgReceiptTime.Float64),
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLatestEndLsn,
			prometheus.CounterValue,
			float64(latestEndLsn.Int64),
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverLatestEndTime,
			prometheus.CounterValue,
			latestEndTime.Float64,
			labels...)

		ch <- prometheus.MustNewConstMetric(
			statWalReceiverUpstreamNode,
			prometheus.GaugeValue,
			float64(upstreamNode.Int64),
			labels...)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
