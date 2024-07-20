{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'PostgreSQL',
        rules: [
          {
            alert: 'PostgreSQLMaxConnectionsReached',
            annotations: {
              description: '{{ $labels.instance }} is exceeding the currently configured maximum Postgres connection limit (current value: {{ $value }}s). Services may be degraded - please take immediate action (you probably need to increase max_connections in the Docker image and re-deploy.',
              summary: 'Postgres connections count is over the maximum amount.',
            },
            expr: |||
              sum by (instance) (pg_stat_activity_count{%(postgresExporterSelector)s})
              >=
              sum by (instance) (pg_settings_max_connections{%(postgresExporterSelector)s})
              -
              sum by (instance) (pg_settings_superuser_reserved_connections{%(postgresExporterSelector)s})
            ||| % $._config,
            'for': '1m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgreSQLHighConnections',
            annotations: {
              description: '{{ $labels.instance }} is exceeding 80% of the currently configured maximum Postgres connection limit (current value: {{ $value }}s). Please check utilization graphs and confirm if this is normal service growth, abuse or an otherwise temporary condition or if new resources need to be provisioned (or the limits increased, which is mostly likely).',
              summary: 'Postgres connections count is over 80% of maximum amount.',
            },
            expr: |||
              sum by (instance) (pg_stat_activity_count{%(postgresExporterSelector)s})
              >
              (
                sum by (instance) (pg_settings_max_connections{%(postgresExporterSelector)s})
                -
                sum by (instance) (pg_settings_superuser_reserved_connections{%(postgresExporterSelector)s})
              ) * 0.8
            ||| % $._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgreSQLDown',
            annotations: {
              description: '{{ $labels.instance }} is rejecting query requests from the exporter, and thus probably not allowing DNS requests to work either. User services should not be effected provided at least 1 node is still alive.',
              summary: 'PostgreSQL is not processing queries.',
            },
            expr: 'pg_up{%(postgresExporterSelector)s} != 1' % $._config,
            'for': '1m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgreSQLSlowQueries',
            annotations: {
              description: 'PostgreSQL high number of slow queries {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }} ',
              summary: 'PostgreSQL high number of slow queries.',
            },
            expr: |||
              avg by (datname) (
                rate (
                  pg_stat_activity_max_tx_duration{datname!~"template.*",%(postgresExporterSelector)s}[2m]
                )
              ) > 2 * 60
            ||| % $._config,
            'for': '2m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgreSQLQPS',
            annotations: {
              description: 'PostgreSQL high number of queries per second on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL high number of queries per second.',
            },
            expr: |||
              avg by (datname) (
                irate(
                  pg_stat_database_xact_commit{datname!~"template.*",%(postgresExporterSelector)s}[5m]
                )
                +
                irate(
                  pg_stat_database_xact_rollback{datname!~"template.*",%(postgresExporterSelector)s}[5m]
                )
              ) > 10000
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgreSQLCacheHitRatio',
            annotations: {
              description: 'PostgreSQL low on cache hit rate on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL low cache hit rate.',
            },
            expr: |||
              avg by (datname) (
                rate(pg_stat_database_blks_hit{datname!~"template.*",%(postgresExporterSelector)s}[5m])
                /
                (
                  rate(
                    pg_stat_database_blks_hit{datname!~"template.*",%(postgresExporterSelector)s}[5m]
                  )
                  +
                  rate(
                    pg_stat_database_blks_read{datname!~"template.*",%(postgresExporterSelector)s}[5m]
                  )
                )
              ) < 0.98
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresHasTooManyRollbacks',
            annotations: {
              description: 'PostgreSQL has too many rollbacks on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL has too many rollbacks.',
            },
            expr: |||
              avg without(pod, instance)
              (rate(pg_stat_database_xact_rollback{db_name!~"template.*|^$"}[5m]) /
              (rate(pg_stat_database_xact_commit{db_name!~"template.*|^$"}[5m])+ rate(pg_stat_database_xact_rollback{db_name!~"template.*|^$"}[5m]))) > 0.10
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresHasHighDeadLocks',
            annotations: {
              description: 'PostgreSQL has too high deadlocks on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL has high number of deadlocks.',
            },
            expr: |||
              max without(pod, instance) (rate(pg_stat_database_deadlocks{db_name!~"template.*|^$"}[5m]) * 60) > 5
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresAcquiredTooManyLocks',
            annotations: {
              description: 'PostgreSQL has acquired too many locks on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL has high number of acquired locks.',
            },
            expr: |||
              max by( server, job, db_name, asserts_env, asserts_site, namespace) ((pg_locks_count{db_name!~"template.*|^$"}) /
              on(instance, asserts_env, asserts_site, namespace) group_left(server) (pg_settings_max_locks_per_transaction{} * pg_settings_max_connections{})) > 0.20
            ||| % $._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresXLOGConsumptionVeryLow',
            annotations: {
              description: 'PostgreSQL instance {{ $labels.instance }} has a very low XLOG consumption rate.',
              summary: 'PostgreSQL XLOG consumption is very low.',
            },
            expr: 'rate(pg_xlog_position_bytes{asserts_env!=""}[5m]) < 200000',
            'for': '5m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresXLOGConsumptionVeryHigh',
            annotations: {
              description: '{{ $labels.instance }} is experiencing very high XLOG consumption rate, which might indicate excessive write operations.',
              summary: 'PostgreSQL very high XLOG consumption rate.',
            },
            expr: 'rate(pg_xlog_position_bytes{asserts_env!=""}[2m]) > 36700160 and on (instance, asserts_env, asserts_site) (pg_replication_is_replica{asserts_env!=""} == 0)',
            'for': '10m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresReplicationStopped',
            annotations: {
              description: 'PostgreSQL instance {{ $labels.instance }} has stopped replication.',
              summary: 'PostgreSQL replication has stopped.',
            },
            expr: 'pg_stat_replication_pg_xlog_location_diff{asserts_env!=""} != 0',
            'for': '5m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresReplicationLagging_More_1Hour',
            annotations: {
              description: '{{ $labels.instance }} replication lag exceeds 1 hour. Check for network issues or load imbalances.',
              summary: 'PostgreSQL replication lagging more than 1 hour.',
            },
            expr: '(pg_replication_lag{asserts_env!=""} > 3600) and on (instance) (pg_replication_is_replica{asserts_env!=""} == 1)',
            'for': '5m',
            labels: {
              asserts_severity: 'warning',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresReplicationLagBytesAreTooLarge',
            annotations: {
              description: '{{ $labels.instance }} replication lag in bytes is too large, which might indicate replication issues or network bottlenecks.',
              summary: 'PostgreSQL replication lag in bytes too large.',
            },
            expr: '(pg_xlog_position_bytes{asserts_env!=""} and pg_replication_is_replica{asserts_env!=""} == 0) - on (job, service, asserts_env, asserts_site) group_right(instance) (pg_xlog_position_bytes{asserts_env!=""} and pg_replication_is_replica{asserts_env!=""} == 1) > 1e+09',
            'for': '5m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresHasReplicationSlotUsed',
            annotations: {
              description: '{{ $labels.instance }} has replication slots that are not used, which might lead to replication lag or data inconsistency.',
              summary: 'PostgreSQL has unused replication slots.',
            },
            expr: 'pg_replication_slots_active{asserts_env!=""} == 0',
            'for': '30m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresReplicationIsStale',
            annotations: {
              description: '{{ $labels.instance }} replication slots have not been updated for a significant period, indicating potential issues with replication.',
              summary: 'PostgreSQL replication slots are stale.',
            },
            expr: 'pg_replication_slots_xmin_age{asserts_env!="", slot_name =~ "^repmgr_slot_[0-9]+"} > 20000',
            'for': '30m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresReplicationRoleChanged',
            annotations: {
              description: '{{ $labels.instance }} replication role has changed. Verify if this is expected or if it indicates a failover.',
              summary: 'PostgreSQL replication role change detected.',
            },
            expr: 'pg_replication_is_replica{asserts_env!=""} and changes(pg_replication_is_replica{asserts_env!=""}[1m]) > 0',
            labels: {
              asserts_severity: 'warning',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresHasExporterErrors',
            annotations: {
              description: '{{ $labels.instance }} exporter is experiencing errors. Verify exporter health and configuration.',
              summary: 'PostgreSQL exporter errors detected.',
            },
            expr: 'pg_exporter_last_scrape_error{asserts_env!=""} > 0',
            'for': '30m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresHasTooManyDeadTuples',
            annotations: {
              description: '{{ $labels.instance }} has too many dead tuples, which may lead to inefficient query performance. Consider vacuuming the database.',
              summary: 'PostgreSQL has too many dead tuples.',
            },
            expr: '(sum without(relname) (pg_stat_user_tables_n_dead_tup{asserts_env!="", db_name!~"template.*|^$"}) > 10000) / ((sum without(relname) (pg_stat_user_tables_n_live_tup{asserts_env!="", db_name!~"template.*|^$"}) + sum without(relname)(pg_stat_user_tables_n_dead_tup{asserts_env!="", db_name!~"template.*|^$"})) > 0) >= 0.1 unless on(instance, asserts_env, asserts_site) (pg_replication_is_replica{asserts_env!=""} == 1)',
            'for': '5m',
            labels: {
              asserts_severity: 'warning',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresTablesNotVaccumed',
            annotations: {
              description: '{{ $labels.instance }} tables have not been vacuumed recently, which may lead to performance degradation.',
              summary: 'PostgreSQL tables not vacuumed.',
            },
            expr: 'group without(pod, instance)(timestamp(pg_stat_user_tables_n_dead_tup{asserts_env!=""} > pg_stat_user_tables_n_live_tup{asserts_env!=""} * on(asserts_env, asserts_site, namespace, job, service, instance, server) group_left pg_settings_autovacuum_vacuum_scale_factor{asserts_env!=""} + on(asserts_env, asserts_site, namespace, job, service, instance, server) group_left pg_settings_autovacuum_vacuum_threshold{asserts_env!=""})) < time() - 36000',
            'for': '30m',
            labels: {
              asserts_severity: 'critical',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresTableNotAnalyzed',
            annotations: {
              description: '{{ $labels.instance }} table has not been analyzed recently, which might lead to inefficient query planning.',
              summary: 'PostgreSQL table not analyzed.',
            },
            expr: '
              group without(pod, instance)(
                timestamp(
                pg_stat_user_tables_n_dead_tup{asserts_env!=""} >
                  pg_stat_user_tables_n_live_tup{asserts_env!=""}
                      * on(asserts_env, asserts_site, namespace, job, service, instance, server) group_left pg_settings_autovacuum_analyze_scale_factor{asserts_env!=""}
                      + on(asserts_env, asserts_site, namespace, job, service, instance, server) group_left pg_settings_autovacuum_analyze_threshold{asserts_env!=""}
                )
                -
                pg_stat_user_tables_last_autoanalyze{asserts_env!=""}
                > 24 * 60 * 60
              )',
            labels: {
              asserts_severity: 'warning',
              asserts_entity_type: 'DataSource',
              asserts_alert_category: 'failure',
            },
          },
          {
            alert: 'PostgresTooManyCheckpointsRequested',
            annotations: {
              description: '{{ $labels.instance }} is requesting too many checkpoints, which may lead to performance degradation.',
              summary: 'PostgreSQL too many checkpoints requested.',
            },
            expr:'
              rate(pg_stat_bgwriter_checkpoints_timed_total{asserts_env!=""}[5m]) /
              (rate(pg_stat_bgwriter_checkpoints_timed_total{asserts_env!=""}[5m]) + rate(pg_stat_bgwriter_checkpoints_req_total{asserts_env!=""}[5m]))
              < 0.5',
            'for': '5m',
            labels: {
              asserts_severity: 'warning',
              asserts_entity_type: 'Service',
              asserts_alert_category: 'failure',
            },
          },
        ],
      },
    ],
  },
}
