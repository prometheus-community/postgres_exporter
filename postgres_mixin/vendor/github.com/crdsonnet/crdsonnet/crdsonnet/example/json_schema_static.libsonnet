local crdsonnet = import '../main.libsonnet';

local schema = import './example_schema.json';

local staticProcessor =
  crdsonnet.processor.new()
  + crdsonnet.processor.withRenderEngineType('static');

crdsonnet.schema.render('customer', schema, staticProcessor)
