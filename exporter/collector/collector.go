package collector

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
	"github.com/prometheus/common/log"

	"github.com/wrouesnel/postgres_exporter/exporter/metricmaps"
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "pg"
	// Subsystems.
	exporter = "exporter"
	// Metric label used for static string data thats handy to send to Prometheus
	// e.g. version
	staticLabelName = "static"
)

// lowestSupportedVersion is the lowest version of Postgres this exporter will support.
var lowestSupportedVersion = semver.MustParse("9.1.0")

// Exporter collects Postgres metrics. It implements prometheus.Collector.
type Exporter struct {
	// Holds a reference to the build in column mappings. Currently this is for testing purposes
	// only, since it just points to the global.
	builtinMetricMaps map[string]map[string]metricmaps.ColumnMapping

	dsn              string
	userQueriesPath  string
	duration         prometheus.Gauge
	error            prometheus.Gauge
	psqlUp           prometheus.Gauge
	userQueriesError *prometheus.GaugeVec
	totalScrapes     prometheus.Counter

	// dbDsn is the connection string used to establish the dbConnection
	dbDsn string
	// dbConnection is used to allow re-using the DB connection between scrapes
	dbConnection *sql.DB

	// Last version used to calculate metric map. If mismatch on scrape,
	// then maps are recalculated.
	lastMapVersion semver.Version
	// Currently active metric map
	metricMap map[string]metricmaps.MetricMapNamespace
	// Currently active query overrides
	queryOverrides map[string]string
	mappingMtx     sync.RWMutex
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
func NewExporter(dsn string, userQueriesPath string) *Exporter {
	return &Exporter{
		builtinMetricMaps: builtinMetricMaps,
		dsn:               dsn,
		userQueriesPath:   userQueriesPath,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from PostgresSQL.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times PostgresSQL was scraped for metrics.",
		}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from PostgreSQL resulted in an error (1 for error, 0 for success).",
		}),
		psqlUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the last scrape of metrics from PostgreSQL was able to connect to the server (1 for yes, 0 for no).",
		}),
		userQueriesError: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "user_queries_load_error",
			Help:      "Whether the user queries file was loaded and parsed successfully (1 for error, 0 for success).",
		}, []string{"filename", "hashsum"}),
		metricMap:      nil,
		queryOverrides: nil,
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// We cannot know in advance what metrics the exporter will generate
	// from Postgres. So we use the poor man's describe method: Run a collect
	// and send the descriptors of all the collected metrics. The problem
	// here is that we need to connect to the Postgres DB. If it is currently
	// unavailable, the descriptors will be incomplete. Since this is a
	// stand-alone exporter and not used as a library within other code
	// implementing additional metrics, the worst that can happen is that we
	// don't detect inconsistent metrics created by this exporter
	// itself. Also, a change in the monitored Postgres instance may change the
	// exported metrics during the runtime of the exporter.

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	ch <- e.psqlUp
	e.userQueriesError.Collect(ch)
}

// Check and update the exporters query maps if the version has changed.
func (e *Exporter) checkMapVersions(ch chan<- prometheus.Metric, db *sql.DB) error {
	log.Debugln("Querying Postgres Version")
	versionRow := db.QueryRow("SELECT version();")
	var versionString string
	err := versionRow.Scan(&versionString)
	if err != nil {
		return fmt.Errorf("Error scanning version string: %v", err)
	}
	semanticVersion, err := parseVersion(versionString)
	if err != nil {
		return fmt.Errorf("Error parsing version string: %v", err)
	}
	if semanticVersion.LT(lowestSupportedVersion) {
		log.Warnln("PostgreSQL version is lower then our lowest supported version! Got", semanticVersion.String(), "minimum supported is", lowestSupportedVersion.String())
	}

	// Check if semantic version changed and recalculate maps if needed.
	if semanticVersion.NE(e.lastMapVersion) || e.metricMap == nil {
		log.Infoln("Semantic Version Changed:", e.lastMapVersion.String(), "->", semanticVersion.String())
		e.mappingMtx.Lock()

		e.metricMap = makeDescMap(semanticVersion, e.builtinMetricMaps)
		e.queryOverrides = makeQueryOverrideMap(semanticVersion, queryOverrides)
		e.lastMapVersion = semanticVersion

		if e.userQueriesPath != "" {
			// Clear the metric while a reload is happening
			e.userQueriesError.Reset()

			// Calculate the hashsum of the useQueries
			userQueriesData, err := ioutil.ReadFile(e.userQueriesPath)
			if err != nil {
				log.Errorln("Failed to reload user queries:", e.userQueriesPath, err)
				e.userQueriesError.WithLabelValues(e.userQueriesPath, "").Set(1)
			} else {
				hashsumStr := fmt.Sprintf("%x", sha256.Sum256(userQueriesData))

				if err := addQueries(userQueriesData, semanticVersion, e.metricMap, e.queryOverrides); err != nil {
					log.Errorln("Failed to reload user queries:", e.userQueriesPath, err)
					e.userQueriesError.WithLabelValues(e.userQueriesPath, hashsumStr).Set(1)
				} else {
					// Mark user queries as successfully loaded
					e.userQueriesError.WithLabelValues(e.userQueriesPath, hashsumStr).Set(0)
				}
			}
		}

		e.mappingMtx.Unlock()
	}

	// Output the version as a special metric
	versionDesc := prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, staticLabelName),
		"Version string as reported by postgres", []string{"version", "short_version"}, nil)

	ch <- prometheus.MustNewConstMetric(versionDesc,
		prometheus.UntypedValue, 1, versionString, semanticVersion.String())
	return nil
}

func (e *Exporter) getDB(conn string) (*sql.DB, error) {
	// Has dsn changed?
	if (e.dbConnection != nil) && (e.dsn != e.dbDsn) {
		err := e.dbConnection.Close()
		log.Warnln("Error while closing obsolete DB connection:", err)
		e.dbConnection = nil
		e.dbDsn = ""
	}

	if e.dbConnection == nil {
		d, err := sql.Open("postgres", conn)
		if err != nil {
			return nil, err
		}

		d.SetMaxOpenConns(1)
		d.SetMaxIdleConns(1)
		e.dbConnection = d
		e.dbDsn = e.dsn
		log.Infoln("Established new database connection.")
	}

	// Always send a ping and possibly invalidate the connection if it fails
	if err := e.dbConnection.Ping(); err != nil {
		cerr := e.dbConnection.Close()
		log.Infoln("Error while closing non-pinging DB connection:", cerr)
		e.dbConnection = nil
		e.psqlUp.Set(0)
		return nil, err
	}

	return e.dbConnection, nil
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
	}(time.Now())

	e.error.Set(0)
	e.totalScrapes.Inc()

	db, err := e.getDB(e.dsn)
	if err != nil {
		loggableDsn := "could not parse DATA_SOURCE_NAME"
		// If the DSN is parseable, log it with a blanked out password
		pDsn, pErr := url.Parse(e.dsn)
		if pErr == nil {
			// Blank user info if not nil
			if pDsn.User != nil {
				pDsn.User = url.UserPassword(pDsn.User.Username(), "PASSWORD_REMOVED")
			}
			loggableDsn = pDsn.String()
		}
		log.Infof("Error opening connection to database (%s): %s", loggableDsn, err)
		e.psqlUp.Set(0)
		e.error.Set(1)
		return
	}

	// Didn't fail, can mark connection as up for this scrape.
	e.psqlUp.Set(1)

	// Check if map versions need to be updated
	if err := e.checkMapVersions(ch, db); err != nil {
		log.Warnln("Proceeding with outdated query maps, as the Postgres version could not be determined:", err)
		e.error.Set(1)
	}

	// Lock the exporter maps
	e.mappingMtx.RLock()
	defer e.mappingMtx.RUnlock()
	if err := querySettings(ch, db); err != nil {
		log.Infof("Error retrieving settings: %s", err)
		e.error.Set(1)
	}

	errMap := queryNamespaceMappings(ch, db, e.metricMap, e.queryOverrides)
	if len(errMap) > 0 {
		e.error.Set(1)
	}
}

