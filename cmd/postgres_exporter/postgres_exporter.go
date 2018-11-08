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
	listenAddress         = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9187").OverrideDefaultFromEnvar("PG_EXPORTER_WEB_LISTEN_ADDRESS").String()
	metricPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").OverrideDefaultFromEnvar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").OverrideDefaultFromEnvar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingMetrics = kingpin.Flag("disable-setting-metrics", "Do not include setting metrics.").Default("false").OverrideDefaultFromEnvar("PG_EXPORTER_DISABLE_SETTING_METRICS").Bool()
	queriesPath           = kingpin.Flag("extend.query-path", "Path to custom queries to run.").Default("").OverrideDefaultFromEnvar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps          = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList    = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,).").Default("").OverrideDefaultFromEnvar("PG_EXPORTER_CONTANT_LABELS").String()
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

// ColumnUsage should be one of several enum values which describe how a
// queried row is to be converted to a Prometheus metric.
type ColumnUsage int

// nolint: golint
const (
	DISCARD      ColumnUsage = iota // Ignore this column
	LABEL        ColumnUsage = iota // Use this column as a label
	COUNTER      ColumnUsage = iota // Use this column as a counter
	GAUGE        ColumnUsage = iota // Use this column as a gauge
	MAPPEDMETRIC ColumnUsage = iota // Use this column with the supplied mapping of text values
	DURATION     ColumnUsage = iota // This column should be interpreted as a text duration (and converted to milliseconds)
)

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (cu *ColumnUsage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}

	columnUsage, err := stringToColumnUsage(value)
	if err != nil {
		return err
	}

	*cu = columnUsage
	return nil
}

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

// ColumnMapping is the user-friendly representation of a prometheus descriptor map
type ColumnMapping struct {
	usage             ColumnUsage        `yaml:"usage"`
	description       string             `yaml:"description"`
	mapping           map[string]float64 `yaml:"metric_mapping"` // Optional column mapping for MAPPEDMETRIC
	supportedVersions semver.Range       `yaml:"pg_version"`     // Semantic version ranges which are supported. Unsupported columns are not queried (internally converted to DISCARD).
}

// UnmarshalYAML implements yaml.Unmarshaller
func (cm *ColumnMapping) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ColumnMapping
	return unmarshal((*plain)(cm))
}

// MetricMapNamespace groups metric maps under a shared set of labels.
type MetricMapNamespace struct {
	labels         []string             // Label names for this namespace
	columnMappings map[string]MetricMap // Column mappings in this namespace
}

// MetricMap stores the prometheus metric description which a given column will
// be mapped to by the collector
type MetricMap struct {
	discard    bool                              // Should metric be discarded during mapping?
	vtype      prometheus.ValueType              // Prometheus valuetype
	desc       *prometheus.Desc                  // Prometheus descriptor
	conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
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

		for column, details := range cmap {
			fmt.Printf("  %-40s %v\n", column, details)
		}
		fmt.Println()
	}
}

var builtinMetricMaps = map[string]map[string]ColumnMapping{
	"pg_stat_bgwriter": {
		"checkpoints_timed":     {COUNTER, "Number of scheduled checkpoints that have been performed", nil, nil},
		"checkpoints_req":       {COUNTER, "Number of requested checkpoints that have been performed", nil, nil},
		"checkpoint_write_time": {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds", nil, nil},
		"checkpoint_sync_time":  {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds", nil, nil},
		"buffers_checkpoint":    {COUNTER, "Number of buffers written during checkpoints", nil, nil},
		"buffers_clean":         {COUNTER, "Number of buffers written by the background writer", nil, nil},
		"maxwritten_clean":      {COUNTER, "Number of times the background writer stopped a cleaning scan because it had written too many buffers", nil, nil},
		"buffers_backend":       {COUNTER, "Number of buffers written directly by a backend", nil, nil},
		"buffers_backend_fsync": {COUNTER, "Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)", nil, nil},
		"buffers_alloc":         {COUNTER, "Number of buffers allocated", nil, nil},
		"stats_reset":           {COUNTER, "Time at which these statistics were last reset", nil, nil},
	},
	"pg_stat_database": {
		"datid":          {LABEL, "OID of a database", nil, nil},
		"datname":        {LABEL, "Name of this database", nil, nil},
		"numbackends":    {GAUGE, "Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.", nil, nil},
		"xact_commit":    {COUNTER, "Number of transactions in this database that have been committed", nil, nil},
		"xact_rollback":  {COUNTER, "Number of transactions in this database that have been rolled back", nil, nil},
		"blks_read":      {COUNTER, "Number of disk blocks read in this database", nil, nil},
		"blks_hit":       {COUNTER, "Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)", nil, nil},
		"tup_returned":   {COUNTER, "Number of rows returned by queries in this database", nil, nil},
		"tup_fetched":    {COUNTER, "Number of rows fetched by queries in this database", nil, nil},
		"tup_inserted":   {COUNTER, "Number of rows inserted by queries in this database", nil, nil},
		"tup_updated":    {COUNTER, "Number of rows updated by queries in this database", nil, nil},
		"tup_deleted":    {COUNTER, "Number of rows deleted by queries in this database", nil, nil},
		"conflicts":      {COUNTER, "Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)", nil, nil},
		"temp_files":     {COUNTER, "Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.", nil, nil},
		"temp_bytes":     {COUNTER, "Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.", nil, nil},
		"deadlocks":      {COUNTER, "Number of deadlocks detected in this database", nil, nil},
		"blk_read_time":  {COUNTER, "Time spent reading data file blocks by backends in this database, in milliseconds", nil, nil},
		"blk_write_time": {COUNTER, "Time spent writing data file blocks by backends in this database, in milliseconds", nil, nil},
		"stats_reset":    {COUNTER, "Time at which these statistics were last reset", nil, nil},
	},
	"pg_stat_database_conflicts": {
		"datid":            {LABEL, "OID of a database", nil, nil},
		"datname":          {LABEL, "Name of this database", nil, nil},
		"confl_tablespace": {COUNTER, "Number of queries in this database that have been canceled due to dropped tablespaces", nil, nil},
		"confl_lock":       {COUNTER, "Number of queries in this database that have been canceled due to lock timeouts", nil, nil},
		"confl_snapshot":   {COUNTER, "Number of queries in this database that have been canceled due to old snapshots", nil, nil},
		"confl_bufferpin":  {COUNTER, "Number of queries in this database that have been canceled due to pinned buffers", nil, nil},
		"confl_deadlock":   {COUNTER, "Number of queries in this database that have been canceled due to deadlocks", nil, nil},
	},
	"pg_locks": {
		"datname": {LABEL, "Name of this database", nil, nil},
		"mode":    {LABEL, "Type of Lock", nil, nil},
		"count":   {GAUGE, "Number of locks", nil, nil},
	},
	"pg_stat_replication": {
		"procpid":          {DISCARD, "Process ID of a WAL sender process", nil, semver.MustParseRange("<9.2.0")},
		"pid":              {DISCARD, "Process ID of a WAL sender process", nil, semver.MustParseRange(">=9.2.0")},
		"usesysid":         {DISCARD, "OID of the user logged into this WAL sender process", nil, nil},
		"usename":          {DISCARD, "Name of the user logged into this WAL sender process", nil, nil},
		"application_name": {DISCARD, "Name of the application that is connected to this WAL sender", nil, nil},
		"client_addr":      {LABEL, "IP address of the client connected to this WAL sender. If this field is null, it indicates that the client is connected via a Unix socket on the server machine.", nil, nil},
		"client_hostname":  {DISCARD, "Host name of the connected client, as reported by a reverse DNS lookup of client_addr. This field will only be non-null for IP connections, and only when log_hostname is enabled.", nil, nil},
		"client_port":      {DISCARD, "TCP port number that the client is using for communication with this WAL sender, or -1 if a Unix socket is used", nil, nil},
		"backend_start": {DISCARD, "with time zone	Time when this process was started, i.e., when the client connected to this WAL sender", nil, nil},
		"backend_xmin":             {DISCARD, "The current backend's xmin horizon.", nil, nil},
		"state":                    {LABEL, "Current WAL sender state", nil, nil},
		"sent_location":            {DISCARD, "Last transaction log position sent on this connection", nil, semver.MustParseRange("<10.0.0")},
		"write_location":           {DISCARD, "Last transaction log position written to disk by this standby server", nil, semver.MustParseRange("<10.0.0")},
		"flush_location":           {DISCARD, "Last transaction log position flushed to disk by this standby server", nil, semver.MustParseRange("<10.0.0")},
		"replay_location":          {DISCARD, "Last transaction log position replayed into the database on this standby server", nil, semver.MustParseRange("<10.0.0")},
		"sent_lsn":                 {DISCARD, "Last transaction log position sent on this connection", nil, semver.MustParseRange(">=10.0.0")},
		"write_lsn":                {DISCARD, "Last transaction log position written to disk by this standby server", nil, semver.MustParseRange(">=10.0.0")},
		"flush_lsn":                {DISCARD, "Last transaction log position flushed to disk by this standby server", nil, semver.MustParseRange(">=10.0.0")},
		"replay_lsn":               {DISCARD, "Last transaction log position replayed into the database on this standby server", nil, semver.MustParseRange(">=10.0.0")},
		"sync_priority":            {DISCARD, "Priority of this standby server for being chosen as the synchronous standby", nil, nil},
		"sync_state":               {DISCARD, "Synchronous state of this standby server", nil, nil},
		"slot_name":                {LABEL, "A unique, cluster-wide identifier for the replication slot", nil, semver.MustParseRange(">=9.2.0")},
		"plugin":                   {DISCARD, "The base name of the shared object containing the output plugin this logical slot is using, or null for physical slots", nil, nil},
		"slot_type":                {DISCARD, "The slot type - physical or logical", nil, nil},
		"datoid":                   {DISCARD, "The OID of the database this slot is associated with, or null. Only logical slots have an associated database", nil, nil},
		"database":                 {DISCARD, "The name of the database this slot is associated with, or null. Only logical slots have an associated database", nil, nil},
		"active":                   {DISCARD, "True if this slot is currently actively being used", nil, nil},
		"active_pid":               {DISCARD, "Process ID of a WAL sender process", nil, nil},
		"xmin":                     {DISCARD, "The oldest transaction that this slot needs the database to retain. VACUUM cannot remove tuples deleted by any later transaction", nil, nil},
		"catalog_xmin":             {DISCARD, "The oldest transaction affecting the system catalogs that this slot needs the database to retain. VACUUM cannot remove catalog tuples deleted by any later transaction", nil, nil},
		"restart_lsn":              {DISCARD, "The address (LSN) of oldest WAL which still might be required by the consumer of this slot and thus won't be automatically removed during checkpoints", nil, nil},
		"pg_current_xlog_location": {DISCARD, "pg_current_xlog_location", nil, nil},
		"pg_current_wal_lsn":       {DISCARD, "pg_current_xlog_location", nil, semver.MustParseRange(">=10.0.0")},
		"pg_xlog_location_diff":    {GAUGE, "Lag in bytes between master and slave", nil, semver.MustParseRange(">=9.2.0 <10.0.0")},
		"pg_wal_lsn_diff":          {GAUGE, "Lag in bytes between master and slave", nil, semver.MustParseRange(">=10.0.0")},
		"confirmed_flush_lsn":      {DISCARD, "LSN position a consumer of a slot has confirmed flushing the data received", nil, nil},
		"write_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written it (but not yet flushed it or applied it). This can be used to gauge the delay that synchronous_commit level remote_write incurred while committing if this server was configured as a synchronous standby.", nil, semver.MustParseRange(">=10.0.0")},
		"flush_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written and flushed it (but not yet applied it). This can be used to gauge the delay that synchronous_commit level remote_flush incurred while committing if this server was configured as a synchronous standby.", nil, semver.MustParseRange(">=10.0.0")},
		"replay_lag":               {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written, flushed and applied it. This can be used to gauge the delay that synchronous_commit level remote_apply incurred while committing if this server was configured as a synchronous standby.", nil, semver.MustParseRange(">=10.0.0")},
	},
	"pg_stat_activity": {
		"datname":         {LABEL, "Name of this database", nil, nil},
		"state":           {LABEL, "connection state", nil, semver.MustParseRange(">=9.2.0")},
		"count":           {GAUGE, "number of connections in this state", nil, nil},
		"max_tx_duration": {GAUGE, "max duration in seconds any active transaction has been running", nil, nil},
	},
}

// OverrideQuery 's are run in-place of simple namespace look ups, and provide
// advanced functionality. But they have a tendency to postgres version specific.
// There aren't too many versions, so we simply store customized versions using
// the semver matching we do for columns.
type OverrideQuery struct {
	versionRange semver.Range
	query        string
}

// Overriding queries for namespaces above.
// TODO: validate this is a closed set in tests, and there are no overlaps
var queryOverrides = map[string][]OverrideQuery{
	"pg_locks": {
		{
			semver.MustParseRange(">0.0.0"),
			`SELECT pg_database.datname,tmp.mode,COALESCE(count,0) as count
			FROM
				(
				  VALUES ('accesssharelock'),
				         ('rowsharelock'),
				         ('rowexclusivelock'),
				         ('shareupdateexclusivelock'),
				         ('sharelock'),
				         ('sharerowexclusivelock'),
				         ('exclusivelock'),
				         ('accessexclusivelock')
				) AS tmp(mode) CROSS JOIN pg_database
			LEFT JOIN
			  (SELECT database, lower(mode) AS mode,count(*) AS count
			  FROM pg_locks WHERE database IS NOT NULL
			  GROUP BY database, lower(mode)
			) AS tmp2
			ON tmp.mode=tmp2.mode and pg_database.oid = tmp2.database ORDER BY 1`,
		},
	},

	"pg_stat_replication": {
		{
			semver.MustParseRange(">=10.0.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_wal_lsn() end) AS pg_current_wal_lsn,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float end) AS pg_wal_lsn_diff
			FROM pg_stat_replication
			`,
		},
		{
			semver.MustParseRange(">=9.2.0 <10.0.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location,
				(case pg_is_in_recovery() when 't' then null else pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float end) AS pg_xlog_location_diff
			FROM pg_stat_replication
			`,
		},
		{
			semver.MustParseRange("<9.2.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location
			FROM pg_stat_replication
			`,
		},
	},

	"pg_stat_activity": {
		// This query only works
		{
			semver.MustParseRange(">=9.2.0"),
			`
			SELECT
				pg_database.datname,
				tmp.state,
				COALESCE(count,0) as count,
				COALESCE(max_tx_duration,0) as max_tx_duration
			FROM
				(
				  VALUES ('active'),
				  		 ('idle'),
				  		 ('idle in transaction'),
				  		 ('idle in transaction (aborted)'),
				  		 ('fastpath function call'),
				  		 ('disabled')
				) AS tmp(state) CROSS JOIN pg_database
			LEFT JOIN
			(
				SELECT
					datname,
					state,
					count(*) AS count,
					MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
				FROM pg_stat_activity GROUP BY datname,state) AS tmp2
				ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname
			`,
		},
		{
			semver.MustParseRange("<9.2.0"),
			`
			SELECT
				datname,
				'unknown' AS state,
				COALESCE(count(*),0) AS count,
				COALESCE(MAX(EXTRACT(EPOCH FROM now() - xact_start))::float,0) AS max_tx_duration
			FROM pg_stat_activity GROUP BY datname
			`,
		},
	},
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
func addQueries(content []byte, pgVersion semver.Version, exporterMap map[string]MetricMapNamespace, queryOverrideMap map[string]string, constLabels prometheus.Labels) error {
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
	partialExporterMap := makeDescMap(pgVersion, metricMaps, constLabels)

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

// Turn the MetricMap column mapping into a prometheus descriptor mapping.
func makeDescMap(pgVersion semver.Version, metricMaps map[string]map[string]ColumnMapping, constLabels prometheus.Labels) map[string]MetricMapNamespace {
	var metricMap = make(map[string]MetricMapNamespace)

	for namespace, mappings := range metricMaps {
		thisMap := make(map[string]MetricMap)

		// Get the constant labels
		var variableLabels []string
		for columnName, columnMapping := range mappings {
			if columnMapping.usage == LABEL {
				variableLabels = append(variableLabels, columnName)
			}
		}

		for columnName, columnMapping := range mappings {
			// Check column version compatibility for the current map
			// Force to discard if not compatible.
			if columnMapping.supportedVersions != nil {
				if !columnMapping.supportedVersions(pgVersion) {
					// It's very useful to be able to see what columns are being
					// rejected.
					log.Debugln(columnName, "is being forced to discard due to version incompatibility.")
					thisMap[columnName] = MetricMap{
						discard: true,
						conversion: func(_ interface{}) (float64, bool) {
							return math.NaN(), true
						},
					}
					continue
				}
			}

			// Determine how to convert the column based on its usage.
			// nolint: dupl
			switch columnMapping.usage {
			case DISCARD, LABEL:
				thisMap[columnName] = MetricMap{
					discard: true,
					conversion: func(_ interface{}) (float64, bool) {
						return math.NaN(), true
					},
				}
			case COUNTER:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.CounterValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, variableLabels, constLabels),
					conversion: func(in interface{}) (float64, bool) {
						return dbToFloat64(in)
					},
				}
			case GAUGE:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, variableLabels, constLabels),
					conversion: func(in interface{}) (float64, bool) {
						return dbToFloat64(in)
					},
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, variableLabels, constLabels),
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
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s_milliseconds", namespace, columnName), columnMapping.description, variableLabels, constLabels),
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

		metricMap[namespace] = MetricMapNamespace{variableLabels, thisMap}
	}

	return metricMap
}

// convert a string to the corresponding ColumnUsage
func stringToColumnUsage(s string) (ColumnUsage, error) {
	var u ColumnUsage
	var err error
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

	return u, err
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
			log.Infoln("Could not parse []byte:", err)
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
	// Holds a reference to the build in column mappings. Currently this is for testing purposes
	// only, since it just points to the global.
	builtinMetricMaps map[string]map[string]ColumnMapping

	dsn                   string
	disableSettingMetrics bool
	disableDefaultMetrics bool
	constLabels           prometheus.Labels
	userQueriesPath       string
	duration              prometheus.Gauge
	error                 prometheus.Gauge
	psqlUp                prometheus.Gauge
	userQueriesError      *prometheus.GaugeVec
	totalScrapes          prometheus.Counter

	// dbDsn is the connection string used to establish the dbConnection
	dbDsn string
	// dbConnection is used to allow re-using the DB connection between scrapes
	dbConnection *sql.DB

	// Last version used to calculate metric map. If mismatch on scrape,
	// then maps are recalculated.
	lastMapVersion semver.Version
	// Currently active metric map
	metricMap map[string]MetricMapNamespace
	// Currently active query overrides
	queryOverrides map[string]string
	mappingMtx     sync.RWMutex
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
func NewExporter(dsn string, disableDefaultMetrics, disableSettingMetrics bool, userQueriesPath string, constLabels prometheus.Labels) *Exporter {
	return &Exporter{
		builtinMetricMaps:     builtinMetricMaps,
		dsn:                   dsn,
		disableSettingMetrics: disableSettingMetrics,
		disableDefaultMetrics: disableDefaultMetrics,
		userQueriesPath:       userQueriesPath,
		constLabels:           constLabels,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   exporter,
			Name:        "last_scrape_duration_seconds",
			Help:        "Duration of the last scrape of metrics from PostgresSQL.",
			ConstLabels: constLabels,
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   exporter,
			Name:        "scrapes_total",
			Help:        "Total number of times PostgresSQL was scraped for metrics.",
			ConstLabels: constLabels,
		}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   exporter,
			Name:        "last_scrape_error",
			Help:        "Whether the last scrape of metrics from PostgreSQL resulted in an error (1 for error, 0 for success).",
			ConstLabels: constLabels,
		}),
		psqlUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   exporter,
			Name:        "up",
			Help:        "Whether the last scrape of metrics from PostgreSQL was able to connect to the server (1 for yes, 0 for no).",
			ConstLabels: constLabels,
		}),
		userQueriesError: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   exporter,
			Name:        "user_queries_load_error",
			Help:        "Whether the user queries file was loaded and parsed successfully (1 for error, 0 for success).",
			ConstLabels: constLabels,
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

func newConstLabels(constantLabelsList string) prometheus.Labels {
	if constantLabelsList == "" {
		return nil
	}

	var constLabels = make(prometheus.Labels)
	parts := strings.Split(constantLabelsList, ",")
	for _, p := range parts {
		keyValue := strings.Split(strings.TrimSpace(p), "=")
		if len(keyValue) != 2 {
			continue
		}
		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])
		if key == "" || value == "" {
			continue
		}
		constLabels[key] = value
	}
	return constLabels
}

func newDesc(subsystem, name, help string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, constLabels,
	)
}

// Query within a namespace mapping and emit metrics. Returns fatal errors if
// the scrape fails, and a slice of errors if they were non-fatal.
func queryNamespaceMapping(ch chan<- prometheus.Metric, db *sql.DB, namespace string, mapping MetricMapNamespace, queryOverrides map[string]string, constLabels prometheus.Labels) ([]error, error) {
	// Check for a query override for this namespace
	query, found := queryOverrides[namespace]

	// Was this query disabled (i.e. nothing sensible can be queried on cu
	// version of PostgreSQL?
	if query == "" && found {
		// Return success (no pertinent data)
		return []error{}, nil
	}

	// Don't fail on a bad scrape of one metric
	var rows *sql.Rows
	var err error

	if !found {
		// I've no idea how to avoid this properly at the moment, but this is
		// an admin tool so you're not injecting SQL right?
		rows, err = db.Query(fmt.Sprintf("SELECT * FROM %s;", namespace)) // nolint: gas, safesql
	} else {
		rows, err = db.Query(query) // nolint: safesql
	}
	if err != nil {
		return []error{}, errors.New(fmt.Sprintln("Error running query on database: ", namespace, err))
	}
	defer rows.Close() // nolint: errcheck

	var columnNames []string
	columnNames, err = rows.Columns()
	if err != nil {
		return []error{}, errors.New(fmt.Sprintln("Error retrieving column list for: ", namespace, err))
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

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return []error{}, errors.New(fmt.Sprintln("Error retrieving rows:", namespace, err))
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
					nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName, columnData[idx])))
					continue
				}

				// Generate the metric
				ch <- prometheus.MustNewConstMetric(metricMapping.desc, metricMapping.vtype, value, labels...)
			} else {
				// Unknown metric. Report as untyped if scan to float64 works, else note an error too.
				metricLabel := fmt.Sprintf("%s_%s", namespace, columnName)
				desc := prometheus.NewDesc(metricLabel, fmt.Sprintf("Unknown metric from %s", namespace), mapping.labels, constLabels)

				// Its not an error to fail here, since the values are
				// unexpected anyway.
				value, ok := dbToFloat64(columnData[idx])
				if !ok {
					nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unparseable column type - discarding: ", namespace, columnName, err)))
					continue
				}
				ch <- prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, value, labels...)
			}
		}
	}
	return nonfatalErrors, nil
}

// Iterate through all the namespace mappings in the exporter and run their
// queries.
func queryNamespaceMappings(ch chan<- prometheus.Metric, db *sql.DB, metricMap map[string]MetricMapNamespace, queryOverrides map[string]string, constLabels prometheus.Labels) map[string]error {
	// Return a map of namespace -> errors
	namespaceErrors := make(map[string]error)

	for namespace, mapping := range metricMap {
		log.Debugln("Querying namespace: ", namespace)
		nonFatalErrors, err := queryNamespaceMapping(ch, db, namespace, mapping, queryOverrides, constLabels)
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
	}

	return namespaceErrors
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
	if !e.disableDefaultMetrics && semanticVersion.LT(lowestSupportedVersion) {
		log.Warnln("PostgreSQL version is lower then our lowest supported version! Got", semanticVersion.String(), "minimum supported is", lowestSupportedVersion.String())
	}

	// Check if semantic version changed and recalculate maps if needed.
	if semanticVersion.NE(e.lastMapVersion) || e.metricMap == nil {
		log.Infoln("Semantic Version Changed:", e.lastMapVersion.String(), "->", semanticVersion.String())
		e.mappingMtx.Lock()

		if e.disableDefaultMetrics {
			e.metricMap = make(map[string]MetricMapNamespace)
			e.queryOverrides = make(map[string]string)
		} else {
			e.metricMap = makeDescMap(semanticVersion, e.builtinMetricMaps, e.constLabels)
			e.queryOverrides = makeQueryOverrideMap(semanticVersion, queryOverrides)
		}

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

				if err := addQueries(userQueriesData, semanticVersion, e.metricMap, e.queryOverrides, e.constLabels); err != nil {
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
		"Version string as reported by postgres", []string{"version", "short_version"}, e.constLabels)

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
	if !e.disableSettingMetrics {
		if err := querySettings(ch, db, e.constLabels); err != nil {
			log.Infof("Error retrieving settings: %s", err)
			e.error.Set(1)
		}
	}

	errMap := queryNamespaceMappings(ch, db, e.metricMap, e.queryOverrides, e.constLabels)
	if len(errMap) > 0 {
		e.error.Set(1)
	}
}

func getDataFromEnv(file string) (dataList []string) {
	fileContents, err := ioutil.ReadFile(os.Getenv(file))
	if err != nil {
		panic(err)
	}
	tmp := strings.TrimSpace(string(fileContents))
	if strings.Contains(tmp, ",") {
		dataList = strings.Split(tmp, ",")
	} else {
		dataList = strings.Split(tmp, "\n")
	}
	return dataList
}

// try to get the DataSource
// DATA_SOURCE_NAME always wins so we do not break older versions
// reading secrets from files wins over secrets in environment variables
// DATA_SOURCE_NAME > DATA_SOURCE_{USER|PASS}_FILE > DATA_SOURCE_{USER|PASS}
func getDataSource() []string {
	var dsn []string
	dsnEnv := os.Getenv("DATA_SOURCE_NAME")
	if len(dsnEnv) == 0 {
		var user, pass []string

		if len(os.Getenv("DATA_SOURCE_USER_FILE")) != 0 {
			user = getDataFromEnv("DATA_SOURCE_USER_FILE")
		} else {
			tmp := os.Getenv("DATA_SOURCE_USER")
			user = strings.Split(tmp, ",")
		}

		if len(os.Getenv("DATA_SOURCE_PASS_FILE")) != 0 {
			pass = getDataFromEnv("DATA_SOURCE_PASS_FILE")
		} else {
			tmp := os.Getenv("DATA_SOURCE_PASS")
			pass = strings.Split(tmp, ",")
		}

		if len(user) != 1 && len(user) < len(pass) {
			panic("Not enough usernames")
		} else if len(pass) != 1 && len(user) > len(pass) {
			panic("Not enough passwords")
		}

		tmp := os.Getenv("DATA_SOURCE_URI")
		uris := strings.Split(tmp, ",")

		if len(uris) < len(user) || len(uris) < len(pass) {
			panic("Not enough uri's")
		}

		for i, uri := range uris {
			currUser := user[0]
			currPass := pass[0]
			if len(user) > 1 {
				currUser = user[i]
			}
			if len(pass) > 1 {
				currPass = pass[i]
			}
			ui := url.UserPassword(currUser, currPass).String()
			dsn = append(dsn, "postgresql://"+ui+"@"+uri)
		}
		return dsn
	}
	dsn = strings.Split(dsnEnv, ",")
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

	var prevPort string
	portFind, _ := regexp.Compile(":[0-9]+/")
	for i, currDsn := range dsn {
		// This is for avoid duplicating metrics
		currDsnDisableDefaultMetrics := *disableDefaultMetrics
		currDsnDisableSettingMetrics := *disableSettingMetrics

		currPort := portFind.FindString(currDsn)
		if currPort != prevPort {
			prevPort = currPort
		} else {
			currDsnDisableDefaultMetrics = true
			currDsnDisableSettingMetrics = true
		}
		// This is need for differing exporters
		constExporterLabels := fmt.Sprintf("exporter=%d", i)
		if len(*constantLabelsList) > 0 {
			constExporterLabels = *constantLabelsList + "," + constExporterLabels
		}
		convertedConstLabels := newConstLabels(constExporterLabels)

		exporter := NewExporter(currDsn, currDsnDisableDefaultMetrics, currDsnDisableSettingMetrics, *queriesPath, convertedConstLabels)
		defer func() {
			if exporter.dbConnection != nil {
				exporter.dbConnection.Close() // nolint: errcheck
			}
		}()

		prometheus.MustRegister(exporter)

		// This is for removing same default metrics
	}

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type:text/plain; charset=UTF-8") // nolint: errcheck
		w.Write(landingPage)                                                     // nolint: errcheck
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
