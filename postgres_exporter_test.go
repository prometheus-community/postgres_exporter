// +build !integration

package main

import (
	. "gopkg.in/check.v1"
	"testing"

	"github.com/blang/semver"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FunctionalSuite struct {
}

var _ = Suite(&FunctionalSuite{})

func (s *FunctionalSuite) SetUpSuite(c *C) {

}

func (s *FunctionalSuite) TestSemanticVersionColumnDiscard(c *C) {
	testMetricMap := map[string]map[string]ColumnMapping{
		"test_namespace": {
			"metric_which_stays":    {COUNTER, "This metric should not be eliminated", nil, nil},
			"metric_which_discards": {COUNTER, "This metric should be forced to DISCARD", nil, nil},
		},
	}

	{
		// No metrics should be eliminated
		resultMap := makeDescMap(semver.MustParse("0.0.1"), testMetricMap)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_stays"].discard,
			Equals,
			false,
		)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_discards"].discard,
			Equals,
			false,
		)
	}

	{
		// Update the map so the discard metric should be eliminated
		discardableMetric := testMetricMap["test_namespace"]["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"]["metric_which_discards"] = discardableMetric

		// Discard metric should be discarded
		resultMap := makeDescMap(semver.MustParse("0.0.1"), testMetricMap)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_stays"].discard,
			Equals,
			false,
		)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_discards"].discard,
			Equals,
			true,
		)
	}

	{
		// Update the map so the discard metric should be kept but has a version
		discardableMetric := testMetricMap["test_namespace"]["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"]["metric_which_discards"] = discardableMetric

		// Discard metric should be discarded
		resultMap := makeDescMap(semver.MustParse("0.0.2"), testMetricMap)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_stays"].discard,
			Equals,
			false,
		)
		c.Check(
			resultMap["test_namespace"].columnMappings["metric_which_discards"].discard,
			Equals,
			false,
		)
	}
}
