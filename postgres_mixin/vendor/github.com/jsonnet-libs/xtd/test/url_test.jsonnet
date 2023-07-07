local xtd = import '../main.libsonnet';
local test = import 'github.com/jsonnet-libs/testonnet/main.libsonnet';

test.new(std.thisFile)
+ test.case.new(
  name='empty',
  test=test.expect.eq(
    actual=xtd.url.escapeString(''),
    expected='',
  )
)

+ test.case.new(
  name='abc',
  test=test.expect.eq(
    actual=xtd.url.escapeString('abc'),
    expected='abc',
  )
)

+ test.case.new(
  name='space',
  test=test.expect.eq(
    actual=xtd.url.escapeString('one two'),
    expected='one%20two',
  )
)

+ test.case.new(
  name='percent',
  test=test.expect.eq(
    actual=xtd.url.escapeString('10%'),
    expected='10%25',
  )
)

+ test.case.new(
  name='complex',
  test=test.expect.eq(
    actual=xtd.url.escapeString(" ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;"),
    expected='%20%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B',
  )
)

+ test.case.new(
  name='exclusions',
  test=test.expect.eq(
    actual=xtd.url.escapeString('hello, world', [',']),
    expected='hello,%20world',
  )
)

+ test.case.new(
  name='multiple exclusions',
  test=test.expect.eq(
    actual=xtd.url.escapeString('hello, world,&', [',', '&']),
    expected='hello,%20world,&',
  )
)

+ test.case.new(
  name='empty',
  test=test.expect.eq(
    actual=xtd.url.encodeQuery({}),
    expected='',
  )
)

+ test.case.new(
  name='simple',
  test=test.expect.eq(
    actual=xtd.url.encodeQuery({ q: 'puppies', oe: 'utf8' }),
    expected='oe=utf8&q=puppies',
  )
)
