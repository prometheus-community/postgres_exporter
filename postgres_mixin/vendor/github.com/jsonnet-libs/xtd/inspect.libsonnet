local d = import 'doc-util/main.libsonnet';

{
  '#': d.pkg(
    name='inspect',
    url='github.com/jsonnet-libs/xtd/inspect.libsonnet',
    help='`inspect` implements helper functions for inspecting Jsonnet',
  ),


  '#inspect':: d.fn(
    |||
      `inspect` reports the structure of a Jsonnet object with a recursion depth of
      `maxDepth` (default maxDepth=10).
    |||,
    [
      d.arg('object', d.T.object),
      d.arg('maxDepth', d.T.number),
      //d.arg('depth', d.T.number), // used for recursion, not exposing in docs
    ]
  ),

  local this = self,
  inspect(object, maxDepth=10, depth=0):
    std.foldl(
      function(acc, p)
        acc + (
          if std.isObject(object[p])
             && depth != maxDepth
          then { [p]+:
            this.inspect(
              object[p],
              maxDepth,
              depth + 1
            ) }
          else {
            [
            (if !std.objectHas(object, p)
             then 'hidden_'
             else '')
            + (if std.isFunction(object[p])
               then 'functions'
               else 'fields')
            ]+: [p],
          }
        ),
      std.objectFieldsAll(object),
      {}
    ),

  '#diff':: d.fn(
    |||
      `diff` returns a JSON object describing the differences between two inputs. It
      attemps to show diffs in nested objects and arrays too.

      Simple example:

      ```jsonnet
      local input1 = {
        same: 'same',
        change: 'this',
        remove: 'removed',
      };

      local input2 = {
        same: 'same',
        change: 'changed',
        add: 'added',
      };

      diff(input1, input2),
      ```

      Output:
      ```json
      {
        "add +": "added",
        "change ~": "~[ this , changed ]",
        "remove -": "removed"
      }
      ```
    |||,
    [
      d.arg('input1', d.T.any),
      d.arg('input2', d.T.any),
    ]
  ),

  diff(input1, input2)::
    if input1 == input2
    then ''
    else if std.isArray(input1) && std.isArray(input2)
    then
      [
        if input1[i] != input2[i]
        then
          this.diff(
            input1[i],
            input2[i]
          )
        else input2[i]
        for i in std.range(0, std.length(input2) - 1)
        if std.length(input1) > i
      ]
      + (if std.length(input1) < std.length(input2)
         then [
           '+ ' + input2[i]
           for i in std.range(std.length(input1), std.length(input2) - 1)
         ]
         else [])
      + (if std.length(input1) > std.length(input2)
         then [
           '- ' + input1[i]
           for i in std.range(std.length(input2), std.length(input1) - 1)
         ]
         else [])

    else if std.isObject(input1) && std.isObject(input2)
    then std.foldl(
           function(acc, k)
             acc + (
               if k in input1 && input1[k] != input2[k]
               then {
                 [k + ' ~']:
                   this.diff(
                     input1[k],
                     input2[k]
                   ),
               }
               else if !(k in input1)
               then {
                 [k + ' +']: input2[k],
               }
               else {}
             ),
           std.objectFields(input2),
           {},
         )
         + {
           [l + ' -']: input1[l]
           for l in std.objectFields(input1)
           if !(l in input2)
         }

    else '~[ %s ]' % std.join(' , ', [std.toString(input1), std.toString(input2)]),
}
