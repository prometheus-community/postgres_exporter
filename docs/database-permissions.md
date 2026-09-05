# Database Permissions

The exporter needs a PostgreSQL role it can connect as, plus enough privileges to read the `pg_stat_*` / `pg_settings` views the enabled collectors query. Two options: use a superuser (simplest, not recommended beyond local testing), or grant a dedicated monitoring role the minimum it needs.

## PostgreSQL 10+: use the built-in `pg_monitor` role

From PostgreSQL 10 onward, the built-in `pg_monitor` role grants read access to the monitoring views the exporter's collectors use. This is the recommended approach.

```sql
-- Create the role (skip if it already exists, e.g. on a managed service like RDS)
CREATE USER postgres_exporter WITH PASSWORD 'password';
ALTER USER postgres_exporter SET SEARCH_PATH TO postgres_exporter,pg_catalog;

-- If deploying as a non-superuser (e.g. on AWS RDS), grant the role to your
-- master/admin user first so it can grant pg_monitor:
-- GRANT postgres_exporter TO <MASTER_USER>;

GRANT CONNECT ON DATABASE postgres TO postgres_exporter;
GRANT pg_monitor TO postgres_exporter;
```

`pg_read_all_stats` is a narrower alternative to `pg_monitor` if you only need read access to statistics views and not the other privileges `pg_monitor` bundles.

Connect using the `postgres` database in your DSN, e.g.:

```
DATA_SOURCE_NAME=postgresql://postgres_exporter:password@localhost:5432/postgres?sslmode=disable
```

## PostgreSQL < 10

Older versions don't have `pg_monitor`. Views in PostgreSQL run with the querying user's permissions, so a non-superuser can't read `pg_stat_activity` / `pg_stat_replication` / `pg_stat_statements` directly. Instead, create `SECURITY DEFINER` functions (as a superuser) that expose the data, and views on top of them that the exporter's role can select from:

```sql
CREATE USER postgres_exporter WITH PASSWORD 'password';
ALTER USER postgres_exporter SET SEARCH_PATH TO postgres_exporter,pg_catalog;
GRANT CONNECT ON DATABASE postgres TO postgres_exporter;

CREATE SCHEMA IF NOT EXISTS postgres_exporter;
GRANT USAGE ON SCHEMA postgres_exporter TO postgres_exporter;

CREATE OR REPLACE FUNCTION get_pg_stat_activity() RETURNS SETOF pg_stat_activity AS
$$ SELECT * FROM pg_catalog.pg_stat_activity; $$
LANGUAGE sql VOLATILE SECURITY DEFINER;

CREATE OR REPLACE VIEW postgres_exporter.pg_stat_activity
AS SELECT * FROM get_pg_stat_activity();
GRANT SELECT ON postgres_exporter.pg_stat_activity TO postgres_exporter;

CREATE OR REPLACE FUNCTION get_pg_stat_replication() RETURNS SETOF pg_stat_replication AS
$$ SELECT * FROM pg_catalog.pg_stat_replication; $$
LANGUAGE sql VOLATILE SECURITY DEFINER;

CREATE OR REPLACE VIEW postgres_exporter.pg_stat_replication
AS SELECT * FROM get_pg_stat_replication();
GRANT SELECT ON postgres_exporter.pg_stat_replication TO postgres_exporter;

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE OR REPLACE FUNCTION get_pg_stat_statements() RETURNS SETOF pg_stat_statements AS
$$ SELECT * FROM public.pg_stat_statements; $$
LANGUAGE sql VOLATILE SECURITY DEFINER;

CREATE OR REPLACE VIEW postgres_exporter.pg_stat_statements
AS SELECT * FROM get_pg_stat_statements();
GRANT SELECT ON postgres_exporter.pg_stat_statements TO postgres_exporter;
```

## Extension-specific collectors

A few collectors need an extension installed in the target database before they'll return data:

- `stat_statements` — requires `CREATE EXTENSION pg_stat_statements;` (and, on managed services like RDS, adding it to `shared_preload_libraries` and rebooting the instance — see [Running Against AWS RDS](aws-rds.md)).
- `buffercache_summary` — requires `CREATE EXTENSION pg_buffercache;`, and only runs on PostgreSQL 16+.

Collector failures are independent and logged per-collector — a missing extension causes that one collector's metrics to be absent (with an error in the logs), not the whole scrape to fail.
