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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func expectAuroraFunctionExists(mock sqlmock.Sqlmock, name string, exists bool) {
	mock.ExpectQuery(sanitizeQuery(auroraFunctionExistsQuery)).
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}

func TestAuroraFunctionExists(t *testing.T) {
	for _, exists := range []bool{true, false} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error opening a stub db connection: %s", err)
		}

		expectAuroraFunctionExists(mock, "aurora_stat_logical_wal_cache", exists)

		got, err := auroraFunctionExists(context.Background(), db, "aurora_stat_logical_wal_cache")
		if err != nil {
			t.Errorf("Error calling auroraFunctionExists: %s", err)
		}
		if got != exists {
			t.Errorf("Expected %v, got %v", exists, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled exceptions: %s", err)
		}
		db.Close()
	}
}
