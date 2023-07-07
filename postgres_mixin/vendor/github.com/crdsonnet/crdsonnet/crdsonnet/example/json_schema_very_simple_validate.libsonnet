local crdsonnet = import '../main.libsonnet';

local schema = {
  type: 'object',
  properties: {
    name: {
      type: 'string',
    },
  },
};

local validateProcessor =
  crdsonnet.processor.new()
  + crdsonnet.processor.withValidation();

local lib = crdsonnet.schema.render('person', schema, validateProcessor);

lib.person.withName(100)  // invalid input
