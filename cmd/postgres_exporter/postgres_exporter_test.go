// +build !integration

package main

import (
	"testing"

	. "gopkg.in/check.v1"

	"os"

	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
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
		resultMap := makeDescMap(semver.MustParse("0.0.1"), prometheus.Labels{}, testMetricMap)
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

	// nolint: dupl
	{
		// Update the map so the discard metric should be eliminated
		discardableMetric := testMetricMap["test_namespace"]["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"]["metric_which_discards"] = discardableMetric

		// Discard metric should be discarded
		resultMap := makeDescMap(semver.MustParse("0.0.1"), prometheus.Labels{}, testMetricMap)
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

	// nolint: dupl
	{
		// Update the map so the discard metric should be kept but has a version
		discardableMetric := testMetricMap["test_namespace"]["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"]["metric_which_discards"] = discardableMetric

		// Discard metric should be discarded
		resultMap := makeDescMap(semver.MustParse("0.0.2"), prometheus.Labels{}, testMetricMap)
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

	var expected = "postgresql://custom_username$&+,%2F%3A;=%3F%40:custom_password$&+,%2F%3A;=%3F%40@localhost:5432/?sslmode=disable"

	dsn := getDataSources()
	if len(dsn) == 0 {
		c.Errorf("Expected one data source, zero found")
	}
	if dsn[0] != expected {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn[0], expected)
	}
}

// test read DATA_SOURCE_NAME from environment
func (s *FunctionalSuite) TestEnvironmentSettingWithDns(c *C) {
	envDsn := "postgresql://user:password@localhost:5432/?sslmode=enabled"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_NAME")

	dsn := getDataSources()
	if len(dsn) == 0 {
		c.Errorf("Expected one data source, zero found")
	}
	if dsn[0] != envDsn {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn[0], envDsn)
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

	dsn := getDataSources()
	if len(dsn) == 0 {
		c.Errorf("Expected one data source, zero found")
	}
	if dsn[0] != envDsn {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn[0], envDsn)
	}
}

func (s *FunctionalSuite) TestPostgresVersionParsing(c *C) {
	type TestCase struct {
		input    string
		expected string
	}

	cases := []TestCase{
		{
			input:    "PostgreSQL 10.1 on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18) 6.3.0 20170516, 64-bit",
			expected: "10.1.0",
		},
		{
			input:    "PostgreSQL 9.5.4, compiled by Visual C++ build 1800, 64-bit",
			expected: "9.5.4",
		},
		{
			input:    "EnterpriseDB 9.6.5.10 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 4.4.7 20120313 (Red Hat 4.4.7-16), 64-bit",
			expected: "9.6.5",
		},
	}

	for _, cs := range cases {
		ver, err := parseVersion(cs.input)
		c.Assert(err, IsNil)
		c.Assert(ver.String(), Equals, cs.expected)
	}
}

func UnsetEnvironment(c *C, d string) {
	err := os.Unsetenv(d)
	c.Assert(err, IsNil)
}
