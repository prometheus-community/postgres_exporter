# playlist

grafonnet.playlist

## Index

* [`fn withInterval(value="5m")`](#fn-withinterval)
* [`fn withItems(value)`](#fn-withitems)
* [`fn withItemsMixin(value)`](#fn-withitemsmixin)
* [`fn withName(value)`](#fn-withname)
* [`fn withUid(value)`](#fn-withuid)
* [`obj items`](#obj-items)
  * [`fn withTitle(value)`](#fn-itemswithtitle)
  * [`fn withType(value)`](#fn-itemswithtype)
  * [`fn withValue(value)`](#fn-itemswithvalue)

## Fields

### fn withInterval

```ts
withInterval(value="5m")
```

Interval sets the time between switching views in a playlist.
FIXME: Is this based on a standardized format or what options are available? Can datemath be used?

### fn withItems

```ts
withItems(value)
```

The ordered list of items that the playlist will iterate over.
FIXME! This should not be optional, but changing it makes the godegen awkward

### fn withItemsMixin

```ts
withItemsMixin(value)
```

The ordered list of items that the playlist will iterate over.
FIXME! This should not be optional, but changing it makes the godegen awkward

### fn withName

```ts
withName(value)
```

Name of the playlist.

### fn withUid

```ts
withUid(value)
```

Unique playlist identifier. Generated on creation, either by the
creator of the playlist of by the application.

### obj items


#### fn items.withTitle

```ts
withTitle(value)
```

Title is an unused property -- it will be removed in the future

#### fn items.withType

```ts
withType(value)
```

Type of the item.

Accepted values for `value` are "dashboard_by_uid", "dashboard_by_id", "dashboard_by_tag"

#### fn items.withValue

```ts
withValue(value)
```

Value depends on type and describes the playlist item.

 - dashboard_by_id: The value is an internal numerical identifier set by Grafana. This
 is not portable as the numerical identifier is non-deterministic between different instances.
 Will be replaced by dashboard_by_uid in the future. (deprecated)
 - dashboard_by_tag: The value is a tag which is set on any number of dashboards. All
 dashboards behind the tag will be added to the playlist.
 - dashboard_by_uid: The value is the dashboard UID
