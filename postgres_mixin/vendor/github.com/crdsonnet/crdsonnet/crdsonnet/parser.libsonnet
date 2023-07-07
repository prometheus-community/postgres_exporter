local schemadb_util = import './schemadb.libsonnet';
{
  local this = self,

  getRefName(ref): std.reverse(std.split(ref, '/'))[0],

  getURIBase(uri): std.join('/', std.splitLimit(uri, '/', 5)[0:3]),
  getURIPath(uri): '/' + std.join('/', std.splitLimit(uri, '/', 5)[3:]),

  parseSchema(key, schema, currentSchema, schemaDB={}, parents=[]):
    // foldStart
    if std.isBoolean(schema)
    then { [key]+: schema }
    else if !std.isObject(schema)
    then error 'Schema is not an object or boolean'
    else
      local schemaToParse =
        if '$ref' in schema
        then this.resolveRef(
          schema['$ref'],
          currentSchema,
          schemaDB
        )
        else schema;

      // shortcut to make it more readable below
      // requires the parseSchema* functions to have the same signature
      local parse(k, f) =
        (if k in schemaToParse
         then
           local parsed = f(
             key,
             schemaToParse[k],
             currentSchema,
             schemaDB,
             parents,
           );
           if parsed != null
           then { [k]: parsed }
           else {}
         else {});

      {
        [key]+:
          schemaToParse
          + parse('properties', this.parseSchemaMap)
          + parse('patternProperties', this.parseSchemaMap)
          + parse('items', this.parseSchemaItems)
          + parse('then', this.parseSchemaSingle)
          + parse('else', this.parseSchemaSingle)
          + parse('prefixItems', this.parseSchemaList)
          + parse('allOf', this.parseSchemaList)
          + parse('anyOf', this.parseSchemaList)
          + parse('oneOf', this.parseSchemaList)
          + { _parents:: parents },
      }
  ,
  // foldEnd

  parseSchemaItems(key, schema, currentSchema, schemaDB, parents):
    self.parseSchemaSingle(key, schema, currentSchema, schemaDB, []),

  parseSchemaSingle(key, schema, currentSchema, schemaDB, parents):
    // foldStart
    local i =
      if std.length(parents) > 0
      then std.length(parents) - 1
      else 0;

    local parsed =
      this.parseSchema(
        key,
        schema,
        currentSchema,
        schemaDB,
        parents[0:i]
      );
    if parsed != null
    then
      if std.isObject(parsed[key])
      then parsed[key] + { _name:: key }
      else parsed[key]
    else {},
  // foldEnd

  parseSchemaMap(key, map, currentSchema, schemaDB, parents):
    // foldStart
    std.foldl(
      function(acc, k)
        acc
        + this.parseSchema(
          k,
          map[k],
          currentSchema,
          schemaDB,
          parents + [k],
        )
        + { [k]+: { _name:: k } },
      std.objectFields(map),
      {}
    ),
  // foldEnd

  parseSchemaList(key, list, currentSchema, schemaDB, parents):
    // foldStart
    [
      local parsed =
        this.parseSchema(
          key,
          item,
          currentSchema,
          schemaDB,
          parents,
        )[key];

      // Due to the nature of list items in JSON they don't have a key we can use as
      // a name. However we can deduct the name from $ref or use $anchor if those are
      // available. The name can later be used to create proper functions.
      local name =
        if std.isObject(item)
           && '$anchor' in item
        then item['$anchor']
        else if std.isObject(item)
                && '$ref' in item
        then this.getRefName(item['$ref'])
        else '';

      // Because order may matter (for example for prefixItems), we return a list.
      parsed + { [if name != '' then '_name']:: name }
      for item in list
    ],
  // foldEnd

  resolveRef(ref, currentSchema, schemaDB):
    // foldStart
    local getFragment(baseURI, ref) =
      local split = std.splitLimit(ref, '#', 2);
      local schema = schemadb_util.get(schemaDB, baseURI + split[0]);
      if schema != {}
      then
        this.resolveRef(
          '#' + split[1],
          schemadb_util.get(schemaDB, baseURI + split[0]),
          schemaDB,
        )
      else {};

    local resolved =
      // Absolute URI
      if std.startsWith(ref, 'https://')
      then
        local baseURI = self.getURIBase(ref);
        local path = self.getURIPath(ref);
        if std.member(ref, '#')
        // Absolute URI with fragment
        then getFragment(baseURI, path)
        // Absolute URI
        else schemadb_util.get(schemaDB, baseURI + path)

      // Relative reference
      else if std.startsWith(ref, '/')
      then
        local baseURI = self.getURIBase(schemadb_util.getID(currentSchema));
        if std.member(ref, '#')
        // Relative reference with fragment
        then getFragment(baseURI, ref)
        // Relative reference
        else schemadb_util.get(schemaDB, baseURI + ref)

      // Fragment only
      else if std.startsWith(ref, '#')
      then
        local split = std.split(ref, '/')[1:];
        local find(schema, keys) =
          local key = keys[0];
          if key in schema
          then
            if std.length(keys) == 1
            then schema[key]
            else find(schema[key], keys[1:])
          else {};
        find(currentSchema, split)

      else {};
    if '$ref' in resolved
    then
      this.resolveRef(
        resolved['$ref'],
        currentSchema,
        schemaDB,
      )
    else resolved,
  // foldEnd
}

// vim: foldmethod=marker foldmarker=foldStart,foldEnd foldlevel=0
