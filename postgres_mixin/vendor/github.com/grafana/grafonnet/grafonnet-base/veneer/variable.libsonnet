local helpers = import '../helpers.libsonnet';
local util = import '../util/main.libsonnet';
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

// The `var` argument should match `dashboard.templating.list`
function(var) {

  '#':: d.package.newSub(
    'variable',
    |||
      Example usage:

      ```jsonnet
      local g = import 'g.libsonnet';
      local var = g.dashboard.variable;

      local customVar =
        var.custom.new(
          'myOptions',
          values=['a', 'b', 'c', 'd'],
        )
        + var.custom.generalOptions.withDescription(
          'This is a variable for my custom options.'
        )
        + var.custom.selectionOptions.withMulti();

      local queryVar =
        var.query.new('queryOptions')
        + var.query.queryTypes.withLabelValues(
          'up',
          'instance',
        )
        + var.query.withDatasource(
          type='prometheus',
          uid='mimir-prod',
        )
        + var.query.selectionOptions.withIncludeAll();


      g.dashboard.new('my dashboard')
      + g.dashboard.withVariables([
        customVar,
        queryVar,
      ])
      ```
    |||,
  ),

  local groupings = {
    generalOptions: [
      'withName',
      'withLabel',
      'withDescription',
    ],
  },

  local general =
    helpers.group(var, groupings)
    + {
      generalOptions+:
        {
          showOnDashboard: {
            '#withLabelAndValue':: d.func.new(''),
            withLabelAndValue(): var.withHide(0),
            '#withValueOnly':: d.func.new(''),
            withValueOnly(): var.withHide(1),
            '#withNothing':: d.func.new(''),
            withNothing(): var.withHide(2),
          },
        },
    },

  local selectionOptions =
    {
      selectionOptions:
        {
          '#withMulti':: d.func.new(
            'Enable selecting multiple values.',
            args=[
              d.arg('value', d.T.boolean, default=true),
            ]
          ),
          withMulti(value=true): {
            multi: value,
          },

          '#withIncludeAll':: d.func.new(
            |||
              `withIncludeAll` enables an option to include all variables.

              Optionally you can set a `customAllValue`.
            |||,
            args=[
              d.arg('value', d.T.boolean, default=true),
              d.arg('customAllValue', d.T.boolean, default=null),
            ]
          ),
          withIncludeAll(value=true, customAllValue=null): {
            includeAll: value,
            [if customAllValue != null then 'allValue']: customAllValue,
          },
        },
    },

  query:
    general
    + selectionOptions
    + {
      '#new':: d.func.new(
        |||
          Create a query template variable.

          `query` argument is optional, this can also be set with `query.queryTypes`.
        |||,
        args=[
          d.arg('name', d.T.string),
          d.arg('query', d.T.string, default=''),
        ]
      ),
      new(name, query=''):
        var.withName(name)
        + var.withType('query')
        + var.withQuery(query),

      '#withDatasource':: d.func.new(
        'Select a datasource for the variable template query.',
        args=[
          d.arg('type', d.T.string),
          d.arg('uid', d.T.string),
        ]
      ),
      withDatasource(type, uid):
        var.datasource.withType(type)
        + var.datasource.withUid(uid),

      '#withDatasourceFromVariable':: d.func.new(
        'Select the datasource from another template variable.',
        args=[
          d.arg('variable', d.T.object),
        ]
      ),
      withDatasourceFromVariable(variable):
        if variable.type == 'datasource'
        then self.withDatasource(variable.query, '${%s}' % variable.name)
        else error "`variable` not of type 'datasource'",

      '#withRegex':: d.func.new(
        |||
          `withRegex` can extract part of a series name or metric node segment. Named
          capture groups can be used to separate the display text and value
          ([see examples](https://grafana.com/docs/grafana/latest/variables/filter-variables-with-regex#filter-and-modify-using-named-text-and-value-capture-groups)).
        |||,
        args=[
          d.arg('value', d.T.string),
        ]
      ),
      withRegex(value): {
        regex: value,
      },

      '#withSort':: d.func.new(
        |||
          Choose how to sort the values in the dropdown.

          This can be called as `withSort(<number>) to use the integer values for each
          option. If `i==0` then it will be ignored and the other arguments will take
          precedence.

          The numerical values are:

          - 1 - Alphabetical (asc)
          - 2 - Alphabetical (desc)
          - 3 - Numerical (asc)
          - 4 - Numerical (desc)
          - 5 - Alphabetical (case-insensitive, asc)
          - 6 - Alphabetical (case-insensitive, desc)
        |||,
        args=[
          d.arg('i', d.T.number, default=0),
          d.arg('type', d.T.string, default='alphabetical'),
          d.arg('asc', d.T.boolean, default=true),
          d.arg('caseInsensitive', d.T.boolean, default=false),
        ],
      ),
      withSort(i=0, type='alphabetical', asc=true, caseInsensitive=false):
        if i != 0  // provide fallback to numerical value
        then { sort: i }
        else
          {
            local mapping = {
              alphabetical:
                if !caseInsensitive
                then
                  if asc
                  then 1
                  else 2
                else
                  if asc
                  then 5
                  else 6,
              numerical:
                if asc
                then 3
                else 4,
            },
            sort: mapping[type],
          },

      // TODO: Expand with Query types to match GUI
      queryTypes: {
        '#withLabelValues':: d.func.new(
          'Construct a Prometheus template variable using `label_values()`.',
          args=[
            d.arg('label', d.T.string),
            d.arg('metric', d.T.string),
          ]
        ),
        withLabelValues(label, metric):
          var.withQuery('label_values(%s, %s)' % [metric, label]),
      },

      // Deliberately undocumented, use `refresh` below
      withRefresh(value): {
        // 1 - On dashboard load
        // 2 - On time range chagne
        refresh: value,
      },

      local withRefresh = self.withRefresh,
      refresh+: {
        '#onLoad':: d.func.new(
          'Refresh label values on dashboard load.'
        ),
        onLoad(): withRefresh(1),

        '#onTime':: d.func.new(
          'Refresh label values on time range change.'
        ),
        onTime(): withRefresh(2),
      },
    },

  custom:
    general
    + selectionOptions
    + {
      '#new':: d.func.new(
        |||
          `new` creates a custom template variable.

          The `values` array accepts an object with key/value keys, if it's not an object
          then it will be added as a string.

          Example:
          ```
          [
            { key: 'mykey', value: 'myvalue' },
            'myvalue',
            12,
          ]
        |||,
        args=[
          d.arg('name', d.T.string),
          d.arg('values', d.T.array),
        ]
      ),
      new(name, values):
        var.withName(name)
        + var.withType('custom')
        + {
          // Make values array available in jsonnet
          values:: [
            if !std.isObject(item)
            then {
              key: std.toString(item),
              value: std.toString(item),
            }
            else item
            for item in values
          ],

          // Render query from values array
          query:
            std.join(',', [
              std.join(' : ', [item.key, item.value])
              for item in self.values
            ]),

          // Set current/options
          current: util.dashboard.getCurrentFromValues(self.values),
          options: util.dashboard.getOptionsFromValues(self.values),
        },

      withQuery(query): {
        values:: util.dashboard.parseCustomQuery(query),
        query: query,
      },
    },

  textbox:
    general
    {
      '#new':: d.func.new(
        '`new` creates a textbox template variable.',
        args=[
          d.arg('name', d.T.string),
          d.arg('default', d.T.string, default=''),
        ]
      ),
      new(name, default=''):
        var.withName(name)
        + var.withType('textbox')
        + {
          local this = self,
          default:: default,
          query: self.default,

          // Set current/options
          keyvaluedict:: [{ key: this.query, value: this.query }],
          current: util.dashboard.getCurrentFromValues(self.keyvaluedict),
          options: util.dashboard.getOptionsFromValues(self.keyvaluedict),
        },
    },

  constant:
    general
    {
      '#new':: d.func.new(
        '`new` creates a hidden constant template variable.',
        args=[
          d.arg('name', d.T.string),
          d.arg('value', d.T.string),
        ]
      ),
      new(name, value=''):
        var.withName(name)
        + var.withType('constant')
        + var.withHide(2)
        + var.withQuery(value),
    },

  datasource:
    general
    + selectionOptions
    + {
      '#new':: d.func.new(
        '`new` creates a datasource template variable.',
        args=[
          d.arg('name', d.T.string),
          d.arg('type', d.T.string),
        ]
      ),
      new(name, type):
        var.withName(name)
        + var.withType('datasource')
        + var.withQuery(type),

      '#withRegex':: d.func.new(
        |||
          `withRegex` filter for which data source instances to choose from in the
          variable value list. Example: `/^prod/`
        |||,
        args=[
          d.arg('value', d.T.string),
        ]
      ),
      withRegex(value): {
        regex: value,
      },
    },

  interval:
    general
    {
      '#new':: d.func.new(
        '`new` creates an interval template variable.',
        args=[
          d.arg('name', d.T.string),
          d.arg('values', d.T.array),
        ]
      ),
      new(name, values):
        var.withName(name)
        + var.withType('interval')
        + {
          // Make values array available in jsonnet
          values:: values,
          // Render query from values array
          query: std.join(',', self.values),

          // Set current/options
          keyvaluedict:: [
            {
              key: item,
              value: item,
            }
            for item in values
          ],
          current: util.dashboard.getCurrentFromValues(self.keyvaluedict),
          options: util.dashboard.getOptionsFromValues(self.keyvaluedict),
        },


      '#withAutoOption':: d.func.new(
        |||
          `withAutoOption` adds an options to dynamically calculate interval by dividing
          time range by the count specified.

          `minInterval' has to be either unit-less or end with one of the following units:
          "y, M, w, d, h, m, s, ms".
        |||,
        args=[
          d.arg('count', d.T.number),
          d.arg('minInterval', d.T.string),
        ]
      ),
      withAutoOption(count=30, minInterval='10s'): {
        auto: true,
        auto_count: count,
        auto_min: minInterval,

        // Add auto item to current/options
        keyvaluedict::
          [{ key: 'auto', value: '$__auto_interval_' + self.name }]
          + super.keyvaluedict,
      },
    },

  adhoc:
    general
    {
      '#new':: d.func.new(
        '`new` creates an adhoc template variable for datasource with `type` and `uid`.',
        args=[
          d.arg('name', d.T.string),
          d.arg('type', d.T.string),
          d.arg('uid', d.T.string),
        ]
      ),
      new(name, type, uid):
        var.withName(name)
        + var.withType('adhoc')
        + var.datasource.withType(type)
        + var.datasource.withUid(uid),

      '#newFromVariable':: d.func.new(
        'Same as `new` but selecting the datasource from another template variable.',
        args=[
          d.arg('name', d.T.string),
          d.arg('variable', d.T.object),
        ]
      ),
      newFromDatasourceVariable(name, variable):
        if variable.type == 'datasource'
        then self.new(name, variable.query, '${%s}' % variable.name)
        else error "`variable` not of type 'datasource'",

    },
}
