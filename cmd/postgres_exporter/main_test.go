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
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

func TestWarnDeprecatedEnvVars(t *testing.T) {
	oldLogger := logger
	t.Cleanup(func() { logger = oldLogger })

	flags := []struct {
		name   string
		env    string
		isBool bool
	}{
		{"web.telemetry-path", webTelemetryPathEnvar, false},
		{"disable-default-metrics", "PG_EXPORTER_DISABLE_DEFAULT_METRICS", true},
		{"auto-discover-databases", "PG_EXPORTER_AUTO_DISCOVER_DATABASES", true},
		{"extend.query-path", "PG_EXPORTER_EXTEND_QUERY_PATH", false},
		{"constantLabels", "PG_EXPORTER_CONSTANT_LABELS", false},
		{"exclude-databases", "PG_EXPORTER_EXCLUDE_DATABASES", false},
		{"include-databases", "PG_EXPORTER_INCLUDE_DATABASES", false},
		{"metric-prefix", "PG_EXPORTER_METRIC_PREFIX", false},
		{"collection-timeout", "PG_EXPORTER_COLLECTION_TIMEOUT", false},
	}
	for _, flag := range flags {
		t.Setenv(flag.env, "")
	}

	for _, flag := range flags {
		t.Run(flag.name, func(t *testing.T) {
			envValue := "env-value-not-to-log"
			cliValue := "cli-value"
			if flag.isBool {
				envValue = "true"
			}
			type testCase struct {
				name string
				env  string
				args []string
				warn bool
			}
			cases := []testCase{
				{"empty", "", nil, false},
				{"environment", envValue, nil, true},
			}
			if flag.isBool {
				cases = append(cases,
					testCase{"false environment", "false", nil, true},
					testCase{"bare flag", "false", []string{"--" + flag.name}, false},
					testCase{"negated flag", envValue, []string{"--no-" + flag.name}, false},
				)
			} else {
				cases = append(cases,
					testCase{"equals value", envValue, []string{"--" + flag.name + "=" + cliValue}, false},
					testCase{"separate value", envValue, []string{"--" + flag.name, cliValue}, false},
					testCase{"explicit empty value", envValue, []string{"--" + flag.name + "="}, false},
				)
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Setenv(flag.env, tc.env)
					app := kingpin.New("test", "")
					for _, model := range kingpin.CommandLine.Model().Flags {
						if !strings.HasPrefix(model.Envar, "PG_EXPORTER_") {
							continue
						}
						clause := app.Flag(model.Name, model.Help).Envar(model.Envar).Default(model.Default...)
						if model.IsBoolFlag() {
							clause.Bool()
						} else {
							clause.String()
						}
					}
					app.Flag("web.telemetry-path", "").Envar(webTelemetryPathEnvar).Default("/metrics").String()
					tracked := trackDeprecatedEnvVarFlags(app)
					if _, err := app.Parse(tc.args); err != nil {
						t.Fatal(err)
					}
					if len(tc.args) > 0 {
						want := cliValue
						if flag.isBool {
							want = "true"
							if strings.HasPrefix(tc.args[0], "--no-") {
								want = "false"
							}
						} else if tc.name == "explicit empty value" {
							want = ""
						}
						if got := app.GetFlag(flag.name).Model().Value.String(); got != want {
							t.Fatalf("flag value = %q, want %q", got, want)
						}
					}
					var output bytes.Buffer
					logger = slog.New(slog.NewTextHandler(&output, nil))
					warnDeprecatedEnvVars(tracked)
					got := output.String()
					if !tc.warn {
						if got != "" {
							t.Fatalf("unexpected warning: %s", got)
						}
						return
					}
					if strings.Count(got, "level=WARN") != 1 || !strings.Contains(got, "env="+flag.env+" flag=--"+flag.name) {
						t.Fatalf("missing deprecation warning for %s: %s", flag.env, got)
					}
					if strings.Contains(got, tc.env) {
						t.Fatalf("warning contains environment value: %s", got)
					}
				})
			}
		})
	}
}

func TestParseCollectionTimeout(t *testing.T) {
	t.Run("valid duration", func(t *testing.T) {
		got, err := parseCollectionTimeout("30s")
		if err != nil {
			t.Fatalf("parseCollectionTimeout() error = %v", err)
		}
		if want := 30 * time.Second; got != want {
			t.Fatalf("parseCollectionTimeout() = %v, want %v", got, want)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		const input = "not-a-duration"

		got, err := parseCollectionTimeout(input)
		if err == nil {
			t.Fatal("parseCollectionTimeout() error = nil, want error")
		}
		if got != 0 {
			t.Fatalf("parseCollectionTimeout() = %v, want 0", got)
		}
		if want := `invalid collection timeout "not-a-duration"`; !strings.Contains(err.Error(), want) {
			t.Fatalf("parseCollectionTimeout() error = %q, want it to contain %q", err, want)
		}
	})
}
