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

package collector

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	toolkit "github.com/prometheus/exporter-toolkit/collector"
)

const (
	// Namespace for all metrics.
	namespace = "pg"

	collectorFlagPrefix = "collector."
	defaultEnabled      = true
	defaultDisabled     = false
)

type Collector interface {
	Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error
}

type collectorConfig struct {
	logger           *slog.Logger
	excludeDatabases []string
}

// registry replaces the factories/collectorState/forcedCollectors maps that
// were duplicated verbatim from node_exporter.
var collectorRegistry = toolkit.NewRegistry[Collector, collectorConfig](namespace)

// scrapeReporter emits pg_scrape_collector_duration_seconds and
// pg_scrape_collector_success.
var scrapeReporter = toolkit.NewScrapeReporter(namespace, "postgres_exporter")

// registerCollector keeps its original signature so the per-collector files
// registering from init() do not change.
func registerCollector(name string, isDefaultEnabled bool, createFunc func(collectorConfig) (Collector, error)) {
	collectorRegistry.Register(toolkit.Descriptor{
		Name:           name,
		DefaultEnabled: isDefaultEnabled,
	}, createFunc)
}

// RegisterCollectorWithMetadata registers a collector along with the operator
// preconditions it needs, so documentation can be generated from the registry.
func RegisterCollectorWithMetadata(d toolkit.Descriptor, createFunc func(collectorConfig) (Collector, error)) {
	collectorRegistry.Register(d, createFunc)
}

// AddFlags registers the --collector.<name> flags. Call before kingpin.Parse.
func AddFlags(app *kingpin.Application) {
	collectorRegistry.AddFlags(app)
}

// DisableDefaultCollectors disables every collector not named explicitly.
func DisableDefaultCollectors() {
	collectorRegistry.DisableDefaults()
}

// WriteCollectorList writes the machine-readable collector inventory.
func WriteCollectorList(w io.Writer) error {
	return collectorRegistry.WriteJSON(w)
}

// WriteCollectorMarkdown writes the docs table for docs/configuration.md.
func WriteCollectorMarkdown(w io.Writer) error {
	return collectorRegistry.WriteMarkdown(w)
}

// EnabledMetrics exposes pg_collector_enabled per collector.
func EnabledMetrics() prometheus.Collector {
	return collectorRegistry.EnabledMetrics()
}

// PostgresCollector implements the prometheus.Collector interface.
type PostgresCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger

	instance          *instance
	CollectionTimeout time.Duration
}

type Option func(*PostgresCollector) error

// NewPostgresCollector creates a new PostgresCollector.
func NewPostgresCollector(logger *slog.Logger, excludeDatabases []string, dsn string, filters []string, options ...Option) (*PostgresCollector, error) {
	p := &PostgresCollector{
		logger: logger,
	}
	// Apply options to customize the collector
	for _, o := range options {
		err := o(p)
		if err != nil {
			return nil, err
		}
	}

	collectors, err := collectorRegistry.Build(func(name string) collectorConfig {
		return collectorConfig{
			logger:           logger.With("collector", name),
			excludeDatabases: excludeDatabases,
		}
	}, filters...)
	if err != nil {
		return nil, err
	}

	p.Collectors = collectors

	if dsn == "" {
		return nil, errors.New("empty dsn")
	}

	instance, err := newInstance(dsn)
	if err != nil {
		return nil, err
	}
	p.instance = instance

	return p, nil
}

func WithCollectionTimeout(s string) Option {
	return func(e *PostgresCollector) error {
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		if duration < 1*time.Millisecond {
			return errors.New("timeout must be greater than 1ms")
		}
		e.CollectionTimeout = duration
		return nil
	}
}

// Describe implements the prometheus.Collector interface.
func (p PostgresCollector) Describe(ch chan<- *prometheus.Desc) {
	scrapeReporter.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (p PostgresCollector) Collect(ch chan<- prometheus.Metric) {
	// copy the instance so that concurrent scrapes have independent instances
	inst := p.instance.copy()

	// Set up the database connection for the collector.
	err := inst.setup()
	defer inst.Close()
	if err != nil {
		p.logger.Error("Error opening connection to database", "err", err)
		return
	}
	p.collectFromConnection(inst, ch)
}

func (p PostgresCollector) collectFromConnection(inst *instance, ch chan<- prometheus.Metric) {
	// Eventually, connect this to the http scraping context
	ctx, cancel := context.WithTimeout(context.Background(), p.CollectionTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(len(p.Collectors))
	for name, c := range p.Collectors {
		go func(name string, c Collector) {
			execute(ctx, name, c, inst, ch, p.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func (p *PostgresCollector) Close() error {
	return p.instance.Close()
}

func execute(ctx context.Context, name string, c Collector, instance *instance, ch chan<- prometheus.Metric, logger *slog.Logger) {
	begin := time.Now()
	err := c.Update(ctx, instance, ch)
	duration := time.Since(begin)

	if err != nil {
		if IsNoDataError(err) {
			logger.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
	}
	scrapeReporter.Report(ch, name, duration, err)
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

func Int32(m sql.NullInt32) float64 {
	mM := 0.0
	if m.Valid {
		mM = float64(m.Int32)
	}
	return mM
}
