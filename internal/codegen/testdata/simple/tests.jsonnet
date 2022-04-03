local bytesString = std.base64('hello');
local allFields = [
  'bool_field',
  'bytes_field',
  'double_field',
  'enum_field',
  'fixed32_field',
  'fixed64_field',
  'float_field',
  'inner1',
  'inner2',
  'int32_field',
  'int64_field',
  'sfixed32_field',
  'sfixed64_field',
  'sint32_field',
  'sint64_field',
  'str_field',
  'uint32_field',
  'uint64_field',
];

local ensureArray = function(valOrArray) if std.type(valOrArray) == 'array' then valOrArray else [valOrArray];

local inSet = function(valOrArray) function(fld) std.setMember(fld, std.set(ensureArray(valOrArray)));
local notInSet = function(valOrArray) function(fld) !std.setMember(fld, std.set(ensureArray(valOrArray)));

local basicTests = [
  {
    name: 'setters',
    summary: 'ensure setters are generated as expected and a good object can be produced',
    file: 'setters.jsonnet',
    result: {
      bool_field: true,
      bytes_field: bytesString,
      double_field: 12.2,
      enum_field: 'FIRST',
      fixed32_field: 7,
      fixed64_field: 8,
      inner1: {
        numbers: ['ONE', 'TWO'],
      },
      inner2: {
        msgs: {
          m1: { numbers: ['TWO'] },
        },
      },
      int32_field: 1,
      int64_field: 2,
      sfixed32_field: 9,
      sfixed64_field: 10,
      sint32_field: 3,
      sint64_field: 4,
      str_field: 'string',
      uint32_field: 5,
      uint64_field: 6,
    },
  },
  {
    name: 'validate-only',
    summary: 'populate with existing object, pass validation',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      bool_field: true,
      bytes_field: bytesString,
      enum_field: 'FIRST',
      fixed32_field: 7,
      fixed64_field: 8,
      float_field: 11.1,
      inner1: {
        numbers: ['ONE', 'TWO'],
      },
      inner2: {
        msgs: {
          m1: { numbers: ['TWO'] },
        },
      },
      int32_field: 1,
      int64_field: 2,
      sfixed32_field: 9,
      sfixed64_field: 10,
      sint32_field: 3,
      sint64_field: 4,
      str_field: 'string',
      uint32_field: 5,
      uint64_field: 6,
    },
  },
  {
    name: 'numbers_as_strings',
    summary: 'ensure that numeric types can be set as strings',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      int32_field: '1',
      int64_field: '2',
      sfixed32_field: '9',
      sfixed64_field: '10',
      sint32_field: '3',
      sint64_field: '4',
      uint32_field: '5',
      uint64_field: '6',
    },
  },
  {
    name: 'allow-aliases',
    summary: 'populate object with alias fields, ensure validation ok',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new(%s)._validate()
    ||| % std.manifestJsonEx($.result, '  '),
    result: {
      boolField: true,
      bytesField: bytesString,
      enumField: 'FIRST',
      fixed32Field: 7,
      fixed64Field: 8,
      floatField: 11.1,
      int32Field: 1,
      int64Field: 2,
      sfixed32Field: 9,
      sfixed64Field: 10,
      sint32Field: 3,
      sint64Field: 4,
      strField: 'string',
      uint32Field: 5,
      uint64Field: 6,
    },
  },
  {
    name: 'partial-fields',
    summary: 'ensure valid partial object allowed',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({int32_field: 200 })
    |||,
    result: {
      int32_field: 200,
    },
  },
  {
    name: 'enum_by_value',
    summary: 'ensure that enums can be set by numeric value, rather than name',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({enum_field: 1 })
    |||,
    result: {
      enum_field: 1,
    },
  },
  {
    name: 'enum_by_str_value',
    summary: 'ensure that enums can be set by string value, rather than name',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({enum_field: '1' })
    |||,
    result: {
      enum_field: '1',
    },
  },
  {
    name: 'canonicalize',
    summary: 'populate object with alias fields, get canonical result',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new(%s)._validate()._normalize()
    ||| % std.manifestJsonEx($.data, '  '),
    data: {
      boolField: true,
      bytesField: bytesString,
      enumField: 'FIRST',
      fixed32Field: 7,
      fixed64Field: 8,
      floatField: 11.1,
      int32Field: 1,
      int64Field: 2,
      sfixed32Field: 9,
      sfixed64Field: 10,
      sint32Field: 3,
      sint64Field: 4,
      strField: 'string',
      uint32Field: 5,
      uint64Field: 6,
    },
    result: {
      bool_field: true,
      bytes_field: bytesString,
      enum_field: 'FIRST',
      fixed32_field: 7,
      fixed64_field: 8,
      float_field: 11.1,
      int32_field: 1,
      int64_field: 2,
      sfixed32_field: 9,
      sfixed64_field: 10,
      sint32_field: 3,
      sint64_field: 4,
      str_field: 'string',
      uint32_field: 5,
      uint64_field: 6,
    },
  },
  {
    name: 'jsonize',
    summary: 'populate object with canonical fields, get json result',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new(%s)._validate()._normalize('json')
    ||| % std.manifestJsonEx($.data, '  '),
    result: {
      boolField: true,
      bytesField: bytesString,
      enumField: 'FIRST',
      fixed32Field: 7,
      fixed64Field: 8,
      floatField: 11.1,
      int32Field: 1,
      int64Field: 2,
      sfixed32Field: 9,
      sfixed64Field: 10,
      sint32Field: 3,
      sint64Field: 4,
      strField: 'string',
      uint32Field: 5,
      uint64Field: 6,
    },
    data: {
      bool_field: true,
      bytes_field: bytesString,
      enum_field: 'FIRST',
      fixed32_field: 7,
      fixed64_field: 8,
      float_field: 11.1,
      int32_field: 1,
      int64_field: 2,
      sfixed32_field: 9,
      sfixed64_field: 10,
      sint32_field: 3,
      sint64_field: 4,
      str_field: 'string',
      uint32_field: 5,
      uint64_field: 6,
    },
  },
];

local generateNonBoolsToBool = function() (
  std.filterMap(
    notInSet(['bool_field', 'inner1', 'inner2', 'enum_field']),
    function(fld) {
      name: 'neg_bool_%s' % fld,
      summary: 'ensure %s field cannot be set to a bool value' % fld,
      code: |||
        local types = import 'types.libsonnet';
        types.testdata.simple.TopMessage._new({ %s: true })._validate()
      ||| % fld,
      err: 'RUNTIME ERROR: testdata.simple.TopMessage.%s: invalid input true' % fld,
    },
    allFields,
  ) + [
    {
      name: 'neg_bool_message',
      summary: 'ensure message fields cannot be set to a bool value',
      code: |||
        local types = import 'types.libsonnet';
        types.testdata.simple.TopMessage._new({ inner1 : true })
      |||,
      err: 'RUNTIME ERROR: testdata.simple.TopMessage.inner1: want object, found boolean',
    },
    {
      name: 'neg_bool_enum',
      summary: 'ensure enum fields cannot be set to a bool value',
      code: |||
        local types = import 'types.libsonnet';
        types.testdata.simple.TopMessage._new({ enum_field : true })
      |||,
      err: 'RUNTIME ERROR: testdata.simple.TopMessage.enum_field: invalid value true for enum testdata.simple.TopLevelEnum',
    },
  ]
);

local generateNonNumbersToNumber = function() (
  std.filterMap(
    inSet(['bool_field', 'bytes_field', 'str_field']),
    function(fld) {
      name: 'neg_number_%s' % fld,
      summary: 'ensure %s field cannot be set to a numeric value' % fld,
      code: |||
        local types = import 'types.libsonnet';
        types.testdata.simple.TopMessage._new({ %s: 1 })
      ||| % fld,
      err: 'RUNTIME ERROR: testdata.simple.TopMessage.%s: invalid input 1 (type=' % fld,
    },
    allFields
  ) + [
    {
      name: 'neg_number_message',
      summary: 'ensure message fields cannot be set to a bool value',
      code: |||
        local types = import 'types.libsonnet';
        types.testdata.simple.TopMessage._new({ inner1 : 1 })
      |||,
      err: 'RUNTIME ERROR: testdata.simple.TopMessage.inner1: want object, found number',
    },
  ]
);

local specificNegativeTests = [
  {
    name: 'neg_bad_field',
    summary: 'ensure that a bad field name cannot be set',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ foo: 'bar' })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage: invalid field(s) ["foo"] found',
  },
  {
    name: 'neg_disallow_aliases',
    summary: 'ensure that alias names of a field cannot be set at the same time',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ int32_field: 1, int32Field: 2 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage (group: alias) - fields ["int32_field", "int32Field"] cannot be set at the same time',
  },
  {
    name: 'neg_disallow_multi_oneofs',
    summary: 'ensure that two fields in a one of group cannot be set',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ float_field: 1, double_field: 2 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage (group: one_double_only) - fields ["float_field", "double_field"] cannot be set at the same time',
  },
  {
    name: 'neg_disallow_multi_oneofs_via_aliases1',
    summary: 'ensure that two fields in a one of group cannot be set when using aliases',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ floatField: 1, doubleField: 2 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage (group: one_double_only) - fields ["floatField", "doubleField"] cannot be set at the same time',
  },
  {
    name: 'neg_disallow_multi_oneofs_via_aliases2',
    summary: 'ensure that two fields in a one of group cannot be set when using aliases and canonical names mixed',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ floatField: 1, double_field: 2 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage (group: one_double_only) - fields ["floatField", "double_field"] cannot be set at the same time',
  },
  {
    name: 'array_for_repeated',
    summary: 'ensure that a repeated field must be an array',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ inner1: { numbers: 'ONE' } })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.inner1.numbers: want array of type testdata.simple.TopMessage.InnerEnum, got string',
  },
  {
    name: 'bad_enum1',
    summary: 'ensure that an unknown value cannot be set on an enum by name',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ enum_field: 'GARBAGE' })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.enum_field: invalid value GARBAGE for enum testdata.simple.TopLevelEnum',
  },
  {
    name: 'bad_enum2',
    summary: 'ensure that an unknown value cannot be set on an enum by value',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ enum_field: 102 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.enum_field: invalid value 102 for enum testdata.simple.TopLevelEnum',
  },
  {
    name: 'bad_enum3',
    summary: 'ensure that an object cannot be used where an enum is expected',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ enum_field: { foo: 'bar' } })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.enum_field: invalid value {"foo": "bar"} for enum testdata.simple.TopLevelEnum',
  },
  {
    name: 'bad_type_for_message',
    summary: 'ensure that an invalid inner message cannot be used and that the error message has a good context',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ inner1: { foo: 'bar'} })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.inner1: invalid field(s) ["foo"] found',
  },
  {
    name: 'bad_type_for_array_message',
    summary: 'ensure that an invalid inner message in a repeated field cannot be used and that the error message has a good context',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ inner1: { numbers: ['FOO'] } })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.inner1.numbers[0]: invalid value FOO for enum testdata.simple.TopMessage.InnerEnum',
  },
  {
    name: 'array_for_map',
    summary: 'ensure an array cannot be used where a map is expected',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage.InnerMessage2._new({ simple_map: [] })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.InnerMessage2.simple_map: want object with values of type string, got array',
  },
  {
    name: 'bad_type_for_map',
    summary: 'ensure map values are validated for correct type',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage.InnerMessage2._new({ simple_map: { foo: 1 } })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.InnerMessage2.simple_map.foo: invalid input 1 (type=number) for type string',
  },
  {
    name: 'bad_type_for_map2',
    summary: 'ensure map values are validated for correct type when values are messages',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage.InnerMessage2._new({ msgs: { foo: { bad: '1'} } })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.InnerMessage2.msgs.foo: invalid field(s) ["bad"] found',
  },
  {
    name: 'negative_uint32',
    summary: 'ensure negative values for unsigned ints are rejected',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ uint32_field: -1 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.uint32_field: bad value -1 (type uint32, less that implicit min 0)',
  },
  {
    name: 'negative_uint32_as_string',
    summary: 'ensure negative values for unsigned ints are rejected even when expressed as a string',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ uint32_field: '-1' })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.uint32_field: bad value -1 (type uint32, less that implicit min 0)',
  },
  {
    name: 'int32_value_too_high',
    summary: 'ensure values that would overflow are caught',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.simple.TopMessage._new({ int32_field: 2147483649 })
    |||,
    err: 'RUNTIME ERROR: testdata.simple.TopMessage.int32_field: bad value 2147483649 (type int32, greater that implicit max -2147483648)',
  },
];


basicTests + generateNonBoolsToBool() + generateNonNumbersToNumber() + specificNegativeTests
