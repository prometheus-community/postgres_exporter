// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres_exporter

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log/level"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

type NamespaceMappings struct {
}

// compile-time check that type implements interface.
var _ NamespaceMetricsAPI = (*NamespaceMappings)(nil)

// MetricMapNamespace groups metric maps under a shared set of labels.
type MetricMapNamespace struct {
	labels         []string             // Label names for this namespace
	columnMappings map[string]MetricMap // Column mappings in this namespace
}

// MetricMap stores the prometheus metric description which a given column will
// be mapped to by the collector
type MetricMap struct {
	discard    bool                              // Should metric be discarded during mapping?
	histogram  bool                              // Should metric be treated as a histogram?
	vtype      prometheus.ValueType              // Prometheus valuetype
	desc       *prometheus.Desc                  // Prometheus descriptor
	conversion func(interface{}) (float64, bool) // Conversion function to turn PG result into float64
}

// Query within a namespace mapping and emit metrics. Returns fatal errors if
// the scrape fails, and a slice of errors if they were non-fatal.
func (n *NamespaceMappings) queryNamespaceMapping(db *sql.DB, namespace string, mapping MetricMapNamespace, serverLabels prometheus.Labels, queryList map[string]string) ([]prometheus.Metric, []error, error) {
	// Check for a query override for this namespace
	query, found := queryList[namespace]

	// Was this query disabled (i.e. nothing sensible can be queried on cu
	// version of PostgreSQL?
	if query == "" && found {
		// Return success (no pertinent data)
		return []prometheus.Metric{}, []error{}, nil
	}

	if !found {
		// Return success (no pertinent data)
		return []prometheus.Metric{}, []error{}, nil
	}

	// Don't fail on a bad scrape of one metric
	var rows *sql.Rows
	var err error

	rows, err = db.Query(query)
	if err != nil {
		return []prometheus.Metric{}, []error{}, fmt.Errorf("Error running query: %s %v", namespace, err)
	}

	var columnNames []string
	columnNames, err = rows.Columns()
	if err != nil {
		return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving column list for: ", namespace, err))
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

	metrics := make([]prometheus.Metric, 0)

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving rows:", namespace, err))
		}

		// Get the label values for this row.
		labels := make([]string, len(mapping.labels))
		for idx, label := range mapping.labels {
			labels[idx], _ = DbToString(columnData[columnIdx[label]])
		}

		// Loop over column names, and match to scan data. Unknown columns
		// will be filled with an untyped metric number *if* they can be
		// converted to float64s. NULLs are allowed and treated as NaN.
		for idx, columnName := range columnNames {
			var metric prometheus.Metric

			if metricMapping, ok := mapping.columnMappings[columnName]; ok {
				// Is this a metricy metric?
				if metricMapping.discard {
					continue
				}
				if metricMapping.histogram {
					var keys []float64
					err = pq.Array(&keys).Scan(columnData[idx])
					if err != nil {
						return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving", columnName, "buckets:", namespace, err))
					}

					var values []int64
					valuesIdx, ok := columnIdx[columnName+"_bucket"]
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Missing column: ", namespace, columnName+"_bucket")))
						continue
					}
					err = pq.Array(&values).Scan(columnData[valuesIdx])
					if err != nil {
						return []prometheus.Metric{}, []error{}, errors.New(fmt.Sprintln("Error retrieving", columnName, "bucket values:", namespace, err))
					}

					buckets := make(map[float64]uint64, len(keys))
					for i, key := range keys {
						if i >= len(values) {
							break
						}
						buckets[key] = uint64(values[i])
					}

					idx, ok = columnIdx[columnName+"_sum"]
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Missing column: ", namespace, columnName+"_sum")))
						continue
					}
					sum, ok := DbToFloat64(columnData[idx])
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName+"_sum", columnData[idx])))
						continue
					}

					idx, ok = columnIdx[columnName+"_count"]
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Missing column: ", namespace, columnName+"_count")))
						continue
					}
					count, ok := DbToUint64(columnData[idx])
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName+"_count", columnData[idx])))
						continue
					}

					metric = prometheus.MustNewConstHistogram(
						metricMapping.desc,
						count, sum, buckets,
						labels...,
					)
				} else {
					value, ok := DbToFloat64(columnData[idx])
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName, columnData[idx])))
						continue
					}
					// Generate the metric
					metric = prometheus.MustNewConstMetric(metricMapping.desc, metricMapping.vtype, value, labels...)
				}
			} else {
				// Unknown metric. Report as untyped if scan to float64 works, else note an error too.
				metricLabel := fmt.Sprintf("%s_%s", namespace, columnName)
				desc := prometheus.NewDesc(metricLabel, fmt.Sprintf("Unknown metric from %s", namespace), mapping.labels, serverLabels)

				// Its not an error to fail here, since the values are
				// unexpected anyway.
				value, ok := DbToFloat64(columnData[idx])
				if !ok {
					nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unparseable column type - discarding: ", namespace, columnName, err)))
					continue
				}
				metric = prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, value, labels...)
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nonfatalErrors, nil
}

// Iterate through all the namespace mappings in the exporter and run their
// queries.
func (n *NamespaceMappings) QueryNamespaceMappings(ch chan<- prometheus.Metric, db *sql.DB, serverLabels prometheus.Labels, queryList map[string]string,
	metricMaps map[string]IntermediateMetricMap,
	semanticVersion semver.Version, versionString string) map[string]error {

	// Return a map of namespace -> errors
	namespaceErrors := make(map[string]error)

	for namespace, mapping := range n.makeDescMap(serverLabels, metricMaps) {

		level.Debug(logger).Log("msg", "Querying namespace", "namespace", namespace)

		metrics, nonFatalErrors, err := n.queryNamespaceMapping(db, namespace, mapping, serverLabels, queryList)
		// Serious error - a namespace disappeared
		if err != nil {
			namespaceErrors[namespace] = err
			level.Info(logger).Log("err", err)
		}
		// Non-serious errors - likely version or parsing problems.
		if len(nonFatalErrors) > 0 {
			for _, err := range nonFatalErrors {
				level.Info(logger).Log("err", err)
			}
		}

		// Emit the metrics into the channel
		for _, metric := range metrics {
			ch <- metric
		}

	}

	versionDesc := prometheus.NewDesc(fmt.Sprintf("%s_%s", Namespace, StaticLabelName),
		"Version string as reported by postgres", []string{"version", "short_version"}, serverLabels)

	// Emit the metric info in the channel
	ch <- prometheus.MustNewConstMetric(versionDesc,
		prometheus.UntypedValue, 1, versionString, semanticVersion.String())

	return namespaceErrors
}

// Turn the MetricMap column mapping into a prometheus descriptor mapping.
func (n *NamespaceMappings) makeDescMap(serverLabels prometheus.Labels, metricMaps map[string]IntermediateMetricMap) map[string]MetricMapNamespace {
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

			// Determine how to convert the column based on its usage.
			// nolint: dupl
			switch columnMapping.Usage {
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
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					conversion: func(in interface{}) (float64, bool) {
						return DbToFloat64(in)
					},
				}
			case GAUGE:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					conversion: func(in interface{}) (float64, bool) {
						return DbToFloat64(in)
					},
				}
			case HISTOGRAM:
				thisMap[columnName] = MetricMap{
					histogram: true,
					vtype:     prometheus.UntypedValue,
					desc:      prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					conversion: func(in interface{}) (float64, bool) {
						return DbToFloat64(in)
					},
				}
				thisMap[columnName+"_bucket"] = MetricMap{
					histogram: true,
					discard:   true,
				}
				thisMap[columnName+"_sum"] = MetricMap{
					histogram: true,
					discard:   true,
				}
				thisMap[columnName+"_count"] = MetricMap{
					histogram: true,
					discard:   true,
				}
			case MAPPEDMETRIC:
				thisMap[columnName] = MetricMap{
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					conversion: func(in interface{}) (float64, bool) {
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
					vtype: prometheus.GaugeValue,
					desc:  prometheus.NewDesc(fmt.Sprintf("%s_%s_milliseconds", namespace, columnName), columnMapping.Description, variableLabels, serverLabels),
					conversion: func(in interface{}) (float64, bool) {
						var durationString string
						switch t := in.(type) {
						case []byte:
							durationString = string(t)
						case string:
							durationString = t
						default:
							level.Error(logger).Log("msg", "Duration conversion metric was not a string")
							return math.NaN(), false
						}

						if durationString == "-1" {
							return math.NaN(), false
						}

						d, err := time.ParseDuration(durationString)
						if err != nil {
							level.Error(logger).Log("msg", "Failed converting result to metric", "column", columnName, "in", in, "err", err)
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

func (n *NamespaceMappings) SetInternalMetrics(ch chan<- prometheus.Metric, duration, totalScrapes, rdsDatabaseConnections, rdsCurrentCapacity float64) {
	durationMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: exporter,
		Name:      "last_scrape_duration_seconds",
		Help:      "Duration of the last scrape of metrics from PostgresSQL.",
	})
	totalScrapesMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: exporter,
		Name:      "scrapes_total",
		Help:      "Total number of times PostgresSQL was scraped for metrics.",
	})
	rdsCurrentCapacityMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "rds_current_capacity",
		Help:      "Current Aurora capacity units",
	})
	rdsDatabaseConnectionsMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "rds_database_connections",
		Help:      "Current Aurora database connections",
	})

	durationMetric.Set(duration)
	totalScrapesMetric.Add(totalScrapes)
	rdsCurrentCapacityMetric.Set(rdsCurrentCapacity)
	rdsDatabaseConnectionsMetric.Set(rdsDatabaseConnections)

	ch <- durationMetric
	ch <- totalScrapesMetric
	ch <- rdsCurrentCapacityMetric
	ch <- rdsDatabaseConnectionsMetric
}
