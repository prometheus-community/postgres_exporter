// Copyright 2021 The Prometheus Authors
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
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) discoverDatabaseDSNs() []string {
	// connstring syntax is complex (and not sure if even regular).
	// we don't need to parse it, so just superficially validate that it starts
	// with a valid-ish keyword pair
	connstringRe := regexp.MustCompile(`^ *[a-zA-Z0-9]+ *= *[^= ]+`)

	dsns := make(map[string]struct{})
	for _, dsn := range e.dsn {
		var dsnURI *url.URL
		var dsnConnstring string

		if strings.HasPrefix(dsn, "postgresql://") {
			var err error
			dsnURI, err = url.Parse(dsn)
			if err != nil {
				level.Error(logger).Log("msg", "Unable to parse DSN as URI", "dsn", loggableDSN(dsn), "err", err)
				continue
			}
		} else if connstringRe.MatchString(dsn) {
			dsnConnstring = dsn
		} else {
			level.Error(logger).Log("msg", "Unable to parse DSN as either URI or connstring", "dsn", loggableDSN(dsn))
			continue
		}

		server, err := e.servers.GetServer(dsn)
		if err != nil {
			level.Error(logger).Log("msg", "Error opening connection to database", "dsn", loggableDSN(dsn), "err", err)
			continue
		}
		dsns[dsn] = struct{}{}

		// If autoDiscoverDatabases is true, set first dsn as master database (Default: false)
		server.master = true

		databaseNames, err := queryDatabases(server)
		if err != nil {
			level.Error(logger).Log("msg", "Error querying databases", "dsn", loggableDSN(dsn), "err", err)
			continue
		}
		for _, databaseName := range databaseNames {
			if contains(e.excludeDatabases, databaseName) {
				continue
			}

			if len(e.includeDatabases) != 0 && !contains(e.includeDatabases, databaseName) {
				continue
			}

			if dsnURI != nil {
				dsnURI.Path = databaseName
				dsn = dsnURI.String()
			} else {
				// replacing one dbname with another is complicated.
				// just append new dbname to override.
				dsn = fmt.Sprintf("%s dbname=%s", dsnConnstring, databaseName)
			}
			dsns[dsn] = struct{}{}
		}
	}

	result := make([]string, len(dsns))
	index := 0
	for dsn := range dsns {
		result[index] = dsn
		index++
	}

	return result
}

func (e *Exporter) scrapeDSN(ch chan<- prometheus.Metric, dsn string) error {
	server, err := e.servers.GetServer(dsn)

	if err != nil {
		return &ErrorConnectToServer{fmt.Sprintf("Error opening connection to database (%s): %s", loggableDSN(dsn), err.Error())}
	}

	// Check if autoDiscoverDatabases is false, set dsn as master database (Default: false)
	if !e.autoDiscoverDatabases {
		server.master = true
	}

	// Check if map versions need to be updated
	if err := e.checkMapVersions(ch, server); err != nil {
		level.Warn(logger).Log("msg", "Proceeding with outdated query maps, as the Postgres version could not be determined", "err", err)
	}

	return server.Scrape(ch, e.disableSettingsMetrics)
}

// try to get the DataSource
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > DATA_SOURCE_{USER|PASS}_FILE > DATA_SOURCE_{USER|PASS}
func getDataSources() ([]string, error) {
	var dsn = os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) != 0 {
		return strings.Split(dsn, ","), nil
	}

	var user, pass, uri string

	dataSourceUserFile := os.Getenv("DATA_SOURCE_USER_FILE")
	if len(dataSourceUserFile) != 0 {
		fileContents, err := ioutil.ReadFile(dataSourceUserFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source user file %s: %s", dataSourceUserFile, err.Error())
		}
		user = strings.TrimSpace(string(fileContents))
	} else {
		user = os.Getenv("DATA_SOURCE_USER")
	}

	dataSourcePassFile := os.Getenv("DATA_SOURCE_PASS_FILE")
	if len(dataSourcePassFile) != 0 {
		fileContents, err := ioutil.ReadFile(dataSourcePassFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source pass file %s: %s", dataSourcePassFile, err.Error())
		}
		pass = strings.TrimSpace(string(fileContents))
	} else {
		pass = os.Getenv("DATA_SOURCE_PASS")
	}

	ui := url.UserPassword(user, pass).String()
	dataSrouceURIFile := os.Getenv("DATA_SOURCE_URI_FILE")
	if len(dataSrouceURIFile) != 0 {
		fileContents, err := ioutil.ReadFile(dataSrouceURIFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading data source URI file %s: %s", dataSrouceURIFile, err.Error())
		}
		uri = strings.TrimSpace(string(fileContents))
	} else {
		uri = os.Getenv("DATA_SOURCE_URI")
	}

	// No datasources found. This allows us to support the multi-target pattern
	// without an explicit datasource.
	if uri == "" {
		return []string{}, nil
	}

	dsn = "postgresql://" + ui + "@" + uri

	return []string{dsn}, nil
}

// dsn represents a parsed datasource. It contains fields for the individual connection components.
type dsn struct {
	scheme   string
	username string
	password string
	host     string
	path     string
	query    string
}

// String makes a dsn safe to print by excluding any passwords. This allows dsn to be used in
// strings and log messages without needing to call a redaction function first.
func (d dsn) String() string {
	if d.password != "" {
		return fmt.Sprintf("%s://%s:******@%s%s?%s", d.scheme, d.username, d.host, d.path, d.query)
	}

	if d.username != "" {
		return fmt.Sprintf("%s://%s@%s%s?%s", d.scheme, d.username, d.host, d.path, d.query)
	}

	return fmt.Sprintf("%s://%s%s?%s", d.scheme, d.host, d.path, d.query)
}

// dsnFromString parses a connection string into a dsn. It will attempt to parse the string as
// a URL and as a set of key=value pairs. If both attempts fail, dsnFromString will return an error.
func dsnFromString(in string) (dsn, error) {
	if strings.HasPrefix(in, "postgresql://") {
		return dsnFromURL(in)
	}

	// Try to parse as key=value pairs
	d, err := dsnFromKeyValue(in)
	if err == nil {
		return d, nil
	}

	return dsn{}, fmt.Errorf("could not understand DSN")
}

// dsnFromURL parses the input as a URL and returns the dsn representation.
func dsnFromURL(in string) (dsn, error) {
	u, err := url.Parse(in)
	if err != nil {
		return dsn{}, err
	}
	pass, _ := u.User.Password()
	user := u.User.Username()

	query := u.Query()

	if queryPass := query.Get("password"); queryPass != "" {
		if pass == "" {
			pass = queryPass
		}
	}
	query.Del("password")

	if queryUser := query.Get("user"); queryUser != "" {
		if user == "" {
			user = queryUser
		}
	}
	query.Del("user")

	d := dsn{
		scheme:   u.Scheme,
		username: user,
		password: pass,
		host:     u.Host,
		path:     u.Path,
		query:    query.Encode(),
	}

	return d, nil
}

// dsnFromKeyValue parses the input as a set of key=value pairs and returns the dsn representation.
func dsnFromKeyValue(in string) (dsn, error) {
	// Attempt to confirm at least one key=value pair before starting the rune parser
	connstringRe := regexp.MustCompile(`^ *[a-zA-Z0-9]+ *= *[^= ]+`)
	if !connstringRe.MatchString(in) {
		return dsn{}, fmt.Errorf("input is not a key-value DSN")
	}

	// Anything other than known fields should be part of the querystring
	query := url.Values{}

	pairs, err := parseKeyValue(in)
	if err != nil {
		return dsn{}, fmt.Errorf("failed to parse key-value DSN: %v", err)
	}

	// Build the dsn from the key=value pairs
	d := dsn{
		scheme: "postgresql",
	}

	hostname := ""
	port := ""

	for k, v := range pairs {
		switch k {
		case "host":
			hostname = v
		case "port":
			port = v
		case "user":
			d.username = v
		case "password":
			d.password = v
		default:
			query.Set(k, v)
		}
	}

	if hostname == "" {
		hostname = "localhost"
	}

	if port == "" {
		d.host = hostname
	} else {
		d.host = fmt.Sprintf("%s:%s", hostname, port)
	}

	d.query = query.Encode()

	return d, nil
}

// parseKeyValue is a key=value parser. It loops over each rune to split out keys and values
// and attempting to honor quoted values. parseKeyValue will return an error if it is unable
// to properly parse the input.
func parseKeyValue(in string) (map[string]string, error) {
	out := map[string]string{}

	inPart := false
	inQuote := false
	part := []rune{}
	key := ""
	for _, c := range in {
		switch {
		case unicode.In(c, unicode.Quotation_Mark):
			if inQuote {
				inQuote = false
			} else {
				inQuote = true
			}
		case unicode.In(c, unicode.White_Space):
			if inPart {
				if inQuote {
					part = append(part, c)
				} else {
					// Are we finishing a key=value?
					if key == "" {
						return out, fmt.Errorf("invalid input")
					}
					out[key] = string(part)
					inPart = false
					part = []rune{}
				}
			} else {
				// Are we finishing a key=value?
				if key == "" {
					return out, fmt.Errorf("invalid input")
				}
				out[key] = string(part)
				inPart = false
				part = []rune{}
				// Do something with the value
			}
		case c == '=':
			if inPart {
				inPart = false
				key = string(part)
				part = []rune{}
			} else {
				return out, fmt.Errorf("invalid input")
			}
		default:
			inPart = true
			part = append(part, c)
		}
	}

	if key != "" && len(part) > 0 {
		out[key] = string(part)
	}

	return out, nil
}
