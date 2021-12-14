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

	"github.com/blang/semver"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Server describes a connection to Postgres.
// Also it contains metrics map and query overrides.
type Server struct {
	db              *sql.DB
	runonserver     string
	dsn             string
	semanticVersion semver.Version
	versionString   string

	// Last version used to calculate metric map. If mismatch on scrape,
	// then maps are recalculated.
	lastMapVersion semver.Version
	// Currently active metric map
	metricMap map[string]MetricMapNamespace
	// Currently active query overrides
	mappingMtx sync.RWMutex
}

// ServerOpt configures a server.
type ServerOpt func(*Server)

// NewServer establishes a new connection using DSN.
func NewServer(dsn string) (*Server, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	level.Info(logger).Log("msg", "Established new database connection")

	// Get postgresql version
	semanticVersion, versionString, err := checkPostgresVersion(db)
	if err != nil {
		level.Info(logger).Log("err", err)
	}

	s := &Server{
		db:              db,
		dsn:             dsn,
		semanticVersion: semanticVersion,
		versionString:   versionString,
	}

	return s, nil
}

// Close disconnects from Postgres.
func (s *Server) Close() error {
	return s.db.Close()
}

func (s *Server) Query(query string) (*sql.Rows, error) {
	return s.db.Query(query)
}

// Scrape loads metrics.
func (s *Server) Scrape(ch chan<- prometheus.Metric) error {
	fingerprint, err := parseFingerprint(s.dsn)
	if err != nil {
		return err
	}

	serverLabels := prometheus.Labels{
		serverLabelName: fingerprint,
	}

	s.mappingMtx.RLock()

	defer s.mappingMtx.RUnlock()

	if err = querySettings(ch, s, serverLabels); err != nil {
		err = fmt.Errorf("error retrieving settings: %s", err)
	}

	errMap := QueryNamespaceMappings(ch, s, serverLabels)
	if len(errMap) > 0 {
		err = fmt.Errorf("queryNamespaceMappings returned %d errors", len(errMap))
	}

	return err
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

func checkPostgresVersion(db *sql.DB) (semver.Version, string, error) {
	level.Debug(logger).Log("msg", "Querying PostgreSQL version")
	versionRow := db.QueryRow("SELECT version();")
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
