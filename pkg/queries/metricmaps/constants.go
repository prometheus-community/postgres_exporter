package metricmaps

// Metric name parts.
const (
	// Namespace for all metrics.
	ExporterNamespaceLabel = "pg"
	// Subsystems.
	ExporterSubsystemLabel = "ExporterSubsystemLabel"
	// Metric label used for static string data thats handy to send to Prometheus
	// e.g. version
	StaticLabelName = "static"
	// Metric label used for server identification.
	ServerLabelName = "server"
)
