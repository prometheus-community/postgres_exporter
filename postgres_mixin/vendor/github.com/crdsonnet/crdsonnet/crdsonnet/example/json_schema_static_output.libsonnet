{
  customer+: {
    withBillingAddress(value): { billing_address: value },
    withBillingAddressMixin(value): { billing_address+: value },
    billing_address+: {
      withCity(value): { billing_address+: { city: value } },
      withCountry(value='United States of America'): { billing_address+: { country: value } },
      withState(value): { billing_address+: { state: value } },
      withStreetAddress(value): { billing_address+: { street_address: value } },
    },
    withFirstName(value): { first_name: value },
    withLastName(value): { last_name: value },
    withShippingAddress(value): { shipping_address: value },
    withShippingAddressMixin(value): { shipping_address+: value },
    shipping_address+: {
      withCity(value): { shipping_address+: { city: value } },
      withCountry(value='United States of America'): { shipping_address+: { country: value } },
      withState(value): { shipping_address+: { state: value } },
      withStreetAddress(value): { shipping_address+: { street_address: value } },
    },
  },
}
