// Copyright The Prometheus Authors
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

package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// dataSourceOpts holds the --datasource.* flag values. An empty field falls
// back to the matching DATA_SOURCE_* env var. The user and password values
// themselves have no flag: CLI flags are readable via /proc, so secrets are
// passed either by env var or by the *File paths below.
type dataSourceOpts struct {
	URI      string
	URIFile  string
	UserFile string
	PassFile string
}

// deprecatedDataSourceEnvVars maps each env var that now has a flag to its
// replacement. DATA_SOURCE_NAME/USER/PASS are absent on purpose: they carry
// secrets, so they get no flag and stay supported.
var deprecatedDataSourceEnvVars = []struct{ envVar, flag string }{
	{"DATA_SOURCE_URI", "--datasource.uri"},
	{"DATA_SOURCE_URI_FILE", "--datasource.uri-file"},
	{"DATA_SOURCE_USER_FILE", "--datasource.user-file"},
	{"DATA_SOURCE_PASS_FILE", "--datasource.pass-file"},
}

func warnDeprecatedDataSourceEnvVars() {
	for _, d := range deprecatedDataSourceEnvVars {
		if os.Getenv(d.envVar) != "" {
			logger.Warn("Configuring the data source via environment variable is DEPRECATED", "env", d.envVar, "flag", d.flag)
		}
	}
}

func flagOrEnvVar(flagVal, envKey string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envKey)
}

// try to get the DataSource
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > DATA_SOURCE_{USER|PASS}_FILE > DATA_SOURCE_{USER|PASS}
func getDataSources(opts dataSourceOpts) ([]string, error) {
	var dsn = os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) != 0 {
		return strings.Split(dsn, ","), nil
	}

	var user, pass, uri string

	dataSourceUserFile := flagOrEnvVar(opts.UserFile, "DATA_SOURCE_USER_FILE")
	if len(dataSourceUserFile) != 0 {
		fileContents, err := os.ReadFile(dataSourceUserFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source user file %s: %s", dataSourceUserFile, err.Error())
		}
		user = strings.TrimSpace(string(fileContents))
	} else {
		user = os.Getenv("DATA_SOURCE_USER")
	}

	dataSourcePassFile := flagOrEnvVar(opts.PassFile, "DATA_SOURCE_PASS_FILE")
	if len(dataSourcePassFile) != 0 {
		fileContents, err := os.ReadFile(dataSourcePassFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source pass file %s: %s", dataSourcePassFile, err.Error())
		}
		pass = strings.TrimSpace(string(fileContents))
	} else {
		pass = os.Getenv("DATA_SOURCE_PASS")
	}

	ui := url.UserPassword(user, pass).String()
	dataSourceURIFile := flagOrEnvVar(opts.URIFile, "DATA_SOURCE_URI_FILE")
	if len(dataSourceURIFile) != 0 {
		fileContents, err := os.ReadFile(dataSourceURIFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source URI file %s: %s", dataSourceURIFile, err.Error())
		}
		uri = strings.TrimSpace(string(fileContents))
	} else {
		uri = flagOrEnvVar(opts.URI, "DATA_SOURCE_URI")
	}

	// No datasources found. This allows us to support the multi-target pattern
	// without an explicit datasource.
	if uri == "" {
		return []string{}, nil
	}

	dsn = "postgresql://" + ui + "@" + uri

	return []string{dsn}, nil
}
