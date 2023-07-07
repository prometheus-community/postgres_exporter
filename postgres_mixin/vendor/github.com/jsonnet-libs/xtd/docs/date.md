---
permalink: /date/
---

# package date

```jsonnet
local date = import "github.com/jsonnet-libs/xtd/date.libsonnet"
```

`time` provides various date related functions.

## Index

* [`fn dayOfWeek(year, month, day)`](#fn-dayofweek)
* [`fn dayOfYear(year, month, day)`](#fn-dayofyear)
* [`fn isLeapYear(year)`](#fn-isleapyear)

## Fields

### fn dayOfWeek

```ts
dayOfWeek(year, month, day)
```

`dayOfWeek` returns the day of the week for the given date. 0=Sunday, 1=Monday, etc.

### fn dayOfYear

```ts
dayOfYear(year, month, day)
```

`dayOfYear` calculates the ordinal day of the year based on the given date. The range of outputs is 1-365
for common years, and 1-366 for leap years.


### fn isLeapYear

```ts
isLeapYear(year)
```

`isLeapYear` returns true if the given year is a leap year.