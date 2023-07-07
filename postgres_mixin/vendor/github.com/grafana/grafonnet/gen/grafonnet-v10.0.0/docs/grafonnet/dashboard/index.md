# dashboard

grafonnet.dashboard

## Subpackages

* [annotation](annotation.md)
* [link](link.md)
* [variable](variable.md)

## Index

* [`fn new(title)`](#fn-new)
* [`fn withAnnotations(value)`](#fn-withannotations)
* [`fn withAnnotationsMixin(value)`](#fn-withannotationsmixin)
* [`fn withDescription(value)`](#fn-withdescription)
* [`fn withEditable(value=true)`](#fn-witheditable)
* [`fn withFiscalYearStartMonth(value=0)`](#fn-withfiscalyearstartmonth)
* [`fn withLinks(value)`](#fn-withlinks)
* [`fn withLinksMixin(value)`](#fn-withlinksmixin)
* [`fn withLiveNow(value)`](#fn-withlivenow)
* [`fn withPanels(value)`](#fn-withpanels)
* [`fn withPanelsMixin(value)`](#fn-withpanelsmixin)
* [`fn withRefresh(value)`](#fn-withrefresh)
* [`fn withRefreshMixin(value)`](#fn-withrefreshmixin)
* [`fn withSchemaVersion(value=36)`](#fn-withschemaversion)
* [`fn withStyle(value="dark")`](#fn-withstyle)
* [`fn withTags(value)`](#fn-withtags)
* [`fn withTagsMixin(value)`](#fn-withtagsmixin)
* [`fn withTemplating(value)`](#fn-withtemplating)
* [`fn withTemplatingMixin(value)`](#fn-withtemplatingmixin)
* [`fn withTimezone(value="browser")`](#fn-withtimezone)
* [`fn withTitle(value)`](#fn-withtitle)
* [`fn withUid(value)`](#fn-withuid)
* [`fn withVariables(value)`](#fn-withvariables)
* [`fn withVariablesMixin(value)`](#fn-withvariablesmixin)
* [`fn withWeekStart(value)`](#fn-withweekstart)
* [`obj graphTooltip`](#obj-graphtooltip)
  * [`fn withSharedCrosshair()`](#fn-graphtooltipwithsharedcrosshair)
  * [`fn withSharedTooltip()`](#fn-graphtooltipwithsharedtooltip)
* [`obj time`](#obj-time)
  * [`fn withFrom(value="now-6h")`](#fn-timewithfrom)
  * [`fn withTo(value="now")`](#fn-timewithto)
* [`obj timepicker`](#obj-timepicker)
  * [`fn withCollapse(value=false)`](#fn-timepickerwithcollapse)
  * [`fn withEnable(value=true)`](#fn-timepickerwithenable)
  * [`fn withHidden(value=false)`](#fn-timepickerwithhidden)
  * [`fn withRefreshIntervals(value=["5s","10s","30s","1m","5m","15m","30m","1h","2h","1d"])`](#fn-timepickerwithrefreshintervals)
  * [`fn withRefreshIntervalsMixin(value=["5s","10s","30s","1m","5m","15m","30m","1h","2h","1d"])`](#fn-timepickerwithrefreshintervalsmixin)
  * [`fn withTimeOptions(value=["5m","15m","1h","6h","12h","24h","2d","7d","30d"])`](#fn-timepickerwithtimeoptions)
  * [`fn withTimeOptionsMixin(value=["5m","15m","1h","6h","12h","24h","2d","7d","30d"])`](#fn-timepickerwithtimeoptionsmixin)

## Fields

### fn new

```ts
new(title)
```

Creates a new dashboard with a title.

### fn withAnnotations

```ts
withAnnotations(value)
```

`withAnnotations` adds an array of annotations to a dashboard.

This function appends passed data to existing values


### fn withAnnotationsMixin

```ts
withAnnotationsMixin(value)
```

`withAnnotationsMixin` adds an array of annotations to a dashboard.

This function appends passed data to existing values


### fn withDescription

```ts
withDescription(value)
```

Description of dashboard.

### fn withEditable

```ts
withEditable(value=true)
```

Whether a dashboard is editable or not.

### fn withFiscalYearStartMonth

```ts
withFiscalYearStartMonth(value=0)
```

The month that the fiscal year starts on.  0 = January, 11 = December

### fn withLinks

```ts
withLinks(value)
```

TODO docs

### fn withLinksMixin

```ts
withLinksMixin(value)
```

TODO docs

### fn withLiveNow

```ts
withLiveNow(value)
```

When set to true, the dashboard will redraw panels at an interval matching the pixel width.
This will keep data "moving left" regardless of the query refresh rate.  This setting helps
avoid dashboards presenting stale live data

### fn withPanels

```ts
withPanels(value)
```



### fn withPanelsMixin

```ts
withPanelsMixin(value)
```



### fn withRefresh

```ts
withRefresh(value)
```

Refresh rate of dashboard. Represented via interval string, e.g. "5s", "1m", "1h", "1d".

### fn withRefreshMixin

```ts
withRefreshMixin(value)
```

Refresh rate of dashboard. Represented via interval string, e.g. "5s", "1m", "1h", "1d".

### fn withSchemaVersion

```ts
withSchemaVersion(value=36)
```

Version of the JSON schema, incremented each time a Grafana update brings
changes to said schema.
TODO this is the existing schema numbering system. It will be replaced by Thema's themaVersion

### fn withStyle

```ts
withStyle(value="dark")
```

Theme of dashboard.

Accepted values for `value` are "dark", "light"

### fn withTags

```ts
withTags(value)
```

Tags associated with dashboard.

### fn withTagsMixin

```ts
withTagsMixin(value)
```

Tags associated with dashboard.

### fn withTemplating

```ts
withTemplating(value)
```

TODO docs

### fn withTemplatingMixin

```ts
withTemplatingMixin(value)
```

TODO docs

### fn withTimezone

```ts
withTimezone(value="browser")
```

Timezone of dashboard. Accepts IANA TZDB zone ID or "browser" or "utc".

### fn withTitle

```ts
withTitle(value)
```

Title of dashboard.

### fn withUid

```ts
withUid(value)
```

Unique dashboard identifier that can be generated by anyone. string (8-40)

### fn withVariables

```ts
withVariables(value)
```

`withVariables` adds an array of variables to a dashboard


### fn withVariablesMixin

```ts
withVariablesMixin(value)
```

`withVariablesMixin` adds an array of variables to a dashboard.

This function appends passed data to existing values


### fn withWeekStart

```ts
withWeekStart(value)
```

TODO docs

### obj graphTooltip


#### fn graphTooltip.withSharedCrosshair

```ts
withSharedCrosshair()
```

Share crosshair on all panels.

#### fn graphTooltip.withSharedTooltip

```ts
withSharedTooltip()
```

Share crosshair and tooltip on all panels.

### obj time


#### fn time.withFrom

```ts
withFrom(value="now-6h")
```



#### fn time.withTo

```ts
withTo(value="now")
```



### obj timepicker


#### fn timepicker.withCollapse

```ts
withCollapse(value=false)
```

Whether timepicker is collapsed or not.

#### fn timepicker.withEnable

```ts
withEnable(value=true)
```

Whether timepicker is enabled or not.

#### fn timepicker.withHidden

```ts
withHidden(value=false)
```

Whether timepicker is visible or not.

#### fn timepicker.withRefreshIntervals

```ts
withRefreshIntervals(value=["5s","10s","30s","1m","5m","15m","30m","1h","2h","1d"])
```

Selectable intervals for auto-refresh.

#### fn timepicker.withRefreshIntervalsMixin

```ts
withRefreshIntervalsMixin(value=["5s","10s","30s","1m","5m","15m","30m","1h","2h","1d"])
```

Selectable intervals for auto-refresh.

#### fn timepicker.withTimeOptions

```ts
withTimeOptions(value=["5m","15m","1h","6h","12h","24h","2d","7d","30d"])
```

TODO docs

#### fn timepicker.withTimeOptionsMixin

```ts
withTimeOptionsMixin(value=["5m","15m","1h","6h","12h","24h","2d","7d","30d"])
```

TODO docs
