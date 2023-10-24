# Using Postgres-Exporter with AWS:RDS

### When using postgres-exporter with Amazon Web Services' RDS, the
  rolname "rdsadmin" and datname "rdsadmin" must be excluded.

I had success running docker container 'quay.io/prometheuscommunity/postgres-exporter:latest'
with queries.yaml as the PG_EXPORTER_EXTEND_QUERY_PATH.  errors
mentioned in issue#335 appeared and I had to modify the
'pg_stat_statements' query with the following:
`WHERE t2.rolname != 'rdsadmin'`

Running postgres-exporter in a container like so:
  ```
  DBNAME='postgres'
  PGUSER='postgres'
  PGPASS='psqlpasswd123'
  PGHOST='name.blahblah.us-east-1.rds.amazonaws.com'
  docker run --rm --detach \
      --name "postgresql_exporter_rds" \
      --publish 9187:9187 \
      --volume=/etc/prometheus/postgresql-exporter/queries.yaml:/var/lib/postgresql/queries.yaml \
      -e DATA_SOURCE_NAME="postgresql://${PGUSER}:${PGPASS}@${PGHOST}:5432/${DBNAME}?sslmode=disable" \
      -e PG_EXPORTER_EXCLUDE_DATABASES=rdsadmin \
      -e PG_EXPORTER_DISABLE_DEFAULT_METRICS=true \
      -e PG_EXPORTER_DISABLE_SETTINGS_METRICS=true \
      -e PG_EXPORTER_EXTEND_QUERY_PATH='/var/lib/postgresql/queries.yaml' \
      quay.io/prometheuscommunity/postgres-exporter
  ```

### Expected changes to RDS:
+ see stackoverflow notes
  (https://stackoverflow.com/questions/43926499/amazon-postgres-rds-pg-stat-statements-not-loaded#43931885)
+ you must also use a specific RDS parameter_group that includes the following:
  ```
  shared_preload_libraries = "pg_stat_statements,pg_hint_plan"
  ```
+ lastly, you must reboot the RDS instance.

