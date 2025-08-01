# Using Postgres-Exporter with AWS:RDS

### When using postgres-exporter with Amazon Web Services' RDS, the
  rolname "rdsadmin" and datname "rdsadmin" must be excluded.

Running postgres-exporter in a container like so:
  ```
  DBNAME='postgres'
  PGUSER='postgres'
  PGPASS='psqlpasswd123'
  PGHOST='name.blahblah.us-east-1.rds.amazonaws.com'
  docker run --rm --detach \
      --name "postgresql_exporter_rds" \
      --publish 9187:9187 \
      -e DATA_SOURCE_NAME="postgresql://${PGUSER}:${PGPASS}@${PGHOST}:5432/${DBNAME}?sslmode=disable" \
      -e PG_EXPORTER_EXCLUDE_DATABASES=rdsadmin \
      -e PG_EXPORTER_DISABLE_DEFAULT_METRICS=false \
      -e PG_EXPORTER_DISABLE_SETTINGS_METRICS=false \
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

