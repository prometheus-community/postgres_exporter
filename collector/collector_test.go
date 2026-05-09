// Copyright 2023 The Prometheus Authors
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
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
)

type labelMap map[string]string

type MetricResult struct {
	labels     labelMap
	value      float64
	metricType dto.MetricType
}

func readMetric(m prometheus.Metric) MetricResult {
	pb := &dto.Metric{}
	m.Write(pb)
	labels := make(labelMap, len(pb.Label))
	for _, v := range pb.Label {
		labels[v.GetName()] = v.GetValue()
	}
	if pb.Gauge != nil {
		return MetricResult{labels: labels, value: pb.GetGauge().GetValue(), metricType: dto.MetricType_GAUGE}
	}
	if pb.Counter != nil {
		return MetricResult{labels: labels, value: pb.GetCounter().GetValue(), metricType: dto.MetricType_COUNTER}
	}
	if pb.Untyped != nil {
		return MetricResult{labels: labels, value: pb.GetUntyped().GetValue(), metricType: dto.MetricType_UNTYPED}
	}
	panic("Unsupported metric type")
}

func sanitizeQuery(q string) string {
	q = strings.Join(strings.Fields(q), " ")
	q = strings.ReplaceAll(q, "(", "\\(")
	q = strings.ReplaceAll(q, "?", "\\?")
	q = strings.ReplaceAll(q, ")", "\\)")
	q = strings.ReplaceAll(q, "[", "\\[")
	q = strings.ReplaceAll(q, "]", "\\]")
	q = strings.ReplaceAll(q, "{", "\\{")
	q = strings.ReplaceAll(q, "}", "\\}")
	q = strings.ReplaceAll(q, "*", "\\*")
	q = strings.ReplaceAll(q, "^", "\\^")
	q = strings.ReplaceAll(q, "$", "\\$")
	return q
}

// We ensure that when the database respond after a long time
// The collection process still occurs in a predictable manner
// Will avoid accumulation of queries on a completely frozen DB
func TestWithConnectionTimeout(t *testing.T) {

	timeoutForQuery := time.Duration(100 * time.Millisecond)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()

	inst := &instance{db: db}

	columns := []string{"pg_roles.rolname", "pg_roles.rolconnlimit"}
	rows := sqlmock.NewRows(columns).AddRow("role1", 2)
	mock.ExpectQuery(pgRolesConnectionLimitsQuery).
		WillDelayFor(30 * time.Second).
		WillReturnRows(rows)

	log_config := promslog.Config{}

	logger := promslog.New(&log_config)

	c, err := NewPostgresCollector(logger, []string{}, "postgresql://local", []string{}, WithCollectionTimeout(timeoutForQuery.String()))
	if err != nil {
		t.Fatalf("error creating NewPostgresCollector: %s", err)
	}
	collector_config := collectorConfig{
		logger:           logger,
		excludeDatabases: []string{},
	}

	collector, err := NewPGRolesCollector(collector_config)
	if err != nil {
		t.Fatalf("error creating collector: %s", err)
	}
	c.Collectors["test"] = collector
	c.instance = inst

	ch := make(chan prometheus.Metric)
	defer close(ch)

	go func() {
		for {
			<-ch
			time.Sleep(1 * time.Millisecond)
		}
	}()

	startTime := time.Now()
	c.collectFromConnection(inst, ch)
	elapsed := time.Since(startTime)

	if elapsed <= timeoutForQuery {
		t.Errorf("elapsed time was %v, should be bigger than timeout=%v", elapsed, timeoutForQuery)
	}

	// Ensure we took more than timeout, but not too much
	if elapsed >= timeoutForQuery+500*time.Millisecond {
		t.Errorf("elapsed time was %v, should not be much bigger than timeout=%v", elapsed, timeoutForQuery)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled exceptions: %s", err)
	}
}

// TestIsAuroraUnsupportedFunction proves the helper only matches Aurora's
// specific feature_not_supported errors and lets every other error fall
// through unchanged.
func TestIsAuroraUnsupportedFunction(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		// Should match — real Aurora errors:
		{
			name: "aurora pg_last_xact_replay_timestamp",
			err:  &pq.Error{Code: "0A000", Message: "pg_last_xact_replay_timestamp() is currently not supported for Aurora"},
			want: true,
		},
		{
			name: "aurora pg_ls_waldir",
			err:  &pq.Error{Code: "0A000", Message: "pg_ls_waldir() is currently not supported for Aurora"},
			want: true,
		},
		// Should NOT match:
		{name: "nil error", err: nil, want: false},
		{name: "plain error (not pq)", err: errors.New("connection refused"), want: false},
		{name: "permission denied (42501)", err: &pq.Error{Code: "42501", Message: "permission denied for function pg_ls_waldir"}, want: false},
		{name: "undefined function (42883)", err: &pq.Error{Code: "42883", Message: "function aurora_replica_status() does not exist"}, want: false},
		{name: "syntax error (42601)", err: &pq.Error{Code: "42601", Message: "syntax error near 'Aurora'"}, want: false},
		{name: "feature_not_supported but not Aurora", err: &pq.Error{Code: "0A000", Message: "this feature is not yet implemented"}, want: false},
		{name: "connection failure (08006)", err: &pq.Error{Code: "08006", Message: "connection failure on Aurora cluster"}, want: false},
		{name: "internal error (XX000)", err: &pq.Error{Code: "XX000", Message: "internal Aurora storage error"}, want: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isAuroraUnsupportedFunction(c.err); got != c.want {
				t.Errorf("isAuroraUnsupportedFunction(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}
