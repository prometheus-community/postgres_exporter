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
	"slices"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestForEachDatabase(t *testing.T) {
	databases := []string{"a", "b", "c", "d", "e"}

	var mu sync.Mutex
	var seen []string
	var maxInFlight, inFlight int32

	forEachDatabase(context.Background(), databases, 2, func(_ context.Context, database string) {
		n := atomic.AddInt32(&inFlight, 1)
		for {
			max := atomic.LoadInt32(&maxInFlight)
			if n <= max || atomic.CompareAndSwapInt32(&maxInFlight, max, n) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)

		mu.Lock()
		seen = append(seen, database)
		mu.Unlock()
	})

	sort.Strings(seen)
	if !slices.Equal(seen, databases) {
		t.Fatalf("forEachDatabase() visited %v, want %v", seen, databases)
	}
	if maxInFlight > 2 {
		t.Fatalf("forEachDatabase() ran %d databases concurrently, want at most 2", maxInFlight)
	}
}

// TestForEachDatabaseRespectsDeadline is a regression test: batching used to
// give each batch its own fresh context.WithTimeout(context.Background(),
// ...), so a caller-supplied deadline could be blown past once enough
// batches ran in sequence. forEachDatabase must instead honor a single
// shared ctx and stop scheduling new work once it expires.
func TestForEachDatabaseRespectsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	databases := make([]string, 100)
	for i := range databases {
		databases[i] = "db"
	}

	var ran int32
	start := time.Now()
	forEachDatabase(ctx, databases, 1, func(ctx context.Context, _ string) {
		atomic.AddInt32(&ran, 1)
		select {
		case <-time.After(5 * time.Millisecond):
		case <-ctx.Done():
		}
	})
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("forEachDatabase() took %s, want it to stop shortly after the 20ms deadline", elapsed)
	}
	if n := atomic.LoadInt32(&ran); n >= int32(len(databases)) {
		t.Fatalf("forEachDatabase() ran all %d databases, want the deadline to cut it short", n)
	}
}
