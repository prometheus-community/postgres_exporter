// Copyright 2021 The Prometheus Authors
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
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	c = config.Handler{
		Config: &config.Config{},
	}

	configFile    = kingpin.Flag("config.file", "Postgres exporter configuration file.").Default("postgres_exporter.yml").String()
	webConfig     = kingpinflag.AddFlags(kingpin.CommandLine, ":9187")
	webConfigFile = kingpin.Flag(
		"web.config",
		"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
	).Default("").String() // added for compatibility reasons to not break it in PMM 2.
	metricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics  = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingsMetrics = kingpin.Flag("disable-settings-metrics", "Do not include pg_settings metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").Bool()
	autoDiscoverDatabases  = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	//queriesPath            = kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps       = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases   = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases   = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix       = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").Default("pg").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	logger             = log.NewNopLogger()
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "pg"
	// Subsystems.
	exporter = "exporter"
	// The name of the exporter.
	exporterName = "postgres_exporter"
	// Metric label used for static string data thats handy to send to Prometheus
	// e.g. version
	staticLabelName = "static"
	// Metric label used for server identification.
	serverLabelName = "server"
)

func main() {
	kingpin.Version(version.Print(exporterName))
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	webConfig.WebConfigFile = webConfigFile
	kingpin.Parse()
	logger = promlog.New(promlogConfig)

	if *onlyDumpMaps {
		dumpMaps()
		return
	}

	if err := c.ReloadConfig(*configFile, logger); err != nil {
		// This is not fatal, but it means that auth must be provided for every dsn.
		level.Warn(logger).Log("msg", "Error loading config", "err", err)
	}

	dsns, err := getDataSources()
	if err != nil {
		level.Error(logger).Log("msg", "Failed reading data sources", "err", err.Error())
		os.Exit(1)
	}

	excludedDatabases := strings.Split(*excludeDatabases, ",")
	logger.Log("msg", "Excluded databases", "databases", fmt.Sprintf("%v", excludedDatabases))

	//if *queriesPath != "" {
	//	level.Warn(logger).Log("msg", "The extended queries.yaml config is DEPRECATED", "file", *queriesPath)
	//}

	if *autoDiscoverDatabases || *excludeDatabases != "" || *includeDatabases != "" {
		level.Warn(logger).Log("msg", "Scraping additional databases via auto discovery is DEPRECATED")
	}

	if *constantLabelsList != "" {
		level.Warn(logger).Log("msg", "Constant labels on all metrics is DEPRECATED")
	}

	servers := NewServers(ServerWithLabels(parseConstLabels(*constantLabelsList)))

	opts := []ExporterOpt{
		CollectorName("exporter"),
		DisableDefaultMetrics(*disableDefaultMetrics),
		DisableSettingsMetrics(*disableSettingsMetrics),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithConstantLabels(*constantLabelsList),
		WithServers(servers),
		ExcludeDatabases(excludedDatabases),
		IncludeDatabases(*includeDatabases),
	}

	exporter := NewExporter(dsns, opts...)
	defer func() {
		exporter.servers.Close()
	}()

	versionCollector := version.NewCollector(exporterName)
	prometheus.MustRegister(versionCollector)

	prometheus.MustRegister(exporter)

	// TODO(@sysadmind): Remove this with multi-target support. We are removing multiple DSN support
	dsn := ""
	if len(dsns) > 0 {
		dsn = dsns[0]
	}

	cleanup, hr, mr, lr := initializePerconaExporters(dsns, servers)
	defer cleanup()

	pe, err := collector.NewPostgresCollector(
		logger,
		excludedDatabases,
		dsn,
		[]string{},
	)
	if err != nil {
		level.Warn(logger).Log("msg", "Failed to create PostgresCollector", "err", err.Error())
	} else {
		prometheus.MustRegister(pe)
	}

	psCollector := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
	goCollector := collectors.NewGoCollector()

	promHandler := newHandler(map[string]prometheus.Collector{
		"exporter":         exporter,
		"custom_query.hr":  hr,
		"custom_query.mr":  mr,
		"custom_query.lr":  lr,
		"standard.process": psCollector,
		"standard.go":      goCollector,
		"version":          versionCollector,
		"postgres":         pe,
	})

	http.Handle(*metricsPath, promHandler)

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
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	http.HandleFunc("/probe", handleProbe(logger, excludedDatabases))

	level.Info(logger).Log("msg", "Listening on address", "address", *webConfig.WebListenAddresses)
	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}

// handler wraps an unfiltered http.Handler but uses a filtered handler,
// created on the fly, if filtering is requested. Create instances with
// newHandler. It used for collectors filtering.
type handler struct {
	unfilteredHandler http.Handler
	collectors        map[string]prometheus.Collector
}

func newHandler(collectors map[string]prometheus.Collector) *handler {
	h := &handler{collectors: collectors}

	innerHandler, err := h.innerHandler()
	if err != nil {
		level.Error(logger).Log("msg", "Couldn't create metrics handler", "error", err)
		os.Exit(1)
	}

	h.unfilteredHandler = innerHandler
	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	level.Debug(logger).Log("msg", "Collect query", "filters", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}

	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		level.Warn(logger).Log("msg", "Couldn't create filtered metrics handler", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err))) // nolint: errcheck
		return
	}

	filteredHandler.ServeHTTP(w, r)
}

func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	registry := prometheus.NewRegistry()

	// register all collectors by default.
	if len(filters) == 0 {
		for name, c := range h.collectors {
			if err := registry.Register(c); err != nil {
				return nil, err
			}
			level.Debug(logger).Log("msg", "Collector was registered", "collector", name)
		}
	}

	// register only filtered collectors.
	for _, name := range filters {
		if c, ok := h.collectors[name]; ok {
			if err := registry.Register(c); err != nil {
				return nil, err
			}
			level.Debug(logger).Log("msg", "Collector was registered", "collector", name)
		}
	}

	handler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			//ErrorLog:       log.NewNopLogger() .NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	return handler, nil
}
