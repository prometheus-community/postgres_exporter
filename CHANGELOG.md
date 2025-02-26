## 0.17.1 / 2025-02-26

* [BUGFIX] Fix: Handle incoming labels with invalid UTF-8 #1131

## 0.17.0 / 2025-02-16

## What's Changed
* [ENHANCEMENT] Add Postgres 17 for CI test by @khiemdoan in https://github.com/prometheus-community/postgres_exporter/pull/1105
* [ENHANCEMENT] Add wait/backend to pg_stat_activity by @fgalind1 in https://github.com/prometheus-community/postgres_exporter/pull/1106
* [ENHANCEMENT] Export last replay age in replication collector by @bitfehler in https://github.com/prometheus-community/postgres_exporter/pull/1085
* [BUGFIX] Fix pg_long_running_transactions time by @jyothikirant-sayukth in https://github.com/prometheus-community/postgres_exporter/pull/1092
* [BUGFIX] Fix to replace dashes with underscore in the metric names by @aagarwalla-fx in https://github.com/prometheus-community/postgres_exporter/pull/1103
* [BIGFIX] Checkpoint related columns in PG 17 have been moved from pg_stat_bgwriter to pg_stat_checkpointer by @n-rodriguez in https://github.com/prometheus-community/postgres_exporter/pull/1072
* [BUGFIX] Fix pg_stat_statements for PG17 by @NevermindZ4 in https://github.com/prometheus-community/postgres_exporter/pull/1114
* [BUGFIX] Handle pg_replication_slots on pg<13 by @michael-todorovic in https://github.com/prometheus-community/postgres_exporter/pull/1098
* [BUGFIX] Fix missing dsn sanitization for logging by @sysadmind in https://github.com/prometheus-community/postgres_exporter/pull/1104

## New Contributors
* @jyothikirant-sayukth made their first contribution in https://github.com/prometheus-community/postgres_exporter/pull/1092
* @aagarwalla-fx made their first contribution in https://github.com/prometheus-community/postgres_exporter/pull/1103
* @NevermindZ4 made their first contribution in https://github.com/prometheus-community/postgres_exporter/pull/1114
* @michael-todorovic made their first contribution in https://github.com/prometheus-community/postgres_exporter/pull/1098
* @fgalind1 made their first contribution in https://github.com/prometheus-community/postgres_exporter/pull/1106

**Full Changelog**: https://github.com/prometheus-community/postgres_exporter/compare/v0.16.0...v0.17.0

## 0.16.0 / 2024-11-10

BREAKING CHANGES:

The logging system has been replaced with log/slog from the stdlib. This change is being made across the prometheus ecosystem. The logging output has changed, but the messages and levels remain the same. The `ts` label for the timestamp has bewen replaced with `time`, the accuracy is less, and the timezone is not forced to UTC. The `caller` field has been replaced by the `source` field, which now includes the full path to the source file. The `level` field now exposes the log level in capital letters.

* [CHANGE] Replace logging system #1073
* [ENHANCEMENT] Add save_wal_size and wal_status to replication_slot collector #1027
* [ENHANCEMENT] Add roles collector and connection limit metrics to database collector #997
* [ENHANCEMENT] Excluded databases log messgae is now info level #1003
* [ENHANCEMENT] Add active_time to stat_database collector #961
* [ENHANCEMENT] Add slot_type label to replication_slot collector #960
* [BUGFIX] Fix walreceiver collectore when no repmgr #1086
* [BUGFIX] Remove logging errors on replicas #1048
* [BUGFIX] Fix active_time query on postgres>=14 #1045

## 0.15.0 / 2023-10-27

* [ENHANCEMENT] Add 1kB and 2kB units #915
* [BUGFIX] Add error log when probe collector creation fails #918
* [BUGFIX] Fix test build failures on 32-bit arch #919
* [BUGFIX] Adjust collector to use separate connection per scrape #936

## 0.14.0 / 2023-09-11

* [CHANGE] Add `state` label to pg_process_idle_seconds #862
* [CHANGE] Change database connections to one per scrape #882 #902
* [ENHANCEMENT] Add wal collector #858
* [ENHANCEMENT] Add database_wraparound collector #834
* [ENHANCEMENT] Add stat_activity_autovacuum collector #840
* [ENHANCEMENT] Add stat_wal_receiver collector #844
* [ENHANCEMENT] Add xlog_location collector #849
* [ENHANCEMENT] Add statio_user_indexes collector #845
* [ENHANCEMENT] Add long_running_transactions collector #836
* [ENHANCEMENT] Add pg_stat_user_tables_size_bytes metric #904
* [BUGFIX] Fix tests on 32-bit systems #857
* [BUGFIX] Fix pg_stat_statements metrics on Postgres 13+ #874 #876
* [BUGFIX] Fix pg_stat_database metrics for NULL stats_reset #877
* [BUGFIX] Fix pg_replication_lag_seconds on Postgres 10+ when master is idle #895

## 0.13.2 / 2023-07-21

* [BUGFIX] Fix type issues on pg_postmaster metrics #828
* [BUGFIX] Fix pg_replication collector instantiation #854
* [BUGFIX] Fix pg_process_idle metrics #855

## 0.13.1 / 2023-06-27

* [BUGFIX] Make collectors not fail on null values #823

## 0.13.0 / 2023-06-21

BREAKING CHANGES:

Please note, the following features are deprecated and may be removed in a future release:
- `auto-discover-databases`
- `extend.query-path`
- `constantLabels`
- `exclude-databases`
- `include-databases`

This exporter is meant to monitor PostgresSQL servers, not the user data/databases. If
you need a generic SQL report exporter https://github.com/burningalchemist/sql_exporter
is recommended.

* [CHANGE] Adjust log level for collector startup #784
* [CHANGE] Move queries from queries.yaml to collectors #801
* [CHANGE] Deprecate extend queries feature #811
* [CHANGE] Deprecate additional database features #815
* [CHANGE] Convert pg_stat_database to new collector #685
* [ENHANCEMENT] Supports alternate postgres:// prefix in URLs #787
* [BUGFIX] Fix pg_setting different help values #771
* [BUGFIX] Fix column type for pg_replication_slots #777
* [BUGFIX] Fix pg_stat_database collector #809

## 0.12.1 / 2023-06-12
* [BUGFIX] Fix column type for pg_replication_slots #777

## 0.12.0 / 2023-03-21

BREAKING CHANGES:

This release changes support for multiple postgres servers to use the
multi-target exporter pattern. This makes it much easier to monitor multiple
PostgreSQL servers from a single exporter by passing the target via URL
params. See the Multi-Target Support section of the README.

* [CHANGE] Add multi-target support #618
* [CHANGE] Add usename and application_name to pg_stat_activity metrics #673
* [FEATURE] Add replication metrics from pg_replication_slots #747
* [BUGFIX] Add dsn type for handling datasources #678
* [BUGFIX] Add 64kB unit for postgres 15 #740
* [BUGFIX] Add 4kB unit for postgres compiled with small blocks #699

## 0.11.1 / 2022-08-01

* [BUGFIX] Fix checkpoint_write_time value type #666
* [BUGFIX] Fix checkpoint_sync_time value type #667

## 0.11.0 / 2022-07-28

NOTE: pg_stat_bgwriter counter metrics had the `_total` suffix added #556

* [CHANGE] refactor pg_stat_bgwriter metrics into standalone collector #556
* [FEATURE] Add pg_database collector #613
* [ENHANCEMENT] Add pg_database_size_bytes metric #613
* [BUGFIX] Avoid parsing error from bogus Azure Flexible Server custom GUC #587
* [BUGFIX] Fix pg_stat_archiver error in 9.4 and earlier. #599
* [BUGFIX] Sanitize setting values because of Aurora irregularity #620

## 0.10.1 / 2022-01-14

* [BUGFIX] Fix broken log-level for values other than debug. #560

## 0.10.0 / 2021-07-08

* [ENHANCEMENT] Add ability to set included databases when autoDiscoverDatabases is enabled #499
* [BUGFIX] fix pg_replication_slots on postgresql versions 9.4 <> 10.0 #537

## 0.9.0 / 2021-03-01

First release under the Prometheus Community organisation.

* [CHANGE] Update build to use standard Prometheus promu/Dockerfile
* [ENHANCEMENT] Remove duplicate column in queries.yml #433
* [ENHANCEMENT] Add query for 'pg_replication_slots' #465
* [ENHANCEMENT] Allow a custom prefix for metric namespace #387
* [ENHANCEMENT] Improve PostgreSQL replication lag detection #395
* [ENHANCEMENT] Support connstring syntax when discovering databases #473
* [ENHANCEMENT] Detect SIReadLock locks in the pg_locks metric #421
* [BUGFIX] Fix pg_database_size_bytes metric in queries.yaml #357
* [BUGFIX] Don't ignore errors in parseUserQueries #362
* [BUGFIX] Fix queries.yaml for AWS RDS #370
* [BUGFIX] Recover when connection cannot be established at startup #415
* [BUGFIX] Don't retry if an error occurs #426
* [BUGFIX] Do not panic on incorrect env #457

## 0.8.0 / 2019-11-25

* Add a build info metric (#323)
* Re-add pg_stat_bgwriter metrics which were accidentally removed in the previous version. (resolves #336)
* Export pg_stat_archiver metrics (#324)
* Add support for 'DATA_SOURCE_URI_FILE' envvar.
* Resolve #329
* Added new field "master" to queries.yaml. (credit to @sfalkon)
  - If "master" is true, query will be call only on once database in instance
* Change queries.yaml for work with autoDiscoveryDatabases options (credit to @sfalkon)
  - added current database name to metrics because any database in cluster maybe have the same table names
  - added "master" field for query instance metrics.

## 0.7.0 / 2019-11-01

Introduces some more significant changes, hence the minor version bump in
such a short time frame.

* Rename pg_database_size to pg_database_size_bytes in queries.yml.
* Add pg_stat_statements to sample queries.yml file.
* Add support for optional namespace caching. (#319)
* Fix some autodiscovery problems (#314) (resolves #308)
* Yaml parsing refactor (#299)
* Don't stop generating fingerprint while encountering value with "=" sign (#318)
  (may resolve problems with passwords and special characters).

## 0.6.0 / 2019-10-30

* Add SQL for grant connect (#303)
* Expose pg_current_wal_lsn_bytes (#307)
* [minor] fix landing page content-type (#305)
* Updated lib/pg driver to 1.2.0 in order to support stronger SCRAM-SHA-256 authentication. This drops support for Go < 1.11 and PostgreSQL < 9.4. (#304)
* Provide more helpful default values for tables that have never been vacuumed (#310)
* Add retries to getServer() (#316)
* Fix pg_up metric returns last calculated value without explicit resetting (#291)
* Discover only databases that are not templates and allow connections (#297)
* Add --exclude-databases option (#298)

## 0.5.1 / 2019-07-09

* Add application_name as a label for pg_stat_replication metrics (#285).

## 0.5.0 / 2019-07-03

It's been far too long since I've done a release and we have a lot of accumulated changes.

* Docker image now runs as a non-root user named "postgres_exporter"
* Add `--auto-discover-databases` option, which automatically discovers and scrapes all databases.
* Add support for boolean data types as metrics
* Replication lag is now expressed as a float and not truncated to an integer.
* When default metrics are disabled, no version metrics are collected anymore either.
* BUGFIX: Fix exporter panic when postgres server goes down.
* Add support for collecting metrics from multiple servers.
* PostgreSQL 11 is now supported in the integration tests.

## 0.4.7 / 2018-10-02

* Added a query for v9.1 pg_stat_activity.
* Add `--constantLabels` flag to allow applying fixed constant labels to metrics.
* queries.yml: dd pg_statio_user_tables.
* Support 'B' suffix in units.

## 0.4.6 / 2018-04-15

* Fix issue #173 - 32 and 64mb unit sizes were not supported in pg_settings.

## 0.4.5 / 2018-02-27

* Add commandline flag to disable default metrics (thanks @hsun-cnnxty)

## 0.4.4 / 2018-03-21

* Bugfix for 0.4.3 which broke pg_up (it would always be 0).
* pg_up is now refreshed based on database Ping() every scrape.
* Re-release of 0.4.4 to fix version numbering.

## 0.4.2 / 2018-02-19

* Adds the following environment variables for overriding defaults:
    * `PG_EXPORTER_WEB_LISTEN_ADDRESS`
    * `PG_EXPORTER_WEB_TELEMETRY_PATH`
    * `PG_EXPORTER_EXTEND_QUERY_PATH`

* Add Content-Type to HTTP landing page.
* Fix Makefile to produce .exe binaries for Windows.

## 0.4.1 / 2017-11-30

* No code changes to v0.4.0 for the exporter.
* First release switching to tar-file based distribution.
* First release with Windows and Darwin cross-builds.\\

## 0.4.0 / 2017-11-29

* Fix panic due to inconsistent label cardinality when using queries.yaml with
  queries which return extra columns.
* Add metric for whether the user queries YAML file parsed correctly. This also
  includes the filename and SHA256 sum allowing tracking of updates.
* Add pg_up metric to indicate whether the exporter was able to connect and
  Ping() the PG instance before a scrape.
* Fix broken link in landing page for `/metrics`

## 0.3.0 / 2017-10-23

* Add support for PostgreSQL 10.

## 0.2.3 / 2017-09-07

* Add support for the 16kB unit when decoding pg_settings. (#101)

## 0.2.2 / 2017-08-04

* Fix DSN logging. The exporter previously never actually logged the DSN when
  database connections failed. This was also masking a logic error which could
  potentially lead to a crash when DSN was unparseable, though no actual
  crash could be produced in testing.

## 0.2.1 / 2017-06-07

* Ignore functions that cannot be executed during replication recovery (#52)
* Add a `-version` flag finally.
* Add confirmed_flush_lsn to pg_stat_replication.

## 0.2.0 / 2017-04-18

* Major change - use pg_settings to retrieve runtime variables. Adds >180
  new metrics and descriptions (big thanks to Matt Bostock for this work).

  Removes the following metrics:
  ```
  pg_runtime_variable_max_connections
  pg_runtime_variable_max_files_per_process
  pg_runtime_variable_max_function_args
  pg_runtime_variable_max_identifier_length
  pg_runtime_variable_max_index_keys
  pg_runtime_variable_max_locks_per_transaction
  pg_runtime_variable_max_pred_locks_per_transaction
  pg_runtime_variable_max_prepared_transactions
  pg_runtime_variable_max_standby_archive_delay_milliseconds
  pg_runtime_variable_max_standby_streaming_delay_milliseconds
  pg_runtime_variable_max_wal_senders
  ```

  They are replaced by equivalent names under `pg_settings` with the exception of
  ```
  pg_runtime_variable_max_standby_archive_delay_milliseconds
  pg_runtime_variable_max_standby_streaming_delay_milliseconds
  ```
  which are replaced with
  ```
  pg_settings_max_standby_archive_delay_seconds
  pg_settings_max_standby_streaming_delay_seconds
  ```

## 0.1.3 / 2017-02-21

* Update the Go build to 1.7.5 to include a fix for NAT handling.
* Fix passwords leaking in DB url error message on connection failure.

## 0.1.2 / 2017-02-07

* Use a connection pool of size 1 to reduce memory churn on target database.

## 0.1.1 / 2016-11-29

* Fix pg_stat_replication metrics not being collected due to semantic version
  filter problem.

## 0.1.0 / 2016-11-21

* Change default port to 9187.
* Fix regressions with pg_stat_replication on older versions of Postgres.
* Add pg_static metric to store version strings as labels.
* Much more thorough testing structure.
* Move to semantic versioning for releases and docker image publications.

## 0.0.1 / 2016-06-03

Initial release for publication.
