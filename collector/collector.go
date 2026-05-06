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
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	factories          = make(map[string]func(collectorConfig) (Collector, error))
	collectorMetadata  = make(map[string]CollectorMetadata)
	collectorConfigMtx = sync.Mutex{}
)

const (
	// Namespace for all metrics.
	namespace = "pg"

	CollectorFlagPrefix = "collector."
	defaultEnabled      = true
	defaultDisabled     = false
)

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

type CollectorMetadata struct {
	Name           string
	DefaultEnabled bool
}

type collectorConfig struct {
	logger               *slog.Logger
	excludeDatabases     []string
	statStatementsConfig statStatementsConfig
}

type collectorOptions struct {
	collectionTimeout    time.Duration
	collectorStates      map[string]bool
	statStatementsConfig statStatementsConfig
}

type statStatementsConfig struct {
	includeQueryStatement bool
	statementLength       uint
	statementLimit        uint
	excludedDatabases     []string
	excludedUsers         []string
}

func registerCollector(name string, isDefaultEnabled bool, createFunc func(collectorConfig) (Collector, error)) {
	collectorConfigMtx.Lock()
	defer collectorConfigMtx.Unlock()
	collectorMetadata[name] = CollectorMetadata{Name: name, DefaultEnabled: isDefaultEnabled}
	factories[name] = createFunc
}

func Collectors() []CollectorMetadata {
	collectorConfigMtx.Lock()
	defer collectorConfigMtx.Unlock()
	collectors := make([]CollectorMetadata, 0, len(collectorMetadata))
	for _, metadata := range collectorMetadata {
		collectors = append(collectors, metadata)
	}
	sort.Slice(collectors, func(i, j int) bool {
		return collectors[i].Name < collectors[j].Name
	})
	return collectors
}

func defaultCollectorStates() map[string]bool {
	states := make(map[string]bool, len(collectorMetadata))
	for name, metadata := range collectorMetadata {
		states[name] = metadata.DefaultEnabled
	}
	return states
}

func newDefaultCollectorOptions() collectorOptions {
	return collectorOptions{
		statStatementsConfig: statStatementsConfig{
			statementLength: DefaultStatStatementsQueryLength,
			statementLimit:  DefaultStatStatementsLimit,
		},
	}
}

// PostgresCollector implements the prometheus.Collector interface.
type PostgresCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger

	instance          *instance
	CollectionTimeout time.Duration
}

type Option func(*collectorOptions) error

// NewPostgresCollector creates a new PostgresCollector.
func NewPostgresCollector(logger *slog.Logger, excludeDatabases []string, dsn string, filters []string, options ...Option) (*PostgresCollector, error) {
	p := &PostgresCollector{
		logger: logger,
	}
	collectorOptions := newDefaultCollectorOptions()
	for _, o := range options {
		err := o(&collectorOptions)
		if err != nil {
			return nil, err
		}
	}
	p.CollectionTimeout = collectorOptions.collectionTimeout

	collectors, err := newCollectors(logger, excludeDatabases, filters, collectorOptions)
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
	return func(o *collectorOptions) error {
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		if duration < 1*time.Millisecond {
			return errors.New("timeout must be greater than 1ms")
		}
		o.collectionTimeout = duration
		return nil
	}
}

func WithCollectorStates(states map[string]bool) Option {
	return func(o *collectorOptions) error {
		collectorConfigMtx.Lock()
		defer collectorConfigMtx.Unlock()

		o.collectorStates = make(map[string]bool, len(states))
		for name, enabled := range states {
			if _, ok := factories[name]; !ok {
				return fmt.Errorf("missing collector: %s", name)
			}
			o.collectorStates[name] = enabled
		}
		return nil
	}
}

func WithStatStatementsConfig(includeQuery bool, queryLength uint, limit uint, excludedDatabases []string, excludedUsers []string) Option {
	return func(o *collectorOptions) error {
		o.statStatementsConfig = statStatementsConfig{
			includeQueryStatement: includeQuery,
			statementLength:       queryLength,
			statementLimit:        limit,
			excludedDatabases:     append([]string(nil), excludedDatabases...),
			excludedUsers:         append([]string(nil), excludedUsers...),
		}
		return nil
	}
}

func newCollectors(logger *slog.Logger, excludeDatabases []string, filters []string, options collectorOptions) (map[string]Collector, error) {
	collectorConfigMtx.Lock()
	defer collectorConfigMtx.Unlock()

	states := defaultCollectorStates()
	for name, enabled := range options.collectorStates {
		states[name] = enabled
	}

	filtered := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := states[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		filtered[filter] = true
	}

	collectors := make(map[string]Collector)
	for key, enabled := range states {
		if !enabled || (len(filtered) > 0 && !filtered[key]) {
			continue
		}
		collector, err := factories[key](collectorConfig{
			logger:               logger.With("collector", key),
			excludeDatabases:     excludeDatabases,
			statStatementsConfig: options.statStatementsConfig,
		})
		if err != nil {
			return nil, err
		}
		collectors[key] = collector
	}
	return collectors, nil
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
