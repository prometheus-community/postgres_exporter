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
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestDatabaseDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    string
		wantErr bool
	}{
		{name: "postgresql URI", dsn: "postgresql://user:pass@localhost:5432/postgres?sslmode=disable", want: "postgresql://user:pass@localhost:5432/other?sslmode=disable"},
		{name: "postgres URI", dsn: "postgres://localhost/postgres", want: "postgres://localhost/other"},
		{name: "connstring", dsn: "host=localhost port=5432 dbname=postgres", want: "host=localhost port=5432 dbname=postgres dbname=other"},
		{name: "unparsable", dsn: "not a dsn", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := databaseDSN(test.dsn, "other")
			if test.wantErr {
				if err == nil {
					t.Fatalf("databaseDSN() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("databaseDSN() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("databaseDSN() = %q, want %q", got, test.want)
			}
		})
	}
}

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

func TestDatabaseDiscoveryCollectAddsAndRemovesDatabases(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	query := sanitizeQuery(`SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()`)

	builds := 0
	build := func(dsn string) (*PostgresCollector, error) {
		builds++
		pc, err := NewPostgresCollector(logger, nil, "postgresql://local", nil, onlyScope(databaseScope))
		if err != nil {
			return nil, err
		}
		return pc, nil
	}

	d := newDatabaseDiscovery(nil, nil)

	ch := make(chan prometheus.Metric, 1024)

	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"datname"}).AddRow("a"))
	d.collect(context.Background(), "postgresql://local", db, ch, logger, build)
	if got, want := builds, 1; got != want {
		t.Fatalf("builds after first collect = %d, want %d", got, want)
	}
	if got, want := len(d.collectors), 1; got != want {
		t.Fatalf("len(collectors) = %d, want %d", got, want)
	}

	// The same database is discovered again: no new collector is built.
	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"datname"}).AddRow("a"))
	d.collect(context.Background(), "postgresql://local", db, ch, logger, build)
	if got, want := builds, 1; got != want {
		t.Fatalf("builds after second collect = %d, want %d (collector should be cached)", got, want)
	}

	// A new database is discovered, and the old one disappears.
	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"datname"}).AddRow("b"))
	d.collect(context.Background(), "postgresql://local", db, ch, logger, build)
	if got, want := builds, 2; got != want {
		t.Fatalf("builds after third collect = %d, want %d", got, want)
	}
	if _, ok := d.collectors["a"]; ok {
		t.Fatal("collector for a removed database is still cached")
	}
	if _, ok := d.collectors["b"]; !ok {
		t.Fatal("collector for the newly discovered database was not built")
	}

	if err := d.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseDiscoveryCollectToleratesBuildError(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	query := sanitizeQuery(`SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()`)
	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"datname"}).AddRow("a"))

	build := func(dsn string) (*PostgresCollector, error) {
		return nil, errors.New("boom")
	}

	d := newDatabaseDiscovery(nil, nil)
	ch := make(chan prometheus.Metric, 1024)
	d.collect(context.Background(), "postgresql://local", db, ch, logger, build)

	if got, want := len(d.collectors), 0; got != want {
		t.Fatalf("len(collectors) = %d, want %d (a failing build must not be cached)", got, want)
	}
}
