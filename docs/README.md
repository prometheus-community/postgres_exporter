# postgres_exporter Documentation

`postgres_exporter` is a [Prometheus](https://prometheus.io/) exporter for PostgreSQL server metrics. It connects to a PostgreSQL instance as a regular database user, queries the server's built-in statistics views on each scrape, and exposes the results as Prometheus metrics over HTTP.

It is a [Prometheus Community](https://github.com/prometheus-community) project.

## What it collects

The exporter reads PostgreSQL's own statistics and catalog views — nothing is instrumented inside your application. Metrics are grouped into [collectors](configuration.md#collectors) that can be enabled or disabled individually, covering areas such as:

- **Connections and activity** — backend counts by state, long-running transactions, idle sessions, autovacuum workers.
- **Database-level statistics** — commits, rollbacks, block reads/hits, deadlocks, temp file usage, size on disk.
- **Table and index I/O** — sequential vs. index scans, tuple counts, dead tuples, vacuum/analyze times.
- **Replication** — replica lag, replication slots, WAL receiver state, WAL generation.
- **Checkpointing and background writer** — checkpoint counts and timings, buffers written.
- **Query statistics** — per-query call counts and timings via the `pg_stat_statements` extension (opt-in).
- **Server configuration** — the values of `pg_settings`, exported as metrics so configuration drift is queryable.

All metrics are prefixed with `pg_` by default (configurable with `--metric-prefix`). The exporter's own health is exposed via `pg_up` and `pg_exporter_*` metrics alongside the standard Go runtime metrics.

Run `postgres_exporter --help` for the authoritative list of collectors and flags in the version you're running.

## Supported PostgreSQL versions

The exporter is tested in CI against PostgreSQL **13, 14, 15, 16, 17, and 18**. Older versions generally still work — the exporter detects the server version at connection time and adjusts its queries — but they are not covered by CI and issues against them are lower priority.

A few collectors are version-gated, for example `stat_checkpointer` requires PostgreSQL 17 or newer.

## Deployment modes

The exporter can be run in either of two ways:

- **Single-target (the default).** One exporter process per PostgreSQL instance, usually deployed as a sidecar, serving that instance's metrics at `/metrics`. This is the recommended setup.
- **Multi-target (`/probe`).** One exporter process scrapes many instances, with the target chosen per HTTP request. Useful when a sidecar isn't possible, such as with managed/SaaS PostgreSQL.

See [Connecting to PostgreSQL](connecting.md) for details on both.

## Documentation

**Start here**

- [Getting Started](getting-started.md) — run the exporter and scrape your first metrics.
- [Docker Images](docker.md) — where the images live and how to run them.

**Configuration and operation**

- [Configuring the Exporter](configuration.md) — collectors, flags, environment variables, and the config file.
- [Connecting to PostgreSQL](connecting.md) — DSNs, single-target vs. multi-target, timeouts.
- [Database Permissions](database-permissions.md) — the SQL grants the exporter needs.
- [Secrets](secrets.md) — ways to supply credentials without putting them on the command line.

**Environment-specific**

- [Running Against AWS RDS](aws-rds.md) — RDS-specific setup notes.

## Dashboards and alerts

The repository ships a [Prometheus mixin](../postgres_mixin/README.md) containing Grafana dashboards and alerting rules built on these metrics.

## Getting help

- Bugs and feature requests: [GitHub issues](https://github.com/prometheus-community/postgres_exporter/issues).
- Questions and discussion: the `#prometheus` channel on [CNCF Slack](https://slack.cncf.io/).
