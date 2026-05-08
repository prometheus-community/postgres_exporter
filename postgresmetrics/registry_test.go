// Copyright 2026 The Prometheus Authors
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

package postgresmetrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestNewRequiresDataSourceName(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigWithDefaults()

	registration, err := New(cfg, promslog.NewNopLogger())
	if err == nil {
		t.Fatal("New() error = nil, want collector creation error")
	}
	if registration != nil {
		t.Fatal("New() registration != nil, want nil")
	}
	if err := registration.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewValidatesCollectionTimeout(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost/postgres"}
	cfg.CollectionTimeout = time.Nanosecond

	registration, err := New(cfg, promslog.NewNopLogger())
	if err == nil {
		t.Fatal("New() error = nil, want timeout validation error")
	}
	if registration != nil {
		t.Fatal("New() registration != nil, want nil")
	}
	if err := registration.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewReturnsCollectorSet(t *testing.T) {
	t.Parallel()

	registration, err := New(testConfig(), promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if registration == nil {
		t.Fatal("New() registration = nil")
	}
	if err := registration.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := registration.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestCollectorSetRegisterRegistersCollectors(t *testing.T) {
	t.Parallel()

	registration, err := New(testConfig(), promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() {
		if err := registration.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	registry := prometheus.NewRegistry()
	if err := registration.Register(registry); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := registry.Gather(); err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	err = registration.Register(registry)
	if err == nil {
		t.Fatal("second Register() error = nil, want duplicate registration error")
	}
	if !strings.Contains(err.Error(), "duplicate metrics collector registration attempted") {
		t.Fatalf("second Register() error = %v, want duplicate registration error", err)
	}
}

func TestNewProbeReturnsCollectorSet(t *testing.T) {
	t.Parallel()

	registration, err := NewProbe(testConfig(), promslog.NewNopLogger(), testDSN)
	if err != nil {
		t.Fatalf("NewProbe() error = %v", err)
	}
	defer func() {
		if err := registration.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	registry := prometheus.NewRegistry()
	if err := registration.Register(registry); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := registry.Gather(); err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
}

func TestNewRegistryRegistersCollectors(t *testing.T) {
	t.Parallel()

	registry, registration, err := NewRegistry(testConfig(), promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	if registry == nil {
		t.Fatal("NewRegistry() registry = nil")
	}
	if registration == nil {
		t.Fatal("NewRegistry() registration = nil")
	}
	if _, err := registry.Gather(); err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	if err := registration.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

const testDSN = "postgresql://localhost/postgres?sslmode=disable"

func testConfig() config.Config {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = []string{testDSN}
	return cfg
}
