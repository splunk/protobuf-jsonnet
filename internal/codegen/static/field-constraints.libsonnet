local valOrDefault = function(obj, name, def={}) if std.objectHas(obj, name) then obj[name] else def;

local friendlyTypes = {
  'google.protobuf.StringValue': 'string',
};

local friendlyTypeName = function(meta) if std.objectHas(friendlyTypes, meta.type) then friendlyTypes[meta.type] else meta.type;

local getValue = function(input) if std.type(input) == 'object' && std.objectHas(input, 'value') then input.value else input;

local identity = function(meta, input, ctx) input;
local inputIdentity = function(input) function(meta, val, ctx) input;

// string constraints
local constCheck = function(typeMeta, input, ctx) (
  if !std.objectHas(typeMeta.constraints, 'const') then input else (
    local constValue = typeMeta.constraints.const;
    if input != constValue
    then
      error '%s: const %s value: want "%s", got "%s"' % [ctx, friendlyTypeName(typeMeta), constValue, std.toString(input)]
    else
      input
  )
);

local inCheck = function(typeMeta, input, ctx) (
  if !std.objectHas(typeMeta.constraints, 'in') then input else (
    local inValues = typeMeta.constraints['in'];
    if !std.member(inValues, input) then
      error '%s: %s in value: want one of %s, got "%s"' % [ctx, friendlyTypeName(typeMeta), std.toString(inValues), std.toString(input)]
    else
      input
  )
);

local notInCheck = function(typeMeta, input, ctx) (
  if !std.objectHas(typeMeta.constraints, 'not_in') then input else (
    local notInValues = typeMeta.constraints.not_in;
    if std.member(notInValues, input) then
      error '%s: %s not_in value: want none of %s, got "%s"' % [ctx, friendlyTypeName(typeMeta), std.toString(notInValues), std.toString(input)]
    else
      input
  )
);

local validateString = function(meta, input, ctx) (
  if !std.objectHas(meta.constraints, 'String_') then input else (
    local typeMeta = { type: meta.type, constraints: meta.constraints.String_ };
    local val = getValue(input);
    local checkers = [
      constCheck,
      inCheck,
      notInCheck,
      inputIdentity(input),
    ];
    std.foldl(function(prev, check) check(typeMeta, prev, ctx), checkers, val)
  )
);

// dispatchers
local dispatchTable = {
  string: validateString,
  'google.protobuf.StringValue': validateString,
};

local dispatchScalar = function(meta, input, ctx) (
  local fn = valOrDefault(dispatchTable, meta.type, identity);
  fn(meta, input, ctx)
);

local dispatchList = function(meta, input, ctx) (
  local constraints = meta.constraints;
  local itemsConstraints = valOrDefault(constraints, 'items');
  local typeConstraints = valOrDefault(itemsConstraints, 'Type');
  std.mapWithIndex(function(i, item) dispatchScalar({ type: meta.type, constraints: typeConstraints }, item, '%s[%d]' % [ctx, i]), input)
);

local dispatchMap = function(meta, input, ctx) (
  local constraints = meta.constraints;
  local itemsConstraints = valOrDefault(constraints, 'values');
  local typeConstraints = valOrDefault(itemsConstraints, 'Type');
  std.foldl(function(prev, name) prev { [name]: dispatchScalar({ type: meta.type, constraints: typeConstraints }, input[name], '%s.%s' % [ctx, name]) },
            std.objectFields(input),
            {})
);

local dispatchTable = {
  '': dispatchScalar,
  list: dispatchList,
  map: dispatchMap,
};

function(field, input, ctx='') (
  // extract only the portions of field meta that we should use. `meta` references in other parts of the code
  // refer to this object.
  local meta = {
    type: field.type,
    constraints: valOrDefault(field, 'constraints'),
  };
  dispatchTable[field.containerType](meta, input, ctx)
)
