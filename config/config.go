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
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	"github.com/prometheus-community/postgres_exporter/internal/connector"
)

const (
	DefaultMetricPrefix      string        = "pg"
	DefaultCollectionTimeout time.Duration = time.Minute

	DefaultLongRunningTransactionsThreshold time.Duration = time.Minute

	// DefaultAutoDiscoverDatabasesMaxConcurrency bounds how many discovered
	// databases are scraped concurrently, so a server with many databases
	// cannot make a single scrape open more connections than it can spare.
	DefaultAutoDiscoverDatabasesMaxConcurrency int = 10

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
	WrapLargeCounters     bool
	DisableDefaultMetrics bool
	AutoDiscoverDatabases bool
	// AutoDiscoverDatabasesMaxConcurrency bounds how many discovered
	// databases are scraped concurrently when AutoDiscoverDatabases is set.
	AutoDiscoverDatabasesMaxConcurrency int
	UserQueriesPath                     string
	ConstantLabels                      string
	ExcludeDatabases                    []string
	IncludeDatabases                    []string
	Collectors                          map[string]bool
	LongRunningTransactions             LongRunningTransactionsConfig
	PGStatStatements                    PGStatStatementsConfig
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

type LongRunningTransactionsConfig struct {
	Threshold time.Duration
}

func NewConfigWithDefaults() Config {
	return Config{
		MetricPrefix:                        DefaultMetricPrefix,
		CollectionTimeout:                   DefaultCollectionTimeout,
		WrapLargeCounters:                   true,
		AutoDiscoverDatabasesMaxConcurrency: DefaultAutoDiscoverDatabasesMaxConcurrency,
		Collectors:                          DefaultCollectorConfig(),
		LongRunningTransactions: LongRunningTransactionsConfig{
			Threshold: DefaultLongRunningTransactionsThreshold,
		},
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
	if c.LongRunningTransactions.Threshold <= 0 {
		return ValidatedConfig{}, fmt.Errorf("long running transactions threshold must be greater than zero")
	}
	if c.AutoDiscoverDatabases && c.AutoDiscoverDatabasesMaxConcurrency <= 0 {
		return ValidatedConfig{}, fmt.Errorf("auto-discover-databases max concurrency must be greater than zero")
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
	IAM      IAM      `yaml:"iam,omitempty"`
	// Add alternative auth modules here
	Options map[string]string `yaml:"options"`
}

type UserPass struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type IAM struct {
	Region   string `yaml:"region,omitempty"`
	RoleARN  string `yaml:"role_arn,omitempty"`
	DBUser   string `yaml:"db_user,omitempty"`
	Database string `yaml:"db_name,omitempty"`
}

type Handler struct {
	sync.RWMutex
	Config *AuthConfig

	// connectors caches driver.Connectors built by AuthModule.Connector, keyed
	// by auth module name and target, so expensive per-connector setup (e.g.
	// the iam module's AWS credentials) isn't rebuilt on every /probe request.
	// It is cleared whenever the auth config is reloaded.
	connectors map[string]driver.Connector

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
	ch.connectors = nil
	ch.Unlock()
}

// maxCachedConnectors bounds the number of connectors Handler retains.
const maxCachedConnectors = 128

// ConnectorFor returns a cached driver.Connector for authModuleName+target,
// building one via authModule.Connector on first use. This works for any auth
// module type, since it caches whatever Connector returns without needing to
// know what's inside it.
func (ch *Handler) ConnectorFor(authModuleName string, authModule AuthModule, target string) (driver.Connector, error) {
	// Only IAM connectors are worth caching: they hold AWS credential state
	// that is expensive to rebuild. Static connectors are cheap.
	if authModule.Type != "iam" {
		return authModule.Connector(target)
	}

	// The key includes the module's own configuration, so a request that
	// resolved its module before a reload can never populate an entry that a
	// request using the reloaded configuration would find.
	moduleJSON, err := json.Marshal(authModule)
	if err != nil {
		return nil, fmt.Errorf("encoding auth module %q: %w", authModuleName, err)
	}
	sum := sha256.Sum256(moduleJSON)
	key := hex.EncodeToString(sum[:]) + "\x00" + target

	// Building an IAM connector does no I/O, so it's cheap to do under the
	// lock; that keeps concurrent probes of one target from building duplicates.
	ch.Lock()
	defer ch.Unlock()
	if conn, ok := ch.connectors[key]; ok {
		return conn, nil
	}
	conn, err := authModule.Connector(target)
	if err != nil {
		return nil, err
	}
	if ch.connectors == nil {
		ch.connectors = make(map[string]driver.Connector)
	} else if len(ch.connectors) >= maxCachedConnectors {
		// Bound memory without discarding every cached credential provider.
		for cachedKey := range ch.connectors {
			delete(ch.connectors, cachedKey)
			break
		}
	}
	ch.connectors[key] = conn
	return conn, nil
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

// Connector builds a driver.Connector
func (m AuthModule) Connector(target string) (driver.Connector, error) {
	if m.Type == "iam" {
		dsn, err := dsnFromString(target)
		if err != nil {
			return nil, err
		}

		// db_user/db_name override target when set, else they're parsed from it.
		dbUser := dsn.username
		if m.IAM.DBUser != "" {
			dbUser = m.IAM.DBUser
		}
		// A key=value DSN (or a ?dbname= URL) carries the database in the
		// query rather than the path.
		database := strings.TrimPrefix(dsn.path, "/")
		if q := dsn.query.Get("dbname"); q != "" {
			database = q
		}
		if name := m.Options["dbname"]; name != "" {
			database = name
		}
		if m.IAM.Database != "" {
			database = m.IAM.Database
		}
		if dbUser == "" {
			return nil, errors.New(`auth module type "iam" requires a db user, from either iam.db_user or target`)
		}
		if database == "" {
			return nil, errors.New(`auth module type "iam" requires a database name, from either iam.db_name or target`)
		}

		// options override target's own query parameters when set, same as for userpass.
		options := url.Values{}
		maps.Copy(options, dsn.query)
		for k, v := range m.Options {
			options.Set(k, v)
		}
		switch sslmode := options.Get("sslmode"); sslmode {
		case "":
			options.Set("sslmode", "require")
		case "disable", "allow", "prefer":
			return nil, fmt.Errorf("IAM authentication requires TLS; sslmode %q permits plaintext connections", sslmode)
		}
		options.Del("user")
		options.Del("password")
		// The database is carried by the path; drop any inherited dbname so
		// it can't override the database selected above.
		options.Del("dbname")
		if m.IAM.Database != "" || m.Options["dbname"] != "" {
			options.Del("database")
		}
		// Resolve the port lib/pq will actually dial: a port query parameter
		// overrides the URL authority, and a portless target falls back to
		// PGPORT. It's then written into the connection string explicitly, so
		// the endpoint we sign for and the one we connect to can't diverge.
		hostname, port := dsn.host, ""
		if h, p, err := net.SplitHostPort(dsn.host); err == nil {
			hostname, port = h, p
		}
		if h := options.Get("host"); h != "" {
			hostname = h
		}
		if q := options.Get("port"); q != "" {
			port = q
		}
		if port == "" {
			port = os.Getenv("PGPORT")
		}
		if port == "" {
			port = "5432"
		}
		options.Del("port")
		endpoint := net.JoinHostPort(hostname, port)

		// base is the connection, minus the password: the same shape every other
		// auth mode builds via GetConnectionString, so a freshly minted IAM token
		// only ever needs to be injected via WithPassword.
		base := DSN{
			scheme:   dsn.scheme,
			username: dbUser,
			host:     endpoint,
			path:     "/" + database,
			query:    options,
		}

		c, err := connector.NewAWSIAMConnector(
			endpoint,
			dbUser,
			m.IAM.Region,
			m.IAM.RoleARN,
			func(token string) string { return base.WithPassword(token).GetConnectionString() },
		)
		if err != nil {
			return nil, fmt.Errorf("configuring iam connector: %w", err)
		}
		return c, nil
	}

	dsn, err := m.ConfigureTarget(target)
	if err != nil {
		return nil, err
	}
	return connector.NewStaticConnector(dsn.GetConnectionString())
}
