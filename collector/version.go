// Copyright The Prometheus Authors
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

package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(versionSubsystem, serverScope, NewVersionCollector)
}

type VersionCollector struct{}

func NewVersionCollector(collectorConfig) (Collector, error) {
	return &VersionCollector{}, nil
}

// The metric is pg_static, not pg_version_*, to stay compatible with the legacy exporter.
var versionDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "", "static"),
	"Version string as reported by postgres",
	[]string{"version", "short_version"}, nil,
)

func (VersionCollector) Update(_ context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	ch <- prometheus.MustNewConstMetric(versionDesc, prometheus.UntypedValue, 1,
		instance.versionString, instance.version.String())
	return nil
}
