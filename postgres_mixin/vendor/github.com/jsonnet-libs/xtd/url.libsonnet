local d = import 'doc-util/main.libsonnet';

{
  '#': d.pkg(
    name='url',
    url='github.com/jsonnet-libs/xtd/url.libsonnet',
    help='`url` implements URL escaping and query building',
  ),

  '#escapeString': d.fn('`escapeString` escapes the given string so it can be safely placed inside an URL, replacing special characters with `%XX` sequences', [d.arg('str', d.T.string), d.arg('excludedChars', d.T.array)]),
  escapeString(str, excludedChars=[])::
    local allowedChars = '0123456789abcdefghijklmnopqrstuvwqxyzABCDEFGHIJKLMNOPQRSTUVWQXYZ';
    local utf8(char) = std.foldl(function(a, b) a + '%%%02X' % b, std.encodeUTF8(char), '');
    local escapeChar(char) = if std.member(excludedChars, char) || std.member(allowedChars, char) then char else utf8(char);
    std.join('', std.map(escapeChar, std.stringChars(str))),

  '#encodeQuery': d.fn('`encodeQuery` takes an object of query parameters and returns them as an escaped `key=value` string', [d.arg('params', d.T.object)]),
  encodeQuery(params)::
    local fmtParam(p) = '%s=%s' % [self.escapeString(p), self.escapeString(params[p])];
    std.join('&', std.map(fmtParam, std.objectFields(params))),
}
