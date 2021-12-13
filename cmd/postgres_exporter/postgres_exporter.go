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

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects Postgres metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn                    string
	iamRoleArn             string
	tenantID               string
	duration               prometheus.Gauge
	error                  prometheus.Gauge
	psqlUp                 prometheus.Gauge
	totalScrapes           prometheus.Counter
	rdsDatabaseConnections prometheus.Gauge
	rdsCurrentCapacity     prometheus.Gauge
}

// ExporterOpt configures Exporter.
type ExporterOpt func(*Exporter)

// AWS role to assume.
func IamRoleArn(s string) ExporterOpt {
	return func(e *Exporter) {
		e.iamRoleArn = s
	}
}

// Tenant ID.
func TenantID(s string) ExporterOpt {
	return func(e *Exporter) {
		e.tenantID = s
	}
}

// Regex used to get the "short-version" from the postgres version field.
var versionRegex = regexp.MustCompile(`^\w+ ((\d+)(\.\d+)?(\.\d+)?)`)
var lowestSupportedVersion = semver.MustParse("9.1.0")

// Parses the version of postgres into the short version string we can use to
// match behaviors.
func parseVersion(versionString string) (semver.Version, error) {
	submatches := versionRegex.FindStringSubmatch(versionString)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}
	return semver.Version{},
		errors.New(fmt.Sprintln("Could not find a postgres version in string:", versionString))
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
func NewExporter(dsn string, opts ...ExporterOpt) *Exporter {
	e := &Exporter{
		dsn: dsn,
	}

	for _, opt := range opts {
		opt(e)
	}

	e.setupInternalMetrics()

	return e
}

func (e *Exporter) setupInternalMetrics() {
	e.duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: exporter,
		Name:      "last_scrape_duration_seconds",
		Help:      "Duration of the last scrape of metrics from PostgresSQL.",
	})
	e.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: exporter,
		Name:      "scrapes_total",
		Help:      "Total number of times PostgresSQL was scraped for metrics.",
	})
	e.error = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: exporter,
		Name:      "last_scrape_error",
		Help:      "Whether the last scrape of metrics from PostgreSQL resulted in an error (1 for error, 0 for success).",
	})
	e.psqlUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "up",
		Help:      "Whether the last scrape of metrics from PostgreSQL was able to connect to the server (1 for yes, 0 for no).",
	})
	e.rdsCurrentCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "rds_current_capacity",
		Help:      "Current Aurora capacity units",
	})
	e.rdsDatabaseConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "rds_database_connections",
		Help:      "Current Aurora database connections",
	})
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	ch <- e.psqlUp
	ch <- e.rdsCurrentCapacity
	ch <- e.rdsDatabaseConnections
}

func newDesc(subsystem, name, help string, labels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, labels,
	)
}

func checkPostgresVersion(db *sql.DB, server string) (semver.Version, string, error) {
	level.Debug(logger).Log("msg", "Querying PostgreSQL version", "server", server)
	versionRow := db.QueryRow("SELECT version();")
	var versionString string
	err := versionRow.Scan(&versionString)
	if err != nil {
		return semver.Version{}, "", fmt.Errorf("Error scanning version string on %q: %v", server, err)
	}
	semanticVersion, err := parseVersion(versionString)
	if err != nil {
		return semver.Version{}, "", fmt.Errorf("Error parsing version string on %q: %v", server, err)
	}

	return semanticVersion, versionString, nil
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	rdsCapacity, err := e.rdsCapacity()
	if err != nil {
		log.Panicf("error check rds status: %w", err)
	}
	rdsConnections, err := e.rdsConnections()
	if err != nil {
		log.Panicf("error check rds status: %w", err)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("rdsCapacity: %d - rdsConnections: %d", rdsCapacity, rdsConnections))

	if rdsCapacity == 0 || rdsConnections == 0 {
		level.Info(logger).Log("msg", "database is not available - nothing to do")
		e.psqlUp.Set(0)
		e.error.Set(0)
		return
	}
	level.Info(logger).Log("msg", "database is up and with connections, collecting data")

	e.psqlUp.Set(1)
	e.error.Set(1)
	e.rdsDatabaseConnections.Set(float64(rdsConnections))
	e.rdsCurrentCapacity.Set(float64(rdsCapacity))

	server, err := NewServer(e.dsn)
	if err != nil {
		log.Panicf("error to open database connection: %w", err)
	}
	semanticVersion, versionString, err := checkPostgresVersion(server.db, server.String())
	versionDesc := prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, staticLabelName),
		"Version string as reported by postgres", []string{"version", "short_version"}, server.labels)

	ch <- prometheus.MustNewConstMetric(versionDesc,
		prometheus.UntypedValue, 1, versionString, semanticVersion.String())

	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
		server.Close()
	}(time.Now())

	e.totalScrapes.Inc()

	server.Scrape(ch)

}
