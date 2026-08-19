# Configuring the Exporter

The exporter is configured through a combination of command-line flags, environment variables, and (optionally) a YAML config file. This page covers the operator-facing settings; database connection and credentials are covered separately in [Connecting to PostgreSQL](connecting.md) and [Secrets](secrets.md).

## Collectors

Metrics are grouped into collectors, each enabled or disabled independently with `--collector.<name>` / `--no-collector.<name>`. Most collectors map directly onto one of PostgreSQL's own statistics views, so the [PostgreSQL monitoring documentation](https://www.postgresql.org/docs/current/monitoring-stats.html) is the reference for what the underlying numbers mean.

The defaults are a reasonable starting point for a general-purpose server. Collectors are off by default when they're expensive, produce high-cardinality metrics, need a specific PostgreSQL version, or require an extension to be installed.

### Connections and activity

| Collector | Default | What it tells you |
|---|---|---|
| `stat_activity` | enabled | Backend counts and the longest running transaction, broken down by database, state, user, application name, backend type, and wait event. The main source of "how busy is this server and what is it waiting on". From `pg_stat_activity`. |
| `locks` | enabled | Number of locks held per database, by lock mode (`accesssharelock`, `rowexclusivelock`, …). Spot lock pile-ups. From `pg_locks`. |
| `roles` | enabled | Per-role connection limit (`rolconnlimit`). Cheap; useful to alert when a role approaches its cap. From `pg_roles`. |
| `long_running_transactions` | disabled | Count of open non-idle transactions and the age of the oldest, excluding autovacuum. A long-lived transaction blocks vacuum and holds back the xmin horizon, so this is worth turning on. |
| `process_idle` | disabled | Histogram of how long idle connections have been idle, by state and application name. Finds connection-pool leaks and `idle in transaction` sessions. |

### Database-level statistics

| Collector | Default | What it tells you |
|---|---|---|
| `stat_database` | enabled | Per-database counters: commits and rollbacks, blocks read from disk vs. found in shared buffers (cache hit ratio), tuples returned/fetched/inserted/updated/deleted, conflicts, deadlocks, temp files and bytes, and I/O timings. The workhorse collector. From `pg_stat_database`. |
| `database` | enabled | Per-database on-disk size and connection limit. Note that it calls `pg_database_size()` once per database on every scrape, which gets expensive with many or very large databases. |
| `stat_user_tables` | enabled | Per-table: sequential vs. index scans, rows inserted/updated/deleted, live and dead tuple counts, rows modified since last analyze, last vacuum/analyze times and counts, and table and index size. How you find missing indexes and tables that aren't being vacuumed. From `pg_stat_user_tables`. |
| `statio_user_tables` | enabled | Per-table block I/O: heap, index, and TOAST blocks read from disk vs. served from shared buffers. From `pg_statio_user_tables`. |
| `statio_user_indexes` | disabled | The same read/hit split per *index* rather than per table. Off by default because it emits a series per index, which adds up quickly on large schemas. From `pg_statio_user_indexes`. |

The per-table and per-index collectors scale with your schema — one set of series per table or index, per database scraped. On a database with thousands of tables this is the first place to look if the exporter's cardinality becomes a problem.

### Vacuum and wraparound

| Collector | Default | What it tells you |
|---|---|---|
| `stat_progress_vacuum` | enabled | Vacuums currently in flight: which phase they're in, heap blocks total/scanned/vacuumed, index vacuum passes, and dead tuple counts. Answers "is this vacuum making progress?". From `pg_stat_progress_vacuum`. |
| `stat_activity_autovacuum` | disabled | The start time of each running autovacuum worker, labeled by the relation it's working on. Shows autovacuum workers that have been stuck for hours. |
| `database_wraparound` | disabled | Age of `datfrozenxid` and `datminmxid` per database, in seconds. This is the early warning for transaction ID wraparound — worth enabling on any database with high write volume. |

### Replication and WAL

| Collector | Default | What it tells you |
|---|---|---|
| `replication` | enabled | Whether this instance is a replica, and if so how far behind it is replaying (in seconds). The simplest replication-lag signal. |
| `stat_replication` | enabled | Seen from the primary: for each connected standby, how many WAL bytes it is behind on sent/written/flushed/replayed, plus the primary's current WAL position. From `pg_stat_replication`. |
| `replication_slots` | enabled | Per slot: whether it's active, and how much WAL the slot is holding back. An inactive slot quietly retaining WAL is a classic way to fill a disk. From `pg_replication_slots`. |
| `stat_archiver` | enabled | WAL archiving: count of archived and failed segments, and seconds since the last successful archive. Catches a broken `archive_command` before it fills `pg_wal`. From `pg_stat_archiver`. |
| `wal` | enabled | Number of WAL segment files and their total size on disk, via `pg_ls_waldir()`. |
| `stat_wal_receiver` | disabled | Seen from a standby: the WAL receiver's LSNs, timeline, upstream node, and last message send/receipt times. Enable this on replicas for a view of the receiving side. From `pg_stat_wal_receiver`. |
| `xlog_location` | disabled | Current WAL write position (or replay position on a replica) as a byte offset. Largely superseded by `stat_replication`; retained for old setups and pre-10 servers. |

### Checkpoints, buffers, and I/O

| Collector | Default | What it tells you |
|---|---|---|
| `stat_bgwriter` | enabled | Background writer activity: buffers it wrote, `maxwritten_clean` stops, and buffers allocated. On PostgreSQL 16 and earlier it additionally reports checkpoint counts, checkpoint write/sync timings, and buffers written directly by backends. PostgreSQL 17 moved those columns out of `pg_stat_bgwriter` — the checkpoint ones are available from `stat_checkpointer`, the backend-write ones live in `pg_stat_io`, which this exporter does not currently collect. From `pg_stat_bgwriter`. |
| `stat_checkpointer` | disabled | Checkpoints and restartpoints: timed vs. requested counts, write and sync time, and buffers written. **PostgreSQL 17+ only** — on older servers the collector logs a warning and reports nothing. Enable it on 17+ to replace the checkpoint metrics that used to come from `stat_bgwriter`. From `pg_stat_checkpointer`. |
| `buffercache_summary` | disabled | Shared buffer pool usage: used, unused, dirty, and pinned buffers plus average usage count. Requires `CREATE EXTENSION pg_buffercache` and PostgreSQL 16+ — see [Database Permissions](database-permissions.md#extension-specific-collectors). |

### Server and query-level

| Collector | Default | What it tells you |
|---|---|---|
| `settings` | enabled | Every boolean, integer, and real setting in `pg_settings`, exported as a metric. Makes server configuration queryable and configuration drift alertable. |
| `postmaster` | disabled | The server process start time, as a Unix timestamp. Use it to derive uptime and to detect unplanned restarts. |
| `stat_statements` | disabled | Per-query aggregates: call counts, total execution time, rows returned, and block read/write time. The most direct way to find your slowest and most frequent queries. Requires the `pg_stat_statements` extension, and it is the highest-cardinality collector here — one series set per distinct query. From `pg_stat_statements`. |

By default `stat_statements` labels its metrics with the query ID only, not the query text. Its sub-flags control that and bound how much it emits:

| Flag | Default | Effect |
|---|---|---|
| `--collector.stat_statements.include_query` | off | Emits an extra `pg_stat_statements_query_id` metric mapping each query ID to its query text, so you can join IDs back to queries. Off by default because the text may contain literals from your queries. |
| `--collector.stat_statements.query_length` | 120 | Maximum length of the query text above, in characters. |
| `--collector.stat_statements.limit` | 100 | Maximum number of statements returned per scrape, ordered by total execution time. The main lever on cardinality. |
| `--collector.stat_statements.exclude_databases` | none | Comma-separated database names to skip. |
| `--collector.stat_statements.exclude_users` | none | Comma-separated user names to skip — useful for excluding the monitoring role's own queries. |

Collectors fail independently: if one errors — a missing extension, an unsupported version, insufficient privileges — its metrics are absent and the error is logged, but the rest of the scrape still succeeds.

Run `postgres_exporter --help` for the exact, current list of flags — new collectors are added over time and these tables can drift.

## Common flags

Flags specific to this exporter:

| Flag | Purpose |
|---|---|
| `--web.telemetry-path` | Metrics path. Default `/metrics`. |
| `--config.file` | Path to the [config file](#config-file). Default `postgres_exporter.yml`. |
| `--collection-timeout` | Per-scrape timeout. Default `1m`. See [Connecting](connecting.md#connection-timeout). |
| `--metric-prefix` | Prefix for emitted metrics. Default `pg`. |

### Flags from the Prometheus toolkit

The exporter's HTTP server is provided by [exporter-toolkit](https://github.com/prometheus/exporter-toolkit), which is shared across Prometheus exporters — so these flags behave the same way here as in `node_exporter` and friends, and the toolkit's own documentation is the authoritative reference:

| Flag | Purpose |
|---|---|
| `--web.listen-address` | Address to serve on. Default `:9187`. Repeatable, to listen on several addresses (e.g. `--web.listen-address=:9187 --web.listen-address=[::1]:9187`). |
| `--web.config.file` | Enables TLS and/or basic auth on the exporter's own HTTP server. File format: [web-configuration.md](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md). See also [Secrets](secrets.md#exporter-http-endpoint). |
| `--web.systemd-socket` | Use systemd socket activation instead of binding a port. Linux only — the flag doesn't exist on other platforms. |

Logging flags come from the shared [`prometheus/common`](https://github.com/prometheus/common) library and are likewise common to all exporters:

| Flag | Purpose |
|---|---|
| `--log.level` | Logging verbosity: `debug`, `info` (default), `warn`, or `error`. |
| `--log.format` | Log output format: `logfmt` (default) or `json`. |

Most flags also have an environment-variable equivalent (e.g. `PG_EXPORTER_WEB_TELEMETRY_PATH`, `PG_EXPORTER_COLLECTION_TIMEOUT`); a CLI flag, if set, always wins over the corresponding environment variable. The full list is in the [main README](../README.md#flags).

## Config file

The config file (`--config.file`, default `postgres_exporter.yml`) is only needed for [multi-target `/probe`](connecting.md#multi-target-mode-probe) authentication — single-target deployments generally don't need one.

```yaml
auth_modules:
  foo: # referenced as ?auth_module=foo on /probe requests
    type: userpass
    userpass:
      username: monitoring_user
      password: monitoring_password
    options:
      # merged into the DSN as key=value query parameters
      sslmode: require
```

The config file is reloaded without restarting the process by sending an HTTP POST to `/-/reload`.

## Deprecated options

These remain for backward compatibility but shouldn't be used in new deployments:

- `--auto-discover-databases`, `--exclude-databases`, `--include-databases` — dynamically discovers and scrapes every database on a server. Prefer explicitly listing DSNs.
- `--extend.query-path` — custom queries defined in a YAML file (see `queries.yaml` for the format), for metrics not covered by a built-in collector.
- `--disable-default-metrics` — turns off every built-in collector, leaving only `--extend.query-path` queries. Mainly useful for unsupported PostgreSQL variants (e.g. Greenplum) or very old versions.
- `--constantLabels` — static `label=value` pairs applied to every metric.
