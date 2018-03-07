package metricmaps

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/common/log"
	"math"
	"github.com/wrouesnel/postgres_exporter/exporter/dbconv"
	"time"
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

//map[string]map[string]ColumnMapping

//type ColumnMap

// ColumnMapping is the user-friendly representation of a prometheus descriptor map
type ColumnMapping struct {
	usage             ColumnUsage        `yaml:"usage"`
	description       string             `yaml:"description"`
	mapping           map[string]float64 `yaml:"metric_mapping"` // Optional column mapping for MAPPEDMETRIC
	// Semantic version ranges which are supported.
	// Unsupported columns are not queried (internally converted to DISCARD).
	supportedVersions semver.Range       `yaml:"pg_version"`
}

// UnmarshalYAML implements yaml.Unmarshaller
func (cm *ColumnMapping) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ColumnMapping
	return unmarshal((*plain)(cm))
}

type ColumnMap map[string]MetricMapping

// MetricMapNamespace groups metric maps under a shared set of Labels.
type MetricMapNamespace struct {
	Labels         []string                 // Label names for this namespace
	ColumnMappings ColumnMap // Column mappings in this namespace
}

type MetricMapNamespaceMapping map[string]MetricMapNamespace

// MetricMapping stores the prometheus metric description which a given column will
// be mapped to by the exporter
type MetricMapping struct {
	Discard    bool                              // Should metric be discarded during mapping?
	Vtype      prometheus.ValueType              // Prometheus valuetype
	Desc       *prometheus.Desc                  // Prometheus descriptor
	Conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
}

type MetricMaps map[string]map[string]ColumnMapping

// OverrideQuerys are run in-place of simple namespace look ups, and provide
// advanced functionality. But they have a tendency to postgres version specific.
// There aren't too many versions, so we simply store customized versions using
// the semver matching we do for columns. This is the user-friendly version of the input
// data structure.
type OverrideQuery struct {
	versionRange semver.Range
	query        string
}

// NamespaceOverrideQueryMapping is the processed structure for override queries.
type NamespaceOverrideQueryMapping map[string]string

// MakeDescMapForVersion turns high-level MetricMaps into MetricMapNamespaceMappings, which can be used to produce
// metrics from database scrapes. Each mapping is applicable only to the version of postgres it is initialized with.
func MakeDescMapForVersion(pgVersion semver.Version, metricMaps MetricMaps) MetricMapNamespaceMapping {
	var metricMap = make(map[string]MetricMapNamespace)

	for namespace, mappings := range metricMaps {
		thisMap := make(map[string]MetricMapping)

		// Get the constant Labels
		var constLabels []string
		for columnName, columnMapping := range mappings {
			if columnMapping.usage == LABEL {
				constLabels = append(constLabels, columnName)
			}
		}

		for columnName, columnMapping := range mappings {
			// Check column version compatibility for the current map
			// Force to Discard if not compatible.
			if columnMapping.supportedVersions != nil {
				if !columnMapping.supportedVersions(pgVersion) {
					// It's very useful to be able to see what columns are being
					// rejected.
					log.Debugln(columnName, "is being forced to Discard due to version incompatibility.")
					thisMap[columnName] = MetricMapping{
						Discard: true,
						Conversion: func(_ interface{}) (float64, bool) {
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
				thisMap[columnName] = MetricMapping{
					Discard: true,
					Conversion: func(_ interface{}) (float64, bool) {
						return math.NaN(), true
					},
				}
			case COUNTER:
				thisMap[columnName] = MetricMapping{
					Vtype: prometheus.CounterValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					Conversion: func(in interface{}) (float64, bool) {
						return dbconv.DbToFloat64(in)
					},
				}
			case GAUGE:
				thisMap[columnName] = MetricMapping{
					Vtype: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					Conversion: func(in interface{}) (float64, bool) {
						return dbconv.DbToFloat64(in)
					},
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMapping{
					Vtype: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.description, constLabels, nil),
					Conversion: func(in interface{}) (float64, bool) {
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
				thisMap[columnName] = MetricMapping{
					Vtype: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s_milliseconds", namespace, columnName), columnMapping.description, constLabels, nil),
					Conversion: func(in interface{}) (float64, bool) {
						var durationString string
						switch t := in.(type) {
						case []byte:
							durationString = string(t)
						case string:
							durationString = t
						default:
							log.Errorln("DURATION Conversion metric was not a string")
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