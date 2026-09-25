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

// writeSecretFile writes value to a temp file and returns its path, standing in
// for a secret mounted from a Kubernetes Secret or Docker secret.
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

// clearDataSourceEnv blanks every DATA_SOURCE_* var for the test, so an
// ambient environment (CI sets DATA_SOURCE_NAME) cannot change the result.
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

// secrets read from the files the DATA_SOURCE_*_FILE env vars point at
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

// DATA_SOURCE_NAME is used as-is, even when user/pass env vars are also set
func TestGetDataSourcesEnvDSNWins(t *testing.T) {
	clearDataSourceEnv(t)
	const envDSN = "postgresql://userDsn:passwordDsn@localhost:55432/?sslmode=disabled"
	t.Setenv("DATA_SOURCE_NAME", envDSN)
	t.Setenv("DATA_SOURCE_USER_FILE", writeSecretFile(t, "user", testUser))
	t.Setenv("DATA_SOURCE_PASS", "envUserPass")

	dsn, err := getDataSources(dataSourceOpts{})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 1 || dsn[0] != envDSN {
		t.Fatalf("getDataSources() = %v, want [%v]", dsn, envDSN)
	}
}

// --datasource.* flags: secrets read from the files they point at, uri from the flag
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

// --datasource.uri-file takes precedence over --datasource.uri
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

// DATA_SOURCE_NAME wins over the datasource.* flags
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

// no datasource configured at all: multi-target pattern, no error
func TestGetDataSourcesEmpty(t *testing.T) {
	clearDataSourceEnv(t)
	dsn, err := getDataSources(dataSourceOpts{})
	if err != nil {
		t.Fatalf("getDataSources() error = %v", err)
	}
	if len(dsn) != 0 {
		t.Fatalf("getDataSources() = %v, want []", dsn)
	}
}
