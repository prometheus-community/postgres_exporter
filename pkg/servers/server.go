package servers

import (
	"database/sql"
	"fmt"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/wrouesnel/postgres_exporter/pkg/pgdbconv"
	"github.com/wrouesnel/postgres_exporter/pkg/queries/metricmaps"
	"sync"
	"time"
)

// Server describes a connection to a Postgresql database, and describes the
// metric maps and overrides in-place for it.
type Server struct {
	db     *sql.DB
	labels prometheus.Labels
	master bool

	// Last version used to calculate metric map. If mismatch on scrape,
	// then maps are recalculated.
	lastMapVersion semver.Version
	// Currently active metric map
	metricMap map[string]metricmaps.MetricMapNamespace
	// Currently active query overrides
	queryOverrides map[string]string
	mappingMtx     sync.RWMutex
	// Currently cached metrics
	metricCache map[string]cachedMetrics
	cacheMtx    sync.Mutex
}

// NewServer establishes a new connection using DSN.
func NewServer(dsn string, opts ...ServerOpt) (*Server, error) {
	fingerprint, err := parseFingerprint(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	log.Infof("Established new database connection to %q.", fingerprint)

	s := &Server{
		db:     db,
		master: false,
		labels: prometheus.Labels{
			metricmaps.ServerLabelName: fingerprint,
		},
		metricCache: make(map[string]cachedMetrics),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Close disconnects from Postgres.
func (s *Server) Close() error {
	return s.db.Close()
}

// Ping checks connection availability and possibly invalidates the connection if it fails.
func (s *Server) Ping() error {
	if err := s.db.Ping(); err != nil {
		if cerr := s.Close(); cerr != nil {
			log.Errorf("Error while closing non-pinging DB connection to %q: %v", s, cerr)
		}
		return err
	}
	return nil
}

// String returns server's fingerprint.
func (s *Server) String() string {
	return s.labels[metricmaps.ServerLabelName]
}

// Scrape loads metrics.
func (s *Server) Scrape(ch chan<- prometheus.Metric, disableSettingsMetrics bool) error {
	s.mappingMtx.RLock()
	defer s.mappingMtx.RUnlock()

	var err error

	if !disableSettingsMetrics && s.master {
		if err = querySettings(ch, s); err != nil {
			err = fmt.Errorf("error retrieving settings: %s", err)
		}
	}

	errMap := queryNamespaceMappings(ch, s)
	if len(errMap) > 0 {
		err = fmt.Errorf("queryNamespaceMappings returned %d errors", len(errMap))
	}

	return err
}

// Query within a namespace mapping and emit metrics. Returns fatal errors if
// the scrape fails, and a slice of errors if they were non-fatal.
func (s *Server) queryNamespaceMapping(namespace string, mapping MetricMapNamespace) ([]prometheus.Metric, []error, error) {
	// Check for a query override for this namespace
	query, found := s.queryOverrides[namespace]

	// Was this query disabled (i.e. nothing sensible can be queried on cu
	// version of PostgreSQL?
	if query == "" && found {
		// Return success (no pertinent data)
		return []prometheus.Metric{}, []error{}, nil
	}

	// Don't fail on a bad scrape of one metric
	var rows *sql.Rows
	var err error

	if !found {
		// I've no idea how to avoid this properly at the moment, but this is
		// an admin tool so you're not injecting SQL right?
		rows, err = s.db.Query(fmt.Sprintf("SELECT * FROM %s;", namespace)) // nolint: gas, safesql
	} else {
		rows, err = s.db.Query(query) // nolint: safesql
	}
	if err != nil {
		return []prometheus.Metric{}, []error{}, fmt.Errorf("Error running query on database %q: %s %v", s, namespace, err)
	}
	defer rows.Close() // nolint: errcheck

	var columnNames []string
	columnNames, err = rows.Columns()
	if err != nil {
		return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving column list for: ", namespace, err))
	}

	// Make a lookup map for the column indices
	var columnIdx = make(map[string]int, len(columnNames))
	for i, n := range columnNames {
		columnIdx[n] = i
	}

	var columnData = make([]interface{}, len(columnNames))
	var scanArgs = make([]interface{}, len(columnNames))
	for i := range columnData {
		scanArgs[i] = &columnData[i]
	}

	nonfatalErrors := []error{}

	metrics := make([]prometheus.Metric, 0)

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving rows:", namespace, err))
		}

		// Get the label values for this row.
		labels := make([]string, len(mapping.labels))
		for idx, label := range mapping.labels {
			labels[idx], _ = pgdbconv.DBToString(columnData[columnIdx[label]])
		}

		// Loop over column names, and match to scan data. Unknown columns
		// will be filled with an untyped metric number *if* they can be
		// converted to float64s. NULLs are allowed and treated as NaN.
		for idx, columnName := range columnNames {
			var metric prometheus.Metric
			if metricMapping, ok := mapping.columnMappings[columnName]; ok {
				// Is this a metricy metric?
				if metricMapping.discard {
					continue
				}

				value, ok := pgdbconv.DBToFloat64(columnData[idx])
				if !ok {
					nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName, columnData[idx])))
					continue
				}
				// Generate the metric
				metric = prometheus.MustNewConstMetric(metricMapping.desc, metricMapping.vtype, value, labels...)
			} else {
				// Unknown metric. Report as untyped if scan to float64 works, else note an error too.
				metricLabel := fmt.Sprintf("%s_%s", namespace, columnName)
				desc := prometheus.NewDesc(metricLabel, fmt.Sprintf("Unknown metric from %s", namespace), mapping.labels, s.labels)

				// Its not an error to fail here, since the values are
				// unexpected anyway.
				value, ok := pgdbconv.DBToFloat64(columnData[idx])
				if !ok {
					nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unparseable column type - discarding: ", namespace, columnName, err)))
					continue
				}
				metric = prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, value, labels...)
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nonfatalErrors, nil
}

// Iterate through all the namespace mappings in the exporter and run their
// queries.
func (s *Server) queryNamespaceMappings(ch chan<- prometheus.Metric) map[string]error {
	// Return a map of namespace -> errors
	namespaceErrors := make(map[string]error)

	scrapeStart := time.Now()

	for namespace, mapping := range s.metricMap {
		log.Debugln("Querying namespace: ", namespace)

		if mapping.master && !s.master {
			log.Debugln("Query skipped...")
			continue
		}

		scrapeMetric := false
		// Check if the metric is cached
		s.cacheMtx.Lock()
		cachedMetric, found := s.metricCache[namespace]
		s.cacheMtx.Unlock()
		// If found, check if needs refresh from cache
		if found {
			if scrapeStart.Sub(cachedMetric.lastScrape).Seconds() > float64(mapping.cacheSeconds) {
				scrapeMetric = true
			}
		} else {
			scrapeMetric = true
		}

		var metrics []prometheus.Metric
		var nonFatalErrors []error
		var err error
		if scrapeMetric {
			metrics, nonFatalErrors, err = s.queryNamespaceMapping(namespace, mapping)
		} else {
			metrics = cachedMetric.metrics
		}

		// Serious error - a namespace disappeared
		if err != nil {
			namespaceErrors[namespace] = err
			log.Infoln(err)
		}
		// Non-serious errors - likely version or parsing problems.
		if len(nonFatalErrors) > 0 {
			for _, err := range nonFatalErrors {
				log.Infoln(err.Error())
			}
		}

		// Emit the metrics into the channel
		for _, metric := range metrics {
			ch <- metric
		}

		if scrapeMetric {
			// Only cache if metric is meaningfully cacheable
			if mapping.cacheSeconds > 0 {
				s.cacheMtx.Lock()
				s.metricCache[namespace] = cachedMetrics{
					metrics:    metrics,
					lastScrape: scrapeStart,
				}
				s.cacheMtx.Unlock()
			}
		}
	}

	return namespaceErrors
}


// AddQueries adds queries to the builtinMetricMaps and queryOverrides maps.
// Added queries do not respect version requirements, because it is assumed
// that the user knows what they are doing with their version of postgres.
func (s *Server) AddQueries() {

}

// Query the pg_settings view containing runtime variables
func (s *Server) querySettings(ch chan<- prometheus.Metric) error {
	log.Debugf("Querying pg_setting view on %q", s)

	// pg_settings docs: https://www.postgresql.org/docs/current/static/view-pg-settings.html
	//
	// NOTE: If you add more vartypes here, you must update the supported
	// types in normaliseUnit() below
	query := "SELECT name, setting, COALESCE(unit, ''), short_desc, vartype FROM pg_settings WHERE vartype IN ('bool', 'integer', 'real');"

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("Error running query on database %q: %s %v", s, metricmaps.ExporterNamespaceLabel, err)
	}
	defer rows.Close() // nolint: errcheck

	for rows.Next() {
		setting := &pgSetting{}
		err = rows.Scan(&setting.name, &setting.setting, &setting.unit, &setting.shortDesc, &setting.vartype)
		if err != nil {
			return fmt.Errorf("Error retrieving rows on %q: %s %v", s, metricmaps.ExporterNamespaceLabel, err)
		}

		ch <- setting.metric(s.labels)
	}
	return nil
}