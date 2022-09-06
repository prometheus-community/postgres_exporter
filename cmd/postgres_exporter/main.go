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

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9187").Envar("PG_EXPORTER_WEB_LISTEN_ADDRESS").String()
	webConfig     = webflag.AddFlags(kingpin.CommandLine)
	webConfigFile = kingpin.Flag(
		"web.config",
		"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
	).Default("").String()
	metricPath             = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics  = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingsMetrics = kingpin.Flag("disable-settings-metrics", "Do not include pg_settings metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").Bool()
	autoDiscoverDatabases  = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically.").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	onlyDumpMaps           = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList     = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,).").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases       = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases       = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix           = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"pg\") prefixes for each of the metrics").Default("pg").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	logger                 = log.NewNopLogger()
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
	kingpin.Parse()
	logger = promlog.New(promlogConfig)

	// landingPage contains the HTML served at '/'.
	// TODO: Make this nicer and more informative.
	var landingPage = []byte(`<html>
	<head><title>Postgres exporter</title></head>
	<body>
	<h1>Postgres exporter</h1>
	<p><a href='` + *metricPath + `'>Metrics</a></p>
	</body>
	</html>
	`)

	if *onlyDumpMaps {
		dumpMaps()
		return
	}

	dsn, err := getDataSources()
	if err != nil {
		level.Error(logger).Log("msg", "Failed reading data sources", "err", err.Error())
		os.Exit(1)
	}

	if len(dsn) == 0 {
		level.Error(logger).Log("msg", "Couldn't find environment variables describing the datasource to use")
		os.Exit(1)
	}

	opts := []ExporterOpt{
		DisableDefaultMetrics(*disableDefaultMetrics),
		DisableSettingsMetrics(*disableSettingsMetrics),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(*excludeDatabases),
		IncludeDatabases(*includeDatabases),
	}

	exporter := NewExporter(dsn, opts...)
	defer func() {
		exporter.servers.Close()
	}()

	versionCollector := version.NewCollector(exporterName)
	prometheus.MustRegister(versionCollector)

	prometheus.MustRegister(exporter)

	cleanup, hr, mr, lr := initializePerconaExporters(dsn, opts)
	defer cleanup()

	pe, err := collector.NewPostgresCollector(
		logger,
		dsn,
		[]string{},
	)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create PostgresCollector", "err", err.Error())
		os.Exit(1)
	}
	prometheus.MustRegister(pe)

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

	http.Handle(*metricPath, promHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                       // nolint: errcheck
	})

	var webCfg string
	if *webConfigFile != "" {
		webCfg = *webConfigFile
	}
	if *webConfig != "" {
		webCfg = *webConfig
	}

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	srv := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(srv, webCfg, logger); err != nil {
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
