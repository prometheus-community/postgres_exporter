local crdsonnet = import '../main.libsonnet';

local schema = import './example_schema.json';

local lib = crdsonnet.schema.render('customer', schema);
local c = lib.customer;

local shippingAddress =
  local address = c.shipping_address;
  address.withStreetAddress('12 Church Lane')
  + address.withCity('West York')
  + address.withState('Dobberton')
  + address.withCountry();

c.withFirstName('John')
+ c.withLastName('Doe')
+ shippingAddress
+ c.withBillingAddress(shippingAddress.shipping_address)
