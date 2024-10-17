local g = import './g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import './variables.libsonnet';

{
  /*
    General overview queries
  */

  qps:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_xact_commit{}[$__rate_interval]
          )
        )
        +
        sum(
          irate(
            pg_stat_database_xact_rollback{}[$__rate_interval]
          )
        )
      |||
    ),

  cacheHitRatio:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum by (datname) (
          rate(
            pg_stat_database_blks_hit{}
          [$__rate_interval])
        ) / (
          sum by (datname) (
            rate(
              pg_stat_database_blks_hit{}
            [$__rate_interval])
          ) + sum by (datname) (
            rate(
              pg_stat_database_blks_read{}
            [$__rate_interval])
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat(|||
      {{datname}}
    |||),

  activeConnections:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
          pg_stat_database_numbackends{}
      |||
    )
    + prometheusQuery.withLegendFormat(|||
      {{datname}}
    |||),

  deadlocks:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          rate(
            pg_stat_database_deadlocks{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat(|||
      deadlocks
    |||),

  conflicts:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          rate(
            pg_stat_database_conflicts{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat(|||
      conflicts
    |||),

  /* ----------------------
        Buffers
  ---------------------- */

  buffersAlloc:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        irate(
          pg_stat_bgwriter_buffers_alloc_total{}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('allocated'),

  buffersBackendFsync:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        irate(
          pg_stat_bgwriter_buffers_backend_fsync_total{}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('backend fsyncs'),

  buffersBackend:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        irate(
          pg_stat_bgwriter_buffers_backend_total{}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('backend'),

  buffersClean:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        irate(
          pg_stat_bgwriter_buffers_clean_total{}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('clean'),

  buffersCheckpoint:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        irate(
          pg_stat_bgwriter_buffers_checkpoint_total{}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('checkpoint'),

  /* ----------------------
        Database Tups
  ---------------------- */

  databaseTupFetched:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_tup_fetched{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat('fetched'),

  databaseTupReturned:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_tup_returned{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat('returned'),

  databaseTupInserted:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_tup_inserted{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat('inserted'),

  databaseTupUpdated:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_tup_updated{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat('updated'),

  databaseTupDeleted:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          irate(
            pg_stat_database_tup_deleted{}[$__rate_interval]
          )
        )
      |||
    )
    + prometheusQuery.withLegendFormat('deleted'),

}
