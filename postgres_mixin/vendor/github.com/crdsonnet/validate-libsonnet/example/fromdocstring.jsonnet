local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';
local validate = import 'validate-libsonnet/main.libsonnet';

{
  '#func'::
    d.func.new(
      'sample function',
      args=[
        d.arg('num', d.T.number),
        d.arg('str', d.T.string),
        d.arg('enum', d.T.string, enums=['valid', 'values']),
      ],
    ),
  func(num, str, enum)::
    assert validate.checkParamsFromDocstring(
      [num, str, enum],
      self['#func'],
    );
    {/* do something here */ },

  return: self.func(100, 'this is a string', 'valid'),
}
