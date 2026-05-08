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

// Package postgresmetrics wires postgres_exporter configuration into typed
// Postgres collectors and Prometheus registries.
package postgresmetrics

import (
	"errors"
	"io"
	"log/slog"

	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// CollectorSet owns typed Postgres collectors and their resources.
type CollectorSet struct {
	collectors []prometheus.Collector
	closers    []io.Closer
	closed     bool
}

// Register registers the typed Postgres collectors on reg. Metrics are produced
// on every subsequent reg.Gather() call, such as during a /metrics scrape.
func (r *CollectorSet) Register(reg prometheus.Registerer) error {
	if r == nil {
		return nil
	}
	if reg == nil {
		return errors.New("nil prometheus registerer")
	}

	var err error
	for _, collector := range r.collectors {
		err = errors.Join(err, reg.Register(collector))
	}
	return err
}

// Close releases resources owned by the registration.
func (r *CollectorSet) Close() error {
	if r == nil || r.closed {
		return nil
	}
	r.closed = true

	var err error
	for _, closer := range r.closers {
		err = errors.Join(err, closer.Close())
	}
	return err
}

// NewRegistry creates a new Prometheus registry and registers core Postgres
// database collectors using cfg.
func NewRegistry(cfg config.Config, logger *slog.Logger) (*prometheus.Registry, *CollectorSet, error) {
	registry := prometheus.NewRegistry()
	registration, err := New(cfg, logger)
	if err != nil {
		return registry, registration, err
	}
	err = registration.Register(registry)
	if err != nil {
		return registry, registration, err
	}
	return registry, registration, nil
}

// New builds long-lived typed Postgres collectors from cfg.
func New(cfg config.Config, logger *slog.Logger) (*CollectorSet, error) {
	logger = defaultLogger(logger)
	postgresCollector, err := collector.NewPostgresCollector(
		logger,
		nil,
		cfg.PrimaryDataSourceName(),
		nil,
		collectorOptionsFromConfig(cfg)...)
	if err != nil {
		return nil, err
	}
	return newCollectorSet(postgresCollector), nil
}

// NewProbe builds per-target probe collectors from cfg and dsn.
func NewProbe(cfg config.Config, logger *slog.Logger, dsn string) (*CollectorSet, error) {
	logger = defaultLogger(logger)
	probeCollector, err := collector.NewProbeCollector(logger, nil, dsn, collectorOptionsFromConfig(cfg)...)
	if err != nil {
		return nil, err
	}
	return newCollectorSet(probeCollector), nil
}

func newCollectorSet(collector prometheus.Collector) *CollectorSet {
	registration := &CollectorSet{
		collectors: []prometheus.Collector{collector},
	}
	if closer, ok := collector.(io.Closer); ok {
		registration.closers = append(registration.closers, closer)
	}
	return registration
}

func collectorOptionsFromConfig(cfg config.Config) []collector.Option {
	options := []collector.Option{
		collector.WithCollectorStates(cfg.CollectorStates()),
		collector.WithStatStatementsConfig(
			cfg.StatStatements.IncludeQuery,
			cfg.StatStatements.QueryLength,
			cfg.StatStatements.Limit,
			cfg.StatStatements.ExcludeDatabases,
			cfg.StatStatements.ExcludeUsers,
		),
	}
	if cfg.CollectionTimeout != 0 {
		options = append(options, collector.WithCollectionTimeoutDuration(cfg.CollectionTimeout))
	}
	return options
}

func defaultLogger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return promslog.NewNopLogger()
	}
	return logger
}
