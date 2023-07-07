local crdsonnet = import '../main.libsonnet';

local schema = import './example_schema.json';

local lib = crdsonnet.schema.render('customer', schema);
local c = lib.customer;

c.withFirstName('John')
+ c.withLastName('Doe')
