# Running Against AWS RDS

RDS is a managed PostgreSQL service, so you don't have superuser access — a few things need to be set up differently than on a self-managed server.

## Excluded database and role

`rdsadmin` is an RDS-internal database and role you can't (and shouldn't) query. Point `DATA_SOURCE_NAME` at your actual database (e.g. `postgres`), same as any other target — see [Connecting to PostgreSQL](connecting.md). If you use the deprecated `--auto-discover-databases` flag to scrape every database on the instance, exclude it with `PG_EXPORTER_EXCLUDE_DATABASES=rdsadmin`.

## Permissions

Follow the standard non-superuser [database permissions](database-permissions.md) setup, granting `pg_monitor` to your exporter role. Since you don't have a true superuser on RDS, run the grant as your RDS master user, and grant the exporter role to the master user first if needed (see the note in [database-permissions.md](database-permissions.md#postgresql-10-use-the-built-in-pg_monitor-role)).

## `pg_stat_statements`

To use the `stat_statements` collector, `pg_stat_statements` must be preloaded via the RDS parameter group, not just created as an extension:

1. In the RDS parameter group attached to your instance, set:
   ```
   shared_preload_libraries = "pg_stat_statements,pg_hint_plan"
   ```
2. Reboot the RDS instance for the change to take effect.
3. Then `CREATE EXTENSION pg_stat_statements;` as usual.

`pg_stat_statements` rows include queries run by RDS's own `rdsadmin` role. Exclude them with `--collector.stat_statements.exclude_users=rdsadmin`.
