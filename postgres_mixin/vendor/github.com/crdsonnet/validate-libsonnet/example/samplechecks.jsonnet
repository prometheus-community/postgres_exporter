// Examples checks:

local customCheck(value) =
  std.member(['a', 'b'], value);

local stringMaxLengthCheck(value) =
  local schema = { type: 'string', maxLength: 1 };
  params.schemaCheck(value, schema);

