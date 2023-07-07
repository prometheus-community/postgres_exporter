local helpers = import './helpers.libsonnet';
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

{
  '#':
    d.package.new(
      'crdsonnet',
      'https://github.com/crdsonnet/crdsonnet/crdsonnet',
      'Generate a *runtime* Jsonnet library directly from JSON Schemas, CRDs or OpenAPI components.',
      std.thisFile,
      'master',
    )
    + d.package.withUsageTemplate(
      '%(json_schema_simple)s' % {
        json_schema_simple: std.strReplace(
          importstr './example/json_schema_simple.libsonnet',
          '../main.libsonnet',
          'github.com/crdsonnet/crdsonnet/crdsonnet/main.libsonnet',
        ),
      }
    ),

  local root = self,
  schemaDB: import './schemadb.libsonnet',
  renderEngine: import './render.libsonnet',
  processor: import './processor.libsonnet',

  schema: {
    '#render': d.fn(
      '`render` returns a library for a `schema`.',
      args=[
        d.arg('name', d.T.string),
        d.arg('schema', d.T.object),
        d.arg('processor', d.T.object, default='processor.new()'),
      ],
    ),
    render(
      name,
      schema,
      processor=root.processor.new(),
    ):
      processor.render(name, schema),
  },

  crd: {
    local this = self,
    '#render': d.fn(
      '`render` returns a library for a `definition`.',
      args=[
        d.arg('definition', d.T.object),
        d.arg('groupSuffix', d.T.string),
        d.arg('processor', d.T.object, default='processor.new()'),
      ],
    ),
    render(
      definition,
      groupSuffix,
      processor=root.processor.new(),
    ):
      local _processor =
        processor
        + root.processor.withSchemaDB(helpers.metadataRefSchemaDB);
      local renderEngine = _processor.renderEngine;
      local grouping = helpers.getGroupKey(definition.spec.group, groupSuffix);
      local name = helpers.camelCaseKind(this.getKind(definition));
      std.foldl(
        function(acc, version)
          local schema = this.getSchemaForVersion(definition, version);
          acc
          + renderEngine.toObject(
            renderEngine.nestInParents(
              [grouping, version.name],
              _processor.render(name, schema)
            )
          )
          + renderEngine.newFunction(
            [grouping, version.name, name]
          )
        ,
        definition.spec.versions,
        renderEngine.nilvalue,
      ),
    getKind(definition):
      definition.spec.names.kind,
    getSchemaForVersion(definition, version):
      version.schema.openAPIV3Schema
      + helpers.properties.withMetadataRef()
      + helpers.properties.withGroupVersionKind(
        definition.spec.group,
        version.name,
        this.getKind(definition)
      ),
  },

  // XRD: Crossplane CompositeResourceDefinition
  // XRDs are very similar to CRDs, processing them requires slightly different behavior.
  xrd:
    self.crd
    + {
      getKind(definition):
        if std.objectHas(definition.spec, 'claimNames')
        then definition.spec.claimNames.kind
        else definition.spec.names.kind,
      getSchemaForVersion(definition, version):
        super.getSchemaForVersion(definition, version)
        + helpers.properties.withCompositeResource(),
    },

  openapi: {
    '#render': d.fn(
      '`render` returns a library for a `component` in an OpenAPI `schema`.',
      args=[
        d.arg('name', d.T.string),
        d.arg('component', d.T.object),
        d.arg('schema', d.T.object),
        d.arg('processor', d.T.object, default='processor.new()'),
      ],
    ),
    render(
      name,
      component,
      schema,
      processor=root.processor.new(),
    ):
      local extendSchema =
        std.mergePatch(
          schema,
          component
          + (if 'x-kubernetes-group-version-kind' in component
             then
               // not sure why this is a list, grabbing the first item
               local gvk = component['x-kubernetes-group-version-kind'][0];
               helpers.properties.withGroupVersionKind(gvk.group, gvk.version, gvk.kind)
             else {})
        );
      processor.render(name, extendSchema)
      + (if 'x-kubernetes-group-version-kind' in component
         then processor.renderEngine.newFunction([name])
         else processor.renderEngine.nilvalue),
  },
}

// Legacy API endpoints
// These endpoints aren't very flexible and require more arguments to add features, this is an anti-pattern. They have been reimplemented to use above modular setup as an example and to verify the modular pattern works. These functions are covered by unit tests.
+ {
  local root = self,
  local defaultRender = 'dynamic',

  fromSchema(name, schema, schemaDB={}, render=defaultRender):
    if name == ''
    then error "name can't be an empty string"
    else
      local _processor =
        root.processor.new()
        + root.processor.withSchemaDB(schemaDB)
        + root.processor.withRenderEngineType(render);
      self.schema.render(name, schema, _processor),

  fromCRD(definition, groupSuffix, schemaDB={}, render=defaultRender):
    local _processor =
      root.processor.new()
      + root.processor.withSchemaDB(schemaDB)
      + root.processor.withRenderEngineType(render);
    self.crd.render(definition, groupSuffix, _processor),

  // XRD: Crossplane CompositeResourceDefinition
  fromXRD(definition, groupSuffix, schemaDB={}, render=defaultRender):
    local _processor =
      root.processor.new()
      + root.processor.withSchemaDB(schemaDB)
      + root.processor.withRenderEngineType(render);
    self.xrd.render(definition, groupSuffix, _processor),

  fromOpenAPI(name, component, schema, schemaDB={}, render=defaultRender):
    if name == ''
    then error "name can't be an empty string"
    else
      local _processor =
        root.processor.new()
        + root.processor.withSchemaDB(schemaDB)
        + root.processor.withRenderEngineType(render);
      self.openapi.render(name, component, schema, _processor),

  // expects schema as rendered by `kubectl get --raw /openapi/v2`
  fromKubernetesOpenAPI(schema, render=defaultRender):
    local _processor =
      root.processor.new()
      + root.processor.withRenderEngineType(render);
    local renderEngine = _processor.renderEngine;
    std.foldl(
      function(acc, d)
        local items = std.reverse(std.split(d, '.'));
        local component = schema.definitions[d];
        local name = helpers.camelCaseKind(items[0]);
        acc
        + renderEngine.toObject(
          renderEngine.nestInParents(
            [items[2], items[1]],
            self.fromOpenAPI(name, component, schema, render=render),
          )
        ),
      std.objectFields(schema.definitions),
      renderEngine.nilvalue
    ),
}
