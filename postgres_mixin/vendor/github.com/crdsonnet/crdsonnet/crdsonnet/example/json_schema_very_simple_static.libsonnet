local crdsonnet = import '../main.libsonnet';

local schema = {
  type: 'object',
  properties: {
    name: {
      type: 'string',
    },
  },
};

local staticProcessor =
  crdsonnet.processor.new()
  + crdsonnet.processor.withRenderEngineType('static');

crdsonnet.schema.render('person', schema, staticProcessor)
