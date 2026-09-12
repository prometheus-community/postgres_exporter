---
title: Secrets
sort_rank: 6
---
# Secrets

There are several ways to supply the database credentials the exporter connects with, and separately, ways to protect the exporter's own HTTP endpoint. Prefer file-based secrets over plain environment variables where your deployment tooling supports it (e.g. Kubernetes Secrets mounted as files, Docker secrets, Vault agent templates) — environment variables are visible to anything that can read the process's environment (e.g. `/proc/<pid>/environ`, `docker inspect`).

## Database credentials

The exporter resolves the connection to use in this order, first match wins:

1. **`DATA_SOURCE_NAME`** — a full DSN (or comma-separated list of DSNs), credentials included. Simplest, but the password ends up in the environment.
2. **`DATA_SOURCE_USER_FILE` / `DATA_SOURCE_PASS_FILE`** — file paths; the exporter reads the username/password from these files at startup. Use this to keep the password out of the environment entirely.
3. **`DATA_SOURCE_USER` / `DATA_SOURCE_PASS`** — plain environment variables, used only if the corresponding `_FILE` variable isn't set.

The host/port/database/query-string portion is supplied separately via `DATA_SOURCE_URI` (or `DATA_SOURCE_URI_FILE`), which never contains credentials.

Recommended for anything beyond local testing:

```bash
docker run \
  -e DATA_SOURCE_URI="my-postgres-host:5432/postgres?sslmode=require" \
  -e DATA_SOURCE_USER=postgres_exporter \
  -e DATA_SOURCE_PASS_FILE=/run/secrets/pg_password \
  -v /path/to/pg_password:/run/secrets/pg_password:ro \
  quay.io/prometheuscommunity/postgres-exporter
```

The mounted secret file must be readable by the container's uid/gid (`65534` in the official image — see [Docker Images](docker.md)).

### IAM authentication (single-target)

As an alternative to a static password, set `DATA_SOURCE_AUTH=iam` to authenticate with AWS RDS/Aurora IAM database authentication instead — the exporter mints a fresh, short-lived token for every connection rather than using a password at all. This takes precedence over `DATA_SOURCE_PASS`/`DATA_SOURCE_PASS_FILE`, which are ignored and don't need to be set.

* `DATA_SOURCE_AUTH=iam` — enables IAM auth for the primary datasource. The host/database still come from `DATA_SOURCE_URI`/`DATA_SOURCE_NAME` as usual; the user (from `DATA_SOURCE_USER`) must have `rds_iam` granted and no password.
* `DATA_SOURCE_REGION` — optional; the AWS region of the target cluster, resolved via the AWS SDK's usual chain (env, shared config, instance metadata) if unset.
* `DATA_SOURCE_ROLE` — optional; an IAM role ARN to assume via STS before minting a token, otherwise uses whatever AWS credentials are already available (e.g. an instance/task role).

See [Running Against AWS RDS](aws-rds.md#iam-database-authentication) for the IAM policy the connecting role/user needs.

### Multi-target (`/probe`) credentials

For the multi-target `/probe` endpoint, credentials are defined in the [config file](configuration.md#config-file) under `auth_modules` instead of environment variables, and selected per-request via `?auth_module=<name>`:

```yaml
auth_modules:
  prod:
    type: userpass
    userpass:
      username: monitoring_user
      password: monitoring_password
```

This keeps credentials out of Prometheus's target list and scrape URLs (which otherwise show up in Prometheus's UI and logs). See [Connecting to PostgreSQL](connecting.md#multi-target-mode-probe) for the full request flow. Note that the config file itself contains plaintext passwords, so its file permissions and any secrets-management layer around it (e.g. templating it from Vault) matter just as much as protecting `DATA_SOURCE_PASS_FILE`.

IAM authentication is also available for multi-target, via an `iam` auth module instead of `userpass` — see [Configuration](configuration.md#config-file).

## Exporter HTTP endpoint

By default, `/metrics` is served over plain HTTP with no authentication — anyone who can reach the port can read your database metrics (which include things like query text, if `stat_statements.include_query` is enabled). Use `--web.config.file` to add TLS and/or HTTP basic auth to the exporter's own listener:

```bash
./postgres_exporter --web.config.file=web-config.yml
```

The file format (certificate paths, basic auth username/bcrypt-hashed password) is documented in the [exporter-toolkit web-configuration docs](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md). In most setups it's simpler to leave this endpoint on plain HTTP and restrict network access instead (firewall rules, a private network, or a reverse proxy that terminates TLS/auth) — but `--web.config.file` is there if you need the exporter to handle it directly.

## Connection-level encryption

`sslmode` (and any other `lib/pq`-supported parameter) is set as part of the DSN's query string — see [Connecting to PostgreSQL](connecting.md). At minimum, use `sslmode=require` (or stricter, e.g. `verify-full`) for connections that cross a network boundary; `sslmode=disable` is only appropriate for connections over a trusted local socket, as in the getting-started examples.
