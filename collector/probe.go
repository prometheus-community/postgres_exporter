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
	"log/slog"
	"sync"

	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

type ProbeCollector struct {
	registry   *prometheus.Registry
	collectors map[string]Collector
	logger     *slog.Logger
	instance   *instance
}

// ProbeOption configures a ProbeCollector.
type ProbeOption func(*ProbeCollector)

// WithProbeAuroraEnabled mirrors WithAuroraEnabled for the multi-target
// (/probe) code path: it propagates the global --aurora.enabled flag down
// to the per-request instance so detectAurora() runs and aurora_* collectors
// actually emit metrics instead of silently returning ErrNoData.
func WithProbeAuroraEnabled(enabled bool) ProbeOption {
	return func(pc *ProbeCollector) {
		pc.instance.auroraSupportEnabled = enabled
	}
}

func NewProbeCollector(logger *slog.Logger, excludeDatabases []string, registry *prometheus.Registry, dsn config.DSN, opts ...ProbeOption) (*ProbeCollector, error) {
	collectors := make(map[string]Collector)
	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()
	for key, enabled := range collectorState {
		// TODO: Handle filters
		// if !*enabled || (len(f) > 0 && !f[key]) {
		// 	continue
		// }
		if !*enabled {
			continue
		}
		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](
				collectorConfig{
					logger:           logger.With("collector", key),
					excludeDatabases: excludeDatabases,
				})
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}

	instance, err := newInstance(dsn.GetConnectionString())
	if err != nil {
		return nil, err
	}

	pc := &ProbeCollector{
		registry:   registry,
		collectors: collectors,
		logger:     logger,
		instance:   instance,
	}
	for _, o := range opts {
		o(pc)
	}
	return pc, nil
}

func (pc *ProbeCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (pc *ProbeCollector) Collect(ch chan<- prometheus.Metric) {
	// Set up the database connection for the collector.
	err := pc.instance.setup()
	defer pc.instance.Close()
	if err != nil {
		pc.logger.Error("Error opening connection to database", "err", err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pc.collectors))
	for name, c := range pc.collectors {
		go func(name string, c Collector) {
			execute(context.TODO(), name, c, pc.instance, ch, pc.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func (pc *ProbeCollector) Close() error {
	return pc.instance.Close()
}
