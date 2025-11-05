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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestPGStatArchiverLagCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"last_archived_wal", "current_lsn"}
	// WAL file 000000010000000000000001 = LSN 16777216 (1 segment * 16MB)
	// Current LSN 0/2000000 = LSN 33554432 (hex 2000000 = decimal 33554432)
	// Lag = 33554432 - 16777216 = 16777216 bytes
	rows := sqlmock.NewRows(columns).
		AddRow("000000010000000000000001", "0/2000000")
	mock.ExpectQuery(sanitizeQuery(statArchiverLagQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatArchiverLagCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatArchiverLagCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 16777216, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestPGStatArchiverLagCollectorNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &Instance{db: db}

	columns := []string{"last_archived_wal", "current_lsn"}
	rows := sqlmock.NewRows(columns)
	mock.ExpectQuery(sanitizeQuery(statArchiverLagQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatArchiverLagCollector{}

		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatArchiverLagCollector.Update: %s", err)
		}
	}()

	expected := []MetricResult{
		{labels: labelMap{}, value: 0, metricType: dto.MetricType_GAUGE},
	}

	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

func TestParseLSNFromWalFile(t *testing.T) {
	tests := []struct {
		name     string
		walFile  string
		expected LSN
		wantErr  bool
	}{
		{
			name:     "first segment",
			walFile:  "000000010000000000000001",
			expected: 16777216, // 1 * 16MB
			wantErr:  false,
		},
		{
			name:     "second segment",
			walFile:  "000000010000000000000002",
			expected: 33554432, // 2 * 16MB
			wantErr:  false,
		},
		{
			name:     "second log ID, first segment",
			walFile:  "000000010000000100000000",
			expected: 4294967296, // 256 * 16MB
			wantErr:  false,
		},
		{
			name:     "WAL file with .history suffix",
			walFile:  "000000010000000000000001.history",
			expected: 16777216, // 1 * 16MB (suffix stripped)
			wantErr:  false,
		},
		{
			name:     "WAL file with .backup suffix",
			walFile:  "000000010000000000000001.00000028.backup",
			expected: 16777216, // 1 * 16MB (suffix stripped)
			wantErr:  false,
		},
		{
			name:     "WAL file with .partial suffix",
			walFile:  "000000010000000000000002.partial",
			expected: 33554432, // 2 * 16MB (suffix stripped)
			wantErr:  false,
		},
		{
			name:     "invalid length",
			walFile:  "00000001000000000000001",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid hex character",
			walFile:  "00000001000000000000000G",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLSNFromWalFile(tt.walFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLSNFromWalFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseLSNFromWalFile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLSNFromLSNString(t *testing.T) {
	tests := []struct {
		name     string
		lsnStr   string
		expected LSN
		wantErr  bool
	}{
		{
			name:     "simple LSN",
			lsnStr:   "0/1000000",
			expected: 16777216, // hex 1000000 = decimal 16777216
			wantErr:  false,
		},
		{
			name:     "another LSN",
			lsnStr:   "0/2000000",
			expected: 33554432, // hex 2000000 = decimal 33554432
			wantErr:  false,
		},
		{
			name:     "LSN with high part",
			lsnStr:   "1/0",
			expected: 4294967296, // 1 << 32
			wantErr:  false,
		},
		{
			name:     "invalid format",
			lsnStr:   "1000000",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty parts",
			lsnStr:   "/",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLSNFromLSNString(tt.lsnStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLSNFromLSNString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseLSNFromLSNString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBytesBetweenLSN(t *testing.T) {
	tests := []struct {
		name        string
		currentLSN  LSN
		archivedLSN LSN
		expected    LSN
	}{
		{
			name:        "normal case",
			currentLSN:  100,
			archivedLSN: 50,
			expected:    50,
		},
		{
			name:        "same LSN",
			currentLSN:  100,
			archivedLSN: 100,
			expected:    0,
		},
		{
			name:        "current less than archived (wraparound)",
			currentLSN:  50,
			archivedLSN: 100,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesBetweenLSN(tt.currentLSN, tt.archivedLSN)
			if got != tt.expected {
				t.Errorf("bytesBetweenLSN() = %v, want %v", got, tt.expected)
			}
		})
	}
}
