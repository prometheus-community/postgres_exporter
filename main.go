package main

import (
	//"bytes"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	//"regexp"
	//"strconv"
	//"strings"
	"math"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

var (
	listenAddress = flag.String(
		"web.listen-address", ":9113",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.",
	)
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "pg"
	// Subsystems.
	exporter = "exporter"
)

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

type ColumnUsage int

const (
	DISCARD      ColumnUsage = iota // Ignore this column
	LABEL        ColumnUsage = iota // Use this column as a label
	COUNTER      ColumnUsage = iota // Use this column as a counter
	GAUGE        ColumnUsage = iota // Use this column as a gauge
	MAPPEDMETRIC ColumnUsage = iota // Use this column with the supplied mapping of text values
)

// Basname of metricset and query to generate the metrics
type NamespaceAndQuery struct {
	namespace string
	query     string
}

// User-friendly representation of a prometheus descriptor map
type ColumnMapping struct {
	usage       ColumnUsage
	description string
	mapping     map[string]float64 // Optional column mapping for MAPPEDMETRIC
}

// Groups metric maps under a shared set of labels
type MetricMapNamespace struct {
	labels         []string             // Label names for this namespace
	columnMappings map[string]MetricMap // Column mappings in this namespace
}

// Stores the prometheus metric description which a given column will be mapped
// to by the collector
type MetricMap struct {
	discard bool                 // Should metric be discarded during mapping?
	vtype   prometheus.ValueType // Prometheus valuetype
	desc    *prometheus.Desc     // Prometheus descriptor
	mapping map[string]float64   // If not nil, maps text values to float64s
}

// Metric descriptors for dynamically created metrics.
var metricMaps = map[NamespaceAndQuery]map[string]ColumnMapping{
	{"pg_database", "select datname,age(datfrozenxid) as xid_age from pg_database"}: map[string]ColumnMapping{
                "datname":      {LABEL, "OID of a database", nil},
                "xid_age":      {COUNTER, "Age of maximum frozen xid", nil},
        },
	{"pg_stat_bgwriter", "select * from pg_stat_bgwriter"}: map[string]ColumnMapping{
		"checkpoints_timed":     {COUNTER, "Number of scheduled checkpoints that have been performed", nil},
		"checkpoints_req":       {COUNTER, "Number of requested checkpoints that have been performed", nil},
		"checkpoint_write_time": {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds", nil},
		"checkpoint_sync_time":  {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds", nil},
		"buffers_checkpoint":    {COUNTER, "Number of buffers written during checkpoints", nil},
		"buffers_clean":         {COUNTER, "Number of buffers written by the background writer", nil},
		"maxwritten_clean":      {COUNTER, "Number of times the background writer stopped a cleaning scan because it had written too many buffers", nil},
		"buffers_backend":       {COUNTER, "Number of buffers written directly by a backend", nil},
		"buffers_backend_fsync": {COUNTER, "Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)", nil},
		"buffers_alloc":         {COUNTER, "Number of buffers allocated", nil},
		"stats_reset":           {COUNTER, "Time at which these statistics were last reset", nil},
	},
	{"pg_stat_database", "select * from pg_stat_database"}: map[string]ColumnMapping{
		"datid":          {LABEL, "OID of a database", nil},
		"datname":        {LABEL, "Name of this database", nil},
		"numbackends":    {GAUGE, "Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.", nil},
		"xact_commit":    {COUNTER, "Number of transactions in this database that have been committed", nil},
		"xact_rollback":  {COUNTER, "Number of transactions in this database that have been rolled back", nil},
		"blks_read":      {COUNTER, "Number of disk blocks read in this database", nil},
		"blks_hit":       {COUNTER, "Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)", nil},
		"tup_returned":   {COUNTER, "Number of rows returned by queries in this database", nil},
		"tup_fetched":    {COUNTER, "Number of rows fetched by queries in this database", nil},
		"tup_inserted":   {COUNTER, "Number of rows inserted by queries in this database", nil},
		"tup_updated":    {COUNTER, "Number of rows updated by queries in this database", nil},
		"tup_deleted":    {COUNTER, "Number of rows deleted by queries in this database", nil},
		"conflicts":      {COUNTER, "Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)", nil},
		"temp_files":     {COUNTER, "Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.", nil},
		"temp_bytes":     {COUNTER, "Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.", nil},
		"deadlocks":      {COUNTER, "Number of deadlocks detected in this database", nil},
		"blk_read_time":  {COUNTER, "Time spent reading data file blocks by backends in this database, in milliseconds", nil},
		"blk_write_time": {COUNTER, "Time spent writing data file blocks by backends in this database, in milliseconds", nil},
		"stats_reset":    {COUNTER, "Time at which these statistics were last reset", nil},
	},
	{"pg_stat_database_conflicts", "select * from pg_stat_database_conflicts"}: map[string]ColumnMapping{
		"datid":            {LABEL, "OID of a database", nil},
		"datname":          {LABEL, "Name of this database", nil},
		"confl_tablespace": {COUNTER, "Number of queries in this database that have been canceled due to dropped tablespaces", nil},
		"confl_lock":       {COUNTER, "Number of queries in this database that have been canceled due to lock timeouts", nil},
		"confl_snapshot":   {COUNTER, "Number of queries in this database that have been canceled due to old snapshots", nil},
		"confl_bufferpin":  {COUNTER, "Number of queries in this database that have been canceled due to pinned buffers", nil},
		"confl_deadlock":   {COUNTER, "Number of queries in this database that have been canceled due to deadlocks", nil},
	},
	{"pg_stat_replication", "select *, pg_current_xlog_location(), pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float from pg_stat_replication"}: map[string]ColumnMapping{
		"pid":              {DISCARD, "Process ID of a WAL sender process", nil},
		"usesysid":         {DISCARD, "OID of the user logged into this WAL sender process", nil},
		"usename":          {DISCARD, "Name of the user logged into this WAL sender process", nil},
		"application_name": {DISCARD, "Name of the application that is connected to this WAL sender", nil},
		"client_addr":      {LABEL, "IP address of the client connected to this WAL sender. If this field is null, it indicates that the client is connected via a Unix socket on the server machine.", nil},
		"client_hostname":  {DISCARD, "Host name of the connected client, as reported by a reverse DNS lookup of client_addr. This field will only be non-null for IP connections, and only when log_hostname is enabled.", nil},
		"client_port":      {DISCARD, "TCP port number that the client is using for communication with this WAL sender, or -1 if a Unix socket is used", nil},
		"backend_start": {DISCARD, "with time zone	Time when this process was started, i.e., when the client connected to this WAL sender", nil},
		"backend_xmin":             {DISCARD, "The current backend's xmin horizon.", nil},
		"state":                    {LABEL, "Current WAL sender state", nil},
		"sent_location":            {DISCARD, "Last transaction log position sent on this connection", nil},
		"write_location":           {DISCARD, "Last transaction log position written to disk by this standby server", nil},
		"flush_location":           {DISCARD, "Last transaction log position flushed to disk by this standby server", nil},
		"replay_location":          {DISCARD, "Last transaction log position replayed into the database on this standby server", nil},
		"sync_priority":            {DISCARD, "Priority of this standby server for being chosen as the synchronous standby", nil},
		"sync_state":               {DISCARD, "Synchronous state of this standby server", nil},
		"pg_current_xlog_location": {DISCARD, "pg_current_xlog_location", nil},
		"pg_xlog_location_diff":    {GAUGE, "Lag in bytes between master and slave", nil},
	},
	{"pg_stat_activity", "SELECT state, count(*) as count, max(EXTRACT(EPOCH FROM now() - xact_start))::float as max_tx_duration, (SELECT setting FROM pg_settings WHERE name = 'max_connections')::float as max_connections FROM pg_stat_activity GROUP BY state"}: map[string]ColumnMapping{
		"state":           {LABEL, "connection state", nil},
		"count":           {GAUGE, "number of connections in this state", nil},
		"max_tx_duration": {GAUGE, "max duration in seconds any active transaction has been running", nil},
		"max_connections": {GAUGE, "maximum number of connections that the db is configured with", nil},
	},
}

// Turn the MetricMap column mapping into a prometheus descriptor mapping.
func makeDescMap(metricMaps map[NamespaceAndQuery]map[string]ColumnMapping) map[NamespaceAndQuery]MetricMapNamespace {
	var metricMap = make(map[NamespaceAndQuery]MetricMapNamespace)

	for namespaceAndQuery, mappings := range metricMaps {
		namespace := namespaceAndQuery.namespace
		thisMap := make(map[string]MetricMap)

		// Get the constant labels
		var constLabels []string
		for columnName, columnMapping := range mappings {
			if columnMapping.usage == LABEL {
				constLabels = append(constLabels, columnName)
			}
		}

		for columnName, columnMapping := range mappings {
			switch columnMapping.usage {
			case DISCARD, LABEL:
				thisMap[columnName] = MetricMap{
					discard: true,
				}
			case COUNTER:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.CounterValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
				}
			case GAUGE:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMap{
					vtype:   prometheus.GaugeValue,
					desc:    prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					mapping: columnMapping.mapping,
				}
			}
		}

		metricMap[namespaceAndQuery] = MetricMapNamespace{constLabels, thisMap}
	}

	return metricMap
}

// Convert database.sql types to float64s for Prometheus consumption. Null types are mapped to NaN. string and []byte
// types are mapped as NaN and !ok
func dbToFloat64(t interface{}) (float64, bool) {
	switch v := t.(type) {
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case time.Time:
		return float64(v.Unix()), true
	case nil:
		return math.NaN(), true
	default:
		return math.NaN(), false
	}
}

// Convert database.sql to string for Prometheus labels. Null types are mapped to empty strings.
func dbToString(t interface{}) (string, bool) {
	switch v := t.(type) {
	case int64:
		return fmt.Sprintf("%v", v), true
	case float64:
		return fmt.Sprintf("%v", v), true
	case time.Time:
		return fmt.Sprintf("%v", v.Unix()), true
	case nil:
		return "", true
	case []byte:
		// Try and convert to string
		return string(v), true
	case string:
		return v, true
	default:
		return "", false
	}
}

// Exporter collects MySQL metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn             string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	metricMap       map[NamespaceAndQuery]MetricMapNamespace
}

// NewExporter returns a new MySQL exporter for the provided DSN.
func NewExporter(dsn string) *Exporter {
	return &Exporter{
		dsn: dsn,
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
		metricMap: makeDescMap(metricMaps),
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
}

func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, nil,
	)
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
	}(time.Now())

	e.error.Set(0)
	e.totalScrapes.Inc()

	db, err := sql.Open("postgres", e.dsn)
	if err != nil {
		log.Println("Error opening connection to database:", err)
		e.error.Set(1)
		return
	}
	defer db.Close()

	for namespaceAndQuery, mapping := range e.metricMap {
		namespace := namespaceAndQuery.namespace
		log.Debugln("Querying namespace: ", namespace, " query: ", namespaceAndQuery.query)

		func() { // Don't fail on a bad scrape of one metric

			rows, err := db.Query(namespaceAndQuery.query)

			if err != nil {
				log.Println("Error running query on database: ", namespace, err)
				e.error.Set(1)
				return
			}
			defer rows.Close()

			var columnNames []string
			columnNames, err = rows.Columns()
			if err != nil {
				log.Println("Error retrieving column list for: ", namespace, err)
				e.error.Set(1)
				return
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

			for rows.Next() {
				err = rows.Scan(scanArgs...)
				if err != nil {
					log.Println("Error retrieving rows:", namespace, err)
					e.error.Set(1)
					return
				}

				// Get the label values for this row
				var labels = make([]string, len(mapping.labels))
				for idx, columnName := range mapping.labels {

					labels[idx], _ = dbToString(columnData[columnIdx[columnName]])
				}

				// Loop over column names, and match to scan data. Unknown columns
				// will be filled with an untyped metric number *if* they can be
				// converted to float64s. NULLs are allowed and treated as NaN.
				for idx, columnName := range columnNames {
					if metricMapping, ok := mapping.columnMappings[columnName]; ok {
						// Is this a metricy metric?
						if metricMapping.discard {
							continue
						}

						value, ok := dbToFloat64(columnData[idx])
						if !ok {
							e.error.Set(1)
							log.Errorln("Unexpected error parsing column: ", namespace, columnName, columnData[idx])
							continue
						}

						// Generate the metric
						ch <- prometheus.MustNewConstMetric(metricMapping.desc, metricMapping.vtype, value, labels...)
					} else {
						// Unknown metric. Report as untyped if scan to float64 works, else note an error too.
						desc := prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), fmt.Sprintf("Unknown metric from %s", namespace), nil, nil)

						// Its not an error to fail here, since the values are
						// unexpected anyway.
						value, ok := dbToFloat64(columnData[idx])
						if !ok {
							log.Warnln("Unparseable column type - discarding: ", namespace, columnName, err)
							continue
						}

						ch <- prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, value, labels...)
					}
				}

			}
		}()
	}
}

func main() {
	flag.Parse()

	dsn := os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) == 0 {
		log.Fatal("couldn't find environment variable DATA_SOURCE_NAME")
	}

	exporter := NewExporter(dsn)
	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
