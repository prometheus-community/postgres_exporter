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
	"time"

	"github.com/prometheus-community/postgres_exporter/collector"
)

const (
	DefaultMetricPrefix   = "pg"
	DefaultAuthConfigFile = "postgres_exporter.yml"
)

const DefaultCollectionTimeout = time.Minute

// Config describes the core PostgreSQL exporter settings without tying callers
// to any particular configuration frontend.
type Config struct {
	DataSourceNames []string

	DisableDefaultMetrics  bool
	DisableSettingsMetrics bool

	MetricPrefix   string
	AuthConfigFile string

	Collectors        []CollectorConfig
	StatStatements    StatStatementsConfig
	CollectionTimeout time.Duration
}

type CollectorConfig struct {
	Name           string
	DefaultEnabled bool
	Enabled        bool
}

type StatStatementsConfig struct {
	IncludeQuery     bool
	QueryLength      uint
	Limit            uint
	ExcludeDatabases []string
	ExcludeUsers     []string
}

// NewConfigWithDefaults returns a Config initialized with exporter defaults.
func NewConfigWithDefaults() Config {
	return Config{
		MetricPrefix:      DefaultMetricPrefix,
		AuthConfigFile:    DefaultAuthConfigFile,
		Collectors:        defaultCollectors(),
		StatStatements:    defaultStatStatements(),
		CollectionTimeout: DefaultCollectionTimeout,
	}
}

// PrimaryDataSourceName returns the first configured DSN, matching the current
// exporter behavior for the typed collector path.
func (c Config) PrimaryDataSourceName() string {
	if len(c.DataSourceNames) == 0 {
		return ""
	}
	return c.DataSourceNames[0]
}

func (c Config) CollectorStates() map[string]bool {
	states := make(map[string]bool, len(c.Collectors))
	for _, collectorConfig := range c.Collectors {
		states[collectorConfig.Name] = collectorConfig.Enabled
	}
	return states
}

func defaultCollectors() []CollectorConfig {
	collectorMetadata := collector.Collectors()
	collectors := make([]CollectorConfig, 0, len(collectorMetadata))
	for _, metadata := range collectorMetadata {
		collectors = append(collectors, CollectorConfig{
			Name:           metadata.Name,
			DefaultEnabled: metadata.DefaultEnabled,
			Enabled:        metadata.DefaultEnabled,
		})
	}
	return collectors
}

func defaultStatStatements() StatStatementsConfig {
	return StatStatementsConfig{
		QueryLength: collector.DefaultStatStatementsQueryLength,
		Limit:       collector.DefaultStatStatementsLimit,
	}
}
