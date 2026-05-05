// Copyright 2022 The Prometheus Authors
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

package config

import (
	"testing"
	"time"
)

func TestNewConfigWithDefaults(t *testing.T) {
	t.Parallel()

	cfg := NewConfigWithDefaults()

	if got, want := cfg.MetricPrefix, DefaultMetricPrefix; got != want {
		t.Fatalf("MetricPrefix = %q, want %q", got, want)
	}
	if got, want := cfg.CollectionTimeout, time.Minute; got != want {
		t.Fatalf("CollectionTimeout = %s, want %s", got, want)
	}
	if got, want := cfg.AuthConfigFile, DefaultAuthConfigFile; got != want {
		t.Fatalf("AuthConfigFile = %q, want %q", got, want)
	}
}

func TestPrimaryDataSourceName(t *testing.T) {
	t.Parallel()

	cfg := Config{
		DataSourceNames: []string{"postgresql://first", "postgresql://second"},
	}

	if got, want := cfg.PrimaryDataSourceName(), "postgresql://first"; got != want {
		t.Fatalf("PrimaryDataSourceName() = %q, want %q", got, want)
	}
}

func TestPrimaryDataSourceNameEmpty(t *testing.T) {
	t.Parallel()

	if got := (Config{}).PrimaryDataSourceName(); got != "" {
		t.Fatalf("PrimaryDataSourceName() = %q, want empty string", got)
	}
}
