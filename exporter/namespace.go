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

package exporter

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log/level"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

// Query within a namespace mapping and emit metrics. Returns fatal errors if
// the scrape fails, and a slice of errors if they were non-fatal.
func queryNamespaceMapping(server *Server, namespace string, mapping MetricMapNamespace) ([]prometheus.Metric, []error, error) {
	// Check for a query override for this namespace
	query, found := server.queryOverrides[namespace]

	// Was this query disabled (i.e. nothing sensible can be queried on cu
	// version of PostgreSQL?
	if query == "" && found {
		// Return success (no pertinent data)
		return []prometheus.Metric{}, []error{}, nil
	}

	// Don't fail on a bad scrape of one metric
	var rows *sql.Rows
	var err error

	if !found {
		// I've no idea how to avoid this properly at the moment, but this is
		// an admin tool so you're not injecting SQL right?
		rows, err = server.db.Query(fmt.Sprintf("SELECT * FROM %s;", namespace)) // nolint: gas
	} else {
		rows, err = server.db.Query(query)
	}
	if err != nil {
		return []prometheus.Metric{}, []error{}, fmt.Errorf("Error running query on database %q: %s %v", server, namespace, err)
	}
	defer rows.Close() // nolint: errcheck

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
			labels[idx], _ = dbToString(columnData[columnIdx[label]])
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
					sum, ok := dbToFloat64(columnData[idx], server.Logger)
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Unexpected error parsing column: ", namespace, columnName+"_sum", columnData[idx])))
						continue
					}

					idx, ok = columnIdx[columnName+"_count"]
					if !ok {
						nonfatalErrors = append(nonfatalErrors, errors.New(fmt.Sprintln("Missing column: ", namespace, columnName+"_count")))
						continue
					}
					count, ok := dbToUint64(columnData[idx], server.Logger)
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
					value, ok := dbToFloat64(columnData[idx], server.Logger)
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
				desc := prometheus.NewDesc(metricLabel, fmt.Sprintf("Unknown metric from %s", namespace), mapping.labels, server.labels)

				// Its not an error to fail here, since the values are
				// unexpected anyway.
				value, ok := dbToFloat64(columnData[idx], server.Logger)
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
func queryNamespaceMappings(ch chan<- prometheus.Metric, server *Server) map[string]error {
	// Return a map of namespace -> errors
	namespaceErrors := make(map[string]error)

	scrapeStart := time.Now()

	for namespace, mapping := range server.metricMap {
		level.Debug(server.Logger).Log("msg", "Querying namespace", "namespace", namespace)

		if mapping.master && !server.master {
			level.Debug(server.Logger).Log("msg", "Query skipped...")
			continue
		}

		// check if the query is to be run on specific database server version range or not
		if len(server.runonserver) > 0 {
			serVersion, _ := semver.Parse(server.lastMapVersion.String())
			runServerRange, _ := semver.ParseRange(server.runonserver)
			if !runServerRange(serVersion) {
				level.Debug(server.Logger).Log("msg", "Query skipped for this database version", "version", server.lastMapVersion.String(), "target_version", server.runonserver)
				continue
			}
		}

		scrapeMetric := false
		// Check if the metric is cached
		server.cacheMtx.Lock()
		cachedMetric, found := server.metricCache[namespace]
		server.cacheMtx.Unlock()
		// If found, check if needs refresh from cache
		if found {
			if scrapeStart.Sub(cachedMetric.lastScrape).Seconds() > float64(mapping.cacheSeconds) {
				scrapeMetric = true
			}
		} else {
			scrapeMetric = true
		}

		var metrics []prometheus.Metric
		var nonFatalErrors []error
		var err error
		if scrapeMetric {
			metrics, nonFatalErrors, err = queryNamespaceMapping(server, namespace, mapping)
		} else {
			metrics = cachedMetric.metrics
		}

		// Serious error - a namespace disappeared
		if err != nil {
			namespaceErrors[namespace] = err
			level.Warn(server.Logger).Log("err", err)
		}
		// Non-serious errors - likely version or parsing problems.
		if len(nonFatalErrors) > 0 {
			for _, err := range nonFatalErrors {
				level.Info(server.Logger).Log("err", err)
			}
		}

		// Emit the metrics into the channel
		for _, metric := range metrics {
			ch <- metric
		}

		if scrapeMetric {
			// Only cache if metric is meaningfully cacheable
			if mapping.cacheSeconds > 0 {
				server.cacheMtx.Lock()
				server.metricCache[namespace] = cachedMetrics{
					metrics:    metrics,
					lastScrape: scrapeStart,
				}
				server.cacheMtx.Unlock()
			}
		}
	}

	return namespaceErrors
}
