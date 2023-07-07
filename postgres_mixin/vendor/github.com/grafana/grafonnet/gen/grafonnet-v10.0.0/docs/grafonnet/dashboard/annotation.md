# annotation



## Index

* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withDatasourceMixin(value)`](#fn-withdatasourcemixin)
* [`fn withEnable(value=true)`](#fn-withenable)
* [`fn withFilter(value)`](#fn-withfilter)
* [`fn withFilterMixin(value)`](#fn-withfiltermixin)
* [`fn withHide(value=false)`](#fn-withhide)
* [`fn withIconColor(value)`](#fn-withiconcolor)
* [`fn withName(value)`](#fn-withname)
* [`fn withTarget(value)`](#fn-withtarget)
* [`fn withTargetMixin(value)`](#fn-withtargetmixin)
* [`fn withType(value)`](#fn-withtype)
* [`obj datasource`](#obj-datasource)
  * [`fn withType(value)`](#fn-datasourcewithtype)
  * [`fn withUid(value)`](#fn-datasourcewithuid)
* [`obj filter`](#obj-filter)
  * [`fn withExclude(value=false)`](#fn-filterwithexclude)
  * [`fn withIds(value)`](#fn-filterwithids)
  * [`fn withIdsMixin(value)`](#fn-filterwithidsmixin)
* [`obj target`](#obj-target)
  * [`fn withLimit(value)`](#fn-targetwithlimit)
  * [`fn withMatchAny(value)`](#fn-targetwithmatchany)
  * [`fn withTags(value)`](#fn-targetwithtags)
  * [`fn withTagsMixin(value)`](#fn-targetwithtagsmixin)
  * [`fn withType(value)`](#fn-targetwithtype)

## Fields

### fn withDatasource

```ts
withDatasource(value)
```

TODO: Should be DataSourceRef

### fn withDatasourceMixin

```ts
withDatasourceMixin(value)
```

TODO: Should be DataSourceRef

### fn withEnable

```ts
withEnable(value=true)
```

When enabled the annotation query is issued with every dashboard refresh

### fn withFilter

```ts
withFilter(value)
```



### fn withFilterMixin

```ts
withFilterMixin(value)
```



### fn withHide

```ts
withHide(value=false)
```

Annotation queries can be toggled on or off at the top of the dashboard.
When hide is true, the toggle is not shown in the dashboard.

### fn withIconColor

```ts
withIconColor(value)
```

Color to use for the annotation event markers

### fn withName

```ts
withName(value)
```

Name of annotation.

### fn withTarget

```ts
withTarget(value)
```

TODO: this should be a regular DataQuery that depends on the selected dashboard
these match the properties of the "grafana" datasouce that is default in most dashboards

### fn withTargetMixin

```ts
withTargetMixin(value)
```

TODO: this should be a regular DataQuery that depends on the selected dashboard
these match the properties of the "grafana" datasouce that is default in most dashboards

### fn withType

```ts
withType(value)
```

TODO -- this should not exist here, it is based on the --grafana-- datasource

### obj datasource


#### fn datasource.withType

```ts
withType(value)
```



#### fn datasource.withUid

```ts
withUid(value)
```



### obj filter


#### fn filter.withExclude

```ts
withExclude(value=false)
```

Should the specified panels be included or excluded

#### fn filter.withIds

```ts
withIds(value)
```

Panel IDs that should be included or excluded

#### fn filter.withIdsMixin

```ts
withIdsMixin(value)
```

Panel IDs that should be included or excluded

### obj target


#### fn target.withLimit

```ts
withLimit(value)
```

Only required/valid for the grafana datasource...
but code+tests is already depending on it so hard to change

#### fn target.withMatchAny

```ts
withMatchAny(value)
```

Only required/valid for the grafana datasource...
but code+tests is already depending on it so hard to change

#### fn target.withTags

```ts
withTags(value)
```

Only required/valid for the grafana datasource...
but code+tests is already depending on it so hard to change

#### fn target.withTagsMixin

```ts
withTagsMixin(value)
```

Only required/valid for the grafana datasource...
but code+tests is already depending on it so hard to change

#### fn target.withType

```ts
withType(value)
```

Only required/valid for the grafana datasource...
but code+tests is already depending on it so hard to change
