# Configuring the Exporter

The exporter is configured through a combination of command-line flags, environment variables, and (optionally) a YAML config file. This page covers the operator-facing settings; database connection and credentials are covered separately in [Connecting to PostgreSQL](connecting.md) and [Secrets](secrets.md).

## Collectors

Metrics are grouped into collectors, each enabled or disabled independently with `--collector.<name>` / `--no-collector.<name>`. Some are on by default, some are opt-in because they're expensive, version-specific, or require extra setup (like `pg_stat_statements`).

| Collector | Default | Notes |
|---|---|---|
| `database` | enabled | |
| `database_wraparound` | disabled | transaction ID wraparound risk |
| `locks` | enabled | |
| `long_running_transactions` | disabled | |
| `postmaster` | disabled | |
| `process_idle` | disabled | |
| `replication` | enabled | |
| `replication_slots` | enabled | |
| `roles` | enabled | |
| `settings` | enabled | values from `pg_settings` |
| `stat_activity` | enabled | |
| `stat_activity_autovacuum` | disabled | |
| `stat_archiver` | enabled | |
| `stat_bgwriter` | enabled | |
| `stat_checkpointer` | disabled | PostgreSQL 17+ only |
| `stat_database` | enabled | |
| `stat_progress_vacuum` | enabled | |
| `stat_replication` | enabled | |
| `stat_statements` | disabled | requires the `pg_stat_statements` extension |
| `stat_user_tables` | enabled | |
| `stat_wal_receiver` | disabled | |
| `statio_user_indexes` | disabled | |
| `statio_user_tables` | enabled | |
| `wal` | enabled | |
| `xlog_location` | disabled | |
| `buffercache_summary` | disabled | requires the `pg_buffercache` extension |

`pg_stat_statements` has its own sub-flags: `--collector.stat_statements.include_query` (off by default — includes raw query text), `--collector.stat_statements.query_length` (default 120), `--collector.stat_statements.limit` (default 100), and `--collector.stat_statements.exclude_databases` / `--collector.stat_statements.exclude_users` (comma-separated).

Run `postgres_exporter --help` for the exact, current list of flags — new collectors are added over time and this table can drift.

## Common flags

| Flag | Purpose |
|---|---|
| `--web.listen-address` | Address to serve metrics/telemetry on. Default `:9187`. |
| `--web.telemetry-path` | Metrics path. Default `/metrics`. |
| `--web.config.file` | Enables TLS and/or basic auth on the exporter's own HTTP server. See [Secrets](secrets.md#exporter-http-endpoint). |
| `--web.systemd-socket` | Use systemd socket activation instead of binding a port (Linux only). |
| `--config.file` | Path to the [config file](#config-file). Default `postgres_exporter.yml`. |
| `--collection-timeout` | Per-scrape timeout. Default `1m`. See [Connecting](connecting.md#connection-timeout). |
| `--metric-prefix` | Prefix for emitted metrics. Default `pg`. |
| `--log.level`, `--log.format` | Logging verbosity (`debug`/`info`/`warn`/`error`) and format (`logfmt`/`json`). |

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
