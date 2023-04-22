{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'PostgreSQL',
        rules: [
          {
            alert: 'PostgreSQLMaxConnectionsReached',
            annotations: {
              description: '{{ $labels.instance }} is exceeding the currently configured maximum Postgres connection limit (current value: {{ $value }}s). Services may be degraded - please take immediate action (you probably need to increase max_connections in the Docker image and re-deploy).',
              summary: 'Postgres connections count is over the maximum amount.',
            },
            expr: |||
              sum by (%(agg)s) (pg_stat_activity_count{%(postgresExporterSelector)s})
              >=
              sum by (%(agg)s) (pg_settings_max_connections{%(postgresExporterSelector)s})
              -
              sum by (%(agg)s) (pg_settings_superuser_reserved_connections{%(postgresExporterSelector)s})
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
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
              sum by (%(agg)s) (pg_stat_activity_count{%(postgresExporterSelector)s})
              >
              (
                sum by (%(agg)s) (pg_settings_max_connections{%(postgresExporterSelector)s})
                -
                sum by (%(agg)s) (pg_settings_superuser_reserved_connections{%(postgresExporterSelector)s})
              ) * 0.8
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
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
              avg by (datname, %(agg)s) (
                rate (
                  pg_stat_activity_max_tx_duration{%(dbNameFilter)s, %(postgresExporterSelector)s}[2m]
                )
              ) > 2 * 60
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
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
              avg by (datname, %(agg)s) (
                irate(
                  pg_stat_database_xact_commit{%(dbNameFilter)s, %(postgresExporterSelector)s}[5m]
                )
                +
                irate(
                  pg_stat_database_xact_rollback{%(dbNameFilter)s, %(postgresExporterSelector)s}[5m]
                )
              ) > 10000
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
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
              avg by (datname, %(agg)s) (
                rate(pg_stat_database_blks_hit{%(dbNameFilter)s, %(postgresExporterSelector)s}[5m])
                /
                (
                  rate(
                    pg_stat_database_blks_hit{%(dbNameFilter)s, %(postgresExporterSelector)s}[5m]
                  )
                  +
                  rate(
                    pg_stat_database_blks_read{%(dbNameFilter)s, %(postgresExporterSelector)s}[5m]
                  )
                )
              ) < 0.98
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
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
              (rate(pg_stat_database_xact_rollback{%(dbNameFilter)s}[5m]) /
              (rate(pg_stat_database_xact_commit{%(dbNameFilter)s}[5m]) + rate(pg_stat_database_xact_rollback{%(dbNameFilter)s}[5m]))) > 0.10
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
              max without(pod, instance) (rate(pg_stat_database_deadlocks{%(dbNameFilter)s}[5m]) * 60) > 5
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
              max by(datname, %(agg)s) (
                (pg_locks_count{%(dbNameFilter)s}) 
                /
                on(%(aggWithoutServer)s) group_left(server) (
                  pg_settings_max_locks_per_transaction{} * pg_settings_max_connections{}
                )
              ) > 0.20
            ||| % $._config { agg: std.join(',', $._config.groupLabels + $._config.instanceLabels), aggWithoutServer: std.join(',', std.filter(function(x) x != "server", $._config.groupLabels + $._config.instanceLabels)) },
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresReplicationLaggingMore1Hour',
            annotations: {
              description: '{{ $labels.instance }} replication lag exceeds 1 hour. Check for network issues or load imbalances.',
              summary: 'PostgreSQL replication lagging more than 1 hour.',
            },
            expr: |||
              (pg_replication_lag{} > 3600) and on (%(agg)s) (pg_replication_is_replica{} == 1)
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresHasReplicationSlotUsed',
            annotations: {
              description: '{{ $labels.instance }} has replication slots that are not used, which might lead to replication lag or data inconsistency.',
              summary: 'PostgreSQL has unused replication slots.',
            },
            expr: 'pg_replication_slots_active{} == 0',
            'for': '30m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'PostgresReplicationRoleChanged',
            annotations: {
              description: '{{ $labels.instance }} replication role has changed. Verify if this is expected or if it indicates a failover.',
              summary: 'PostgreSQL replication role change detected.',
            },
            expr: 'pg_replication_is_replica{} and changes(pg_replication_is_replica{}[1m]) > 0',
            labels: {
              severity: 'warning',
            },
          },
          {
            alert: 'PostgresHasExporterErrors',
            annotations: {
              description: '{{ $labels.instance }} exporter is experiencing errors. Verify exporter health and configuration.',
              summary: 'PostgreSQL exporter errors detected.',
            },
            expr: 'pg_exporter_last_scrape_error{} > 0',
            'for': '30m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'PostgresTablesNotVaccumed',
            annotations: {
              description: '{{ $labels.instance }} tables have not been vacuumed recently within the last hour, which may lead to performance degradation.',
              summary: 'PostgreSQL tables not vacuumed.',
            },
            expr: |||
              group without(pod, instance)(
                timestamp(
                  pg_stat_user_tables_n_dead_tup{} >
                    pg_stat_user_tables_n_live_tup{}
                      * on(%(agg)s) group_left pg_settings_autovacuum_vacuum_scale_factor{}
                      + on(%(agg)s) group_left pg_settings_autovacuum_vacuum_threshold{}
                )
                < time() - 36000
              )
            ||| % $._config { agg: std.join(', ', $._config.groupLabels + $._config.instanceLabels) },
            'for': '30m',
            labels: {
              severity: 'critical',
            },
          },
          {
            alert: 'PostgresTooManyCheckpointsRequested',
            annotations: {
              description: '{{ $labels.instance }} is requesting too many checkpoints, which may lead to performance degradation.',
              summary: 'PostgreSQL too many checkpoints requested.',
            },
            expr: |||
              rate(pg_stat_bgwriter_checkpoints_timed_total{}[5m]) /
              (rate(pg_stat_bgwriter_checkpoints_timed_total{}[5m]) + rate(pg_stat_bgwriter_checkpoints_req_total{}[5m]))
              < 0.5
            |||,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
          },
        ],
      },
    ],
  },
}
