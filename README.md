# PostgreSQL Server Exporter for Aurora Serverless

Prometheus exporter for PostgreSQL server metrics.

PostgreSQL versions: `10`

This is a fork from https://github.com/prometheus-community/postgres_exporter


## Quick Start
This package is available for Docker:
```
# Start an example database
docker run --net=host -it --rm -e POSTGRES_PASSWORD=password postgres
# Connect to it
docker run \
  --net=host \
  -e DATA_SOURCE_NAME="postgresql://postgres:password@localhost:5432/postgres?sslmode=disable" \
  quay.io/prometheuscommunity/postgres-exporter
```

## Building and running

    git clone https://github.com/prometheus-community/postgres_exporter.git
    cd postgres_exporter
    make build
    ./postgres_exporter <flags>

To build the Docker image:

    make promu
    promu crossbuild -p linux/amd64 -p linux/armv7 -p linux/amd64 -p linux/ppc64le
    make docker

This will build the docker image as `prometheuscommunity/postgres_exporter:${branch}`.

### Flags

* `help`
  Show context-sensitive help (also try --help-long and --help-man).

* `version`
  Show application version.

* `iam-role-arn`
  A list of databases to remove when autoDiscoverDatabases is enabled.

* `tenant-id`
  A list of databases to only include when autoDiscoverDatabases is enabled.

* `log.level`
  Set logging level: one of `debug`, `info`, `warn`, `error`.

* `log.format`
  Set the log format: one of `logfmt`, `json`.

* `web.config.file`
  Configuration file to use TLS and/or basic authentication. The format of the
  file is described [in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

### Environment Variables

The following environment variables configure the exporter:

* `DATA_SOURCE_NAME`
  the default legacy format. Accepts URI form and key=value form arguments. The
  URI may contain the username and password to connect with.

* `PG_IAM_ROLE_ARN`
  AWS IAM role arn 

* `PG_TENANT_ID`
  Tenant ID

### Setting the Postgres server's data source name

The PostgreSQL server's [data source name](http://en.wikipedia.org/wiki/Data_source_name)
must be set via the `DATA_SOURCE_NAME` environment variable.

For running it locally on a default Debian/Ubuntu install, this will work (transpose to init script as appropriate):

    sudo -u postgres DATA_SOURCE_NAME="user=postgres host=/var/run/postgresql/ sslmode=disable" postgres_exporter

Also, you can set a list of sources to scrape different instances from the one exporter setup. Just define a comma separated string.

    sudo -u postgres DATA_SOURCE_NAME="port=5432,port=6432" postgres_exporter

See the [github.com/lib/pq](http://github.com/lib/pq) module for other ways to format the connection string.


