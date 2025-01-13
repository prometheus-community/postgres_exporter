local g = import 'g.libsonnet';

local dashboard = g.dashboard;
local row = g.panel.row;

local panels = import './panels.libsonnet';
local variables = import './variables.libsonnet';
local queries = import './queries.libsonnet';

// import config
local c = import '../config.libsonnet';

{
  grafanaDashboards+:: {
    'overview.json':
      dashboard.new('%s Overview' % $._config.dashboardNamePrefix)
      + dashboard.withTags($._config.dashboardTags)
      + dashboard.withRefresh('1m')
      + dashboard.time.withFrom(value='now-1h')
      + dashboard.graphTooltip.withSharedCrosshair()
      + dashboard.withVariables([
        variables.datasource,
      ])
      + dashboard.withPanels(
        g.util.grid.makeGrid([
          row.new('Overview')
          + row.withPanels([
            panels.stat.qps('QPS', queries.qps),
            panels.timeSeries.ratio1('Cache Hit Ratio', queries.cacheHitRatio),
            panels.timeSeries.base('Active Connections', queries.activeConnections)
          ]),
          row.new('server')
          + row.withPanels([
            panels.timeSeries.base('Conflicts/Deadlocks', [queries.conflicts, queries.deadlocks]),
            panels.timeSeries.base('Buffers', [queries.buffersAlloc, queries.buffersBackendFsync, queries.buffersBackend, queries.buffersClean, queries.buffersCheckpoint]),
            panels.timeSeries.base('Rows', [queries.databaseTupFetched, queries.databaseTupReturned, queries.databaseTupInserted, queries.databaseTupUpdated, queries.databaseTupDeleted]),
          ]),
        ])
      )
  }
}
