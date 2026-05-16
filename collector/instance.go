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
	"database/sql"
	"fmt"
	"regexp"
	"sync"

	"github.com/blang/semver/v4"
)

// auroraProbe caches the result of detectAurora() so the aurora_version()
// query runs at most once per process. The parent instance and every copy
// produced by copy() share the same pointer.
type auroraProbe struct {
	once     sync.Once
	isAurora bool
}

type instance struct {
	dsn     string
	db      *sql.DB
	version semver.Version

	// isAurora is set by setup() via the shared auroraProbe (one-shot
	// SELECT aurora_version() per process). Aurora-specific collectors
	// gate on this field; on non-Aurora servers they short-circuit to
	// ErrNoData.
	isAurora    bool
	auroraProbe *auroraProbe
}

func newInstance(dsn string) (*instance, error) {
	i := &instance{
		dsn:         dsn,
		auroraProbe: &auroraProbe{},
	}

	// "Create" a database handle to verify the DSN provided is valid.
	// Open is not guaranteed to create a connection.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.Close()

	return i, nil
}

// copy returns a copy of the instance.
func (i *instance) copy() *instance {
	return &instance{
		dsn:         i.dsn,
		auroraProbe: i.auroraProbe,
	}
}

func (i *instance) setup() error {
	db, err := sql.Open("postgres", i.dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	i.db = db

	version, err := queryVersion(i.db)
	if err != nil {
		return fmt.Errorf("error querying postgresql version: %w", err)
	} else {
		i.version = version
	}

	if i.auroraProbe != nil {
		i.auroraProbe.once.Do(func() {
			i.auroraProbe.isAurora = detectAurora(i.db)
		})
		i.isAurora = i.auroraProbe.isAurora
	}
	return nil
}

// detectAurora reports whether the connected server is Amazon Aurora
// PostgreSQL. It calls aurora_version(), an Aurora-only built-in: any
// error means we are not on Aurora. The check is best-effort and
// silently returns false on failure.
func detectAurora(db *sql.DB) bool {
	var v string
	return db.QueryRow("SELECT aurora_version()").Scan(&v) == nil
}

func (i *instance) getDB() *sql.DB {
	return i.db
}

func (i *instance) Close() error {
	if i.db == nil {
		// setup() was never called (or failed before sql.Open); nothing to
		// close. This guard lets callers defer Close() right after
		// construction without worrying about ordering.
		return nil
	}
	return i.db.Close()
}

// Regex used to get the "short-version" from the postgres version field.
// The result of SELECT version() is something like "PostgreSQL 9.6.2 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 6.2.1 20160830, 64-bit"
var versionRegex = regexp.MustCompile(`^\w+ ((\d+)(\.\d+)?(\.\d+)?)`)
var serverVersionRegex = regexp.MustCompile(`^((\d+)(\.\d+)?(\.\d+)?)`)

func queryVersion(db *sql.DB) (semver.Version, error) {
	var version string
	err := db.QueryRow("SELECT version();").Scan(&version)
	if err != nil {
		return semver.Version{}, err
	}
	submatches := versionRegex.FindStringSubmatch(version)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}

	// We could also try to parse the version from the server_version field.
	// This is of the format 13.3 (Debian 13.3-1.pgdg100+1)
	err = db.QueryRow("SHOW server_version;").Scan(&version)
	if err != nil {
		return semver.Version{}, err
	}
	submatches = serverVersionRegex.FindStringSubmatch(version)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}
	return semver.Version{}, fmt.Errorf("could not parse version from %q", version)
}
