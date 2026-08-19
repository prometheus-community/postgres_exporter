# Getting Started

## Install

The exporter is a single static binary with no runtime dependencies, distributed as a container image and as release tarballs. Pick whichever fits your environment — the rest of this page is the same either way.

### Docker

```bash
docker pull quay.io/prometheuscommunity/postgres-exporter
```

See [Docker Images](docker.md) for the other registries, the tagging scheme, and details about the image itself.

### Prebuilt binary

Download the tarball for your platform from the [releases page](https://github.com/prometheus-community/postgres_exporter/releases) and extract it:

```bash
VERSION=0.20.1
curl -LO "https://github.com/prometheus-community/postgres_exporter/releases/download/v${VERSION}/postgres_exporter-${VERSION}.linux-amd64.tar.gz"
tar xzf "postgres_exporter-${VERSION}.linux-amd64.tar.gz"
cd "postgres_exporter-${VERSION}.linux-amd64"
```

Builds are published for Linux, macOS, Windows, and several BSDs across a range of architectures.

### Kubernetes

The `prometheus-postgres-exporter` [Helm chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-postgres-exporter) is the supported way to deploy the exporter on Kubernetes. It's a Prometheus Community project like this one, but it lives in the [helm-charts](https://github.com/prometheus-community/helm-charts) repository and has its own maintainers — report chart problems there, not here.

### From source

```bash
git clone https://github.com/prometheus-community/postgres_exporter.git
cd postgres_exporter
make build
```

### Elsewhere

The exporter is also packaged by various Linux distributions, the FreeBSD ports tree, and other third parties. Those packages are not produced or supported by this project; their contents and versions may differ from the official releases above.

## Run the exporter

Point the exporter at an existing PostgreSQL server. With Docker:

```bash
docker run \
  --net=host \
  -e DATA_SOURCE_URI="my-postgres-host:5432/postgres?sslmode=disable" \
  -e DATA_SOURCE_USER=postgres \
  -e DATA_SOURCE_PASS=password \
  quay.io/prometheuscommunity/postgres-exporter
```

Or, with the binary:

```bash
export DATA_SOURCE_URI="my-postgres-host:5432/postgres?sslmode=disable"
export DATA_SOURCE_USER=postgres
export DATA_SOURCE_PASS=password
./postgres_exporter
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
