# validate-libsonnet

Type checking is a common grievance in the jsonnet eco-system, this library is an
aid to validate function parameters and other values.

Here's a comprehensive example validating the function arguments against the
arguments documented by docsonnet:

```jsonnet
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';
local validate = import 'github.com/crdsonnet/validate-libsonnet/main.libsonnet';

{
  '#func'::
    d.func.new(
      'sample function',
      args=[
        d.arg('num', d.T.number),
        d.arg('str', d.T.string),
        d.arg('enum', d.T.string, enums=['valid', 'values']),
      ],
    ),
  func(num, str, enum)::
    assert validate.checkParamsFromDocstring(
      [num, str, enum],
      self['#func'],
    );
    {/* do something here */ },

  return: self.func(100, 'this is a string', 'valid'),
}

```

A failure output would look like this:

```
TRACE: vendor/github.com/crdsonnet/validate-libsonnet/main.libsonnet:63 
Invalid parameters:
  Parameter enum is invalid:
    Value "invalid" MUST match schema:
      {
        "enum": [
          "valid",
          "values"
        ],
        "type": "string"
      }
  Parameter str is invalid:
    Value 20 MUST match schema:
      {
        "type": "string"
      }
RUNTIME ERROR: Assertion failed
	fromdocstring.jsonnet:(15:5)-(19:31)	
	fromdocstring.jsonnet:21:11-40	object <anonymous>
	Field "return"	
	During manifestation	


```


## Install

```
jb install github.com/crdsonnet/validate-libsonnet@master
```

## Usage

```jsonnet
local validate = import 'github.com/crdsonnet/validate-libsonnet/main.libsonnet'
```

## Index

* [`fn checkParameters(checks)`](#fn-checkparameters)
* [`fn checkParamsFromDocstring(params, docstring)`](#fn-checkparamsfromdocstring)
* [`fn getChecksFromDocstring(params, docstring)`](#fn-getchecksfromdocstring)
* [`fn schemaCheck(param, schema)`](#fn-schemacheck)

## Fields

### fn checkParameters

```ts
checkParameters(checks)
```

`checkParameters` validates parameters against their `checks`.

```jsonnet
local validate = import 'github.com/crdsonnet/validate-libsonnet/main.libsonnet';

local func(arg) =
  assert validate.checkParameters({
    arg: std.isString(arg),
  });
  {/* do something here */ };

func('this is a string')

```

A failure output would look like this:

```
TRACE: vendor/github.com/crdsonnet/validate-libsonnet/main.libsonnet:63 
Invalid parameters:
  Parameter enum is invalid:
    Value "invalid" MUST match schema:
      {
        "enum": [
          "valid",
          "values"
        ],
        "type": "string"
      }
  Parameter str is invalid:
    Value 20 MUST match schema:
      {
        "type": "string"
      }
RUNTIME ERROR: Assertion failed
	fromdocstring.jsonnet:(15:5)-(19:31)	
	fromdocstring.jsonnet:21:11-40	object <anonymous>
	Field "return"	
	During manifestation	


```


### fn checkParamsFromDocstring

```ts
checkParamsFromDocstring(params, docstring)
```

`checkParamsFromDocstring` validates `params` against a docsonnet `docstring` object.

```jsonnet
local d = import 'github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet';
local validate = import 'github.com/crdsonnet/validate-libsonnet/main.libsonnet';

{
  '#func'::
    d.func.new(
      'sample function',
      args=[
        d.arg('num', d.T.number),
        d.arg('str', d.T.string),
        d.arg('enum', d.T.string, enums=['valid', 'values']),
      ],
    ),
  func(num, str, enum)::
    assert validate.checkParamsFromDocstring(
      [num, str, enum],
      self['#func'],
    );
    {/* do something here */ },

  return: self.func(100, 'this is a string', 'valid'),
}

```

A failure output would look like this:

```
TRACE: vendor/github.com/crdsonnet/validate-libsonnet/main.libsonnet:63 
Invalid parameters:
  Parameter enum is invalid:
    Value "invalid" MUST match schema:
      {
        "enum": [
          "valid",
          "values"
        ],
        "type": "string"
      }
  Parameter str is invalid:
    Value 20 MUST match schema:
      {
        "type": "string"
      }
RUNTIME ERROR: Assertion failed
	fromdocstring.jsonnet:(15:5)-(19:31)	
	fromdocstring.jsonnet:21:11-40	object <anonymous>
	Field "return"	
	During manifestation	


```


### fn getChecksFromDocstring

```ts
getChecksFromDocstring(params, docstring)
```

`getChecksFromDocstring` returns checks for `params` derived from a docsonnet `docstring` object.

### fn schemaCheck

```ts
schemaCheck(param, schema)
```

`schemaCheck` validates `param` against a JSON `schema`. Note that this function does not resolve "$ref" and recursion.
