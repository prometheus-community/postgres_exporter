package pgexporter

import "strings"

// ExporterOpt configures Exporter.
type ExporterOpt func(*Exporter)

// DisableDefaultMetrics configures default metrics export.
func DisableDefaultMetrics(b bool) ExporterOpt {
	return func(e *Exporter) {
		e.disableDefaultMetrics = b
	}
}

// DisableSettingsMetrics configures pg_settings export.
func DisableSettingsMetrics(b bool) ExporterOpt {
	return func(e *Exporter) {
		e.disableSettingsMetrics = b
	}
}

// AutoDiscoverDatabases allows scraping all databases on a database server.
func AutoDiscoverDatabases(b bool) ExporterOpt {
	return func(e *Exporter) {
		e.autoDiscoverDatabases = b
	}
}

// ExcludeDatabases allows to filter out result from AutoDiscoverDatabases
func ExcludeDatabases(s string) ExporterOpt {
	return func(e *Exporter) {
		e.excludeDatabases = strings.Split(s, ",")
	}
}

// WithUserQueriesPath configures user's queries path.
func WithUserQueriesPath(p string) ExporterOpt {
	return func(e *Exporter) {
		e.userQueriesPath = p
	}
}

// WithConstantLabels configures constant labels.
func WithConstantLabels(s string) ExporterOpt {
	return func(e *Exporter) {
		e.constantLabels = parseConstLabels(s)
	}
}
