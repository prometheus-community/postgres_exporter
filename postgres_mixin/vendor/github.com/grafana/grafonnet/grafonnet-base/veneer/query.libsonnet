local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';


// Upstream schema `DataQuery.datasource` is not properly defined, this bit of veneer
// provides a generic way for setting the datasource on a query type.
local datasourceFunction(type) = {
  '#withDatasource':: d.func.new(
    'Set the datasource for this query.',
    args=[
      d.arg('value', d.T.string),
    ]
  ),
  withDatasource(value): {
    datasource+: {
      type: type,
      uid: value,
    },
  },
};
local veneer = {
  loki+:
    {
      '#new':: d.func.new(
        'Creates a new loki query target for panels.',
        args=[
          d.arg('datasource', d.T.string),
          d.arg('expr', d.T.string),
        ]
      ),
      new(datasource, expr):
        self.withDatasource(datasource)
        + self.withExpr(expr),

    }
    + datasourceFunction('loki'),

  prometheus+:
    {
      '#new':: d.func.new(
        'Creates a new prometheus query target for panels.',
        args=[
          d.arg('datasource', d.T.string),
          d.arg('expr', d.T.string),
        ]
      ),
      new(datasource, expr):
        self.withDatasource(datasource)
        + self.withExpr(expr),

      '#withIntervalFactor':: d.func.new(
        'Set the interval factor for this query.',
        args=[
          d.arg('value', d.T.string),
        ]
      ),
      withIntervalFactor(value): {
        intervalFactor: value,
      },

      '#withLegendFormat':: d.func.new(
        'Set the legend format for this query.',
        args=[
          d.arg('value', d.T.string),
        ]
      ),
      withLegendFormat(value): {
        legendFormat: value,
      },
    }
    + datasourceFunction('prometheus'),

  tempo+:
    {
      '#new':: d.func.new(
        'Creates a new tempo query target for panels.',
        args=[
          d.arg('datasource', d.T.string),
          d.arg('query', d.T.string),
          d.arg('filters', d.T.array),
        ]
      ),
      new(datasource, query, filters):
        self.withDatasource(datasource)
        + self.withQuery(query)
        + self.withFilters(filters),
    }
    + datasourceFunction('tempo'),
};

function(name) std.get(veneer, name, default={})
