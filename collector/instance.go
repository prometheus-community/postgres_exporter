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

	"github.com/blang/semver/v4"
)

type Instance struct {
	dsn     string
	db      *sql.DB
	version semver.Version
	closeDB bool // whether we should close the connection on Close()
}

func NewInstance(dsn string) (*Instance, error) {
	i := &Instance{
		dsn: dsn,
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
func (i *Instance) copy() *Instance {
	return &Instance{
		dsn: i.dsn,
	}
}

func (i *Instance) setup() error {
	db, err := sql.Open("postgres", i.dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	i.db = db
	i.closeDB = true // we created this connection, so we should close it

	version, err := queryVersion(i.db)
	if err != nil {
		return fmt.Errorf("error querying postgresql version: %w", err)
	} else {
		i.version = version
	}
	return nil
}

// SetupWithConnection sets up the instance with an existing database connection.
func (i *Instance) SetupWithConnection(db *sql.DB) error {
	i.db = db
	i.closeDB = false // we're borrowing this connection, don't close it

	version, err := queryVersion(i.db)
	if err != nil {
		return fmt.Errorf("error querying postgresql version: %w", err)
	}
	i.version = version
	return nil
}

func (i *Instance) getDB() *sql.DB {
	return i.db
}

func (i *Instance) Close() error {
	if i.closeDB {
		return i.db.Close()
	}
	return nil
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

// InstanceFactory creates instances for collectors to use
type InstanceFactory func() (*Instance, error)

// InstanceFactoryFromTemplate creates a factory that copies from a template instance
// and creates a new database connection for each call
func InstanceFactoryFromTemplate(template *Instance) InstanceFactory {
	return func() (*Instance, error) {
		inst := template.copy()
		err := inst.setup() // Creates new connection, sets closeDB=true
		if err != nil {
			return nil, err
		}
		return inst, nil
	}
}
