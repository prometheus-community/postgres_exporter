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
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
)

// The deprecated PG_EXPORTER_* environment variables were removed in v0.22.0:
// no flag may bind them anymore, and setting them must not change flag values.
func TestRemovedEnvVarsHaveNoEffect(t *testing.T) {
	for _, flag := range kingpin.CommandLine.Model().Flags {
		if strings.HasPrefix(flag.Envar, "PG_EXPORTER_") {
			t.Errorf("flag %s still binds removed environment variable %s", flag.Name, flag.Envar)
		}
	}

	t.Setenv("PG_EXPORTER_METRIC_PREFIX", "env-value-not-to-apply")
	app := kingpin.New("test", "")
	got := app.Flag("metric-prefix", "").Default("pg").String()
	if _, err := app.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if *got != "pg" {
		t.Fatalf("metric-prefix = %q, want default %q", *got, "pg")
	}
}
