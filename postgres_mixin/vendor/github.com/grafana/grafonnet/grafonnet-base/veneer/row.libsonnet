local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

function(name, panel)
  {
    '#new':: d.func.new(
      'Creates a new %s panel with a title.' % name,
      args=[d.arg('title', d.T.string)]
    ),
    new(title):
      self.withTitle(title)
      + self.withType(),
  }
