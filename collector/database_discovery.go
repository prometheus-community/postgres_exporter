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
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// databaseDiscovery fans database-scoped collectors out across every other
// database on the same PostgreSQL server as a PostgresCollector's primary
// connection. It is reconciled against pg_database on every scrape, so
// databases created or dropped after startup are picked up without an
// exporter restart.
type databaseDiscovery struct {
	includeDatabases []string
	excludeDatabases []string

	mu         sync.Mutex
	collectors map[string]*PostgresCollector // keyed by database name
}

func newDatabaseDiscovery(includeDatabases, excludeDatabases []string) *databaseDiscovery {
	return &databaseDiscovery{
		includeDatabases: includeDatabases,
		excludeDatabases: excludeDatabases,
		collectors:       make(map[string]*PostgresCollector),
	}
}

// connstringRe superficially validates a key=value DSN: connstring syntax is
// complex (and not sure if even regular), and we don't need to parse it, so
// we just check that it starts with a valid-ish keyword pair.
var connstringRe = regexp.MustCompile(`^ *[a-zA-Z0-9]+ *= *[^= ]+`)

// databaseDSN returns dsn rewritten to connect to database instead of
// whichever database it originally pointed at.
func databaseDSN(dsn, database string) (string, error) {
	if strings.HasPrefix(dsn, "postgresql://") || strings.HasPrefix(dsn, "postgres://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", fmt.Errorf("malformed dsn: %w", err)
		}
		u.Path = database
		return u.String(), nil
	}
	if connstringRe.MatchString(dsn) {
		// Replacing one dbname with another is complicated, so just append a
		// new dbname to override any existing one.
		return fmt.Sprintf("%s dbname=%s", dsn, database), nil
	}
	return "", errors.New("unable to parse dsn as either a URI or a connstring")
}

// discoverDatabases returns every database on the server db is connected to,
// besides the one that connection is currently attached to, filtered by
// includeDatabases and excludeDatabases.
func discoverDatabases(ctx context.Context, db *sql.DB, includeDatabases, excludeDatabases []string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, err
		}
		if slices.Contains(excludeDatabases, database) {
			continue
		}
		if len(includeDatabases) != 0 && !slices.Contains(includeDatabases, database) {
			continue
		}
		databases = append(databases, database)
	}
	return databases, rows.Err()
}

// collect reconciles the cached per-database collectors against the
// databases currently on the server, then scrapes every one of them.
func (d *databaseDiscovery) collect(ctx context.Context, primaryDSN string, db *sql.DB, ch chan<- prometheus.Metric, logger *slog.Logger, build func(dsn string) (*PostgresCollector, error)) {
	databases, err := discoverDatabases(ctx, db, d.includeDatabases, d.excludeDatabases)
	if err != nil {
		logger.Error("failed to discover databases", "err", err)
		return
	}

	d.mu.Lock()
	next := make(map[string]*PostgresCollector, len(databases))
	for _, database := range databases {
		if pc, ok := d.collectors[database]; ok {
			next[database] = pc
			continue
		}
		dsn, err := databaseDSN(primaryDSN, database)
		if err != nil {
			logger.Error("failed to build dsn for discovered database", "database", database, "err", err)
			continue
		}
		pc, err := build(dsn)
		if err != nil {
			logger.Error("failed to build collector for discovered database", "database", database, "err", err)
			continue
		}
		next[database] = pc
	}
	for database, pc := range d.collectors {
		if _, ok := next[database]; !ok {
			if err := pc.Close(); err != nil {
				logger.Error("failed to close collector for removed database", "database", database, "err", err)
			}
		}
	}
	d.collectors = next

	snapshot := make([]*PostgresCollector, 0, len(next))
	for _, pc := range next {
		snapshot = append(snapshot, pc)
	}
	d.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(snapshot))
	for _, pc := range snapshot {
		go func(pc *PostgresCollector) {
			defer wg.Done()
			pc.Collect(ch)
		}(pc)
	}
	wg.Wait()
}

// Close closes every collector currently cached.
func (d *databaseDiscovery) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	var err error
	for _, pc := range d.collectors {
		err = errors.Join(err, pc.Close())
	}
	return err
}
