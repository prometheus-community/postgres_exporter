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
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
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

	ctx, cancel := context.WithTimeout(context.Background(), timeoutForQuery)
	defer cancel()

	startTime := time.Now()
	c.collectFromConnection(ctx, inst, ch, nil)
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

func TestCollectFromConnectionFiltersToMatchingScope(t *testing.T) {
	logger := promslog.NewNopLogger()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening a stub db connection: %s", err)
	}
	defer db.Close()
	inst := &instance{db: db}

	c, err := NewPostgresCollector(logger, nil, "postgresql://local", nil)
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}

	var mu sync.Mutex
	ran := make(map[string]bool)
	for name := range c.Collectors {
		c.Collectors[name] = collectorFunc(func(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
			mu.Lock()
			ran[name] = true
			mu.Unlock()
			return nil
		})
	}

	ch := make(chan prometheus.Metric, 1024)
	scope := databaseScope
	c.collectFromConnection(context.Background(), inst, ch, &scope)
	close(ch)
	for range ch {
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	mu.Lock()
	defer mu.Unlock()
	for name, didRun := range ran {
		want := factories[name].scope == databaseScope
		if didRun != want {
			t.Errorf("collector %q ran = %v, want %v", name, didRun, want)
		}
	}
}

// collectorFunc adapts a function to the Collector interface.
type collectorFunc func(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error

func (f collectorFunc) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	return f(ctx, instance, ch)
}

// fakeServers stands in for PostgreSQL servers so a test can make new
// connections fail and recover between scrapes. reject decides, per DSN,
// whether the server turns the new connection away.
type fakeServers struct {
	mu        sync.Mutex
	reject    func(dsn string) bool
	databases []string
}

func (f *fakeServers) setReject(reject func(dsn string) bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reject = reject
}

func (f *fakeServers) open(dsn string) (*sql.DB, error) {
	f.mu.Lock()
	rejected := f.reject != nil && f.reject(dsn)
	databases := f.databases
	f.mu.Unlock()

	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mock.MatchExpectationsInOrder(false)
	if rejected {
		refused := errors.New("pq: sorry, too many clients already")
		mock.ExpectQuery("SELECT version").WillReturnError(refused)
		mock.ExpectQuery("SELECT current_database").WillReturnError(refused)
		return db, nil
	}
	parsed, err := config.NewDSN(dsn)
	if err != nil {
		return nil, err
	}
	mock.ExpectQuery("SELECT version").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 17.0 on x86_64-pc-linux-gnu"))
	mock.ExpectQuery("SELECT current_database").WillReturnRows(sqlmock.NewRows([]string{"current_database", "is_aurora"}).AddRow(parsed.Database(), false))
	rows := sqlmock.NewRows([]string{"datname"})
	for _, database := range databases {
		rows.AddRow(database)
	}
	mock.ExpectQuery("SELECT datname FROM pg_database").WillReturnRows(rows)
	return db, nil
}

// newConnectionTestCollector returns a PostgresCollector whose connections
// go to servers and whose only collector always succeeds.
func newConnectionTestCollector(t *testing.T, servers *fakeServers, dsn string, options ...Option) *PostgresCollector {
	t.Helper()
	c, err := NewPostgresCollector(promslog.NewNopLogger(), nil, dsn, nil, options...)
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}
	c.instance.open = servers.open
	c.Collectors = map[string]Collector{
		"stub": collectorFunc(func(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error { return nil }),
	}
	return c
}

const connectionSuccessHeader = `
# HELP pg_scrape_connection_success postgres_exporter: Whether a new connection to the target could be opened during this scrape.
# TYPE pg_scrape_connection_success gauge
`

// assertCollected scrapes c and requires the named metrics to be exactly
// want, label sets included.
func assertCollected(t *testing.T, c prometheus.Collector, want string, names ...string) {
	t.Helper()
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), names...); err != nil {
		t.Fatal(err)
	}
}

// TestPostgresCollectorReportsConnectionHealth covers
// https://github.com/prometheus-community/postgres_exporter/issues/1387:
// when a new connection cannot be made, the failure must show up as a metric
// that keeps the series identity of the success path, and not as the absence
// of the collectors' metrics or as failures of every collector.
func TestPostgresCollectorReportsConnectionHealth(t *testing.T) {
	servers := &fakeServers{}
	c := newConnectionTestCollector(t, servers, "postgresql://exporter@db.example:5432/postgres")
	const collectorSuccess = `
# HELP pg_scrape_collector_success postgres_exporter: Whether a collector succeeded.
# TYPE pg_scrape_collector_success gauge
pg_scrape_collector_success{collector="stub"} 1
`
	up := connectionSuccessHeader + `pg_scrape_connection_success{datname="postgres",server="db.example:5432"} 1` + "\n"
	down := connectionSuccessHeader + `pg_scrape_connection_success{datname="postgres",server="db.example:5432"} 0` + "\n"

	assertCollected(t, c, up, "pg_scrape_connection_success")
	assertCollected(t, c, collectorSuccess, "pg_scrape_collector_success")

	// The server starts refusing new connections, e.g. the connection limit
	// is reached or pg_hba.conf no longer matches. Only the connection
	// metric remains, and it keeps its labels.
	servers.setReject(func(string) bool { return true })
	assertCollected(t, c, down, "pg_scrape_connection_success")
	assertCollected(t, c, "", "pg_scrape_collector_success")

	servers.setReject(nil)
	assertCollected(t, c, up, "pg_scrape_connection_success")
	assertCollected(t, c, collectorSuccess, "pg_scrape_collector_success")
}

// TestPostgresCollectorReportsRefusedConnection uses a real refused TCP
// connection instead of a fake driver.
func TestPostgresCollectorReportsRefusedConnection(t *testing.T) {
	// Nothing listens on port 1, so setup() fails fast.
	dsn := "postgresql://postgres@127.0.0.1:1/postgres?sslmode=disable&connect_timeout=2"
	c, err := NewPostgresCollector(promslog.NewNopLogger(), nil, dsn, []string{databaseSubsystem})
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}
	want := connectionSuccessHeader + `pg_scrape_connection_success{datname="postgres",server="127.0.0.1:1"} 0` + "\n"
	assertCollected(t, c, want)
}

func TestPostgresCollectorConnectionHealthWithDatabaseDiscovery(t *testing.T) {
	servers := &fakeServers{databases: []string{"app", "shop"}}
	c := newConnectionTestCollector(t, servers, "postgresql://exporter@db.example:5432/postgres",
		WithDatabaseDiscovery(nil, nil, 2))
	rejectDatabases := func(names ...string) func(string) bool {
		return func(dsn string) bool {
			parsed, err := config.NewDSN(dsn)
			return err == nil && slices.Contains(names, parsed.Database())
		}
	}
	series := func(datname string, value int) string {
		return fmt.Sprintf("pg_scrape_connection_success{datname=%q,server=\"db.example:5432\"} %d\n", datname, value)
	}
	const collectorSuccess = `
# HELP pg_scrape_collector_success postgres_exporter: Whether a collector succeeded.
# TYPE pg_scrape_collector_success gauge
pg_scrape_collector_success{collector="stub",datname="postgres"} 1
`

	assertCollected(t, c, connectionSuccessHeader+series("app", 1)+series("postgres", 1)+series("shop", 1), "pg_scrape_connection_success")
	assertCollected(t, c, collectorSuccess, "pg_scrape_collector_success")

	// A discovered database that refuses connections reports 0 under its
	// own name while the others keep reporting 1.
	servers.setReject(rejectDatabases("app"))
	assertCollected(t, c, connectionSuccessHeader+series("app", 0)+series("postgres", 1)+series("shop", 1), "pg_scrape_connection_success")

	servers.setReject(nil)
	assertCollected(t, c, connectionSuccessHeader+series("app", 1)+series("postgres", 1)+series("shop", 1), "pg_scrape_connection_success")

	// With the primary connection refused the other databases cannot be
	// discovered, so only the primary reports.
	servers.setReject(rejectDatabases("postgres"))
	assertCollected(t, c, connectionSuccessHeader+series("postgres", 0), "pg_scrape_connection_success")
	assertCollected(t, c, "", "pg_scrape_collector_success")

	servers.setReject(nil)
	assertCollected(t, c, connectionSuccessHeader+series("app", 1)+series("postgres", 1)+series("shop", 1), "pg_scrape_connection_success")
}

func TestPostgresCollectorConnectionHealthForAdditionalTargets(t *testing.T) {
	servers := &fakeServers{}
	c := newConnectionTestCollector(t, servers, "postgresql://exporter@db.example:5432/postgres",
		WithAdditionalConnectionTargets([]string{
			"postgresql://exporter@replica.example/postgres",
			"postgresql://exporter:one@other.example:5433/appdb",
			"postgresql://exporter:two@other.example:5433/appdb",
		}))
	series := func(server, datname string, value int) string {
		return fmt.Sprintf("pg_scrape_connection_success{datname=%q,server=%q} %d\n", datname, server, value)
	}
	want := func(primary, replica, other int) string {
		return connectionSuccessHeader +
			series("db.example:5432", "postgres", primary) +
			series("other.example:5433", "appdb", other) +
			series("replica.example:5432", "postgres", replica)
	}

	assertCollected(t, c, want(1, 1, 1), "pg_scrape_connection_success")

	servers.setReject(func(dsn string) bool { return strings.Contains(dsn, "replica.example") })
	assertCollected(t, c, want(1, 0, 1), "pg_scrape_connection_success")

	// Two DSNs for the same server and database share a series, which is 0
	// as soon as one of them fails.
	servers.setReject(func(dsn string) bool { return strings.Contains(dsn, "two") })
	assertCollected(t, c, want(1, 1, 0), "pg_scrape_connection_success")

	servers.setReject(func(string) bool { return true })
	assertCollected(t, c, want(0, 0, 0), "pg_scrape_connection_success")

	servers.setReject(nil)
	assertCollected(t, c, want(1, 1, 1), "pg_scrape_connection_success")
}

// A DSN that cannot be parsed still reports its failure, with empty labels,
// and the metric does not leak the password.
func TestPostgresCollectorReportsUnparseableDSN(t *testing.T) {
	c, err := NewPostgresCollector(promslog.NewNopLogger(), nil, "postgresql://demo:bad%zzpw@127.0.0.1:1/demo?sslmode=disable", nil)
	if err != nil {
		t.Fatalf("NewPostgresCollector() error = %v", err)
	}
	want := connectionSuccessHeader + `pg_scrape_connection_success{datname="",server=""} 0` + "\n"
	assertCollected(t, c, want)
}

func TestRegisterCollectorRejectsUnknownConfig(t *testing.T) {
	const name = "not_in_default_config"

	defer func() {
		if recover() == nil {
			t.Fatalf("registerCollector(%q) did not panic", name)
		}
	}()

	registerCollector(name, serverScope, func(collectorConfig) (Collector, error) {
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

type fakeCollector struct {
	err error
}

func (f fakeCollector) Update(ctx context.Context, inst *instance, ch chan<- prometheus.Metric) error {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("fake_metric", "help", nil, nil), prometheus.GaugeValue, 1)
	return f.err
}

func TestExecuteDropsMetricsOnError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{"success", nil, []string{"fake_metric", "pg_scrape_collector_duration_seconds", "pg_scrape_collector_success"}},
		{"error", errors.New("boom"), []string{"pg_scrape_collector_duration_seconds", "pg_scrape_collector_success"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan prometheus.Metric, 10)
			execute(context.Background(), "fake", fakeCollector{err: tt.err}, nil, ch, promslog.NewNopLogger(), false)
			close(ch)
			var got []string
			success := -1.0
			for m := range ch {
				name := m.Desc().String()
				for _, w := range []string{"pg_scrape_collector_duration_seconds", "pg_scrape_collector_success", "fake_metric"} {
					if strings.Contains(name, `"`+w+`"`) {
						got = append(got, w)
						if w == "pg_scrape_collector_success" {
							success = readMetric(m).value
						}
					}
				}
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got metrics %v, want %v", got, tt.want)
			}
			if (tt.err == nil) != (success == 1) {
				t.Fatalf("unexpected success value %v", success)
			}
		})
	}
}
