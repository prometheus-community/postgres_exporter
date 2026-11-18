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

// User and password have no flag of their own: flags are readable via /proc, so
// those secrets arrive by env var or through the *File paths.
type dataSourceOpts struct {
	URI      string
	URIFile  string
	UserFile string
	PassFile string
}

// try to get the DataSource
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > {UserFile|PassFile} > DATA_SOURCE_{USER|PASS}
func getDataSources(opts dataSourceOpts) ([]string, error) {
	var dsn = os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) != 0 {
		return strings.Split(dsn, ","), nil
	}

	var user, pass, uri string

	if len(opts.UserFile) != 0 {
		fileContents, err := os.ReadFile(opts.UserFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source user file %s: %s", opts.UserFile, err.Error())
		}
		user = strings.TrimSpace(string(fileContents))
	} else {
		user = os.Getenv("DATA_SOURCE_USER")
	}

	if len(opts.PassFile) != 0 {
		fileContents, err := os.ReadFile(opts.PassFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source pass file %s: %s", opts.PassFile, err.Error())
		}
		pass = strings.TrimSpace(string(fileContents))
	} else {
		pass = os.Getenv("DATA_SOURCE_PASS")
	}

	ui := url.UserPassword(user, pass).String()
	if len(opts.URIFile) != 0 {
		fileContents, err := os.ReadFile(opts.URIFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source URI file %s: %s", opts.URIFile, err.Error())
		}
		uri = strings.TrimSpace(string(fileContents))
	} else {
		uri = opts.URI
	}

	// No datasources found. This allows us to support the multi-target pattern
	// without an explicit datasource.
	if uri == "" {
		return []string{}, nil
	}

	dsn = "postgresql://" + ui + "@" + uri

	return []string{dsn}, nil
}
