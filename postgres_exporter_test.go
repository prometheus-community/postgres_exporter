// +build !integration

package main

import (
	. "gopkg.in/check.v1"
	"testing"

	"github.com/blang/semver"
	"os"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FunctionalSuite struct {
	e *Exporter
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

// test read username and password from file
func (s *FunctionalSuite) TestEnvironmentSettingWithSecretsFiles(c *C) {

	err := os.Setenv("DATA_SOURCE_USER_FILE", "./tests/username_file")
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_USER_FILE")

	err = os.Setenv("DATA_SOURCE_PASS_FILE", "./tests/userpass_file")
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_PASS_FILE")

	err = os.Setenv("DATA_SOURCE_URI", "localhost:5432/?sslmode=disable")
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_URI")

	var expected = "postgresql://custom_username:custom_password@localhost:5432/?sslmode=disable"

	dsn := getDataSource()
	if dsn != expected {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, expected)
	}
}

// test read DATA_SOURCE_NAME from environment
func (s *FunctionalSuite) TestEnvironmentSettingWithDns(c *C) {

	envDsn := "postgresql://user:password@localhost:5432/?sslmode=enabled"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_NAME")

	dsn := getDataSource()
	if dsn != envDsn {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, envDsn)
	}
}

// test DATA_SOURCE_NAME is used even if username and password environment variables are set
func (s *FunctionalSuite) TestEnvironmentSettingWithDnsAndSecrets(c *C) {

	envDsn := "postgresql://userDsn:passwordDsn@localhost:55432/?sslmode=disabled"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_NAME")

	err = os.Setenv("DATA_SOURCE_USER_FILE", "./tests/username_file")
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_USER_FILE")

	err = os.Setenv("DATA_SOURCE_PASS", "envUserPass")
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_PASS")

	dsn := getDataSource()
	if dsn != envDsn {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, envDsn)
	}
}

func UnsetEnvironment(c *C, d string) {
	err := os.Unsetenv(d)
	c.Assert(err, IsNil)
}
