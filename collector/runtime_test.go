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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestConfigCollectorDefaultsHaveRegisteredFactories(t *testing.T) {
	for name := range config.DefaultCollectorConfig() {
		if _, ok := factories[name]; !ok {
			t.Errorf("collector %q has a configured default but no registered factory", name)
		}
	}
}

func TestNewRuntimeRequiresValidatedConfig(t *testing.T) {
	runtime, err := NewRuntime(config.ValidatedConfig{}, promslog.NewNopLogger())
	if err == nil {
		t.Fatal("NewRuntime() error = nil, want error")
	}
	if runtime != nil {
		t.Fatalf("NewRuntime() runtime = %v, want nil", runtime)
	}
}

func TestNewRuntimeCollectorsWithoutDataSource(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	if got, want := len(runtime.Collectors()), 1; got != want {
		t.Fatalf("len(Collectors()) = %d, want %d", got, want)
	}
}

func TestNewRuntimeCollectorsWithDataSource(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	if got, want := len(runtime.Collectors()), 2; got != want {
		t.Fatalf("len(Collectors()) = %d, want %d", got, want)
	}
}

func TestNewRuntimePropagatesWrapLargeCounters(t *testing.T) {
	dsn := "postgresql://localhost:5432/postgres?sslmode=disable"
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{dsn}
	cfg.WrapLargeCounters = false
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	pc := runtime.postgresCollector
	if pc.instance.wrapLargeCounters {
		t.Fatal("postgres collector wrapLargeCounters = true, want false")
	}
}

func TestRuntimeFilteredCollectors(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	cfg.Collectors[config.CollectorStatStatements] = true
	cfg.Collectors[config.CollectorLocks] = false
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	var enabled, allButStatStatements []string
	for name := range runtime.postgresCollector.Collectors {
		enabled = append(enabled, name)
		if name != config.CollectorStatStatements {
			allButStatStatements = append(allButStatStatements, name)
		}
	}
	slices.Sort(enabled)
	slices.Sort(allButStatStatements)

	tests := []struct {
		name       string
		include    []string
		exclude    []string
		want       []string
		wantLegacy bool
		wantErr    string
	}{
		{
			name:       "no filter",
			want:       enabled,
			wantLegacy: true,
		},
		{
			name:    "include selects only listed collectors without the legacy exporter",
			include: []string{config.CollectorStatStatements, config.CollectorDatabase, config.CollectorStatStatements},
			want:    []string{config.CollectorDatabase, config.CollectorStatStatements},
		},
		{
			name:       "exclude removes listed collectors and keeps the legacy exporter",
			exclude:    []string{config.CollectorStatStatements, config.CollectorLocks},
			want:       allButStatStatements,
			wantLegacy: true,
		},
		{
			name:    "include and exclude together",
			include: []string{config.CollectorDatabase},
			exclude: []string{config.CollectorStatStatements},
			wantErr: "combined collect[] and exclude[]",
		},
		{
			name:    "include unknown collector",
			include: []string{"no_such_collector"},
			wantErr: `unknown collector "no_such_collector"`,
		},
		{
			name:    "exclude unknown collector",
			exclude: []string{"no_such_collector"},
			wantErr: `unknown collector "no_such_collector"`,
		},
		{
			name:    "include collector disabled by configuration",
			include: []string{config.CollectorLocks},
			wantErr: `collector "locks" is disabled`,
		},
		{
			name:    "exclude every enabled collector",
			exclude: enabled,
			wantErr: "disables every enabled collector",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collectors, err := runtime.FilteredCollectors(test.include, test.exclude)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("FilteredCollectors() error = %v, want error containing %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FilteredCollectors() error = %v", err)
			}

			var got []string
			hasLegacy := false
			for _, c := range collectors {
				switch c := c.(type) {
				case *PostgresCollector:
					for name := range c.Collectors {
						got = append(got, name)
					}
				default:
					hasLegacy = hasLegacy || c == prometheus.Collector(runtime.exporter)
				}
			}
			slices.Sort(got)
			if !slices.Equal(got, test.want) {
				t.Fatalf("collectors = %v, want %v", got, test.want)
			}
			if hasLegacy != test.wantLegacy {
				t.Fatalf("legacy exporter present = %t, want %t", hasLegacy, test.wantLegacy)
			}
			if got := len(runtime.postgresCollector.Collectors); got != len(enabled) {
				t.Fatalf("runtime collectors = %d after filtering, want %d", got, len(enabled))
			}
		})
	}
}

func TestHasDatabaseScopedCollectors(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	cfg.AutoDiscoverDatabases = true
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	tests := []struct {
		name    string
		include []string
		want    bool
	}{
		{name: "server-scoped only", include: []string{config.CollectorSettings, config.CollectorDatabase}, want: false},
		{name: "database-scoped only", include: []string{config.CollectorStatUserTables}, want: true},
		{name: "mixed", include: []string{config.CollectorSettings, config.CollectorStatUserTables}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collectors, err := runtime.FilteredCollectors(test.include, nil)
			if err != nil {
				t.Fatalf("FilteredCollectors() error = %v", err)
			}
			pc := collectors[0].(*PostgresCollector)
			if pc.databaseDiscovery == nil {
				t.Fatal("databaseDiscovery is nil, want set")
			}
			if got := pc.hasDatabaseScopedCollectors(); got != test.want {
				t.Fatalf("hasDatabaseScopedCollectors() = %t, want %t", got, test.want)
			}
		})
	}
	if !runtime.postgresCollector.hasDatabaseScopedCollectors() {
		t.Fatal("hasDatabaseScopedCollectors() = false for the unfiltered collector, want true")
	}
}

func TestRuntimeFilteredCollectorsWithoutDataSource(t *testing.T) {
	validated, err := config.NewConfigWithDefaults().Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	collectors, err := runtime.FilteredCollectors([]string{config.CollectorDatabase}, nil)
	if err != nil {
		t.Fatalf("FilteredCollectors() error = %v", err)
	}
	if len(collectors) != 0 {
		t.Fatalf("FilteredCollectors() = %v, want none", collectors)
	}
}

func TestNewRuntimeOnlyEnablesDatabaseDiscoveryWhenAutoDiscoverDatabasesIsSet(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	if runtime.postgresCollector.databaseDiscovery != nil {
		t.Fatal("databaseDiscovery is set, want nil when AutoDiscoverDatabases is false")
	}
}

func TestNewRuntimeEnablesDatabaseDiscoveryWithAutoDiscoverDatabases(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	cfg.AutoDiscoverDatabases = true
	cfg.IncludeDatabases = []string{"included"}
	cfg.ExcludeDatabases = []string{"excluded"}
	cfg.AutoDiscoverDatabasesMaxConcurrency = 7
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	discovery := runtime.postgresCollector.databaseDiscovery
	if discovery == nil {
		t.Fatal("databaseDiscovery is nil, want set when AutoDiscoverDatabases is true")
	}
	if got, want := discovery.includeDatabases, cfg.IncludeDatabases; len(got) != 1 || got[0] != want[0] {
		t.Errorf("includeDatabases = %v, want %v", got, want)
	}
	if got, want := discovery.excludeDatabases, cfg.ExcludeDatabases; len(got) != 1 || got[0] != want[0] {
		t.Errorf("excludeDatabases = %v, want %v", got, want)
	}
	if got, want := discovery.maxConcurrency, cfg.AutoDiscoverDatabasesMaxConcurrency; got != want {
		t.Errorf("maxConcurrency = %d, want %d", got, want)
	}
}

func TestNewRuntimePropagatesLongRunningTransactionsThreshold(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	cfg.Collectors[config.CollectorLongRunningTransactions] = true
	cfg.LongRunningTransactions.Threshold = 5 * time.Minute
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	pc := runtime.postgresCollector
	collector, ok := pc.Collectors[config.CollectorLongRunningTransactions].(*PGLongRunningTransactionsCollector)
	if !ok {
		t.Fatalf("collector type = %T, want *PGLongRunningTransactionsCollector", collector)
	}
	if got, want := collector.threshold, 5*time.Minute; got != want {
		t.Fatalf("threshold = %v, want %v", got, want)
	}
}
