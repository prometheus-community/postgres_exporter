package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// Query the pg_settings view containing runtime variables
func querySettings(ch chan<- prometheus.Metric, db *sql.DB) error {
	log.Debugln("Querying pg_setting view")

	// pg_settings docs: https://www.postgresql.org/docs/current/static/view-pg-settings.html
	//
	// NOTE: If you add more vartypes here, you must update the supported
	// types in normaliseUnit() below
	query := "SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real');"

	rows, err := db.Query(query)
	if err != nil {
		return errors.New(fmt.Sprintln("Error running query on database: ", namespace, err))
	}
	defer rows.Close()

	for rows.Next() {
		s := &pgSetting{}
		err = rows.Scan(&s.name, &s.setting, &s.unit, &s.shortDesc, &s.vartype)
		if err != nil {
			return errors.New(fmt.Sprintln("Error retrieving rows:", namespace, err))
		}

		ch <- s.metric()
	}

	return nil
}

// pgSetting is represents a PostgreSQL runtime variable as returned by the
// pg_settings view.
type pgSetting struct {
	name, setting, unit, shortDesc, vartype string
}

func (s *pgSetting) metric() prometheus.Metric {
	var (
		err       error
		name      = strings.Replace(s.name, ".", "_", -1)
		unit      = s.unit
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

	desc := newDesc(subsystem, name, shortDesc)
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val)
}

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
	case "kB", "MB", "GB", "TB", "8kB", "16MB":
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
	case "16MB":
		val *= math.Pow(2, 24)
	}

	return
}
