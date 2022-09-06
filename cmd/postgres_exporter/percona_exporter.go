package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/blang/semver"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"path/filepath"
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
)

func initializePerconaExporters(dsn []string, opts []ExporterOpt) (func(), *Exporter, *Exporter, *Exporter) {
	queriesPath := map[MetricResolution]string{
		HR: *collectCustomQueryHrDirectory,
		MR: *collectCustomQueryMrDirectory,
		LR: *collectCustomQueryLrDirectory,
	}

	defaultOpts := []ExporterOpt{CollectorName("exporter")}
	defaultOpts = append(defaultOpts, opts...)
	defaultExporter := NewExporter(
		dsn,
		defaultOpts...,
	)
	prometheus.MustRegister(defaultExporter)

	hrExporter := NewExporter(dsn,
		CollectorName("custom_query.hr"),
		DisableDefaultMetrics(true),
		DisableSettingsMetrics(true),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithUserQueriesEnabled(map[MetricResolution]bool{
			HR: *collectCustomQueryHr,
			MR: false,
			LR: false,
		}),
		WithUserQueriesPath(queriesPath),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(*excludeDatabases),
	)
	prometheus.MustRegister(hrExporter)

	mrExporter := NewExporter(dsn,
		CollectorName("custom_query.mr"),
		DisableDefaultMetrics(true),
		DisableSettingsMetrics(true),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithUserQueriesEnabled(map[MetricResolution]bool{
			HR: false,
			MR: *collectCustomQueryMr,
			LR: false,
		}),
		WithUserQueriesPath(queriesPath),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(*excludeDatabases),
	)
	prometheus.MustRegister(mrExporter)

	lrExporter := NewExporter(dsn,
		CollectorName("custom_query.lr"),
		DisableDefaultMetrics(true),
		DisableSettingsMetrics(true),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithUserQueriesEnabled(map[MetricResolution]bool{
			HR: false,
			MR: false,
			LR: *collectCustomQueryLr,
		}),
		WithUserQueriesPath(queriesPath),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(*excludeDatabases),
	)
	prometheus.MustRegister(lrExporter)

	return func() {
		defaultExporter.servers.Close()
		hrExporter.servers.Close()
		mrExporter.servers.Close()
		lrExporter.servers.Close()
	}, hrExporter, mrExporter, lrExporter
}

func (e *Exporter) loadCustomQueries(res MetricResolution, version semver.Version, server *Server) {
	if e.userQueriesPath[res] != "" {
		fi, err := ioutil.ReadDir(e.userQueriesPath[res])
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("failed read dir %q for custom query", e.userQueriesPath[res]),
				"err", err)
			return
		}

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
	userQueriesData, err := ioutil.ReadFile(path)
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
