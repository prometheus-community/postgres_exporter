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
	"os"
	"path/filepath"
	"testing"
)

const secretsFileDSN = "postgresql://custom_username$&+,%2F%3A;=%3F%40:custom_password$&+,%2F%3A;=%3F%40@localhost:5432/?sslmode=disable"

func writeSecretFile(t *testing.T, name, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(value+"\n"), 0600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	return path
}

const (
	testUser = "custom_username$&+,/:;=?@"
	testPass = "custom_password$&+,/:;=?@"
)

// CI sets DATA_SOURCE_NAME, which would otherwise win over anything a test sets.
func clearDataSourceEnv(t *testing.T) {
	for _, envVar := range []string{
		"DATA_SOURCE_NAME",
		"DATA_SOURCE_URI", "DATA_SOURCE_URI_FILE",
		"DATA_SOURCE_USER", "DATA_SOURCE_USER_FILE",
		"DATA_SOURCE_PASS", "DATA_SOURCE_PASS_FILE",
	} {
		t.Setenv(envVar, "")
	}
}

func TestGetDataSourcesWithSecretsFiles(t *testing.T) {
	clearDataSourceEnv(t)
	t.Setenv("DATA_SOURCE_USER_FILE", writeSecretFile(t, "user", testUser))
	t.Setenv("DATA_SOURCE_PASS_FILE", writeSecretFile(t, "pass", testPass))
	t.Setenv("DATA_SOURCE_URI", "localhost:5432/?sslmode=disable")

	dsn, err := getDataSources(dataSourceOpts{})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 1 || dsn[0] != secretsFileDSN {
		t.Fatalf("getDataSources() = %v, want [%v]", dsn, secretsFileDSN)
	}
}

func TestGetDataSourcesFromFlags(t *testing.T) {
	clearDataSourceEnv(t)
	dsn, err := getDataSources(dataSourceOpts{
		UserFile: writeSecretFile(t, "user", testUser),
		PassFile: writeSecretFile(t, "pass", testPass),
		URI:      "localhost:5432/?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 1 || dsn[0] != secretsFileDSN {
		t.Fatalf("getDataSources() = %v, want [%v]", dsn, secretsFileDSN)
	}
}

func TestGetDataSourcesURIFileFlagWins(t *testing.T) {
	clearDataSourceEnv(t)
	const want = "postgresql://:@from-file:5432/?sslmode=disable"

	dsn, err := getDataSources(dataSourceOpts{
		URI:     "ignored:5432/?sslmode=disable",
		URIFile: writeSecretFile(t, "uri", "from-file:5432/?sslmode=disable"),
	})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 1 || dsn[0] != want {
		t.Fatalf("getDataSources() = %v, want [%v]", dsn, want)
	}
}

func TestGetDataSourcesEnvDSNWinsOverFlags(t *testing.T) {
	clearDataSourceEnv(t)
	const envDSN = "postgresql://envUser:envPass@localhost:5432/?sslmode=disable"
	t.Setenv("DATA_SOURCE_NAME", envDSN)

	dsn, err := getDataSources(dataSourceOpts{
		UserFile: writeSecretFile(t, "user", testUser),
		PassFile: writeSecretFile(t, "pass", testPass),
		URI:      "localhost:5432/?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 1 || dsn[0] != envDSN {
		t.Fatalf("getDataSources() = %v, want [%v]", dsn, envDSN)
	}
}
