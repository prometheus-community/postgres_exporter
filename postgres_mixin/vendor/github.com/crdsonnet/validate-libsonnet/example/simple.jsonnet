local validate = import 'validate-libsonnet/main.libsonnet';

local func(arg) =
  assert validate.checkParameters({
    arg: std.isString(arg),
  });
  {/* do something here */ };

func('this is a string')
