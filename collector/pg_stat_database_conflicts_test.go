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
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartystreets/goconvey/convey"
)

var statDatabaseConflictsTestColumns = []string{
	"datid",
	"datname",
	"confl_tablespace",
	"confl_lock",
	"confl_snapshot",
	"confl_bufferpin",
	"confl_deadlock",
}

func TestPGStatDatabaseConflictsCollector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	rows := sqlmock.NewRows(statDatabaseConflictsTestColumns).
		AddRow("1", "postgres", 1, 2, 3, 4, 5).
		AddRow("2", "template1", 0, 0, 0, 0, 0)
	mock.ExpectQuery(sanitizeQuery(statDatabaseConflictsQuery)).WillReturnRows(rows)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		c := PGStatDatabaseConflictsCollector{}
		if err := c.Update(context.Background(), inst, ch); err != nil {
			t.Errorf("Error calling PGStatDatabaseConflictsCollector.Update: %s", err)
		}
	}()

	pg := labelMap{"datid": "1", "datname": "postgres"}
	t1 := labelMap{"datid": "2", "datname": "template1"}
	expected := []MetricResult{
		{labels: pg, metricType: dto.MetricType_COUNTER, value: 1},
		{labels: pg, metricType: dto.MetricType_COUNTER, value: 2},
		{labels: pg, metricType: dto.MetricType_COUNTER, value: 3},
		{labels: pg, metricType: dto.MetricType_COUNTER, value: 4},
		{labels: pg, metricType: dto.MetricType_COUNTER, value: 5},
		{labels: t1, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: t1, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: t1, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: t1, metricType: dto.MetricType_COUNTER, value: 0},
		{labels: t1, metricType: dto.MetricType_COUNTER, value: 0},
	}
	convey.Convey("Metrics comparison", t, func() {
		for _, expect := range expected {
			m := readMetric(<-ch)
			convey.So(expect, convey.ShouldResemble, m)
		}
	})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

// Metric names are public API: they must stay what the legacy exporter emitted.
func TestPGStatDatabaseConflictsMetricNames(t *testing.T) {
	for desc, name := range map[*prometheus.Desc]string{
		statDatabaseConflictsTablespaceDesc: "pg_stat_database_conflicts_confl_tablespace",
		statDatabaseConflictsLockDesc:       "pg_stat_database_conflicts_confl_lock",
		statDatabaseConflictsSnapshotDesc:   "pg_stat_database_conflicts_confl_snapshot",
		statDatabaseConflictsBufferpinDesc:  "pg_stat_database_conflicts_confl_bufferpin",
		statDatabaseConflictsDeadlockDesc:   "pg_stat_database_conflicts_confl_deadlock",
	} {
		if got := desc.String(); !strings.Contains(got, `fqName: "`+name+`"`) {
			t.Errorf("desc %s does not contain fqName %q", got, name)
		}
	}
}
