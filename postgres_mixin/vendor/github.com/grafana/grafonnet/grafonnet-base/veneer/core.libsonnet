local util = import '../util/main.libsonnet';
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

local veneer = {
  dashboard+: {
    // Remove legacy panels (heatmap, graph), new users should not create those.
    // Schemas are also underdeveloped.
    panels:: {},

    '#new':: d.func.new(
      'Creates a new dashboard with a title.',
      args=[d.arg('title', d.T.string)]
    ),
    new(title):
      self.withTitle(title)
      + self.withSchemaVersion()
      + self.withTimezone('utc')
      + self.time.withFrom('now-6h')
      + self.time.withTo('now'),

    // Hide functions covered by objects
    '#withTime':: {},
    '#withTimeMixin':: {},
    '#withTimepicker':: {},
    '#withTimepickerMixin':: {},
    '#withGraphTooltip':: {},
    '#withGraphTooltipMixin':: {},

    // Hide internal values
    '#withGnetId':: {},
    '#withId':: {},
    '#withRevision':: {},
    '#withVersion':: {},
    '#snapshot':: {},  // Snapshots can't be created through code afaik
    '#withSnapshot':: {},
    '#withSnapshotMixin':: {},


    local withGraphTooltip = super.withGraphTooltip,
    graphTooltip+: {
      // 0 - Default
      // 1 - Shared crosshair
      // 2 - Shared tooltip
      '#withSharedCrosshair':: d.func.new(
        'Share crosshair on all panels.',
      ),
      withSharedCrosshair():
        withGraphTooltip(1),

      '#withSharedTooltip':: d.func.new(
        'Share crosshair and tooltip on all panels.',
      ),
      withSharedTooltip():
        withGraphTooltip(2),
    },

    // Manual veneer for annotations
    '#annotations':: {},
    annotation: (import './annotation.libsonnet')(self.annotations.list),
    '#withAnnotations':
      d.func.new(
        |||
          `withAnnotations` adds an array of annotations to a dashboard.

          This function appends passed data to existing values
        |||,
        args=[d.arg('value', d.T.array)]
      ),
    withAnnotations(value): self.annotations.withList(value),
    '#withAnnotationsMixin':
      d.func.new(
        |||
          `withAnnotationsMixin` adds an array of annotations to a dashboard.

          This function appends passed data to existing values
        |||,
        args=[d.arg('value', d.T.array)]
      ),
    withAnnotationsMixin(value): self.annotations.withListMixin(value),

    // Manual veneer for links (matches UI)
    '#links':: {},
    link: (import './link.libsonnet')(self.links),

    // Manual veneer for variables (matches UI)
    variable: (import './variable.libsonnet')(self.templating.list),

    '#withVariables':
      d.func.new(
        |||
          `withVariables` adds an array of variables to a dashboard
        |||,
        args=[d.arg('value', d.T.array)]
      ),
    withVariables(value): self.templating.withList(value),

    '#withVariablesMixin':
      d.func.new(
        |||
          `withVariablesMixin` adds an array of variables to a dashboard.

          This function appends passed data to existing values
        |||,
        args=[d.arg('value', d.T.array)]
      ),
    withVariablesMixin(value): self.templating.withListMixin(value),


    // Hide from docs but keep available for backwards compatibility, use `variable` subpackage instead.
    '#templateVariable':: {},
    templateVariable:: self.templating.list,
    '#templating':: {},
    templating+: {
      list+: {
        local this = self,

        '#new':: d.func.new(
          'Create a template variable.',
          args=[
            d.arg('name', d.T.string),
            d.arg('type', d.T.string, default='query'),
          ]
        ),
        new(name, type='query'):
          {
            name: name,
            type: type,
            [if type == 'custom' then 'query']: '',
            [if type == 'custom' then 'current']:
              util.dashboard.getOptionsForCustomQuery(self.query).current,
            [if type == 'custom' then 'options']:
              util.dashboard.getOptionsForCustomQuery(self.query).options,
          },

        withType(value):
          super.withType(value)
          + {
            [if value == 'custom' then 'query']: '',
            [if value == 'custom' then 'current']:
              util.dashboard.getOptionsForCustomQuery(self.query).current,
            [if value == 'custom' then 'options']:
              util.dashboard.getOptionsForCustomQuery(self.query).options,
          },

        query+: {
          '#withLabelValues':: d.func.new(
            'Construct a Prometheus template variable using `label_values()`.',
            args=[
              d.arg('label', d.T.string),
              d.arg('metric', d.T.string),
            ]
          ),
          withLabelValues(label, metric): {
            query: 'label_values(%s, %s)' % [metric, label],
          },
        },

        '#withRegex':: d.func.new(
          'Filter the values with a regex.',
          args=[
            d.arg('value', d.T.string),
          ]
        ),
        withRegex(value): {
          regex: value,
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
          onLoad():
            withRefresh(1),

          '#onTime':: d.func.new(
            'Refresh label values on time range change.'
          ),
          onTime():
            withRefresh(2),
        },

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
          'Provide option to select "All" values.',
          args=[
            d.arg('value', d.T.boolean, default=true),
          ]
        ),
        withIncludeAll(value=true): {
          includeAll: value,
        },

        '#withAllValue':: d.func.new(
          |||
            Provide value to use with the `withIncludeAll`, this will also enable
            includeAll by default.
          |||,
          args=[
            d.arg('value', d.T.string),
          ]
        ),
        withAllValue(value):
          self.withIncludeAll(true)
          + {
            allValue: value,
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
                alphabethical:
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

        datasource+: {
          '#new':: d.func.new(
            'Select a datasource for the variable template query.',
            args=[
              d.arg('type', d.T.string),
              d.arg('uid', d.T.string),
            ]
          ),
          new(type, uid):
            self.withType(type)
            + self.withUid(uid),

          '#fromVariable':: d.func.new(
            'Select the datasource from another template variable.',
            args=[
              d.arg('variable', d.T.object),
            ]
          ),
          fromVariable(variable):
            if variable.type == 'datasource'
            then
              self.new(variable.query, '${%s}' % variable.name)
            else
              error "`variable` not of type 'datasource'",
        },
      },
    },

    withPanels(value): {
      _panels:: if std.isArray(value) then value else [value],
      panels: util.panel.setPanelIDs(self._panels),
    },
    withPanelsMixin(value): {
      _panels+:: if std.isArray(value) then value else [value],
      panels: util.panel.setPanelIDs(self._panels),
    },
  },
};

function(name) std.get(veneer, name, default={})
