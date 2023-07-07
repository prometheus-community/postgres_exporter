# elasticsearch

grafonnet.query.elasticsearch

## Index

* [`fn withAlias(value)`](#fn-withalias)
* [`fn withBucketAggs(value)`](#fn-withbucketaggs)
* [`fn withBucketAggsMixin(value)`](#fn-withbucketaggsmixin)
* [`fn withDatasource(value)`](#fn-withdatasource)
* [`fn withHide(value)`](#fn-withhide)
* [`fn withMetrics(value)`](#fn-withmetrics)
* [`fn withMetricsMixin(value)`](#fn-withmetricsmixin)
* [`fn withQuery(value)`](#fn-withquery)
* [`fn withQueryType(value)`](#fn-withquerytype)
* [`fn withRefId(value)`](#fn-withrefid)
* [`fn withTimeField(value)`](#fn-withtimefield)
* [`obj bucketAggs`](#obj-bucketaggs)
  * [`obj DateHistogram`](#obj-bucketaggsdatehistogram)
    * [`fn withField(value)`](#fn-bucketaggsdatehistogramwithfield)
    * [`fn withId(value)`](#fn-bucketaggsdatehistogramwithid)
    * [`fn withSettings(value)`](#fn-bucketaggsdatehistogramwithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggsdatehistogramwithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggsdatehistogramwithtype)
    * [`obj settings`](#obj-bucketaggsdatehistogramsettings)
      * [`fn withInterval(value)`](#fn-bucketaggsdatehistogramsettingswithinterval)
      * [`fn withMinDocCount(value)`](#fn-bucketaggsdatehistogramsettingswithmindoccount)
      * [`fn withOffset(value)`](#fn-bucketaggsdatehistogramsettingswithoffset)
      * [`fn withTimeZone(value)`](#fn-bucketaggsdatehistogramsettingswithtimezone)
      * [`fn withTrimEdges(value)`](#fn-bucketaggsdatehistogramsettingswithtrimedges)
  * [`obj Filters`](#obj-bucketaggsfilters)
    * [`fn withId(value)`](#fn-bucketaggsfilterswithid)
    * [`fn withSettings(value)`](#fn-bucketaggsfilterswithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggsfilterswithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggsfilterswithtype)
    * [`obj settings`](#obj-bucketaggsfilterssettings)
      * [`fn withFilters(value)`](#fn-bucketaggsfilterssettingswithfilters)
      * [`fn withFiltersMixin(value)`](#fn-bucketaggsfilterssettingswithfiltersmixin)
      * [`obj filters`](#obj-bucketaggsfilterssettingsfilters)
        * [`fn withLabel(value)`](#fn-bucketaggsfilterssettingsfilterswithlabel)
        * [`fn withQuery(value)`](#fn-bucketaggsfilterssettingsfilterswithquery)
  * [`obj GeoHashGrid`](#obj-bucketaggsgeohashgrid)
    * [`fn withField(value)`](#fn-bucketaggsgeohashgridwithfield)
    * [`fn withId(value)`](#fn-bucketaggsgeohashgridwithid)
    * [`fn withSettings(value)`](#fn-bucketaggsgeohashgridwithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggsgeohashgridwithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggsgeohashgridwithtype)
    * [`obj settings`](#obj-bucketaggsgeohashgridsettings)
      * [`fn withPrecision(value)`](#fn-bucketaggsgeohashgridsettingswithprecision)
  * [`obj Histogram`](#obj-bucketaggshistogram)
    * [`fn withField(value)`](#fn-bucketaggshistogramwithfield)
    * [`fn withId(value)`](#fn-bucketaggshistogramwithid)
    * [`fn withSettings(value)`](#fn-bucketaggshistogramwithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggshistogramwithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggshistogramwithtype)
    * [`obj settings`](#obj-bucketaggshistogramsettings)
      * [`fn withInterval(value)`](#fn-bucketaggshistogramsettingswithinterval)
      * [`fn withMinDocCount(value)`](#fn-bucketaggshistogramsettingswithmindoccount)
  * [`obj Nested`](#obj-bucketaggsnested)
    * [`fn withField(value)`](#fn-bucketaggsnestedwithfield)
    * [`fn withId(value)`](#fn-bucketaggsnestedwithid)
    * [`fn withSettings(value)`](#fn-bucketaggsnestedwithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggsnestedwithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggsnestedwithtype)
  * [`obj Terms`](#obj-bucketaggsterms)
    * [`fn withField(value)`](#fn-bucketaggstermswithfield)
    * [`fn withId(value)`](#fn-bucketaggstermswithid)
    * [`fn withSettings(value)`](#fn-bucketaggstermswithsettings)
    * [`fn withSettingsMixin(value)`](#fn-bucketaggstermswithsettingsmixin)
    * [`fn withType(value)`](#fn-bucketaggstermswithtype)
    * [`obj settings`](#obj-bucketaggstermssettings)
      * [`fn withMinDocCount(value)`](#fn-bucketaggstermssettingswithmindoccount)
      * [`fn withMissing(value)`](#fn-bucketaggstermssettingswithmissing)
      * [`fn withOrder(value)`](#fn-bucketaggstermssettingswithorder)
      * [`fn withOrderBy(value)`](#fn-bucketaggstermssettingswithorderby)
      * [`fn withSize(value)`](#fn-bucketaggstermssettingswithsize)
* [`obj metrics`](#obj-metrics)
  * [`obj Count`](#obj-metricscount)
    * [`fn withHide(value)`](#fn-metricscountwithhide)
    * [`fn withId(value)`](#fn-metricscountwithid)
    * [`fn withType(value)`](#fn-metricscountwithtype)
  * [`obj MetricAggregationWithSettings`](#obj-metricsmetricaggregationwithsettings)
    * [`obj Average`](#obj-metricsmetricaggregationwithsettingsaverage)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsaveragewithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsaveragesettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingsaveragesettingswithmissing)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsaveragesettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsaveragesettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsaveragesettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsaveragesettingsscriptwithinline)
    * [`obj BucketScript`](#obj-metricsmetricaggregationwithsettingsbucketscript)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithid)
      * [`fn withPipelineVariables(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithpipelinevariables)
      * [`fn withPipelineVariablesMixin(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithpipelinevariablesmixin)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptwithtype)
      * [`obj pipelineVariables`](#obj-metricsmetricaggregationwithsettingsbucketscriptpipelinevariables)
        * [`fn withName(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptpipelinevariableswithname)
        * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptpipelinevariableswithpipelineagg)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsbucketscriptsettings)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptsettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsbucketscriptsettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsbucketscriptsettingsscriptwithinline)
    * [`obj CumulativeSum`](#obj-metricsmetricaggregationwithsettingscumulativesum)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithid)
      * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingscumulativesumsettings)
        * [`fn withFormat(value)`](#fn-metricsmetricaggregationwithsettingscumulativesumsettingswithformat)
    * [`obj Derivative`](#obj-metricsmetricaggregationwithsettingsderivative)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithid)
      * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsderivativewithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsderivativesettings)
        * [`fn withUnit(value)`](#fn-metricsmetricaggregationwithsettingsderivativesettingswithunit)
    * [`obj ExtendedStats`](#obj-metricsmetricaggregationwithsettingsextendedstats)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithid)
      * [`fn withMeta(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithmeta)
      * [`fn withMetaMixin(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithmetamixin)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatswithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsextendedstatssettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatssettingswithmissing)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatssettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatssettingswithscriptmixin)
        * [`fn withSigma(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatssettingswithsigma)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsextendedstatssettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsextendedstatssettingsscriptwithinline)
    * [`obj Logs`](#obj-metricsmetricaggregationwithsettingslogs)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingslogswithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingslogswithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingslogswithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingslogswithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingslogswithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingslogssettings)
        * [`fn withLimit(value)`](#fn-metricsmetricaggregationwithsettingslogssettingswithlimit)
    * [`obj Max`](#obj-metricsmetricaggregationwithsettingsmax)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsmaxwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsmaxsettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingsmaxsettingswithmissing)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsmaxsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsmaxsettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsmaxsettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsmaxsettingsscriptwithinline)
    * [`obj Min`](#obj-metricsmetricaggregationwithsettingsmin)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsminwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsminwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsminwithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsminwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsminwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsminwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsminsettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingsminsettingswithmissing)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsminsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsminsettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsminsettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsminsettingsscriptwithinline)
    * [`obj MovingAverage`](#obj-metricsmetricaggregationwithsettingsmovingaverage)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithid)
      * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsmovingaveragewithtype)
    * [`obj MovingFunction`](#obj-metricsmetricaggregationwithsettingsmovingfunction)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithid)
      * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsmovingfunctionsettings)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionsettingswithscriptmixin)
        * [`fn withShift(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionsettingswithshift)
        * [`fn withWindow(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionsettingswithwindow)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingsmovingfunctionsettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingsmovingfunctionsettingsscriptwithinline)
    * [`obj Percentiles`](#obj-metricsmetricaggregationwithsettingspercentiles)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingspercentileswithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingspercentilessettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingswithmissing)
        * [`fn withPercents(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingswithpercents)
        * [`fn withPercentsMixin(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingswithpercentsmixin)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingspercentilessettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingspercentilessettingsscriptwithinline)
    * [`obj Rate`](#obj-metricsmetricaggregationwithsettingsrate)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsratewithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsratewithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsratewithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsratewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsratewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsratewithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsratesettings)
        * [`fn withMode(value)`](#fn-metricsmetricaggregationwithsettingsratesettingswithmode)
        * [`fn withUnit(value)`](#fn-metricsmetricaggregationwithsettingsratesettingswithunit)
    * [`obj RawData`](#obj-metricsmetricaggregationwithsettingsrawdata)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsrawdatawithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsrawdatawithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsrawdatawithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsrawdatawithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsrawdatawithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsrawdatasettings)
        * [`fn withSize(value)`](#fn-metricsmetricaggregationwithsettingsrawdatasettingswithsize)
    * [`obj RawDocument`](#obj-metricsmetricaggregationwithsettingsrawdocument)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentwithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsrawdocumentsettings)
        * [`fn withSize(value)`](#fn-metricsmetricaggregationwithsettingsrawdocumentsettingswithsize)
    * [`obj SerialDiff`](#obj-metricsmetricaggregationwithsettingsserialdiff)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithid)
      * [`fn withPipelineAgg(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsserialdiffsettings)
        * [`fn withLag(value)`](#fn-metricsmetricaggregationwithsettingsserialdiffsettingswithlag)
    * [`obj Sum`](#obj-metricsmetricaggregationwithsettingssum)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingssumwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingssumwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingssumwithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingssumwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingssumwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingssumwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingssumsettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingssumsettingswithmissing)
        * [`fn withScript(value)`](#fn-metricsmetricaggregationwithsettingssumsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricsmetricaggregationwithsettingssumsettingswithscriptmixin)
        * [`obj script`](#obj-metricsmetricaggregationwithsettingssumsettingsscript)
          * [`fn withInline(value)`](#fn-metricsmetricaggregationwithsettingssumsettingsscriptwithinline)
    * [`obj TopMetrics`](#obj-metricsmetricaggregationwithsettingstopmetrics)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingstopmetricswithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingstopmetricswithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingstopmetricswithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingstopmetricswithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingstopmetricswithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingstopmetricssettings)
        * [`fn withMetrics(value)`](#fn-metricsmetricaggregationwithsettingstopmetricssettingswithmetrics)
        * [`fn withMetricsMixin(value)`](#fn-metricsmetricaggregationwithsettingstopmetricssettingswithmetricsmixin)
        * [`fn withOrder(value)`](#fn-metricsmetricaggregationwithsettingstopmetricssettingswithorder)
        * [`fn withOrderBy(value)`](#fn-metricsmetricaggregationwithsettingstopmetricssettingswithorderby)
    * [`obj UniqueCount`](#obj-metricsmetricaggregationwithsettingsuniquecount)
      * [`fn withField(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithfield)
      * [`fn withHide(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithhide)
      * [`fn withId(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithid)
      * [`fn withSettings(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountwithtype)
      * [`obj settings`](#obj-metricsmetricaggregationwithsettingsuniquecountsettings)
        * [`fn withMissing(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountsettingswithmissing)
        * [`fn withPrecisionThreshold(value)`](#fn-metricsmetricaggregationwithsettingsuniquecountsettingswithprecisionthreshold)
  * [`obj PipelineMetricAggregation`](#obj-metricspipelinemetricaggregation)
    * [`obj BucketScript`](#obj-metricspipelinemetricaggregationbucketscript)
      * [`fn withHide(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithhide)
      * [`fn withId(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithid)
      * [`fn withPipelineVariables(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithpipelinevariables)
      * [`fn withPipelineVariablesMixin(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithpipelinevariablesmixin)
      * [`fn withSettings(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricspipelinemetricaggregationbucketscriptwithtype)
      * [`obj pipelineVariables`](#obj-metricspipelinemetricaggregationbucketscriptpipelinevariables)
        * [`fn withName(value)`](#fn-metricspipelinemetricaggregationbucketscriptpipelinevariableswithname)
        * [`fn withPipelineAgg(value)`](#fn-metricspipelinemetricaggregationbucketscriptpipelinevariableswithpipelineagg)
      * [`obj settings`](#obj-metricspipelinemetricaggregationbucketscriptsettings)
        * [`fn withScript(value)`](#fn-metricspipelinemetricaggregationbucketscriptsettingswithscript)
        * [`fn withScriptMixin(value)`](#fn-metricspipelinemetricaggregationbucketscriptsettingswithscriptmixin)
        * [`obj script`](#obj-metricspipelinemetricaggregationbucketscriptsettingsscript)
          * [`fn withInline(value)`](#fn-metricspipelinemetricaggregationbucketscriptsettingsscriptwithinline)
    * [`obj CumulativeSum`](#obj-metricspipelinemetricaggregationcumulativesum)
      * [`fn withField(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithfield)
      * [`fn withHide(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithhide)
      * [`fn withId(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithid)
      * [`fn withPipelineAgg(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithsettingsmixin)
      * [`fn withType(value)`](#fn-metricspipelinemetricaggregationcumulativesumwithtype)
      * [`obj settings`](#obj-metricspipelinemetricaggregationcumulativesumsettings)
        * [`fn withFormat(value)`](#fn-metricspipelinemetricaggregationcumulativesumsettingswithformat)
    * [`obj Derivative`](#obj-metricspipelinemetricaggregationderivative)
      * [`fn withField(value)`](#fn-metricspipelinemetricaggregationderivativewithfield)
      * [`fn withHide(value)`](#fn-metricspipelinemetricaggregationderivativewithhide)
      * [`fn withId(value)`](#fn-metricspipelinemetricaggregationderivativewithid)
      * [`fn withPipelineAgg(value)`](#fn-metricspipelinemetricaggregationderivativewithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricspipelinemetricaggregationderivativewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricspipelinemetricaggregationderivativewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricspipelinemetricaggregationderivativewithtype)
      * [`obj settings`](#obj-metricspipelinemetricaggregationderivativesettings)
        * [`fn withUnit(value)`](#fn-metricspipelinemetricaggregationderivativesettingswithunit)
    * [`obj MovingAverage`](#obj-metricspipelinemetricaggregationmovingaverage)
      * [`fn withField(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithfield)
      * [`fn withHide(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithhide)
      * [`fn withId(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithid)
      * [`fn withPipelineAgg(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithpipelineagg)
      * [`fn withSettings(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithsettings)
      * [`fn withSettingsMixin(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithsettingsmixin)
      * [`fn withType(value)`](#fn-metricspipelinemetricaggregationmovingaveragewithtype)

## Fields

### fn withAlias

```ts
withAlias(value)
```

Alias pattern

### fn withBucketAggs

```ts
withBucketAggs(value)
```

List of bucket aggregations

### fn withBucketAggsMixin

```ts
withBucketAggsMixin(value)
```

List of bucket aggregations

### fn withDatasource

```ts
withDatasource(value)
```

For mixed data sources the selected datasource is on the query level.
For non mixed scenarios this is undefined.
TODO find a better way to do this ^ that's friendly to schema
TODO this shouldn't be unknown but DataSourceRef | null

### fn withHide

```ts
withHide(value)
```

true if query is disabled (ie should not be returned to the dashboard)
Note this does not always imply that the query should not be executed since
the results from a hidden query may be used as the input to other queries (SSE etc)

### fn withMetrics

```ts
withMetrics(value)
```

List of metric aggregations

### fn withMetricsMixin

```ts
withMetricsMixin(value)
```

List of metric aggregations

### fn withQuery

```ts
withQuery(value)
```

Lucene query

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

### fn withTimeField

```ts
withTimeField(value)
```

Name of time field

### obj bucketAggs


#### obj bucketAggs.DateHistogram


##### fn bucketAggs.DateHistogram.withField

```ts
withField(value)
```



##### fn bucketAggs.DateHistogram.withId

```ts
withId(value)
```



##### fn bucketAggs.DateHistogram.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.DateHistogram.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.DateHistogram.withType

```ts
withType(value)
```



##### obj bucketAggs.DateHistogram.settings


###### fn bucketAggs.DateHistogram.settings.withInterval

```ts
withInterval(value)
```



###### fn bucketAggs.DateHistogram.settings.withMinDocCount

```ts
withMinDocCount(value)
```



###### fn bucketAggs.DateHistogram.settings.withOffset

```ts
withOffset(value)
```



###### fn bucketAggs.DateHistogram.settings.withTimeZone

```ts
withTimeZone(value)
```



###### fn bucketAggs.DateHistogram.settings.withTrimEdges

```ts
withTrimEdges(value)
```



#### obj bucketAggs.Filters


##### fn bucketAggs.Filters.withId

```ts
withId(value)
```



##### fn bucketAggs.Filters.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.Filters.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.Filters.withType

```ts
withType(value)
```



##### obj bucketAggs.Filters.settings


###### fn bucketAggs.Filters.settings.withFilters

```ts
withFilters(value)
```



###### fn bucketAggs.Filters.settings.withFiltersMixin

```ts
withFiltersMixin(value)
```



###### obj bucketAggs.Filters.settings.filters


####### fn bucketAggs.Filters.settings.filters.withLabel

```ts
withLabel(value)
```



####### fn bucketAggs.Filters.settings.filters.withQuery

```ts
withQuery(value)
```



#### obj bucketAggs.GeoHashGrid


##### fn bucketAggs.GeoHashGrid.withField

```ts
withField(value)
```



##### fn bucketAggs.GeoHashGrid.withId

```ts
withId(value)
```



##### fn bucketAggs.GeoHashGrid.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.GeoHashGrid.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.GeoHashGrid.withType

```ts
withType(value)
```



##### obj bucketAggs.GeoHashGrid.settings


###### fn bucketAggs.GeoHashGrid.settings.withPrecision

```ts
withPrecision(value)
```



#### obj bucketAggs.Histogram


##### fn bucketAggs.Histogram.withField

```ts
withField(value)
```



##### fn bucketAggs.Histogram.withId

```ts
withId(value)
```



##### fn bucketAggs.Histogram.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.Histogram.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.Histogram.withType

```ts
withType(value)
```



##### obj bucketAggs.Histogram.settings


###### fn bucketAggs.Histogram.settings.withInterval

```ts
withInterval(value)
```



###### fn bucketAggs.Histogram.settings.withMinDocCount

```ts
withMinDocCount(value)
```



#### obj bucketAggs.Nested


##### fn bucketAggs.Nested.withField

```ts
withField(value)
```



##### fn bucketAggs.Nested.withId

```ts
withId(value)
```



##### fn bucketAggs.Nested.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.Nested.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.Nested.withType

```ts
withType(value)
```



#### obj bucketAggs.Terms


##### fn bucketAggs.Terms.withField

```ts
withField(value)
```



##### fn bucketAggs.Terms.withId

```ts
withId(value)
```



##### fn bucketAggs.Terms.withSettings

```ts
withSettings(value)
```



##### fn bucketAggs.Terms.withSettingsMixin

```ts
withSettingsMixin(value)
```



##### fn bucketAggs.Terms.withType

```ts
withType(value)
```



##### obj bucketAggs.Terms.settings


###### fn bucketAggs.Terms.settings.withMinDocCount

```ts
withMinDocCount(value)
```



###### fn bucketAggs.Terms.settings.withMissing

```ts
withMissing(value)
```



###### fn bucketAggs.Terms.settings.withOrder

```ts
withOrder(value)
```



Accepted values for `value` are "desc", "asc"

###### fn bucketAggs.Terms.settings.withOrderBy

```ts
withOrderBy(value)
```



###### fn bucketAggs.Terms.settings.withSize

```ts
withSize(value)
```



### obj metrics


#### obj metrics.Count


##### fn metrics.Count.withHide

```ts
withHide(value)
```



##### fn metrics.Count.withId

```ts
withId(value)
```



##### fn metrics.Count.withType

```ts
withType(value)
```



#### obj metrics.MetricAggregationWithSettings


##### obj metrics.MetricAggregationWithSettings.Average


###### fn metrics.MetricAggregationWithSettings.Average.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Average.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Average.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Average.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Average.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Average.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Average.settings


####### fn metrics.MetricAggregationWithSettings.Average.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.Average.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.Average.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.Average.settings.script


######## fn metrics.MetricAggregationWithSettings.Average.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.BucketScript


###### fn metrics.MetricAggregationWithSettings.BucketScript.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withPipelineVariables

```ts
withPipelineVariables(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withPipelineVariablesMixin

```ts
withPipelineVariablesMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.BucketScript.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.BucketScript.pipelineVariables


####### fn metrics.MetricAggregationWithSettings.BucketScript.pipelineVariables.withName

```ts
withName(value)
```



####### fn metrics.MetricAggregationWithSettings.BucketScript.pipelineVariables.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### obj metrics.MetricAggregationWithSettings.BucketScript.settings


####### fn metrics.MetricAggregationWithSettings.BucketScript.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.BucketScript.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.BucketScript.settings.script


######## fn metrics.MetricAggregationWithSettings.BucketScript.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.CumulativeSum


###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.CumulativeSum.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.CumulativeSum.settings


####### fn metrics.MetricAggregationWithSettings.CumulativeSum.settings.withFormat

```ts
withFormat(value)
```



##### obj metrics.MetricAggregationWithSettings.Derivative


###### fn metrics.MetricAggregationWithSettings.Derivative.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Derivative.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Derivative.settings


####### fn metrics.MetricAggregationWithSettings.Derivative.settings.withUnit

```ts
withUnit(value)
```



##### obj metrics.MetricAggregationWithSettings.ExtendedStats


###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withMeta

```ts
withMeta(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withMetaMixin

```ts
withMetaMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.ExtendedStats.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.ExtendedStats.settings


####### fn metrics.MetricAggregationWithSettings.ExtendedStats.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.ExtendedStats.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.ExtendedStats.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### fn metrics.MetricAggregationWithSettings.ExtendedStats.settings.withSigma

```ts
withSigma(value)
```



####### obj metrics.MetricAggregationWithSettings.ExtendedStats.settings.script


######## fn metrics.MetricAggregationWithSettings.ExtendedStats.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.Logs


###### fn metrics.MetricAggregationWithSettings.Logs.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Logs.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Logs.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Logs.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Logs.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Logs.settings


####### fn metrics.MetricAggregationWithSettings.Logs.settings.withLimit

```ts
withLimit(value)
```



##### obj metrics.MetricAggregationWithSettings.Max


###### fn metrics.MetricAggregationWithSettings.Max.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Max.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Max.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Max.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Max.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Max.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Max.settings


####### fn metrics.MetricAggregationWithSettings.Max.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.Max.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.Max.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.Max.settings.script


######## fn metrics.MetricAggregationWithSettings.Max.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.Min


###### fn metrics.MetricAggregationWithSettings.Min.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Min.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Min.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Min.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Min.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Min.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Min.settings


####### fn metrics.MetricAggregationWithSettings.Min.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.Min.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.Min.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.Min.settings.script


######## fn metrics.MetricAggregationWithSettings.Min.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.MovingAverage


###### fn metrics.MetricAggregationWithSettings.MovingAverage.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingAverage.withType

```ts
withType(value)
```



##### obj metrics.MetricAggregationWithSettings.MovingFunction


###### fn metrics.MetricAggregationWithSettings.MovingFunction.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.MovingFunction.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.MovingFunction.settings


####### fn metrics.MetricAggregationWithSettings.MovingFunction.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.MovingFunction.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### fn metrics.MetricAggregationWithSettings.MovingFunction.settings.withShift

```ts
withShift(value)
```



####### fn metrics.MetricAggregationWithSettings.MovingFunction.settings.withWindow

```ts
withWindow(value)
```



####### obj metrics.MetricAggregationWithSettings.MovingFunction.settings.script


######## fn metrics.MetricAggregationWithSettings.MovingFunction.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.Percentiles


###### fn metrics.MetricAggregationWithSettings.Percentiles.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Percentiles.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Percentiles.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Percentiles.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Percentiles.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Percentiles.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Percentiles.settings


####### fn metrics.MetricAggregationWithSettings.Percentiles.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.Percentiles.settings.withPercents

```ts
withPercents(value)
```



####### fn metrics.MetricAggregationWithSettings.Percentiles.settings.withPercentsMixin

```ts
withPercentsMixin(value)
```



####### fn metrics.MetricAggregationWithSettings.Percentiles.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.Percentiles.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.Percentiles.settings.script


######## fn metrics.MetricAggregationWithSettings.Percentiles.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.Rate


###### fn metrics.MetricAggregationWithSettings.Rate.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Rate.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Rate.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Rate.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Rate.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Rate.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Rate.settings


####### fn metrics.MetricAggregationWithSettings.Rate.settings.withMode

```ts
withMode(value)
```



####### fn metrics.MetricAggregationWithSettings.Rate.settings.withUnit

```ts
withUnit(value)
```



##### obj metrics.MetricAggregationWithSettings.RawData


###### fn metrics.MetricAggregationWithSettings.RawData.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.RawData.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.RawData.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.RawData.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.RawData.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.RawData.settings


####### fn metrics.MetricAggregationWithSettings.RawData.settings.withSize

```ts
withSize(value)
```



##### obj metrics.MetricAggregationWithSettings.RawDocument


###### fn metrics.MetricAggregationWithSettings.RawDocument.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.RawDocument.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.RawDocument.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.RawDocument.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.RawDocument.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.RawDocument.settings


####### fn metrics.MetricAggregationWithSettings.RawDocument.settings.withSize

```ts
withSize(value)
```



##### obj metrics.MetricAggregationWithSettings.SerialDiff


###### fn metrics.MetricAggregationWithSettings.SerialDiff.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.SerialDiff.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.SerialDiff.settings


####### fn metrics.MetricAggregationWithSettings.SerialDiff.settings.withLag

```ts
withLag(value)
```



##### obj metrics.MetricAggregationWithSettings.Sum


###### fn metrics.MetricAggregationWithSettings.Sum.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.Sum.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.Sum.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.Sum.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.Sum.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.Sum.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.Sum.settings


####### fn metrics.MetricAggregationWithSettings.Sum.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.Sum.settings.withScript

```ts
withScript(value)
```



####### fn metrics.MetricAggregationWithSettings.Sum.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.MetricAggregationWithSettings.Sum.settings.script


######## fn metrics.MetricAggregationWithSettings.Sum.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.MetricAggregationWithSettings.TopMetrics


###### fn metrics.MetricAggregationWithSettings.TopMetrics.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.TopMetrics.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.TopMetrics.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.TopMetrics.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.TopMetrics.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.TopMetrics.settings


####### fn metrics.MetricAggregationWithSettings.TopMetrics.settings.withMetrics

```ts
withMetrics(value)
```



####### fn metrics.MetricAggregationWithSettings.TopMetrics.settings.withMetricsMixin

```ts
withMetricsMixin(value)
```



####### fn metrics.MetricAggregationWithSettings.TopMetrics.settings.withOrder

```ts
withOrder(value)
```



####### fn metrics.MetricAggregationWithSettings.TopMetrics.settings.withOrderBy

```ts
withOrderBy(value)
```



##### obj metrics.MetricAggregationWithSettings.UniqueCount


###### fn metrics.MetricAggregationWithSettings.UniqueCount.withField

```ts
withField(value)
```



###### fn metrics.MetricAggregationWithSettings.UniqueCount.withHide

```ts
withHide(value)
```



###### fn metrics.MetricAggregationWithSettings.UniqueCount.withId

```ts
withId(value)
```



###### fn metrics.MetricAggregationWithSettings.UniqueCount.withSettings

```ts
withSettings(value)
```



###### fn metrics.MetricAggregationWithSettings.UniqueCount.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.MetricAggregationWithSettings.UniqueCount.withType

```ts
withType(value)
```



###### obj metrics.MetricAggregationWithSettings.UniqueCount.settings


####### fn metrics.MetricAggregationWithSettings.UniqueCount.settings.withMissing

```ts
withMissing(value)
```



####### fn metrics.MetricAggregationWithSettings.UniqueCount.settings.withPrecisionThreshold

```ts
withPrecisionThreshold(value)
```



#### obj metrics.PipelineMetricAggregation


##### obj metrics.PipelineMetricAggregation.BucketScript


###### fn metrics.PipelineMetricAggregation.BucketScript.withHide

```ts
withHide(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withId

```ts
withId(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withPipelineVariables

```ts
withPipelineVariables(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withPipelineVariablesMixin

```ts
withPipelineVariablesMixin(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withSettings

```ts
withSettings(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.PipelineMetricAggregation.BucketScript.withType

```ts
withType(value)
```



###### obj metrics.PipelineMetricAggregation.BucketScript.pipelineVariables


####### fn metrics.PipelineMetricAggregation.BucketScript.pipelineVariables.withName

```ts
withName(value)
```



####### fn metrics.PipelineMetricAggregation.BucketScript.pipelineVariables.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### obj metrics.PipelineMetricAggregation.BucketScript.settings


####### fn metrics.PipelineMetricAggregation.BucketScript.settings.withScript

```ts
withScript(value)
```



####### fn metrics.PipelineMetricAggregation.BucketScript.settings.withScriptMixin

```ts
withScriptMixin(value)
```



####### obj metrics.PipelineMetricAggregation.BucketScript.settings.script


######## fn metrics.PipelineMetricAggregation.BucketScript.settings.script.withInline

```ts
withInline(value)
```



##### obj metrics.PipelineMetricAggregation.CumulativeSum


###### fn metrics.PipelineMetricAggregation.CumulativeSum.withField

```ts
withField(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withHide

```ts
withHide(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withId

```ts
withId(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withSettings

```ts
withSettings(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.PipelineMetricAggregation.CumulativeSum.withType

```ts
withType(value)
```



###### obj metrics.PipelineMetricAggregation.CumulativeSum.settings


####### fn metrics.PipelineMetricAggregation.CumulativeSum.settings.withFormat

```ts
withFormat(value)
```



##### obj metrics.PipelineMetricAggregation.Derivative


###### fn metrics.PipelineMetricAggregation.Derivative.withField

```ts
withField(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withHide

```ts
withHide(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withId

```ts
withId(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withSettings

```ts
withSettings(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.PipelineMetricAggregation.Derivative.withType

```ts
withType(value)
```



###### obj metrics.PipelineMetricAggregation.Derivative.settings


####### fn metrics.PipelineMetricAggregation.Derivative.settings.withUnit

```ts
withUnit(value)
```



##### obj metrics.PipelineMetricAggregation.MovingAverage


###### fn metrics.PipelineMetricAggregation.MovingAverage.withField

```ts
withField(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withHide

```ts
withHide(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withId

```ts
withId(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withPipelineAgg

```ts
withPipelineAgg(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withSettings

```ts
withSettings(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withSettingsMixin

```ts
withSettingsMixin(value)
```



###### fn metrics.PipelineMetricAggregation.MovingAverage.withType

```ts
withType(value)
```


