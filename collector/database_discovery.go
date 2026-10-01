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
	"slices"
	"sync"
)

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

// forEachDatabase runs collect for every entry in databases, using at most
// maxConcurrency workers, all sharing ctx. Sharing a single ctx across every
// worker (rather than giving each database its own fresh deadline) ensures
// the deadline applies to the whole call: a caller-supplied
// context.WithTimeout still expires on schedule no matter how many batches
// of databases it takes to work through the list. A maxConcurrency <= 0 is
// treated as 1.
func forEachDatabase(ctx context.Context, databases []string, maxConcurrency int, collect func(context.Context, string)) {
	if len(databases) == 0 {
		return
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	if maxConcurrency > len(databases) {
		maxConcurrency = len(databases)
	}

	jobs := make(chan string)
	var wg sync.WaitGroup
	wg.Add(maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			defer wg.Done()
			for database := range jobs {
				if ctx.Err() != nil {
					return
				}
				collect(ctx, database)
			}
		}()
	}

	for _, database := range databases {
		select {
		case jobs <- database:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		}
	}
	close(jobs)
	wg.Wait()
}
