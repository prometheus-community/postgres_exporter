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
	"log/slog"
	"sync"

	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

// multiInstanceCollector fans PostgresCollector out across every DSN
// currently reported by targets. A DSN marked exporter.Target.Primary (one
// that was explicitly configured, rather than produced by database
// autodiscovery) gets every enabled collector, matching the exporter's
// historical single-DSN behavior. Any other DSN only gets collectors
// registered with databaseScope: instance-wide (clusterScope) collectors
// would otherwise report the exact same rows once per additional database
// and panic the registry with a duplicate-metric error.
//
// The collector set is reconciled against the target list on every Collect
// call, so databases created or dropped after startup are picked up without
// an exporter restart, mirroring the dynamic behavior database autodiscovery
// already has for the legacy exporter.Exporter metrics.
type multiInstanceCollector struct {
	logger  *slog.Logger
	targets func() []exporter.Target
	build   func(dsn string, primary bool) (*PostgresCollector, error)

	mu         sync.Mutex
	collectors map[string]*PostgresCollector
}

func newMultiInstanceCollector(logger *slog.Logger, targets func() []exporter.Target, build func(dsn string, primary bool) (*PostgresCollector, error)) *multiInstanceCollector {
	return &multiInstanceCollector{
		logger:     logger,
		targets:    targets,
		build:      build,
		collectors: make(map[string]*PostgresCollector),
	}
}

// init eagerly builds collectors for the current primary targets, so startup
// still fails fast on a bad DSN, exactly as it did before autodiscovery could
// fan a single Runtime out into multiple PostgresCollectors. Non-primary
// (discovered) targets are best-effort here and are otherwise built lazily
// from Collect.
func (m *multiInstanceCollector) init() error {
	targets := m.targets()

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range targets {
		pc, err := m.build(t.DSN, t.Primary)
		if err != nil {
			if t.Primary {
				return err
			}
			m.logger.Error("failed to build collector for discovered database", "dsn", exporter.LoggableDSN(t.DSN), "err", err)
			continue
		}
		m.collectors[t.DSN] = pc
	}
	return nil
}

// Describe implements prometheus.Collector.
func (m *multiInstanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements prometheus.Collector.
func (m *multiInstanceCollector) Collect(ch chan<- prometheus.Metric) {
	snapshot := m.refresh()

	var wg sync.WaitGroup
	wg.Add(len(snapshot))
	for _, pc := range snapshot {
		go func(pc *PostgresCollector) {
			defer wg.Done()
			pc.Collect(ch)
		}(pc)
	}
	wg.Wait()
}

// refresh reconciles the cached collectors against the current target list:
// it builds collectors for newly seen DSNs, closes collectors for DSNs that
// disappeared (e.g. a database was dropped), and returns a snapshot to
// scrape.
func (m *multiInstanceCollector) refresh() []*PostgresCollector {
	targets := m.targets()

	m.mu.Lock()
	defer m.mu.Unlock()

	next := make(map[string]*PostgresCollector, len(targets))
	for _, t := range targets {
		if pc, ok := m.collectors[t.DSN]; ok {
			next[t.DSN] = pc
			continue
		}
		pc, err := m.build(t.DSN, t.Primary)
		if err != nil {
			m.logger.Error("failed to build collector for target", "dsn", exporter.LoggableDSN(t.DSN), "err", err)
			continue
		}
		next[t.DSN] = pc
	}

	for dsn, pc := range m.collectors {
		if _, ok := next[dsn]; !ok {
			if err := pc.Close(); err != nil {
				m.logger.Error("failed to close collector for removed target", "dsn", exporter.LoggableDSN(dsn), "err", err)
			}
		}
	}

	m.collectors = next

	snapshot := make([]*PostgresCollector, 0, len(next))
	for _, pc := range next {
		snapshot = append(snapshot, pc)
	}
	return snapshot
}

// Close closes every collector currently cached.
func (m *multiInstanceCollector) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var err error
	for _, pc := range m.collectors {
		err = errors.Join(err, pc.Close())
	}
	return err
}
