// Copyright 2026 The Prometheus Authors
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
	"log/slog"
	"net/url"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
)

const sequenceOverflowSubsystem = "sequence_overflow"

func init() {
	registerCollector(sequenceOverflowSubsystem, defaultDisabled, NewPGSequenceOverflowCollector)
}

type PGSequenceOverflowCollector struct {
	log               *slog.Logger
	excludedDatabases []string
}

func NewPGSequenceOverflowCollector(config collectorConfig) (Collector, error) {
	return &PGSequenceOverflowCollector{
		log:               config.logger,
		excludedDatabases: config.excludeDatabases,
	}, nil
}

var (
	sequenceOverflowLastValue = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sequenceOverflowSubsystem, "last_value"),
		"Last value consumed from the sequence.",
		[]string{"datname", "schemaname", "sequence", "sequence_datatype", "owned_by", "column_datatype"},
		prometheus.Labels{},
	)
	sequenceOverflowSequenceRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sequenceOverflowSubsystem, "sequence_ratio"),
		"Ratio of the last sequence value to the maximum value of the sequence datatype (0–1).",
		[]string{"datname", "schemaname", "sequence", "sequence_datatype", "owned_by", "column_datatype"},
		prometheus.Labels{},
	)
	sequenceOverflowColumnRatio = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sequenceOverflowSubsystem, "column_ratio"),
		"Ratio of the last sequence value to the maximum value of the owning column datatype (0–1).",
		[]string{"datname", "schemaname", "sequence", "sequence_datatype", "owned_by", "column_datatype"},
		prometheus.Labels{},
	)

	sequenceOverflowDatabaseQuery = "SELECT datname FROM pg_database WHERE datallowconn AND NOT datistemplate;"

	sequenceOverflowQuery = `
SELECT
    seqs.relname AS sequence,
    format_type(s.seqtypid, NULL) AS sequence_datatype,
    CONCAT(tbls.relname, '.', attrs.attname) AS owned_by,
    format_type(attrs.atttypid, atttypmod) AS column_datatype,
    ns.nspname AS schemaname,
    COALESCE(pg_sequence_last_value(seqs.oid::regclass), 0) AS last_sequence_value,
    COALESCE(CASE format_type(s.seqtypid, NULL)
        WHEN 'smallint' THEN pg_sequence_last_value(seqs.oid::regclass) / 32767::float
        WHEN 'integer'  THEN pg_sequence_last_value(seqs.oid::regclass) / 2147483647::float
        WHEN 'bigint'   THEN pg_sequence_last_value(seqs.oid::regclass) / 9223372036854775807::float
    END, 0) AS sequence_ratio,
    COALESCE(CASE format_type(attrs.atttypid, NULL)
        WHEN 'smallint' THEN pg_sequence_last_value(seqs.oid::regclass) / 32767::float
        WHEN 'integer'  THEN pg_sequence_last_value(seqs.oid::regclass) / 2147483647::float
        WHEN 'bigint'   THEN pg_sequence_last_value(seqs.oid::regclass) / 9223372036854775807::float
    END, 0) AS column_ratio
FROM pg_depend d
JOIN pg_class AS seqs ON seqs.relkind = 'S' AND seqs.oid = d.objid
JOIN pg_class AS tbls ON tbls.relkind = 'r' AND tbls.oid = d.refobjid
JOIN pg_attribute AS attrs ON attrs.attrelid = d.refobjid AND attrs.attnum = d.refobjsubid
JOIN pg_sequence s ON s.seqrelid = seqs.oid
JOIN pg_namespace ns ON ns.oid = seqs.relnamespace
WHERE d.deptype = 'a'
  AND d.classid = 1259;
`
)

// dsnForDatabase returns a new DSN with the database name replaced.
// It updates both the URL path and the dbname query parameter (if present),
// since libpq gives precedence to the query parameter over the path.
func dsnForDatabase(dsn string, datname string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("could not parse DSN: %w", err)
	}
	u.Path = "/" + datname
	if q := u.Query(); q.Get("dbname") != "" {
		q.Set("dbname", datname)
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

// Update implements Collector and exposes sequence integer overflow metrics
// for all non-template databases on the instance.
// It is called by the Prometheus registry when collecting metrics.
// Note: pg_sequence_last_value requires USAGE privilege on each sequence,
// or pg_read_all_data. Without it the function returns NULL, which COALESCE converts
// to 0, causing all metrics to show 0 with no visible error.
func (c *PGSequenceOverflowCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	// Fetch the list of connectable, non-template databases.
	rows, err := db.QueryContext(ctx, sequenceOverflowDatabaseQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var datname sql.NullString
		if err := rows.Scan(&datname); err != nil {
			return err
		}
		if !datname.Valid {
			continue
		}
		if slices.Contains(c.excludedDatabases, datname.String) {
			continue
		}
		databases = append(databases, datname.String)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Query sequence metrics from each database individually.
	for _, datname := range databases {
		newDSN, err := dsnForDatabase(instance.dsn, datname)
		if err != nil {
			c.log.Debug("Skipping database", "datname", datname, "err", err)
			continue
		}
		dbConn, err := sql.Open("postgres", newDSN)
		if err != nil {
			c.log.Debug("Skipping database", "datname", datname, "err", err)
			continue
		}
		if err := c.updateDatabase(ctx, dbConn, datname, ch); err != nil {
			c.log.Debug("Skipping database", "datname", datname, "err", err)
		}
		dbConn.Close()
	}
	return nil
}

func (c *PGSequenceOverflowCollector) updateDatabase(ctx context.Context, db *sql.DB, datname string, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx, sequenceOverflowQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		// sequence_datatype, owned_by, and column_datatype are sourced from NOT NULL
		// catalog columns (seqtypid, relname/attname, atttypid) and cannot be NULL
		// given the JOIN structure.
		var sequence, sequenceDatatype, ownedBy, columnDatatype, schemaname sql.NullString
		var lastValue, sequenceRatio, columnRatio float64

		if err := rows.Scan(&sequence, &sequenceDatatype, &ownedBy, &columnDatatype, &schemaname, &lastValue, &sequenceRatio, &columnRatio); err != nil {
			return err
		}

		if !sequence.Valid {
			c.log.Debug("Skipping sequence with NULL name", "datname", datname)
			continue
		}

		schemanameLabel := "unknown"
		if schemaname.Valid {
			schemanameLabel = schemaname.String
		}

		ch <- prometheus.MustNewConstMetric(
			sequenceOverflowLastValue,
			prometheus.GaugeValue,
			lastValue,
			datname, schemanameLabel, sequence.String, sequenceDatatype.String, ownedBy.String, columnDatatype.String,
		)
		ch <- prometheus.MustNewConstMetric(
			sequenceOverflowSequenceRatio,
			prometheus.GaugeValue,
			sequenceRatio,
			datname, schemanameLabel, sequence.String, sequenceDatatype.String, ownedBy.String, columnDatatype.String,
		)
		ch <- prometheus.MustNewConstMetric(
			sequenceOverflowColumnRatio,
			prometheus.GaugeValue,
			columnRatio,
			datname, schemanameLabel, sequence.String, sequenceDatatype.String, ownedBy.String, columnDatatype.String,
		)
	}
	return rows.Err()
}
