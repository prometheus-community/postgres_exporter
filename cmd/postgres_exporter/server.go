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

package postgres_exporter

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Server describes a connection to Postgres.
// Also it contains metrics map and query overrides.
type Server struct {
	Db  *sql.DB
	Dsn string

	// Currently active query overrides
	mappingMtx sync.RWMutex

	NsMap       NamespaceMetricsAPI
	SettMetrics SettingsMetricsAPI
}

// compile-time check that type implements interface.
var _ ServerAPI = (*Server)(nil)

// ServerOpt configures a server.
type ServerOpt func(*Server)

// Open establishes a new connection using DSN.
func (s *Server) Open() error {

	db, err := sql.Open("postgres", s.Dsn)
	if err != nil {
		return err
	}
	s.Db = db

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	level.Info(logger).Log("msg", "Established new database connection")

	return nil
}

// Close disconnects from Postgres.
func (s *Server) Close() error {
	return s.Db.Close()
}

func (s *Server) Query(query string) (*sql.Rows, error) {
	return s.Db.Query(query)
}

// Scrape loads metrics.
func (s *Server) Scrape(ch chan<- prometheus.Metric, totalScrapes, rdsDatabaseConnections, rdsCurrentCapacity float64) error {
	fingerprint, err := ParseFingerprint(s.Dsn)
	if err != nil {
		return err
	}

	// track scrap duration and send it in the channel
	defer func(begun time.Time) {
		duration := time.Since(begun).Seconds()

		s.NsMap.SetInternalMetrics(ch, duration, totalScrapes, rdsDatabaseConnections, rdsCurrentCapacity)

	}(time.Now())

	serverLabels := prometheus.Labels{
		ServerLabelName: fingerprint,
	}

	s.mappingMtx.RLock()
	defer s.mappingMtx.RUnlock()

	if err = s.SettMetrics.QuerySettings(ch, s.Db, serverLabels); err != nil {
		return fmt.Errorf("error retrieving settings: %s", err)
	}

	semanticVersion, versionString, err := s.checkPostgresVersion()
	if err != nil {
		return fmt.Errorf("error checking postgres version: %s", err)
	}

	errMap := s.NsMap.QueryNamespaceMappings(ch, s.Db, serverLabels, Queries(), MetricMaps(), semanticVersion, versionString)
	if len(errMap) > 0 {
		return fmt.Errorf("queryNamespaceMappings returned %d errors", len(errMap))
	}

	return nil
}

func (s *Server) checkPostgresVersion() (semver.Version, string, error) {
	level.Debug(logger).Log("msg", "Querying PostgreSQL version")
	versionRow := s.Db.QueryRow("SELECT version();")
	var versionString string
	err := versionRow.Scan(&versionString)
	if err != nil {
		return semver.Version{}, "", fmt.Errorf("Error scanning version string: %v", err)
	}
	semanticVersion, err := parseVersion(versionString)
	if err != nil {
		return semver.Version{}, "", fmt.Errorf("Error parsing version string: %v", err)
	}

	return semanticVersion, versionString, nil
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
