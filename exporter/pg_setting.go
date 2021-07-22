// Copyright 2021 The Prometheus Authors
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

package exporter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Query the pg_settings view containing runtime variables
func querySettings(ch chan<- prometheus.Metric, server *Server) error {
	level.Debug(server.Logger).Log("msg", "Querying pg_setting view", "server", server)

	// pg_settings docs: https://www.postgresql.org/docs/current/static/view-pg-settings.html
	//
	// NOTE: If you add more vartypes here, you must update the supported
	// types in normaliseUnit() below
	query := "SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real');"

	rows, err := server.db.Query(query)
	if err != nil {
		return fmt.Errorf("Error running query on database %q: %s %v", server, namespace, err)
	}
	defer rows.Close() // nolint: errcheck

	for rows.Next() {
		s := &pgSetting{}
		err = rows.Scan(&s.name, &s.setting, &s.unit, &s.shortDesc, &s.vartype)
		if err != nil {
			return fmt.Errorf("Error retrieving rows on %q: %s %v", server, namespace, err)
		}

		ch <- s.metric(server.labels)
	}

	return nil
}

// pgSetting is represents a PostgreSQL runtime variable as returned by the
// pg_settings view.
type pgSetting struct {
	name, setting, unit, shortDesc, vartype string
}

func (s *pgSetting) metric(labels prometheus.Labels) prometheus.Metric {
	var (
		err       error
		name      = strings.Replace(s.name, ".", "_", -1)
		unit      = s.unit // nolint: ineffassign
		shortDesc = s.shortDesc
		subsystem = "settings"
		val       float64
	)

	switch s.vartype {
	case "bool":
		if s.setting == "on" {
			val = 1
		}
	case "integer", "real":
		if val, unit, err = s.normaliseUnit(); err != nil {
			// Panic, since we should recognise all units
			// and don't want to silently exlude metrics
			panic(err)
		}

		if len(unit) > 0 {
			name = fmt.Sprintf("%s_%s", name, unit)
			shortDesc = fmt.Sprintf("%s [Units converted to %s.]", shortDesc, unit)
		}
	default:
		// Panic because we got a type we didn't ask for
		panic(fmt.Sprintf("Unsupported vartype %q", s.vartype))
	}

	desc := newDesc(subsystem, name, shortDesc, labels)
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val)
}

// TODO: fix linter override
// nolint: nakedret
func (s *pgSetting) normaliseUnit() (val float64, unit string, err error) {
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
	case "B", "kB", "MB", "GB", "TB", "8kB", "16kB", "32kB", "16MB", "32MB", "64MB":
		unit = "bytes"
	default:
		err = fmt.Errorf("Unknown unit for runtime variable: %q", s.unit)
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
	case "8kB":
		val *= math.Pow(2, 13)
	case "16kB":
		val *= math.Pow(2, 14)
	case "32kB":
		val *= math.Pow(2, 15)
	case "16MB":
		val *= math.Pow(2, 24)
	case "32MB":
		val *= math.Pow(2, 25)
	case "64MB":
		val *= math.Pow(2, 26)
	}

	return
}
