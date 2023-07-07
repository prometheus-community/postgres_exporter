# azureMonitor

grafonnet.query.azureMonitor

## Index

* [`fn withAzureLogAnalytics(value)`](#fn-withazureloganalytics)
* [`fn withAzureLogAnalyticsMixin(value)`](#fn-withazureloganalyticsmixin)
* [`fn withAzureMonitor(value)`](#fn-withazuremonitor)
* [`fn withAzureMonitorMixin(value)`](#fn-withazuremonitormixin)
* [`fn withAzureResourceGraph(value)`](#fn-withazureresourcegraph)
* [`fn withAzureResourceGraphMixin(value)`](#fn-withazureresourcegraphmixin)
* [`fn withAzureTraces(value)`](#fn-withazuretraces)
* [`fn withAzureTracesMixin(value)`](#fn-withazuretracesmixin)
* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withGrafanaTemplateVariableFn(value)`](#fn-withgrafanatemplatevariablefn)
* [`fn withGrafanaTemplateVariableFnMixin(value)`](#fn-withgrafanatemplatevariablefnmixin)
* [`fn withHide(value)`](#fn-withhide)
* [`fn withNamespace(value)`](#fn-withnamespace)
* [`fn withQueryType(value)`](#fn-withquerytype)
* [`fn withRefId(value)`](#fn-withrefid)
* [`fn withRegion(value)`](#fn-withregion)
* [`fn withResource(value)`](#fn-withresource)
* [`fn withResourceGroup(value)`](#fn-withresourcegroup)
* [`fn withSubscription(value)`](#fn-withsubscription)
* [`fn withSubscriptions(value)`](#fn-withsubscriptions)
* [`fn withSubscriptionsMixin(value)`](#fn-withsubscriptionsmixin)
* [`obj azureLogAnalytics`](#obj-azureloganalytics)
  * [`fn withQuery(value)`](#fn-azureloganalyticswithquery)
  * [`fn withResource(value)`](#fn-azureloganalyticswithresource)
  * [`fn withResources(value)`](#fn-azureloganalyticswithresources)
  * [`fn withResourcesMixin(value)`](#fn-azureloganalyticswithresourcesmixin)
  * [`fn withResultFormat(value)`](#fn-azureloganalyticswithresultformat)
  * [`fn withWorkspace(value)`](#fn-azureloganalyticswithworkspace)
* [`obj azureMonitor`](#obj-azuremonitor)
  * [`fn withAggregation(value)`](#fn-azuremonitorwithaggregation)
  * [`fn withAlias(value)`](#fn-azuremonitorwithalias)
  * [`fn withAllowedTimeGrainsMs(value)`](#fn-azuremonitorwithallowedtimegrainsms)
  * [`fn withAllowedTimeGrainsMsMixin(value)`](#fn-azuremonitorwithallowedtimegrainsmsmixin)
  * [`fn withCustomNamespace(value)`](#fn-azuremonitorwithcustomnamespace)
  * [`fn withDimension(value)`](#fn-azuremonitorwithdimension)
  * [`fn withDimensionFilter(value)`](#fn-azuremonitorwithdimensionfilter)
  * [`fn withDimensionFilters(value)`](#fn-azuremonitorwithdimensionfilters)
  * [`fn withDimensionFiltersMixin(value)`](#fn-azuremonitorwithdimensionfiltersmixin)
  * [`fn withMetricDefinition(value)`](#fn-azuremonitorwithmetricdefinition)
  * [`fn withMetricName(value)`](#fn-azuremonitorwithmetricname)
  * [`fn withMetricNamespace(value)`](#fn-azuremonitorwithmetricnamespace)
  * [`fn withRegion(value)`](#fn-azuremonitorwithregion)
  * [`fn withResourceGroup(value)`](#fn-azuremonitorwithresourcegroup)
  * [`fn withResourceName(value)`](#fn-azuremonitorwithresourcename)
  * [`fn withResourceUri(value)`](#fn-azuremonitorwithresourceuri)
  * [`fn withResources(value)`](#fn-azuremonitorwithresources)
  * [`fn withResourcesMixin(value)`](#fn-azuremonitorwithresourcesmixin)
  * [`fn withTimeGrain(value)`](#fn-azuremonitorwithtimegrain)
  * [`fn withTimeGrainUnit(value)`](#fn-azuremonitorwithtimegrainunit)
  * [`fn withTop(value)`](#fn-azuremonitorwithtop)
  * [`obj dimensionFilters`](#obj-azuremonitordimensionfilters)
    * [`fn withDimension(value)`](#fn-azuremonitordimensionfilterswithdimension)
    * [`fn withFilter(value)`](#fn-azuremonitordimensionfilterswithfilter)
    * [`fn withFilters(value)`](#fn-azuremonitordimensionfilterswithfilters)
    * [`fn withFiltersMixin(value)`](#fn-azuremonitordimensionfilterswithfiltersmixin)
    * [`fn withOperator(value)`](#fn-azuremonitordimensionfilterswithoperator)
  * [`obj resources`](#obj-azuremonitorresources)
    * [`fn withMetricNamespace(value)`](#fn-azuremonitorresourceswithmetricnamespace)
    * [`fn withRegion(value)`](#fn-azuremonitorresourceswithregion)
    * [`fn withResourceGroup(value)`](#fn-azuremonitorresourceswithresourcegroup)
    * [`fn withResourceName(value)`](#fn-azuremonitorresourceswithresourcename)
    * [`fn withSubscription(value)`](#fn-azuremonitorresourceswithsubscription)
* [`obj azureResourceGraph`](#obj-azureresourcegraph)
  * [`fn withQuery(value)`](#fn-azureresourcegraphwithquery)
  * [`fn withResultFormat(value)`](#fn-azureresourcegraphwithresultformat)
* [`obj azureTraces`](#obj-azuretraces)
  * [`fn withFilters(value)`](#fn-azuretraceswithfilters)
  * [`fn withFiltersMixin(value)`](#fn-azuretraceswithfiltersmixin)
  * [`fn withOperationId(value)`](#fn-azuretraceswithoperationid)
  * [`fn withQuery(value)`](#fn-azuretraceswithquery)
  * [`fn withResources(value)`](#fn-azuretraceswithresources)
  * [`fn withResourcesMixin(value)`](#fn-azuretraceswithresourcesmixin)
  * [`fn withResultFormat(value)`](#fn-azuretraceswithresultformat)
  * [`fn withTraceTypes(value)`](#fn-azuretraceswithtracetypes)
  * [`fn withTraceTypesMixin(value)`](#fn-azuretraceswithtracetypesmixin)
  * [`obj filters`](#obj-azuretracesfilters)
    * [`fn withFilters(value)`](#fn-azuretracesfilterswithfilters)
    * [`fn withFiltersMixin(value)`](#fn-azuretracesfilterswithfiltersmixin)
    * [`fn withOperation(value)`](#fn-azuretracesfilterswithoperation)
    * [`fn withProperty(value)`](#fn-azuretracesfilterswithproperty)
* [`obj grafanaTemplateVariableFn`](#obj-grafanatemplatevariablefn)
  * [`fn withAppInsightsGroupByQuery(value)`](#fn-grafanatemplatevariablefnwithappinsightsgroupbyquery)
  * [`fn withAppInsightsGroupByQueryMixin(value)`](#fn-grafanatemplatevariablefnwithappinsightsgroupbyquerymixin)
  * [`fn withAppInsightsMetricNameQuery(value)`](#fn-grafanatemplatevariablefnwithappinsightsmetricnamequery)
  * [`fn withAppInsightsMetricNameQueryMixin(value)`](#fn-grafanatemplatevariablefnwithappinsightsmetricnamequerymixin)
  * [`fn withMetricDefinitionsQuery(value)`](#fn-grafanatemplatevariablefnwithmetricdefinitionsquery)
  * [`fn withMetricDefinitionsQueryMixin(value)`](#fn-grafanatemplatevariablefnwithmetricdefinitionsquerymixin)
  * [`fn withMetricNamesQuery(value)`](#fn-grafanatemplatevariablefnwithmetricnamesquery)
  * [`fn withMetricNamesQueryMixin(value)`](#fn-grafanatemplatevariablefnwithmetricnamesquerymixin)
  * [`fn withMetricNamespaceQuery(value)`](#fn-grafanatemplatevariablefnwithmetricnamespacequery)
  * [`fn withMetricNamespaceQueryMixin(value)`](#fn-grafanatemplatevariablefnwithmetricnamespacequerymixin)
  * [`fn withResourceGroupsQuery(value)`](#fn-grafanatemplatevariablefnwithresourcegroupsquery)
  * [`fn withResourceGroupsQueryMixin(value)`](#fn-grafanatemplatevariablefnwithresourcegroupsquerymixin)
  * [`fn withResourceNamesQuery(value)`](#fn-grafanatemplatevariablefnwithresourcenamesquery)
  * [`fn withResourceNamesQueryMixin(value)`](#fn-grafanatemplatevariablefnwithresourcenamesquerymixin)
  * [`fn withSubscriptionsQuery(value)`](#fn-grafanatemplatevariablefnwithsubscriptionsquery)
  * [`fn withSubscriptionsQueryMixin(value)`](#fn-grafanatemplatevariablefnwithsubscriptionsquerymixin)
  * [`fn withUnknownQuery(value)`](#fn-grafanatemplatevariablefnwithunknownquery)
  * [`fn withUnknownQueryMixin(value)`](#fn-grafanatemplatevariablefnwithunknownquerymixin)
  * [`fn withWorkspacesQuery(value)`](#fn-grafanatemplatevariablefnwithworkspacesquery)
  * [`fn withWorkspacesQueryMixin(value)`](#fn-grafanatemplatevariablefnwithworkspacesquerymixin)
  * [`obj AppInsightsGroupByQuery`](#obj-grafanatemplatevariablefnappinsightsgroupbyquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnappinsightsgroupbyquerywithkind)
    * [`fn withMetricName(value)`](#fn-grafanatemplatevariablefnappinsightsgroupbyquerywithmetricname)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnappinsightsgroupbyquerywithrawquery)
  * [`obj AppInsightsMetricNameQuery`](#obj-grafanatemplatevariablefnappinsightsmetricnamequery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnappinsightsmetricnamequerywithkind)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnappinsightsmetricnamequerywithrawquery)
  * [`obj MetricDefinitionsQuery`](#obj-grafanatemplatevariablefnmetricdefinitionsquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithkind)
    * [`fn withMetricNamespace(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithmetricnamespace)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithrawquery)
    * [`fn withResourceGroup(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithresourcegroup)
    * [`fn withResourceName(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithresourcename)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnmetricdefinitionsquerywithsubscription)
  * [`obj MetricNamesQuery`](#obj-grafanatemplatevariablefnmetricnamesquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithkind)
    * [`fn withMetricNamespace(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithmetricnamespace)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithrawquery)
    * [`fn withResourceGroup(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithresourcegroup)
    * [`fn withResourceName(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithresourcename)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnmetricnamesquerywithsubscription)
  * [`obj MetricNamespaceQuery`](#obj-grafanatemplatevariablefnmetricnamespacequery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithkind)
    * [`fn withMetricNamespace(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithmetricnamespace)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithrawquery)
    * [`fn withResourceGroup(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithresourcegroup)
    * [`fn withResourceName(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithresourcename)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnmetricnamespacequerywithsubscription)
  * [`obj ResourceGroupsQuery`](#obj-grafanatemplatevariablefnresourcegroupsquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnresourcegroupsquerywithkind)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnresourcegroupsquerywithrawquery)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnresourcegroupsquerywithsubscription)
  * [`obj ResourceNamesQuery`](#obj-grafanatemplatevariablefnresourcenamesquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnresourcenamesquerywithkind)
    * [`fn withMetricNamespace(value)`](#fn-grafanatemplatevariablefnresourcenamesquerywithmetricnamespace)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnresourcenamesquerywithrawquery)
    * [`fn withResourceGroup(value)`](#fn-grafanatemplatevariablefnresourcenamesquerywithresourcegroup)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnresourcenamesquerywithsubscription)
  * [`obj SubscriptionsQuery`](#obj-grafanatemplatevariablefnsubscriptionsquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnsubscriptionsquerywithkind)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnsubscriptionsquerywithrawquery)
  * [`obj UnknownQuery`](#obj-grafanatemplatevariablefnunknownquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnunknownquerywithkind)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnunknownquerywithrawquery)
  * [`obj WorkspacesQuery`](#obj-grafanatemplatevariablefnworkspacesquery)
    * [`fn withKind(value)`](#fn-grafanatemplatevariablefnworkspacesquerywithkind)
    * [`fn withRawQuery(value)`](#fn-grafanatemplatevariablefnworkspacesquerywithrawquery)
    * [`fn withSubscription(value)`](#fn-grafanatemplatevariablefnworkspacesquerywithsubscription)

## Fields

### fn withAzureLogAnalytics

```ts
withAzureLogAnalytics(value)
```

Azure Monitor Logs sub-query properties

### fn withAzureLogAnalyticsMixin

```ts
withAzureLogAnalyticsMixin(value)
```

Azure Monitor Logs sub-query properties

### fn withAzureMonitor

```ts
withAzureMonitor(value)
```



### fn withAzureMonitorMixin

```ts
withAzureMonitorMixin(value)
```



### fn withAzureResourceGraph

```ts
withAzureResourceGraph(value)
```



### fn withAzureResourceGraphMixin

```ts
withAzureResourceGraphMixin(value)
```



### fn withAzureTraces

```ts
withAzureTraces(value)
```

Application Insights Traces sub-query properties

### fn withAzureTracesMixin

```ts
withAzureTracesMixin(value)
```

Application Insights Traces sub-query properties

### fn withDatasource

```ts
withDatasource(value)
```

For mixed data sources the selected datasource is on the query level.
For non mixed scenarios this is undefined.
TODO find a better way to do this ^ that's friendly to schema
TODO this shouldn't be unknown but DataSourceRef | null

### fn withGrafanaTemplateVariableFn

```ts
withGrafanaTemplateVariableFn(value)
```



### fn withGrafanaTemplateVariableFnMixin

```ts
withGrafanaTemplateVariableFnMixin(value)
```



### fn withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

### fn withNamespace

```ts
withNamespace(value)
```



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

### fn withRegion

```ts
withRegion(value)
```

Azure Monitor query type.
queryType: #AzureQueryType

### fn withResource

```ts
withResource(value)
```



### fn withResourceGroup

```ts
withResourceGroup(value)
```

Template variables params. These exist for backwards compatiblity with legacy template variables.

### fn withSubscription

```ts
withSubscription(value)
```

Azure subscription containing the resource(s) to be queried.

### fn withSubscriptions

```ts
withSubscriptions(value)
```

Subscriptions to be queried via Azure Resource Graph.

### fn withSubscriptionsMixin

```ts
withSubscriptionsMixin(value)
```

Subscriptions to be queried via Azure Resource Graph.

### obj azureLogAnalytics


#### fn azureLogAnalytics.withQuery

```ts
withQuery(value)
```

KQL query to be executed.

#### fn azureLogAnalytics.withResource

```ts
withResource(value)
```

@deprecated Use resources instead

#### fn azureLogAnalytics.withResources

```ts
withResources(value)
```

Array of resource URIs to be queried.

#### fn azureLogAnalytics.withResourcesMixin

```ts
withResourcesMixin(value)
```

Array of resource URIs to be queried.

#### fn azureLogAnalytics.withResultFormat

```ts
withResultFormat(value)
```



Accepted values for `value` are "table", "time_series", "trace"

#### fn azureLogAnalytics.withWorkspace

```ts
withWorkspace(value)
```

Workspace ID. This was removed in Grafana 8, but remains for backwards compat

### obj azureMonitor


#### fn azureMonitor.withAggregation

```ts
withAggregation(value)
```

The aggregation to be used within the query. Defaults to the primaryAggregationType defined by the metric.

#### fn azureMonitor.withAlias

```ts
withAlias(value)
```

Aliases can be set to modify the legend labels. e.g. {{ resourceGroup }}. See docs for more detail.

#### fn azureMonitor.withAllowedTimeGrainsMs

```ts
withAllowedTimeGrainsMs(value)
```

Time grains that are supported by the metric.

#### fn azureMonitor.withAllowedTimeGrainsMsMixin

```ts
withAllowedTimeGrainsMsMixin(value)
```

Time grains that are supported by the metric.

#### fn azureMonitor.withCustomNamespace

```ts
withCustomNamespace(value)
```

Used as the value for the metricNamespace property when it's different from the resource namespace.

#### fn azureMonitor.withDimension

```ts
withDimension(value)
```

@deprecated This property was migrated to dimensionFilters and should only be accessed in the migration

#### fn azureMonitor.withDimensionFilter

```ts
withDimensionFilter(value)
```

@deprecated This property was migrated to dimensionFilters and should only be accessed in the migration

#### fn azureMonitor.withDimensionFilters

```ts
withDimensionFilters(value)
```

Filters to reduce the set of data returned. Dimensions that can be filtered on are defined by the metric.

#### fn azureMonitor.withDimensionFiltersMixin

```ts
withDimensionFiltersMixin(value)
```

Filters to reduce the set of data returned. Dimensions that can be filtered on are defined by the metric.

#### fn azureMonitor.withMetricDefinition

```ts
withMetricDefinition(value)
```

@deprecated Use metricNamespace instead

#### fn azureMonitor.withMetricName

```ts
withMetricName(value)
```

The metric to query data for within the specified metricNamespace. e.g. UsedCapacity

#### fn azureMonitor.withMetricNamespace

```ts
withMetricNamespace(value)
```

metricNamespace is used as the resource type (or resource namespace).
It's usually equal to the target metric namespace. e.g. microsoft.storage/storageaccounts
Kept the name of the variable as metricNamespace to avoid backward incompatibility issues.

#### fn azureMonitor.withRegion

```ts
withRegion(value)
```

The Azure region containing the resource(s).

#### fn azureMonitor.withResourceGroup

```ts
withResourceGroup(value)
```

@deprecated Use resources instead

#### fn azureMonitor.withResourceName

```ts
withResourceName(value)
```

@deprecated Use resources instead

#### fn azureMonitor.withResourceUri

```ts
withResourceUri(value)
```

@deprecated Use resourceGroup, resourceName and metricNamespace instead

#### fn azureMonitor.withResources

```ts
withResources(value)
```

Array of resource URIs to be queried.

#### fn azureMonitor.withResourcesMixin

```ts
withResourcesMixin(value)
```

Array of resource URIs to be queried.

#### fn azureMonitor.withTimeGrain

```ts
withTimeGrain(value)
```

The granularity of data points to be queried. Defaults to auto.

#### fn azureMonitor.withTimeGrainUnit

```ts
withTimeGrainUnit(value)
```

@deprecated

#### fn azureMonitor.withTop

```ts
withTop(value)
```

Maximum number of records to return. Defaults to 10.

#### obj azureMonitor.dimensionFilters


##### fn azureMonitor.dimensionFilters.withDimension

```ts
withDimension(value)
```

Name of Dimension to be filtered on.

##### fn azureMonitor.dimensionFilters.withFilter

```ts
withFilter(value)
```

@deprecated filter is deprecated in favour of filters to support multiselect.

##### fn azureMonitor.dimensionFilters.withFilters

```ts
withFilters(value)
```

Values to match with the filter.

##### fn azureMonitor.dimensionFilters.withFiltersMixin

```ts
withFiltersMixin(value)
```

Values to match with the filter.

##### fn azureMonitor.dimensionFilters.withOperator

```ts
withOperator(value)
```

String denoting the filter operation. Supports 'eq' - equals,'ne' - not equals, 'sw' - starts with. Note that some dimensions may not support all operators.

#### obj azureMonitor.resources


##### fn azureMonitor.resources.withMetricNamespace

```ts
withMetricNamespace(value)
```



##### fn azureMonitor.resources.withRegion

```ts
withRegion(value)
```



##### fn azureMonitor.resources.withResourceGroup

```ts
withResourceGroup(value)
```



##### fn azureMonitor.resources.withResourceName

```ts
withResourceName(value)
```



##### fn azureMonitor.resources.withSubscription

```ts
withSubscription(value)
```



### obj azureResourceGraph


#### fn azureResourceGraph.withQuery

```ts
withQuery(value)
```

Azure Resource Graph KQL query to be executed.

#### fn azureResourceGraph.withResultFormat

```ts
withResultFormat(value)
```

Specifies the format results should be returned as. Defaults to table.

### obj azureTraces


#### fn azureTraces.withFilters

```ts
withFilters(value)
```

Filters for property values.

#### fn azureTraces.withFiltersMixin

```ts
withFiltersMixin(value)
```

Filters for property values.

#### fn azureTraces.withOperationId

```ts
withOperationId(value)
```

Operation ID. Used only for Traces queries.

#### fn azureTraces.withQuery

```ts
withQuery(value)
```

KQL query to be executed.

#### fn azureTraces.withResources

```ts
withResources(value)
```

Array of resource URIs to be queried.

#### fn azureTraces.withResourcesMixin

```ts
withResourcesMixin(value)
```

Array of resource URIs to be queried.

#### fn azureTraces.withResultFormat

```ts
withResultFormat(value)
```



Accepted values for `value` are "table", "time_series", "trace"

#### fn azureTraces.withTraceTypes

```ts
withTraceTypes(value)
```

Types of events to filter by.

#### fn azureTraces.withTraceTypesMixin

```ts
withTraceTypesMixin(value)
```

Types of events to filter by.

#### obj azureTraces.filters


##### fn azureTraces.filters.withFilters

```ts
withFilters(value)
```

Values to filter by.

##### fn azureTraces.filters.withFiltersMixin

```ts
withFiltersMixin(value)
```

Values to filter by.

##### fn azureTraces.filters.withOperation

```ts
withOperation(value)
```

Comparison operator to use. Either equals or not equals.

##### fn azureTraces.filters.withProperty

```ts
withProperty(value)
```

Property name, auto-populated based on available traces.

### obj grafanaTemplateVariableFn


#### fn grafanaTemplateVariableFn.withAppInsightsGroupByQuery

```ts
withAppInsightsGroupByQuery(value)
```



#### fn grafanaTemplateVariableFn.withAppInsightsGroupByQueryMixin

```ts
withAppInsightsGroupByQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withAppInsightsMetricNameQuery

```ts
withAppInsightsMetricNameQuery(value)
```



#### fn grafanaTemplateVariableFn.withAppInsightsMetricNameQueryMixin

```ts
withAppInsightsMetricNameQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withMetricDefinitionsQuery

```ts
withMetricDefinitionsQuery(value)
```

@deprecated Use MetricNamespaceQuery instead

#### fn grafanaTemplateVariableFn.withMetricDefinitionsQueryMixin

```ts
withMetricDefinitionsQueryMixin(value)
```

@deprecated Use MetricNamespaceQuery instead

#### fn grafanaTemplateVariableFn.withMetricNamesQuery

```ts
withMetricNamesQuery(value)
```



#### fn grafanaTemplateVariableFn.withMetricNamesQueryMixin

```ts
withMetricNamesQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withMetricNamespaceQuery

```ts
withMetricNamespaceQuery(value)
```



#### fn grafanaTemplateVariableFn.withMetricNamespaceQueryMixin

```ts
withMetricNamespaceQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withResourceGroupsQuery

```ts
withResourceGroupsQuery(value)
```



#### fn grafanaTemplateVariableFn.withResourceGroupsQueryMixin

```ts
withResourceGroupsQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withResourceNamesQuery

```ts
withResourceNamesQuery(value)
```



#### fn grafanaTemplateVariableFn.withResourceNamesQueryMixin

```ts
withResourceNamesQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withSubscriptionsQuery

```ts
withSubscriptionsQuery(value)
```



#### fn grafanaTemplateVariableFn.withSubscriptionsQueryMixin

```ts
withSubscriptionsQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withUnknownQuery

```ts
withUnknownQuery(value)
```



#### fn grafanaTemplateVariableFn.withUnknownQueryMixin

```ts
withUnknownQueryMixin(value)
```



#### fn grafanaTemplateVariableFn.withWorkspacesQuery

```ts
withWorkspacesQuery(value)
```



#### fn grafanaTemplateVariableFn.withWorkspacesQueryMixin

```ts
withWorkspacesQueryMixin(value)
```



#### obj grafanaTemplateVariableFn.AppInsightsGroupByQuery


##### fn grafanaTemplateVariableFn.AppInsightsGroupByQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "AppInsightsGroupByQuery"

##### fn grafanaTemplateVariableFn.AppInsightsGroupByQuery.withMetricName

```ts
withMetricName(value)
```



##### fn grafanaTemplateVariableFn.AppInsightsGroupByQuery.withRawQuery

```ts
withRawQuery(value)
```



#### obj grafanaTemplateVariableFn.AppInsightsMetricNameQuery


##### fn grafanaTemplateVariableFn.AppInsightsMetricNameQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "AppInsightsMetricNameQuery"

##### fn grafanaTemplateVariableFn.AppInsightsMetricNameQuery.withRawQuery

```ts
withRawQuery(value)
```



#### obj grafanaTemplateVariableFn.MetricDefinitionsQuery


##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "MetricDefinitionsQuery"

##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withMetricNamespace

```ts
withMetricNamespace(value)
```



##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withResourceGroup

```ts
withResourceGroup(value)
```



##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withResourceName

```ts
withResourceName(value)
```



##### fn grafanaTemplateVariableFn.MetricDefinitionsQuery.withSubscription

```ts
withSubscription(value)
```



#### obj grafanaTemplateVariableFn.MetricNamesQuery


##### fn grafanaTemplateVariableFn.MetricNamesQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "MetricNamesQuery"

##### fn grafanaTemplateVariableFn.MetricNamesQuery.withMetricNamespace

```ts
withMetricNamespace(value)
```



##### fn grafanaTemplateVariableFn.MetricNamesQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.MetricNamesQuery.withResourceGroup

```ts
withResourceGroup(value)
```



##### fn grafanaTemplateVariableFn.MetricNamesQuery.withResourceName

```ts
withResourceName(value)
```



##### fn grafanaTemplateVariableFn.MetricNamesQuery.withSubscription

```ts
withSubscription(value)
```



#### obj grafanaTemplateVariableFn.MetricNamespaceQuery


##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "MetricNamespaceQuery"

##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withMetricNamespace

```ts
withMetricNamespace(value)
```



##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withResourceGroup

```ts
withResourceGroup(value)
```



##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withResourceName

```ts
withResourceName(value)
```



##### fn grafanaTemplateVariableFn.MetricNamespaceQuery.withSubscription

```ts
withSubscription(value)
```



#### obj grafanaTemplateVariableFn.ResourceGroupsQuery


##### fn grafanaTemplateVariableFn.ResourceGroupsQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "ResourceGroupsQuery"

##### fn grafanaTemplateVariableFn.ResourceGroupsQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.ResourceGroupsQuery.withSubscription

```ts
withSubscription(value)
```



#### obj grafanaTemplateVariableFn.ResourceNamesQuery


##### fn grafanaTemplateVariableFn.ResourceNamesQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "ResourceNamesQuery"

##### fn grafanaTemplateVariableFn.ResourceNamesQuery.withMetricNamespace

```ts
withMetricNamespace(value)
```



##### fn grafanaTemplateVariableFn.ResourceNamesQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.ResourceNamesQuery.withResourceGroup

```ts
withResourceGroup(value)
```



##### fn grafanaTemplateVariableFn.ResourceNamesQuery.withSubscription

```ts
withSubscription(value)
```



#### obj grafanaTemplateVariableFn.SubscriptionsQuery


##### fn grafanaTemplateVariableFn.SubscriptionsQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "SubscriptionsQuery"

##### fn grafanaTemplateVariableFn.SubscriptionsQuery.withRawQuery

```ts
withRawQuery(value)
```



#### obj grafanaTemplateVariableFn.UnknownQuery


##### fn grafanaTemplateVariableFn.UnknownQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "UnknownQuery"

##### fn grafanaTemplateVariableFn.UnknownQuery.withRawQuery

```ts
withRawQuery(value)
```



#### obj grafanaTemplateVariableFn.WorkspacesQuery


##### fn grafanaTemplateVariableFn.WorkspacesQuery.withKind

```ts
withKind(value)
```



Accepted values for `value` are "WorkspacesQuery"

##### fn grafanaTemplateVariableFn.WorkspacesQuery.withRawQuery

```ts
withRawQuery(value)
```



##### fn grafanaTemplateVariableFn.WorkspacesQuery.withSubscription

```ts
withSubscription(value)
```


