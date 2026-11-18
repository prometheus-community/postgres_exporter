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

//go:build integration

package collector

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// TestPGReplicationLagQuerySemantics runs pgReplicationQuery on a real server
// with the recovery functions replaced by fixed values, so the lag expression
// can be checked for replica states the CI primary cannot reach.
func TestPGReplicationLagQuerySemantics(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	const streamingCheck = "EXISTS (SELECT 1 FROM pg_stat_wal_receiver WHERE status = 'streaming')"

	for _, tc := range []struct {
		name           string
		receive        string
		replay         string
		streaming      bool
		wantLagMinimum float64
		wantLagMaximum float64
	}{
		// After the WAL receiver restarts on an idle cluster, the receive
		// position is the start of the segment it streams from, behind replay.
		{name: "receive behind replay, streaming", receive: "0/1000000", replay: "0/14EFCE0", streaming: true, wantLagMinimum: 0, wantLagMaximum: 0},
		// Replaying from the archive with no WAL receiver: receive stays at its
		// last value while replay advances, and the replica can still be behind.
		{name: "receive behind replay, not streaming", receive: "0/1000000", replay: "0/14EFCE0", streaming: false, wantLagMinimum: 3599, wantLagMaximum: 3700},
		{name: "receive equals replay, streaming", receive: "0/14EFCE0", replay: "0/14EFCE0", streaming: true, wantLagMinimum: 0, wantLagMaximum: 0},
		{name: "receive equals replay, not streaming", receive: "0/14EFCE0", replay: "0/14EFCE0", streaming: false, wantLagMinimum: 0, wantLagMaximum: 0},
		// WAL received but not replayed yet: the lag is the replay age.
		{name: "receive ahead of replay", receive: "0/2000060", replay: "0/14EFCE0", streaming: true, wantLagMinimum: 3599, wantLagMaximum: 3700},
	} {
		t.Run(tc.name, func(t *testing.T) {
			replacements := []string{
				"NOT pg_is_in_recovery()", "false",
				"pg_is_in_recovery()", "true",
				"pg_last_wal_receive_lsn ()", "'" + tc.receive + "'::pg_lsn",
				"pg_last_wal_replay_lsn ()", "'" + tc.replay + "'::pg_lsn",
				"pg_last_xact_replay_timestamp()", "(now() - interval '1 hour')",
				streamingCheck, strconv.FormatBool(tc.streaming),
			}
			// A replacement whose target is not in the query would leave the
			// real function in place and test the CI server's state instead.
			for i := 0; i < len(replacements); i += 2 {
				if !strings.Contains(pgReplicationQuery, replacements[i]) {
					t.Fatalf("pgReplicationQuery does not contain %q", replacements[i])
				}
			}
			query := strings.NewReplacer(replacements...).Replace(pgReplicationQuery)
			for _, fn := range []string{"pg_is_in_recovery", "pg_last_wal_receive_lsn", "pg_last_wal_replay_lsn", "pg_last_xact_replay_timestamp", "pg_stat_wal_receiver"} {
				if strings.Contains(query, fn) {
					t.Fatalf("%s left in the query after the replacements:\n%s", fn, query)
				}
			}

			var lag, replayAge float64
			var isReplica int64
			if err := db.QueryRow(query).Scan(&lag, &isReplica, &replayAge); err != nil {
				t.Fatalf("query error = %v\n%s", err, query)
			}
			if isReplica != 1 {
				t.Errorf("is_replica = %d, want 1", isReplica)
			}
			if lag < tc.wantLagMinimum || lag > tc.wantLagMaximum {
				t.Errorf("lag = %v, want between %v and %v", lag, tc.wantLagMinimum, tc.wantLagMaximum)
			}
		})
	}
}
