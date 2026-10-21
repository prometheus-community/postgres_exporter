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
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/lib/pq"
)

// TestRedactSecretsRemovesPasswords feeds RedactSecrets the error messages
// that the URL parser, NewDSN and the pq driver actually produce for each
// connection string, alongside messages that quote the connection string
// directly, and checks that no part of the password survives.
func TestRedactSecretsRemovesPasswords(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		secrets []string
		// extra holds messages that quote the connection string in a form
		// none of the parsers produce on their own.
		extra []string
	}{
		{
			// The form GetDataSources builds from DATA_SOURCE_USER and
			// DATA_SOURCE_PASS (issue #1408).
			name:    "url with encoded space in password",
			dsn:     "postgresql://demo:dummy%20secret-PW@127.0.0.1:5999/demo?sslmode=disable",
			secrets: []string{"dummy secret-PW", "dummy%20secret-PW"},
			extra: []string{
				// The message logged by v0.20.1, quoting pq.ParseURL's output.
				`malformed dsn "dbname='demo' host='127.0.0.1' password='dummy secret-PW' port='5999' sslmode='disable' user='demo'"`,
			},
		},
		{
			name:    "url with invalid escape in password",
			dsn:     "postgresql://demo:bad%zzpw@127.0.0.1:5999/demo",
			secrets: []string{"bad%zzpw"},
		},
		{
			name:    "postgres scheme with unescaped at sign in password",
			dsn:     "postgres://demo:p@ss@127.0.0.1:5999/demo",
			secrets: []string{"p@ss"},
		},
		{
			name:    "url with password query parameter",
			dsn:     "postgresql://127.0.0.1:5999/demo?user=demo&password=hunter2&sslmode=disable",
			secrets: []string{"hunter2"},
		},
		{
			name:    "url without scheme",
			dsn:     "demo:hunter2@127.0.0.1:5999/demo",
			secrets: []string{"hunter2"},
		},
		{
			name:    "key value with quoted password containing a space",
			dsn:     "host=127.0.0.1 port=5999 user=demo password='dummy secret-PW'",
			secrets: []string{"dummy secret-PW"},
		},
		{
			// pq reports the part after the space as a keyword missing its
			// "=", quoting it.
			name:    "key value with unquoted password containing a space",
			dsn:     "host=127.0.0.1 port=5999 user=demo password=dummy secret-PW",
			secrets: []string{"dummy", "secret-PW"},
		},
		{
			name:    "key value with unterminated quote",
			dsn:     "host=127.0.0.1 port=5999 user=demo password='dummy secret-PW",
			secrets: []string{"dummy secret-PW"},
		},
		{
			name:    "key value with escaped quote in password",
			dsn:     `host=127.0.0.1 user=demo password='it\'s a secret' sslmode=disable`,
			secrets: []string{`it's a secret`, `it\'s a secret`},
		},
		{
			name:    "key value with trailing newline after unterminated quote",
			dsn:     "host=127.0.0.1 port=5999 user=demo password='dummy secret-PW\n",
			secrets: []string{"dummy secret-PW"},
		},
		{
			name:    "key value with trailing newline after unquoted password containing a space",
			dsn:     "host=127.0.0.1 port=5999 user=demo password=dummy secret-PW\n",
			secrets: []string{"dummy", "secret-PW"},
		},
		{
			name:    "key value with newline in unquoted password",
			dsn:     "host=127.0.0.1 port=5999 user=demo password=dummy\nsecret-PW sslmode=disable",
			secrets: []string{"dummy", "secret-PW"},
		},
		{
			name:    "key value with newline in unterminated quote",
			dsn:     "host=127.0.0.1 port=5999 user=demo password='dummy\nsecret-PW",
			secrets: []string{"dummy\nsecret-PW", "dummy", "secret-PW"},
		},
		{
			name:    "key value with pairs separated by newlines",
			dsn:     "host=127.0.0.1\nport=5999\nuser=demo\npassword=dummy secret-PW\nsslmode=disable",
			secrets: []string{"dummy", "secret-PW"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := append([]string{
				tt.dsn,
				fmt.Sprintf("%q", tt.dsn),
			}, tt.extra...)
			if _, err := url.Parse(tt.dsn); err != nil {
				messages = append(messages, err.Error())
			}
			if _, err := NewDSN(tt.dsn); err != nil {
				messages = append(messages, err.Error())
			}
			if _, err := pq.NewConfig(tt.dsn); err != nil {
				messages = append(messages, err.Error())
			}

			for _, msg := range messages {
				got := RedactSecrets(msg, tt.dsn)
				for _, secret := range tt.secrets {
					if strings.Contains(got, secret) {
						t.Errorf("RedactSecrets(%q) = %q, still contains %q", msg, got, secret)
					}
				}
			}
		})
	}
}

func TestRedactSecretsKeepsContext(t *testing.T) {
	tests := []struct {
		name string
		text string
		dsn  string
		want string
	}{
		{
			name: "password replaced in place",
			text: `parse "postgresql://demo:hunter2@127.0.0.1:5999/demo": bad`,
			dsn:  "postgresql://demo:hunter2@127.0.0.1:5999/demo",
			want: `parse "postgresql://demo:PASSWORD_REMOVED@127.0.0.1:5999/demo": bad`,
		},
		{
			name: "dsn without password left alone",
			text: `dial tcp 127.0.0.1:5999: connect: connection refused`,
			dsn:  "postgresql://demo@127.0.0.1:5999/demo",
			want: `dial tcp 127.0.0.1:5999: connect: connection refused`,
		},
		{
			name: "empty dsn left alone",
			text: "something failed",
			dsn:  "",
			want: "something failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactSecrets(tt.text, tt.dsn); got != tt.want {
				t.Errorf("RedactSecrets() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRedactError(t *testing.T) {
	if err := RedactError(nil, "postgresql://demo:hunter2@localhost/demo"); err != nil {
		t.Fatalf("RedactError(nil) = %v, want nil", err)
	}

	dsn := "postgresql://demo:hunter2@localhost/demo"
	err := RedactError(fmt.Errorf("connecting to %s: %w", dsn, context.DeadlineExceeded), dsn)
	if strings.Contains(err.Error(), "hunter2") {
		t.Errorf("RedactError() = %q, still contains the password", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("errors.Is(RedactError(), context.DeadlineExceeded) = false, want true")
	}
}

func TestNewDSNErrorOmitsInput(t *testing.T) {
	_, err := NewDSN("postgresql://demo:bad%zzpw@127.0.0.1:5999/demo")
	if err == nil {
		t.Fatal("NewDSN() error = nil, want an error")
	}
	if strings.Contains(err.Error(), "bad%zzpw") {
		t.Errorf("NewDSN() error = %q, contains the password", err)
	}
}
