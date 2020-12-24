// These are specialized integration tests. We only build them when we're doing
// a lot of additional work to keep the external docker environment they require
// working.
// +build integration

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

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

func (s *IntegrationSuite) SetUpSuite(c *C) {
	dsn := os.Getenv("DATA_SOURCE_NAME")
	c.Assert(dsn, Not(Equals), "")

	exporter := NewExporter(strings.Split(dsn, ","))
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

	for _, dsn := range s.e.dsn {
		// Open a database connection
		server, err := NewServer(dsn)
		c.Assert(server, NotNil)
		c.Assert(err, IsNil)

		// Do a version update
		err = s.e.checkMapVersions(ch, server)
		c.Assert(err, IsNil)

		err = querySettings(ch, server)
		if !c.Check(err, Equals, nil) {
			fmt.Println("## ERRORS FOUND")
			fmt.Println(err)
		}

		// This should never happen in our test cases.
		errMap := queryNamespaceMappings(ch, server)
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
// the exporter. Related to https://github.com/wrouesnel/postgres_exporter/issues/93
// although not a replication of the scenario.
func (s *IntegrationSuite) TestInvalidDsnDoesntCrash(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for range ch {
		}
	}()

	// Send a bad DSN
	exporter := NewExporter([]string{"invalid dsn"})
	c.Assert(exporter, NotNil)
	exporter.scrape(ch)

	// Send a DSN to a non-listening port.
	exporter = NewExporter([]string{"postgresql://nothing:nothing@127.0.0.1:1/nothing"})
	c.Assert(exporter, NotNil)
	exporter.scrape(ch)
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

	exporter := NewExporter(strings.Split(dsn, ","))
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
	exporter.scrape(ch)
}
