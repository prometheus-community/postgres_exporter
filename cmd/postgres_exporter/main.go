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
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	prometheuscollectors "github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/exporter-toolkit/bootstrap"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	configFile                          = kingpin.Flag("config.file", "Postgres exporter configuration file.").Default("postgres_exporter.yml").String()
	disableDefaultMetrics               = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	autoDiscoverDatabases               = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	autoDiscoverDatabasesMaxConcurrency = kingpin.Flag("auto-discover-databases.max-concurrency", "Maximum number of discovered databases to scrape concurrently in a single scrape.").Default(fmt.Sprintf("%d", config.DefaultAutoDiscoverDatabasesMaxConcurrency)).Int()
	queriesPath                         = kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps                        = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList                  = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases                    = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases                    = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix                        = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").Default("pg").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	collectionTimeout                   = kingpin.Flag("collection-timeout", "Timeout for collecting the statistics when the database is slow").Default("1m").Envar("PG_EXPORTER_COLLECTION_TIMEOUT").String()
	wrapLargeCounters                   = kingpin.Flag("wrap-large-counters", "Wrap 64-bit counters at 2^53 to avoid floating point rounding.").Default("true").Bool()
	datasourceURI                       = kingpin.Flag("datasource.uri", "Data source URI, in host:port/dbname form.").Default("").String()
	datasourceURIFile                   = kingpin.Flag("datasource.uri-file", "Path to a file containing the data source URI. Takes precedence over datasource.uri.").Default("").String()
	datasourceUserFile                  = kingpin.Flag("datasource.user-file", "Path to a file containing the data source user name.").Default("").String()
	datasourcePassFile                  = kingpin.Flag("datasource.pass-file", "Path to a file containing the data source password.").Default("").String()
	collectorFlags                      = newCollectorFlags()

	longRunningTransactionsThreshold = kingpin.Flag(
		"collector.long_running_transactions.threshold",
		"Minimum transaction duration to consider long running.",
	).Default(config.DefaultLongRunningTransactionsThreshold.String()).Duration()

	statStatementsFlags = newPGStatStatementsFlags()
	logger              = promslog.NewNopLogger()
)

// The name of the exporter.
const exporterName = "postgres_exporter"

// webTelemetryPathEnvar is deprecated and will be removed in v0.22.0.
// Use --web.telemetry-path instead.
const webTelemetryPathEnvar = "PG_EXPORTER_WEB_TELEMETRY_PATH"

// Bootstrap parses flags and builds the handler before starting the listener.
// Returning errDumpMaps from the handler factory stops startup before database
// initialization, so main can dump the maps and exit successfully.
var errDumpMaps = errors.New("dumpmaps requested")

type collectorFlagSet map[string]*bool

type pgStatStatementsFlags struct {
	includeQuery     *bool
	queryLength      *uint
	limit            *uint
	excludeDatabases *string
	excludeUsers     *string
}

func newCollectorFlags() collectorFlagSet {
	defaults := config.DefaultCollectorConfig()
	names := make([]string, 0, len(defaults))
	for name := range defaults {
		names = append(names, name)
	}
	sort.Strings(names)

	flags := make(collectorFlagSet, len(defaults))
	for _, name := range names {
		helpDefaultState := "disabled"
		if defaults[name] {
			helpDefaultState = "enabled"
		}
		flags[name] = kingpin.Flag(
			"collector."+name,
			fmt.Sprintf("Enable the %s collector (default: %s).", name, helpDefaultState),
		).Default(fmt.Sprintf("%v", defaults[name])).Bool()
	}
	return flags
}

func newPGStatStatementsFlags() pgStatStatementsFlags {
	return pgStatStatementsFlags{
		includeQuery: kingpin.Flag(
			"collector.stat_statements.include_query",
			"Enable selecting statement query together with queryId. (default: disabled)",
		).Default(strconv.FormatBool(config.DefaultPGStatStatementsIncludeQuery)).Bool(),
		queryLength: kingpin.Flag(
			"collector.stat_statements.query_length",
			"Maximum length of the statement text.",
		).Default(strconv.FormatUint(uint64(config.DefaultPGStatStatementsQueryLength), 10)).Uint(),
		limit: kingpin.Flag(
			"collector.stat_statements.limit",
			"Maximum number of statements to return.",
		).Default(fmt.Sprintf("%d", config.DefaultPGStatStatementsLimit)).Uint(),
		excludeDatabases: kingpin.Flag(
			"collector.stat_statements.exclude_databases",
			"Comma-separated list of database names to exclude. (default: none)",
		).Default("").String(),
		excludeUsers: kingpin.Flag(
			"collector.stat_statements.exclude_users",
			"Comma-separated list of user names to exclude. (default: none)",
		).Default("").String(),
	}
}

func main() {
	server := &exporterServer{}
	runner := bootstrap.New(bootstrap.Config{
		App:            kingpin.CommandLine,
		Name:           exporterName,
		Description:    "Prometheus PostgreSQL server Exporter",
		DefaultAddress: ":9187",
		LandingConfig: web.LandingConfig{
			Name: "Postgres Exporter",
		},
		MetricsHandlerFactory: server.newMetricsHandler,
	})
	// This flag previously belonged to postgres_exporter and read this env var.
	// Bootstrap registers it without an env binding. Restore that binding until
	// v0.22.0, preserving explicit CLI flags' precedence.
	kingpin.CommandLine.GetFlag("web.telemetry-path").Envar(webTelemetryPathEnvar)

	err := runner.Run()
	server.close()
	if errors.Is(err, errDumpMaps) {
		exporter.DumpMaps()
		return
	}
	if err != nil {
		if runner.Logger != nil {
			runner.Logger.Error("Error running exporter", "err", err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

type exporterServer struct {
	runtime    *collector.Runtime
	stopReload func()
}

func (s *exporterServer) close() {
	if s.stopReload != nil {
		s.stopReload()
	}
	if s.runtime != nil {
		if err := s.runtime.Close(); err != nil {
			logger.Error("Failed to close runtime", "err", err)
		}
	}
}

// newMetricsHandler wires up the exporter and registers postgres_exporter's
// endpoints once the toolkit has parsed flags.
func (s *exporterServer) newMetricsHandler(b *bootstrap.Bootstrap) (http.Handler, error) {
	logger = b.Logger

	if *onlyDumpMaps {
		return nil, errDumpMaps
	}

	registry := prometheus.NewRegistry()
	authHandler, err := config.NewHandler(registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create config handler: %w", err)
	}

	if err := authHandler.ReloadAuthConfig(*configFile, logger); err != nil {
		// This is not fatal, but it means that auth must be provided for every dsn.
		logger.Warn("Error loading config", "err", err)
	}

	warnDeprecatedDataSourceEnvVars()

	if os.Getenv(webTelemetryPathEnvar) != "" {
		logger.Warn(webTelemetryPathEnvar + " is DEPRECATED, use --web.telemetry-path instead")
	}

	dsns, err := getDataSources(dataSourceOpts{
		URI:      *datasourceURI,
		URIFile:  *datasourceURIFile,
		UserFile: *datasourceUserFile,
		PassFile: *datasourcePassFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed reading data sources: %w", err)
	}

	cfg, err := buildConfig(dsns)
	if err != nil {
		return nil, fmt.Errorf("failed building config: %w", err)
	}
	logger.Info("Excluded databases", "databases", fmt.Sprintf("%v", cfg.ExcludeDatabases))

	if cfg.UserQueriesPath != "" {
		logger.Warn("The extended queries.yaml config is DEPRECATED", "file", cfg.UserQueriesPath)
	}

	if cfg.AutoDiscoverDatabases || *excludeDatabases != "" || *includeDatabases != "" {
		logger.Warn("Scraping additional databases via auto discovery is DEPRECATED")
	}

	if cfg.ConstantLabels != "" {
		logger.Warn("Constant labels on all metrics is DEPRECATED")
	}

	validatedConfig, err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	runtime, err := collector.NewRuntime(validatedConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}
	s.runtime = runtime

	if !b.DisableExporterMetrics {
		registry.MustRegister(
			versioncollector.NewCollector(exporterName),
			prometheuscollectors.NewGoCollector(),
			prometheuscollectors.NewProcessCollector(prometheuscollectors.ProcessCollectorOpts{}),
		)
	}
	for _, collector := range runtime.Collectors() {
		registry.MustRegister(collector)
	}

	b.Handle("/probe", b.MaxRequestsHandler(handleProbe(logger, authHandler, cfg)))
	b.Handle("/debug/", http.DefaultServeMux)

	metricsHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	b.HandleFunc("/-/reload", handleReload(authHandler, *configFile))

	reloadSignals := make(chan os.Signal, 1)
	reloadDone := make(chan struct{})
	signal.Notify(reloadSignals, syscall.SIGHUP)
	go func() {
		defer close(reloadDone)
		handleReloadSignal(reloadSignals, authHandler, *configFile, logger)
	}()
	s.stopReload = func() {
		signal.Stop(reloadSignals)
		close(reloadSignals)
		<-reloadDone
	}

	return metricsHandler, nil
}

func buildConfig(dsns []string) (config.Config, error) {
	parsedCollectionTimeout, err := parseCollectionTimeout(*collectionTimeout)
	if err != nil {
		return config.Config{}, err
	}

	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = dsns
	cfg.MetricPrefix = *metricPrefix
	cfg.CollectionTimeout = parsedCollectionTimeout
	cfg.WrapLargeCounters = *wrapLargeCounters
	cfg.DisableDefaultMetrics = *disableDefaultMetrics
	cfg.AutoDiscoverDatabases = *autoDiscoverDatabases
	cfg.AutoDiscoverDatabasesMaxConcurrency = *autoDiscoverDatabasesMaxConcurrency
	cfg.UserQueriesPath = *queriesPath
	cfg.ConstantLabels = *constantLabelsList
	cfg.ExcludeDatabases = splitList(*excludeDatabases)
	cfg.IncludeDatabases = splitList(*includeDatabases)
	cfg.Collectors = collectorFlags.states()
	cfg.LongRunningTransactions = config.LongRunningTransactionsConfig{
		Threshold: *longRunningTransactionsThreshold,
	}
	cfg.PGStatStatements = config.PGStatStatementsConfig{
		IncludeQuery:     *statStatementsFlags.includeQuery,
		QueryLength:      *statStatementsFlags.queryLength,
		Limit:            *statStatementsFlags.limit,
		ExcludeDatabases: splitList(*statStatementsFlags.excludeDatabases),
		ExcludeUsers:     splitList(*statStatementsFlags.excludeUsers),
	}
	return cfg, nil
}

func (flags collectorFlagSet) states() map[string]bool {
	states := make(map[string]bool, len(flags))
	for name, value := range flags {
		states[name] = *value
	}
	return states
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseCollectionTimeout(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid collection timeout %q: %w", value, err)
	}
	return duration, nil
}
