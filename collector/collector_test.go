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
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus-community/postgres_exporter/config"
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

func TestPostgresCollectorWrapLargeCounters(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    bool
	}{
		{name: "enabled by default", want: true},
		{name: "disabled by option", options: []Option{WithWrapLargeCounters(false)}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collector, err := NewPostgresCollector(
				promslog.NewNopLogger(),
				nil,
				"postgresql://local",
				nil,
				test.options...,
			)
			if err != nil {
				t.Fatalf("creating collector: %v", err)
			}

			if got := collector.instance.wrapLargeCounters; got != test.want {
				t.Fatalf("instance wrapLargeCounters = %t, want %t", got, test.want)
			}
			if got := collector.instance.copy().wrapLargeCounters; got != test.want {
				t.Fatalf("copied instance wrapLargeCounters = %t, want %t", got, test.want)
			}
		})
	}
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

func TestNewPostgresCollectorCreatesIndependentCollectorInstances(t *testing.T) {
	logger := promslog.NewNopLogger()
	dsn := "postgresql://local"

	first, err := NewPostgresCollector(logger, []string{"first"}, dsn, []string{})
	if err != nil {
		t.Fatalf("NewPostgresCollector() first error = %v", err)
	}
	second, err := NewPostgresCollector(logger, []string{"second"}, dsn, []string{})
	if err != nil {
		t.Fatalf("NewPostgresCollector() second error = %v", err)
	}

	firstDatabase, ok := first.Collectors[databaseSubsystem].(*PGDatabaseCollector)
	if !ok {
		t.Fatalf("first database collector type = %T, want *PGDatabaseCollector", first.Collectors[databaseSubsystem])
	}
	secondDatabase, ok := second.Collectors[databaseSubsystem].(*PGDatabaseCollector)
	if !ok {
		t.Fatalf("second database collector type = %T, want *PGDatabaseCollector", second.Collectors[databaseSubsystem])
	}
	if firstDatabase == secondDatabase {
		t.Fatal("database collector instances are shared, want independent instances")
	}
	if got, want := firstDatabase.excludedDatabases[0], "first"; got != want {
		t.Fatalf("first excluded database = %q, want %q", got, want)
	}
	if got, want := secondDatabase.excludedDatabases[0], "second"; got != want {
		t.Fatalf("second excluded database = %q, want %q", got, want)
	}
}

func TestNewPostgresCollectorUsesCollectorStateOverrides(t *testing.T) {
	logger := promslog.NewNopLogger()
	dsn := "postgresql://local"

	c, err := NewPostgresCollector(
		logger,
		nil,
		dsn,
		nil,
		WithCollectorStates(map[string]bool{databaseSubsystem: false}),
	)
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}
	if _, ok := c.Collectors[databaseSubsystem]; ok {
		t.Fatal("database collector is enabled, want disabled")
	}
}

// TestPostgresCollectorReportsFailureOnConnectionError guards against
// https://github.com/prometheus-community/postgres_exporter/issues/1387:
// when the per-scrape connection can't be established, Collect must still
// report pg_scrape_collector_success{collector=...} 0 for every enabled
// collector, rather than silently emitting no metrics at all.
func TestPostgresCollectorReportsFailureOnConnectionError(t *testing.T) {
	logger := promslog.NewNopLogger()
	// Nothing listens on this port, so inst.setup() fails fast with a
	// connection error without depending on a real unreachable network dial.
	dsn := "postgresql://postgres@127.0.0.1:1/postgres?sslmode=disable&connect_timeout=2"

	c, err := NewPostgresCollector(logger, nil, dsn, []string{databaseSubsystem})
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}

	ch := make(chan prometheus.Metric, len(c.Collectors)*2)
	c.Collect(ch)
	close(ch)

	var sawSuccessZero, sawDurationZero bool
	count := 0
	for m := range ch {
		count++
		result := readMetric(m)
		if result.labels["collector"] != databaseSubsystem {
			t.Fatalf("unexpected collector label %q", result.labels["collector"])
		}
		if result.value != 0 {
			t.Fatalf("metric value = %v, want 0 (connection never succeeded)", result.value)
		}
		switch m.Desc().String() {
		case scrapeSuccessDesc.String():
			sawSuccessZero = true
		case scrapeDurationDesc.String():
			sawDurationZero = true
		default:
			t.Fatalf("unexpected metric desc %v", m.Desc())
		}
	}

	if count != 2 {
		t.Fatalf("got %d metrics, want 2 (duration+success for the single enabled collector)", count)
	}
	if !sawSuccessZero {
		t.Fatal("pg_scrape_collector_success was not reported on connection failure")
	}
	if !sawDurationZero {
		t.Fatal("pg_scrape_collector_duration_seconds was not reported on connection failure")
	}
}

func TestRegisterCollectorRejectsUnknownConfig(t *testing.T) {
	const name = "not_in_default_config"

	defer func() {
		if recover() == nil {
			t.Fatalf("registerCollector(%q) did not panic", name)
		}
	}()

	registerCollector(name, func(collectorConfig) (Collector, error) {
		return nil, nil
	})
}

func TestNewPGStatStatementsCollectorUsesConfig(t *testing.T) {
	logger := promslog.NewNopLogger()

	collector, err := NewPGStatStatementsCollector(collectorConfig{
		logger: logger,
		pgStatStatementsConfig: config.PGStatStatementsConfig{
			IncludeQuery:     true,
			QueryLength:      42,
			Limit:            7,
			ExcludeDatabases: []string{"postgres"},
			ExcludeUsers:     []string{"monitor"},
		},
	})
	if err != nil {
		t.Fatalf("NewPGStatStatementsCollector() error = %v", err)
	}
	got, ok := collector.(*PGStatStatementsCollector)
	if !ok {
		t.Fatalf("collector type = %T, want *PGStatStatementsCollector", collector)
	}
	if !got.includeQueryStatement {
		t.Fatal("includeQueryStatement = false, want true")
	}
	if got.statementLength != 42 {
		t.Fatalf("statementLength = %d, want 42", got.statementLength)
	}
	if got.statementLimit != 7 {
		t.Fatalf("statementLimit = %d, want 7", got.statementLimit)
	}
	if got.excludedDatabases[0] != "postgres" {
		t.Fatalf("excludedDatabases[0] = %q, want postgres", got.excludedDatabases[0])
	}
	if got.excludedUsers[0] != "monitor" {
		t.Fatalf("excludedUsers[0] = %q, want monitor", got.excludedUsers[0])
	}
}
