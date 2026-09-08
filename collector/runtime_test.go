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
	"testing"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
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
