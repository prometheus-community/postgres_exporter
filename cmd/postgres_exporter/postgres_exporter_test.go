//go:build !integration
// +build !integration

package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FunctionalSuite struct{}

var _ = Suite(&FunctionalSuite{})

func (s *FunctionalSuite) SetUpSuite(c *C) {
}

func (s *FunctionalSuite) TestSemanticVersionColumnDiscard(c *C) {
	testMetricMap := map[string]intermediateMetricMap{
		"test_namespace": {
			map[string]ColumnMapping{
				"metric_which_stays":    {COUNTER, "This metric should not be eliminated", nil, nil},
				"metric_which_discards": {COUNTER, "This metric should be forced to DISCARD", nil, nil},
			},
			true,
			0,
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
		discardableMetric := testMetricMap["test_namespace"].columnMappings["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"].columnMappings["metric_which_discards"] = discardableMetric

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
		discardableMetric := testMetricMap["test_namespace"].columnMappings["metric_which_discards"]
		discardableMetric.supportedVersions = semver.MustParseRange(">0.0.1")
		testMetricMap["test_namespace"].columnMappings["metric_which_discards"] = discardableMetric

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

	expected := "postgresql://custom_username$&+,%2F%3A;=%3F%40:custom_password$&+,%2F%3A;=%3F%40@localhost:5432/?sslmode=disable"

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

// test DSN including SSL enabled
func (s *FunctionalSuite) TestSSL(c *C) {
	// Thr driver doesn't support sslmode=prefer
	envDsn := "host=127.0.0.1 port=5433 user=root password=root database=postgres sslmode=require"
	err := os.Setenv("DATA_SOURCE_NAME", envDsn)
	c.Assert(err, IsNil)
	defer UnsetEnvironment(c, "DATA_SOURCE_NAME")

	server, err := NewServer(envDsn)
	c.Assert(err, IsNil)
	c.Assert(server, NotNil)

	// Just to check getDataSources can handle ssl options
	dsn := getDataSources()
	if len(dsn) == 0 {
		c.Errorf("Expected one data source, zero found")
	}
	if dsn[0] != envDsn {
		c.Errorf("Expected Username to be read from file. Found=%v, expected=%v", dsn[0], envDsn)
	}

	type row struct {
		Pid         int    // pid
		Ssl         bool   // ssl
		Version     string // version
		Cipher      string // cipher
		Bits        int    // bits
		Compression bool   // compression
		Clientdn    string // clientdn
	}
	res := row{}

	db, err := sql.Open("postgres", dsn[0])
	c.Assert(err, IsNil)
	version, err := postgresVersion(db)
	c.Assert(err, IsNil)
	if version < 90500 {
		c.Skip("This test needs PostgreSQL 9.5+")
	}

	// This query returns information about the current connection. Ssl field should be true
	sqlstr := "SELECT pid, ssl, version, cipher, bits, compression FROM pg_stat_ssl WHERE pid=pg_backend_pid()"
	err = db.QueryRow(sqlstr).Scan(&res.Pid, &res.Ssl, &res.Version, &res.Cipher, &res.Bits, &res.Compression)
	c.Assert(err, IsNil)
	c.Assert(res.Ssl, Equals, true)
	c.Assert(db.Close(), IsNil)
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

func (s *FunctionalSuite) TestParseFingerprint(c *C) {
	cases := []struct {
		url         string
		fingerprint string
		err         string
	}{
		{
			url:         "postgresql://userDsn:passwordDsn@localhost:55432/?sslmode=disabled",
			fingerprint: "localhost:55432",
		},
		{
			url:         "postgresql://userDsn:passwordDsn%3D@localhost:55432/?sslmode=disabled",
			fingerprint: "localhost:55432",
		},
		{
			url:         "port=1234",
			fingerprint: "localhost:1234",
		},
		{
			url:         "host=example",
			fingerprint: "example:5432",
		},
		{
			url: "xyz",
			err: "malformed dsn \"xyz\"",
		},
	}

	for _, cs := range cases {
		f, err := parseFingerprint(cs.url)
		if cs.err == "" {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, NotNil)
			c.Assert(err.Error(), Equals, cs.err)
		}
		c.Assert(f, Equals, cs.fingerprint)
	}
}

func (s *FunctionalSuite) TestParseConstLabels(c *C) {
	cases := []struct {
		s      string
		labels prometheus.Labels
	}{
		{
			s: "a=b",
			labels: prometheus.Labels{
				"a": "b",
			},
		},
		{
			s:      "",
			labels: prometheus.Labels{},
		},
		{
			s: "a=b, c=d",
			labels: prometheus.Labels{
				"a": "b",
				"c": "d",
			},
		},
		{
			s: "a=b, xyz",
			labels: prometheus.Labels{
				"a": "b",
			},
		},
		{
			s:      "a=",
			labels: prometheus.Labels{},
		},
	}

	for _, cs := range cases {
		labels := parseConstLabels(cs.s)
		if !reflect.DeepEqual(labels, cs.labels) {
			c.Fatalf("labels not equal (%v -> %v)", labels, cs.labels)
		}
	}
}

func UnsetEnvironment(c *C, d string) {
	err := os.Unsetenv(d)
	c.Assert(err, IsNil)
}

// test boolean metric type gets converted to float
func (s *FunctionalSuite) TestBooleanConversionToValueAndString(c *C) {
	type TestCase struct {
		input          interface{}
		expectedString string
		expectedValue  float64
		expectedOK     bool
	}

	cases := []TestCase{
		{
			input:          true,
			expectedString: "true",
			expectedValue:  1.0,
			expectedOK:     true,
		},
		{
			input:          false,
			expectedString: "false",
			expectedValue:  0.0,
			expectedOK:     true,
		},
	}

	for _, cs := range cases {
		value, ok := dbToFloat64(cs.input)
		c.Assert(value, Equals, cs.expectedValue)
		c.Assert(ok, Equals, cs.expectedOK)

		str, ok := dbToString(cs.input)
		c.Assert(str, Equals, cs.expectedString)
		c.Assert(ok, Equals, cs.expectedOK)
	}
}

func postgresVersion(db *sql.DB) (int, error) {
	version := 0
	err := db.QueryRow("SHOW server_version_num").Scan(&version)
	return version, err
}

func (s *FunctionalSuite) TestParseUserQueries(c *C) {
	userQueriesData, err := ioutil.ReadFile("./tests/user_queries_ok.yaml")
	if err == nil {
		metricMaps, newQueryOverrides, err := parseUserQueries(userQueriesData)
		c.Assert(err, Equals, nil)
		c.Assert(metricMaps, NotNil)
		c.Assert(newQueryOverrides, NotNil)

		if len(metricMaps) != 2 {
			c.Errorf("Expected 2 metrics from user file, got %d", len(metricMaps))
		}
	}
}
