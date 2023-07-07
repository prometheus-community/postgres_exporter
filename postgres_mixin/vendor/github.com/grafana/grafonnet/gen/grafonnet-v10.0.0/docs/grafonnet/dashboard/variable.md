# variable

Example usage:

```jsonnet
local g = import 'g.libsonnet';
local var = g.dashboard.variable;

local customVar =
  var.custom.new(
    'myOptions',
    values=['a', 'b', 'c', 'd'],
  )
  + var.custom.generalOptions.withDescription(
    'This is a variable for my custom options.'
  )
  + var.custom.selectionOptions.withMulti();

local queryVar =
  var.query.new('queryOptions')
  + var.query.queryTypes.withLabelValues(
    'up',
    'instance',
  )
  + var.query.withDatasource(
    type='prometheus',
    uid='mimir-prod',
  )
  + var.query.selectionOptions.withIncludeAll();


g.dashboard.new('my dashboard')
+ g.dashboard.withVariables([
  customVar,
  queryVar,
])
```


## Index

* [`obj adhoc`](#obj-adhoc)
  * [`fn new(name, type, uid)`](#fn-adhocnew)
  * [`fn newFromVariable(name, variable)`](#fn-adhocnewfromvariable)
  * [`obj generalOptions`](#obj-adhocgeneraloptions)
    * [`fn withDescription(value)`](#fn-adhocgeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-adhocgeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-adhocgeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-adhocgeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-adhocgeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-adhocgeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-adhocgeneraloptionsshowondashboardwithvalueonly)
* [`obj constant`](#obj-constant)
  * [`fn new(name, value)`](#fn-constantnew)
  * [`obj generalOptions`](#obj-constantgeneraloptions)
    * [`fn withDescription(value)`](#fn-constantgeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-constantgeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-constantgeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-constantgeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-constantgeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-constantgeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-constantgeneraloptionsshowondashboardwithvalueonly)
* [`obj custom`](#obj-custom)
  * [`fn new(name, values)`](#fn-customnew)
  * [`obj generalOptions`](#obj-customgeneraloptions)
    * [`fn withDescription(value)`](#fn-customgeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-customgeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-customgeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-customgeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-customgeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-customgeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-customgeneraloptionsshowondashboardwithvalueonly)
  * [`obj selectionOptions`](#obj-customselectionoptions)
    * [`fn withIncludeAll(value=true, customAllValue)`](#fn-customselectionoptionswithincludeall)
    * [`fn withMulti(value=true)`](#fn-customselectionoptionswithmulti)
* [`obj datasource`](#obj-datasource)
  * [`fn new(name, type)`](#fn-datasourcenew)
  * [`fn withRegex(value)`](#fn-datasourcewithregex)
  * [`obj generalOptions`](#obj-datasourcegeneraloptions)
    * [`fn withDescription(value)`](#fn-datasourcegeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-datasourcegeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-datasourcegeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-datasourcegeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-datasourcegeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-datasourcegeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-datasourcegeneraloptionsshowondashboardwithvalueonly)
  * [`obj selectionOptions`](#obj-datasourceselectionoptions)
    * [`fn withIncludeAll(value=true, customAllValue)`](#fn-datasourceselectionoptionswithincludeall)
    * [`fn withMulti(value=true)`](#fn-datasourceselectionoptionswithmulti)
* [`obj interval`](#obj-interval)
  * [`fn new(name, values)`](#fn-intervalnew)
  * [`fn withAutoOption(count, minInterval)`](#fn-intervalwithautooption)
  * [`obj generalOptions`](#obj-intervalgeneraloptions)
    * [`fn withDescription(value)`](#fn-intervalgeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-intervalgeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-intervalgeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-intervalgeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-intervalgeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-intervalgeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-intervalgeneraloptionsshowondashboardwithvalueonly)
* [`obj query`](#obj-query)
  * [`fn new(name, query="")`](#fn-querynew)
  * [`fn withDatasource(type, uid)`](#fn-querywithdatasource)
  * [`fn withDatasourceFromVariable(variable)`](#fn-querywithdatasourcefromvariable)
  * [`fn withRegex(value)`](#fn-querywithregex)
  * [`fn withSort(i=0, type="alphabetical", asc=true, caseInsensitive=false)`](#fn-querywithsort)
  * [`obj generalOptions`](#obj-querygeneraloptions)
    * [`fn withDescription(value)`](#fn-querygeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-querygeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-querygeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-querygeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-querygeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-querygeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-querygeneraloptionsshowondashboardwithvalueonly)
  * [`obj queryTypes`](#obj-queryquerytypes)
    * [`fn withLabelValues(label, metric)`](#fn-queryquerytypeswithlabelvalues)
  * [`obj refresh`](#obj-queryrefresh)
    * [`fn onLoad()`](#fn-queryrefreshonload)
    * [`fn onTime()`](#fn-queryrefreshontime)
  * [`obj selectionOptions`](#obj-queryselectionoptions)
    * [`fn withIncludeAll(value=true, customAllValue)`](#fn-queryselectionoptionswithincludeall)
    * [`fn withMulti(value=true)`](#fn-queryselectionoptionswithmulti)
* [`obj textbox`](#obj-textbox)
  * [`fn new(name, default="")`](#fn-textboxnew)
  * [`obj generalOptions`](#obj-textboxgeneraloptions)
    * [`fn withDescription(value)`](#fn-textboxgeneraloptionswithdescription)
    * [`fn withLabel(value)`](#fn-textboxgeneraloptionswithlabel)
    * [`fn withName(value)`](#fn-textboxgeneraloptionswithname)
    * [`obj showOnDashboard`](#obj-textboxgeneraloptionsshowondashboard)
      * [`fn withLabelAndValue()`](#fn-textboxgeneraloptionsshowondashboardwithlabelandvalue)
      * [`fn withNothing()`](#fn-textboxgeneraloptionsshowondashboardwithnothing)
      * [`fn withValueOnly()`](#fn-textboxgeneraloptionsshowondashboardwithvalueonly)

## Fields

### obj adhoc


#### fn adhoc.new

```ts
new(name, type, uid)
```

`new` creates an adhoc template variable for datasource with `type` and `uid`.

#### fn adhoc.newFromVariable

```ts
newFromVariable(name, variable)
```

Same as `new` but selecting the datasource from another template variable.

#### obj adhoc.generalOptions


##### fn adhoc.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn adhoc.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn adhoc.generalOptions.withName

```ts
withName(value)
```



##### obj adhoc.generalOptions.showOnDashboard


###### fn adhoc.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn adhoc.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn adhoc.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



### obj constant


#### fn constant.new

```ts
new(name, value)
```

`new` creates a hidden constant template variable.

#### obj constant.generalOptions


##### fn constant.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn constant.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn constant.generalOptions.withName

```ts
withName(value)
```



##### obj constant.generalOptions.showOnDashboard


###### fn constant.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn constant.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn constant.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



### obj custom


#### fn custom.new

```ts
new(name, values)
```

`new` creates a custom template variable.

The `values` array accepts an object with key/value keys, if it's not an object
then it will be added as a string.

Example:
```
[
  { key: 'mykey', value: 'myvalue' },
  'myvalue',
  12,
]


#### obj custom.generalOptions


##### fn custom.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn custom.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn custom.generalOptions.withName

```ts
withName(value)
```



##### obj custom.generalOptions.showOnDashboard


###### fn custom.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn custom.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn custom.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



#### obj custom.selectionOptions


##### fn custom.selectionOptions.withIncludeAll

```ts
withIncludeAll(value=true, customAllValue)
```

`withIncludeAll` enables an option to include all variables.

Optionally you can set a `customAllValue`.


##### fn custom.selectionOptions.withMulti

```ts
withMulti(value=true)
```

Enable selecting multiple values.

### obj datasource


#### fn datasource.new

```ts
new(name, type)
```

`new` creates a datasource template variable.

#### fn datasource.withRegex

```ts
withRegex(value)
```

`withRegex` filter for which data source instances to choose from in the
variable value list. Example: `/^prod/`


#### obj datasource.generalOptions


##### fn datasource.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn datasource.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn datasource.generalOptions.withName

```ts
withName(value)
```



##### obj datasource.generalOptions.showOnDashboard


###### fn datasource.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn datasource.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn datasource.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



#### obj datasource.selectionOptions


##### fn datasource.selectionOptions.withIncludeAll

```ts
withIncludeAll(value=true, customAllValue)
```

`withIncludeAll` enables an option to include all variables.

Optionally you can set a `customAllValue`.


##### fn datasource.selectionOptions.withMulti

```ts
withMulti(value=true)
```

Enable selecting multiple values.

### obj interval


#### fn interval.new

```ts
new(name, values)
```

`new` creates an interval template variable.

#### fn interval.withAutoOption

```ts
withAutoOption(count, minInterval)
```

`withAutoOption` adds an options to dynamically calculate interval by dividing
time range by the count specified.

`minInterval' has to be either unit-less or end with one of the following units:
"y, M, w, d, h, m, s, ms".


#### obj interval.generalOptions


##### fn interval.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn interval.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn interval.generalOptions.withName

```ts
withName(value)
```



##### obj interval.generalOptions.showOnDashboard


###### fn interval.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn interval.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn interval.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



### obj query


#### fn query.new

```ts
new(name, query="")
```

Create a query template variable.

`query` argument is optional, this can also be set with `query.queryTypes`.


#### fn query.withDatasource

```ts
withDatasource(type, uid)
```

Select a datasource for the variable template query.

#### fn query.withDatasourceFromVariable

```ts
withDatasourceFromVariable(variable)
```

Select the datasource from another template variable.

#### fn query.withRegex

```ts
withRegex(value)
```

`withRegex` can extract part of a series name or metric node segment. Named
capture groups can be used to separate the display text and value
([see examples](https://grafana.com/docs/grafana/latest/variables/filter-variables-with-regex#filter-and-modify-using-named-text-and-value-capture-groups)).


#### fn query.withSort

```ts
withSort(i=0, type="alphabetical", asc=true, caseInsensitive=false)
```

Choose how to sort the values in the dropdown.

This can be called as `withSort(<number>) to use the integer values for each
option. If `i==0` then it will be ignored and the other arguments will take
precedence.

The numerical values are:

- 1 - Alphabetical (asc)
- 2 - Alphabetical (desc)
- 3 - Numerical (asc)
- 4 - Numerical (desc)
- 5 - Alphabetical (case-insensitive, asc)
- 6 - Alphabetical (case-insensitive, desc)


#### obj query.generalOptions


##### fn query.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn query.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn query.generalOptions.withName

```ts
withName(value)
```



##### obj query.generalOptions.showOnDashboard


###### fn query.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn query.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn query.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```



#### obj query.queryTypes


##### fn query.queryTypes.withLabelValues

```ts
withLabelValues(label, metric)
```

Construct a Prometheus template variable using `label_values()`.

#### obj query.refresh


##### fn query.refresh.onLoad

```ts
onLoad()
```

Refresh label values on dashboard load.

##### fn query.refresh.onTime

```ts
onTime()
```

Refresh label values on time range change.

#### obj query.selectionOptions


##### fn query.selectionOptions.withIncludeAll

```ts
withIncludeAll(value=true, customAllValue)
```

`withIncludeAll` enables an option to include all variables.

Optionally you can set a `customAllValue`.


##### fn query.selectionOptions.withMulti

```ts
withMulti(value=true)
```

Enable selecting multiple values.

### obj textbox


#### fn textbox.new

```ts
new(name, default="")
```

`new` creates a textbox template variable.

#### obj textbox.generalOptions


##### fn textbox.generalOptions.withDescription

```ts
withDescription(value)
```



##### fn textbox.generalOptions.withLabel

```ts
withLabel(value)
```



##### fn textbox.generalOptions.withName

```ts
withName(value)
```



##### obj textbox.generalOptions.showOnDashboard


###### fn textbox.generalOptions.showOnDashboard.withLabelAndValue

```ts
withLabelAndValue()
```



###### fn textbox.generalOptions.showOnDashboard.withNothing

```ts
withNothing()
```



###### fn textbox.generalOptions.showOnDashboard.withValueOnly

```ts
withValueOnly()
```


