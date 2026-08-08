# Getting Started

## Run the exporter

The fastest way to try the exporter is with Docker, pointed at an existing PostgreSQL server:

```bash
docker run \
  --net=host \
  -e DATA_SOURCE_URI="localhost:5432/postgres?sslmode=disable" \
  -e DATA_SOURCE_USER=postgres \
  -e DATA_SOURCE_PASS=password \
  quay.io/prometheuscommunity/postgres-exporter
```

See [Connecting to PostgreSQL](connecting.md) for the different ways to specify the database connection, and [Secrets](secrets.md) for how to avoid putting passwords on the command line or in plain environment variables.

## Verify it's working

By default the exporter listens on port `9187` and serves metrics at `/metrics`:

```bash
curl "http://localhost:9187/metrics"
```

You should see PostgreSQL metrics prefixed with `pg_`, along with the exporter's own internal metrics (`go_*`, `promhttp_*`, etc).

## Point Prometheus at it

Add a scrape job to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: postgres
    static_configs:
      - targets: ["127.0.0.1:9187"] # replace with the exporter's actual host:port
```

## Next steps

- [Configure](configuration.md) which collectors are enabled and how the exporter is deployed.
- Grant the exporter's database user the right [permissions](database-permissions.md).
- Review [secrets](secrets.md) options before running this in anything beyond a local test.
