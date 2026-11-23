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
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type Runtime struct {
	exporter          *exporter.Exporter
	postgresCollector *PostgresCollector
	collectorStates   map[string]bool
}

func NewRuntime(validatedConfig config.ValidatedConfig, logger *slog.Logger) (*Runtime, error) {
	if !validatedConfig.Valid() {
		return nil, errors.New("config has not been validated; obtain a ValidatedConfig from Config.Validate")
	}
	if logger == nil {
		logger = slog.Default()
	}
	cfg := validatedConfig.Config()

	exporterCollector := exporter.NewExporter(cfg.DataSourceNames, logger, exporterOptions(cfg)...)
	collectorStates := config.DefaultCollectorConfig()
	maps.Copy(collectorStates, cfg.Collectors)
	runtime := &Runtime{
		exporter:        exporterCollector,
		collectorStates: collectorStates,
	}

	if len(cfg.DataSourceNames) == 0 {
		return runtime, nil
	}

	opts := []Option{
		WithCollectionTimeout(cfg.CollectionTimeout.String()),
		WithCollectorStates(cfg.Collectors),
		WithLongRunningTransactionsConfig(cfg.LongRunningTransactions),
		WithPGStatStatementsConfig(cfg.PGStatStatements),
		WithWrapLargeCounters(cfg.WrapLargeCounters),
	}
	if cfg.AutoDiscoverDatabases {
		// Run the database-scoped collectors (stat_user_tables,
		// statio_user_tables, statio_user_indexes, ...) against every other
		// database on the server too. Server-scoped collectors are
		// unaffected: they still only ever run once, against
		// DataSourceNames[0], so they are never duplicated.
		opts = append(opts, WithDatabaseDiscovery(cfg.IncludeDatabases, cfg.ExcludeDatabases, cfg.AutoDiscoverDatabasesMaxConcurrency))
	}

	postgresCollector, err := NewPostgresCollector(
		logger,
		cfg.ExcludeDatabases,
		cfg.DataSourceNames[0],
		nil,
		opts...,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("create postgres collector: %w", err)
	}
	runtime.postgresCollector = postgresCollector

	return runtime, nil
}

func (r *Runtime) Collectors() []prometheus.Collector {
	collectors := []prometheus.Collector{r.exporter}
	if r.postgresCollector != nil {
		collectors = append(collectors, r.postgresCollector)
	}
	return collectors
}

func (r *Runtime) FilteredCollectors(include, exclude []string) ([]prometheus.Collector, error) {
	if len(include) > 0 && len(exclude) > 0 {
		return nil, errors.New("combined collect[] and exclude[] filters are not allowed")
	}
	if len(include) == 0 && len(exclude) == 0 {
		return r.Collectors(), nil
	}
	for _, name := range slices.Concat(include, exclude) {
		if _, ok := r.collectorStates[name]; !ok {
			return nil, fmt.Errorf("unknown collector %q", name)
		}
	}

	selected := make(map[string]bool)
	var collectors []prometheus.Collector
	if len(include) > 0 {
		for _, name := range include {
			if !r.collectorStates[name] {
				return nil, fmt.Errorf("collector %q is disabled", name)
			}
			selected[name] = true
		}
	} else {
		for name, enabled := range r.collectorStates {
			if enabled && !slices.Contains(exclude, name) {
				selected[name] = true
			}
		}
		if len(selected) == 0 {
			return nil, errors.New("exclude[] filter disables every enabled collector")
		}
		collectors = append(collectors, r.exporter)
	}
	if r.postgresCollector != nil {
		collectors = append(collectors, r.postgresCollector.withCollectors(selected))
	}
	return collectors, nil
}

func (r *Runtime) Close() error {
	var err error
	if r.exporter != nil {
		r.exporter.CloseServers()
	}
	if r.postgresCollector != nil {
		err = errors.Join(err, r.postgresCollector.Close())
	}
	return err
}

func exporterOptions(cfg config.Config) []exporter.ExporterOpt {
	return []exporter.ExporterOpt{
		exporter.DisableDefaultMetrics(cfg.DisableDefaultMetrics),
		exporter.AutoDiscoverDatabases(cfg.AutoDiscoverDatabases),
		exporter.WithCollectionTimeout(cfg.CollectionTimeout),
		exporter.WithUserQueriesPath(cfg.UserQueriesPath),
		exporter.WithConstantLabels(cfg.ConstantLabels),
		exporter.ExcludeDatabases(cfg.ExcludeDatabases),
		exporter.IncludeDatabases(strings.Join(cfg.IncludeDatabases, ",")),
		exporter.WithMetricPrefix(cfg.MetricPrefix),
		exporter.WrapLargeCounters(cfg.WrapLargeCounters),
	}
}
