// Copyright 2025 PlanetScale Inc.
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
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const archiverLagSubsystem = "stat_archiver"

func init() {
	registerCollector(archiverLagSubsystem, defaultEnabled, NewPGStatArchiverLagCollector)
}

type PGStatArchiverLagCollector struct{}

func NewPGStatArchiverLagCollector(collectorConfig) (Collector, error) {
	return &PGStatArchiverLagCollector{}, nil
}

var (
	statArchiverLagBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, archiverLagSubsystem, "lag_bytes"),
		"Archiver lag in bytes (difference between current WAL position and last archived WAL)",
		[]string{},
		prometheus.Labels{},
	)

	statArchiverLagQuery = `
    SELECT
      last_archived_wal,
      CASE WHEN pg_is_in_recovery() THEN NULL ELSE pg_current_wal_lsn() END AS current_lsn
    FROM pg_stat_archiver
    WHERE last_archived_wal IS NOT NULL
      AND last_archived_wal != ''
  `
)

// LSN represents a PostgreSQL Log Sequence Number, a 64-bit unsigned integer
// representing a byte position in the WAL.
type LSN uint64

const (
	// walSegmentSizeBytes is the size of a WAL segment in bytes (16MB)
	walSegmentSizeBytes = 16 * 1024 * 1024 // 16777216
	// segmentsPerLogID is the number of segments per log ID (256)
	segmentsPerLogID = 256
)

// parseLSNFromWalFile parses a WAL file name (e.g., "000000010000000000000001") and returns
// the LSN position in bytes. The WAL file format is:
// - Positions 1-8: timeline ID (8 hex chars)
// - Positions 9-16: log ID (8 hex chars)
// - Positions 17-24: segment ID (8 hex chars)
// Returns LSN = logID * 256 segments * 16MB + segmentID * 16MB
//
// Handles WAL files with suffixes like .backup, .history, .partial by stripping them first.
func parseLSNFromWalFile(walFile string) (LSN, error) {
	// Strip suffix if present (e.g., .backup, .history, .partial)
	if idx := strings.Index(walFile, "."); idx != -1 {
		walFile = walFile[:idx]
	}

	if len(walFile) != 24 {
		return 0, fmt.Errorf("WAL file name must be exactly 24 hex chars, got %d: %q", len(walFile), walFile)
	}

	// Validate all characters are hex
	for i, r := range walFile {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') && (r < 'a' || r > 'f') {
			return 0, fmt.Errorf("WAL file name contains invalid hex character at position %d: %q", i+1, string(r))
		}
	}

	// Extract log ID (positions 9-16, 0-indexed: 8-15)
	logIDHex := walFile[8:16]
	logID, err := parseHexUint32(logIDHex)
	if err != nil {
		return 0, fmt.Errorf("parse log ID from %q: %w", logIDHex, err)
	}

	// Extract segment ID (positions 17-24, 0-indexed: 16-23)
	segIDHex := walFile[16:24]
	segID, err := parseHexUint32(segIDHex)
	if err != nil {
		return 0, fmt.Errorf("parse segment ID from %q: %w", segIDHex, err)
	}

	// Calculate LSN: logID * 256 segments * 16MB + segmentID * 16MB
	lsnBytes := LSN(logID)*segmentsPerLogID*walSegmentSizeBytes + LSN(segID)*walSegmentSizeBytes
	return lsnBytes, nil
}

// parseLSNFromLSNString parses a PostgreSQL LSN string (e.g., "0/12345678") and returns
// the LSN position in bytes. PostgreSQL LSNs represent byte positions in the WAL.
// The format is "high/low" where both are hex numbers representing a 64-bit byte offset:
// LSN = (high << 32) | low
func parseLSNFromLSNString(lsnStr string) (LSN, error) {
	parts := strings.Split(lsnStr, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("LSN string must be in format 'high/low', got: %q", lsnStr)
	}

	highStr, lowStr := parts[0], parts[1]
	if highStr == "" || lowStr == "" {
		return 0, fmt.Errorf("LSN string parts cannot be empty: %q", lsnStr)
	}

	high, err := strconv.ParseUint(highStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse high part %q of LSN string %q: %w", highStr, lsnStr, err)
	}

	low, err := strconv.ParseUint(lowStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse low part %q of LSN string %q: %w", lowStr, lsnStr, err)
	}

	// LSN = (high << 32) | low
	return LSN(high<<32 | low), nil
}

// parseHexUint32 parses a hex string (8 hex chars) and returns a uint32.
func parseHexUint32(hexStr string) (uint32, error) {
	if len(hexStr) != 8 {
		return 0, fmt.Errorf("hex string must be exactly 8 chars, got %d: %q", len(hexStr), hexStr)
	}

	val, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("parse hex %q: %w", hexStr, err)
	}
	return uint32(val), nil
}

// bytesBetweenLSN calculates the difference in bytes between two LSN positions.
// Returns the difference, clamped to 0 if currentLSN is less than archivedLSN
// (which can happen during wraparound or timeline switches).
func bytesBetweenLSN(currentLSN, archivedLSN LSN) LSN {
	if currentLSN < archivedLSN {
		return 0
	}
	return currentLSN - archivedLSN
}

func (PGStatArchiverLagCollector) Update(ctx context.Context, instance *Instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx, statArchiverLagQuery)

	var lastArchivedWal sql.NullString
	var currentLSN sql.NullString

	err := row.Scan(&lastArchivedWal, &currentLSN)
	if err != nil {
		// If no rows found (no WAL segments archived yet), return 0 lag
		if err == sql.ErrNoRows {
			ch <- prometheus.MustNewConstMetric(
				statArchiverLagBytesDesc,
				prometheus.GaugeValue,
				0,
			)
			return nil
		}
		return err
	}

	// If either value is null, return 0 lag
	if !lastArchivedWal.Valid || !currentLSN.Valid {
		ch <- prometheus.MustNewConstMetric(
			statArchiverLagBytesDesc,
			prometheus.GaugeValue,
			0,
		)
		return nil
	}

	// Parse LSN from WAL file name
	archivedLSN, err := parseLSNFromWalFile(lastArchivedWal.String)
	if err != nil {
		return fmt.Errorf("parse archived WAL file %q: %w", lastArchivedWal.String, err)
	}

	// Parse current LSN from PostgreSQL LSN string format
	currentLSNBytes, err := parseLSNFromLSNString(currentLSN.String)
	if err != nil {
		return fmt.Errorf("parse current LSN %q: %w", currentLSN.String, err)
	}

	// Calculate lag
	lagBytes := bytesBetweenLSN(currentLSNBytes, archivedLSN)

	ch <- prometheus.MustNewConstMetric(
		statArchiverLagBytesDesc,
		prometheus.GaugeValue,
		float64(lagBytes),
	)

	return nil
}
