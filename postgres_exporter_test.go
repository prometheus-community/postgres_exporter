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

func TestEnvironmentSettingWithSecretsFiles(t *testing.T) {

	err := os.Setenv("DATA_SOURCE_USER_FILE", "./tests/username_file")
	if err != nil {
		t.Errorf("DATA_SOURCE_USER_FILE could not be set")
	}

	err = os.Setenv("DATA_SOURCE_PASS_FILE", "./tests/userpass_file")
	if err != nil {
		t.Errorf("DATA_SOURCE_PASS_FILE could not be set")
	}

	err = os.Setenv("DATA_SOURCE_URI", "localhost:5432/?sslmode=disable")
	if err != nil {
		t.Errorf("DATA_SOURCE_URI could not be set")
	}

	var expected = "postgresql://custom_username:custom_password@localhost:5432/?sslmode=disable"

	dsn := getDataSource()
	if dsn != expected {
		t.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, expected)
	}
}

func TestEnvironmentSettingWithDns(t *testing.T) {

	envDsn := "postgresql://user:password@localhost:5432/?sslmode=enabled"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	if err != nil {
		t.Errorf("DATA_SOURCE_NAME could not be set")
	}

	dsn := getDataSource()
	if dsn != envDsn {
		t.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, envDsn)
	}
}

// test DATA_SOURCE_NAME is used even if username and password environment wariables are set
func TestEnvironmentSettingWithDnsAndSecrets(t *testing.T) {

	envDsn := "postgresql://userDsn:passwordDsn@localhost:55432/?sslmode=disabled"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	if err != nil {
		t.Errorf("DATA_SOURCE_NAME could not be set")
	}

	err = os.Setenv("DATA_SOURCE_USER_FILE", "./tests/username_file")
	if err != nil {
		t.Errorf("DATA_SOURCE_USER_FILE could not be set")
	}

	err = os.Setenv("DATA_SOURCE_PASS", "envUserPass")
	if err != nil {
		t.Errorf("DATA_SOURCE_PASS could not be set")
	}

	dsn := getDataSource()
	if dsn != envDsn {
		t.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn, envDsn)
	}
}
