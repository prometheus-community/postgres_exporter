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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

const (
	DefaultMetricPrefix      string        = "pg"
	DefaultCollectionTimeout time.Duration = time.Minute

	DefaultPGStatStatementsIncludeQuery bool = false
	DefaultPGStatStatementsQueryLength  uint = 120
	DefaultPGStatStatementsLimit        uint = 100
)

const (
	CollectorBuffercacheSummary      = "buffercache_summary"
	CollectorDatabase                = "database"
	CollectorDatabaseWraparound      = "database_wraparound"
	CollectorLocks                   = "locks"
	CollectorLongRunningTransactions = "long_running_transactions"
	CollectorPostmaster              = "postmaster"
	CollectorProcessIdle             = "process_idle"
	CollectorReplication             = "replication"
	CollectorReplicationSlots        = "replication_slots"
	CollectorRoles                   = "roles"
	CollectorSettings                = "settings"
	CollectorStatActivity            = "stat_activity"
	CollectorStatActivityAutovacuum  = "stat_activity_autovacuum"
	CollectorStatArchiver            = "stat_archiver"
	CollectorStatBGWriter            = "stat_bgwriter"
	CollectorStatCheckpointer        = "stat_checkpointer"
	CollectorStatDatabase            = "stat_database"
	CollectorStatProgressVacuum      = "stat_progress_vacuum"
	CollectorStatReplication         = "stat_replication"
	CollectorStatStatements          = "stat_statements"
	CollectorStatUserTables          = "stat_user_tables"
	CollectorStatWalReceiver         = "stat_wal_receiver"
	CollectorStatioUserIndexes       = "statio_user_indexes"
	CollectorStatioUserTables        = "statio_user_tables"
	CollectorWal                     = "wal"
	CollectorXlogLocation            = "xlog_location"
)

type Config struct {
	DataSourceNames       []string
	MetricPrefix          string
	CollectionTimeout     time.Duration
	DisableDefaultMetrics bool
	AutoDiscoverDatabases bool
	UserQueriesPath       string
	ConstantLabels        string
	ExcludeDatabases      []string
	IncludeDatabases      []string
	Collectors            map[string]bool
	PGStatStatements      PGStatStatementsConfig

	validated bool
}

type PGStatStatementsConfig struct {
	IncludeQuery     bool
	QueryLength      uint
	Limit            uint
	ExcludeDatabases []string
	ExcludeUsers     []string
}

func NewConfigWithDefaults() Config {
	return Config{
		MetricPrefix:      DefaultMetricPrefix,
		CollectionTimeout: DefaultCollectionTimeout,
		Collectors:        DefaultCollectorConfig(),
		PGStatStatements: PGStatStatementsConfig{
			IncludeQuery: DefaultPGStatStatementsIncludeQuery,
			QueryLength:  DefaultPGStatStatementsQueryLength,
			Limit:        DefaultPGStatStatementsLimit,
		},
	}
}

func (c *Config) Validate() error {
	c.validated = false

	if c.MetricPrefix == "" {
		return fmt.Errorf("metric prefix must not be empty")
	}
	if c.CollectionTimeout <= 0 {
		return fmt.Errorf("collection timeout must be greater than zero")
	}
	for i, dsn := range c.DataSourceNames {
		if dsn == "" {
			return fmt.Errorf("data source name at index %d must not be empty", i)
		}
	}
	if c.PGStatStatements.QueryLength == 0 {
		return fmt.Errorf("pg_stat_statements query length must be greater than zero")
	}
	if c.PGStatStatements.Limit == 0 {
		return fmt.Errorf("pg_stat_statements limit must be greater than zero")
	}
	for name := range c.Collectors {
		if name == "" {
			return fmt.Errorf("collector name must not be empty")
		}
		if _, ok := DefaultCollectorConfig()[name]; !ok {
			return fmt.Errorf("unknown collector %q", name)
		}
	}

	c.validated = true
	return nil
}

func (c Config) Validated() bool {
	return c.validated
}

func DefaultCollectorConfig() map[string]bool {
	return map[string]bool{
		CollectorBuffercacheSummary:      false,
		CollectorDatabase:                true,
		CollectorDatabaseWraparound:      false,
		CollectorLocks:                   true,
		CollectorLongRunningTransactions: false,
		CollectorPostmaster:              false,
		CollectorProcessIdle:             false,
		CollectorReplication:             true,
		CollectorReplicationSlots:        true,
		CollectorRoles:                   true,
		CollectorSettings:                true,
		CollectorStatActivity:            true,
		CollectorStatActivityAutovacuum:  false,
		CollectorStatArchiver:            true,
		CollectorStatBGWriter:            true,
		CollectorStatCheckpointer:        false,
		CollectorStatDatabase:            true,
		CollectorStatProgressVacuum:      true,
		CollectorStatReplication:         true,
		CollectorStatStatements:          false,
		CollectorStatUserTables:          true,
		CollectorStatWalReceiver:         false,
		CollectorStatioUserIndexes:       false,
		CollectorStatioUserTables:        true,
		CollectorWal:                     true,
		CollectorXlogLocation:            false,
	}
}

type AuthConfig struct {
	AuthModules map[string]AuthModule `yaml:"auth_modules"`
}

type AuthModule struct {
	Type     string   `yaml:"type"`
	UserPass UserPass `yaml:"userpass,omitempty"`
	// Add alternative auth modules here
	Options map[string]string `yaml:"options"`
}

type UserPass struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Handler struct {
	sync.RWMutex
	Config *AuthConfig

	configReloadSuccess prometheus.Gauge
	configReloadSeconds prometheus.Gauge
}

func NewHandler(registerer prometheus.Registerer) (*Handler, error) {
	if registerer == nil {
		return nil, errors.New("registerer is required")
	}
	h := &Handler{
		Config: &AuthConfig{},
		configReloadSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "postgres_exporter",
			Name:      "config_last_reload_successful",
			Help:      "Postgres exporter config loaded successfully.",
		}),
		configReloadSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "postgres_exporter",
			Name:      "config_last_reload_success_timestamp_seconds",
			Help:      "Timestamp of the last successful configuration reload.",
		}),
	}
	registerer.MustRegister(h.configReloadSuccess, h.configReloadSeconds)

	return h, nil
}

func (ch *Handler) GetAuthConfig() *AuthConfig {
	ch.RLock()
	defer ch.RUnlock()
	return ch.Config
}

func (ch *Handler) ReloadAuthConfig(f string, logger *slog.Logger) error {
	var err error
	defer func() {
		ch.observeReload(err)
	}()

	config, err := LoadAuthConfig(f)
	if err != nil {
		return err
	}

	ch.SetAuthConfig(config)
	return nil
}

func (ch *Handler) observeReload(err error) {
	if ch.configReloadSuccess == nil {
		return
	}
	if err != nil {
		ch.configReloadSuccess.Set(0)
		return
	}
	ch.configReloadSuccess.Set(1)
	if ch.configReloadSeconds != nil {
		ch.configReloadSeconds.SetToCurrentTime()
	}
}

func LoadAuthConfig(f string) (*AuthConfig, error) {
	yamlReader, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("error opening config file %q: %s", f, err)
	}
	defer yamlReader.Close()

	config, err := DecodeAuthConfig(yamlReader)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %s", f, err)
	}
	return config, nil
}

func DecodeAuthConfig(r io.Reader) (*AuthConfig, error) {
	config := &AuthConfig{}
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (ch *Handler) SetAuthConfig(config *AuthConfig) {
	ch.Lock()
	ch.Config = config
	ch.Unlock()
}

func (m AuthModule) ConfigureTarget(target string) (DSN, error) {
	dsn, err := dsnFromString(target)
	if err != nil {
		return DSN{}, err
	}

	// Set the credentials from the authentication module
	// TODO(@sysadmind): What should the order of precedence be?
	if m.Type == "userpass" {
		if m.UserPass.Username != "" {
			dsn.username = m.UserPass.Username
		}
		if m.UserPass.Password != "" {
			dsn.password = m.UserPass.Password
		}
	}

	for k, v := range m.Options {
		dsn.query.Set(k, v)
	}

	return dsn, nil
}
