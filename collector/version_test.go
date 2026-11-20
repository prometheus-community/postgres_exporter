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
	"testing"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

func TestVersionCollector(t *testing.T) {
	const raw = "PostgreSQL 17.2 (Debian 17.2-1.pgdg120+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit"
	inst := &instance{versionString: raw, version: semver.MustParse("17.2.0")}

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		if err := (VersionCollector{}).Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling VersionCollector.Update: %s", err)
		}
	}()

	expected := MetricResult{
		labels:     labelMap{"version": raw, "short_version": "17.2.0"},
		value:      1,
		metricType: dto.MetricType_UNTYPED,
	}
	convey.Convey("Metrics comparison", t, func() {
		convey.So(readMetric(<-ch), convey.ShouldResemble, expected)
	})
}
