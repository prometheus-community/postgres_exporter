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
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/semaphore"
)

type ProbeCollector struct {
	registry   *prometheus.Registry
	collectors map[string]Collector
	logger     log.Logger
	instance   *instance
	connSema   *semaphore.Weighted
	ctx        context.Context
}

func NewProbeCollector(ctx context.Context, logger log.Logger, excludeDatabases []string, registry *prometheus.Registry, dsn config.DSN, connSema *semaphore.Weighted) (*ProbeCollector, error) {
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
					logger:           log.With(logger, "collector", key),
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

	return &ProbeCollector{
		registry:   registry,
		collectors: collectors,
		logger:     logger,
		instance:   instance,
		connSema:   connSema,
		ctx:        ctx,
	}, nil
}

func (pc *ProbeCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (pc *ProbeCollector) Collect(ch chan<- prometheus.Metric) {
	if err := pc.connSema.Acquire(pc.ctx, 1); err != nil {
		level.Warn(pc.logger).Log("msg", "Failed to acquire semaphore", "err", err)
		return
	}
	defer pc.connSema.Release(1)

	// Set up the database connection for the collector.
	err := pc.instance.setup()
	if err != nil {
		level.Error(pc.logger).Log("msg", "Error opening connection to database", "err", err)
		return
	}
	defer pc.instance.Close()

	wg := sync.WaitGroup{}
	wg.Add(len(pc.collectors))
	for name, c := range pc.collectors {
		go func(name string, c Collector) {
			execute(pc.ctx, name, c, pc.instance, ch, pc.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func (pc *ProbeCollector) Close() error {
	return pc.instance.Close()
}
