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
	"context"
	"database/sql"
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

// TestIsNoDataError pins the trivial identity of ErrNoData. The wrapper
// exists so callers do not depend on the package-private sentinel value
// directly; if that contract drifts, this test fails immediately.
func TestIsNoDataError(t *testing.T) {
	if !IsNoDataError(ErrNoData) {
		t.Error("IsNoDataError(ErrNoData) must be true")
	}
	if IsNoDataError(nil) {
		t.Error("IsNoDataError(nil) must be false")
	}
	if IsNoDataError(errFnDoesNotExist) {
		t.Error("IsNoDataError must reject unrelated errors")
	}
}

// TestInt32 covers the small helper that drops sql.NullInt32 values to
// float64 with a NaN-style zero default. The two interesting cases are
// "valid" (returns the value) and "invalid" (returns 0).
func TestInt32(t *testing.T) {
	if got := Int32(sql.NullInt32{Int32: 42, Valid: true}); got != 42 {
		t.Errorf("Int32(valid=42) = %v, want 42", got)
	}
	if got := Int32(sql.NullInt32{Valid: false}); got != 0 {
		t.Errorf("Int32(invalid) = %v, want 0", got)
	}
}

// TestExecuteSuccessfulCollectorEmitsSuccessOne and the companion
// "failed" / "no data" tests verify the execute() wrapper records the
// right pg_scrape_collector_success value depending on what the
// Collector.Update returns.
func TestExecuteSuccessfulCollectorEmitsSuccessOne(t *testing.T) {
	got := runExecuteAndReadSuccess(t, func(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error {
		return nil
	})
	if got != 1 {
		t.Errorf("scrape_collector_success = %v, want 1", got)
	}
}

func TestExecuteFailedCollectorEmitsSuccessZero(t *testing.T) {
	got := runExecuteAndReadSuccess(t, func(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error {
		return errFnDoesNotExist
	})
	if got != 0 {
		t.Errorf("scrape_collector_success = %v, want 0", got)
	}
}

func TestExecuteNoDataCollectorEmitsSuccessZero(t *testing.T) {
	got := runExecuteAndReadSuccess(t, func(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error {
		return ErrNoData
	})
	if got != 0 {
		t.Errorf("scrape_collector_success = %v, want 0", got)
	}
}

// runExecuteAndReadSuccess wraps the Collector under test in a tiny
// adapter, drains the duration metric, and returns the success metric so
// the caller can assert against it.
func runExecuteAndReadSuccess(t *testing.T, update func(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error) float64 {
	t.Helper()
	ch := make(chan prometheus.Metric, 4)
	c := updateFn(update)
	execute(context.Background(), "fake", c, &instance{}, ch, promslog.NewNopLogger())
	close(ch)

	var success float64 = -1
	for m := range ch {
		mr := readMetric(m)
		// Two metrics emitted: duration_seconds and success. We only care
		// about success here.
		desc := m.Desc().String()
		if strings.Contains(desc, "collector_success") {
			success = mr.value
		}
	}
	if success == -1 {
		t.Fatal("scrape_collector_success metric was not emitted")
	}
	return success
}

// updateFn lets a test inline its Update implementation without declaring
// a struct per case.
type updateFn func(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error

func (f updateFn) Update(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error {
	return f(ctx, inst, ch)
}

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
