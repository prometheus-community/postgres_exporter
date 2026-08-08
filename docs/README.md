# postgres_exporter Documentation

Documentation for running and operating `postgres_exporter`, a Prometheus exporter for PostgreSQL server metrics.

- [Getting Started](getting-started.md) — run the exporter and scrape your first metrics.
- [Docker Images](docker.md) — where the images live and how to run them.
- [Configuring the Exporter](configuration.md) — flags, environment variables, and the config file.
- [Connecting to PostgreSQL](connecting.md) — how the exporter builds connections and scrapes targets.
- [Database Permissions](database-permissions.md) — the SQL grants the exporter needs.
- [Secrets](secrets.md) — ways to supply credentials without putting them on the command line.
- [Running Against AWS RDS](aws-rds.md) — RDS-specific setup notes.
