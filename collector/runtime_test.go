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
	"maps"
	"testing"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/common/promslog"
)

func TestConfigCollectorDefaultsMatchRegisteredCollectors(t *testing.T) {
	if got, want := config.DefaultCollectorConfig(), DefaultCollectorStates(); !maps.Equal(got, want) {
		t.Fatalf("DefaultCollectorConfig() = %v, want registered collector defaults %v", got, want)
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
