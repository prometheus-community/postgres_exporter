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
	"errors"
	"testing"

	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/common/promslog"
)

func TestMultiInstanceCollectorInitFailsFastOnPrimaryError(t *testing.T) {
	logger := promslog.NewNopLogger()
	targets := func() []exporter.Target {
		return []exporter.Target{{DSN: "postgresql://primary", Primary: true}}
	}
	build := func(dsn string, primary bool) (*PostgresCollector, error) {
		return nil, errors.New("boom")
	}

	m := newMultiInstanceCollector(logger, targets, build)
	if err := m.init(); err == nil {
		t.Fatal("init() error = nil, want error when a primary target fails to build")
	}
}

func TestMultiInstanceCollectorInitToleratesDiscoveredTargetError(t *testing.T) {
	logger := promslog.NewNopLogger()
	targets := func() []exporter.Target {
		return []exporter.Target{
			{DSN: "postgresql://primary", Primary: true},
			{DSN: "bad", Primary: false},
		}
	}
	build := func(dsn string, primary bool) (*PostgresCollector, error) {
		if dsn == "bad" {
			return nil, errors.New("boom")
		}
		return NewPostgresCollector(logger, nil, dsn, nil)
	}

	m := newMultiInstanceCollector(logger, targets, build)
	if err := m.init(); err != nil {
		t.Fatalf("init() error = %v, want nil (a failing discovered target must not be fatal)", err)
	}
	if got, want := len(m.collectors), 1; got != want {
		t.Fatalf("len(collectors) = %d, want %d", got, want)
	}
}

func TestMultiInstanceCollectorRefreshAddsAndRemovesTargets(t *testing.T) {
	logger := promslog.NewNopLogger()
	dsns := []string{"postgresql://a"}
	targets := func() []exporter.Target {
		out := make([]exporter.Target, len(dsns))
		for i, dsn := range dsns {
			out[i] = exporter.Target{DSN: dsn, Primary: i == 0}
		}
		return out
	}
	builds := 0
	build := func(dsn string, primary bool) (*PostgresCollector, error) {
		builds++
		return NewPostgresCollector(logger, nil, dsn, nil)
	}

	m := newMultiInstanceCollector(logger, targets, build)
	if err := m.init(); err != nil {
		t.Fatalf("init() error = %v", err)
	}
	if got, want := builds, 1; got != want {
		t.Fatalf("builds after init = %d, want %d", got, want)
	}

	// A new database appears.
	dsns = []string{"postgresql://a", "postgresql://b"}
	if got, want := len(m.refresh()), 2; got != want {
		t.Fatalf("len(refresh()) = %d, want %d", got, want)
	}
	if got, want := builds, 2; got != want {
		t.Fatalf("builds after first refresh = %d, want %d", got, want)
	}

	// Refreshing again with the same targets must reuse cached collectors.
	if got, want := len(m.refresh()), 2; got != want {
		t.Fatalf("len(refresh()) = %d, want %d", got, want)
	}
	if got, want := builds, 2; got != want {
		t.Fatalf("builds after second refresh = %d, want %d (collectors should be cached)", got, want)
	}

	// The database disappears.
	dsns = []string{"postgresql://a"}
	if got, want := len(m.refresh()), 1; got != want {
		t.Fatalf("len(refresh()) = %d, want %d", got, want)
	}
	if _, ok := m.collectors["postgresql://b"]; ok {
		t.Fatal("collector for a removed target is still cached")
	}
}
