package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	"crypto/sha256"

	"github.com/blang/semver"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
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





// TODO: revisit this with the semver system
func dumpMaps() {
	// TODO: make this function part of the exporter
	for name, cmap := range builtinMetricMaps {
		query, ok := queryOverrides[name]
		if !ok {
			fmt.Println(name)
		} else {
			for _, queryOverride := range query {
				fmt.Println(name, queryOverride.versionRange, queryOverride.query)
			}
		}

		for column, details := range cmap {
			fmt.Printf("  %-40s %v\n", column, details)
		}
		fmt.Println()
	}
}





// Convert the query override file to the version-specific query override file
// for the exporter.
func makeQueryOverrideMap(pgVersion semver.Version, queryOverrides map[string][]OverrideQuery) map[string]string {
	resultMap := make(map[string]string)
	for name, overrideDef := range queryOverrides {
		// Find a matching semver. We make it an error to have overlapping
		// ranges at test-time, so only 1 should ever match.
		matched := false
		for _, queryDef := range overrideDef {
			if queryDef.versionRange(pgVersion) {
				resultMap[name] = queryDef.query
				matched = true
				break
			}
		}
		if !matched {
			log.Warnln("No query matched override for", name, "- disabling metric space.")
			resultMap[name] = ""
		}
	}

	return resultMap
}

// Add queries to the builtinMetricMaps and queryOverrides maps. Added queries do not
// respect version requirements, because it is assumed that the user knows
// what they are doing with their version of postgres.
//
// This function modifies metricMap and queryOverrideMap to contain the new
// queries.
// TODO: test code for all cu.
// TODO: use proper struct type system
// TODO: the YAML this supports is "non-standard" - we should move away from it.
func addQueries(content []byte, pgVersion semver.Version, exporterMap map[string]MetricMapNamespace, queryOverrideMap map[string]string) error {
	var extra map[string]interface{}

	err := yaml.Unmarshal(content, &extra)
	if err != nil {
		return err
	}

	// Stores the loaded map representation
	metricMaps := make(map[string]map[string]ColumnMapping)
	newQueryOverrides := make(map[string]string)

	for metric, specs := range extra {
		log.Debugln("New user metric namespace from YAML:", metric)
		for key, value := range specs.(map[interface{}]interface{}) {
			switch key.(string) {
			case "query":
				query := value.(string)
				newQueryOverrides[metric] = query

			case "metrics":
				for _, c := range value.([]interface{}) {
					column := c.(map[interface{}]interface{})

					for n, a := range column {
						var columnMapping ColumnMapping

						// Fetch the metric map we want to work on.
						metricMap, ok := metricMaps[metric]
						if !ok {
							// Namespace for metric not found - add it.
							metricMap = make(map[string]ColumnMapping)
							metricMaps[metric] = metricMap
						}

						// Get name.
						name := n.(string)

						for attrKey, attrVal := range a.(map[interface{}]interface{}) {
							switch attrKey.(string) {
							case "usage":
								usage, err := stringToColumnUsage(attrVal.(string))
								if err != nil {
									return err
								}
								columnMapping.usage = usage
							case "description":
								columnMapping.description = attrVal.(string)
							}
						}

						// TODO: we should support cu
						columnMapping.mapping = nil
						// Should we support this for users?
						columnMapping.supportedVersions = nil

						metricMap[name] = columnMapping
					}
				}
			}
		}
	}

	// Convert the loaded metric map into exporter representation
	partialExporterMap := makeDescMap(pgVersion, metricMaps)

	// Merge the two maps (which are now quite flatteend)
	for k, v := range partialExporterMap {
		_, found := exporterMap[k]
		if found {
			log.Debugln("Overriding metric", k, "from user YAML file.")
		} else {
			log.Debugln("Adding new metric", k, "from user YAML file.")
		}
		exporterMap[k] = v
	}

	// Merge the query override map
	for k, v := range newQueryOverrides {
		_, found := queryOverrideMap[k]
		if found {
			log.Debugln("Overriding query override", k, "from user YAML file.")
		} else {
			log.Debugln("Adding new query override", k, "from user YAML file.")
		}
		queryOverrideMap[k] = v
	}

	return nil
}









// try to get the DataSource
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > DATA_SOURCE_{USER|FILE}_FILE > DATA_SOURCE_{USER|FILE}
func getDataSource() string {
	var dsn = os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) == 0 {
		var user string
		var pass string

		if len(os.Getenv("DATA_SOURCE_USER_FILE")) != 0 {
			fileContents, err := ioutil.ReadFile(os.Getenv("DATA_SOURCE_USER_FILE"))
			if err != nil {
				panic(err)
			}
			user = strings.TrimSpace(string(fileContents))
		} else {
			user = os.Getenv("DATA_SOURCE_USER")
		}

		if len(os.Getenv("DATA_SOURCE_PASS_FILE")) != 0 {
			fileContents, err := ioutil.ReadFile(os.Getenv("DATA_SOURCE_PASS_FILE"))
			if err != nil {
				panic(err)
			}
			pass = strings.TrimSpace(string(fileContents))
		} else {
			pass = os.Getenv("DATA_SOURCE_PASS")
		}

		uri := os.Getenv("DATA_SOURCE_URI")
		dsn = "postgresql://" + user + ":" + pass + "@" + uri
	}

	return dsn
}

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
		dumpMaps()
		return
	}

	dsn := getDataSource()
	if len(dsn) == 0 {
		log.Fatal("couldn't find environment variables describing the datasource to use")
	}

	exporter := NewExporter(dsn, *queriesPath)
	defer func() {
		if exporter.dbConnection != nil {
			exporter.dbConnection.Close() // nolint: errcheck
		}
	}()

	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type:text/plain; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                                     // nolint: errcheck
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
