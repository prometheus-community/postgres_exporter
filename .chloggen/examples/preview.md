## 0.x.0 / YYYY-MM-DD

### Upgrade notes

**Breaking change: `web`** ([#1428](https://github.com/prometheus-community/postgres_exporter/issues/1428))

Requests above `--web.max-requests` receive HTTP 503.
Set `--web.max-requests=0` to retain unlimited concurrency.

**Deprecation: `config`** ([#1451](https://github.com/prometheus-community/postgres_exporter/issues/1451))

These variables continue to work until v0.22.0. Update deployments before upgrading.

| Deprecated variable | Replacement |
| --- | --- |
| `PG_EXPORTER_METRIC_PREFIX` | `--metric-prefix` |

### Changes

* [CHANGE] Enforce a limit on concurrent scrapes ([#1428](https://github.com/prometheus-community/postgres_exporter/issues/1428))
* [CHANGE] Deprecate environment variables that have flag equivalents ([#1451](https://github.com/prometheus-community/postgres_exporter/issues/1451))
* [FEATURE] Add the aurora_replica_status collector for Amazon Aurora PostgreSQL ([#1311](https://github.com/prometheus-community/postgres_exporter/issues/1311))

  Enable it with `--collector.aurora_replica_status`.
* [FEATURE] settings: export whether a setting requires a server restart ([#1396](https://github.com/prometheus-community/postgres_exporter/issues/1396))

  The `pg_settings_pending_restart{name}` gauge is 1 when the setting needs a restart.
* [BUGFIX] stat_database: report errors encountered while reading query results ([#1399](https://github.com/prometheus-community/postgres_exporter/issues/1399))
