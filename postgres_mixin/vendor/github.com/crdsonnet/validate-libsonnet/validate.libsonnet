{
  local root = self,

  validate(object, schema)::
    if std.isBoolean(schema)
    then schema
    else if schema == {}
    then true
    else root.process(
      object,
      schema,
      root.genericTestCases
    ),

  process(object, schema, testcases):
    std.all([
      testcases[keyword](object, schema)
      for keyword in std.objectFields(testcases)
      if keyword in schema
    ]),

  types: {
    boolean(object, schema):
      std.isBoolean(object),

    'null'(object, schema):
      std.type(object) == 'null',

    string(object, schema)::
      // foldStart
      if !std.isString(object)
      then false
      else root.process(
        object,
        schema,
        root.typeTestCases.string
      ),
    // foldEnd

    number(object, schema)::
      // foldStart
      if !std.isNumber(object)
      then false
      else root.process(
        object,
        schema,
        root.typeTestCases.number
      ),
    // foldEnd

    integer(object, schema)::
      // foldStart
      if !std.isNumber(object)
         && std.mod(object, 1) != 0
      then false
      else root.process(
        object,
        schema,
        root.typeTestCases.number
      ),
    // foldEnd

    object(object, schema)::
      // foldStart
      if !std.isObject(object)
      then false
      else root.process(
        object,
        schema,
        root.typeTestCases.object
      ),
    // foldEnd

    array(object, schema)::
      // foldStart
      if !std.isArray(object)
      then false
      else root.process(
        object,
        schema,
        root.typeTestCases.array
      ),
    // foldEnd
  },

  notImplemented(key, schema):
    std.trace('JSON Schema attribute `%s` not implemented.' % key, true),

  genericTestCases: {
    enum(object, schema):
      std.member(schema.enum, object),

    const(object, schema):
      object == schema.const,

    not(object, schema):
      !root.validate(object, schema.not),

    allOf(object, schema):
      // foldStart
      std.all([
        root.validate(object, s)
        for s in schema.allOf
      ]),
    // foldEnd

    anyOf(object, schema):
      // foldStart
      std.any([
        root.validate(object, s)
        for s in schema.anyOf
      ]),
    // foldEnd

    oneOf(object, schema):
      // foldStart
      std.length([
        true
        for s in schema.oneOf
        if root.validate(object, s)
      ]) == 1,
    // foldEnd

    'if'(object, schema):
      // foldStart
      if root.validate(
        object,
        std.mergePatch(
          schema { 'if': true, 'then': true },
          schema['if']
        )
      )
      then
        if 'then' in schema
        then
          root.validate(
            object,
            std.mergePatch(
              schema { 'if': true, 'then': true },
              schema['then']
            )
          )
        else true
      else
        if 'else' in schema
        then
          root.validate(
            object,
            std.mergePatch(
              schema { 'if': true, 'then': true },
              schema['else']
            )
          )
        else true,
    // foldEnd

    type(object, schema):
      // foldStart
      if std.isBoolean(schema.type)
      then object != null

      else if std.isArray(schema.type)
      then std.any([
        root.types[t](object, schema)
        for t in schema.type
      ])

      else root.types[schema.type](object, schema),
    // foldEnd
  },

  typeTestCases: {
    string: {
      // foldStart
      minLength(object, schema):
        std.length(object) >= schema.minLength,

      maxLength(object, schema):
        std.length(object) <= schema.maxLength,

      pattern(object, schema):
        root.notImplemented('pattern', schema),

      // vocabulary specific
      //format(object, schema):
      //  root.notImplemented('format', schema),
    },  // foldEnd

    number: {
      // foldStart
      multipleOf(object, schema): std.mod(object, schema.multipleOf) == 0,
      minimum(object, schema): object >= schema.minimum,
      maximum(object, schema): object <= schema.maximum,

      exclusiveMinimum(object, schema):
        if std.isBoolean(schema.exclusiveMinimum)  // Draft 4
        then
          if 'minimum' in schema
          then
            if schema.exclusiveMinimum
            then object > schema.minimum
            else object >= schema.minimum
          else true  // invalid schema doesn't mean invalid object
        else object > schema.exclusiveMinimum,

      exclusiveMaximum(object, schema):
        if std.isBoolean(schema.exclusiveMaximum)  // Draft 4
        then
          if 'maximum' in schema
          then
            if schema.exclusiveMaximum
            then object > schema.maximum
            else object >= schema.maximum
          else true  // invalid schema doesn't mean invalid object
        else object > schema.exclusiveMaximum,
    },  // foldEnd

    object: {
      // foldStart
      patternProperties(object, schema):
        root.notImplemented('patternProperties', schema),
      dependentRequired(object, schema):
        root.notImplemented('dependentRequired', schema),
      unevaluatedProperties(object, schema):
        root.notImplemented('unevaluatedProperties', schema),
      additionalProperties(object, schema):
        root.notImplemented('additionalProperties', schema),

      properties(object, schema):
        std.all([
          root.validate(object[property], schema.properties[property])
          for property in std.objectFields(schema.properties)
          if property in object
        ]),

      required(object, schema):
        std.all([
          std.member(std.objectFields(object), property)
          for property in schema.required
        ]),

      propertyNames(object, schema):
        std.all([
          self.string(property, schema.propertyNames)
          for property in std.objectFields(schema)
        ]),

      minProperties(object, schema):
        std.count(std.objectFields(object)) >= schema.minProperties,

      maxProperties(object, schema):
        std.count(std.objectFields(object)) <= schema.maxProperties,

    },  // foldEnd

    array: {
      // foldStart
      minItems(object, schema):
        std.length(object) >= schema.minItems,

      maxItems(object, schema):
        std.length(object) <= schema.maxItems,

      uniqueItems(object, schema):
        local f = function(x) std.md5(std.manifestJson(x));
        if schema.uniqueItems
        then std.set(object, f) == std.sort(object, f)
        else true,

      prefixItems(object, schema):
        if std.length(schema.prefixItems) > 0
        then
          local lengthCheck =
            if 'items' in schema
               && std.isBoolean(schema.items)
               && !schema.items
            then std.length(object) == std.length(schema.prefixItems)
            else std.length(object) >= std.length(schema.prefixItems);

          if !lengthCheck
          then false
          else
            std.all([
              root.validate(object[i], schema.prefixItems[i])
              for i in std.range(0, std.length(schema.prefixItems) - 1)
            ])
        else true,

      items(object, schema):
        if std.isBoolean(schema.items)
        then true  // only valid in the context of prefixItems
        else
          if std.length(object) == 0
          then true  // validated by prefixItems and min/maxLength
          else
            local count =
              if 'prefixItems' in schema
              then std.length(schema.prefixItems)
              else 0;
            std.all([
              root.validate(item, schema.items)
              for item in object[count:]
            ]),

      contains(object, schema):
        local validated = [
          true
          for item in object
          if root.validate(item, schema.contains)
        ];
        std.any(validated)
        && std.all([
          if 'minContains' in schema
          then std.length(validated) >= schema.minContains
          else true,
          if 'maxContains' in schema
          then std.length(validated) <= schema.maxContains
          else true,
        ]),

      unevaluatedItems(object, schema):
        root.notImplemented('unevaluatedItems', schema),
    },  // foldEnd
  },
}

// vim: foldmethod=marker foldmarker=foldStart,foldEnd foldlevel=0
