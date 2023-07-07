local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

// The `anno` argument should match `dashboard.annotations.list`
function(anno)
  anno {
    '#':: d.package.newSub(
      'annotation',
      '',
    ),

    // TODO: provide API that matches the UI
  }
