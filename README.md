[![Build Status](https://travis-ci.org/wrouesnel/postgres_exporter.svg?branch=master)](https://travis-ci.org/wrouesnel/postgres_exporter)
[![Coverage Status](https://coveralls.io/repos/github/wrouesnel/postgres_exporter/badge.svg?branch=master)](https://coveralls.io/github/wrouesnel/postgres_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/wrouesnel/postgres_exporter)](https://goreportcard.com/report/github.com/wrouesnel/postgres_exporter)

# PostgreSQL Server Exporter

Prometheus exporter for PostgreSQL server metrics.

CI Tested PostgreSQL versions: `9.1`, `9.2`, `9.3`, `9.4`, `9.5`, `9.6`, `10`, `11`

## Quick Start
This package is available for Docker:
```
# Start an example database
docker run --net=host -it --rm -e POSTGRES_PASSWORD=password postgres
# Connect to it
docker run --net=host -e DATA_SOURCE_NAME="postgresql://postgres:password@localhost:5432/?sslmode=disable" wrouesnel/postgres_exporter
# Connect to multiple databases (separate by comma)
docker run --net=host -e DATA_SOURCE_NAME="postgresql://postgres:password@localhost:5432/?sslmode=disable,postgresql://postgres:password@localhost:5432/test_db?sslmode=disable" wrouesnel/postgres_exporter

```

## Building and running

The build system is based on [Mage](https://magefile.org)

The default make file behavior is to build the binary:
```
$ go get github.com/wrouesnel/postgres_exporter
$ cd ${GOPATH-$HOME/go}/src/github.com/wrouesnel/postgres_exporter
$ go run mage.go
$ export DATA_SOURCE_NAME="postgresql://login:password@hostname:port/dbname"
$ ./postgres_exporter <flags>
```

To build the dockerfile, run `go run mage.go docker`.

This will build the docker image as `wrouesnel/postgres_exporter:latest`. This
is a minimal docker image containing *just* postgres_exporter. By default no SSL
certificates are included, if you need to use SSL you should either bind-mount
`/etc/ssl/certs/ca-certificates.crt` or derive a new image containing them.

### Vendoring
Package vendoring is handled with [`govendor`](https://github.com/kardianos/govendor)

### Flags

* `web.listen-address`
  Address to listen on for web interface and telemetry. Default is `:9187`.

* `web.telemetry-path`
  Path under which to expose metrics. Default is `/metrics`.

* `disable-default-metrics`
  Use only metrics supplied from `queries.yaml` via `--extend.query-path`

* `disable-setting-metrics`
  Removes metrics like pg_setting

* `extend.query-path`
  Path to a YAML file containing custom queries to run. Check out [`queries.yaml`](queries.yaml)
  for examples of the format.
 
* `dumpmaps`
  Do not run - print the internal representation of the metric maps. Useful when debugging a custom
  queries file.
  
* `log.level`
  Set logging level: one of `debug`, `info`, `warn`, `error`, `fatal`

* `log.format`
  Set the log output target and format. e.g. `logger:syslog?appname=bob&local=7` or `logger:stdout?json=true`
  Defaults to `logger:stderr`.

* `constantLabels`
  Labels to set in all metrics. A list of `label=value` pairs, separated by commas.

### Environment Variables

The following environment variables configure the exporter:

* `DATA_SOURCE_NAME`
  the default legacy format. Accepts URI form and key=value form arguments. The
  URI may contain the username and password to connect with.

* `DATA_SOURCE_URI`
   an alternative to DATA_SOURCE_NAME which exclusively accepts the raw URI
   without a username and password component.

* `DATA_SOURCE_USER`
  When using `DATA_SOURCE_URI`, this environment variable is used to specify
  the username.
* `DATA_SOURCE_USER_FILE`
  The same, but reads the username from a file.

* `DATA_SOURCE_PASS`
  When using `DATA_SOURCE_URI`, this environment variable is used to specify
  the password to connect with.
* `DATA_SOURCE_PASS_FILE`
  The same as above but reads the password from a file.
  
* `PG_EXPORTER_WEB_LISTEN_ADDRESS`
  Address to listen on for web interface and telemetry. Default is `:9187`.

* `PG_EXPORTER_WEB_TELEMETRY_PATH`
  Path under which to expose metrics. Default is `/metrics`.

* `PG_EXPORTER_DISABLE_DEFAULT_METRICS`
  Use only metrics supplied from `queries.yaml`. Value can be `true` or `false`. Default is `false`.

* `PG_EXPORTER_EXTEND_QUERY_PATH`
  Path to a YAML file containing custom queries to run. Check out [`queries.yaml`](queries.yaml)
  for examples of the format.

* `PG_EXPORTER_CONSTANT_LABELS`
  Labels to set in all metrics. A list of `label=value` pairs, separated by commas.
  
Settings set by environment variables starting with `PG_` will be overwritten by the corresponding CLI flag if given.

### Setting the Postgres server's data source name

The PostgreSQL server's [data source name](http://en.wikipedia.org/wiki/Data_source_name)
must be set via the `DATA_SOURCE_NAME` environment variable.

For running it locally on a default Debian/Ubuntu install, this will work (transpose to init script as appropriate):

    sudo -u postgres DATA_SOURCE_NAME="user=postgres host=/var/run/postgresql/ sslmode=disable" postgres_exporter

See the [github.com/lib/pq](http://github.com/lib/pq) module for other ways to format the connection string.

### Adding new metrics

The exporter will attempt to dynamically export additional metrics if they are added in the
future, but they will be marked as "untyped". Additional metric maps can be easily created
from Postgres documentation by copying the tables and using the following Python snippet:

```python
x = """tab separated raw text of a documentation table"""
for l in StringIO(x):
    column, ctype, description = l.split('\t')
    print """"{0}" : {{ prometheus.CounterValue, prometheus.NewDesc("pg_stat_database_{0}", "{2}", nil, nil) }}, """.format(column.strip(), ctype, description.strip())
```
Adjust the value of the resultant prometheus value type appropriately. This helps build
rich self-documenting metrics for the exporter.

### Adding new metrics via a config file

The -extend.query-path command-line argument specifies a YAML file containing additional queries to run.
Some examples are provided in [queries.yaml](queries.yaml).

### Disabling default metrics
To work with non-officially-supported postgres versions you can try disabling (e.g. 8.2.15) 
or a variant of postgres (e.g. Greenplum) you can disable the default metrics with the `--disable-default-metrics`
flag. This removes all built-in metrics, and uses only metrics defined by queries in the `queries.yaml` file you supply
(so you must supply one, otherwise the exporter will return nothing but internal statuses and not your database).

### Running as non-superuser

To be able to collect metrics from `pg_stat_activity` and `pg_stat_replication` 
as  non-superuser you have to create views as a superuser, and assign permissions 
separately to those.  

In PostgreSQL, views run with the permissions of the user that created them so 
they can act as security barriers.

```sql
CREATE USER postgres_exporter PASSWORD 'password';
ALTER USER postgres_exporter SET SEARCH_PATH TO postgres_exporter,pg_catalog;

-- If deploying as non-superuser (for example in AWS RDS), uncomment the GRANT
-- line below and replace <MASTER_USER> with your root user.
-- GRANT postgres_exporter TO <MASTER_USER>
CREATE SCHEMA postgres_exporter AUTHORIZATION postgres_exporter;

CREATE VIEW postgres_exporter.pg_stat_activity
AS
  SELECT * from pg_catalog.pg_stat_activity;

GRANT SELECT ON postgres_exporter.pg_stat_activity TO postgres_exporter;

CREATE VIEW postgres_exporter.pg_stat_replication AS
  SELECT * from pg_catalog.pg_stat_replication;

GRANT SELECT ON postgres_exporter.pg_stat_replication TO postgres_exporter;
```

> **NOTE**
> <br />Remember to use `postgres` database name in the connection string:
> ```
> DATA_SOURCE_NAME=postgresql://postgres_exporter:password@localhost:5432/postgres?sslmode=disable
> ```

# Hacking
* To build a copy for your current architecture run `go run mage.go binary` or just `go run mage.go`
  This will create a symlink to the just built binary in the root directory.
* To build release tar balls run `go run mage.go release`.
* Build system is a bit temperamental at the moment since the conversion to mage - I am working on getting it
  to be a perfect out of the box experience, but am time-constrained on it at the moment.
