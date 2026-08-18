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
	"github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/exporter-toolkit/bootstrap"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	c = newConfigHandler()

	configFile            = kingpin.Flag("config.file", "Postgres exporter configuration file.").Default("postgres_exporter.yml").String()
	disableDefaultMetrics = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	autoDiscoverDatabases = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	queriesPath           = kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps          = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList    = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases      = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases      = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix          = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").Default("pg").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	collectionTimeout     = kingpin.Flag("collection-timeout", "Timeout for collecting the statistics when the database is slow").Default("1m").Envar("PG_EXPORTER_COLLECTION_TIMEOUT").String()
	logger                = promslog.NewNopLogger()
)

// The name of the exporter.
const exporterName = "postgres_exporter"

func newConfigHandler() *config.Handler {
	handler, err := config.NewHandler(prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}
	return handler
}

func main() {
	runner := bootstrap.New(bootstrap.Config{
		App:              kingpin.CommandLine,
		Name:             exporterName,
		Description:      "Prometheus PostgreSQL server Exporter",
		DefaultAddress:   ":9187",
		MetricsPathEnvar: "PG_EXPORTER_WEB_TELEMETRY_PATH",
		LandingConfig: web.LandingConfig{
			Name: "Postgres Exporter",
		},
		MetricsHandlerFactory: newMetricsHandler,
	})

	err := runner.Run()
	if pgExporter != nil {
		pgExporter.CloseServers()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// pgExporter is the legacy exporter, kept package level so that its database
// connections can be closed once the HTTP server stops.
var pgExporter *exporter.Exporter

// newMetricsHandler wires up the exporter once the toolkit has parsed the
// command line, and registers the endpoints that are specific to
// postgres_exporter.
func newMetricsHandler(b *bootstrap.Bootstrap) (http.Handler, error) {
	logger = b.Logger

	if *onlyDumpMaps {
		exporter.DumpMaps()
		os.Exit(0)
	}

	if err := c.ReloadConfig(*configFile, logger); err != nil {
		// This is not fatal, but it means that auth must be provided for every dsn.
		logger.Warn("Error loading config", "err", err)
	}

	dsns, err := exporter.GetDataSources()
	if err != nil {
		return nil, fmt.Errorf("failed reading data sources: %w", err)
	}

	excludedDatabases := strings.Split(*excludeDatabases, ",")
	logger.Info("Excluded databases", "databases", fmt.Sprintf("%v", excludedDatabases))

	if *queriesPath != "" {
		logger.Warn("The extended queries.yaml config is DEPRECATED", "file", *queriesPath)
	}

	if *autoDiscoverDatabases || *excludeDatabases != "" || *includeDatabases != "" {
		logger.Warn("Scraping additional databases via auto discovery is DEPRECATED")
	}

	if *constantLabelsList != "" {
		logger.Warn("Constant labels on all metrics is DEPRECATED")
	}

	opts := []exporter.ExporterOpt{
		exporter.DisableDefaultMetrics(*disableDefaultMetrics),
		exporter.AutoDiscoverDatabases(*autoDiscoverDatabases),
		exporter.WithUserQueriesPath(*queriesPath),
		exporter.WithConstantLabels(*constantLabelsList),
		exporter.ExcludeDatabases(excludedDatabases),
		exporter.IncludeDatabases(*includeDatabases),
		exporter.WithMetricPrefix(*metricPrefix),
	}

	pgExporter = exporter.NewExporter(dsns, logger, opts...)

	prometheus.MustRegister(versioncollector.NewCollector(exporterName))
	prometheus.MustRegister(pgExporter)

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
		collector.WithCollectionTimeout(*collectionTimeout))
	if err != nil {
		logger.Warn("Failed to create PostgresCollector", "err", err.Error())
	} else {
		prometheus.MustRegister(pe)
	}

	if b.DisableExporterMetrics {
		prometheus.Unregister(collectors.NewGoCollector())
		prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	b.HandleFunc("/probe", handleProbe(logger, excludedDatabases))
	// net/http/pprof registers its handlers on the default mux.
	b.Handle("/debug/pprof/", http.DefaultServeMux)

	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		MaxRequestsInFlight: b.MaxRequests,
	}), nil
}
