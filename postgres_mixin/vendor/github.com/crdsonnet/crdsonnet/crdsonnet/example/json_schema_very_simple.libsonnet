local crdsonnet = import '../main.libsonnet';

local schema = {
  type: 'object',
  properties: {
    name: {
      type: 'string',
    },
  },
};

local lib = crdsonnet.schema.render('person', schema);

lib.person.withName('John')
