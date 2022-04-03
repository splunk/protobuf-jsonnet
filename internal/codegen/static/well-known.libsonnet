local dispatch = import 'dispatch.libsonnet';
local validate = dispatch();
local normalize = dispatch('normalizer', false);
local isValue = function(input) std.type(input) == 'object' && std.objectHas(input, 'value') && std.length(input) == 1;

// turn boolean result function into a check
local check = function(t, fn) (
  function(input, ctx='') (
    if fn(input) then input else error '%s: invalid input %s (type=%s) for type %s' % [ctx, std.toString(input), std.type(input), t]
  )
);

// string-ish types
local isString = function(input) std.type(input) == 'string';
local isStringOrValue = function(input) isString(input) || (isValue(input) && isString(input.value));

local stringTable = {
  string: { validator: check('string', isString) },
  'google.protobuf.StringValue': { validator: check('google.protobuf.StringValue', isStringOrValue) },
  bytes: { validator: check('bytes', isString) },
  'google.protobuf.BytesValue': { validator: check('google.protobuf.BytesValue', isStringOrValue) },
};

// integer types
local min32 = -2147483648;
local max32 = 2147483648;
local min64 = -9223372036854775808;
local max64 = 9223372036854775808;

local wellKnownInts = {
  int32: {
    min: min32,
    max: max32,
    wrapper: false,
  },
  'google.protobuf.Int32Value': $.int32 { wrapper: true },
  sint32: $.int32,
  fixed32: $.int32,
  sfixed32: $.int32,

  int64: $.int32 {
    min: min64,
    max: max64,
  },
  'google.protobuf.Int64Value': $.int64 { wrapper: true },
  sint64: $.int64,
  fixed64: $.int64,
  sfixed64: $.int64,

  uint32: $.int32 { min: 0 },
  'google.protobuf.UInt32Value': $.uint32 { wrapper: true },

  uint64: $.int64 { min: 0 },
  'google.protobuf.UInt64Value': $.uint64 { wrapper: true },
};

local validateInteger0 = function(type, input, ctx) (
  local meta = wellKnownInts[type];
  local v0 = if meta.wrapper && isValue(input) then validateInteger0(type, input.value, ctx) else input;
  local v1 = if std.type(v0) == 'string' then std.parseInt(v0) else v0;
  local v2 = if std.type(v1) != 'number'
  then
    error '%s: invalid input %s (type=%s)' % [ctx, std.toString(v1), std.type(v1)]
  else (
    local v3 = if v1 < meta.min
    then
      error '%s: bad value %d (type %s, less that implicit min %d)' % [ctx, v1, type, meta.min]
    else
      v1;
    local v4 = if v3 > meta.max
    then
      error '%s: bad value %d (type %s, greater that implicit max %d)' % [ctx, v3, type, meta.min]
    else
      v3;
    v4
  );
  v2
);

local validateInteger = function(type, input, ctx) std.foldl(
  function(prev, fn) fn(type, prev, ctx),
  [
    validateInteger0,
    function(type, prev, ctx) input,  // restore uder input that may have changed with validateInteger0
  ],
  input
);

local intTable = std.foldl(function(prev, type) prev {
  [type]: { validator: function(input, ctx) validateInteger(type, input, ctx) },
}, std.objectFields(wellKnownInts), {});

// floating point
local isNumber = function(input) std.type(input) == 'number' || isString(input);  // JSON spec allows string
local isNumberOrValue = function(input) isNumber(input) || (isValue(input) && isNumber(input.value));

local floatTable = {
  double: { validator: check('double', isNumber) },
  float: { validator: check('float', isNumber) },
  'google.protobuf.FloatValue': { validator: check('google.protobuf.FloatValue', isNumberOrValue) },
  'google.protobuf.DoubleValue': { validator: check('google.protobuf.DoubleValue', isNumberOrValue) },
};

// bool
local isBool = function(input) std.type(input) == 'boolean';
local isBoolOrValue = function(input) isBool(input) || (isValue(input) && isBool(input.value));

local boolTable = {
  bool: { validator: check('bool', isBool) },
  'google.protobuf.BoolValue': { validator: check('google.protobuf.BoolValue', isBoolOrValue) },
};

// Any
local withoutAtType = function(object) (
  local keys = std.objectFields(object);
  std.foldl(function(prev, key) if key == '@type' then prev else prev { [key]: object[key] }, keys, {})
);

local processAny = function(fn, updateContext=true) function(input, ctx='') (
  local obj0 = if std.type(input) == 'object' then input else error '%s: Any field was not an object, got %s' % [ctx, std.type(input)];
  if !std.objectHas(obj0, '@type') then obj0 else (
    local atType = obj0['@type'];
    if std.type(atType) != 'string' then error '%s: Any @type attribute: want string, got %s' % [ctx, std.type(atType)]
    else (
      local typeSplit = std.splitLimit(atType, '/', 2);
      if std.length(typeSplit) != 2 then std.trace('WARN: %s: not processing unexpected @type %s' % [ctx, atType], obj0)
      else (
        local typeName = typeSplit[1];
        local validated = fn(
          typeName,
          withoutAtType(obj0),
          if updateContext then '%s(type:%s)' % [ctx, typeName] else ctx,
        );
        validated { '@type': atType }  // restore the atType
      )
    )
  )
);

local validateAny = processAny(validate);
local normalizeAny = processAny(normalize, false);

// duration
local durationValueValidator = function(input, ctx='') (
  local fieldValidators = [
    function(input) (
      if std.objectHas(input, 'seconds')
      then input { seconds: validate('int64', input.seconds, ctx + '.seconds') }
      else input
    ),
    function(input) (
      if std.objectHas(input, 'nanos')
      then input { nanos: validate('int32', input.nanos, ctx + '.nanos') }
      else input
    ),
    function(input) (
      local bad = std.filter(function(k) k != 'seconds' && k != 'nanos', std.objectFields(input));
      if std.length(bad) > 0 then
        error '%s: invalid field(s) %s for type google.protobuf.Duration' % [ctx, std.toString(bad)]
      else
        input
    ),
  ];
  local fieldValidationOutput = std.foldl(function(prev, fn) fn(prev), fieldValidators, input);
  fieldValidationOutput
);

local validateDuration = function(input, ctx='') (
  if std.type(input) == 'string'
  then
    input
  else
    durationValueValidator(input, ctx)
);

stringTable +
intTable +
floatTable +
boolTable +
{
  'google.protobuf.Struct': { validator: check('google.protobuf.Struct', function(input) std.type(input) == 'object') },
  'google.protobuf.Any': { validator: validateAny, normalizer: normalizeAny },
  'google.protobuf.Duration': { validator: validateDuration },
}
