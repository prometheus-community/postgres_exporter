package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var Version string = "0.0.1"

var (
	listenAddress = flag.String(
		"web.listen-address", ":9113",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.",
	)
	queriesPath = flag.String(
		"extend.query-path", "",
		"Path to custom queries to run.",
	)
	onlyDumpMaps = flag.Bool(
		"dumpmaps", false,
		"Do not run, simply dump the maps.",
	)
	pathpassword = flag.String(
		"pathpassword", "",
		"Specifie path to file contain password for postgres Server",
	)
	password = flag.String(
		"password", "none",
		"Specifie password for postgres Server",
	)
	port = flag.String(
		"port", "5432",
		"Specifie port for postgres Server",
	)
	host = flag.String(
		"port", "localhost",
		"Specifie host for postgres Server",
	)
	params = flag.String(
		"params", "?sslmode=disable",
		"Specifie params for dsn (default: ?sslmode=disable)",
	)
	user = flag.String(
		"user", "postgres",
		"Specifie the user postgres",
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
	DURATION     ColumnUsage = iota // This column should be interpreted as a text duration (and converted to milliseconds)
)

// Which metric mapping should be acquired using "SHOW" queries
const SHOW_METRIC = "pg_runtime_variables"

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
	discard    bool                              // Should metric be discarded during mapping?
	vtype      prometheus.ValueType              // Prometheus valuetype
	desc       *prometheus.Desc                  // Prometheus descriptor
	conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
}

// Metric descriptors for dynamically created metrics.
var variableMaps = map[string]map[string]ColumnMapping{
	"pg_runtime_variable": map[string]ColumnMapping{
		"max_connections":                {GAUGE, "Sets the maximum number of concurrent connections.", nil},
		"max_files_per_process":          {GAUGE, "Sets the maximum number of simultaneously open files for each server process.", nil},
		"max_function_args":              {GAUGE, "Shows the maximum number of function arguments.", nil},
		"max_identifier_length":          {GAUGE, "Shows the maximum identifier length.", nil},
		"max_index_keys":                 {GAUGE, "Shows the maximum number of index keys.", nil},
		"max_locks_per_transaction":      {GAUGE, "Sets the maximum number of locks per transaction.", nil},
		"max_pred_locks_per_transaction": {GAUGE, "Sets the maximum number of predicate locks per transaction.", nil},
		"max_prepared_transactions":      {GAUGE, "Sets the maximum number of simultaneously prepared transactions.", nil},
		//"max_stack_depth" : { GAUGE, "Sets the maximum number of concurrent connections.", nil }, // No dehumanize support yet
		"max_standby_archive_delay":   {DURATION, "Sets the maximum delay before canceling queries when a hot standby server is processing archived WAL data.", nil},
		"max_standby_streaming_delay": {DURATION, "Sets the maximum delay before canceling queries when a hot standby server is processing streamed WAL data.", nil},
		"max_wal_senders":             {GAUGE, "Sets the maximum number of simultaneously running WAL sender processes.", nil},
	},
}

func dumpMaps() {
	for name, cmap := range metricMaps {
		query, ok := queryOverrides[name]
		if ok {
			fmt.Printf("%s: %s\n", name, query)
		} else {
			fmt.Println(name)
		}
		for column, details := range cmap {
			fmt.Printf("  %-40s %v\n", column, details)
		}
		fmt.Println()
	}
}

var metricMaps = map[string]map[string]ColumnMapping{
	"pg_stat_bgwriter": map[string]ColumnMapping{
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
	"pg_stat_database": map[string]ColumnMapping{
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
	"pg_stat_database_conflicts": map[string]ColumnMapping{
		"datid":            {LABEL, "OID of a database", nil},
		"datname":          {LABEL, "Name of this database", nil},
		"confl_tablespace": {COUNTER, "Number of queries in this database that have been canceled due to dropped tablespaces", nil},
		"confl_lock":       {COUNTER, "Number of queries in this database that have been canceled due to lock timeouts", nil},
		"confl_snapshot":   {COUNTER, "Number of queries in this database that have been canceled due to old snapshots", nil},
		"confl_bufferpin":  {COUNTER, "Number of queries in this database that have been canceled due to pinned buffers", nil},
		"confl_deadlock":   {COUNTER, "Number of queries in this database that have been canceled due to deadlocks", nil},
	},
	"pg_locks": map[string]ColumnMapping{
		"datname": {LABEL, "Name of this database", nil},
		"mode":    {LABEL, "Type of Lock", nil},
		"count":   {GAUGE, "Number of locks", nil},
	},
	"pg_stat_replication": map[string]ColumnMapping{
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
		"slot_name":                {LABEL, "A unique, cluster-wide identifier for the replication slot", nil},
		"plugin":                   {DISCARD, "The base name of the shared object containing the output plugin this logical slot is using, or null for physical slots", nil},
		"slot_type":                {DISCARD, "The slot type - physical or logical", nil},
		"datoid":                   {DISCARD, "The OID of the database this slot is associated with, or null. Only logical slots have an associated database", nil},
		"database":                 {DISCARD, "The name of the database this slot is associated with, or null. Only logical slots have an associated database", nil},
		"active":                   {DISCARD, "True if this slot is currently actively being used", nil},
		"active_pid":               {DISCARD, "Process ID of a WAL sender process", nil},
		"xmin":                     {DISCARD, "The oldest transaction that this slot needs the database to retain. VACUUM cannot remove tuples deleted by any later transaction", nil},
		"catalog_xmin":             {DISCARD, "The oldest transaction affecting the system catalogs that this slot needs the database to retain. VACUUM cannot remove catalog tuples deleted by any later transaction", nil},
		"restart_lsn":              {DISCARD, "The address (LSN) of oldest WAL which still might be required by the consumer of this slot and thus won't be automatically removed during checkpoints", nil},
		"pg_current_xlog_location": {DISCARD, "pg_current_xlog_location", nil},
		"pg_xlog_location_diff":    {GAUGE, "Lag in bytes between master and slave", nil},
	},
	"pg_stat_activity": map[string]ColumnMapping{
		"datname":         {LABEL, "Name of this database", nil},
		"state":           {LABEL, "connection state", nil},
		"count":           {GAUGE, "number of connections in this state", nil},
		"max_tx_duration": {GAUGE, "max duration in seconds any active transaction has been running", nil},
	},
}

// Overriding queries for namespaces above.
var queryOverrides = map[string]string{
	"pg_locks": `
        SELECT pg_database.datname,tmp.mode,COALESCE(count,0) as count FROM
        (VALUES ('accesssharelock'),('rowsharelock'),('rowexclusivelock'),('shareupdateexclusivelock'),('sharelock'),('sharerowexclusivelock'),('exclusivelock'),('accessexclusivelock')) AS tmp(mode) CROSS JOIN pg_database
        LEFT JOIN
        (SELECT database, lower(mode) AS mode,count(*) AS count
        FROM pg_locks WHERE database IS NOT NULL
        GROUP BY database, lower(mode)
      ) AS tmp2
      ON tmp.mode=tmp2.mode and pg_database.oid = tmp2.database ORDER BY 1`,

	"pg_stat_replication": `
        SELECT *, pg_current_xlog_location(), pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float FROM pg_stat_replication
				INNER JOIN pg_replication_slots ON pg_stat_replication.pid=pg_replication_slots.active_pid`,

	"pg_stat_activity": `
      SELECT
          pg_database.datname,
          tmp.state,
          COALESCE(count,0) as count, 
          COALESCE(max_tx_duration,0) as max_tx_duration
      FROM
          (VALUES ('active'),('idle'),('idle in transaction'),('idle in transaction (aborted)'),('fastpath function call'),('disabled')) as tmp(state) CROSS JOIN pg_database
      LEFT JOIN
          (SELECT
              datname,
              state,
              count(*) AS count,
              MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
          FROM pg_stat_activity GROUP BY datname,state) as tmp2 
      ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname`,
}

// Add queries to the metricMaps and queryOverrides maps
func addQueries(queriesPath string) (err error) {
	var extra map[string]interface{}

	content, err := ioutil.ReadFile(queriesPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, &extra)
	if err != nil {
		return err
	}

	for metric, specs := range extra {
		for key, value := range specs.(map[interface{}]interface{}) {
			switch key.(string) {
			case "query":
				query := value.(string)
				queryOverrides[metric] = query

			case "metrics":
				for _, c := range value.([]interface{}) {
					column := c.(map[interface{}]interface{})

					for n, a := range column {
						var cmap ColumnMapping

						metric_map, ok := metricMaps[metric]
						if !ok {
							metric_map = make(map[string]ColumnMapping)
						}

						name := n.(string)

						for attr_key, attr_val := range a.(map[interface{}]interface{}) {
							switch attr_key.(string) {
							case "usage":
								usage, err := stringToColumnUsage(attr_val.(string))
								if err != nil {
									return err
								}
								cmap.usage = usage
							case "description":
								cmap.description = attr_val.(string)
							}
						}

						cmap.mapping = nil

						metric_map[name] = cmap

						metricMaps[metric] = metric_map
					}
				}
			}
		}
	}

	return
}

// Turn the MetricMap column mapping into a prometheus descriptor mapping.
func makeDescMap(metricMaps map[string]map[string]ColumnMapping) map[string]MetricMapNamespace {
	var metricMap = make(map[string]MetricMapNamespace)

	for namespace, mappings := range metricMaps {
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
					conversion: func(in interface{}) (float64, bool) {
						return math.NaN(), true
					},
				}
			case COUNTER:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.CounterValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					conversion: func(in interface{}) (float64, bool) {
						return dbToFloat64(in)
					},
				}
			case GAUGE:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					conversion: func(in interface{}) (float64, bool) {
						return dbToFloat64(in)
					},
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					conversion: func(in interface{}) (float64, bool) {
						text, ok := in.(string)
						if !ok {
							return math.NaN(), false
						}

						val, ok := columnMapping.mapping[text]
						if !ok {
							return math.NaN(), false
						}
						return val, true
					},
				}
			case DURATION:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s_milliseconds", namespace, columnName), columnMapping.description, constLabels, nil),
					conversion: func(in interface{}) (float64, bool) {
						var durationString string
						switch t := in.(type) {
						case []byte:
							durationString = string(t)
						case string:
							durationString = t
						default:
							log.Errorln("DURATION conversion metric was not a string")
							return math.NaN(), false
						}

						if durationString == "-1" {
							return math.NaN(), false
						}

						d, err := time.ParseDuration(durationString)
						if err != nil {
							log.Errorln("Failed converting result to metric:", columnName, in, err)
							return math.NaN(), false
						}
						return float64(d / time.Millisecond), true
					},
				}
			}
		}

		metricMap[namespace] = MetricMapNamespace{constLabels, thisMap}
	}

	return metricMap
}

// convert a string to the corresponding ColumnUsage
func stringToColumnUsage(s string) (u ColumnUsage, err error) {
	switch s {
	case "DISCARD":
		u = DISCARD

	case "LABEL":
		u = LABEL

	case "COUNTER":
		u = COUNTER

	case "GAUGE":
		u = GAUGE

	case "MAPPEDMETRIC":
		u = MAPPEDMETRIC

	case "DURATION":
		u = DURATION

	default:
		err = fmt.Errorf("wrong ColumnUsage given : %s", s)
	}

	return
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
	case []byte:
		// Try and convert to string and then parse to a float64
		strV := string(v)
		result, err := strconv.ParseFloat(strV, 64)
		if err != nil {
			return math.NaN(), false
		}
		return result, true
	case string:
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Infoln("Could not parse string:", err)
			return math.NaN(), false
		}
		return result, true
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

// Exporter collects Postgres metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn             string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	variableMap     map[string]MetricMapNamespace
	metricMap       map[string]MetricMapNamespace
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
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
		variableMap: makeDescMap(variableMaps),
		metricMap:   makeDescMap(metricMaps),
	}
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func ReadPassword(pathpass string) string {
	dat, err := ioutil.ReadFile(pathpass)
	check(err)
    return string(dat)
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
		log.Infoln("Error opening connection to database:", err)
		e.error.Set(1)
		return
	}
	defer db.Close()

	log.Debugln("Querying SHOW variables")
	for _, mapping := range e.variableMap {
		for columnName, columnMapping := range mapping.columnMappings {
			// Check for a discard request on this value
			if columnMapping.discard {
				continue
			}

			// Use SHOW to get the value
			row := db.QueryRow(fmt.Sprintf("SHOW %s;", columnName))

			var val interface{}
			err := row.Scan(&val)
			if err != nil {
				log.Errorln("Error scanning runtime variable:", columnName, err)
				continue
			}

			fval, ok := columnMapping.conversion(val)
			if !ok {
				e.error.Set(1)
				log.Errorln("Unexpected error parsing column: ", namespace, columnName, val)
				continue
			}

			ch <- prometheus.MustNewConstMetric(columnMapping.desc, columnMapping.vtype, fval)
		}
	}

	for namespace, mapping := range e.metricMap {
		log.Debugln("Querying namespace: ", namespace)
		func() {
			query, er := queryOverrides[namespace]
			if er == false {
				query = fmt.Sprintf("SELECT * FROM %s;", namespace)
			}

			// Don't fail on a bad scrape of one metric
			rows, err := db.Query(query)
			if err != nil {
				log.Infoln("Error running query on database: ", namespace, err)
				e.error.Set(1)
				return
			}
			defer rows.Close()

			var columnNames []string
			columnNames, err = rows.Columns()
			if err != nil {
				log.Infoln("Error retrieving column list for: ", namespace, err)
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
					log.Infoln("Error retrieving rows:", namespace, err)
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

	if *queriesPath != "" {
		err := addQueries(*queriesPath)
		if err != nil {
			log.Warnln("Unparseable queries file - discarding merge: ", *queriesPath, err)
		}
	}

	if *onlyDumpMaps {
		dumpMaps()
		return
	}

	if *password == "none" {
		*password = ReadPassword(*pathpassword)
	}
	
	dsn := os.Getenv("DATA_SOURCE_NAME")
	
	if len(dsn) == 0 {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", *user, *password, *host, *port, *params)
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