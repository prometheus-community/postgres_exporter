# librarypanel

grafonnet.librarypanel

## Index

* [`fn withDescription(value)`](#fn-withdescription)
* [`fn withFolderUid(value)`](#fn-withfolderuid)
* [`fn withMeta(value)`](#fn-withmeta)
* [`fn withMetaMixin(value)`](#fn-withmetamixin)
* [`fn withModel(value)`](#fn-withmodel)
* [`fn withModelMixin(value)`](#fn-withmodelmixin)
* [`fn withName(value)`](#fn-withname)
* [`fn withSchemaVersion(value)`](#fn-withschemaversion)
* [`fn withType(value)`](#fn-withtype)
* [`fn withUid(value)`](#fn-withuid)
* [`fn withVersion(value)`](#fn-withversion)
* [`obj meta`](#obj-meta)
  * [`fn withConnectedDashboards(value)`](#fn-metawithconnecteddashboards)
  * [`fn withCreated(value)`](#fn-metawithcreated)
  * [`fn withCreatedBy(value)`](#fn-metawithcreatedby)
  * [`fn withCreatedByMixin(value)`](#fn-metawithcreatedbymixin)
  * [`fn withFolderName(value)`](#fn-metawithfoldername)
  * [`fn withFolderUid(value)`](#fn-metawithfolderuid)
  * [`fn withUpdated(value)`](#fn-metawithupdated)
  * [`fn withUpdatedBy(value)`](#fn-metawithupdatedby)
  * [`fn withUpdatedByMixin(value)`](#fn-metawithupdatedbymixin)
  * [`obj createdBy`](#obj-metacreatedby)
    * [`fn withAvatarUrl(value)`](#fn-metacreatedbywithavatarurl)
    * [`fn withId(value)`](#fn-metacreatedbywithid)
    * [`fn withName(value)`](#fn-metacreatedbywithname)
  * [`obj updatedBy`](#obj-metaupdatedby)
    * [`fn withAvatarUrl(value)`](#fn-metaupdatedbywithavatarurl)
    * [`fn withId(value)`](#fn-metaupdatedbywithid)
    * [`fn withName(value)`](#fn-metaupdatedbywithname)

## Fields

### fn withDescription

```ts
withDescription(value)
```

Panel description

### fn withFolderUid

```ts
withFolderUid(value)
```

Folder UID

### fn withMeta

```ts
withMeta(value)
```



### fn withMetaMixin

```ts
withMetaMixin(value)
```



### fn withModel

```ts
withModel(value)
```

TODO: should be the same panel schema defined in dashboard
Typescript: Omit<Panel, 'gridPos' | 'id' | 'libraryPanel'>;

### fn withModelMixin

```ts
withModelMixin(value)
```

TODO: should be the same panel schema defined in dashboard
Typescript: Omit<Panel, 'gridPos' | 'id' | 'libraryPanel'>;

### fn withName

```ts
withName(value)
```

Panel name (also saved in the model)

### fn withSchemaVersion

```ts
withSchemaVersion(value)
```

Dashboard version when this was saved (zero if unknown)

### fn withType

```ts
withType(value)
```

The panel type (from inside the model)

### fn withUid

```ts
withUid(value)
```

Library element UID

### fn withVersion

```ts
withVersion(value)
```

panel version, incremented each time the dashboard is updated.

### obj meta


#### fn meta.withConnectedDashboards

```ts
withConnectedDashboards(value)
```



#### fn meta.withCreated

```ts
withCreated(value)
```



#### fn meta.withCreatedBy

```ts
withCreatedBy(value)
```



#### fn meta.withCreatedByMixin

```ts
withCreatedByMixin(value)
```



#### fn meta.withFolderName

```ts
withFolderName(value)
```



#### fn meta.withFolderUid

```ts
withFolderUid(value)
```



#### fn meta.withUpdated

```ts
withUpdated(value)
```



#### fn meta.withUpdatedBy

```ts
withUpdatedBy(value)
```



#### fn meta.withUpdatedByMixin

```ts
withUpdatedByMixin(value)
```



#### obj meta.createdBy


##### fn meta.createdBy.withAvatarUrl

```ts
withAvatarUrl(value)
```



##### fn meta.createdBy.withId

```ts
withId(value)
```



##### fn meta.createdBy.withName

```ts
withName(value)
```



#### obj meta.updatedBy


##### fn meta.updatedBy.withAvatarUrl

```ts
withAvatarUrl(value)
```



##### fn meta.updatedBy.withId

```ts
withId(value)
```



##### fn meta.updatedBy.withName

```ts
withName(value)
```


