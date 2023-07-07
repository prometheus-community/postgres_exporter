local xtd = import 'github.com/jsonnet-libs/xtd/main.libsonnet';

{
  camelCaseKind(kind):
    local s = xtd.camelcase.split(kind);
    std.asciiLower(s[0]) + std.join('', s[1:]),

  getGroupKey(group, suffix):
    // If no dedicated API group, then use 'nogroup' key for consistency
    if suffix == group
    then 'nogroup'
    else std.split(std.strReplace(
      group,
      '.' + suffix,
      ''
    ), '.')[0],

  metadataRefSchemaDB: {
    'https://objectmeta/schema': import 'objectmeta.json',
  },

  properties: {
    withMetadataRef(): {
      // foldStart
      properties+: {
        metadata+: {
          '$ref': 'https://objectmeta/schema#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta',
        },
      },
    },
    // foldEnd

    withGroupVersionKind(group, version, kind): {
      // foldStart
      properties+: {
        apiVersion+: {
          const:
            if group == ''
            then version
            else group
                 + '/'
                 + version,
        },
        kind+: {
          const: kind,
        },
      },
    },
    // foldEnd

    withCompositeResource(): {
      // foldStart
      properties+: {
        spec+: {
          properties+: {
            compositionRef: {
              properties: { name: { type: 'string' } },
              required: ['name'],
              type: 'object',
            },
            compositionRevisionRef: {
              properties: { name: { type: 'string' } },
              required: ['name'],
              type: 'object',
            },
            compositionSelector: {
              properties: {
                matchLabels: {
                  additionalProperties: { type: 'string' },
                  type: 'object',
                },
              },
              required: ['matchLabels'],
              type: 'object',
            },
            compositionUpdatePolicy: {
              enum: [
                'Automatic',
                'Manual',
              ],
              type: 'string',
            },
            writeConnectionSecretToRef: {
              properties: { name: { type: 'string' } },
              required: ['name'],
              type: 'object',
            },
          },
        },
      },
    },
    // foldEnd
  },
}

// vim: foldmethod=marker foldmarker=foldStart,foldEnd foldlevel=0
