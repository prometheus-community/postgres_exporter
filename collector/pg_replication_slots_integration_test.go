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
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPGReplicationSlotsCollectorPhysicalSlot(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	version, err := queryVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("queryVersion() error = %v", err)
	}

	before := serverEpoch(t, db)
	const slotName = "pgexp_integration_physical_slot"
	if _, err := db.Exec(`SELECT pg_create_physical_replication_slot($1)`, slotName); err != nil {
		t.Fatalf("pg_create_physical_replication_slot() error = %v", err)
	}
	defer func() {
		if _, err := db.Exec(`SELECT pg_drop_replication_slot($1)`, slotName); err != nil {
			t.Logf("pg_drop_replication_slot() error = %v", err)
		}
	}()

	inst := &instance{db: db, version: version}
	ch := make(chan prometheus.Metric)
	errCh := make(chan error, 1)
	go func() {
		errCh <- PGReplicationSlotsCollector{}.Update(context.Background(), inst, ch)
		close(ch)
	}()

	got := map[string]float64{}
	for metric := range ch {
		m := readMetric(metric)
		if m.labels["slot_name"] != slotName {
			continue
		}
		got[metric.Desc().String()] = m.value
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	after := serverEpoch(t, db)

	if v, ok := got[replicationSlotsIsActiveDesc.String()]; !ok || v != 0 {
		t.Fatalf("pg_replication_slots_slot_is_active = %v (reported: %v), want 0", v, ok)
	}
	// conflicting is NULL for physical slots.
	if _, ok := got[replicationSlotsConflictingDesc.String()]; ok {
		t.Errorf("pg_replication_slots_conflicting reported for a physical slot")
	}

	abovePG17 := version.GTE(semver.MustParse("17.0.0"))
	for _, desc := range []*prometheus.Desc{replicationSlotsFailoverDesc, replicationSlotsSyncedDesc} {
		v, ok := got[desc.String()]
		switch {
		case abovePG17 && (!ok || v != 0):
			t.Errorf("%s = %v (reported: %v), want 0", desc, v, ok)
		case !abovePG17 && ok:
			t.Errorf("%s reported on PostgreSQL %s", desc, version)
		}
	}

	inactiveSince, ok := got[replicationSlotsInactiveSinceDesc.String()]
	switch {
	case !abovePG17 && ok:
		t.Errorf("pg_replication_slots_inactive_since_timestamp_seconds reported on PostgreSQL %s", version)
	case abovePG17 && !ok:
		t.Errorf("pg_replication_slots_inactive_since_timestamp_seconds not reported on PostgreSQL %s", version)
	case abovePG17 && (inactiveSince < before-1 || inactiveSince > after+1):
		t.Errorf("pg_replication_slots_inactive_since_timestamp_seconds = %v, want between %v and %v", inactiveSince, before, after)
	}
}

// serverEpoch returns the database server's current time in seconds since
// the Unix epoch.
func serverEpoch(t *testing.T, db *sql.DB) float64 {
	t.Helper()
	var epoch float64
	if err := db.QueryRow(`SELECT EXTRACT(EPOCH FROM clock_timestamp())::float8`).Scan(&epoch); err != nil {
		t.Fatalf("querying clock_timestamp(): %v", err)
	}
	return epoch
}
