# heatmap

grafonnet.panel.heatmap

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
* [`obj fieldConfig`](#obj-fieldconfig)
  * [`obj defaults`](#obj-fieldconfigdefaults)
    * [`obj custom`](#obj-fieldconfigdefaultscustom)
      * [`fn withHideFrom(value)`](#fn-fieldconfigdefaultscustomwithhidefrom)
      * [`fn withHideFromMixin(value)`](#fn-fieldconfigdefaultscustomwithhidefrommixin)
      * [`fn withScaleDistribution(value)`](#fn-fieldconfigdefaultscustomwithscaledistribution)
      * [`fn withScaleDistributionMixin(value)`](#fn-fieldconfigdefaultscustomwithscaledistributionmixin)
      * [`obj hideFrom`](#obj-fieldconfigdefaultscustomhidefrom)
        * [`fn withLegend(value)`](#fn-fieldconfigdefaultscustomhidefromwithlegend)
        * [`fn withTooltip(value)`](#fn-fieldconfigdefaultscustomhidefromwithtooltip)
        * [`fn withViz(value)`](#fn-fieldconfigdefaultscustomhidefromwithviz)
      * [`obj scaleDistribution`](#obj-fieldconfigdefaultscustomscaledistribution)
        * [`fn withLinearThreshold(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithlinearthreshold)
        * [`fn withLog(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithlog)
        * [`fn withType(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithtype)
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
  * [`fn withCalculate(value=false)`](#fn-optionswithcalculate)
  * [`fn withCalculation(value)`](#fn-optionswithcalculation)
  * [`fn withCalculationMixin(value)`](#fn-optionswithcalculationmixin)
  * [`fn withCellGap(value=1)`](#fn-optionswithcellgap)
  * [`fn withCellRadius(value)`](#fn-optionswithcellradius)
  * [`fn withCellValues(value={})`](#fn-optionswithcellvalues)
  * [`fn withCellValuesMixin(value={})`](#fn-optionswithcellvaluesmixin)
  * [`fn withColor(value={"exponent": 0.5,"fill": "dark-orange","reverse": false,"scheme": "Oranges","steps": 64})`](#fn-optionswithcolor)
  * [`fn withColorMixin(value={"exponent": 0.5,"fill": "dark-orange","reverse": false,"scheme": "Oranges","steps": 64})`](#fn-optionswithcolormixin)
  * [`fn withExemplars(value)`](#fn-optionswithexemplars)
  * [`fn withExemplarsMixin(value)`](#fn-optionswithexemplarsmixin)
  * [`fn withFilterValues(value={"le": 0.000000001})`](#fn-optionswithfiltervalues)
  * [`fn withFilterValuesMixin(value={"le": 0.000000001})`](#fn-optionswithfiltervaluesmixin)
  * [`fn withLegend(value)`](#fn-optionswithlegend)
  * [`fn withLegendMixin(value)`](#fn-optionswithlegendmixin)
  * [`fn withRowsFrame(value)`](#fn-optionswithrowsframe)
  * [`fn withRowsFrameMixin(value)`](#fn-optionswithrowsframemixin)
  * [`fn withShowValue(value)`](#fn-optionswithshowvalue)
  * [`fn withTooltip(value)`](#fn-optionswithtooltip)
  * [`fn withTooltipMixin(value)`](#fn-optionswithtooltipmixin)
  * [`fn withYAxis(value)`](#fn-optionswithyaxis)
  * [`fn withYAxisMixin(value)`](#fn-optionswithyaxismixin)
  * [`obj calculation`](#obj-optionscalculation)
    * [`fn withXBuckets(value)`](#fn-optionscalculationwithxbuckets)
    * [`fn withXBucketsMixin(value)`](#fn-optionscalculationwithxbucketsmixin)
    * [`fn withYBuckets(value)`](#fn-optionscalculationwithybuckets)
    * [`fn withYBucketsMixin(value)`](#fn-optionscalculationwithybucketsmixin)
    * [`obj xBuckets`](#obj-optionscalculationxbuckets)
      * [`fn withMode(value)`](#fn-optionscalculationxbucketswithmode)
      * [`fn withScale(value)`](#fn-optionscalculationxbucketswithscale)
      * [`fn withScaleMixin(value)`](#fn-optionscalculationxbucketswithscalemixin)
      * [`fn withValue(value)`](#fn-optionscalculationxbucketswithvalue)
      * [`obj scale`](#obj-optionscalculationxbucketsscale)
        * [`fn withLinearThreshold(value)`](#fn-optionscalculationxbucketsscalewithlinearthreshold)
        * [`fn withLog(value)`](#fn-optionscalculationxbucketsscalewithlog)
        * [`fn withType(value)`](#fn-optionscalculationxbucketsscalewithtype)
    * [`obj yBuckets`](#obj-optionscalculationybuckets)
      * [`fn withMode(value)`](#fn-optionscalculationybucketswithmode)
      * [`fn withScale(value)`](#fn-optionscalculationybucketswithscale)
      * [`fn withScaleMixin(value)`](#fn-optionscalculationybucketswithscalemixin)
      * [`fn withValue(value)`](#fn-optionscalculationybucketswithvalue)
      * [`obj scale`](#obj-optionscalculationybucketsscale)
        * [`fn withLinearThreshold(value)`](#fn-optionscalculationybucketsscalewithlinearthreshold)
        * [`fn withLog(value)`](#fn-optionscalculationybucketsscalewithlog)
        * [`fn withType(value)`](#fn-optionscalculationybucketsscalewithtype)
  * [`obj cellValues`](#obj-optionscellvalues)
    * [`fn withCellValues(value)`](#fn-optionscellvalueswithcellvalues)
    * [`fn withCellValuesMixin(value)`](#fn-optionscellvalueswithcellvaluesmixin)
    * [`obj CellValues`](#obj-optionscellvaluescellvalues)
      * [`fn withDecimals(value)`](#fn-optionscellvaluescellvalueswithdecimals)
      * [`fn withUnit(value)`](#fn-optionscellvaluescellvalueswithunit)
  * [`obj color`](#obj-optionscolor)
    * [`fn withHeatmapColorOptions(value)`](#fn-optionscolorwithheatmapcoloroptions)
    * [`fn withHeatmapColorOptionsMixin(value)`](#fn-optionscolorwithheatmapcoloroptionsmixin)
    * [`obj HeatmapColorOptions`](#obj-optionscolorheatmapcoloroptions)
      * [`fn withExponent(value)`](#fn-optionscolorheatmapcoloroptionswithexponent)
      * [`fn withFill(value)`](#fn-optionscolorheatmapcoloroptionswithfill)
      * [`fn withMax(value)`](#fn-optionscolorheatmapcoloroptionswithmax)
      * [`fn withMin(value)`](#fn-optionscolorheatmapcoloroptionswithmin)
      * [`fn withMode(value)`](#fn-optionscolorheatmapcoloroptionswithmode)
      * [`fn withReverse(value)`](#fn-optionscolorheatmapcoloroptionswithreverse)
      * [`fn withScale(value)`](#fn-optionscolorheatmapcoloroptionswithscale)
      * [`fn withScheme(value)`](#fn-optionscolorheatmapcoloroptionswithscheme)
      * [`fn withSteps(value)`](#fn-optionscolorheatmapcoloroptionswithsteps)
  * [`obj exemplars`](#obj-optionsexemplars)
    * [`fn withColor(value)`](#fn-optionsexemplarswithcolor)
  * [`obj filterValues`](#obj-optionsfiltervalues)
    * [`fn withFilterValueRange(value)`](#fn-optionsfiltervalueswithfiltervaluerange)
    * [`fn withFilterValueRangeMixin(value)`](#fn-optionsfiltervalueswithfiltervaluerangemixin)
    * [`obj FilterValueRange`](#obj-optionsfiltervaluesfiltervaluerange)
      * [`fn withGe(value)`](#fn-optionsfiltervaluesfiltervaluerangewithge)
      * [`fn withLe(value)`](#fn-optionsfiltervaluesfiltervaluerangewithle)
  * [`obj legend`](#obj-optionslegend)
    * [`fn withShow(value)`](#fn-optionslegendwithshow)
  * [`obj rowsFrame`](#obj-optionsrowsframe)
    * [`fn withLayout(value)`](#fn-optionsrowsframewithlayout)
    * [`fn withValue(value)`](#fn-optionsrowsframewithvalue)
  * [`obj tooltip`](#obj-optionstooltip)
    * [`fn withShow(value)`](#fn-optionstooltipwithshow)
    * [`fn withYHistogram(value)`](#fn-optionstooltipwithyhistogram)
  * [`obj yAxis`](#obj-optionsyaxis)
    * [`fn withAxisCenteredZero(value)`](#fn-optionsyaxiswithaxiscenteredzero)
    * [`fn withAxisColorMode(value)`](#fn-optionsyaxiswithaxiscolormode)
    * [`fn withAxisGridShow(value)`](#fn-optionsyaxiswithaxisgridshow)
    * [`fn withAxisLabel(value)`](#fn-optionsyaxiswithaxislabel)
    * [`fn withAxisPlacement(value)`](#fn-optionsyaxiswithaxisplacement)
    * [`fn withAxisSoftMax(value)`](#fn-optionsyaxiswithaxissoftmax)
    * [`fn withAxisSoftMin(value)`](#fn-optionsyaxiswithaxissoftmin)
    * [`fn withAxisWidth(value)`](#fn-optionsyaxiswithaxiswidth)
    * [`fn withDecimals(value)`](#fn-optionsyaxiswithdecimals)
    * [`fn withMax(value)`](#fn-optionsyaxiswithmax)
    * [`fn withMin(value)`](#fn-optionsyaxiswithmin)
    * [`fn withReverse(value)`](#fn-optionsyaxiswithreverse)
    * [`fn withScaleDistribution(value)`](#fn-optionsyaxiswithscaledistribution)
    * [`fn withScaleDistributionMixin(value)`](#fn-optionsyaxiswithscaledistributionmixin)
    * [`fn withUnit(value)`](#fn-optionsyaxiswithunit)
    * [`obj scaleDistribution`](#obj-optionsyaxisscaledistribution)
      * [`fn withLinearThreshold(value)`](#fn-optionsyaxisscaledistributionwithlinearthreshold)
      * [`fn withLog(value)`](#fn-optionsyaxisscaledistributionwithlog)
      * [`fn withType(value)`](#fn-optionsyaxisscaledistributionwithtype)
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

Creates a new heatmap panel with a title.

### obj datasource


#### fn datasource.withType

```ts
withType(value)
```



#### fn datasource.withUid

```ts
withUid(value)
```



### obj fieldConfig


#### obj fieldConfig.defaults


##### obj fieldConfig.defaults.custom


###### fn fieldConfig.defaults.custom.withHideFrom

```ts
withHideFrom(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withHideFromMixin

```ts
withHideFromMixin(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withScaleDistribution

```ts
withScaleDistribution(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withScaleDistributionMixin

```ts
withScaleDistributionMixin(value)
```

TODO docs

###### obj fieldConfig.defaults.custom.hideFrom


####### fn fieldConfig.defaults.custom.hideFrom.withLegend

```ts
withLegend(value)
```



####### fn fieldConfig.defaults.custom.hideFrom.withTooltip

```ts
withTooltip(value)
```



####### fn fieldConfig.defaults.custom.hideFrom.withViz

```ts
withViz(value)
```



###### obj fieldConfig.defaults.custom.scaleDistribution


####### fn fieldConfig.defaults.custom.scaleDistribution.withLinearThreshold

```ts
withLinearThreshold(value)
```



####### fn fieldConfig.defaults.custom.scaleDistribution.withLog

```ts
withLog(value)
```



####### fn fieldConfig.defaults.custom.scaleDistribution.withType

```ts
withType(value)
```

TODO docs

Accepted values for `value` are "linear", "log", "ordinal", "symlog"

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


#### fn options.withCalculate

```ts
withCalculate(value=false)
```

Controls if the heatmap should be calculated from data

#### fn options.withCalculation

```ts
withCalculation(value)
```



#### fn options.withCalculationMixin

```ts
withCalculationMixin(value)
```



#### fn options.withCellGap

```ts
withCellGap(value=1)
```

Controls gap between cells

#### fn options.withCellRadius

```ts
withCellRadius(value)
```

Controls cell radius

#### fn options.withCellValues

```ts
withCellValues(value={})
```

Controls cell value unit

#### fn options.withCellValuesMixin

```ts
withCellValuesMixin(value={})
```

Controls cell value unit

#### fn options.withColor

```ts
withColor(value={"exponent": 0.5,"fill": "dark-orange","reverse": false,"scheme": "Oranges","steps": 64})
```

Controls the color options

#### fn options.withColorMixin

```ts
withColorMixin(value={"exponent": 0.5,"fill": "dark-orange","reverse": false,"scheme": "Oranges","steps": 64})
```

Controls the color options

#### fn options.withExemplars

```ts
withExemplars(value)
```

Controls exemplar options

#### fn options.withExemplarsMixin

```ts
withExemplarsMixin(value)
```

Controls exemplar options

#### fn options.withFilterValues

```ts
withFilterValues(value={"le": 0.000000001})
```

Filters values between a given range

#### fn options.withFilterValuesMixin

```ts
withFilterValuesMixin(value={"le": 0.000000001})
```

Filters values between a given range

#### fn options.withLegend

```ts
withLegend(value)
```

Controls legend options

#### fn options.withLegendMixin

```ts
withLegendMixin(value)
```

Controls legend options

#### fn options.withRowsFrame

```ts
withRowsFrame(value)
```

Controls frame rows options

#### fn options.withRowsFrameMixin

```ts
withRowsFrameMixin(value)
```

Controls frame rows options

#### fn options.withShowValue

```ts
withShowValue(value)
```

| *{
	layout: ui.HeatmapCellLayout & "auto" // TODO: fix after remove when https://github.com/grafana/cuetsy/issues/74 is fixed
}
Controls the display of the value in the cell

#### fn options.withTooltip

```ts
withTooltip(value)
```

Controls tooltip options

#### fn options.withTooltipMixin

```ts
withTooltipMixin(value)
```

Controls tooltip options

#### fn options.withYAxis

```ts
withYAxis(value)
```

Configuration options for the yAxis

#### fn options.withYAxisMixin

```ts
withYAxisMixin(value)
```

Configuration options for the yAxis

#### obj options.calculation


##### fn options.calculation.withXBuckets

```ts
withXBuckets(value)
```



##### fn options.calculation.withXBucketsMixin

```ts
withXBucketsMixin(value)
```



##### fn options.calculation.withYBuckets

```ts
withYBuckets(value)
```



##### fn options.calculation.withYBucketsMixin

```ts
withYBucketsMixin(value)
```



##### obj options.calculation.xBuckets


###### fn options.calculation.xBuckets.withMode

```ts
withMode(value)
```



Accepted values for `value` are "size", "count"

###### fn options.calculation.xBuckets.withScale

```ts
withScale(value)
```

TODO docs

###### fn options.calculation.xBuckets.withScaleMixin

```ts
withScaleMixin(value)
```

TODO docs

###### fn options.calculation.xBuckets.withValue

```ts
withValue(value)
```

The number of buckets to use for the axis in the heatmap

###### obj options.calculation.xBuckets.scale


####### fn options.calculation.xBuckets.scale.withLinearThreshold

```ts
withLinearThreshold(value)
```



####### fn options.calculation.xBuckets.scale.withLog

```ts
withLog(value)
```



####### fn options.calculation.xBuckets.scale.withType

```ts
withType(value)
```

TODO docs

Accepted values for `value` are "linear", "log", "ordinal", "symlog"

##### obj options.calculation.yBuckets


###### fn options.calculation.yBuckets.withMode

```ts
withMode(value)
```



Accepted values for `value` are "size", "count"

###### fn options.calculation.yBuckets.withScale

```ts
withScale(value)
```

TODO docs

###### fn options.calculation.yBuckets.withScaleMixin

```ts
withScaleMixin(value)
```

TODO docs

###### fn options.calculation.yBuckets.withValue

```ts
withValue(value)
```

The number of buckets to use for the axis in the heatmap

###### obj options.calculation.yBuckets.scale


####### fn options.calculation.yBuckets.scale.withLinearThreshold

```ts
withLinearThreshold(value)
```



####### fn options.calculation.yBuckets.scale.withLog

```ts
withLog(value)
```



####### fn options.calculation.yBuckets.scale.withType

```ts
withType(value)
```

TODO docs

Accepted values for `value` are "linear", "log", "ordinal", "symlog"

#### obj options.cellValues


##### fn options.cellValues.withCellValues

```ts
withCellValues(value)
```

Controls cell value options

##### fn options.cellValues.withCellValuesMixin

```ts
withCellValuesMixin(value)
```

Controls cell value options

##### obj options.cellValues.CellValues


###### fn options.cellValues.CellValues.withDecimals

```ts
withDecimals(value)
```

Controls the number of decimals for cell values

###### fn options.cellValues.CellValues.withUnit

```ts
withUnit(value)
```

Controls the cell value unit

#### obj options.color


##### fn options.color.withHeatmapColorOptions

```ts
withHeatmapColorOptions(value)
```

Controls various color options

##### fn options.color.withHeatmapColorOptionsMixin

```ts
withHeatmapColorOptionsMixin(value)
```

Controls various color options

##### obj options.color.HeatmapColorOptions


###### fn options.color.HeatmapColorOptions.withExponent

```ts
withExponent(value)
```

Controls the exponent when scale is set to exponential

###### fn options.color.HeatmapColorOptions.withFill

```ts
withFill(value)
```

Controls the color fill when in opacity mode

###### fn options.color.HeatmapColorOptions.withMax

```ts
withMax(value)
```

Sets the maximum value for the color scale

###### fn options.color.HeatmapColorOptions.withMin

```ts
withMin(value)
```

Sets the minimum value for the color scale

###### fn options.color.HeatmapColorOptions.withMode

```ts
withMode(value)
```

Controls the color mode of the heatmap

Accepted values for `value` are "opacity", "scheme"

###### fn options.color.HeatmapColorOptions.withReverse

```ts
withReverse(value)
```

Reverses the color scheme

###### fn options.color.HeatmapColorOptions.withScale

```ts
withScale(value)
```

Controls the color scale of the heatmap

Accepted values for `value` are "linear", "exponential"

###### fn options.color.HeatmapColorOptions.withScheme

```ts
withScheme(value)
```

Controls the color scheme used

###### fn options.color.HeatmapColorOptions.withSteps

```ts
withSteps(value)
```

Controls the number of color steps

#### obj options.exemplars


##### fn options.exemplars.withColor

```ts
withColor(value)
```

Sets the color of the exemplar markers

#### obj options.filterValues


##### fn options.filterValues.withFilterValueRange

```ts
withFilterValueRange(value)
```

Controls the value filter range

##### fn options.filterValues.withFilterValueRangeMixin

```ts
withFilterValueRangeMixin(value)
```

Controls the value filter range

##### obj options.filterValues.FilterValueRange


###### fn options.filterValues.FilterValueRange.withGe

```ts
withGe(value)
```

Sets the filter range to values greater than or equal to the given value

###### fn options.filterValues.FilterValueRange.withLe

```ts
withLe(value)
```

Sets the filter range to values less than or equal to the given value

#### obj options.legend


##### fn options.legend.withShow

```ts
withShow(value)
```

Controls if the legend is shown

#### obj options.rowsFrame


##### fn options.rowsFrame.withLayout

```ts
withLayout(value)
```



Accepted values for `value` are "le", "ge", "unknown", "auto"

##### fn options.rowsFrame.withValue

```ts
withValue(value)
```

Sets the name of the cell when not calculating from data

#### obj options.tooltip


##### fn options.tooltip.withShow

```ts
withShow(value)
```

Controls if the tooltip is shown

##### fn options.tooltip.withYHistogram

```ts
withYHistogram(value)
```

Controls if the tooltip shows a histogram of the y-axis values

#### obj options.yAxis


##### fn options.yAxis.withAxisCenteredZero

```ts
withAxisCenteredZero(value)
```



##### fn options.yAxis.withAxisColorMode

```ts
withAxisColorMode(value)
```

TODO docs

Accepted values for `value` are "text", "series"

##### fn options.yAxis.withAxisGridShow

```ts
withAxisGridShow(value)
```



##### fn options.yAxis.withAxisLabel

```ts
withAxisLabel(value)
```



##### fn options.yAxis.withAxisPlacement

```ts
withAxisPlacement(value)
```

TODO docs

Accepted values for `value` are "auto", "top", "right", "bottom", "left", "hidden"

##### fn options.yAxis.withAxisSoftMax

```ts
withAxisSoftMax(value)
```



##### fn options.yAxis.withAxisSoftMin

```ts
withAxisSoftMin(value)
```



##### fn options.yAxis.withAxisWidth

```ts
withAxisWidth(value)
```



##### fn options.yAxis.withDecimals

```ts
withDecimals(value)
```

Controls the number of decimals for yAxis values

##### fn options.yAxis.withMax

```ts
withMax(value)
```

Sets the maximum value for the yAxis

##### fn options.yAxis.withMin

```ts
withMin(value)
```

Sets the minimum value for the yAxis

##### fn options.yAxis.withReverse

```ts
withReverse(value)
```

Reverses the yAxis

##### fn options.yAxis.withScaleDistribution

```ts
withScaleDistribution(value)
```

TODO docs

##### fn options.yAxis.withScaleDistributionMixin

```ts
withScaleDistributionMixin(value)
```

TODO docs

##### fn options.yAxis.withUnit

```ts
withUnit(value)
```

Sets the yAxis unit

##### obj options.yAxis.scaleDistribution


###### fn options.yAxis.scaleDistribution.withLinearThreshold

```ts
withLinearThreshold(value)
```



###### fn options.yAxis.scaleDistribution.withLog

```ts
withLog(value)
```



###### fn options.yAxis.scaleDistribution.withType

```ts
withType(value)
```

TODO docs

Accepted values for `value` are "linear", "log", "ordinal", "symlog"

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
