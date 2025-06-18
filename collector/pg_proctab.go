package collector

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const proctabSubsystem = "proctab"

func init() {
	registerCollector(proctabSubsystem, defaultDisabled, NewPGProctabCollector)
}

type PGProctabCollector struct {
	log *slog.Logger
}

func NewPGProctabCollector(config collectorConfig) (Collector, error) {
	return &PGProctabCollector{
		log: config.logger,
	}, nil
}

var (
	pgMemusedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"memused",
		),
		"used memory (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)

	pgMemfreeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"memfree",
		),
		"free memory (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)

	pgMemsharedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"memshared",
		),
		"shared memory (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)

	pgMembuffersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"membuffers",
		),
		"buffered memory (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)

	pgMemcachedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"memcached",
		),
		"cached memory (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)
	pgSwapusedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"swapused",
		),
		"swap used (from /proc/meminfo) in bytes",
		[]string{}, nil,
	)

	// Loadavg metrics
	pgLoad1Desc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"load1",
		),
		"load1 load Average",
		[]string{}, nil,
	)

	// CPU metrics
	pgCpuUserDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"cpu_user",
		),
		"PG User cpu time",
		[]string{}, nil,
	)
	pgCpuNiceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"cpu_nice",
		),
		"PG nice cpu time (running queries)",
		[]string{}, nil,
	)
	pgCpuSystemDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"cpu_system",
		),
		"PG system cpu time",
		[]string{}, nil,
	)
	pgCpuIdleDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"cpu_idle",
		),
		"PG idle cpu time",
		[]string{}, nil,
	)
	pgCpuIowaitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			proctabSubsystem,
			"cpu_iowait",
		),
		"PG iowait time",
		[]string{}, nil,
	)

	memoryQuery = `
   select
	   memused,
	   memfree,
	   memshared,
	   membuffers,
	   memcached,
	   swapused
	 from
	 pg_memusage()
	`

	load1Query = `
   select
	   load1
	 from
	   pg_loadavg()
	`
	cpuQuery = `
   select
	   "user",
		 nice,
		 system,
		 idle,
		 iowait
	from
	  pg_cputime()
	`
)

func emitMemMetric(m sql.NullInt64, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	mM := 0.0
	if m.Valid {
		mM = float64(m.Int64 * 1024)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, mM)
}
func emitCpuMetric(m sql.NullInt64, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	mM := 0.0
	if m.Valid {
		mM = float64(m.Int64)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, mM)
}

// Update implements Collector and exposes database locks.
// It is called by the Prometheus registry when collecting metrics.
func (c PGProctabCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx, memoryQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var memused, memfree, memshared, membuffers, memcached, swapused sql.NullInt64

	for rows.Next() {
		if err := rows.Scan(&memused, &memfree, &memshared, &membuffers, &memcached, &swapused); err != nil {
			return err
		}
		emitMemMetric(memused, pgMemusedDesc, ch)
		emitMemMetric(memfree, pgMemfreeDesc, ch)
		emitMemMetric(memshared, pgMemsharedDesc, ch)
		emitMemMetric(membuffers, pgMembuffersDesc, ch)
		emitMemMetric(memcached, pgMemcachedDesc, ch)
		emitMemMetric(swapused, pgSwapusedDesc, ch)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = db.QueryContext(ctx, load1Query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var load1 sql.NullFloat64
	for rows.Next() {
		if err := rows.Scan(&load1); err != nil {
			return err
		}
		load1Metric := 0.0
		if load1.Valid {
			load1Metric = load1.Float64
		}
		ch <- prometheus.MustNewConstMetric(
			pgLoad1Desc,
			prometheus.GaugeValue, load1Metric,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = db.QueryContext(ctx, cpuQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	var user, nice, system, idle, iowait sql.NullInt64
	for rows.Next() {
		if err := rows.Scan(&user, &nice, &system, &idle, &iowait); err != nil {
			return err
		}
		emitCpuMetric(user, pgCpuUserDesc, ch)
		emitCpuMetric(nice, pgCpuNiceDesc, ch)
		emitCpuMetric(system, pgCpuSystemDesc, ch)
		emitCpuMetric(idle, pgCpuIdleDesc, ch)
		emitCpuMetric(iowait, pgCpuIowaitDesc, ch)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil

}
