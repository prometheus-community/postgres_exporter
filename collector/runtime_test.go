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
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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

// recordingDriver is a database/sql driver that counts the connectors it is
// asked to open and fails every connection.
type recordingDriver struct {
	mu     sync.Mutex
	opened int
}

func (d *recordingDriver) Open(dsn string) (driver.Conn, error) {
	c, err := d.OpenConnector(dsn)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

func (d *recordingDriver) OpenConnector(string) (driver.Connector, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.opened++
	return failingConnector{d}, nil
}

func (d *recordingDriver) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.opened
}

type failingConnector struct{ d *recordingDriver }

func (failingConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("recording driver doesn't connect")
}

func (c failingConnector) Driver() driver.Driver { return c.d }

var recordingDrivers atomic.Int32

// registerRecordingDriver registers a new recordingDriver under a unique
// name, since database/sql can't unregister one.
func registerRecordingDriver() (string, *recordingDriver) {
	d := &recordingDriver{}
	name := fmt.Sprintf("recording-%d", recordingDrivers.Add(1))
	sql.Register(name, d)
	return name, d
}

func TestNewRuntimeConnectsThroughConfiguredDriver(t *testing.T) {
	name, d := registerRecordingDriver()
	cfg := config.NewConfigWithDefaults()
	// Nothing listens on port 1, so a path that ignores DriverName and uses
	// lib/pq fails without reaching the recording driver.
	cfg.DataSourceNames = []string{"postgresql://exporter@127.0.0.1:1/postgres?sslmode=disable"}
	cfg.DriverName = name
	cfg.CollectionTimeout = 200 * time.Millisecond
	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	runtime, err := NewRuntime(validated, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer runtime.Close()

	// The legacy exporter and the native collector open their own
	// connections; each has to use the configured driver, e.g. the IAM one.
	for _, c := range runtime.Collectors() {
		before := d.count()
		ch := make(chan prometheus.Metric)
		go func() {
			c.Collect(ch)
			close(ch)
		}()
		for range ch {
		}
		if d.count() == before {
			t.Errorf("%T didn't connect through driver %q", c, name)
		}
	}
}
