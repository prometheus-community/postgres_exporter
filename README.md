# PostgresSQL Server Exporter

Prometheus exporter for PostgresSQL server metrics.
Supported Postgres versions: 9.1 and up.

## Quick Start
This package is available for Docker:
```
docker run -e DATA_SOURCE_NAME="login:password@(hostname:port)/dbname" -p 9113:9113 wrouesnel/postgres_exporter
```

## Building and running
The default make file behavior is to build the binary:
```
make
export DATA_SOURCE_NAME="login:password@(hostname:port)/dbname"
./postgres_exporter <flags>
```

To build the dockerfile, run `make docker`. 

This will build the docker image as `wrouesnel/postgres_exporter:latest`. This 
is a minimal docker image containing *just* postgres_exporter. By default no SSL 
certificates are included, if you need to use SSL you should either bind-mount 
`/etc/ssl/certs/ca-certificates.crt` or derive a new image containing them.

### Vendoring
Package vendoring is handled with [`govendor`](https://github.com/kardianos/govendor)

### Flags

Name               | Description
-------------------|------------
web.listen-address | Address to listen on for web interface and telemetry.
web.telemetry-path | Path under which to expose metrics.

### Setting the Postgres server's data source name

The PostgresSQL server's [data source name](http://en.wikipedia.org/wiki/Data_source_name)
must be set via the `DATA_SOURCE_NAME` environment variable.
The format of this variable is described at https://github.com/go-sql-driver/mysql#dsn-data-source-name.

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
Adjust the value of the resultant prometheus value type appropriately. This 
helps build rich self-documenting metrics for the exporter.
