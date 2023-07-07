# prometheus

grafonnet.query.prometheus

## Index

* [`fn new(datasource, expr)`](#fn-new)
* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withEditorMode(value)`](#fn-witheditormode)
* [`fn withExemplar(value)`](#fn-withexemplar)
* [`fn withExpr(value)`](#fn-withexpr)
* [`fn withFormat(value)`](#fn-withformat)
* [`fn withHide(value)`](#fn-withhide)
* [`fn withInstant(value)`](#fn-withinstant)
* [`fn withIntervalFactor(value)`](#fn-withintervalfactor)
* [`fn withLegendFormat(value)`](#fn-withlegendformat)
* [`fn withQueryType(value)`](#fn-withquerytype)
* [`fn withRange(value)`](#fn-withrange)
* [`fn withRefId(value)`](#fn-withrefid)

## Fields

### fn new

```ts
new(datasource, expr)
```

Creates a new prometheus query target for panels.

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

### fn withExemplar

```ts
withExemplar(value)
```

Execute an additional query to identify interesting raw samples relevant for the given expr

### fn withExpr

```ts
withExpr(value)
```

The actual expression/query that will be evaluated by Prometheus

### fn withFormat

```ts
withFormat(value)
```



Accepted values for `value` are "time_series", "table", "heatmap"

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

Returns only the latest value that Prometheus has scraped for the requested time series

### fn withIntervalFactor

```ts
withIntervalFactor(value)
```

Set the interval factor for this query.

### fn withLegendFormat

```ts
withLegendFormat(value)
```

Set the legend format for this query.

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

Returns a Range vector, comprised of a set of time series containing a range of data points over time for each time series

### fn withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.
