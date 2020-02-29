package main

import (
	"errors"
	"fmt"
	"github.com/wrouesnel/postgres_exporter/pkg/queries/metricmaps"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/wrouesnel/postgres_exporter/pkg/servers"
)

// Branch is set during build to the git branch.
var Branch string

// BuildDate is set during build to the ISO-8601 date and time.
var BuildDate string

// Revision is set during build to the git commit revision.
var Revision string

// Version is set during build to the git describe version
// (semantic version)-(commitish) form.
var Version = "0.0.1-rev"

// VersionShort is set during build to the semantic version.
var VersionShort = "0.0.1"

var (
	listenAddress          = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9187").Envar("PG_EXPORTER_WEB_LISTEN_ADDRESS").String()
	metricPath             = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics  = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingsMetrics = kingpin.Flag("disable-settings-metrics", "Do not include pg_settings metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").Bool()
	autoDiscoverDatabases  = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically.").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	queriesPath            = kingpin.Flag("extend.query-path", "Path to custom queries to run.").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps           = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList     = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,).").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases       = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
)

// Regex used to get the "short-version" from the postgres version field.
var versionRegex = regexp.MustCompile(`^\w+ ((\d+)(\.\d+)?(\.\d+)?)`)
var lowestSupportedVersion = semver.MustParse("9.1.0")

// Parses the version of postgres into the short version string we can use to
// match behaviors.
func parseVersion(versionString string) (semver.Version, error) {
	submatches := versionRegex.FindStringSubmatch(versionString)
	if len(submatches) > 1 {
		return semver.ParseTolerant(submatches[1])
	}
	return semver.Version{},
		errors.New(fmt.Sprintln("Could not find a postgres version in string:", versionString))
}

// ErrorConnectToServer is a connection to PgSQL server error
type ErrorConnectToServer struct {
	Msg string
}

// Error returns error
func (e *ErrorConnectToServer) Error() string {
	return e.Msg
}

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

		for column, details := range cmap.columnMappings {
			fmt.Printf("  %-40s %v\n", column, details)
		}
		fmt.Println()
	}
}

// Add queries to the builtinMetricMaps and queryOverrides maps. Added queries do not
// respect version requirements, because it is assumed that the user knows
// what they are doing with their version of postgres.
//
// This function modifies metricMap and queryOverrideMap to contain the new
// queries.
// TODO: the YAML this supports is "non-standard" - we should move away from it.
func addQueries(content []byte, pgVersion semver.Version, server *Server) error {
	metricMaps, newQueryOverrides, err := parseUserQueries(content)
	if err != nil {
		return err
	}
	// Convert the loaded metric map into exporter representation
	partialExporterMap := makeDescMap(pgVersion, server.labels, metricMaps)

	// Merge the two maps (which are now quite flatteend)
	for k, v := range partialExporterMap {
		_, found := server.metricMap[k]
		if found {
			log.Debugln("Overriding metric", k, "from user YAML file.")
		} else {
			log.Debugln("Adding new metric", k, "from user YAML file.")
		}
		server.metricMap[k] = v
	}

	// Merge the query override map
	for k, v := range newQueryOverrides {
		_, found := server.queryOverrides[k]
		if found {
			log.Debugln("Overriding query override", k, "from user YAML file.")
		} else {
			log.Debugln("Adding new query override", k, "from user YAML file.")
		}
		server.queryOverrides[k] = v
	}
	return nil
}

func parseConstLabels(s string) prometheus.Labels {
	labels := make(prometheus.Labels)

	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return labels
	}

	parts := strings.Split(s, ",")
	for _, p := range parts {
		keyValue := strings.Split(strings.TrimSpace(p), "=")
		if len(keyValue) != 2 {
			log.Errorf(`Wrong constant labels format %q, should be "key=value"`, p)
			continue
		}
		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])
		if key == "" || value == "" {
			continue
		}
		labels[key] = value
	}

	return labels
}

func newDesc(subsystem, name, help string, labels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, labels,
	)
}

func queryDatabases(server *servers.Server) ([]string, error) {
	rows, err := server.db.Query("SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()") // nolint: safesql
	if err != nil {
		return nil, fmt.Errorf("Error retrieving databases: %v", err)
	}
	defer rows.Close() // nolint: errcheck

	var databaseName string
	result := make([]string, 0)
	for rows.Next() {
		err = rows.Scan(&databaseName)
		if err != nil {
			return nil, errors.New(fmt.Sprintln("Error retrieving rows:", err))
		}
		result = append(result, databaseName)
	}

	return result, nil
}

// getDataSources tries to get a datasource connection ID.
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > DATA_SOURCE_{USER|PASS}_FILE > DATA_SOURCE_{USER|PASS}
func getDataSources() []string {
	var dsn = os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) == 0 {
		var user string
		var pass string
		var uri string

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

		ui := url.UserPassword(user, pass).String()

		if len(os.Getenv("DATA_SOURCE_URI_FILE")) != 0 {
			fileContents, err := ioutil.ReadFile(os.Getenv("DATA_SOURCE_URI_FILE"))
			if err != nil {
				panic(err)
			}
			uri = strings.TrimSpace(string(fileContents))
		} else {
			uri = os.Getenv("DATA_SOURCE_URI")
		}

		dsn = "postgresql://" + ui + "@" + uri

		return []string{dsn}
	}
	return strings.Split(dsn, ",")
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

	dsn := getDataSources()
	if len(dsn) == 0 {
		log.Fatal("couldn't find environment variables describing the datasource to use")
	}

	exporter := NewExporter(dsn,
		DisableDefaultMetrics(*disableDefaultMetrics),
		DisableSettingsMetrics(*disableSettingsMetrics),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithUserQueriesPath(*queriesPath),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(*excludeDatabases),
	)
	defer func() {
		exporter.servers.Close()
	}()

	// Setup build info metric.
	version.Branch = Branch
	version.BuildDate = BuildDate
	version.Revision = Revision
	version.Version = VersionShort
	prometheus.MustRegister(version.NewCollector("postgres_exporter"))

	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                       // nolint: errcheck
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
