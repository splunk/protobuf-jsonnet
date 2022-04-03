local bytesString = std.base64('hello');
local basicTests = [
  {
    name: 'wrapped_as_primitive',
    summary: 'ensure that standard protobuf message wrappers can be set with simplified syntax',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      str_field: 'foo',
      bytes_field: bytesString,
      int32_field: 1,
      int64_field: 2,
      uint32_field: 3,
      uint64_field: 4,
      float_field: 5.1,
      double_field: 6.2,
      bool_field: true,
      duration_field: '0.25s',
      any_field: { foo: 'bar' },
      struct_field: { bar: 'baz' },
    },
  },
  {
    name: 'wrapped_as_messages',
    summary: 'ensure that standard protobuf message wrappers can be set with message syntax',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      str_field: { value: 'foo' },
      bytes_field: { value: bytesString },
      int32_field: { value: 1 },
      int64_field: { value: 2 },
      uint32_field: { value: 3 },
      uint64_field: { value: 4 },
      float_field: { value: 5.1 },
      double_field: { value: 6.2 },
      bool_field: { value: true },
      duration_field: { seconds: 1, nanos: 25000000 },
    },
  },
  {
    name: 'numbers_as_strings',
    summary: 'ensure that standard protobuf message wrappers can be set with numbers as strings',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      int32_field: '1',
      int64_field: '2',
      uint32_field: '3',
      uint64_field: '4',
      float_field: '5.1',
      double_field: '6.2',
    },
  }
  {
    name: 'any_validation',
    summary: 'ensure that any messages with @type attributes are validated',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({
        any_field: {
            '@type': 'namespace/testdata.wellknown.TopMessage.Config',
            name: 'foo',
            val: 'bar',
        },
      })
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.any_field(type:testdata.wellknown.TopMessage.Config): invalid field(s) ["val"] found',
  },
  {
    name: 'any_validation_unknown_type',
    summary: 'ensure that any messages with an unknown @type attribute validates',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new(%s)
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      any_field: {
        '@type': 'namespace/testdata.wellknown.TopMessage.XXX',
        name: 'foo',
        val: 'bar',
      },
    },
  },
  {
    name: 'any_validation_no_namespace',
    summary: 'ensure that any messages with an a non-namespaced @name skips validation',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new(%s)
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      any_field: {
        '@type': 'testdata.wellknown.TopMessage.Config',
        name: 'foo',
        val: 'bar',
      },
    },
  },
];

local negativeTests = [
  {
    name: 'neg_bool_as_string',
    summary: 'ensure that an any field cannot be a scalar value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({bool_field: 'true'})
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.bool_field: invalid input true (type=string) for type google.protobuf.BoolValue',
  },
  {
    name: 'neg_any_scalar',
    summary: 'ensure that an any field cannot be a scalar value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({any_field: 'foo'})
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.any_field: Any field was not an object, got string',
  },
  {
    name: 'neg_any_array',
    summary: 'ensure that an any field cannot be an array value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({any_field: ['foo']})
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.any_field: Any field was not an object, got array',
  },
  {
    name: 'neg_struct_scalar',
    summary: 'ensure that a struct field cannot be a scalar value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({struct_field: 'foo'})
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.struct_field: invalid input foo (type=string) for type google.protobuf.Struct',
  },
  {
    name: 'neg_struct_array',
    summary: 'ensure that a struct field cannot be an array value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({struct_field: ['foo']})
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.struct_field: invalid input ["foo"] (type=array) for type google.protobuf.Struct',
  },
  {
    name: 'neg_uint32_wrapper',
    summary: 'ensure that a uint32 wrapper cannot be negative',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({uint32_field: { value: '-1'} })
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.uint32_field: bad value -1 (type google.protobuf.UInt32Value, less that implicit min 0)',
  },
  {
    name: 'neg_uint32_wrapper_as_number',
    summary: 'ensure that a uint32 wrapper cannot be negative',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({uint32_field: -1 })
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.uint32_field: bad value -1 (type google.protobuf.UInt32Value, less that implicit min 0)',
  },
  {
    name: 'neg_int32_value_too_low',
    summary: 'ensure that a 32 bit wrapper does not underflow',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({int32_field: -2147483649 })
    |||,
    err: 'RUNTIME ERROR: testdata.wellknown.TopMessage.int32_field: bad value -2147483649 (type google.protobuf.Int32Value, less that implicit min -2147483648)',
  },
];

local badWrappersTest = std.map(function(fld) (
  {
    name: 'neg_bad_wrap_%s' % fld,
    summary: 'ensure that a %s field cannot have a bad message',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.wellknown.TopMessage._new({ %s: { value: '1', value2: 'foo' } })
    ||| % fld,
    err: if fld == 'duration_field' then
      'RUNTIME ERROR: testdata.wellknown.TopMessage.duration_field: invalid field(s) ["value", "value2"] for type google.protobuf.Duration'
    else 'RUNTIME ERROR: testdata.wellknown.TopMessage.%s: invalid input {"value": "1", "value2": "foo"} (type=object)' % fld,
  }
), [
  'str_field',
  'bytes_field',
  'int32_field',
  'int64_field',
  'uint32_field',
  'uint64_field',
  'float_field',
  'double_field',
  'bool_field',
  'duration_field',
]);

basicTests + negativeTests + badWrappersTest
