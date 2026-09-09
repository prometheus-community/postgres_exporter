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

func TestDiscoverDatabases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"datname"}).
		AddRow("keep").
		AddRow("excluded").
		AddRow("not_included")
	mock.ExpectQuery(sanitizeQuery(`SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()`)).
		WillReturnRows(rows)

	got, err := discoverDatabases(context.Background(), db, []string{"keep", "excluded"}, []string{"excluded"})
	if err != nil {
		t.Fatalf("discoverDatabases() error = %v", err)
	}
	if want := []string{"keep"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("discoverDatabases() = %v, want %v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
