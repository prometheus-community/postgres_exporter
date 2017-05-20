package main

import (
	"fmt"
	"net/http"
	"runtime"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/wrouesnel/postgres_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version is set during build to the git describe version
// (semantic version)-(commitish) form.
var Version = "0.0.1"

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9187").OverrideDefaultFromEnvar("PG_EXPORTER_WEB_LISTEN_ADDRESS").String()
	metricPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").OverrideDefaultFromEnvar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	queriesPath   = kingpin.Flag("extend.query-path", "Path to custom queries to run.").Default("").OverrideDefaultFromEnvar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps  = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
)

func main() {
	kingpin.Version(fmt.Sprintf("postgres_exporter %s (built with %s)\n", Version, runtime.Version()))
	log.AddFlags(kingpin.CommandLine)
	kingpin.Parse()

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
		collector.DumpMaps()
		return
	}

	dsn := collector.GetDataSource()
	if len(dsn) == 0 {
		log.Fatal("couldn't find environment variables describing the datasource to use")
	}

	exporter := collector.NewExporter(dsn, *queriesPath)
	defer exporter.Close()

	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type:text/plain; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                                     // nolint: errcheck
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
