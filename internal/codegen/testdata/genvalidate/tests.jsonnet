local validInput = {
  float_field: 23.2,
  str_field: 'hello',
  int32_field: 10,
  boolField: true,
  inner: {
    name: 'foo',
  },
  str_array: ['foo', 'bar'],
  str_map: { foo: 'bar', bar: 'baz' },
};

local without = function(name) std.foldl(
  function(prev, fld) if fld == name then prev else prev { [fld]: validInput[fld] },
  std.objectFields(validInput),
  {}
);

local template = function(data) |||
  local types = import 'types.libsonnet';
  types.testdata.genvalidate.TopMessage._new(%s)._validate()
||| % std.manifestJsonEx(data, '  ');

local basicTests = [
  {
    name: 'valid_object',
    summary: 'ensure a valid object can be produced by respecting constraints',
    code: template($.result),
    result: validInput,
  },
  {
    name: 'required_one_of',
    summary: 'ensure that required one-ofs need to be set when requested',
    code: template($.data),
    data: without('float_field'),
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage (group: exactly_one_floater) - at least one field of ["float_field", "floatField", "double_field", "doubleField"] must be set',
  },
  {
    name: 'required_message',
    summary: 'ensure that required messages need to be set when requested',
    code: template($.data),
    data: without('inner'),
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage - field "inner" must be set',
  },
  {
    name: 'required_list',
    summary: 'ensure that required repeated fields need to be set when requested',
    code: template($.data),
    data: without('str_array'),
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage (group: alias) - at least one field of ["str_array", "strArray"] must be set',
  },
  {
    name: 'required_map',
    summary: 'ensure that required map fields need to be set when requested',
    code: template($.data),
    data: without('str_map'),
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage (group: alias) - at least one field of ["str_map", "strMap"] must be set',
  },
];

local requiredScalars = function() (
  std.map(function(fld) (
    {
      name: 'required_%s' % fld,
      summary: 'ensure that required %s need to be set when requested' % fld,
      code: template($.data),
      data: without(fld),
      err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage (group: alias) - at least one field of [',

    }
  ), ['str_field', 'int32_field', 'boolField'])
);

local constraintChecks = [
  {
    name: 'const_string',
    summary: 'check string constant',
    code: template($.data),
    data: validInput { foo_string: 'bar' },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.foo_string: const string value: want "foo", got "bar"',
  },
  {
    name: 'const_string_msg',
    summary: 'check string constant',
    code: template($.data),
    data: validInput { foo_string_msg: 'bar' },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.foo_string_msg: const string value: want "foo", got "bar"',
  },
  {
    name: 'in_string',
    summary: 'check string in values',
    code: template($.data),
    data: validInput { foo_or_bar_string: 'baz' },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.foo_or_bar_string: string in value: want one of ["foo", "bar"], got "baz"',
  },
  {
    name: 'in_string_msg',
    summary: 'check string in values',
    code: template($.data),
    data: validInput { foo_or_bar_string_msg: { value: 'baz' } },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.foo_or_bar_string_msg: string in value: want one of ["foo", "bar"], got "baz"',
  },
  {
    name: 'not_in_string',
    summary: 'check string not_in values',
    code: template($.data),
    data: validInput { not_foo_or_bar_string: 'foo' },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.not_foo_or_bar_string: string not_in value: want none of ["foo", "bar"], got "foo"',
  },
  {
    name: 'not_in_string_msg',
    summary: 'check string not_in values',
    code: template($.data),
    data: validInput { not_foo_or_bar_string_msg: 'bar' },
    err: 'RUNTIME ERROR: testdata.genvalidate.TopMessage.not_foo_or_bar_string_msg: string not_in value: want none of ["foo", "bar"], got "bar"',
  },
];


basicTests + requiredScalars() + constraintChecks
