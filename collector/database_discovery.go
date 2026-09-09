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
