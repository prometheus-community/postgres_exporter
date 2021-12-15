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

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	. "github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter"
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

func main() {
	kingpin.Version(version.Print(ExporterName))
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
		DumpMaps()
		return
	}

	sess, err := NewAWSSession(*iamRoleArn)
	if err != nil {
		panic(fmt.Sprintf("error to create the aws session: %v", err))
	}

	rdsClient := rds.New(sess)
	cloudWatchClient := cloudwatch.New(sess)
	rdsMetrics := AwsUtils{
		RdsClient:        rdsClient,
		CloudwatchClient: cloudWatchClient,
	}
	server := Server{
		Dsn:         os.Getenv("DATA_SOURCE_NAME"),
		NsMap:       &NamespaceMappings{},
		SettMetrics: &SettingsMetrics{},
	}

	opts := []ExporterOpt{
		TenantID(*tenantID),
		RdsMetrics(&rdsMetrics),
		ServerInstance(&server),
	}

	exporter := NewExporter(opts...)
	prometheus.MustRegister(version.NewCollector(ExporterName))

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
