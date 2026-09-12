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
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/blang/semver/v4"
	"github.com/prometheus-community/postgres_exporter/config"

	"github.com/prometheus-community/postgres_exporter/internal/connector"
)

type instance struct {
	dsn               string
	connector         driver.Connector
	db                *sql.DB
	version           semver.Version
	database          string
	wrapLargeCounters bool
}

// copy returns a copy of the instance.
func (i *instance) copy() *instance {
	return &instance{
		dsn:               i.dsn,
		connector:         i.connector,
		wrapLargeCounters: i.wrapLargeCounters,
	}
}

// withDatabase returns a new, unconnected instance whose dsn points at
// database instead of whichever database i's dsn originally targeted. It is
// used to scrape a database discovered alongside i's own primary connection,
// on the same PostgreSQL server.
func (i *instance) withDatabase(database string) (*instance, error) {
	dsn, err := config.NewDSN(i.dsn)
	if err != nil {
		return nil, fmt.Errorf("malformed dsn: %w", err)
	}
	var conn driver.Connector
	if iam, ok := i.connector.(*connector.AWSIAMConnector); ok {
		// Mint tokens as usual, but point each connection at database.
		baseDSN := iam.DSN
		conn, err = connector.NewAWSIAMConnector(iam.DBEndpoint, iam.DBUser, iam.Region, iam.RoleARN, func(token string) string {
			d, derr := config.NewDSN(baseDSN(token))
			if derr != nil {
				return baseDSN(token)
			}
			return d.WithDatabase(database).GetConnectionString()
		})
		if err != nil {
			return nil, err
		}
	}
	other := &instance{
		dsn:               dsn.WithDatabase(database).GetConnectionString(),
		connector:         conn,
		wrapLargeCounters: i.wrapLargeCounters,
		// database is on the same PostgreSQL server as i, so it runs the same
		// version; skip re-querying it over the new connection.
		version: i.version,
	}
	return other, nil
}

func (i *instance) setup(ctx context.Context) error {
	// Without a custom connector, connect with the DSN as sql.Open would.
	if i.connector == nil {
		c, err := connector.NewStaticConnector(i.dsn)
		if err != nil {
			return err
		}
		i.connector = c
	}
	db := sql.OpenDB(i.connector)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	i.db = db

	if i.version.EQ(semver.Version{}) {
		version, err := queryVersion(ctx, i.db)
		if err != nil {
			return fmt.Errorf("error querying postgresql version: %w", err)
		}
		i.version = version
	}

	if err := i.db.QueryRowContext(ctx, "SELECT current_database()").Scan(&i.database); err != nil {
		return fmt.Errorf("error querying current database: %w", err)
	}
	return nil
}

func (i *instance) getDB() *sql.DB {
	return i.db
}

func (i *instance) Close() error {
	if i.db == nil {
		return nil
	}
	err := i.db.Close()
	i.db = nil
	return err
}

// Regex used to get the "short-version" from the postgres version field.
// The result of SELECT version() is something like "PostgreSQL 9.6.2 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 6.2.1 20160830, 64-bit"
var versionRegex = regexp.MustCompile(`^\w+ ((\d+)(\.\d+)?(\.\d+)?)`)
var serverVersionRegex = regexp.MustCompile(`^((\d+)(\.\d+)?(\.\d+)?)`)

func queryVersion(ctx context.Context, db *sql.DB) (semver.Version, error) {
	var version string
	err := db.QueryRowContext(ctx, "SELECT version();").Scan(&version)
	if err != nil {
		return semver.Version{}, err
	}
	submatches := versionRegex.FindStringSubmatch(version)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}

	// We could also try to parse the version from the server_version field.
	// This is of the format 13.3 (Debian 13.3-1.pgdg100+1)
	err = db.QueryRowContext(ctx, "SHOW server_version;").Scan(&version)
	if err != nil {
		return semver.Version{}, err
	}
	submatches = serverVersionRegex.FindStringSubmatch(version)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}
	return semver.Version{}, fmt.Errorf("could not parse version from %q", version)
}
