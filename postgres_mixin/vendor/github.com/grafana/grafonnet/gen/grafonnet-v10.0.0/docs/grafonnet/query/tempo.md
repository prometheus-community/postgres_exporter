# tempo

grafonnet.query.tempo

## Index

* [`fn new(datasource, query, filters)`](#fn-new)
* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withFilters(value)`](#fn-withfilters)
* [`fn withFiltersMixin(value)`](#fn-withfiltersmixin)
* [`fn withHide(value)`](#fn-withhide)
* [`fn withLimit(value)`](#fn-withlimit)
* [`fn withMaxDuration(value)`](#fn-withmaxduration)
* [`fn withMinDuration(value)`](#fn-withminduration)
* [`fn withQuery(value)`](#fn-withquery)
* [`fn withQueryType(value)`](#fn-withquerytype)
* [`fn withRefId(value)`](#fn-withrefid)
* [`fn withSearch(value)`](#fn-withsearch)
* [`fn withServiceMapQuery(value)`](#fn-withservicemapquery)
* [`fn withServiceName(value)`](#fn-withservicename)
* [`fn withSpanName(value)`](#fn-withspanname)
* [`obj filters`](#obj-filters)
  * [`fn withId(value)`](#fn-filterswithid)
  * [`fn withOperator(value)`](#fn-filterswithoperator)
  * [`fn withScope(value)`](#fn-filterswithscope)
  * [`fn withTag(value)`](#fn-filterswithtag)
  * [`fn withValue(value)`](#fn-filterswithvalue)
  * [`fn withValueMixin(value)`](#fn-filterswithvaluemixin)
  * [`fn withValueType(value)`](#fn-filterswithvaluetype)

## Fields

### fn new

```ts
new(datasource, query, filters)
```

Creates a new tempo query target for panels.

### fn withDatasource

```ts
withDatasource(value)
```

Set the datasource for this query.

### fn withFilters

```ts
withFilters(value)
```



### fn withFiltersMixin

```ts
withFiltersMixin(value)
```



### fn withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

### fn withLimit

```ts
withLimit(value)
```

Defines the maximum number of traces that are returned from Tempo

### fn withMaxDuration

```ts
withMaxDuration(value)
```

Define the maximum duration to select traces. Use duration format, for example: 1.2s, 100ms

### fn withMinDuration

```ts
withMinDuration(value)
```

Define the minimum duration to select traces. Use duration format, for example: 1.2s, 100ms

### fn withQuery

```ts
withQuery(value)
```

TraceQL query or trace ID

### fn withQueryType

```ts
withQueryType(value)
```

Specify the query flavor
TODO make this required and give it a default

### fn withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.

### fn withSearch

```ts
withSearch(value)
```

Logfmt query to filter traces by their tags. Example: http.status_code=200 error=true

### fn withServiceMapQuery

```ts
withServiceMapQuery(value)
```

Filters to be included in a PromQL query to select data for the service graph. Example: {client="app",service="app"}

### fn withServiceName

```ts
withServiceName(value)
```

Query traces by service name

### fn withSpanName

```ts
withSpanName(value)
```

Query traces by span name

### obj filters


#### fn filters.withId

```ts
withId(value)
```

Uniquely identify the filter, will not be used in the query generation

#### fn filters.withOperator

```ts
withOperator(value)
```

The operator that connects the tag to the value, for example: =, >, !=, =~

#### fn filters.withScope

```ts
withScope(value)
```

static fields are pre-set in the UI, dynamic fields are added by the user

Accepted values for `value` are "unscoped", "resource", "span"

#### fn filters.withTag

```ts
withTag(value)
```

The tag for the search filter, for example: .http.status_code, .service.name, status

#### fn filters.withValue

```ts
withValue(value)
```

The value for the search filter

#### fn filters.withValueMixin

```ts
withValueMixin(value)
```

The value for the search filter

#### fn filters.withValueType

```ts
withValueType(value)
```

The type of the value, used for example to check whether we need to wrap the value in quotes when generating the query
