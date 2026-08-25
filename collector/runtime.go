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
	"strings"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type Runtime struct {
	exporter          *exporter.Exporter
	postgresCollector *PostgresCollector
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
	runtime := &Runtime{
		exporter: exporterCollector,
	}

	if len(cfg.DataSourceNames) == 0 {
		return runtime, nil
	}

	postgresCollector, err := NewPostgresCollector(
		logger,
		cfg.ExcludeDatabases,
		cfg.DataSourceNames[0],
		nil,
		WithCollectionTimeout(cfg.CollectionTimeout.String()),
		WithCollectorStates(cfg.Collectors),
		WithLongRunningTransactionsConfig(cfg.LongRunningTransactions),
		WithPGStatStatementsConfig(cfg.PGStatStatements),
		WithWrapLargeCounters(cfg.WrapLargeCounters),
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
		exporter.WithUserQueriesPath(cfg.UserQueriesPath),
		exporter.WithConstantLabels(cfg.ConstantLabels),
		exporter.ExcludeDatabases(cfg.ExcludeDatabases),
		exporter.IncludeDatabases(strings.Join(cfg.IncludeDatabases, ",")),
		exporter.WithMetricPrefix(cfg.MetricPrefix),
		exporter.WrapLargeCounters(cfg.WrapLargeCounters),
	}
}
