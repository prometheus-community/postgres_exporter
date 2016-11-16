// These are specialized integration tests. We only build them when we're doing
// a lot of additional work to keep the external docker environment they require
// working.
// +build integration

package main

import (
	"os"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/prometheus/client_golang/prometheus"
	"database/sql"
	_ "github.com/lib/pq"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type IntegrationSuite struct{
	e *Exporter
}

var _ = Suite(&IntegrationSuite{})

func (s *IntegrationSuite) SetUpSuite(c *C) {
	dsn := os.Getenv("DATA_SOURCE_NAME")
	c.Assert(dsn, Not(Equals), "")

	exporter := NewExporter(dsn)
	c.Assert(exporter, NotNil)
	// Assign the exporter to the suite
	s.e = exporter

	prometheus.MustRegister(exporter)
}

func (s *IntegrationSuite) TestAllNamespacesReturnResults(c *C) {
	// Setup a dummy channel to consume metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		for _ = range ch {}
	}()

	// Open a database connection
	db, err := sql.Open("postgres", s.e.dsn)
	c.Assert(db, NotNil)
	c.Assert(err, IsNil)
	defer db.Close()

	// Check the show variables work
	nonFatalErrors := queryShowVariables(ch, db, s.e.variableMap)
	c.Check(len(nonFatalErrors), Equals, 0)

	// This should never happen in our test cases.
	errMap := queryNamespaceMappings(ch, db, s.e.metricMap)
	c.Check(len(errMap), Equals, 0)
}