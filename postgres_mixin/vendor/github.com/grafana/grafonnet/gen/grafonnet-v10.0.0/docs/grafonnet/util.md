# util

Helper functions that work well with Grafonnet.

## Index

* [`obj dashboard`](#obj-dashboard)
  * [`fn getOptionsForCustomQuery(query)`](#fn-dashboardgetoptionsforcustomquery)
* [`obj grid`](#obj-grid)
  * [`fn makeGrid(panels, panelWidth, panelHeight)`](#fn-gridmakegrid)
* [`obj panel`](#obj-panel)
  * [`fn setPanelIDs(panels)`](#fn-panelsetpanelids)
* [`obj string`](#obj-string)
  * [`fn slugify(string)`](#fn-stringslugify)

## Fields

### obj dashboard


#### fn dashboard.getOptionsForCustomQuery

```ts
getOptionsForCustomQuery(query)
```

`getOptionsForCustomQuery` provides values for the `options` and `current` fields.
These are required for template variables of type 'custom'but do not automatically
get populated by Grafana when importing a dashboard from JSON.

This is a bit of a hack and should always be called on functions that set `type` on
a template variable (see the dashboard.templating.list veneer). Ideally Grafana
populates these fields from the `query` value but this provides a backwards
compatible solution.


### obj grid


#### fn grid.makeGrid

```ts
makeGrid(panels, panelWidth, panelHeight)
```

`makeGrid` returns an array of `panels` organized in a grid with equal `panelWidth`
and `panelHeight`. Row panels are used as "linebreaks", if a Row panel is collapsed,
then all panels below it will be folded into the row.

This function will use the full grid of 24 columns, setting `panelWidth` to a value
that can divide 24 into equal parts will fill up the page nicely. (1, 2, 3, 4, 6, 8, 12)
Other value for `panelWidth` will leave a gap on the far right.


### obj panel


#### fn panel.setPanelIDs

```ts
setPanelIDs(panels)
```

`setPanelIDs` ensures that all `panels` have a unique ID, this functions is used in
`dashboard.withPanels` and `dashboard.withPanelsMixin` to provide a consistent
experience.

used in ../veneer/dashboard.libsonnet


### obj string


#### fn string.slugify

```ts
slugify(string)
```

`slugify` will create a simple slug from `string`, keeping only alphanumeric
characters and replacing spaces with dashes.

