package metricmaps

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/wrouesnel/postgres_exporter/pkg/pgdbconv"
	"math"
	"time"
)

// Turn the MetricMap column mapping into a prometheus descriptor mapping.
func makeDescMap(pgVersion semver.Version, serverLabels prometheus.Labels, metricMaps map[string]IntermediateMetricMap) map[string]MetricMapNamespace {
	var metricMap = make(map[string]MetricMapNamespace)

	for namespace, intermediateMappings := range metricMaps {
		thisMap := make(map[string]MetricMap)

		// Get the constant labels
		var variableLabels []string
		for columnName, columnMapping := range intermediateMappings.ColumnMappings {
			if columnMapping.Usage == LABEL {
				variableLabels = append(variableLabels, columnName)
			}
		}

		for columnName, columnMapping := range intermediateMappings.ColumnMappings {
			// Check column version compatibility for the current map
			// Force to discard if not compatible.
			if columnMapping.SupportedVersions != nil {
				if !columnMapping.SupportedVersions.Range(pgVersion) {
					// It's very useful to be able to see what columns are being
					// rejected.
					log.Debugln(columnName, "is being forced to discard due to version incompatibility.")
					thisMap[columnName] = MetricMap{
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
			switch columnMapping.Usage {
			case DISCARD, LABEL:
				thisMap[columnName] = MetricMap{
					Discard: true,
					Conversion: func(_ interface{}) (float64, bool) {
						return math.NaN(), true
					},
				}
			case COUNTER:
				thisMap[columnName] = MetricMap{
					ValueType: prometheus.CounterValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					Conversion: func(in interface{}) (float64, bool) {
						return pgdbconv.DBToFloat64(in)
					},
				}
			case GAUGE:
				thisMap[columnName] = MetricMap{
					ValueType: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					Conversion: func(in interface{}) (float64, bool) {
						return pgdbconv.DBToFloat64(in)
					},
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMap{
					ValueType: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					Conversion: func(in interface{}) (float64, bool) {
						text, ok := in.(string)
						if !ok {
							return math.NaN(), false
						}

						val, ok := columnMapping.Mapping[text]
						if !ok {
							return math.NaN(), false
						}
						return val, true
					},
				}
			case DURATION:
				thisMap[columnName] = MetricMap{
					ValueType: prometheus.GaugeValue,
					Desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s_milliseconds", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					Conversion: func(in interface{}) (float64, bool) {
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

		metricMap[namespace] = MetricMapNamespace{
			Labels:         variableLabels,
			ColumnMappings: thisMap,
			Master:         intermediateMappings.Master,
			CacheSeconds:   intermediateMappings.CacheSeconds}
	}

	return metricMap
}
