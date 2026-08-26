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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/internal/metricutil"
	"github.com/prometheus/client_golang/prometheus"
)

var factories = make(map[string]func(collectorConfig) (Collector, error))

// Namespace for all metrics.
const namespace = "pg"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"postgres_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"postgres_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

type Collector interface {
	Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error
}

type collectorConfig struct {
	logger                        *slog.Logger
	excludeDatabases              []string
	longRunningTransactionsConfig config.LongRunningTransactionsConfig
	pgStatStatementsConfig        config.PGStatStatementsConfig
}

func registerCollector(name string, createFunc func(collectorConfig) (Collector, error)) {
	if _, ok := config.DefaultCollectorConfig()[name]; !ok {
		panic(fmt.Sprintf("collector %q is not declared in config.DefaultCollectorConfig", name))
	}
	factories[name] = createFunc
}

// PostgresCollector implements the prometheus.Collector interface.
type PostgresCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger

	instance                *instance
	CollectionTimeout       time.Duration
	collectorStates         map[string]bool
	longRunningTransactions config.LongRunningTransactionsConfig
	pgStatStatements        config.PGStatStatementsConfig
	wrapLargeCounters       bool
}

type Option func(*PostgresCollector) error

// NewPostgresCollector creates a new PostgresCollector.
func NewPostgresCollector(logger *slog.Logger, excludeDatabases []string, dsn string, filters []string, options ...Option) (*PostgresCollector, error) {
	p := &PostgresCollector{
		logger:                  logger,
		collectorStates:         config.DefaultCollectorConfig(),
		longRunningTransactions: defaultLongRunningTransactionsConfig(),
		pgStatStatements:        defaultPGStatStatementsConfig(),
		CollectionTimeout:       time.Minute,
		wrapLargeCounters:       true,
	}
	// Apply options to customize the collector
	for _, o := range options {
		err := o(p)
		if err != nil {
			return nil, err
		}
	}

	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := p.collectorStates[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	for key, enabled := range p.collectorStates {
		if !enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		factory, ok := factories[key]
		if !ok {
			return nil, fmt.Errorf("missing collector factory: %s", key)
		}
		collector, err := factory(collectorConfig{
			logger:                        logger.With("collector", key),
			excludeDatabases:              excludeDatabases,
			longRunningTransactionsConfig: p.longRunningTransactions,
			pgStatStatementsConfig:        p.pgStatStatements,
		})
		if err != nil {
			return nil, err
		}
		collectors[key] = collector
	}

	p.Collectors = collectors

	if dsn == "" {
		return nil, errors.New("empty dsn")
	}

	instance, err := newInstance(dsn)
	if err != nil {
		return nil, err
	}
	instance.wrapLargeCounters = p.wrapLargeCounters
	p.instance = instance

	return p, nil
}

func WithCollectorStates(states map[string]bool) Option {
	return func(e *PostgresCollector) error {
		merged := config.DefaultCollectorConfig()
		for name, enabled := range states {
			if _, ok := factories[name]; !ok {
				return fmt.Errorf("missing collector: %s", name)
			}
			merged[name] = enabled
		}
		e.collectorStates = merged
		return nil
	}
}

func WithPGStatStatementsConfig(cfg config.PGStatStatementsConfig) Option {
	return func(e *PostgresCollector) error {
		e.pgStatStatements = withPGStatStatementsDefaults(cfg)
		return nil
	}
}

func WithLongRunningTransactionsConfig(cfg config.LongRunningTransactionsConfig) Option {
	return func(e *PostgresCollector) error {
		e.longRunningTransactions = cfg
		return nil
	}
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

// WithWrapLargeCounters configures wrapping 64-bit counters at 2^53.
func WithWrapLargeCounters(wrap bool) Option {
	return func(p *PostgresCollector) error {
		p.wrapLargeCounters = wrap
		return nil
	}
}

// Describe implements the prometheus.Collector interface.
func (p PostgresCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
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
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			logger.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
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

func int64CounterValue(value sql.NullInt64, wrap bool) float64 {
	if !value.Valid {
		return 0
	}

	return metricutil.Int64CounterValue(value.Int64, wrap)
}
