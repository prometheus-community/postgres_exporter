# timeSeries

grafonnet.panel.timeSeries

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
      * [`fn withAxisCenteredZero(value)`](#fn-fieldconfigdefaultscustomwithaxiscenteredzero)
      * [`fn withAxisColorMode(value)`](#fn-fieldconfigdefaultscustomwithaxiscolormode)
      * [`fn withAxisGridShow(value)`](#fn-fieldconfigdefaultscustomwithaxisgridshow)
      * [`fn withAxisLabel(value)`](#fn-fieldconfigdefaultscustomwithaxislabel)
      * [`fn withAxisPlacement(value)`](#fn-fieldconfigdefaultscustomwithaxisplacement)
      * [`fn withAxisSoftMax(value)`](#fn-fieldconfigdefaultscustomwithaxissoftmax)
      * [`fn withAxisSoftMin(value)`](#fn-fieldconfigdefaultscustomwithaxissoftmin)
      * [`fn withAxisWidth(value)`](#fn-fieldconfigdefaultscustomwithaxiswidth)
      * [`fn withBarAlignment(value)`](#fn-fieldconfigdefaultscustomwithbaralignment)
      * [`fn withBarMaxWidth(value)`](#fn-fieldconfigdefaultscustomwithbarmaxwidth)
      * [`fn withBarWidthFactor(value)`](#fn-fieldconfigdefaultscustomwithbarwidthfactor)
      * [`fn withDrawStyle(value)`](#fn-fieldconfigdefaultscustomwithdrawstyle)
      * [`fn withFillBelowTo(value)`](#fn-fieldconfigdefaultscustomwithfillbelowto)
      * [`fn withFillColor(value)`](#fn-fieldconfigdefaultscustomwithfillcolor)
      * [`fn withFillOpacity(value)`](#fn-fieldconfigdefaultscustomwithfillopacity)
      * [`fn withGradientMode(value)`](#fn-fieldconfigdefaultscustomwithgradientmode)
      * [`fn withHideFrom(value)`](#fn-fieldconfigdefaultscustomwithhidefrom)
      * [`fn withHideFromMixin(value)`](#fn-fieldconfigdefaultscustomwithhidefrommixin)
      * [`fn withLineColor(value)`](#fn-fieldconfigdefaultscustomwithlinecolor)
      * [`fn withLineInterpolation(value)`](#fn-fieldconfigdefaultscustomwithlineinterpolation)
      * [`fn withLineStyle(value)`](#fn-fieldconfigdefaultscustomwithlinestyle)
      * [`fn withLineStyleMixin(value)`](#fn-fieldconfigdefaultscustomwithlinestylemixin)
      * [`fn withLineWidth(value)`](#fn-fieldconfigdefaultscustomwithlinewidth)
      * [`fn withPointColor(value)`](#fn-fieldconfigdefaultscustomwithpointcolor)
      * [`fn withPointSize(value)`](#fn-fieldconfigdefaultscustomwithpointsize)
      * [`fn withPointSymbol(value)`](#fn-fieldconfigdefaultscustomwithpointsymbol)
      * [`fn withScaleDistribution(value)`](#fn-fieldconfigdefaultscustomwithscaledistribution)
      * [`fn withScaleDistributionMixin(value)`](#fn-fieldconfigdefaultscustomwithscaledistributionmixin)
      * [`fn withShowPoints(value)`](#fn-fieldconfigdefaultscustomwithshowpoints)
      * [`fn withSpanNulls(value)`](#fn-fieldconfigdefaultscustomwithspannulls)
      * [`fn withSpanNullsMixin(value)`](#fn-fieldconfigdefaultscustomwithspannullsmixin)
      * [`fn withStacking(value)`](#fn-fieldconfigdefaultscustomwithstacking)
      * [`fn withStackingMixin(value)`](#fn-fieldconfigdefaultscustomwithstackingmixin)
      * [`fn withThresholdsStyle(value)`](#fn-fieldconfigdefaultscustomwiththresholdsstyle)
      * [`fn withThresholdsStyleMixin(value)`](#fn-fieldconfigdefaultscustomwiththresholdsstylemixin)
      * [`fn withTransform(value)`](#fn-fieldconfigdefaultscustomwithtransform)
      * [`obj hideFrom`](#obj-fieldconfigdefaultscustomhidefrom)
        * [`fn withLegend(value)`](#fn-fieldconfigdefaultscustomhidefromwithlegend)
        * [`fn withTooltip(value)`](#fn-fieldconfigdefaultscustomhidefromwithtooltip)
        * [`fn withViz(value)`](#fn-fieldconfigdefaultscustomhidefromwithviz)
      * [`obj lineStyle`](#obj-fieldconfigdefaultscustomlinestyle)
        * [`fn withDash(value)`](#fn-fieldconfigdefaultscustomlinestylewithdash)
        * [`fn withDashMixin(value)`](#fn-fieldconfigdefaultscustomlinestylewithdashmixin)
        * [`fn withFill(value)`](#fn-fieldconfigdefaultscustomlinestylewithfill)
      * [`obj scaleDistribution`](#obj-fieldconfigdefaultscustomscaledistribution)
        * [`fn withLinearThreshold(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithlinearthreshold)
        * [`fn withLog(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithlog)
        * [`fn withType(value)`](#fn-fieldconfigdefaultscustomscaledistributionwithtype)
      * [`obj stacking`](#obj-fieldconfigdefaultscustomstacking)
        * [`fn withGroup(value)`](#fn-fieldconfigdefaultscustomstackingwithgroup)
        * [`fn withMode(value)`](#fn-fieldconfigdefaultscustomstackingwithmode)
      * [`obj thresholdsStyle`](#obj-fieldconfigdefaultscustomthresholdsstyle)
        * [`fn withMode(value)`](#fn-fieldconfigdefaultscustomthresholdsstylewithmode)
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
  * [`fn withLegend(value)`](#fn-optionswithlegend)
  * [`fn withLegendMixin(value)`](#fn-optionswithlegendmixin)
  * [`fn withTimezone(value)`](#fn-optionswithtimezone)
  * [`fn withTimezoneMixin(value)`](#fn-optionswithtimezonemixin)
  * [`fn withTooltip(value)`](#fn-optionswithtooltip)
  * [`fn withTooltipMixin(value)`](#fn-optionswithtooltipmixin)
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

Creates a new timeSeries panel with a title.

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


###### fn fieldConfig.defaults.custom.withAxisCenteredZero

```ts
withAxisCenteredZero(value)
```



###### fn fieldConfig.defaults.custom.withAxisColorMode

```ts
withAxisColorMode(value)
```

TODO docs

Accepted values for `value` are "text", "series"

###### fn fieldConfig.defaults.custom.withAxisGridShow

```ts
withAxisGridShow(value)
```



###### fn fieldConfig.defaults.custom.withAxisLabel

```ts
withAxisLabel(value)
```



###### fn fieldConfig.defaults.custom.withAxisPlacement

```ts
withAxisPlacement(value)
```

TODO docs

Accepted values for `value` are "auto", "top", "right", "bottom", "left", "hidden"

###### fn fieldConfig.defaults.custom.withAxisSoftMax

```ts
withAxisSoftMax(value)
```



###### fn fieldConfig.defaults.custom.withAxisSoftMin

```ts
withAxisSoftMin(value)
```



###### fn fieldConfig.defaults.custom.withAxisWidth

```ts
withAxisWidth(value)
```



###### fn fieldConfig.defaults.custom.withBarAlignment

```ts
withBarAlignment(value)
```

TODO docs

Accepted values for `value` are -1, 0, 1

###### fn fieldConfig.defaults.custom.withBarMaxWidth

```ts
withBarMaxWidth(value)
```



###### fn fieldConfig.defaults.custom.withBarWidthFactor

```ts
withBarWidthFactor(value)
```



###### fn fieldConfig.defaults.custom.withDrawStyle

```ts
withDrawStyle(value)
```

TODO docs

Accepted values for `value` are "line", "bars", "points"

###### fn fieldConfig.defaults.custom.withFillBelowTo

```ts
withFillBelowTo(value)
```



###### fn fieldConfig.defaults.custom.withFillColor

```ts
withFillColor(value)
```



###### fn fieldConfig.defaults.custom.withFillOpacity

```ts
withFillOpacity(value)
```



###### fn fieldConfig.defaults.custom.withGradientMode

```ts
withGradientMode(value)
```

TODO docs

Accepted values for `value` are "none", "opacity", "hue", "scheme"

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

###### fn fieldConfig.defaults.custom.withLineColor

```ts
withLineColor(value)
```



###### fn fieldConfig.defaults.custom.withLineInterpolation

```ts
withLineInterpolation(value)
```

TODO docs

Accepted values for `value` are "linear", "smooth", "stepBefore", "stepAfter"

###### fn fieldConfig.defaults.custom.withLineStyle

```ts
withLineStyle(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withLineStyleMixin

```ts
withLineStyleMixin(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withLineWidth

```ts
withLineWidth(value)
```



###### fn fieldConfig.defaults.custom.withPointColor

```ts
withPointColor(value)
```



###### fn fieldConfig.defaults.custom.withPointSize

```ts
withPointSize(value)
```



###### fn fieldConfig.defaults.custom.withPointSymbol

```ts
withPointSymbol(value)
```



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

###### fn fieldConfig.defaults.custom.withShowPoints

```ts
withShowPoints(value)
```

TODO docs

Accepted values for `value` are "auto", "never", "always"

###### fn fieldConfig.defaults.custom.withSpanNulls

```ts
withSpanNulls(value)
```

Indicate if null values should be treated as gaps or connected.
When the value is a number, it represents the maximum delta in the
X axis that should be considered connected.  For timeseries, this is milliseconds

###### fn fieldConfig.defaults.custom.withSpanNullsMixin

```ts
withSpanNullsMixin(value)
```

Indicate if null values should be treated as gaps or connected.
When the value is a number, it represents the maximum delta in the
X axis that should be considered connected.  For timeseries, this is milliseconds

###### fn fieldConfig.defaults.custom.withStacking

```ts
withStacking(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withStackingMixin

```ts
withStackingMixin(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withThresholdsStyle

```ts
withThresholdsStyle(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withThresholdsStyleMixin

```ts
withThresholdsStyleMixin(value)
```

TODO docs

###### fn fieldConfig.defaults.custom.withTransform

```ts
withTransform(value)
```

TODO docs

Accepted values for `value` are "constant", "negative-Y"

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



###### obj fieldConfig.defaults.custom.lineStyle


####### fn fieldConfig.defaults.custom.lineStyle.withDash

```ts
withDash(value)
```



####### fn fieldConfig.defaults.custom.lineStyle.withDashMixin

```ts
withDashMixin(value)
```



####### fn fieldConfig.defaults.custom.lineStyle.withFill

```ts
withFill(value)
```



Accepted values for `value` are "solid", "dash", "dot", "square"

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

###### obj fieldConfig.defaults.custom.stacking


####### fn fieldConfig.defaults.custom.stacking.withGroup

```ts
withGroup(value)
```



####### fn fieldConfig.defaults.custom.stacking.withMode

```ts
withMode(value)
```

TODO docs

Accepted values for `value` are "none", "normal", "percent"

###### obj fieldConfig.defaults.custom.thresholdsStyle


####### fn fieldConfig.defaults.custom.thresholdsStyle.withMode

```ts
withMode(value)
```

TODO docs

Accepted values for `value` are "off", "line", "dashed", "area", "line+area", "dashed+area", "series"

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

#### fn options.withTimezone

```ts
withTimezone(value)
```



#### fn options.withTimezoneMixin

```ts
withTimezoneMixin(value)
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
