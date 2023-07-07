# cloudWatch

grafonnet.query.cloudWatch

## Index

* [`obj CloudWatchAnnotationQuery`](#obj-cloudwatchannotationquery)
  * [`fn withAccountId(value)`](#fn-cloudwatchannotationquerywithaccountid)
  * [`fn withActionPrefix(value)`](#fn-cloudwatchannotationquerywithactionprefix)
  * [`fn withAlarmNamePrefix(value)`](#fn-cloudwatchannotationquerywithalarmnameprefix)
  * [`fn withDatasource(value)`](#fn-cloudwatchannotationquerywithdatasource)
  * [`fn withDimensions(value)`](#fn-cloudwatchannotationquerywithdimensions)
  * [`fn withDimensionsMixin(value)`](#fn-cloudwatchannotationquerywithdimensionsmixin)
  * [`fn withHide(value)`](#fn-cloudwatchannotationquerywithhide)
  * [`fn withMatchExact(value)`](#fn-cloudwatchannotationquerywithmatchexact)
  * [`fn withMetricName(value)`](#fn-cloudwatchannotationquerywithmetricname)
  * [`fn withNamespace(value)`](#fn-cloudwatchannotationquerywithnamespace)
  * [`fn withPeriod(value)`](#fn-cloudwatchannotationquerywithperiod)
  * [`fn withPrefixMatching(value)`](#fn-cloudwatchannotationquerywithprefixmatching)
  * [`fn withQueryMode(value)`](#fn-cloudwatchannotationquerywithquerymode)
  * [`fn withQueryType(value)`](#fn-cloudwatchannotationquerywithquerytype)
  * [`fn withRefId(value)`](#fn-cloudwatchannotationquerywithrefid)
  * [`fn withRegion(value)`](#fn-cloudwatchannotationquerywithregion)
  * [`fn withStatistic(value)`](#fn-cloudwatchannotationquerywithstatistic)
  * [`fn withStatistics(value)`](#fn-cloudwatchannotationquerywithstatistics)
  * [`fn withStatisticsMixin(value)`](#fn-cloudwatchannotationquerywithstatisticsmixin)
* [`obj CloudWatchLogsQuery`](#obj-cloudwatchlogsquery)
  * [`fn withDatasource(value)`](#fn-cloudwatchlogsquerywithdatasource)
  * [`fn withExpression(value)`](#fn-cloudwatchlogsquerywithexpression)
  * [`fn withHide(value)`](#fn-cloudwatchlogsquerywithhide)
  * [`fn withId(value)`](#fn-cloudwatchlogsquerywithid)
  * [`fn withLogGroupNames(value)`](#fn-cloudwatchlogsquerywithloggroupnames)
  * [`fn withLogGroupNamesMixin(value)`](#fn-cloudwatchlogsquerywithloggroupnamesmixin)
  * [`fn withLogGroups(value)`](#fn-cloudwatchlogsquerywithloggroups)
  * [`fn withLogGroupsMixin(value)`](#fn-cloudwatchlogsquerywithloggroupsmixin)
  * [`fn withQueryMode(value)`](#fn-cloudwatchlogsquerywithquerymode)
  * [`fn withQueryType(value)`](#fn-cloudwatchlogsquerywithquerytype)
  * [`fn withRefId(value)`](#fn-cloudwatchlogsquerywithrefid)
  * [`fn withRegion(value)`](#fn-cloudwatchlogsquerywithregion)
  * [`fn withStatsGroups(value)`](#fn-cloudwatchlogsquerywithstatsgroups)
  * [`fn withStatsGroupsMixin(value)`](#fn-cloudwatchlogsquerywithstatsgroupsmixin)
  * [`obj logGroups`](#obj-cloudwatchlogsqueryloggroups)
    * [`fn withAccountId(value)`](#fn-cloudwatchlogsqueryloggroupswithaccountid)
    * [`fn withAccountLabel(value)`](#fn-cloudwatchlogsqueryloggroupswithaccountlabel)
    * [`fn withArn(value)`](#fn-cloudwatchlogsqueryloggroupswitharn)
    * [`fn withName(value)`](#fn-cloudwatchlogsqueryloggroupswithname)
* [`obj CloudWatchMetricsQuery`](#obj-cloudwatchmetricsquery)
  * [`fn withAccountId(value)`](#fn-cloudwatchmetricsquerywithaccountid)
  * [`fn withAlias(value)`](#fn-cloudwatchmetricsquerywithalias)
  * [`fn withDatasource(value)`](#fn-cloudwatchmetricsquerywithdatasource)
  * [`fn withDimensions(value)`](#fn-cloudwatchmetricsquerywithdimensions)
  * [`fn withDimensionsMixin(value)`](#fn-cloudwatchmetricsquerywithdimensionsmixin)
  * [`fn withExpression(value)`](#fn-cloudwatchmetricsquerywithexpression)
  * [`fn withHide(value)`](#fn-cloudwatchmetricsquerywithhide)
  * [`fn withId(value)`](#fn-cloudwatchmetricsquerywithid)
  * [`fn withLabel(value)`](#fn-cloudwatchmetricsquerywithlabel)
  * [`fn withMatchExact(value)`](#fn-cloudwatchmetricsquerywithmatchexact)
  * [`fn withMetricEditorMode(value)`](#fn-cloudwatchmetricsquerywithmetriceditormode)
  * [`fn withMetricName(value)`](#fn-cloudwatchmetricsquerywithmetricname)
  * [`fn withMetricQueryType(value)`](#fn-cloudwatchmetricsquerywithmetricquerytype)
  * [`fn withNamespace(value)`](#fn-cloudwatchmetricsquerywithnamespace)
  * [`fn withPeriod(value)`](#fn-cloudwatchmetricsquerywithperiod)
  * [`fn withQueryMode(value)`](#fn-cloudwatchmetricsquerywithquerymode)
  * [`fn withQueryType(value)`](#fn-cloudwatchmetricsquerywithquerytype)
  * [`fn withRefId(value)`](#fn-cloudwatchmetricsquerywithrefid)
  * [`fn withRegion(value)`](#fn-cloudwatchmetricsquerywithregion)
  * [`fn withSql(value)`](#fn-cloudwatchmetricsquerywithsql)
  * [`fn withSqlExpression(value)`](#fn-cloudwatchmetricsquerywithsqlexpression)
  * [`fn withSqlMixin(value)`](#fn-cloudwatchmetricsquerywithsqlmixin)
  * [`fn withStatistic(value)`](#fn-cloudwatchmetricsquerywithstatistic)
  * [`fn withStatistics(value)`](#fn-cloudwatchmetricsquerywithstatistics)
  * [`fn withStatisticsMixin(value)`](#fn-cloudwatchmetricsquerywithstatisticsmixin)
  * [`obj sql`](#obj-cloudwatchmetricsquerysql)
    * [`fn withFrom(value)`](#fn-cloudwatchmetricsquerysqlwithfrom)
    * [`fn withFromMixin(value)`](#fn-cloudwatchmetricsquerysqlwithfrommixin)
    * [`fn withGroupBy(value)`](#fn-cloudwatchmetricsquerysqlwithgroupby)
    * [`fn withGroupByMixin(value)`](#fn-cloudwatchmetricsquerysqlwithgroupbymixin)
    * [`fn withLimit(value)`](#fn-cloudwatchmetricsquerysqlwithlimit)
    * [`fn withOrderBy(value)`](#fn-cloudwatchmetricsquerysqlwithorderby)
    * [`fn withOrderByDirection(value)`](#fn-cloudwatchmetricsquerysqlwithorderbydirection)
    * [`fn withOrderByMixin(value)`](#fn-cloudwatchmetricsquerysqlwithorderbymixin)
    * [`fn withSelect(value)`](#fn-cloudwatchmetricsquerysqlwithselect)
    * [`fn withSelectMixin(value)`](#fn-cloudwatchmetricsquerysqlwithselectmixin)
    * [`fn withWhere(value)`](#fn-cloudwatchmetricsquerysqlwithwhere)
    * [`fn withWhereMixin(value)`](#fn-cloudwatchmetricsquerysqlwithwheremixin)
    * [`obj from`](#obj-cloudwatchmetricsquerysqlfrom)
      * [`fn withQueryEditorFunctionExpression(value)`](#fn-cloudwatchmetricsquerysqlfromwithqueryeditorfunctionexpression)
      * [`fn withQueryEditorFunctionExpressionMixin(value)`](#fn-cloudwatchmetricsquerysqlfromwithqueryeditorfunctionexpressionmixin)
      * [`fn withQueryEditorPropertyExpression(value)`](#fn-cloudwatchmetricsquerysqlfromwithqueryeditorpropertyexpression)
      * [`fn withQueryEditorPropertyExpressionMixin(value)`](#fn-cloudwatchmetricsquerysqlfromwithqueryeditorpropertyexpressionmixin)
      * [`obj QueryEditorFunctionExpression`](#obj-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpression)
        * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionwithname)
        * [`fn withParameters(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionwithparameters)
        * [`fn withParametersMixin(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionwithparametersmixin)
        * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionwithtype)
        * [`obj parameters`](#obj-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionparameters)
          * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionparameterswithname)
          * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorfunctionexpressionparameterswithtype)
      * [`obj QueryEditorPropertyExpression`](#obj-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpression)
        * [`fn withProperty(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionwithproperty)
        * [`fn withPropertyMixin(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionwithpropertymixin)
        * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionwithtype)
        * [`obj property`](#obj-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionproperty)
          * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionpropertywithname)
          * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlfromqueryeditorpropertyexpressionpropertywithtype)
    * [`obj groupBy`](#obj-cloudwatchmetricsquerysqlgroupby)
      * [`fn withExpressions(value)`](#fn-cloudwatchmetricsquerysqlgroupbywithexpressions)
      * [`fn withExpressionsMixin(value)`](#fn-cloudwatchmetricsquerysqlgroupbywithexpressionsmixin)
      * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlgroupbywithtype)
    * [`obj orderBy`](#obj-cloudwatchmetricsquerysqlorderby)
      * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlorderbywithname)
      * [`fn withParameters(value)`](#fn-cloudwatchmetricsquerysqlorderbywithparameters)
      * [`fn withParametersMixin(value)`](#fn-cloudwatchmetricsquerysqlorderbywithparametersmixin)
      * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlorderbywithtype)
      * [`obj parameters`](#obj-cloudwatchmetricsquerysqlorderbyparameters)
        * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlorderbyparameterswithname)
        * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlorderbyparameterswithtype)
    * [`obj select`](#obj-cloudwatchmetricsquerysqlselect)
      * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlselectwithname)
      * [`fn withParameters(value)`](#fn-cloudwatchmetricsquerysqlselectwithparameters)
      * [`fn withParametersMixin(value)`](#fn-cloudwatchmetricsquerysqlselectwithparametersmixin)
      * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlselectwithtype)
      * [`obj parameters`](#obj-cloudwatchmetricsquerysqlselectparameters)
        * [`fn withName(value)`](#fn-cloudwatchmetricsquerysqlselectparameterswithname)
        * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlselectparameterswithtype)
    * [`obj where`](#obj-cloudwatchmetricsquerysqlwhere)
      * [`fn withExpressions(value)`](#fn-cloudwatchmetricsquerysqlwherewithexpressions)
      * [`fn withExpressionsMixin(value)`](#fn-cloudwatchmetricsquerysqlwherewithexpressionsmixin)
      * [`fn withType(value)`](#fn-cloudwatchmetricsquerysqlwherewithtype)

## Fields

### obj CloudWatchAnnotationQuery


#### fn CloudWatchAnnotationQuery.withAccountId

```ts
withAccountId(value)
```

The ID of the AWS account to query for the metric, specifying `all` will query all accounts that the monitoring account is permitted to query.

#### fn CloudWatchAnnotationQuery.withActionPrefix

```ts
withActionPrefix(value)
```

Use this parameter to filter the results of the operation to only those alarms
that use a certain alarm action. For example, you could specify the ARN of
an SNS topic to find all alarms that send notifications to that topic.
e.g. `arn:aws:sns:us-east-1:123456789012:my-app-` would match `arn:aws:sns:us-east-1:123456789012:my-app-action`
but not match `arn:aws:sns:us-east-1:123456789012:your-app-action`

#### fn CloudWatchAnnotationQuery.withAlarmNamePrefix

```ts
withAlarmNamePrefix(value)
```

An alarm name prefix. If you specify this parameter, you receive information
about all alarms that have names that start with this prefix.
e.g. `my-team-service-` would match `my-team-service-high-cpu` but not match `your-team-service-high-cpu`

#### fn CloudWatchAnnotationQuery.withDatasource

```ts
withDatasource(value)
```

For mixed data sources the selected datasource is on the query level.
For non mixed scenarios this is undefined.
TODO find a better way to do this ^ that's friendly to schema
TODO this shouldn't be unknown but DataSourceRef | null

#### fn CloudWatchAnnotationQuery.withDimensions

```ts
withDimensions(value)
```

A name/value pair that is part of the identity of a metric. For example, you can get statistics for a specific EC2 instance by specifying the InstanceId dimension when you search for metrics.

#### fn CloudWatchAnnotationQuery.withDimensionsMixin

```ts
withDimensionsMixin(value)
```

A name/value pair that is part of the identity of a metric. For example, you can get statistics for a specific EC2 instance by specifying the InstanceId dimension when you search for metrics.

#### fn CloudWatchAnnotationQuery.withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

#### fn CloudWatchAnnotationQuery.withMatchExact

```ts
withMatchExact(value)
```

Only show metrics that exactly match all defined dimension names.

#### fn CloudWatchAnnotationQuery.withMetricName

```ts
withMetricName(value)
```

Name of the metric

#### fn CloudWatchAnnotationQuery.withNamespace

```ts
withNamespace(value)
```

A namespace is a container for CloudWatch metrics. Metrics in different namespaces are isolated from each other, so that metrics from different applications are not mistakenly aggregated into the same statistics. For example, Amazon EC2 uses the AWS/EC2 namespace.

#### fn CloudWatchAnnotationQuery.withPeriod

```ts
withPeriod(value)
```

The length of time associated with a specific Amazon CloudWatch statistic. Can be specified by a number of seconds, 'auto', or as a duration string e.g. '15m' being 15 minutes

#### fn CloudWatchAnnotationQuery.withPrefixMatching

```ts
withPrefixMatching(value)
```

Enable matching on the prefix of the action name or alarm name, specify the prefixes with actionPrefix and/or alarmNamePrefix

#### fn CloudWatchAnnotationQuery.withQueryMode

```ts
withQueryMode(value)
```



Accepted values for `value` are "Metrics", "Logs", "Annotations"

#### fn CloudWatchAnnotationQuery.withQueryType

```ts
withQueryType(value)
```

Specify the query flavor
TODO make this required and give it a default

#### fn CloudWatchAnnotationQuery.withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.

#### fn CloudWatchAnnotationQuery.withRegion

```ts
withRegion(value)
```

AWS region to query for the metric

#### fn CloudWatchAnnotationQuery.withStatistic

```ts
withStatistic(value)
```

Metric data aggregations over specified periods of time. For detailed definitions of the statistics supported by CloudWatch, see https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html.

#### fn CloudWatchAnnotationQuery.withStatistics

```ts
withStatistics(value)
```

@deprecated use statistic

#### fn CloudWatchAnnotationQuery.withStatisticsMixin

```ts
withStatisticsMixin(value)
```

@deprecated use statistic

### obj CloudWatchLogsQuery


#### fn CloudWatchLogsQuery.withDatasource

```ts
withDatasource(value)
```

For mixed data sources the selected datasource is on the query level.
For non mixed scenarios this is undefined.
TODO find a better way to do this ^ that's friendly to schema
TODO this shouldn't be unknown but DataSourceRef | null

#### fn CloudWatchLogsQuery.withExpression

```ts
withExpression(value)
```

The CloudWatch Logs Insights query to execute

#### fn CloudWatchLogsQuery.withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

#### fn CloudWatchLogsQuery.withId

```ts
withId(value)
```



#### fn CloudWatchLogsQuery.withLogGroupNames

```ts
withLogGroupNames(value)
```

@deprecated use logGroups

#### fn CloudWatchLogsQuery.withLogGroupNamesMixin

```ts
withLogGroupNamesMixin(value)
```

@deprecated use logGroups

#### fn CloudWatchLogsQuery.withLogGroups

```ts
withLogGroups(value)
```

Log groups to query

#### fn CloudWatchLogsQuery.withLogGroupsMixin

```ts
withLogGroupsMixin(value)
```

Log groups to query

#### fn CloudWatchLogsQuery.withQueryMode

```ts
withQueryMode(value)
```



Accepted values for `value` are "Metrics", "Logs", "Annotations"

#### fn CloudWatchLogsQuery.withQueryType

```ts
withQueryType(value)
```

Specify the query flavor
TODO make this required and give it a default

#### fn CloudWatchLogsQuery.withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.

#### fn CloudWatchLogsQuery.withRegion

```ts
withRegion(value)
```

AWS region to query for the logs

#### fn CloudWatchLogsQuery.withStatsGroups

```ts
withStatsGroups(value)
```

Fields to group the results by, this field is automatically populated whenever the query is updated

#### fn CloudWatchLogsQuery.withStatsGroupsMixin

```ts
withStatsGroupsMixin(value)
```

Fields to group the results by, this field is automatically populated whenever the query is updated

#### obj CloudWatchLogsQuery.logGroups


##### fn CloudWatchLogsQuery.logGroups.withAccountId

```ts
withAccountId(value)
```

AccountId of the log group

##### fn CloudWatchLogsQuery.logGroups.withAccountLabel

```ts
withAccountLabel(value)
```

Label of the log group

##### fn CloudWatchLogsQuery.logGroups.withArn

```ts
withArn(value)
```

ARN of the log group

##### fn CloudWatchLogsQuery.logGroups.withName

```ts
withName(value)
```

Name of the log group

### obj CloudWatchMetricsQuery


#### fn CloudWatchMetricsQuery.withAccountId

```ts
withAccountId(value)
```

The ID of the AWS account to query for the metric, specifying `all` will query all accounts that the monitoring account is permitted to query.

#### fn CloudWatchMetricsQuery.withAlias

```ts
withAlias(value)
```

Deprecated: use label
@deprecated use label

#### fn CloudWatchMetricsQuery.withDatasource

```ts
withDatasource(value)
```

For mixed data sources the selected datasource is on the query level.
For non mixed scenarios this is undefined.
TODO find a better way to do this ^ that's friendly to schema
TODO this shouldn't be unknown but DataSourceRef | null

#### fn CloudWatchMetricsQuery.withDimensions

```ts
withDimensions(value)
```

A name/value pair that is part of the identity of a metric. For example, you can get statistics for a specific EC2 instance by specifying the InstanceId dimension when you search for metrics.

#### fn CloudWatchMetricsQuery.withDimensionsMixin

```ts
withDimensionsMixin(value)
```

A name/value pair that is part of the identity of a metric. For example, you can get statistics for a specific EC2 instance by specifying the InstanceId dimension when you search for metrics.

#### fn CloudWatchMetricsQuery.withExpression

```ts
withExpression(value)
```

Math expression query

#### fn CloudWatchMetricsQuery.withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

#### fn CloudWatchMetricsQuery.withId

```ts
withId(value)
```

ID can be used to reference other queries in math expressions. The ID can include numbers, letters, and underscore, and must start with a lowercase letter.

#### fn CloudWatchMetricsQuery.withLabel

```ts
withLabel(value)
```

Change the time series legend names using dynamic labels. See https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/graph-dynamic-labels.html for more details.

#### fn CloudWatchMetricsQuery.withMatchExact

```ts
withMatchExact(value)
```

Only show metrics that exactly match all defined dimension names.

#### fn CloudWatchMetricsQuery.withMetricEditorMode

```ts
withMetricEditorMode(value)
```



Accepted values for `value` are 0, 1

#### fn CloudWatchMetricsQuery.withMetricName

```ts
withMetricName(value)
```

Name of the metric

#### fn CloudWatchMetricsQuery.withMetricQueryType

```ts
withMetricQueryType(value)
```



Accepted values for `value` are 0, 1

#### fn CloudWatchMetricsQuery.withNamespace

```ts
withNamespace(value)
```

A namespace is a container for CloudWatch metrics. Metrics in different namespaces are isolated from each other, so that metrics from different applications are not mistakenly aggregated into the same statistics. For example, Amazon EC2 uses the AWS/EC2 namespace.

#### fn CloudWatchMetricsQuery.withPeriod

```ts
withPeriod(value)
```

The length of time associated with a specific Amazon CloudWatch statistic. Can be specified by a number of seconds, 'auto', or as a duration string e.g. '15m' being 15 minutes

#### fn CloudWatchMetricsQuery.withQueryMode

```ts
withQueryMode(value)
```



Accepted values for `value` are "Metrics", "Logs", "Annotations"

#### fn CloudWatchMetricsQuery.withQueryType

```ts
withQueryType(value)
```

Specify the query flavor
TODO make this required and give it a default

#### fn CloudWatchMetricsQuery.withRefId

```ts
withRefId(value)
```

A unique identifier for the query within the list of targets.
In server side expressions, the refId is used as a variable name to identify results.
By default, the UI will assign A->Z; however setting meaningful names may be useful.

#### fn CloudWatchMetricsQuery.withRegion

```ts
withRegion(value)
```

AWS region to query for the metric

#### fn CloudWatchMetricsQuery.withSql

```ts
withSql(value)
```



#### fn CloudWatchMetricsQuery.withSqlExpression

```ts
withSqlExpression(value)
```

When the metric query type is `metricQueryType` is set to `Query`, this field is used to specify the query string.

#### fn CloudWatchMetricsQuery.withSqlMixin

```ts
withSqlMixin(value)
```



#### fn CloudWatchMetricsQuery.withStatistic

```ts
withStatistic(value)
```

Metric data aggregations over specified periods of time. For detailed definitions of the statistics supported by CloudWatch, see https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html.

#### fn CloudWatchMetricsQuery.withStatistics

```ts
withStatistics(value)
```

@deprecated use statistic

#### fn CloudWatchMetricsQuery.withStatisticsMixin

```ts
withStatisticsMixin(value)
```

@deprecated use statistic

#### obj CloudWatchMetricsQuery.sql


##### fn CloudWatchMetricsQuery.sql.withFrom

```ts
withFrom(value)
```

FROM part of the SQL expression

##### fn CloudWatchMetricsQuery.sql.withFromMixin

```ts
withFromMixin(value)
```

FROM part of the SQL expression

##### fn CloudWatchMetricsQuery.sql.withGroupBy

```ts
withGroupBy(value)
```



##### fn CloudWatchMetricsQuery.sql.withGroupByMixin

```ts
withGroupByMixin(value)
```



##### fn CloudWatchMetricsQuery.sql.withLimit

```ts
withLimit(value)
```

LIMIT part of the SQL expression

##### fn CloudWatchMetricsQuery.sql.withOrderBy

```ts
withOrderBy(value)
```



##### fn CloudWatchMetricsQuery.sql.withOrderByDirection

```ts
withOrderByDirection(value)
```

The sort order of the SQL expression, `ASC` or `DESC`

##### fn CloudWatchMetricsQuery.sql.withOrderByMixin

```ts
withOrderByMixin(value)
```



##### fn CloudWatchMetricsQuery.sql.withSelect

```ts
withSelect(value)
```



##### fn CloudWatchMetricsQuery.sql.withSelectMixin

```ts
withSelectMixin(value)
```



##### fn CloudWatchMetricsQuery.sql.withWhere

```ts
withWhere(value)
```



##### fn CloudWatchMetricsQuery.sql.withWhereMixin

```ts
withWhereMixin(value)
```



##### obj CloudWatchMetricsQuery.sql.from


###### fn CloudWatchMetricsQuery.sql.from.withQueryEditorFunctionExpression

```ts
withQueryEditorFunctionExpression(value)
```



###### fn CloudWatchMetricsQuery.sql.from.withQueryEditorFunctionExpressionMixin

```ts
withQueryEditorFunctionExpressionMixin(value)
```



###### fn CloudWatchMetricsQuery.sql.from.withQueryEditorPropertyExpression

```ts
withQueryEditorPropertyExpression(value)
```



###### fn CloudWatchMetricsQuery.sql.from.withQueryEditorPropertyExpressionMixin

```ts
withQueryEditorPropertyExpressionMixin(value)
```



###### obj CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression


####### fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.withName

```ts
withName(value)
```



####### fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.withParameters

```ts
withParameters(value)
```



####### fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.withParametersMixin

```ts
withParametersMixin(value)
```



####### fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.withType

```ts
withType(value)
```



####### obj CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.parameters


######## fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.parameters.withName

```ts
withName(value)
```



######## fn CloudWatchMetricsQuery.sql.from.QueryEditorFunctionExpression.parameters.withType

```ts
withType(value)
```



###### obj CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression


####### fn CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.withProperty

```ts
withProperty(value)
```



####### fn CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.withPropertyMixin

```ts
withPropertyMixin(value)
```



####### fn CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.withType

```ts
withType(value)
```



####### obj CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.property


######## fn CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.property.withName

```ts
withName(value)
```



######## fn CloudWatchMetricsQuery.sql.from.QueryEditorPropertyExpression.property.withType

```ts
withType(value)
```



Accepted values for `value` are "string"

##### obj CloudWatchMetricsQuery.sql.groupBy


###### fn CloudWatchMetricsQuery.sql.groupBy.withExpressions

```ts
withExpressions(value)
```



###### fn CloudWatchMetricsQuery.sql.groupBy.withExpressionsMixin

```ts
withExpressionsMixin(value)
```



###### fn CloudWatchMetricsQuery.sql.groupBy.withType

```ts
withType(value)
```



Accepted values for `value` are "and", "or"

##### obj CloudWatchMetricsQuery.sql.orderBy


###### fn CloudWatchMetricsQuery.sql.orderBy.withName

```ts
withName(value)
```



###### fn CloudWatchMetricsQuery.sql.orderBy.withParameters

```ts
withParameters(value)
```



###### fn CloudWatchMetricsQuery.sql.orderBy.withParametersMixin

```ts
withParametersMixin(value)
```



###### fn CloudWatchMetricsQuery.sql.orderBy.withType

```ts
withType(value)
```



###### obj CloudWatchMetricsQuery.sql.orderBy.parameters


####### fn CloudWatchMetricsQuery.sql.orderBy.parameters.withName

```ts
withName(value)
```



####### fn CloudWatchMetricsQuery.sql.orderBy.parameters.withType

```ts
withType(value)
```



##### obj CloudWatchMetricsQuery.sql.select


###### fn CloudWatchMetricsQuery.sql.select.withName

```ts
withName(value)
```



###### fn CloudWatchMetricsQuery.sql.select.withParameters

```ts
withParameters(value)
```



###### fn CloudWatchMetricsQuery.sql.select.withParametersMixin

```ts
withParametersMixin(value)
```



###### fn CloudWatchMetricsQuery.sql.select.withType

```ts
withType(value)
```



###### obj CloudWatchMetricsQuery.sql.select.parameters


####### fn CloudWatchMetricsQuery.sql.select.parameters.withName

```ts
withName(value)
```



####### fn CloudWatchMetricsQuery.sql.select.parameters.withType

```ts
withType(value)
```



##### obj CloudWatchMetricsQuery.sql.where


###### fn CloudWatchMetricsQuery.sql.where.withExpressions

```ts
withExpressions(value)
```



###### fn CloudWatchMetricsQuery.sql.where.withExpressionsMixin

```ts
withExpressionsMixin(value)
```



###### fn CloudWatchMetricsQuery.sql.where.withType

```ts
withType(value)
```



Accepted values for `value` are "and", "or"
