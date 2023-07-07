local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

{
  '#': d.package.newSub(
    'schemaDB',
    '`schemaDB` provides an interface to describe a schemaDB.',
  ),

  '#get': d.fn(
    "`get` gets a schema from a 'db'.",
    args=[
      d.arg('db', d.T.object),
      d.arg('name', d.T.string),
    ],
  ),
  get(db, name):
    std.get(db, name, {}),

  '#add': d.fn(
    "`add` adds a schema to a 'db', expects a schema to have either am `$id` or `id` field.",
    args=[d.arg('schema', d.T.object)],
  ),
  add(schema):
    local id = self.getID(schema);
    if id == ''
    then error "Can't add schema without id"
    else { [id]: schema },

  '#getID': d.fn(
    '`getID` gets the ID from a schema, either `$id` or `id` are returned.',
    args=[d.arg('schema', d.T.object)],
  ),
  getID(schema):
    std.get(
      schema,
      '$id',
      std.get(
        schema,
        'id',
        ''
      )
    ),

}
