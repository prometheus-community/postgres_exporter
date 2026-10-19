// Copyright 2022 The Prometheus Authors
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

package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// DSN represents a parsed datasource. It contains fields for the individual connection components.
type DSN struct {
	scheme   string
	username string
	password string
	host     string
	path     string
	query    url.Values
}

// String makes a dsn safe to print by excluding any passwords. This allows dsn to be used in
// strings and log messages without needing to call a redaction function first.
func (d DSN) String() string {
	if d.password != "" {
		return fmt.Sprintf("%s://%s:******@%s%s?%s", d.scheme, d.username, d.host, d.path, d.query.Encode())
	}

	if d.username != "" {
		return fmt.Sprintf("%s://%s@%s%s?%s", d.scheme, d.username, d.host, d.path, d.query.Encode())
	}

	return fmt.Sprintf("%s://%s%s?%s", d.scheme, d.host, d.path, d.query.Encode())
}

// GetConnectionString returns the connection string to pass to the driver
// for database connections. This value should not be logged.
func (d DSN) GetConnectionString() string {
	// A host that names a filesystem path (e.g. "/var/run/postgresql/") or
	// abstract socket (prefixed with "@") addresses a Unix-domain socket, as
	// documented for the "host" keyword at
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS.
	// The pq driver's own URL parser rejects a percent-encoded "/" in the URL
	// authority (see lib/pq's connector.go doc comment on convertURL), so a
	// socket path can only survive serialization in key=value form.
	if isSocketHost(d.host) {
		return d.keyValueConnectionString()
	}

	u := url.URL{
		Scheme:   d.scheme,
		Host:     d.host,
		Path:     d.path,
		RawQuery: d.query.Encode(),
	}

	// Username and Password
	if d.username != "" {
		u.User = url.UserPassword(d.username, d.password)
	}

	return u.String()
}

// isSocketHost reports whether host addresses a Unix-domain socket rather
// than a TCP host, per the rules of the "host" connection parameter: an
// absolute path, or a value prefixed with "@" for an abstract socket.
func isSocketHost(host string) bool {
	h := host
	if hh, _, err := net.SplitHostPort(host); err == nil {
		h = hh
	}
	return strings.HasPrefix(h, "/") || strings.HasPrefix(h, "@")
}

// keyValueConnectionString renders d as a key=value connection string, the
// only form the driver accepts for a Unix-domain socket host. Parts are
// sorted so the result is deterministic despite the map iteration used to
// walk the query string.
func (d DSN) keyValueConnectionString() string {
	host, port := d.host, ""
	if h, p, err := net.SplitHostPort(d.host); err == nil {
		host, port = h, p
	}

	escaper := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	var parts []string
	set := func(k, v string) {
		if v == "" {
			return
		}
		parts = append(parts, fmt.Sprintf("%s='%s'", k, escaper.Replace(v)))
	}

	set("host", host)
	set("port", port)
	set("user", d.username)
	set("password", d.password)
	if d.path != "" {
		set("dbname", strings.TrimPrefix(d.path, "/"))
	}
	for k, values := range d.query {
		for _, v := range values {
			set(k, v)
		}
	}

	sort.Strings(parts)
	return strings.Join(parts, " ")
}

// WithDatabase returns a copy of d pointed at database instead of whichever
// database it originally targeted. The override is always carried in the
// query string (as "dbname"), regardless of how the database was originally
// specified, since a "dbname" left behind in the query string would
// otherwise take precedence over the path when the driver parses the
// resulting connection string.
func (d DSN) WithDatabase(database string) DSN {
	d.path = ""
	query := url.Values{}
	for k, v := range d.query {
		if k == "dbname" {
			continue
		}
		query[k] = v
	}
	query.Set("dbname", database)
	d.query = query
	return d
}

// NewDSN parses a connection string into a DSN.
func NewDSN(in string) (DSN, error) {
	return dsnFromString(in)
}

// dsnFromString parses a connection string into a dsn. It will attempt to parse the string as
// a URL and as a set of key=value pairs. If both attempts fail, dsnFromString will return an error.
func dsnFromString(in string) (DSN, error) {
	if strings.HasPrefix(in, "postgresql://") || strings.HasPrefix(in, "postgres://") {
		return dsnFromURL(in)
	}

	// Try to parse as key=value pairs
	d, err := dsnFromKeyValue(in)
	if err == nil {
		return d, nil
	}

	// Input that starts like a key=value string must not fall back to the URL
	// parser. url.Parse accepts some malformed key=value strings (for example
	// "host=/tmp password=a b") and puts the whole input, password included,
	// into the host or path, where String() would print it.
	if keyValueRe.MatchString(in) {
		return DSN{}, err
	}

	// Parse the string as a URL, with the scheme prefixed
	d, err = dsnFromURL(fmt.Sprintf("postgresql://%s", in))
	if err == nil {
		return d, nil
	}

	return DSN{}, fmt.Errorf("could not understand DSN")
}

// dsnFromURL parses the input as a URL and returns the dsn representation.
func dsnFromURL(in string) (DSN, error) {
	u, err := url.Parse(in)
	if err != nil {
		return DSN{}, err
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

	d := DSN{
		scheme:   u.Scheme,
		username: user,
		password: pass,
		host:     u.Host,
		path:     u.Path,
		query:    query,
	}

	return d, nil
}

// keyValueRe matches the start of a key=value connection string: a keyword
// followed by "=".
var keyValueRe = regexp.MustCompile(`^\s*[a-zA-Z0-9_]+\s*=`)

// dsnFromKeyValue parses the input as a set of key=value pairs and returns the dsn representation.
func dsnFromKeyValue(in string) (DSN, error) {
	// Attempt to confirm at least one key=value pair before starting the rune parser
	if !keyValueRe.MatchString(in) {
		return DSN{}, fmt.Errorf("input is not a key-value DSN")
	}

	// Anything other than known fields should be part of the querystring
	query := url.Values{}

	pairs, err := parseKeyValue(in)
	if err != nil {
		return DSN{}, fmt.Errorf("failed to parse key-value DSN: %v", err)
	}

	// Build the dsn from the key=value pairs
	d := DSN{
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

	d.query = query

	return d, nil
}

// parseKeyValue parses a keyword/value connection string the way libpq does
// (https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-KEYWORD-VALUE):
// settings are separated by whitespace, spaces around "=" are optional, a value
// ends at the first whitespace unless it is quoted, and a backslash escapes the
// next character. A value may contain "=". An unterminated quote or a trailing
// backslash is an error. Unlike libpq, double quotes are accepted as well as
// single quotes.
func parseKeyValue(in string) (map[string]string, error) {
	out := map[string]string{}
	s := []rune(in)
	i := 0
	skipSpaces := func() {
		for i < len(s) && unicode.IsSpace(s[i]) {
			i++
		}
	}

	for {
		skipSpaces()
		if i == len(s) {
			return out, nil
		}

		start := i
		for i < len(s) && !unicode.IsSpace(s[i]) && s[i] != '=' {
			i++
		}
		key := string(s[start:i])
		skipSpaces()
		if key == "" || i == len(s) || s[i] != '=' {
			return out, fmt.Errorf("invalid input")
		}
		i++ // skip "="
		skipSpaces()

		var quote rune
		if i < len(s) && (s[i] == '\'' || s[i] == '"') {
			quote = s[i]
			i++
		}
		value := []rune{}
		closed := quote == 0
		for i < len(s) {
			c := s[i]
			i++
			if quote != 0 && c == quote {
				closed = true
				break
			}
			if quote == 0 && unicode.IsSpace(c) {
				break
			}
			if c == '\\' {
				if i == len(s) {
					return out, fmt.Errorf("invalid input")
				}
				c = s[i]
				i++
			}
			value = append(value, c)
		}
		if !closed {
			return out, fmt.Errorf("invalid input")
		}
		out[key] = string(value)
	}
}
