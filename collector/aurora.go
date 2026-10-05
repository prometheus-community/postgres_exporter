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

// auroraFunctionExistsQuery uses to_regproc, which returns NULL instead of
// raising an error when the function doesn't exist.
const auroraFunctionExistsQuery = `SELECT to_regproc($1) IS NOT NULL`

// auroraFunctionExists reports whether an Aurora function exists on the
// server. Some Aurora functions only exist in newer Aurora PostgreSQL
// versions; their collectors check first, so older versions report no data
// instead of an error on every scrape.
func auroraFunctionExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, auroraFunctionExistsQuery, name).Scan(&exists)
	return exists, err
}
