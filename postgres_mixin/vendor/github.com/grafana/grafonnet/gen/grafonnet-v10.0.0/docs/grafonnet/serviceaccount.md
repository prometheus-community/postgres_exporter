# serviceaccount

grafonnet.serviceaccount

## Index

* [`fn withAccessControl(value)`](#fn-withaccesscontrol)
* [`fn withAccessControlMixin(value)`](#fn-withaccesscontrolmixin)
* [`fn withAvatarUrl(value)`](#fn-withavatarurl)
* [`fn withId(value)`](#fn-withid)
* [`fn withIsDisabled(value)`](#fn-withisdisabled)
* [`fn withLogin(value)`](#fn-withlogin)
* [`fn withName(value)`](#fn-withname)
* [`fn withOrgId(value)`](#fn-withorgid)
* [`fn withRole(value)`](#fn-withrole)
* [`fn withTeams(value)`](#fn-withteams)
* [`fn withTeamsMixin(value)`](#fn-withteamsmixin)
* [`fn withTokens(value)`](#fn-withtokens)

## Fields

### fn withAccessControl

```ts
withAccessControl(value)
```

AccessControl metadata associated with a given resource.

### fn withAccessControlMixin

```ts
withAccessControlMixin(value)
```

AccessControl metadata associated with a given resource.

### fn withAvatarUrl

```ts
withAvatarUrl(value)
```

AvatarUrl is the service account's avatar URL. It allows the frontend to display a picture in front
of the service account.

### fn withId

```ts
withId(value)
```

ID is the unique identifier of the service account in the database.

### fn withIsDisabled

```ts
withIsDisabled(value)
```

IsDisabled indicates if the service account is disabled.

### fn withLogin

```ts
withLogin(value)
```

Login of the service account.

### fn withName

```ts
withName(value)
```

Name of the service account.

### fn withOrgId

```ts
withOrgId(value)
```

OrgId is the ID of an organisation the service account belongs to.

### fn withRole

```ts
withRole(value)
```

OrgRole is a Grafana Organization Role which can be 'Viewer', 'Editor', 'Admin'.

Accepted values for `value` are "Admin", "Editor", "Viewer"

### fn withTeams

```ts
withTeams(value)
```

Teams is a list of teams the service account belongs to.

### fn withTeamsMixin

```ts
withTeamsMixin(value)
```

Teams is a list of teams the service account belongs to.

### fn withTokens

```ts
withTokens(value)
```

Tokens is the number of active tokens for the service account.
Tokens are used to authenticate the service account against Grafana.
