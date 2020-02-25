package metricmaps

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// MappingOptions is a copy of ColumnMapping used only for parsing
type MappingOptions struct {
	Usage             string             `yaml:"usage"`
	Description       string             `yaml:"description"`
	Mapping           map[string]float64 `yaml:"metric_mapping"` // Optional column mapping for MAPPEDMETRIC
	SupportedVersions semver.Range       `yaml:"pg_version"`     // Semantic version ranges which are supported. Unsupported columns are not queried (internally converted to DISCARD).
}

// QueryOverrides ensures our query types are consistent
type QueryOverrides map[string]string

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
	Labels         []string             // Label names for this namespace
	ColumnMappings map[string]MetricMap // Column mappings in this namespace
	Master         bool                 // Call query only for master database
	CacheSeconds   uint64               // Number of seconds this metric namespace can be cached. 0 disables.
}

// MetricMap stores the prometheus metric description which a given column will
// be mapped to by the collector
type MetricMap struct {
	Discard    bool                              // Should metric be discarded during mapping?
	ValueType      prometheus.ValueType              // Prometheus valuetype
	Desc       *prometheus.Desc                  // Prometheus descriptor
	Conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
}