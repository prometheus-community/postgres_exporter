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

package config

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	redactedSecret = "SECRET_REMOVED"
	redactedDSN    = "DSN_REMOVED"

	// minSecretLength skips short values such as "0" that would garble the
	// message. Passwords are removed at any length.
	minSecretLength = 8
)

var (
	// queryParamRe matches a parameter in a URL query string.
	queryParamRe = regexp.MustCompile(`[?&]([^=&#]*)=([^&#]*)`)
	// keyValueKeywordRe matches a keyword and its "=" in a key=value
	// connection string. A keyword's value runs to the next match.
	keyValueKeywordRe = regexp.MustCompile(`(?:^|\s)([A-Za-z_]+)\s*=\s*`)
)

// RedactError returns err with its message passed through RedactSecrets.
// Use it on any error from parsing or connecting with dsn before logging it:
// the URL parser and the pq driver both echo the connection string. The
// original error stays reachable through errors.Unwrap.
func RedactError(err error, dsn string) error {
	if err == nil {
		return nil
	}
	return &redactedError{msg: RedactSecrets(err.Error(), dsn), err: err}
}

type redactedError struct {
	msg string
	err error
}

func (e *redactedError) Error() string { return e.msg }
func (e *redactedError) Unwrap() error { return e.err }

// RedactSecrets returns text with copies of dsn replaced by DSN.String and
// any value String omits removed. dsn may be malformed.
func RedactSecrets(text, dsn string) string {
	if dsn == "" {
		return text
	}

	// Whole copies go first: removing values would break the match.
	safe := redactedDSN
	if d, err := NewDSN(dsn); err == nil {
		safe = d.String()
	}
	var dsnReplacements []string
	for _, form := range encodedForms(dsn) {
		dsnReplacements = append(dsnReplacements, form, safe)
	}
	text = replace(text, dsnReplacements)

	// Catch fragments, such as the driver quoting part of a malformed value.
	var replacements []string
	for _, secret := range secretCandidates(dsn) {
		for _, form := range encodedForms(secret) {
			replacements = append(replacements, form, redactedSecret)
		}
	}
	return replace(text, replacements)
}

// replace applies the old, new pairs in replacements to text, longest old
// string first so that a secret is never partially replaced by one of its own
// substrings.
func replace(text string, replacements []string) string {
	type pair struct{ old, new string }
	pairs := make([]pair, 0, len(replacements)/2)
	for i := 0; i < len(replacements); i += 2 {
		if replacements[i] != "" {
			pairs = append(pairs, pair{replacements[i], replacements[i+1]})
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool { return len(pairs[i].old) > len(pairs[j].old) })

	flat := make([]string, 0, len(pairs)*2)
	for _, p := range pairs {
		flat = append(flat, p.old, p.new)
	}
	return strings.NewReplacer(flat...).Replace(text)
}

// secretCandidates returns the values in dsn, which may be malformed, that
// String omits.
func secretCandidates(dsn string) []string {
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		out = append(out, s)
		if u, err := url.PathUnescape(s); err == nil && u != s {
			out = append(out, u)
		}
		if u, err := url.QueryUnescape(s); err == nil && u != s {
			out = append(out, u)
		}
	}
	addValue := func(key, value string) {
		if loggableParams[key] || (key != "password" && len(value) < minSecretLength) {
			return
		}
		add(value)
	}

	if d, err := NewDSN(dsn); err == nil {
		add(d.password)
		for k, values := range d.query {
			for _, v := range values {
				addValue(k, v)
			}
		}
	}

	// The extractors below all run regardless of which form dsn appears to
	// be in, since NewDSN also accepts a URL with its scheme left off.

	// Without a successful parse the authority cannot be located reliably,
	// so take everything up to the last "@" as user info.
	rest := strings.TrimPrefix(strings.TrimPrefix(dsn, "postgresql://"), "postgres://")
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		if _, pass, found := strings.Cut(rest[:at], ":"); found {
			add(pass)
		}
	}

	for _, m := range queryParamRe.FindAllStringSubmatch(dsn, -1) {
		key := m[1]
		if k, err := url.QueryUnescape(key); err == nil {
			key = k
		}
		addValue(key, m[2])
	}

	matches := keyValueKeywordRe.FindAllStringSubmatchIndex(dsn, -1)
	for i, m := range matches {
		end := len(dsn)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		key := dsn[m[2]:m[3]]
		value := strings.TrimSpace(dsn[m[1]:end])
		addValue(key, value)
		addValue(key, unquoteKeyValue(value))
		for field := range strings.FieldsSeq(value) {
			addValue(key, field)
			addValue(key, strings.Trim(field, `'`))
		}
	}
	return out
}

// unquoteKeyValue strips the quoting and backslash escaping of a key=value
// connection string value. An unterminated quote is tolerated.
func unquoteKeyValue(s string) string {
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")
	return strings.NewReplacer(`\'`, `'`, `\\`, `\`).Replace(s)
}

// encodedForms returns s together with the escaped forms it can take when it
// is embedded in a connection string or quoted in an error message.
func encodedForms(s string) []string {
	forms := []string{
		s,
		url.PathEscape(s),
		url.QueryEscape(s),
		strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s),
	}
	if ui := url.UserPassword("", s).String(); len(ui) > 1 {
		forms = append(forms, ui[1:]) // strip the leading ":"
	}
	if q := strconv.Quote(s); len(q) > 2 {
		forms = append(forms, q[1:len(q)-1]) // as formatted by %q
	}
	return forms
}
