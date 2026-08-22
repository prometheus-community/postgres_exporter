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
	"maps"
	"os"
	"slices"
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
}

// ValidatedConfig is the result of a successful Config.Validate call. It holds
// a private deep copy of the validated Config, so later mutations of the
// original cannot invalidate it. Consumers that require validated
// configuration (e.g. collector.NewRuntime) accept this type instead of
// Config, making validation impossible to skip.
type ValidatedConfig struct {
	inner Config
	ok    bool
}

// Valid reports whether this value was produced by a successful
// Config.Validate call. It only returns false for zero-value ValidatedConfig
// structs that bypassed validation.
func (v ValidatedConfig) Valid() bool {
	return v.ok
}

// Config returns a deep copy of the validated configuration. Mutating the
// returned value does not affect the validated state.
func (v ValidatedConfig) Config() Config {
	return v.inner.clone()
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

// Validate checks the configuration and, on success, returns a
// ValidatedConfig holding a deep copy of it. Validation runs against the copy,
// so concurrent mutations of the caller's Config cannot affect the outcome.
func (c Config) Validate() (ValidatedConfig, error) {
	c = c.clone()

	if c.MetricPrefix == "" {
		return ValidatedConfig{}, fmt.Errorf("metric prefix must not be empty")
	}
	if c.CollectionTimeout <= 0 {
		return ValidatedConfig{}, fmt.Errorf("collection timeout must be greater than zero")
	}
	for i, dsn := range c.DataSourceNames {
		if dsn == "" {
			return ValidatedConfig{}, fmt.Errorf("data source name at index %d must not be empty", i)
		}
	}
	if c.PGStatStatements.QueryLength <= 0 {
		return ValidatedConfig{}, fmt.Errorf("pg_stat_statements query length must be greater than zero")
	}
	if c.PGStatStatements.Limit <= 0 {
		return ValidatedConfig{}, fmt.Errorf("pg_stat_statements limit must be greater than zero")
	}
	for name := range c.Collectors {
		if name == "" {
			return ValidatedConfig{}, fmt.Errorf("collector name must not be empty")
		}
		if _, ok := DefaultCollectorConfig()[name]; !ok {
			return ValidatedConfig{}, fmt.Errorf("unknown collector %q", name)
		}
	}

	return ValidatedConfig{inner: c, ok: true}, nil
}

// clone returns a copy of the Config with all reference-bearing fields
// (slices and maps) deep-copied, so the copy shares no mutable state with the
// original.
func (c Config) clone() Config {
	c.DataSourceNames = slices.Clone(c.DataSourceNames)
	c.ExcludeDatabases = slices.Clone(c.ExcludeDatabases)
	c.IncludeDatabases = slices.Clone(c.IncludeDatabases)
	c.Collectors = maps.Clone(c.Collectors)
	c.PGStatStatements.ExcludeDatabases = slices.Clone(c.PGStatStatements.ExcludeDatabases)
	c.PGStatStatements.ExcludeUsers = slices.Clone(c.PGStatStatements.ExcludeUsers)
	return c
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
	if f == "" {
		return &AuthConfig{}, nil
	}

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
