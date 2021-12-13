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
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
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
	onlyDumpMaps  = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	iamRoleArn    = kingpin.Flag("iam-role-arn", "AWS IAM role to assume and query the aurora serverless status").Default("").Envar("PG_IAM_ROLE_ARN").String()
	tenantID      = kingpin.Flag("tenant-id", "Tenant ID").Default("").Envar("PG_TENANT_ID").String()
	logger        = log.NewNopLogger()
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
	<p><a href='/metrics'>Metrics</a></p>
	</body>
	</html>
	`)

	if *onlyDumpMaps {
		dumpMaps()
		return
	}

	opts := []ExporterOpt{
		IamRoleArn(*iamRoleArn),
		TenantID(*tenantID),
	}

	exporter := NewExporter(os.Getenv("DATA_SOURCE_NAME"), opts...)
	prometheus.MustRegister(version.NewCollector(exporterName))

	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                       // nolint: errcheck
	})

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	srv := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}
