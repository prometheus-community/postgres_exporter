package metricmaps

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// QueryMap is the global holder structure for the new unified configuration format.
type QueryMap struct {
	Global *QueryConfig	`yaml:"global"`
	ByServer map[string]*QueryConfig `yaml:"by_server"`
}

// QueryConfig holds a specific mapping and query override config
type QueryConfig struct {
	MetricMap MetricMaps	`yaml:"metric_maps"`
	QueryOverrides QueryOverrides `yaml:"query_override"`
}

// MappingOptions is a copy of ColumnMapping used only for parsing
type MappingOptions struct {
	Usage             string             `yaml:"usage"`
	Description       string             `yaml:"description"`
	Mapping           map[string]float64 `yaml:"metric_mapping"` // Optional column mapping for MAPPEDMETRIC
	SupportedVersions semver.Range       `yaml:"pg_version"`     // Semantic version ranges which are supported. Unsupported columns are not queried (internally converted to DISCARD).
}

// QueryOverrides ensures our query types are consistent
type QueryOverrides map[string][]OverrideQuery

// MetricMaps is a stub type to assist with checking
type MetricMaps map[string]IntermediateMetricMap

// IntermediateMetricMap holds the partially loaded metric map parsing.
// This is mainly so we can parse cacheSeconds around.
type IntermediateMetricMap struct {
	ColumnMappings map[string]ColumnMapping
	Master         bool
	CacheSeconds   uint64
}

// MetricMapNamespace groups metric maps under a shared set of labels.
type MetricMapNamespace struct {
	Labels         []string             // Label names for this ExporterNamespaceLabel
	ColumnMappings map[string]MetricMap // Column mappings in this ExporterNamespaceLabel
	Master         bool                 // Call query only for master database
	CacheSeconds   uint64               // Number of seconds this metric ExporterNamespaceLabel can be cached. 0 disables.
}

// MetricMap stores the prometheus metric description which a given column will
// be mapped to by the collector
type MetricMap struct {
	Discard    bool                              // Should metric be discarded during mapping?
	ValueType  prometheus.ValueType              // Prometheus valuetype
	Desc       *prometheus.Desc                  // Prometheus descriptor
	Conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
}
