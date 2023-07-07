---
permalink: /url/
---

# package url

```jsonnet
local url = import "github.com/jsonnet-libs/xtd/url.libsonnet"
```

`url` implements URL escaping and query building

## Index

* [`fn encodeQuery(params)`](#fn-encodequery)
* [`fn escapeString(str, excludedChars)`](#fn-escapestring)

## Fields

### fn encodeQuery

```ts
encodeQuery(params)
```

`encodeQuery` takes an object of query parameters and returns them as an escaped `key=value` string

### fn escapeString

```ts
escapeString(str, excludedChars)
```

`escapeString` escapes the given string so it can be safely placed inside an URL, replacing special characters with `%XX` sequences