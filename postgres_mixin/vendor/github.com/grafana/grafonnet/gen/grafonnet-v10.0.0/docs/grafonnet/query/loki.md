# loki

grafonnet.query.loki

## Index

* [`fn new(datasource, expr)`](#fn-new)
* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withEditorMode(value)`](#fn-witheditormode)
* [`fn withExpr(value)`](#fn-withexpr)
* [`fn withHide(value)`](#fn-withhide)
* [`fn withInstant(value)`](#fn-withinstant)
* [`fn withLegendFormat(value)`](#fn-withlegendformat)
* [`fn withMaxLines(value)`](#fn-withmaxlines)
* [`fn withQueryType(value)`](#fn-withquerytype)
* [`fn withRange(value)`](#fn-withrange)
* [`fn withRefId(value)`](#fn-withrefid)
* [`fn withResolution(value)`](#fn-withresolution)

## Fields

### fn new

```ts
new(datasource, expr)
```

Creates a new loki query target for panels.

### fn withDatasource

```ts
withDatasource(value)
```

Set the datasource for this query.

### fn withEditorMode

```ts
withEditorMode(value)
```



Accepted values for `value` are "code", "builder"

### fn withExpr

```ts
withExpr(value)
```

The LogQL query.

### fn withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

### fn withInstant

```ts
withInstant(value)
```

@deprecated, now use queryType.

### fn withLegendFormat

```ts
withLegendFormat(value)
```

Used to override the name of the series.

### fn withMaxLines

```ts
withMaxLines(value)
```

Used to limit the number of log rows returned.

### fn withQueryType

```ts
withQueryType(value)
```

Specify the query flavor
TODO make this required and give it a default

### fn withRange

```ts
withRange(value)
```

@deprecated, now use queryType.

### fn withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.

### fn withResolution

```ts
withResolution(value)
```

Used to scale the interval value.
