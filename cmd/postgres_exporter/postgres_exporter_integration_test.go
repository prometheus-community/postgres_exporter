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

// These are specialized integration tests. We only build them when we're doing
// a lot of additional work to keep the external docker environment they require
// working.
// +build integration

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type IntegrationSuite struct {
	e *Exporter
}

var _ = Suite(&IntegrationSuite{})

var testScrapeDuration = 10 * time.Second

func (s *IntegrationSuite) SetUpSuite(c *C) {
	dsn := os.Getenv("DATA_SOURCE_NAME")
	c.Assert(dsn, Not(Equals), "")

	exporter := NewExporter(strings.Split(dsn, ","), &testScrapeDuration)
	c.Assert(exporter, NotNil)
	// Assign the exporter to the suite
	s.e = exporter

	prometheus.MustRegister(exporter)
}

// TODO: it would be nice if cu didn't mostly just recreate the scrape function
func (s *IntegrationSuite) TestAllNamespacesReturnResults(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for range ch {
		}
	}()

	ctx := context.Background()

	for _, dsn := range s.e.dsn {
		// Open a database connection
		server, err := NewServer(dsn)
		c.Assert(server, NotNil)
		c.Assert(err, IsNil)

		// Do a version update
		err = s.e.checkMapVersions(ctx, ch, server)
		c.Assert(err, IsNil)

		err = querySettings(ctx, ch, server)
		if !c.Check(err, Equals, nil) {
			fmt.Println("## ERRORS FOUND")
			fmt.Println(err)
		}

		// This should never happen in our test cases.
		errMap := queryNamespaceMappings(ctx, ch, server)
		if !c.Check(len(errMap), Equals, 0) {
			fmt.Println("## NAMESPACE ERRORS FOUND")
			for namespace, err := range errMap {
				fmt.Println(namespace, ":", err)
			}
		}
		server.Close()
	}
}

// TestInvalidDsnDoesntCrash tests that specifying an invalid DSN doesn't crash
// the exporter. Related to https://github.com/prometheus-community/postgres_exporter/issues/93
// although not a replication of the scenario.
func (s *IntegrationSuite) TestInvalidDsnDoesntCrash(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for range ch {
		}
	}()

	// Send a bad DSN
	ctx := context.Background()
	exporter := NewExporter([]string{"invalid dsn"}, &testScrapeDuration)
	c.Assert(exporter, NotNil)
	exporter.scrape(ctx, ch)

	// Send a DSN to a non-listening port.
	exporter = NewExporter([]string{"postgresql://nothing:nothing@127.0.0.1:1/nothing"}, &testScrapeDuration)
	c.Assert(exporter, NotNil)
	exporter.scrape(ctx, ch)
}

// TestUnknownMetricParsingDoesntCrash deliberately deletes all the column maps out
// of an exporter to test that the default metric handling code can cope with unknown columns.
func (s *IntegrationSuite) TestUnknownMetricParsingDoesntCrash(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for range ch {
		}
	}()

	dsn := os.Getenv("DATA_SOURCE_NAME")
	c.Assert(dsn, Not(Equals), "")

	exporter := NewExporter(strings.Split(dsn, ","), &testScrapeDuration)
	c.Assert(exporter, NotNil)

	// Convert the default maps into a list of empty maps.
	emptyMaps := make(map[string]intermediateMetricMap, 0)
	for k := range exporter.builtinMetricMaps {
		emptyMaps[k] = intermediateMetricMap{
			map[string]ColumnMapping{},
			true,
			0,
		}
	}
	exporter.builtinMetricMaps = emptyMaps

	// scrape the exporter and make sure it works
	exporter.scrape(context.Background(), ch)
}

// TestExtendQueriesDoesntCrash tests that specifying extend.query-path doesn't
// crash.
func (s *IntegrationSuite) TestExtendQueriesDoesntCrash(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for range ch {
		}
	}()

	dsn := os.Getenv("DATA_SOURCE_NAME")
	c.Assert(dsn, Not(Equals), "")

	exporter := NewExporter(
		strings.Split(dsn, ","), &testScrapeDuration,
		WithUserQueriesPath("../user_queries_test.yaml"),
	)
	c.Assert(exporter, NotNil)

	// scrape the exporter and make sure it works
	exporter.scrape(context.Background(), ch)
}

func (s *IntegrationSuite) TestAutoDiscoverDatabases(c *C) {
	dsn := os.Getenv("DATA_SOURCE_NAME")

	exporter := NewExporter(
		strings.Split(dsn, ","),
		&testScrapeDuration,
	)
	c.Assert(exporter, NotNil)

	dsns := exporter.discoverDatabaseDSNs(context.Background())

	c.Assert(len(dsns), Equals, 2)
}
