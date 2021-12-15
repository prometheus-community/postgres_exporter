package postgres_exporter

const (
	// DISCARD ignores a column
	DISCARD ColumnUsage = iota
	// LABEL identifies a column as a label
	LABEL ColumnUsage = iota
	// COUNTER identifies a column as a counter
	COUNTER ColumnUsage = iota
	// GAUGE identifies a column as a gauge
	GAUGE ColumnUsage = iota
	// MAPPEDMETRIC identifies a column as a mapping of text values
	MAPPEDMETRIC ColumnUsage = iota
	// DURATION identifies a column as a text duration (and converted to milliseconds)
	DURATION ColumnUsage = iota
	// HISTOGRAM identifies a column as a histogram
	HISTOGRAM ColumnUsage = iota

	// Namespace for all metrics.
	Namespace = "pg"
	// Subsystems.
	exporter = "exporter"
	// The name of the exporter.
	ExporterName = "postgres_exporter"
	// Metric label used for static string data thats handy to send to Prometheus
	// e.g. version
	StaticLabelName = "static"
	// Metric label used for server identification.
	ServerLabelName = "server"
)
