// These are specialized integration tests. We only build them when we're doing
// a lot of additional work to keep the external docker environment they require
// working.
// +build integration

package main

import (
	"os"
	"testing"

	. "gopkg.in/check.v1"

	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
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

	exporter := NewExporter(dsn, false, false, "", nil)
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

	// Open a database connection
	db, err := sql.Open("postgres", s.e.dsn)
	c.Assert(db, NotNil)
	c.Assert(err, IsNil)
	defer db.Close()

	// Do a version update
	err = s.e.checkMapVersions(ch, db)
	c.Assert(err, IsNil)

	err = querySettings(ch, db, nil)
	if !c.Check(err, Equals, nil) {
		fmt.Println("## ERRORS FOUND")
		fmt.Println(err)
	}

	// This should never happen in our test cases.
	errMap := queryNamespaceMappings(ch, db, s.e.metricMap, s.e.queryOverrides, nil)
	if !c.Check(len(errMap), Equals, 0) {
		fmt.Println("## NAMESPACE ERRORS FOUND")
		for namespace, err := range errMap {
			fmt.Println(namespace, ":", err)
		}
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
	exporter := NewExporter("invalid dsn", false, false, *queriesPath, nil)
	c.Assert(exporter, NotNil)
	exporter.scrape(ch)

	// Send a DSN to a non-listening port.
	exporter = NewExporter("postgresql://nothing:nothing@127.0.0.1:1/nothing", false, false, *queriesPath, nil)
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

	exporter := NewExporter(dsn, false, false, "", nil)
	c.Assert(exporter, NotNil)

	// Convert the default maps into a list of empty maps.
	emptyMaps := make(map[string]map[string]ColumnMapping, 0)
	for k := range exporter.builtinMetricMaps {
		emptyMaps[k] = map[string]ColumnMapping{}
	}
	exporter.builtinMetricMaps = emptyMaps

	// scrape the exporter and make sure it works
	exporter.scrape(ch)
}
