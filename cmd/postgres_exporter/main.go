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
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/postgres_exporter/collectors"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus-community/postgres_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	c *config.Handler

	configFile            = kingpin.Flag("config.file", "Postgres exporter configuration file.").Default("postgres_exporter.yml").String()
	webConfig             = kingpinflag.AddFlags(kingpin.CommandLine, ":9187")
	metricsPath           = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	autoDiscoverDatabases = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	queriesPath           = kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps          = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList    = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases      = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases      = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix          = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").Default("pg").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	collectionTimeout     = kingpin.Flag("collection-timeout", "Timeout for collecting the statistics when the database is slow").Default("1m").Envar("PG_EXPORTER_COLLECTION_TIMEOUT").String()
	collectorFlags        = newCollectorFlags()
	statStatementsFlags   = newPGStatStatementsFlags()
	logger                = promslog.NewNopLogger()
)

// The name of the exporter.
const exporterName = "postgres_exporter"

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
	kingpin.Version(version.Print(exporterName))
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promslog.New(promslogConfig)

	if *onlyDumpMaps {
		exporter.DumpMaps()
		return
	}

	registry := prometheus.NewRegistry()
	var err error
	c, err = config.NewHandler(registry)
	if err != nil {
		logger.Error("Failed to create config handler", "err", err)
		os.Exit(1)
	}

	if err := c.ReloadAuthConfig(*configFile, logger); err != nil {
		// This is not fatal, but it means that auth must be provided for every dsn.
		logger.Warn("Error loading config", "err", err)
	}

	dsns, err := exporter.GetDataSources()
	if err != nil {
		logger.Error("Failed reading data sources", "err", err.Error())
		os.Exit(1)
	}

	cfg := buildConfig(dsns)
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

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid config", "err", err)
		os.Exit(1)
	}

	pgRuntime, err := collectors.NewRuntime(&cfg, logger)
	if err != nil {
		logger.Error("Failed to create runtime", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := pgRuntime.Close(); err != nil {
			logger.Error("Failed to close runtime", "err", err)
		}
	}()

	registry.MustRegister(versioncollector.NewCollector(exporterName))
	for _, collector := range pgRuntime.Collectors() {
		registry.MustRegister(collector)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Postgres Exporter",
			Description: "Prometheus PostgreSQL server Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error("error creating landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	http.HandleFunc("/probe", handleProbe(logger, cfg))

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		logger.Error("Error running HTTP server", "err", err)
		os.Exit(1)
	}
}

func buildConfig(dsns []string) config.Config {
	cfg := config.NewConfigWithDefaults()
	cfg.DataSourceNames = dsns
	cfg.MetricPrefix = *metricPrefix
	cfg.CollectionTimeout = mustParseDurationFlag(*collectionTimeout)
	cfg.DisableDefaultMetrics = *disableDefaultMetrics
	cfg.AutoDiscoverDatabases = *autoDiscoverDatabases
	cfg.UserQueriesPath = *queriesPath
	cfg.ConstantLabels = *constantLabelsList
	cfg.ExcludeDatabases = splitList(*excludeDatabases)
	cfg.IncludeDatabases = splitList(*includeDatabases)
	cfg.Collectors = collectorFlags.states()
	cfg.PGStatStatements = config.PGStatStatementsConfig{
		IncludeQuery:     *statStatementsFlags.includeQuery,
		QueryLength:      *statStatementsFlags.queryLength,
		Limit:            *statStatementsFlags.limit,
		ExcludeDatabases: splitList(*statStatementsFlags.excludeDatabases),
		ExcludeUsers:     splitList(*statStatementsFlags.excludeUsers),
	}
	return cfg
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

func mustParseDurationFlag(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Keep Kingpin's current parse-later behavior but fail clearly during
		// config validation/construction.
		return 0
	}
	return duration
}
