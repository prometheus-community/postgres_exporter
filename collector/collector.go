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

// collectorScope describes whether a collector's data is specific to the
// database it is connected to, or shared across the whole PostgreSQL server.
type collectorScope int

const (
	// serverScope collectors read catalog views or functions that report on
	// the whole PostgreSQL server (e.g. pg_stat_activity, pg_stat_bgwriter,
	// pg_stat_statements) regardless of which database the connection happens
	// to be attached to. They must only ever be run once per server: running
	// them again from a second connection to a different database on the same
	// server would re-report the exact same rows and panic the registry with
	// a duplicate-metric error.
	serverScope collectorScope = iota
	// databaseScope collectors read catalog views that are scoped to the
	// currently connected database (e.g. pg_stat_user_tables). They report
	// different data depending on which database the connection is attached
	// to, and must be run once per database to get complete coverage.
	databaseScope
)

type registeredCollector struct {
	scope  collectorScope
	create func(collectorConfig) (Collector, error)
}

var factories = make(map[string]registeredCollector)

// Namespace for all metrics.
const namespace = "pg"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"postgres_exporter: Duration of a collector scrape.",
		[]string{"collector", "datname"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"postgres_exporter: Whether a collector succeeded.",
		[]string{"collector", "datname"},
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

func registerCollector(name string, scope collectorScope, createFunc func(collectorConfig) (Collector, error)) {
	if _, ok := config.DefaultCollectorConfig()[name]; !ok {
		panic(fmt.Sprintf("collector %q is not declared in config.DefaultCollectorConfig", name))
	}
	factories[name] = registeredCollector{scope: scope, create: createFunc}
}

// PostgresCollector implements the prometheus.Collector interface.
type PostgresCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger

	instance                *instance
	CollectionTimeout       time.Duration
	collectorStates         map[string]bool
	excludeDatabases        []string
	longRunningTransactions config.LongRunningTransactionsConfig
	pgStatStatements        config.PGStatStatementsConfig
	wrapLargeCounters       bool
	databaseDiscovery       *databaseDiscoveryConfig
}

// databaseDiscoveryConfig holds the include/exclude filters for
// WithDatabaseDiscovery. Its presence on a PostgresCollector (non-nil) is
// what enables discovery; it carries no other state; the databases it names
// are discovered fresh on every scrape.
type databaseDiscoveryConfig struct {
	includeDatabases []string
	excludeDatabases []string
}

type Option func(*PostgresCollector) error

// NewPostgresCollector creates a new PostgresCollector.
func NewPostgresCollector(logger *slog.Logger, excludeDatabases []string, dsn string, filters []string, options ...Option) (*PostgresCollector, error) {
	p := &PostgresCollector{
		logger:                  logger,
		excludeDatabases:        excludeDatabases,
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
		collector, err := factory.create(collectorConfig{
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

// WithDatabaseDiscovery makes the collector discover every other database on
// the same PostgreSQL server as its primary connection, and run the
// database-scoped collectors against each one over its own connection, in
// addition to the primary database. The set of databases is re-evaluated on
// every scrape, so databases created or dropped after startup are picked up
// without a restart.
//
// Server-scoped collectors are unaffected by this option: they still only
// ever run once, against the primary connection, so they are never
// duplicated no matter how many databases are discovered.
func WithDatabaseDiscovery(includeDatabases, excludeDatabases []string) Option {
	return func(p *PostgresCollector) error {
		p.databaseDiscovery = &databaseDiscoveryConfig{
			includeDatabases: includeDatabases,
			excludeDatabases: excludeDatabases,
		}
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
	p.collectFromConnection(inst, ch, nil)

	if p.databaseDiscovery != nil {
		p.collectDiscoveredDatabases(inst, ch)
	}
}

// collectDiscoveredDatabases discovers every other database on the same
// PostgreSQL server as primary, then runs the database-scoped Collectors
// against each one over its own connection. The set of databases is
// rediscovered on every call, so databases created or dropped after startup
// are picked up without a restart; no state is kept between calls.
func (p PostgresCollector) collectDiscoveredDatabases(primary *instance, ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), p.CollectionTimeout)
	defer cancel()

	databases, err := discoverDatabases(ctx, primary.getDB(), p.databaseDiscovery.includeDatabases, p.databaseDiscovery.excludeDatabases)
	if err != nil {
		p.logger.Error("failed to discover databases", "err", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))
	for _, database := range databases {
		go func(database string) {
			defer wg.Done()

			inst, err := primary.withDatabase(database)
			if err != nil {
				p.logger.Error("failed to build connection for discovered database", "database", database, "err", err)
				return
			}
			defer inst.Close()
			if err := inst.setup(); err != nil {
				p.logger.Error("failed to connect to discovered database", "database", database, "err", err)
				return
			}
			scope := databaseScope
			p.collectFromConnection(inst, ch, &scope)
		}(database)
	}
	wg.Wait()
}

// collectFromConnection runs the Collectors matching scope against inst,
// sending their metrics to ch. A nil scope runs every Collector.
func (p PostgresCollector) collectFromConnection(inst *instance, ch chan<- prometheus.Metric, scope *collectorScope) {
	// Eventually, connect this to the http scraping context
	ctx, cancel := context.WithTimeout(context.Background(), p.CollectionTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	for name, c := range p.Collectors {
		if scope != nil && factories[name].scope != *scope {
			continue
		}
		wg.Add(1)
		go func(name string, c Collector) {
			defer wg.Done()
			execute(ctx, name, c, inst, ch, p.logger)
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
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name, instance.database)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name, instance.database)
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
