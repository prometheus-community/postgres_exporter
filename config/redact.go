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
	redactedPassword = "PASSWORD_REMOVED"
	redactedDSN      = "DSN_REMOVED"
)

var (
	// queryPasswordRe matches a password supplied as a URL query parameter.
	queryPasswordRe = regexp.MustCompile(`[?&]password=([^&#]*)`)
	// keyValuePasswordRe matches a password keyword in a key=value connection
	// string and everything after it up to the next keyword (or the end). A
	// malformed value, such as an unquoted password containing a space, can
	// leave part of the password where the driver expects a keyword, and the
	// driver then quotes that fragment in its error. The driver treats any
	// whitespace, newlines included, as a separator, so the s flag lets the
	// value span lines rather than leaving the match to fail at the first
	// newline.
	keyValuePasswordRe = regexp.MustCompile(`(?s)(?:^|\s)password\s*=\s*(.*?)(?:\s+[A-Za-z_]+\s*=|$)`)
)

// RedactError returns err with every copy of dsn's password, and of dsn
// itself, removed from its message. Use it on any error produced while
// parsing dsn or connecting with it before logging or wrapping that error:
// the standard library's URL parser and the pq driver both echo the
// connection string they were given. The original error stays reachable
// through errors.Unwrap so that errors.Is and errors.As keep working.
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

// RedactSecrets returns text with every copy of dsn's password, and of dsn
// itself, removed. It works on malformed connection strings too, since those
// are the ones that end up quoted in error messages.
func RedactSecrets(text, dsn string) string {
	if dsn == "" {
		return text
	}

	var replacements []string
	for _, secret := range passwordCandidates(dsn) {
		for _, form := range encodedForms(secret) {
			replacements = append(replacements, form, redactedPassword)
		}
	}
	// Passwords are replaced first so that an error quoting a parseable dsn
	// stays useful. Whatever copy of dsn is still left contains a password
	// that could not be located, so it is removed whole.
	text = replace(text, replacements)

	var dsnReplacements []string
	for _, form := range encodedForms(dsn) {
		dsnReplacements = append(dsnReplacements, form, redactedDSN)
	}
	return replace(text, dsnReplacements)
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

// passwordCandidates returns every string in dsn that may be all or part of a
// password. It is deliberately lenient: dsn may be malformed.
func passwordCandidates(dsn string) []string {
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

	if d, err := NewDSN(dsn); err == nil {
		add(d.password)
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

	for _, m := range queryPasswordRe.FindAllStringSubmatch(dsn, -1) {
		add(m[1])
	}

	for _, m := range keyValuePasswordRe.FindAllStringSubmatch(dsn, -1) {
		value := strings.TrimSpace(m[1])
		add(value)
		add(unquoteKeyValue(value))
		for field := range strings.FieldsSeq(value) {
			add(field)
			add(strings.Trim(field, `'`))
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
