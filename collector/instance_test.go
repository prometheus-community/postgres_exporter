// Copyright 2025 The Prometheus Authors
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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestNewInstanceAllocatesAuroraProbe verifies that every instance gets a
// non-nil auroraProbe at construction. Without it the setup() code path
// would silently skip Aurora detection.
func TestNewInstanceAllocatesAuroraProbe(t *testing.T) {
	i, err := newInstance("postgres://user:pass@localhost:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("newInstance: %v", err)
	}
	if i.auroraProbe == nil {
		t.Fatal("auroraProbe must be allocated by newInstance")
	}
}

// TestInstanceCopySharesAuroraProbe is a regression test for the "detect
// every scrape" bug. The parent instance and every copy returned by copy()
// MUST share the same *auroraProbe pointer so that sync.Once collapses all
// detection attempts into one for the lifetime of the process.
func TestInstanceCopySharesAuroraProbe(t *testing.T) {
	i, err := newInstance("postgres://user:pass@localhost:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("newInstance: %v", err)
	}

	c1 := i.copy()
	c2 := i.copy()
	c3 := c1.copy() // copy of a copy must still share the probe

	for name, c := range map[string]*instance{"c1": c1, "c2": c2, "c3": c3} {
		if c.auroraProbe != i.auroraProbe {
			t.Errorf("%s.auroraProbe pointer must equal parent's", name)
		}
	}
}

// TestAuroraProbeRunsOnce verifies the sync.Once semantics on auroraProbe.
// Even if setup() is invoked many times (every scrape, in every copy),
// detectAurora() must run exactly once and the cached value is reused.
func TestAuroraProbeRunsOnce(t *testing.T) {
	probe := &auroraProbe{}

	calls := 0
	for i := 0; i < 100; i++ {
		probe.once.Do(func() {
			calls++
			probe.isAurora = true
		})
	}
	if calls != 1 {
		t.Errorf("once.Do should fire exactly once, fired %d times", calls)
	}
	if !probe.isAurora {
		t.Error("cached isAurora value should persist after first call")
	}
}

// TestDetectAuroraTrue verifies detectAurora returns true when
// aurora_version() succeeds, mimicking an Aurora server.
func TestDetectAuroraTrue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT aurora_version\(\)`).
		WillReturnRows(sqlmock.NewRows([]string{"aurora_version"}).AddRow("15.4"))

	if !detectAurora(db) {
		t.Error("detectAurora must return true when aurora_version() succeeds")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestDetectAuroraFalse verifies detectAurora returns false when
// aurora_version() errors (the function does not exist on plain PostgreSQL).
func TestDetectAuroraFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT aurora_version\(\)`).
		WillReturnError(errFnDoesNotExist)

	if detectAurora(db) {
		t.Error("detectAurora must return false when aurora_version() errors")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// errFnDoesNotExist is a generic error mirroring the message plain
// PostgreSQL returns for an unknown function — we only need a non-nil error
// to drive the negative path.
var errFnDoesNotExist = &fakeErr{msg: `pq: function aurora_version() does not exist`}

type fakeErr struct{ msg string }

func (e *fakeErr) Error() string { return e.msg }

// TestInstanceCopyPreservesDSN verifies that copy() carries over the DSN —
// it is required for the fresh per-scrape instance to reconnect to the
// right server.
func TestInstanceCopyPreservesDSN(t *testing.T) {
	const dsn = "postgres://user:pass@example.invalid:5432/db?sslmode=disable"
	i, err := newInstance(dsn)
	if err != nil {
		t.Fatalf("newInstance: %v", err)
	}

	c := i.copy()
	if c.dsn != dsn {
		t.Errorf("copy.dsn = %q, want %q", c.dsn, dsn)
	}
	if c.db != nil {
		t.Error("copy.db must be nil — setup() establishes a fresh connection")
	}
}

// TestInstanceGetDB ensures getDB() returns whatever was placed in the db
// field. It is a tiny accessor but the rest of the package depends on it
// returning the live connection set up during setup().
func TestInstanceGetDB(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	i := &instance{db: db}
	if got := i.getDB(); got != db {
		t.Errorf("getDB() = %p, want %p", got, db)
	}
}

// TestQueryVersionFromSelect verifies the happy path: SELECT version()
// returns a "PostgreSQL X.Y.Z ..." string that the regex parses.
func TestQueryVersionFromSelect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT version\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 16.4 on aarch64-unknown-linux-gnu, compiled by gcc 12.2.0, 64-bit"))

	v, err := queryVersion(db)
	if err != nil {
		t.Fatalf("queryVersion: %v", err)
	}
	if v.Major != 16 || v.Minor != 4 {
		t.Errorf("queryVersion = %v, want 16.4.x", v)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestQueryVersionFallbackToServerVersion: when SELECT version() returns
// something the regex cannot parse, queryVersion must fall through to
// SHOW server_version and try the looser regex there.
func TestQueryVersionFallbackToServerVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// First call returns an unparseable string (no leading word + version).
	mock.ExpectQuery(`SELECT version\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("???"))
	// Fallback path queries SHOW server_version.
	mock.ExpectQuery(`SHOW server_version;`).
		WillReturnRows(sqlmock.NewRows([]string{"server_version"}).
			AddRow("13.3 (Debian 13.3-1.pgdg100+1)"))

	v, err := queryVersion(db)
	if err != nil {
		t.Fatalf("queryVersion: %v", err)
	}
	if v.Major != 13 || v.Minor != 3 {
		t.Errorf("queryVersion = %v, want 13.3.x", v)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestQueryVersionBothUnparseable: when neither query yields a recognizable
// version, queryVersion returns a non-nil error to make the failure loud.
func TestQueryVersionBothUnparseable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT version\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("???"))
	mock.ExpectQuery(`SHOW server_version;`).
		WillReturnRows(sqlmock.NewRows([]string{"server_version"}).AddRow("???"))

	if _, err := queryVersion(db); err == nil {
		t.Error("queryVersion must return an error when nothing parses")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestQueryVersionFirstQueryFails: a raw SQL error on the first query is
// returned verbatim — we don't paper over connection errors as version
// parse failures.
func TestQueryVersionFirstQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT version\(\);`).WillReturnError(errFnDoesNotExist)

	if _, err := queryVersion(db); err == nil {
		t.Error("queryVersion must propagate the underlying error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestSetupRunsAuroraDetection verifies that setup() probes aurora_version()
// once and caches the result on the shared auroraProbe. Detection is always
// on; aurora_* collectors gate on instance.isAurora at Update() time.
func TestSetupRunsAuroraDetection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT version\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 15.4 on aarch64"))
	mock.ExpectQuery(`SELECT aurora_version\(\)`).
		WillReturnRows(sqlmock.NewRows([]string{"aurora_version"}).AddRow("15.4"))

	i := &instance{
		db:          db,
		auroraProbe: &auroraProbe{},
	}
	v, err := queryVersion(i.db)
	if err != nil {
		t.Fatalf("queryVersion: %v", err)
	}
	i.version = v

	if i.auroraProbe != nil {
		i.auroraProbe.once.Do(func() {
			i.auroraProbe.isAurora = detectAurora(i.db)
		})
		i.isAurora = i.auroraProbe.isAurora
	}

	if !i.isAurora {
		t.Error("isAurora must be true after detect succeeds")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
