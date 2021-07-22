// Copyright 2021 The Prometheus Authors
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

package exporter

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Server describes a connection to Postgres.
// Also it contains metrics map and query overrides.
type Server struct {
	db          *sql.DB
	labels      prometheus.Labels
	master      bool
	runonserver string

	// Last version used to calculate metric map. If mismatch on scrape,
	// then maps are recalculated.
	lastMapVersion semver.Version
	// Currently active metric map
	metricMap map[string]MetricMapNamespace
	// Currently active query overrides
	queryOverrides map[string]string
	mappingMtx     sync.RWMutex
	// Currently cached metrics
	metricCache map[string]cachedMetrics
	cacheMtx    sync.Mutex

	Logger log.Logger
}

// ServerOpt configures a server.
type ServerOpt func(*Server)

// ServerWithLabels configures a set of labels.
func ServerWithLabels(labels prometheus.Labels) ServerOpt {
	return func(s *Server) {
		for k, v := range labels {
			s.labels[k] = v
		}
	}
}

// NewServer establishes a new connection using DSN.
func NewServer(dsn string, logger log.Logger, opts ...ServerOpt) (*Server, error) {
	fingerprint, err := parseFingerprint(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	level.Info(logger).Log("msg", "Established new database connection", "fingerprint", fingerprint)

	s := &Server{
		db:     db,
		master: false,
		labels: prometheus.Labels{
			serverLabelName: fingerprint,
		},
		metricCache: make(map[string]cachedMetrics),
		Logger:      logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Close disconnects from Postgres.
func (s *Server) Close() error {
	return s.db.Close()
}

// Ping checks connection availability and possibly invalidates the connection if it fails.
func (s *Server) Ping() error {
	if err := s.db.Ping(); err != nil {
		if cerr := s.Close(); cerr != nil {
			level.Error(s.Logger).Log("msg", "Error while closing non-pinging DB connection", "server", s, "err", cerr)
		}
		return err
	}
	return nil
}

// String returns server's fingerprint.
func (s *Server) String() string {
	return s.labels[serverLabelName]
}

// Scrape loads metrics.
func (s *Server) Scrape(ch chan<- prometheus.Metric, disableSettingsMetrics bool) error {
	s.mappingMtx.RLock()
	defer s.mappingMtx.RUnlock()

	var err error

	if !disableSettingsMetrics && s.master {
		if err = querySettings(ch, s); err != nil {
			err = fmt.Errorf("error retrieving settings: %s", err)
		}
	}

	errMap := queryNamespaceMappings(ch, s)
	if len(errMap) > 0 {
		err = fmt.Errorf("queryNamespaceMappings returned %d errors", len(errMap))
	}

	return err
}

// Servers contains a collection of servers to Postgres.
type Servers struct {
	m       sync.Mutex
	servers map[string]*Server
	opts    []ServerOpt
	logger  log.Logger
}

// NewServers creates a collection of servers to Postgres.
func NewServers(logger log.Logger, opts ...ServerOpt) *Servers {
	return &Servers{
		servers: make(map[string]*Server),
		opts:    opts,
		logger:  logger,
	}
}

// GetServer returns established connection from a collection.
func (s *Servers) GetServer(dsn string) (*Server, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var err error
	var ok bool
	errCount := 0 // start at zero because we increment before doing work
	retries := 1
	var server *Server
	for {
		if errCount++; errCount > retries {
			return nil, err
		}
		server, ok = s.servers[dsn]
		if !ok {
			server, err = NewServer(dsn, s.logger, s.opts...)
			if err != nil {
				time.Sleep(time.Duration(errCount) * time.Second)
				continue
			}
			s.servers[dsn] = server
		}
		if err = server.Ping(); err != nil {
			delete(s.servers, dsn)
			time.Sleep(time.Duration(errCount) * time.Second)
			continue
		}
		break
	}
	return server, nil
}

// Close disconnects from all known servers.
func (s *Servers) Close() {
	s.m.Lock()
	defer s.m.Unlock()
	for _, server := range s.servers {
		if err := server.Close(); err != nil {
			level.Error(s.logger).Log("msg", "Failed to close connection", "server", server, "err", err)
		}
	}
}
