local helpers = import '../helpers.libsonnet';
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';

// The `link` argument should match `dashboard.links`
function(link) {

  '#':: d.package.newSub(
    'link',
    '',
  ),

  local groupings = {
    options: [
      'withAsDropdown',
      'withKeepTime',
      'withIncludeVars',
      'withTargetBlank',
    ],
    linkOptions: [
      'withTooltip',
      'withIcon',
    ],
  },

  local grouped = helpers.group(link, groupings),

  dashboards:
    {
      '#new':: d.func.new(
        |||
          Create links to dashboards based on `tags`.
        |||,
        args=[
          d.arg('title', d.T.string),
          d.arg('tags', d.T.array),
        ]
      ),
      new(title, tags):
        link.withTitle(title)
        + link.withType('dashboards')
        + link.withTags(tags),

      options: grouped.options,
    },

  link:
    grouped.linkOptions {
      '#new':: d.func.new(
        |||
          Create link to an arbitrary URL.
        |||,
        args=[
          d.arg('title', d.T.string),
          d.arg('url', d.T.string),
        ]
      ),
      new(title, url):
        link.withTitle(title)
        + link.withType('link')
        + link.withUrl(url),

      options: grouped.options,
    },

}
