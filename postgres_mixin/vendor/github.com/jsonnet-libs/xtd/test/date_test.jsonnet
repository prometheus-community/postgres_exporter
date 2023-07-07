local xtd = import '../main.libsonnet';
local test = import 'github.com/jsonnet-libs/testonnet/main.libsonnet';

test.new(std.thisFile)

+ test.case.new(
  name='Leap Year commonYear',
  test=test.expect.eq(
    actual=xtd.date.isLeapYear(1995),
    expected=false,
  )
)

+ test.case.new(
  name='Leap Year fourYearCycle',
  test=test.expect.eq(
    actual=xtd.date.isLeapYear(1996),
    expected=true,
  )
)

+ test.case.new(
  name='Leap Year fourHundredYearCycle',
  test=test.expect.eq(
    actual=xtd.date.isLeapYear(2000),
    expected=true,
  )
)

+ test.case.new(
  name='Leap Year hundredYearCycle',
  test=test.expect.eq(
    actual=xtd.date.isLeapYear(2100),
    expected=false,
  )
)

+ test.case.new(
  name='Day Of Week leapYearStart',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(2000, 1, 1),
    expected=6,
  )
)

+ test.case.new(
  name='Day Of Week leapYearEnd',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(2000, 12, 31),
    expected=0,
  )
)

+ test.case.new(
  name='Day Of Week commonYearStart',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(1995, 1, 1),
    expected=0,
  )
)

+ test.case.new(
  name='Day Of Week commonYearEnd',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(2003, 12, 31),
    expected=3,
  )
)

+ test.case.new(
  name='Day Of Week leapYearMid',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(2024, 7, 19),
    expected=5,
  )
)

+ test.case.new(
  name='Day Of Week commonYearMid',
  test=test.expect.eq(
    actual=xtd.date.dayOfWeek(2023, 6, 15),
    expected=4,
  )
)
+ test.case.new(
  name='Day Of Year leapYearStart',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(2000, 1, 1),
    expected=1,
  )
)

+ test.case.new(
  name='Day Of Year leapYearEnd',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(2000, 12, 31),
    expected=366,
  )
)

+ test.case.new(
  name='Day Of Year commonYearStart',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(1995, 1, 1),
    expected=1,
  )
)

+ test.case.new(
  name='Day Of Year commonYearEnd',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(2003, 12, 31),
    expected=365,
  )
)

+ test.case.new(
  name='Day Of Year leapYearMid',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(2024, 7, 19),
    expected=201,
  )
)

+ test.case.new(
  name='Day Of Year commonYearMid',
  test=test.expect.eq(
    actual=xtd.date.dayOfYear(2023, 6, 15),
    expected=166,
  )
)
