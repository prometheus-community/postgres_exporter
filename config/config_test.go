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
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewConfigWithDefaults(t *testing.T) {
	cfg := NewConfigWithDefaults()

	if got, want := cfg.MetricPrefix, DefaultMetricPrefix; got != want {
		t.Fatalf("MetricPrefix = %q, want %q", got, want)
	}
	if got, want := cfg.CollectionTimeout, DefaultCollectionTimeout; got != want {
		t.Fatalf("CollectionTimeout = %v, want %v", got, want)
	}
	if cfg.DisableDefaultMetrics {
		t.Fatal("DisableDefaultMetrics = true, want false")
	}
	if got, want := cfg.PGStatStatements.IncludeQuery, DefaultPGStatStatementsIncludeQuery; got != want {
		t.Fatalf("PGStatStatements.IncludeQuery = %t, want %t", got, want)
	}
	if len(cfg.Collectors) == 0 {
		t.Fatal("Collectors is empty, want default collector config")
	}
	if got, want := cfg.Collectors[CollectorDatabase], true; got != want {
		t.Fatalf("Collectors[%q] = %t, want %t", CollectorDatabase, got, want)
	}
	if got, want := cfg.Collectors[CollectorStatStatements], false; got != want {
		t.Fatalf("Collectors[%q] = %t, want %t", CollectorStatStatements, got, want)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}

	if cfg.Validated() {
		t.Fatal("Validated() = true before Validate, want false")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !cfg.Validated() {
		t.Fatal("Validated() = false after Validate, want true")
	}
}

func TestConfigValidateFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{
			name: "empty metric prefix",
			mutate: func(cfg *Config) {
				cfg.MetricPrefix = ""
			},
			want: "metric prefix must not be empty",
		},
		{
			name: "zero collection timeout",
			mutate: func(cfg *Config) {
				cfg.CollectionTimeout = 0
			},
			want: "collection timeout must be greater than zero",
		},
		{
			name: "empty data source",
			mutate: func(cfg *Config) {
				cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres", ""}
			},
			want: "data source name at index 1 must not be empty",
		},
		{
			name: "zero pg_stat_statements query length",
			mutate: func(cfg *Config) {
				cfg.PGStatStatements.QueryLength = 0
			},
			want: "pg_stat_statements query length must be greater than zero",
		},
		{
			name: "empty collector name",
			mutate: func(cfg *Config) {
				cfg.Collectors[""] = true
			},
			want: "collector name must not be empty",
		},
		{
			name: "unknown collector name",
			mutate: func(cfg *Config) {
				cfg.Collectors["does_not_exist"] = true
			},
			want: `unknown collector "does_not_exist"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := NewConfigWithDefaults()
			cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
			test.mutate(&cfg)

			err := cfg.Validate()
			if err == nil || err.Error() != test.want {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
			if cfg.Validated() {
				t.Fatal("Validated() = true after failed Validate, want false")
			}
		})
	}
}

func TestConfigValidateAcceptsNoDataSourcesForMultiTargetMode(t *testing.T) {
	cfg := NewConfigWithDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateAcceptsCustomTimeout(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.CollectionTimeout = 30 * time.Second
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLoadAuthConfigFile(t *testing.T) {
	config, err := LoadAuthConfig("testdata/config-good.yaml")
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}
	if len(config.AuthModules) == 0 {
		t.Fatal("LoadAuthConfig() loaded no auth modules")
	}
}

func TestDecodeAuthConfig(t *testing.T) {
	config, err := DecodeAuthConfig(strings.NewReader(`
auth_modules:
  module:
    type: userpass
    userpass:
      username: user
      password: pass
`))
	if err != nil {
		t.Fatalf("DecodeAuthConfig() error = %v", err)
	}
	if got, want := config.AuthModules["module"].UserPass.Username, "user"; got != want {
		t.Fatalf("username = %q, want %q", got, want)
	}
}

func TestReloadAuthConfig(t *testing.T) {
	ch, err := NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	if err := ch.ReloadAuthConfig("testdata/config-good.yaml", nil); err != nil {
		t.Errorf("error loading config: %s", err)
	}
}

func TestNewHandlerRequiresRegisterer(t *testing.T) {
	handler, err := NewHandler(nil)
	if err == nil {
		t.Fatal("NewHandler() error = nil, want error")
	}
	if handler != nil {
		t.Fatalf("NewHandler() handler = %v, want nil", handler)
	}
}

func TestLoadBadConfigs(t *testing.T) {
	ch, err := NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "testdata/config-bad-auth-module.yaml",
			want:  "error parsing config file \"testdata/config-bad-auth-module.yaml\": yaml: unmarshal errors:\n  line 3: field pretendauth not found in type config.AuthModule",
		},
		{
			input: "testdata/config-bad-extra-field.yaml",
			want:  "error parsing config file \"testdata/config-bad-extra-field.yaml\": yaml: unmarshal errors:\n  line 8: field doesNotExist not found in type config.AuthModule",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := ch.ReloadAuthConfig(test.input, nil)
			if got == nil || got.Error() != test.want {
				t.Fatalf("ReloadAuthConfig(%q) = %v, want %s", test.input, got, test.want)
			}
		})
	}
}
