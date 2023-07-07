local helpers = import '../helpers.libsonnet';
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

local groupings = {
  panelOptions: [
    'withTitle',
    'withDescription',  // both found in root and fieldConfig.defaults ???
    'withTransparent',
    'withLinks',  // depend on sub package
    'withLinksMixin',
    'withRepeat',  // to veneer // missing maxPerRow
    'withRepeatDirection',
  ],

  queryOptions: [
    'withDatasource',  // In as-code, default to per-query datasources
    'withDatasourceMixin',
    'withMaxDataPoints',
    'withInterval',  //minInterval
    //'queryCachingTTL',  // not in schema
    'withTimeFrom',  //relativeTime
    'withTimeShift',
    //'hideTimeOverride', // not in schema
    'withTargets',  // query, expression or recorded query, not clear from schema
    'withTargetsMixin',
    'withTransformations',  // depend on very bare sub package for a very useful feature
    'withTransformationsMixin',
  ],

  standardOptions: [  // 'fieldConfig.defaults.
    'fieldConfig.defaults.withUnit',
    'fieldConfig.defaults.withMin',
    'fieldConfig.defaults.withMax',
    'fieldConfig.defaults.withDecimals',
    'fieldConfig.defaults.withDisplayName',
    'fieldConfig.defaults.color',
    'fieldConfig.defaults.withNoValue',
    'fieldConfig.defaults.withLinks',  // known as 'Data links' in UI, uses links subpackage
    'fieldConfig.defaults.withLinksMixin',
    'fieldConfig.defaults.withMappings',  // known as 'Value mappings' in UI, uses valueMapping subpackage
    'fieldConfig.defaults.withMappingsMixin',

    // fieldOverrides needs to recieve more attention in Grafonnet, the JSON is unintuitive
    // matcher = obj, properties = array, unclear in current grafonnet
    'fieldConfig.withOverrides',  // known as 'Overrides' in UI, uses fieldOverrides subpackage
    'fieldConfig.withOverridesMixin',
  ],

  'standardOptions.thresholds': [
    'fieldConfig.defaults.thresholds.withMode',
    'fieldConfig.defaults.thresholds.withSteps',
    'fieldConfig.defaults.thresholds.withStepsMixin',
  ],
};

local subPackages = [
  {
    from: 'links',
    to: 'link',
    docstring: '',
  },
  {
    from: 'transformations',
    to: 'transformation',
    docstring: '',
  },
  {
    from: 'fieldConfig.defaults.mappings',
    to: 'valueMapping',
    docstring: '',
  },
  {
    from: 'fieldConfig.defaults.thresholds.steps',
    to: 'thresholdStep',
    docstring: '',
  },
];

local toRemove = [
  // Access through more specific attributes
  '#withFieldConfig',
  '#withFieldConfigMixin',
  '#withGridPos',
  '#withGridPosMixin',
  '#withOptions',
  '#withOptionsMixin',
  'fieldConfig.#withDefaults',
  'fieldConfig.#withDefaultsMixin',
  'fieldConfig.defaults.#withColor',
  'fieldConfig.defaults.#withColorMixin',
  'fieldConfig.defaults.#withCustom',
  'fieldConfig.defaults.#withCustomMixin',
  'fieldConfig.defaults.#withThresholds',
  'fieldConfig.defaults.#withThresholdsMixin',

  // Internal
  '#withId',
  '#withPluginVersion',  // The current PluginVersion value should come from the schema, this should be set on `new()`, 9.4/9.5 schema's don't have a value.
  '#withRepeatPanelId',
  '#withType',

  // Not in UI
  '#withLibraryPanel',
  '#withLibraryPanelMixin',
  '#withTags',  // seems to be related to search
  '#withTagsMixin',
  'fieldConfig.defaults.#withDescription',
  'fieldConfig.defaults.#withDisplayNameFromDS',
  'fieldConfig.defaults.#withFilterable',  // only found in overrides
  'fieldConfig.defaults.#withPath',  // also related to overrides
  'fieldConfig.defaults.#withWriteable',

  // Old fields, not used anymore
  '#withThresholds',
  '#withThresholdsMixin',
  '#withTimeRegions',
  '#withTimeRegionsMixin',
];


function(name, panel)
  helpers.regroup(panel, groupings)
  + helpers.repackage(panel, subPackages)
  + helpers.removePaths(panel, toRemove)
  + {
    '#new':: d.func.new(
      'Creates a new %s panel with a title.' % name,
      args=[d.arg('title', d.T.string)]
    ),
    new(title):
      self.withTitle(title)
      + self.withType()
      // Default to Mixed datasource so panels can be datasource agnostic, this
      // requires query targets to explicitly set datasource, which is a lot more
      // interesting from a reusability standpoint.
      + self.datasource.withType('datasource')
      + self.datasource.withUid('-- Mixed --'),

    fieldConfig+: {
      '#overrides':: {},
      overrides+:: {},
    },
    local overrides = super.fieldConfig.overrides,
    fieldOverride:
      local matchers = [
        'byName',
        'byRegexp',
        'byType',
        'byQuery',
        'byValue',  // TODO: byValue takes more complex `options` than string
      ];
      {
        '#':: d.package.newSub(
          'fieldOverride',
          |||
            Overrides allow you to customize visualization settings for specific fields or
            series. This is accomplished by adding an override rule that targets
            a particular set of fields and that can each define multiple options.

            ```jsonnet
            fieldOverride.byType.new('number')
            + fieldOverride.byType.withPropertiesFromOptions(
              panel.standardOptions.withDecimals(2)
              + panel.standardOptions.withUnit('s')
            )
            ```
          |||
        ),
      } + {
        [matcher]: {
          '#new':: d.fn(
            '`new` creates a new override of type `%s`.' % matcher,
            args=[
              d.arg('value', d.T.string),
            ]
          ),
          new(value):
            overrides.matcher.withId(matcher)
            + overrides.matcher.withOptions(value),

          '#withProperty':: d.fn(
            |||
              `withProperty` adds a property that needs to be overridden. This function can
              be called multiple time, adding more properties.
            |||,
            args=[
              d.arg('id', d.T.string),
              d.arg('value', d.T.any),
            ]
          ),
          withProperty(id, value):
            overrides.withPropertiesMixin([
              overrides.properties.withId(id)
              + overrides.properties.withValue(value),
            ]),

          '#withPropertiesFromOptions':: d.fn(
            |||
              `withPropertiesFromOptions` takes an object with properties that need to be
              overridden. See example code above.
            |||,
            args=[
              d.arg('options', d.T.object),
            ]
          ),
          withPropertiesFromOptions(options):
            local infunc(input, path=[]) =
              std.foldl(
                function(acc, p)
                  acc + (
                    if std.isObject(input[p])
                    then infunc(input[p], path=path + [p])
                    else
                      overrides.withPropertiesMixin([
                        overrides.properties.withId(std.join('.', path + [p]))
                        + overrides.properties.withValue(input[p]),
                      ])
                  ),
                std.objectFields(input),
                {}
              );
            infunc(options.fieldConfig.defaults),
        }
        for matcher in matchers
      },
  }
