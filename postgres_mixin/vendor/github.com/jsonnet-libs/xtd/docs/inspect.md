---
permalink: /inspect/
---

# package inspect

```jsonnet
local inspect = import "github.com/jsonnet-libs/xtd/inspect.libsonnet"
```

`inspect` implements helper functions for inspecting Jsonnet

## Index

* [`fn diff(input1, input2)`](#fn-diff)
* [`fn inspect(object, maxDepth)`](#fn-inspect)

## Fields

### fn diff

```ts
diff(input1, input2)
```

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


### fn inspect

```ts
inspect(object, maxDepth)
```

`inspect` reports the structure of a Jsonnet object with a recursion depth of
`maxDepth` (default maxDepth=10).
