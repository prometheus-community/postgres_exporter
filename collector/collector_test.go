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

func TestNewPGStatStatementsCollectorUsesConfig(t *testing.T) {
	logger := promslog.NewNopLogger()

	collector, err := NewPGStatStatementsCollector(collectorConfig{
		logger: logger,
		pgStatStatementsConfig: PGStatStatementsConfig{
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
