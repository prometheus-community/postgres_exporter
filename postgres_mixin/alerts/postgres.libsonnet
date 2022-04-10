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
              summary: '{{ $labels.instance }} has maxed out Postgres connections.',
            },
            expr: 'sum(pg_stat_activity_count) by (instance) >= sum(pg_settings_max_connections) by (instance) - sum(pg_settings_superuser_reserved_connections) by (instance)',
            'for': '1m',
            labels: {
              severity: 'email',
            },
          },
          {
            alert: 'PostgreSQLHighConnections',
            annotations: {
              description: '{{ $labels.instance }} is exceeding 80% of the currently configured maximum Postgres connection limit (current value: {{ $value }}s). Please check utilization graphs and confirm if this is normal service growth, abuse or an otherwise temporary condition or if new resources need to be provisioned (or the limits increased, which is mostly likely).',
              summary: '{{ $labels.instance }} is over 80% of max Postgres connections.',
            },
            expr: 'sum(pg_stat_activity_count) by (instance) > (sum(pg_settings_max_connections) by (instance) - sum(pg_settings_superuser_reserved_connections) by (instance)) * 0.8',
            'for': '10m',
            labels: {
              severity: 'email',
            },
          },
          {
            alert: 'PostgreSQLDown',
            annotations: {
              description: '{{ $labels.instance }} is rejecting query requests from the exporter, and thus probably not allowing DNS requests to work either. User services should not be effected provided at least 1 node is still alive.',
              summary: 'PostgreSQL is not processing queries: {{ $labels.instance }}',
            },
            expr: 'pg_up != 1',
            'for': '1m',
            labels: {
              severity: 'email',
            },
          },
          {
            alert: 'PostgreSQLSlowQueries',
            annotations: {
              description: 'PostgreSQL high number of slow queries {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }} ',
              summary: 'PostgreSQL high number of slow on {{ $labels.cluster }} for database {{ $labels.datname }} ',
            },
            expr: 'avg(rate(pg_stat_activity_max_tx_duration{datname!~"template.*"}[2m])) by (datname) > 2 * 60',
            'for': '2m',
            labels: {
              severity: 'email',
            },
          },
          {
            alert: 'PostgreSQLQPS',
            annotations: {
              description: 'PostgreSQL high number of queries per second on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL high number of queries per second {{ $labels.cluster }} for database {{ $labels.datname }}',
            },
            expr: 'avg(irate(pg_stat_database_xact_commit{datname!~"template.*"}[5m]) + irate(pg_stat_database_xact_rollback{datname!~"template.*"}[5m])) by (datname) > 10000',
            'for': '5m',
            labels: {
              severity: 'email',
            },
          },
          {
            alert: 'PostgreSQLCacheHitRatio',
            annotations: {
              description: 'PostgreSQL low on cache hit rate on {{ $labels.cluster }} for database {{ $labels.datname }} with a value of {{ $value }}',
              summary: 'PostgreSQL low cache hit rate on {{ $labels.cluster }} for database {{ $labels.datname }}',
            },
            expr: 'avg(rate(pg_stat_database_blks_hit{datname!~"template.*"}[5m]) / (rate(pg_stat_database_blks_hit{datname!~"template.*"}[5m]) + rate(pg_stat_database_blks_read{datname!~"template.*"}[5m]))) by (datname) < 0.98',
            'for': '5m',
            labels: {
              severity: 'email',
            },
          },
        ],
      },
    ],
  },
}
