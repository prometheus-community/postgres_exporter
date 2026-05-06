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
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/postgres_exporter/collector"
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

type legacyFlags struct {
	AutoDiscoverDatabases bool
	UserQueriesPath       string
	ConstantLabels        string
	ExcludeDatabases      string
	IncludeDatabases      string
}

var (
	c                              = newAuthConfigHandler()
	cfg                            = config.NewConfigWithDefaults()
	legacyMetricsFlags             = legacyFlags{}
	statStatementsExcludeDatabases string
	statStatementsExcludeUsers     string

	webConfig    = kingpinflag.AddFlags(kingpin.CommandLine, ":9187")
	metricsPath  = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	onlyDumpMaps = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	logger       = promslog.NewNopLogger()
)

func init() {
	kingpin.Flag("config.file", "Postgres exporter configuration file.").
		Default(cfg.AuthConfigFile).
		StringVar(&cfg.AuthConfigFile)
	kingpin.Flag("disable-default-metrics", "Do not include default metrics.").
		Default("false").
		Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").
		BoolVar(&cfg.DisableDefaultMetrics)
	kingpin.Flag("disable-settings-metrics", "Do not include pg_settings metrics.").
		Default("false").
		Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").
		BoolVar(&cfg.DisableSettingsMetrics)
	kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").
		Default(cfg.MetricPrefix).
		Envar("PG_EXPORTER_METRIC_PREFIX").
		StringVar(&cfg.MetricPrefix)
	kingpin.Flag("collection-timeout", "Timeout for collecting the statistics when the database is slow").
		Default(cfg.CollectionTimeout.String()).
		Envar("PG_EXPORTER_COLLECTION_TIMEOUT").
		DurationVar(&cfg.CollectionTimeout)
	registerCollectorFlags()
	registerStatStatementsFlags()

	kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").
		Default("false").
		Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").
		BoolVar(&legacyMetricsFlags.AutoDiscoverDatabases)
	kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").
		Default("").
		Envar("PG_EXPORTER_EXTEND_QUERY_PATH").
		StringVar(&legacyMetricsFlags.UserQueriesPath)
	kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").
		Default("").
		Envar("PG_EXPORTER_CONSTANT_LABELS").
		StringVar(&legacyMetricsFlags.ConstantLabels)
	kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").
		Default("").
		Envar("PG_EXPORTER_EXCLUDE_DATABASES").
		StringVar(&legacyMetricsFlags.ExcludeDatabases)
	kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").
		Default("").
		Envar("PG_EXPORTER_INCLUDE_DATABASES").
		StringVar(&legacyMetricsFlags.IncludeDatabases)
}

func registerCollectorFlags() {
	for i := range cfg.Collectors {
		helpDefaultState := "disabled"
		if cfg.Collectors[i].DefaultEnabled {
			helpDefaultState = "enabled"
		}
		kingpin.Flag(
			collector.CollectorFlagPrefix+cfg.Collectors[i].Name,
			fmt.Sprintf("Enable the %s collector (default: %s).", cfg.Collectors[i].Name, helpDefaultState),
		).
			Default(fmt.Sprintf("%v", cfg.Collectors[i].DefaultEnabled)).
			BoolVar(&cfg.Collectors[i].Enabled)
	}
}

func registerStatStatementsFlags() {
	flagPrefix := collector.CollectorFlagPrefix + collector.StatStatementsCollectorName
	kingpin.Flag(flagPrefix+".include_query", "Enable selecting statement query together with queryId. (default: disabled)").
		Default("false").
		BoolVar(&cfg.StatStatements.IncludeQuery)
	kingpin.Flag(flagPrefix+".query_length", "Maximum length of the statement text.").
		Default(fmt.Sprintf("%d", cfg.StatStatements.QueryLength)).
		UintVar(&cfg.StatStatements.QueryLength)
	kingpin.Flag(flagPrefix+".limit", "Maximum number of statements to return.").
		Default(fmt.Sprintf("%d", cfg.StatStatements.Limit)).
		UintVar(&cfg.StatStatements.Limit)
	kingpin.Flag(flagPrefix+".exclude_databases", "Comma-separated list of database names to exclude. (default: none)").
		Default("").
		StringVar(&statStatementsExcludeDatabases)
	kingpin.Flag(flagPrefix+".exclude_users", "Comma-separated list of user names to exclude. (default: none)").
		Default("").
		StringVar(&statStatementsExcludeUsers)
}

// The name of the exporter.
const exporterName = "postgres_exporter"

func newAuthConfigHandler() *config.AuthConfigHandler {
	handler, err := config.NewAuthConfigHandler(prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}
	return handler
}

func main() {
	kingpin.Version(version.Print(exporterName))
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promslog.New(promslogConfig)
	cfg.StatStatements.ExcludeDatabases = parseCommaSeparatedList(statStatementsExcludeDatabases)
	cfg.StatStatements.ExcludeUsers = parseCommaSeparatedList(statStatementsExcludeUsers)

	if *onlyDumpMaps {
		exporter.DumpMaps()
		return
	}

	if err := c.ReloadAuthConfig(cfg.AuthConfigFile, logger); err != nil {
		// This is not fatal, but it means that auth must be provided for every dsn.
		logger.Warn("Error loading config", "err", err)
	}

	dsns, err := exporter.GetDataSources()
	if err != nil {
		logger.Error("Failed reading data sources", "err", err.Error())
		os.Exit(1)
	}
	cfg.DataSourceNames = dsns

	excludedDatabases := strings.Split(legacyMetricsFlags.ExcludeDatabases, ",")
	logger.Info("Excluded databases", "databases", fmt.Sprintf("%v", excludedDatabases))

	if legacyMetricsFlags.UserQueriesPath != "" {
		logger.Warn("The extended queries.yaml config is DEPRECATED", "file", legacyMetricsFlags.UserQueriesPath)
	}

	if legacyMetricsFlags.AutoDiscoverDatabases || legacyMetricsFlags.ExcludeDatabases != "" || legacyMetricsFlags.IncludeDatabases != "" {
		logger.Warn("Scraping additional databases via auto discovery is DEPRECATED")
	}

	if legacyMetricsFlags.ConstantLabels != "" {
		logger.Warn("Constant labels on all metrics is DEPRECATED")
	}

	opts := []exporter.ExporterOpt{
		exporter.DisableDefaultMetrics(cfg.DisableDefaultMetrics),
		exporter.DisableSettingsMetrics(cfg.DisableSettingsMetrics),
		exporter.AutoDiscoverDatabases(legacyMetricsFlags.AutoDiscoverDatabases),
		exporter.WithUserQueriesPath(legacyMetricsFlags.UserQueriesPath),
		exporter.WithConstantLabels(legacyMetricsFlags.ConstantLabels),
		exporter.ExcludeDatabases(excludedDatabases),
		exporter.IncludeDatabases(legacyMetricsFlags.IncludeDatabases),
		exporter.WithMetricPrefix(cfg.MetricPrefix),
	}

	exporter := exporter.NewExporter(dsns, logger, opts...)
	defer func() {
		exporter.CloseServers()
	}()

	prometheus.MustRegister(versioncollector.NewCollector(exporterName))

	prometheus.MustRegister(exporter)

	// TODO(@sysadmind): Remove this with multi-target support. We are removing multiple DSN support
	dsn := ""
	if len(dsns) > 0 {
		dsn = dsns[0]
	}

	pe, err := collector.NewPostgresCollector(
		logger,
		excludedDatabases,
		dsn,
		[]string{},
		collector.WithCollectionTimeout(cfg.CollectionTimeout.String()),
		collector.WithCollectorStates(cfg.CollectorStates()),
		collector.WithStatStatementsConfig(
			cfg.StatStatements.IncludeQuery,
			cfg.StatStatements.QueryLength,
			cfg.StatStatements.Limit,
			cfg.StatStatements.ExcludeDatabases,
			cfg.StatStatements.ExcludeUsers,
		))
	if err != nil {
		logger.Warn("Failed to create PostgresCollector", "err", err.Error())
	} else {
		prometheus.MustRegister(pe)
	}

	http.Handle(*metricsPath, promhttp.Handler())

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

	http.HandleFunc("/probe", handleProbe(logger, excludedDatabases))

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		logger.Error("Error running HTTP server", "err", err)
		os.Exit(1)
	}
}

func parseCommaSeparatedList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
