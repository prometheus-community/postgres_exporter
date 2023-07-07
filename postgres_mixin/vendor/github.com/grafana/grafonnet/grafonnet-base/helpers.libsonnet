local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';
local xtd = import 'github.com/jsonnet-libs/xtd/main.libsonnet';

{
  local root = self,

  // Grabs last item in an array
  last(arr): xtd.array.slice(arr, -1)[0],

  // Returns the whole array except the last item
  allButLast(arr): xtd.array.slice(arr, 0, -1),

  // Gets the content from source on a specified JSONPath
  getContent(source, path):
    xtd.jsonpath.getJSONPath(source, path),

  // Sets the content on a specified ~JSONPath
  setContent(content, path):
    std.foldr(
      function(k, acc)
        { [k]+: acc },
      xtd.string.splitEscape(path, '.'),
      content
    ),

  // Hides the content in source on a specified ~JSONPath
  hideContent(source, path):
    local splitPath = xtd.string.splitEscape(path, '.');
    local content = root.getContent(source, path);
    std.foldr(
      function(k, acc)
        { [k]+: acc },
      root.allButLast(splitPath),
      { [root.last(splitPath)]:: content }
    ),

  // Removes the content in source on a specified ~JSONPath
  removeContent(source, path):
    local splitPath = xtd.string.splitEscape(path, '.');
    std.foldr(
      function(k, acc)
        { [k]+: acc },
      root.allButLast(splitPath),
      { [root.last(splitPath)]:: {} }
    ),

  // Transform moves the content from JSONPath `from` to JSONPath `to` in `source`
  transform(source, from, to):
    local content = root.getContent(source, from);
    if content == null
    then {}
    else root.setContent(content, to)
  ,

  // This functions transforms the canonical groupings representation to an array that can
  // be processed by `transform()`. Example groupings object:
  //   local groupings = {
  //     toPath: [
  //       'from.path.one',
  //       'from.path.two',
  //     ],
  //   },
  groupingsToTransformArray(groupings, keyPrefix='', keySuffix='', separator='.'):
    [
      {
        local splitFromPath = xtd.string.splitEscape(fromPath, '.'),
        local lastKey =
          keyPrefix
          + root.last(splitFromPath)
          + keySuffix,

        from:
          std.join(
            separator,
            root.allButLast(splitFromPath)
            + [lastKey],
          ),
        to:
          toPath
          + separator
          + lastKey,
      }
      for toPath in std.objectFields(groupings)
      for fromPath in groupings[toPath]
    ],

  // Transforms a groupings object from source, including their docstring counterparts
  group(source, groupings):
    std.foldl(
      function(acc, mapping)
        acc
        + root.transform(source, mapping.from, mapping.to),
      root.groupingsToTransformArray(groupings)
      + root.groupingsToTransformArray(groupings, '#'),  // also regroup docstrings
      {}
    ),

  // Transforms a groupings object from source, including their docstring counterparts
  // Regroup means it gets merged with the `source`.
  regroup(source, groupings, base=source):
    std.foldl(
      function(acc, mapping)
        acc
        + root.transform(
          source,
          mapping.from,
          mapping.to
        ),
      root.groupingsToTransformArray(groupings)
      + root.groupingsToTransformArray(groupings, '#'),  // also regroup docstrings
      base
    )
    + std.foldl(
      function(acc, mapping)
        acc
        + root.removeContent(
          source,
          mapping.from
        ),
      root.groupingsToTransformArray(groupings, '#'),
      base
    ),

  // Creates a (docs) subpackage from `source` and places it at `to`.
  makeSubpackage(source, from, to, docstring=''):
    local splitFrom = xtd.string.splitEscape(from, '.');
    local splitTo = xtd.string.splitEscape(to, '.');
    local content = {
      '#':: d.package.newSub(root.last(splitTo), docstring),
    };

    root.removeContent(  // Hide docstring on root
      source,
      std.join(
        '.',
        root.allButLast(splitFrom)
        + ['#' + root.last(splitFrom)],
      )
    )
    + root.transform(source, from, to)
    + root.setContent(content, to),

  // Repackages a field from `source` according to the mapping data.
  //   local data = [
  //     {
  //       from: 'fromPath',
  //       to: 'toPath',
  //       docstring: '',
  //     },
  //   ],
  repackage(source, data):
    std.foldl(
      function(acc, mapping)
        acc +
        root.makeSubpackage(
          source,
          mapping.from,
          mapping.to,
          mapping.docstring
        ),
      data,
      {}
    ),

  // Remove fields from `source`, `paths` is an array of path strings
  removePaths(source, paths):
    std.foldl(
      function(acc, path)
        acc + root.removeContent(source, path),
      paths,
      {}
    ),
}
