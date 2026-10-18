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
)

// functionExistsQuery uses to_regproc, which returns NULL instead of
// raising an error when the function doesn't exist.
const functionExistsQuery = `SELECT to_regproc($1) IS NOT NULL`

// functionExists reports whether a function exists on the server. Some
// functions only exist in newer versions or on some types of instances, e.g.
// Amazon Aurora PostgreSQL; their collectors check first, so other servers
// report no data instead of an error on every scrape.
func functionExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, functionExistsQuery, name).Scan(&exists)
	return exists, err
}
