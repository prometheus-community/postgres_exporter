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
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/postgres_exporter/internal/connector"
)

func TestNewConfigWithDefaults(t *testing.T) {
	cfg := NewConfigWithDefaults()

	if got, want := cfg.MetricPrefix, DefaultMetricPrefix; got != want {
		t.Fatalf("MetricPrefix = %q, want %q", got, want)
	}
	if got, want := cfg.CollectionTimeout, DefaultCollectionTimeout; got != want {
		t.Fatalf("CollectionTimeout = %v, want %v", got, want)
	}
	if got, want := cfg.LongRunningTransactions.Threshold, DefaultLongRunningTransactionsThreshold; got != want {
		t.Fatalf("LongRunningTransactions.Threshold = %v, want %v", got, want)
	}
	if !cfg.WrapLargeCounters {
		t.Fatal("WrapLargeCounters = false, want true")
	}
	if cfg.DisableDefaultMetrics {
		t.Fatal("DisableDefaultMetrics = true, want false")
	}
	if got, want := cfg.PGStatStatements.IncludeQuery, DefaultPGStatStatementsIncludeQuery; got != want {
		t.Fatalf("PGStatStatements.IncludeQuery = %t, want %t", got, want)
	}
	if len(cfg.Collectors) == 0 {
		t.Fatal("Collectors is empty, want default collector config")
	}
	if got, want := cfg.Collectors[CollectorDatabase], true; got != want {
		t.Fatalf("Collectors[%q] = %t, want %t", CollectorDatabase, got, want)
	}
	if got, want := cfg.Collectors[CollectorStatStatements], false; got != want {
		t.Fatalf("Collectors[%q] = %t, want %t", CollectorStatStatements, got, want)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}

	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !validated.Valid() {
		t.Fatal("Valid() = false after successful Validate, want true")
	}
	if got, want := validated.Config().MetricPrefix, cfg.MetricPrefix; got != want {
		t.Fatalf("Config().MetricPrefix = %q, want %q", got, want)
	}
}

func TestValidatedConfigZeroValueIsInvalid(t *testing.T) {
	var validated ValidatedConfig
	if validated.Valid() {
		t.Fatal("Valid() = true for zero-value ValidatedConfig, want false")
	}
}

func TestValidatedConfigIsIsolatedFromOriginal(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
	cfg.PGStatStatements.ExcludeDatabases = []string{"template0"}

	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	cfg.Collectors[CollectorDatabase] = false
	cfg.DataSourceNames[0] = "mutated"
	cfg.PGStatStatements.ExcludeDatabases[0] = "mutated"

	got := validated.Config()
	if !got.Collectors[CollectorDatabase] {
		t.Fatalf("Collectors[%q] = false after mutating original, want true", CollectorDatabase)
	}
	if want := "postgresql://localhost:5432/postgres?sslmode=disable"; got.DataSourceNames[0] != want {
		t.Fatalf("DataSourceNames[0] = %q after mutating original, want %q", got.DataSourceNames[0], want)
	}
	if want := "template0"; got.PGStatStatements.ExcludeDatabases[0] != want {
		t.Fatalf("PGStatStatements.ExcludeDatabases[0] = %q after mutating original, want %q", got.PGStatStatements.ExcludeDatabases[0], want)
	}
}

func TestValidatedConfigAccessorReturnsCopy(t *testing.T) {
	cfg := NewConfigWithDefaults()

	validated, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	first := validated.Config()
	first.Collectors[CollectorDatabase] = false

	second := validated.Config()
	if !second.Collectors[CollectorDatabase] {
		t.Fatalf("Collectors[%q] = false after mutating a previous copy, want true", CollectorDatabase)
	}
}

func TestConfigValidateFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{
			name: "empty metric prefix",
			mutate: func(cfg *Config) {
				cfg.MetricPrefix = ""
			},
			want: "metric prefix must not be empty",
		},
		{
			name: "zero collection timeout",
			mutate: func(cfg *Config) {
				cfg.CollectionTimeout = 0
			},
			want: "collection timeout must be greater than zero",
		},
		{
			name: "zero long running transactions threshold",
			mutate: func(cfg *Config) {
				cfg.LongRunningTransactions.Threshold = 0
			},
			want: "long running transactions threshold must be greater than zero",
		},
		{
			name: "empty data source",
			mutate: func(cfg *Config) {
				cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres", ""}
			},
			want: "data source name at index 1 must not be empty",
		},
		{
			name: "zero pg_stat_statements query length",
			mutate: func(cfg *Config) {
				cfg.PGStatStatements.QueryLength = 0
			},
			want: "pg_stat_statements query length must be greater than zero",
		},
		{
			name: "zero pg_stat_statements limit",
			mutate: func(cfg *Config) {
				cfg.PGStatStatements.Limit = 0
			},
			want: "pg_stat_statements limit must be greater than zero",
		},
		{
			name: "empty collector name",
			mutate: func(cfg *Config) {
				cfg.Collectors[""] = true
			},
			want: "collector name must not be empty",
		},
		{
			name: "unknown collector name",
			mutate: func(cfg *Config) {
				cfg.Collectors["does_not_exist"] = true
			},
			want: `unknown collector "does_not_exist"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := NewConfigWithDefaults()
			cfg.DataSourceNames = []string{"postgresql://localhost:5432/postgres?sslmode=disable"}
			test.mutate(&cfg)

			validated, err := cfg.Validate()
			if err == nil || err.Error() != test.want {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
			if validated.Valid() {
				t.Fatal("Valid() = true after failed Validate, want false")
			}
		})
	}
}

func TestConfigValidateAcceptsNoDataSourcesForMultiTargetMode(t *testing.T) {
	cfg := NewConfigWithDefaults()
	if _, err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateAcceptsCustomTimeout(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.CollectionTimeout = 30 * time.Second
	if _, err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLoadAuthConfigFile(t *testing.T) {
	config, err := LoadAuthConfig("testdata/config-good.yaml")
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}
	if len(config.AuthModules) == 0 {
		t.Fatal("LoadAuthConfig() loaded no auth modules")
	}
	if got, want := config.AuthModules["second"].IAM.Region, "us-east-1"; got != want {
		t.Fatalf(`AuthModules["second"].IAM.Region = %q, want %q`, got, want)
	}
}

func TestLoadAuthConfigEmptyPath(t *testing.T) {
	config, err := LoadAuthConfig("")
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("LoadAuthConfig() config = nil, want empty config")
	}
	if len(config.AuthModules) != 0 {
		t.Fatalf("LoadAuthConfig() loaded %d auth modules, want 0", len(config.AuthModules))
	}
}

func TestDecodeAuthConfig(t *testing.T) {
	config, err := DecodeAuthConfig(strings.NewReader(`
auth_modules:
  module:
    type: userpass
    userpass:
      username: user
      password: pass
`))
	if err != nil {
		t.Fatalf("DecodeAuthConfig() error = %v", err)
	}
	if got, want := config.AuthModules["module"].UserPass.Username, "user"; got != want {
		t.Fatalf("username = %q, want %q", got, want)
	}
}

func TestAuthModuleConnectorDefaultsToStatic(t *testing.T) {
	m := AuthModule{
		Type: "userpass",
		UserPass: UserPass{
			Username: "user",
			Password: "pass",
		},
	}

	conn, err := m.Connector("postgresql://localhost:5432/postgres")
	if err != nil {
		t.Fatalf("Connector() error = %v", err)
	}
	if _, ok := conn.(*connector.StaticConnector); !ok {
		t.Fatalf("Connector() type = %T, want *connector.StaticConnector", conn)
	}
}

func TestAuthModuleConnectorIAMRequiresDBUserAndDatabase(t *testing.T) {
	tests := []struct {
		name string
		iam  IAM
	}{
		{"missing both", IAM{Region: "us-east-1"}},
		{"missing database", IAM{Region: "us-east-1", DBUser: "iamuser"}},
		{"missing db_user", IAM{Region: "us-east-1", Database: "mydb"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AuthModule{Type: "iam", IAM: tt.iam}
			if _, err := m.Connector("postgresql://db.example.com:5432/"); err == nil {
				t.Fatal("Connector() error = nil, want error for missing iam.db_user/iam.db_name")
			}
		})
	}
}

func TestAuthModuleConnectorIAMRegionAndRoleARNAreOptional(t *testing.T) {
	m := AuthModule{
		Type: "iam",
		IAM: IAM{
			DBUser:   "iamuser",
			Database: "mydb",
		},
	}

	conn, err := m.Connector("postgresql://db.example.com:5432/")
	if err != nil {
		t.Fatalf("Connector() error = %v, want no error with region/role_arn unset", err)
	}

	iamConn, ok := conn.(*connector.AWSIAMConnector)
	if !ok {
		t.Fatalf("Connector() type = %T, want *connector.AWSIAMConnector", conn)
	}
	if iamConn.Region != "" {
		t.Errorf("Region = %q, want empty (resolved by the AWS SDK at connect time)", iamConn.Region)
	}
	if iamConn.RoleARN != "" {
		t.Errorf("RoleARN = %q, want empty (use ambient credentials directly)", iamConn.RoleARN)
	}
}

func TestAuthModuleConnectorIAM(t *testing.T) {
	m := AuthModule{
		Type: "iam",
		IAM: IAM{
			Region:   "us-east-1",
			RoleARN:  "arn:aws:iam::123456789012:role/rds-connect",
			DBUser:   "iamuser",
			Database: "mydb",
		},
	}

	// iam.db_user/iam.db_name override target's own targetuser/targetdb.
	conn, err := m.Connector("postgresql://targetuser@db.example.com:5432/targetdb")
	if err != nil {
		t.Fatalf("Connector() error = %v", err)
	}

	iamConn, ok := conn.(*connector.AWSIAMConnector)
	if !ok {
		t.Fatalf("Connector() type = %T, want *connector.AWSIAMConnector", conn)
	}
	if got, want := iamConn.DBEndpoint, "db.example.com:5432"; got != want {
		t.Errorf("DBEndpoint = %q, want %q", got, want)
	}
	if got, want := iamConn.DBUser, "iamuser"; got != want {
		t.Errorf("DBUser = %q, want %q", got, want)
	}
	if got, want := iamConn.Database, "mydb"; got != want {
		t.Errorf("Database = %q, want %q", got, want)
	}
	if got, want := iamConn.Region, "us-east-1"; got != want {
		t.Errorf("Region = %q, want %q", got, want)
	}
	if got, want := iamConn.RoleARN, "arn:aws:iam::123456789012:role/rds-connect"; got != want {
		t.Errorf("RoleARN = %q, want %q", got, want)
	}
}

func TestReloadAuthConfig(t *testing.T) {
	ch, err := NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	if err := ch.ReloadAuthConfig("testdata/config-good.yaml", nil); err != nil {
		t.Errorf("error loading config: %s", err)
	}
}

func TestNewHandlerRequiresRegisterer(t *testing.T) {
	handler, err := NewHandler(nil)
	if err == nil {
		t.Fatal("NewHandler() error = nil, want error")
	}
	if handler != nil {
		t.Fatalf("NewHandler() handler = %v, want nil", handler)
	}
}

func TestLoadBadConfigs(t *testing.T) {
	ch, err := NewHandler(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "testdata/config-bad-auth-module.yaml",
			want:  "error parsing config file \"testdata/config-bad-auth-module.yaml\": yaml: unmarshal errors:\n  line 3: field pretendauth not found in type config.AuthModule",
		},
		{
			input: "testdata/config-bad-extra-field.yaml",
			want:  "error parsing config file \"testdata/config-bad-extra-field.yaml\": yaml: unmarshal errors:\n  line 8: field doesNotExist not found in type config.AuthModule",
		},
		{
			input: "testdata/config-bad-iam-extra-field.yaml",
			want:  "error parsing config file \"testdata/config-bad-iam-extra-field.yaml\": yaml: unmarshal errors:\n  line 8: field doesNotExist not found in type config.IAM",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := ch.ReloadAuthConfig(test.input, nil)
			if got == nil || got.Error() != test.want {
				t.Fatalf("ReloadAuthConfig(%q) = %v, want %s", test.input, got, test.want)
			}
		})
	}
}
