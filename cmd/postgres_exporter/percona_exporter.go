package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/semaphore"
)

type MetricResolution string

const (
	LR MetricResolution = "lr"
	MR MetricResolution = "mr"
	HR MetricResolution = "hr"
)

var (
	collectCustomQueryLr          = kingpin.Flag("collect.custom_query.lr", "Enable custom queries with low resolution directory.").Default("false").Envar("PG_EXPORTER_EXTEND_QUERY_LR").Bool()
	collectCustomQueryMr          = kingpin.Flag("collect.custom_query.mr", "Enable custom queries with medium resolution directory.").Default("false").Envar("PG_EXPORTER_EXTEND_QUERY_MR").Bool()
	collectCustomQueryHr          = kingpin.Flag("collect.custom_query.hr", "Enable custom queries with high resolution directory.").Default("false").Envar("PG_EXPORTER_EXTEND_QUERY_HR").Bool()
	collectCustomQueryLrDirectory = kingpin.Flag("collect.custom_query.lr.directory", "Path to custom queries with low resolution directory.").Envar("PG_EXPORTER_EXTEND_QUERY_LR_PATH").String()
	collectCustomQueryMrDirectory = kingpin.Flag("collect.custom_query.mr.directory", "Path to custom queries with medium resolution directory.").Envar("PG_EXPORTER_EXTEND_QUERY_MR_PATH").String()
	collectCustomQueryHrDirectory = kingpin.Flag("collect.custom_query.hr.directory", "Path to custom queries with high resolution directory.").Envar("PG_EXPORTER_EXTEND_QUERY_HR_PATH").String()

	maxConnections = kingpin.Flag("max-connections", "Maximum number of connections to use").Default("5").Envar("PG_EXPORTER_MAX_CONNECTIONS").Int64()
)

// Handler returns a http.Handler that serves metrics. Can be used instead of
// run for hooking up custom HTTP servers.
func Handler(logger log.Logger, dsns []string, connSema *semaphore.Weighted, globalCollectors map[string]prometheus.Collector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seconds, err := strconv.Atoi(r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"))
		// To support also older ones vmagents.
		if err != nil {
			seconds = 10
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
		defer cancel()

		filters := r.URL.Query()["collect[]"]
		level.Debug(logger).Log("msg", "Collect query", "filters", filters)

		var f Filters
		if len(filters) == 0 {
			f.EnableAllCollectors = true
		} else {
			for _, filter := range filters {
				switch filter {
				case "standard.process":
					f.EnableProcessCollector = true
				case "standard.go":
					f.EnableGoCollector = true
				case "version":
					f.EnableVersionCollector = true
				case "exporter":
					f.EnableDefaultCollector = true
				case "custom_query.hr":
					f.EnableHRCollector = true
				case "custom_query.mr":
					f.EnableMRCollector = true
				case "custom_query.lr":
					f.EnableLRCollector = true
				case "postgres":
					f.EnablePostgresCollector = true
				}
			}
		}

		registry := makeRegistry(ctx, dsns, connSema, globalCollectors, f)

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			ErrorLog:      stdlog.New(log.NewStdlibAdapter(logger), "handler", 0),
		})

		h.ServeHTTP(w, r)
	})
}

// Filters is a struct to enable or disable collectors.
type Filters struct {
	EnableAllCollectors     bool
	EnableLRCollector       bool
	EnableMRCollector       bool
	EnableHRCollector       bool
	EnableDefaultCollector  bool
	EnableGoCollector       bool
	EnableVersionCollector  bool
	EnableProcessCollector  bool
	EnablePostgresCollector bool
}

// makeRegistry creates a new prometheus registry with default and percona exporters.
func makeRegistry(ctx context.Context, dsns []string, connSema *semaphore.Weighted, globalCollectors map[string]prometheus.Collector, filters Filters) *prometheus.Registry {
	registry := prometheus.NewRegistry()

	excludedDatabases := strings.Split(*excludeDatabases, ",")
	logger.Log("msg", "Excluded databases", "databases", fmt.Sprintf("%v", excludedDatabases))

	queriesPath := map[MetricResolution]string{
		HR: *collectCustomQueryHrDirectory,
		MR: *collectCustomQueryMrDirectory,
		LR: *collectCustomQueryLrDirectory,
	}

	opts := []ExporterOpt{
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		ExcludeDatabases(excludedDatabases),
		WithConnectionsSemaphore(connSema),
		WithContext(ctx),
	}

	if filters.EnableAllCollectors || filters.EnableDefaultCollector {
		defaultExporter := NewExporter(dsns, append(
			opts,
			CollectorName("exporter"),
			WithConstantLabels(*constantLabelsList), // This option depends on collectors name, so keep it after CollectorName option
			DisableDefaultMetrics(*disableDefaultMetrics),
			DisableSettingsMetrics(*disableSettingsMetrics),
			IncludeDatabases(*includeDatabases),
		)...)
		registry.MustRegister(defaultExporter)
	}

	if filters.EnableAllCollectors || filters.EnableHRCollector {
		hrExporter := NewExporter(dsns,
			append(opts,
				CollectorName("custom_query.hr"),
				WithConstantLabels(*constantLabelsList), // This option depends on collectors name, so keep it after CollectorName option
				WithUserQueriesEnabled(HR),
				WithEnabled(*collectCustomQueryHr),
				DisableDefaultMetrics(true),
				DisableSettingsMetrics(true),
				WithUserQueriesPath(queriesPath),
			)...)
		registry.MustRegister(hrExporter)

	}

	if filters.EnableAllCollectors || filters.EnableMRCollector {
		mrExporter := NewExporter(dsns,
			append(opts,
				CollectorName("custom_query.mr"),
				WithConstantLabels(*constantLabelsList), // This option depends on collectors name, so keep it after CollectorName option
				WithUserQueriesEnabled(MR),
				WithEnabled(*collectCustomQueryMr),
				DisableDefaultMetrics(true),
				DisableSettingsMetrics(true),
				WithUserQueriesPath(queriesPath),
			)...)
		registry.MustRegister(mrExporter)
	}

	if filters.EnableAllCollectors || filters.EnableLRCollector {
		lrExporter := NewExporter(dsns,
			append(opts,
				CollectorName("custom_query.lr"),
				WithConstantLabels(*constantLabelsList), // This option depends on collectors name, so keep it after CollectorName option
				WithUserQueriesEnabled(LR),
				WithEnabled(*collectCustomQueryLr),
				DisableDefaultMetrics(true),
				DisableSettingsMetrics(true),
				WithUserQueriesPath(queriesPath),
			)...)
		registry.MustRegister(lrExporter)
	}

	if filters.EnableAllCollectors || filters.EnableGoCollector {
		registry.MustRegister(globalCollectors["standard.go"])
	}

	if filters.EnableAllCollectors || filters.EnableProcessCollector {
		registry.MustRegister(globalCollectors["standard.process"])
	}

	if filters.EnableAllCollectors || filters.EnableVersionCollector {
		registry.MustRegister(globalCollectors["version"])
	}

	if filters.EnableAllCollectors || filters.EnablePostgresCollector {
		// This chunk moved here from main.go
		// TODO(@sysadmind): Remove this with multi-target support. We are removing multiple DSN support
		dsn := ""
		if len(dsns) > 0 {
			dsn = dsns[0]
		}

		pe, err := collector.NewPostgresCollector(
			logger,
			excludedDatabases,
			dsn,
			[]string{},
			collector.WithContext(ctx),
			collector.WithConnectionsSemaphore(connSema),
		)
		if err != nil {
			level.Error(logger).Log("msg", "Failed to create PostgresCollector", "err", err.Error())
		} else {
			registry.MustRegister(pe)
		}
	}

	return registry
}

func (e *Exporter) loadCustomQueries(res MetricResolution, version semver.Version, server *Server) {
	if e.userQueriesPath[res] != "" {
		fi, err := os.ReadDir(e.userQueriesPath[res])
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("failed read dir %q for custom query", e.userQueriesPath[res]),
				"err", err)
			return
		}
		level.Debug(logger).Log("msg", fmt.Sprintf("reading dir %q for custom query", e.userQueriesPath[res]))

		for _, v := range fi {
			if v.IsDir() {
				continue
			}

			if filepath.Ext(v.Name()) == ".yml" || filepath.Ext(v.Name()) == ".yaml" {
				path := filepath.Join(e.userQueriesPath[res], v.Name())
				e.addCustomQueriesFromFile(path, version, server)
			}
		}
	}
}

func (e *Exporter) addCustomQueriesFromFile(path string, version semver.Version, server *Server) {
	// Calculate the hashsum of the useQueries
	userQueriesData, err := os.ReadFile(path)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to reload user queries:"+path, "err", err)
		e.userQueriesError.WithLabelValues(path, "").Set(1)
		return
	}

	hashsumStr := fmt.Sprintf("%x", sha256.Sum256(userQueriesData))

	if err := addQueries(userQueriesData, version, server); err != nil {
		level.Error(logger).Log("msg", "Failed to reload user queries:"+path, "err", err)
		e.userQueriesError.WithLabelValues(path, hashsumStr).Set(1)
		return
	}

	// Mark user queries as successfully loaded
	e.userQueriesError.WithLabelValues(path, hashsumStr).Set(0)
}
