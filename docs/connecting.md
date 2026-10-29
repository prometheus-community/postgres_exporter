---
title: Connecting to PostgreSQL
sort_rank: 3
---
# Connecting to PostgreSQL

## How the exporter connects

The exporter connects using a standard PostgreSQL [connection string / DSN](https://www.postgresql.org/docs/current/libpq-connstring.html), e.g. `postgresql://user:pass@host:5432/dbname?sslmode=disable`. Any parameter supported by [github.com/lib/pq](https://github.com/lib/pq) can be included in the DSN's query string (`sslmode`, `connect_timeout`, `application_name`, etc.).

Each scrape opens a connection using that DSN, runs the enabled collectors' queries, and closes it — the exporter keeps at most one open connection per configured target at a time. On each successful connection to a target, the exporter also detects the PostgreSQL server version, which determines which version-specific queries collectors use.

## Single-target mode (the default)

In single-target mode, the exporter scrapes one DSN configured at startup (via the `--datasource.*` flags or `DATA_SOURCE_NAME` — see [Secrets](secrets.md)) and always serves their metrics at `/metrics`. This is the typical deployment: one exporter process per database instance, usually as a sidecar.

## Multi-target mode (`/probe`)

For cases where running a sidecar isn't possible (for example, a SaaS-managed PostgreSQL instance) or not preferred, the exporter supports the [multi-target pattern](https://prometheus.io/docs/guides/multi-target-exporter/): a single exporter process can scrape many different servers, chosen per-request.

Send the target as a query parameter:

```
http://exporter:9187/probe?target=my-db-host:5432
```

Prometheus is typically configured to do this via relabeling so the target list can come from service discovery:

```yaml
scrape_configs:
  - job_name: 'postgres'
    static_configs:
      - targets:
        - server1:5432
        - server2:5432
    metrics_path: /probe
    params:
      auth_module: [foo]
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:9187 # the exporter's real host:port
```

Credentials for `/probe` targets come from an `auth_module` defined in the [config file](configuration.md#config-file), referenced via the `auth_module` query/HTTP parameter — this avoids putting usernames and passwords in the target URL.

## Detecting connection problems

On every scrape the exporter reports `pg_scrape_connection_success{server, datname}`: `1` if it could open a new connection to the target and run its initial queries, `0` if not. A connection limit that is reached, a `pg_hba.conf` change, rotated or revoked credentials, and TLS failures all give `0`. The series is emitted whether or not the connection succeeds and keeps the same labels, which are taken from the DSN (`server` is `host:port`, `datname` is the database it names) rather than from the connection.

It is reported for every configured DSN, for each database found by `--auto-discover-databases` (as its own `datname`), and for every `/probe` request. Collectors only run against the first configured DSN; the others get the connection check only. If the connection to the first DSN fails, the other databases of that server cannot be discovered, so only the first one reports `0`.

This is a different question from `pg_scrape_collector_success`, which says whether one collector's query worked, and a failed connection does not set it to `0` for every collector. `pg_up` is reported by the older long-lived connection and can stay `1` while new connections are refused. Alert on `pg_scrape_connection_success == 0` to find out that the exporter cannot connect.

## Connection timeout

`PG_EXPORTER_COLLECTION_TIMEOUT` (default `1m`) bounds how long a scrape is allowed to take before the exporter drops the connection. This protects against connections piling up if the database is slow to respond (e.g. during a large DDL operation holding locks). Values less than `1ms` are invalid.

## Running against non-superusers or unusual variants

- Non-superuser connections need specific grants — see [Database Permissions](database-permissions.md).
