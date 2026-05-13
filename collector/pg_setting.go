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
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const settingsSubsystem = "settings"

func init() {
	registerCollector(settingsSubsystem, defaultEnabled, NewPGSettingsCollector)
}

type PGSettingsCollector struct {
	log *slog.Logger
}

func NewPGSettingsCollector(config collectorConfig) (Collector, error) {
	return &PGSettingsCollector{log: config.logger}, nil
}

var (
	settingUnits = []string{
		"ms", "s", "min", "h", "d",
		"B", "kB", "MB", "GB", "TB",
	}

	// pg_settings docs: https://www.postgresql.org/docs/current/static/view-pg-settings.html
	//
	// NOTE: If you add more vartypes here, you must update the supported
	// types in normaliseUnit() below.
	//
	// Settings intentionally ignored due to invalid format:
	// - `sync_commit_cancel_wait`, specific to Azure Postgres, see https://github.com/prometheus-community/postgres_exporter/issues/523
	// - `google_dataplex.max_messages`, specific to Google Cloud SQL, see https://github.com/prometheus-community/postgres_exporter/issues/1240
	pgSettingsQuery = "SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real') AND name NOT IN ('sync_commit_cancel_wait', 'google_dataplex.max_messages');"
)

// Update implements Collector and exposes PostgreSQL runtime settings.
func (c PGSettingsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx, pgSettingsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		s := &pgSetting{}
		if err := rows.Scan(&s.name, &s.setting, &s.unit, &s.shortDesc, &s.vartype); err != nil {
			return err
		}
		metric, err := s.metric()
		if err != nil {
			c.log.Warn("Error normalising unit for setting", "setting", s.name, "value", s.setting, "unit", s.unit, "error", err)
			continue
		}
		ch <- metric
	}

	return rows.Err()
}

// pgSetting represents a PostgreSQL runtime variable as returned by the
// pg_settings view.
type pgSetting struct {
	name, setting, unit, shortDesc, vartype string
}

func (s *pgSetting) metric() (prometheus.Metric, error) {
	var (
		err       error
		name      = strings.ReplaceAll(strings.ReplaceAll(s.name, ".", "_"), "-", "_")
		unit      = s.unit // nolint: ineffassign
		shortDesc = fmt.Sprintf("Server Parameter: %s", s.name)
		val       float64
	)

	switch s.vartype {
	case "bool":
		if s.setting == "on" {
			val = 1
		}
	case "integer", "real":
		if val, unit, err = s.normaliseUnit(); err != nil {
			return nil, err
		}

		if len(unit) > 0 {
			name = fmt.Sprintf("%s_%s", name, unit)
			shortDesc = fmt.Sprintf("%s [Units converted to %s.]", shortDesc, unit)
		}
	default:
		return nil, fmt.Errorf("pgsetting: unsupported vartype %q", s.vartype)
	}

	desc := prometheus.NewDesc(prometheus.BuildFQName(namespace, settingsSubsystem, name), shortDesc, nil, nil)
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val), nil
}

// Removes units from any of the setting values.
// This is mostly because of a irregularity regarding AWS RDS Aurora
// https://github.com/prometheus-community/postgres_exporter/issues/619
func (s *pgSetting) sanitizeValue() {
	for _, unit := range settingUnits {
		if strings.HasSuffix(s.setting, unit) {
			endPos := len(s.setting) - len(unit) - 1
			s.setting = s.setting[:endPos]
			return
		}
	}
}

// TODO: fix linter override
// nolint: nakedret
func (s *pgSetting) normaliseUnit() (val float64, unit string, err error) {
	s.sanitizeValue()

	val, err = strconv.ParseFloat(s.setting, 64)
	if err != nil {
		return val, unit, fmt.Errorf("Error converting setting %q value %q to float: %s", s.name, s.setting, err)
	}

	// Units defined in: https://www.postgresql.org/docs/current/static/config-setting.html
	switch s.unit {
	case "":
		return
	case "ms", "s", "min", "h", "d":
		unit = "seconds"
	case "B", "kB", "MB", "GB", "TB", "1kB", "2kB", "4kB", "8kB", "16kB", "32kB", "64kB", "16MB", "32MB", "64MB":
		unit = "bytes"
	default:
		err = fmt.Errorf("unknown unit for runtime variable: %q", s.unit)
		return
	}

	// -1 is special, don't modify the value
	if val == -1 {
		return
	}

	switch s.unit {
	case "ms":
		val /= 1000
	case "min":
		val *= 60
	case "h":
		val *= 60 * 60
	case "d":
		val *= 60 * 60 * 24
	case "kB":
		val *= math.Pow(2, 10)
	case "MB":
		val *= math.Pow(2, 20)
	case "GB":
		val *= math.Pow(2, 30)
	case "TB":
		val *= math.Pow(2, 40)
	case "1kB":
		val *= math.Pow(2, 10)
	case "2kB":
		val *= math.Pow(2, 11)
	case "4kB":
		val *= math.Pow(2, 12)
	case "8kB":
		val *= math.Pow(2, 13)
	case "16kB":
		val *= math.Pow(2, 14)
	case "32kB":
		val *= math.Pow(2, 15)
	case "64kB":
		val *= math.Pow(2, 16)
	case "16MB":
		val *= math.Pow(2, 24)
	case "32MB":
		val *= math.Pow(2, 25)
	case "64MB":
		val *= math.Pow(2, 26)
	}

	return
}
