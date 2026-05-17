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

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// newProbeTestDSN returns a DSN suitable for probe-collector construction.
// The exported ConfigureTarget path is the only way for tests to obtain
// a config.DSN value because the parser is package-private.
func newProbeTestDSN(t *testing.T) config.DSN {
	t.Helper()
	auth := config.AuthModule{
		Type:     "userpass",
		UserPass: config.UserPass{Username: "user", Password: "pass"},
	}
	dsn, err := auth.ConfigureTarget("postgres://localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("ConfigureTarget: %v", err)
	}
	return dsn
}

// TestNewProbeCollectorAllocatesAuroraProbe verifies that the per-request
// instance gets a non-nil auroraProbe so detectAurora() can run on the
// first scrape (and exactly once for the lifetime of this ProbeCollector).
func TestNewProbeCollectorAllocatesAuroraProbe(t *testing.T) {
	logger := promslog.NewNopLogger()
	registry := prometheus.NewRegistry()

	pc, err := NewProbeCollector(logger, nil, registry, newProbeTestDSN(t))
	if err != nil {
		t.Fatalf("NewProbeCollector: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	if pc.instance == nil {
		t.Fatal("ProbeCollector.instance must not be nil")
	}
	if pc.instance.auroraProbe == nil {
		t.Error("instance.auroraProbe must be allocated for the probe path")
	}
}

// TestNewProbeCollectorRespectsCollectorState verifies that NewProbeCollector
// only wires up collectors whose state flag is true. It does so without
// caring about specific collector names — we just flip a known collector's
// state and observe whether it ends up in the resulting collectors map.
func TestNewProbeCollectorRespectsCollectorState(t *testing.T) {
	// Snapshot to restore.
	snapshot := make(map[string]bool, len(collectorState))
	for k, v := range collectorState {
		snapshot[k] = *v
	}
	t.Cleanup(func() {
		for k, v := range snapshot {
			*collectorState[k] = v
		}
	})

	// Disable everything, then enable exactly one collector.
	for _, state := range collectorState {
		*state = false
	}
	target := "database"
	if _, ok := collectorState[target]; !ok {
		t.Skipf("collector %q not registered, skipping", target)
	}
	*collectorState[target] = true

	logger := promslog.NewNopLogger()
	registry := prometheus.NewRegistry()
	pc, err := NewProbeCollector(logger, nil, registry, newProbeTestDSN(t))
	if err != nil {
		t.Fatalf("NewProbeCollector: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	if _, ok := pc.collectors[target]; !ok {
		t.Errorf("expected collector %q to be wired up, got %v", target, keysOf(pc.collectors))
	}
	for name := range pc.collectors {
		if name != target {
			t.Errorf("unexpected collector %q wired up", name)
		}
	}
}

// TestProbeCollectorDescribeIsEmpty pins the documented behavior: probe
// collectors emit their metrics via MustNewConstMetric and intentionally
// produce nothing from Describe(). If that contract changes we want a
// loud failure here so we revisit prometheus registration semantics.
func TestProbeCollectorDescribeIsEmpty(t *testing.T) {
	logger := promslog.NewNopLogger()
	registry := prometheus.NewRegistry()

	pc, err := NewProbeCollector(logger, nil, registry, newProbeTestDSN(t))
	if err != nil {
		t.Fatalf("NewProbeCollector: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	ch := make(chan *prometheus.Desc, 8)
	pc.Describe(ch)
	close(ch)
	for d := range ch {
		t.Errorf("Describe emitted an unexpected descriptor: %v", d)
	}
}

func keysOf(m map[string]Collector) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
