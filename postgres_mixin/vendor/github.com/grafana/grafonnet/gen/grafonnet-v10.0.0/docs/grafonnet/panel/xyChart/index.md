# xyChart

grafonnet.panel.xyChart

## Subpackages

* [fieldOverride](fieldOverride.md)
* [link](link.md)
* [thresholdStep](thresholdStep.md)
* [transformation](transformation.md)
* [valueMapping](valueMapping.md)

## Index

* [`fn new(title)`](#fn-new)
* [`obj datasource`](#obj-datasource)
  * [`fn withType(value)`](#fn-datasourcewithtype)
  * [`fn withUid(value)`](#fn-datasourcewithuid)
* [`obj gridPos`](#obj-gridpos)
  * [`fn withH(value=9)`](#fn-gridposwithh)
  * [`fn withStatic(value)`](#fn-gridposwithstatic)
  * [`fn withW(value=12)`](#fn-gridposwithw)
  * [`fn withX(value=0)`](#fn-gridposwithx)
  * [`fn withY(value=0)`](#fn-gridposwithy)
* [`obj libraryPanel`](#obj-librarypanel)
  * [`fn withName(value)`](#fn-librarypanelwithname)
  * [`fn withUid(value)`](#fn-librarypanelwithuid)
* [`obj options`](#obj-options)
  * [`fn withDims(value)`](#fn-optionswithdims)
  * [`fn withDimsMixin(value)`](#fn-optionswithdimsmixin)
  * [`fn withLegend(value)`](#fn-optionswithlegend)
  * [`fn withLegendMixin(value)`](#fn-optionswithlegendmixin)
  * [`fn withSeries(value)`](#fn-optionswithseries)
  * [`fn withSeriesMapping(value)`](#fn-optionswithseriesmapping)
  * [`fn withSeriesMixin(value)`](#fn-optionswithseriesmixin)
  * [`fn withTooltip(value)`](#fn-optionswithtooltip)
  * [`fn withTooltipMixin(value)`](#fn-optionswithtooltipmixin)
  * [`obj dims`](#obj-optionsdims)
    * [`fn withExclude(value)`](#fn-optionsdimswithexclude)
    * [`fn withExcludeMixin(value)`](#fn-optionsdimswithexcludemixin)
    * [`fn withFrame(value)`](#fn-optionsdimswithframe)
    * [`fn withX(value)`](#fn-optionsdimswithx)
  * [`obj legend`](#obj-optionslegend)
    * [`fn withAsTable(value)`](#fn-optionslegendwithastable)
    * [`fn withCalcs(value)`](#fn-optionslegendwithcalcs)
    * [`fn withCalcsMixin(value)`](#fn-optionslegendwithcalcsmixin)
    * [`fn withDisplayMode(value)`](#fn-optionslegendwithdisplaymode)
    * [`fn withIsVisible(value)`](#fn-optionslegendwithisvisible)
    * [`fn withPlacement(value)`](#fn-optionslegendwithplacement)
    * [`fn withShowLegend(value)`](#fn-optionslegendwithshowlegend)
    * [`fn withSortBy(value)`](#fn-optionslegendwithsortby)
    * [`fn withSortDesc(value)`](#fn-optionslegendwithsortdesc)
    * [`fn withWidth(value)`](#fn-optionslegendwithwidth)
  * [`obj series`](#obj-optionsseries)
    * [`fn withAxisCenteredZero(value)`](#fn-optionsserieswithaxiscenteredzero)
    * [`fn withAxisColorMode(value)`](#fn-optionsserieswithaxiscolormode)
    * [`fn withAxisGridShow(value)`](#fn-optionsserieswithaxisgridshow)
    * [`fn withAxisLabel(value)`](#fn-optionsserieswithaxislabel)
    * [`fn withAxisPlacement(value)`](#fn-optionsserieswithaxisplacement)
    * [`fn withAxisSoftMax(value)`](#fn-optionsserieswithaxissoftmax)
    * [`fn withAxisSoftMin(value)`](#fn-optionsserieswithaxissoftmin)
    * [`fn withAxisWidth(value)`](#fn-optionsserieswithaxiswidth)
    * [`fn withHideFrom(value)`](#fn-optionsserieswithhidefrom)
    * [`fn withHideFromMixin(value)`](#fn-optionsserieswithhidefrommixin)
    * [`fn withLabel(value)`](#fn-optionsserieswithlabel)
    * [`fn withLabelValue(value)`](#fn-optionsserieswithlabelvalue)
    * [`fn withLabelValueMixin(value)`](#fn-optionsserieswithlabelvaluemixin)
    * [`fn withLineColor(value)`](#fn-optionsserieswithlinecolor)
    * [`fn withLineColorMixin(value)`](#fn-optionsserieswithlinecolormixin)
    * [`fn withLineStyle(value)`](#fn-optionsserieswithlinestyle)
    * [`fn withLineStyleMixin(value)`](#fn-optionsserieswithlinestylemixin)
    * [`fn withLineWidth(value)`](#fn-optionsserieswithlinewidth)
    * [`fn withName(value)`](#fn-optionsserieswithname)
    * [`fn withPointColor(value)`](#fn-optionsserieswithpointcolor)
    * [`fn withPointColorMixin(value)`](#fn-optionsserieswithpointcolormixin)
    * [`fn withPointSize(value)`](#fn-optionsserieswithpointsize)
    * [`fn withPointSizeMixin(value)`](#fn-optionsserieswithpointsizemixin)
    * [`fn withScaleDistribution(value)`](#fn-optionsserieswithscaledistribution)
    * [`fn withScaleDistributionMixin(value)`](#fn-optionsserieswithscaledistributionmixin)
    * [`fn withShow(value)`](#fn-optionsserieswithshow)
    * [`fn withX(value)`](#fn-optionsserieswithx)
    * [`fn withY(value)`](#fn-optionsserieswithy)
    * [`obj hideFrom`](#obj-optionsserieshidefrom)
      * [`fn withLegend(value)`](#fn-optionsserieshidefromwithlegend)
      * [`fn withTooltip(value)`](#fn-optionsserieshidefromwithtooltip)
      * [`fn withViz(value)`](#fn-optionsserieshidefromwithviz)
    * [`obj labelValue`](#obj-optionsserieslabelvalue)
      * [`fn withField(value)`](#fn-optionsserieslabelvaluewithfield)
      * [`fn withFixed(value)`](#fn-optionsserieslabelvaluewithfixed)
      * [`fn withMode(value)`](#fn-optionsserieslabelvaluewithmode)
    * [`obj lineColor`](#obj-optionsserieslinecolor)
      * [`fn withField(value)`](#fn-optionsserieslinecolorwithfield)
      * [`fn withFixed(value)`](#fn-optionsserieslinecolorwithfixed)
    * [`obj lineStyle`](#obj-optionsserieslinestyle)
      * [`fn withDash(value)`](#fn-optionsserieslinestylewithdash)
      * [`fn withDashMixin(value)`](#fn-optionsserieslinestylewithdashmixin)
      * [`fn withFill(value)`](#fn-optionsserieslinestylewithfill)
    * [`obj pointColor`](#obj-optionsseriespointcolor)
      * [`fn withField(value)`](#fn-optionsseriespointcolorwithfield)
      * [`fn withFixed(value)`](#fn-optionsseriespointcolorwithfixed)
    * [`obj pointSize`](#obj-optionsseriespointsize)
      * [`fn withField(value)`](#fn-optionsseriespointsizewithfield)
      * [`fn withFixed(value)`](#fn-optionsseriespointsizewithfixed)
      * [`fn withMax(value)`](#fn-optionsseriespointsizewithmax)
      * [`fn withMin(value)`](#fn-optionsseriespointsizewithmin)
      * [`fn withMode(value)`](#fn-optionsseriespointsizewithmode)
    * [`obj scaleDistribution`](#obj-optionsseriesscaledistribution)
      * [`fn withLinearThreshold(value)`](#fn-optionsseriesscaledistributionwithlinearthreshold)
      * [`fn withLog(value)`](#fn-optionsseriesscaledistributionwithlog)
      * [`fn withType(value)`](#fn-optionsseriesscaledistributionwithtype)
  * [`obj tooltip`](#obj-optionstooltip)
    * [`fn withMode(value)`](#fn-optionstooltipwithmode)
    * [`fn withSort(value)`](#fn-optionstooltipwithsort)
* [`obj panelOptions`](#obj-paneloptions)
  * [`fn withDescription(value)`](#fn-paneloptionswithdescription)
  * [`fn withLinks(value)`](#fn-paneloptionswithlinks)
  * [`fn withLinksMixin(value)`](#fn-paneloptionswithlinksmixin)
  * [`fn withRepeat(value)`](#fn-paneloptionswithrepeat)
  * [`fn withRepeatDirection(value="h")`](#fn-paneloptionswithrepeatdirection)
  * [`fn withTitle(value)`](#fn-paneloptionswithtitle)
  * [`fn withTransparent(value=false)`](#fn-paneloptionswithtransparent)
* [`obj queryOptions`](#obj-queryoptions)
  * [`fn withDatasource(value)`](#fn-queryoptionswithdatasource)
  * [`fn withDatasourceMixin(value)`](#fn-queryoptionswithdatasourcemixin)
  * [`fn withInterval(value)`](#fn-queryoptionswithinterval)
  * [`fn withMaxDataPoints(value)`](#fn-queryoptionswithmaxdatapoints)
  * [`fn withTargets(value)`](#fn-queryoptionswithtargets)
  * [`fn withTargetsMixin(value)`](#fn-queryoptionswithtargetsmixin)
  * [`fn withTimeFrom(value)`](#fn-queryoptionswithtimefrom)
  * [`fn withTimeShift(value)`](#fn-queryoptionswithtimeshift)
  * [`fn withTransformations(value)`](#fn-queryoptionswithtransformations)
  * [`fn withTransformationsMixin(value)`](#fn-queryoptionswithtransformationsmixin)
* [`obj standardOptions`](#obj-standardoptions)
  * [`fn withDecimals(value)`](#fn-standardoptionswithdecimals)
  * [`fn withDisplayName(value)`](#fn-standardoptionswithdisplayname)
  * [`fn withLinks(value)`](#fn-standardoptionswithlinks)
  * [`fn withLinksMixin(value)`](#fn-standardoptionswithlinksmixin)
  * [`fn withMappings(value)`](#fn-standardoptionswithmappings)
  * [`fn withMappingsMixin(value)`](#fn-standardoptionswithmappingsmixin)
  * [`fn withMax(value)`](#fn-standardoptionswithmax)
  * [`fn withMin(value)`](#fn-standardoptionswithmin)
  * [`fn withNoValue(value)`](#fn-standardoptionswithnovalue)
  * [`fn withOverrides(value)`](#fn-standardoptionswithoverrides)
  * [`fn withOverridesMixin(value)`](#fn-standardoptionswithoverridesmixin)
  * [`fn withUnit(value)`](#fn-standardoptionswithunit)
  * [`obj color`](#obj-standardoptionscolor)
    * [`fn withFixedColor(value)`](#fn-standardoptionscolorwithfixedcolor)
    * [`fn withMode(value)`](#fn-standardoptionscolorwithmode)
    * [`fn withSeriesBy(value)`](#fn-standardoptionscolorwithseriesby)
  * [`obj thresholds`](#obj-standardoptionsthresholds)
    * [`fn withMode(value)`](#fn-standardoptionsthresholdswithmode)
    * [`fn withSteps(value)`](#fn-standardoptionsthresholdswithsteps)
    * [`fn withStepsMixin(value)`](#fn-standardoptionsthresholdswithstepsmixin)

## Fields

### fn new

```ts
new(title)
```

Creates a new xyChart panel with a title.

### obj datasource


#### fn datasource.withType

```ts
withType(value)
```



#### fn datasource.withUid

```ts
withUid(value)
```



### obj gridPos


#### fn gridPos.withH

```ts
withH(value=9)
```

Panel

#### fn gridPos.withStatic

```ts
withStatic(value)
```

true if fixed

#### fn gridPos.withW

```ts
withW(value=12)
```

Panel

#### fn gridPos.withX

```ts
withX(value=0)
```

Panel x

#### fn gridPos.withY

```ts
withY(value=0)
```

Panel y

### obj libraryPanel


#### fn libraryPanel.withName

```ts
withName(value)
```



#### fn libraryPanel.withUid

```ts
withUid(value)
```



### obj options


#### fn options.withDims

```ts
withDims(value)
```



#### fn options.withDimsMixin

```ts
withDimsMixin(value)
```



#### fn options.withLegend

```ts
withLegend(value)
```

TODO docs

#### fn options.withLegendMixin

```ts
withLegendMixin(value)
```

TODO docs

#### fn options.withSeries

```ts
withSeries(value)
```



#### fn options.withSeriesMapping

```ts
withSeriesMapping(value)
```



Accepted values for `value` are "auto", "manual"

#### fn options.withSeriesMixin

```ts
withSeriesMixin(value)
```



#### fn options.withTooltip

```ts
withTooltip(value)
```

TODO docs

#### fn options.withTooltipMixin

```ts
withTooltipMixin(value)
```

TODO docs

#### obj options.dims


##### fn options.dims.withExclude

```ts
withExclude(value)
```



##### fn options.dims.withExcludeMixin

```ts
withExcludeMixin(value)
```



##### fn options.dims.withFrame

```ts
withFrame(value)
```



##### fn options.dims.withX

```ts
withX(value)
```



#### obj options.legend


##### fn options.legend.withAsTable

```ts
withAsTable(value)
```



##### fn options.legend.withCalcs

```ts
withCalcs(value)
```



##### fn options.legend.withCalcsMixin

```ts
withCalcsMixin(value)
```



##### fn options.legend.withDisplayMode

```ts
withDisplayMode(value)
```

TODO docs
Note: "hidden" needs to remain as an option for plugins compatibility

Accepted values for `value` are "list", "table", "hidden"

##### fn options.legend.withIsVisible

```ts
withIsVisible(value)
```



##### fn options.legend.withPlacement

```ts
withPlacement(value)
```

TODO docs

Accepted values for `value` are "bottom", "right"

##### fn options.legend.withShowLegend

```ts
withShowLegend(value)
```



##### fn options.legend.withSortBy

```ts
withSortBy(value)
```



##### fn options.legend.withSortDesc

```ts
withSortDesc(value)
```



##### fn options.legend.withWidth

```ts
withWidth(value)
```



#### obj options.series


##### fn options.series.withAxisCenteredZero

```ts
withAxisCenteredZero(value)
```



##### fn options.series.withAxisColorMode

```ts
withAxisColorMode(value)
```

TODO docs

Accepted values for `value` are "text", "series"

##### fn options.series.withAxisGridShow

```ts
withAxisGridShow(value)
```



##### fn options.series.withAxisLabel

```ts
withAxisLabel(value)
```



##### fn options.series.withAxisPlacement

```ts
withAxisPlacement(value)
```

TODO docs

Accepted values for `value` are "auto", "top", "right", "bottom", "left", "hidden"

##### fn options.series.withAxisSoftMax

```ts
withAxisSoftMax(value)
```



##### fn options.series.withAxisSoftMin

```ts
withAxisSoftMin(value)
```



##### fn options.series.withAxisWidth

```ts
withAxisWidth(value)
```



##### fn options.series.withHideFrom

```ts
withHideFrom(value)
```

TODO docs

##### fn options.series.withHideFromMixin

```ts
withHideFromMixin(value)
```

TODO docs

##### fn options.series.withLabel

```ts
withLabel(value)
```

TODO docs

Accepted values for `value` are "auto", "never", "always"

##### fn options.series.withLabelValue

```ts
withLabelValue(value)
```



##### fn options.series.withLabelValueMixin

```ts
withLabelValueMixin(value)
```



##### fn options.series.withLineColor

```ts
withLineColor(value)
```



##### fn options.series.withLineColorMixin

```ts
withLineColorMixin(value)
```



##### fn options.series.withLineStyle

```ts
withLineStyle(value)
```

TODO docs

##### fn options.series.withLineStyleMixin

```ts
withLineStyleMixin(value)
```

TODO docs

##### fn options.series.withLineWidth

```ts
withLineWidth(value)
```



##### fn options.series.withName

```ts
withName(value)
```



##### fn options.series.withPointColor

```ts
withPointColor(value)
```



##### fn options.series.withPointColorMixin

```ts
withPointColorMixin(value)
```



##### fn options.series.withPointSize

```ts
withPointSize(value)
```



##### fn options.series.withPointSizeMixin

```ts
withPointSizeMixin(value)
```



##### fn options.series.withScaleDistribution

```ts
withScaleDistribution(value)
```

TODO docs

##### fn options.series.withScaleDistributionMixin

```ts
withScaleDistributionMixin(value)
```

TODO docs

##### fn options.series.withShow

```ts
withShow(value)
```



Accepted values for `value` are "points", "lines", "points+lines"

##### fn options.series.withX

```ts
withX(value)
```



##### fn options.series.withY

```ts
withY(value)
```



##### obj options.series.hideFrom


###### fn options.series.hideFrom.withLegend

```ts
withLegend(value)
```



###### fn options.series.hideFrom.withTooltip

```ts
withTooltip(value)
```



###### fn options.series.hideFrom.withViz

```ts
withViz(value)
```



##### obj options.series.labelValue


###### fn options.series.labelValue.withField

```ts
withField(value)
```

fixed: T -- will be added by each element

###### fn options.series.labelValue.withFixed

```ts
withFixed(value)
```



###### fn options.series.labelValue.withMode

```ts
withMode(value)
```



Accepted values for `value` are "fixed", "field", "template"

##### obj options.series.lineColor


###### fn options.series.lineColor.withField

```ts
withField(value)
```

fixed: T -- will be added by each element

###### fn options.series.lineColor.withFixed

```ts
withFixed(value)
```



##### obj options.series.lineStyle


###### fn options.series.lineStyle.withDash

```ts
withDash(value)
```



###### fn options.series.lineStyle.withDashMixin

```ts
withDashMixin(value)
```



###### fn options.series.lineStyle.withFill

```ts
withFill(value)
```



Accepted values for `value` are "solid", "dash", "dot", "square"

##### obj options.series.pointColor


###### fn options.series.pointColor.withField

```ts
withField(value)
```

fixed: T -- will be added by each element

###### fn options.series.pointColor.withFixed

```ts
withFixed(value)
```



##### obj options.series.pointSize


###### fn options.series.pointSize.withField

```ts
withField(value)
```

fixed: T -- will be added by each element

###### fn options.series.pointSize.withFixed

```ts
withFixed(value)
```



###### fn options.series.pointSize.withMax

```ts
withMax(value)
```



###### fn options.series.pointSize.withMin

```ts
withMin(value)
```



###### fn options.series.pointSize.withMode

```ts
withMode(value)
```



Accepted values for `value` are "linear", "quad"

##### obj options.series.scaleDistribution


###### fn options.series.scaleDistribution.withLinearThreshold

```ts
withLinearThreshold(value)
```



###### fn options.series.scaleDistribution.withLog

```ts
withLog(value)
```



###### fn options.series.scaleDistribution.withType

```ts
withType(value)
```

TODO docs

Accepted values for `value` are "linear", "log", "ordinal", "symlog"

#### obj options.tooltip


##### fn options.tooltip.withMode

```ts
withMode(value)
```

TODO docs

Accepted values for `value` are "single", "multi", "none"

##### fn options.tooltip.withSort

```ts
withSort(value)
```

TODO docs

Accepted values for `value` are "asc", "desc", "none"

### obj panelOptions


#### fn panelOptions.withDescription

```ts
withDescription(value)
```

Description.

#### fn panelOptions.withLinks

```ts
withLinks(value)
```

Panel links.
TODO fill this out - seems there are a couple variants?

#### fn panelOptions.withLinksMixin

```ts
withLinksMixin(value)
```

Panel links.
TODO fill this out - seems there are a couple variants?

#### fn panelOptions.withRepeat

```ts
withRepeat(value)
```

Name of template variable to repeat for.

#### fn panelOptions.withRepeatDirection

```ts
withRepeatDirection(value="h")
```

Direction to repeat in if 'repeat' is set.
"h" for horizontal, "v" for vertical.
TODO this is probably optional

Accepted values for `value` are "h", "v"

#### fn panelOptions.withTitle

```ts
withTitle(value)
```

Panel title.

#### fn panelOptions.withTransparent

```ts
withTransparent(value=false)
```

Whether to display the panel without a background.

### obj queryOptions


#### fn queryOptions.withDatasource

```ts
withDatasource(value)
```

The datasource used in all targets.

#### fn queryOptions.withDatasourceMixin

```ts
withDatasourceMixin(value)
```

The datasource used in all targets.

#### fn queryOptions.withInterval

```ts
withInterval(value)
```

TODO docs
TODO tighter constraint

#### fn queryOptions.withMaxDataPoints

```ts
withMaxDataPoints(value)
```

TODO docs

#### fn queryOptions.withTargets

```ts
withTargets(value)
```

TODO docs

#### fn queryOptions.withTargetsMixin

```ts
withTargetsMixin(value)
```

TODO docs

#### fn queryOptions.withTimeFrom

```ts
withTimeFrom(value)
```

TODO docs
TODO tighter constraint

#### fn queryOptions.withTimeShift

```ts
withTimeShift(value)
```

TODO docs
TODO tighter constraint

#### fn queryOptions.withTransformations

```ts
withTransformations(value)
```



#### fn queryOptions.withTransformationsMixin

```ts
withTransformationsMixin(value)
```



### obj standardOptions


#### fn standardOptions.withDecimals

```ts
withDecimals(value)
```

Significant digits (for display)

#### fn standardOptions.withDisplayName

```ts
withDisplayName(value)
```

The display value for this field.  This supports template variables blank is auto

#### fn standardOptions.withLinks

```ts
withLinks(value)
```

The behavior when clicking on a result

#### fn standardOptions.withLinksMixin

```ts
withLinksMixin(value)
```

The behavior when clicking on a result

#### fn standardOptions.withMappings

```ts
withMappings(value)
```

Convert input values into a display string

#### fn standardOptions.withMappingsMixin

```ts
withMappingsMixin(value)
```

Convert input values into a display string

#### fn standardOptions.withMax

```ts
withMax(value)
```



#### fn standardOptions.withMin

```ts
withMin(value)
```



#### fn standardOptions.withNoValue

```ts
withNoValue(value)
```

Alternative to empty string

#### fn standardOptions.withOverrides

```ts
withOverrides(value)
```



#### fn standardOptions.withOverridesMixin

```ts
withOverridesMixin(value)
```



#### fn standardOptions.withUnit

```ts
withUnit(value)
```

Numeric Options

#### obj standardOptions.color


##### fn standardOptions.color.withFixedColor

```ts
withFixedColor(value)
```

Stores the fixed color value if mode is fixed

##### fn standardOptions.color.withMode

```ts
withMode(value)
```

The main color scheme mode

##### fn standardOptions.color.withSeriesBy

```ts
withSeriesBy(value)
```

TODO docs

Accepted values for `value` are "min", "max", "last"

#### obj standardOptions.thresholds


##### fn standardOptions.thresholds.withMode

```ts
withMode(value)
```



Accepted values for `value` are "absolute", "percentage"

##### fn standardOptions.thresholds.withSteps

```ts
withSteps(value)
```

Must be sorted by 'value', first value is always -Infinity

##### fn standardOptions.thresholds.withStepsMixin

```ts
withStepsMixin(value)
```

Must be sorted by 'value', first value is always -Infinity
