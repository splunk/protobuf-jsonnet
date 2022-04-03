local dispatch = import 'dispatch.libsonnet';
local constraintsCheck = import 'field-constraints.libsonnet';

// a dispatch function for repeated fields.
local dispatchArray = function(inner, updateContext=true) (
  function(typeName, input, ctx) (
    local t = std.type(input);
    if t != 'array'
    then
      error '%s: want array of type %s, got %s' % [ctx, typeName, t]
    else
      std.mapWithIndex(function(i, item) inner(
        typeName,
        item,
        if updateContext then '%s[%d]' % [ctx, i] else ctx,
      ), input)
  )
);

// a dispatch function for map fields.
local dispatchMap = function(inner, updateContext=true) (
  function(typeName, input, ctx) (
    local t = std.type(input);
    if t != 'object'
    then
      error '%s: want object with values of type %s, got %s' % [ctx, typeName, t]
    else
      std.foldl(function(prev, name) prev { [name]: inner(
                  typeName,
                  input[name],
                  if updateContext then '%s.%s' % [ctx, name] else ctx,
                ) },
                std.objectFields(input),
                {})
  )
);

// validation map for various container types.
local containerValidateMap = {
  '': dispatch(),
  list: dispatchArray($['']),
  map: dispatchMap($['']),
};

// normalization map for various container types.
local containerNormalizeMap = {
  '': dispatch('normalizer', false),
  list: dispatchArray($[''], false),
  map: dispatchMap($[''], false),
};

local generator = function(type, fields0, oneOfs) (
  // normalize metadata by adding missing fields with default values
  local addOptionalFields = function(meta) (
    local x1 = if std.objectHas(meta, 'required') then meta else meta { required: false };
    local x2 = if std.objectHas(x1, 'containerType') then x1 else x1 { containerType: '' };
    local x3 = if std.objectHas(x2, 'constraints') then x2 else x2 { constraints: {} };
    x3
  );
  // create the fields map from the one passed in, ensuring that all meta objects have the standard set of expected fields.
  local fields = std.foldl(function(prev, key) prev { [key]: addOptionalFields(fields0[key]) }, std.objectFields(fields0), {});

  // make a map of metadata keyed by all field names including canonical names and JSON aliases
  local allFields = std.foldl(
    function(prev, name) (
      local meta = fields[name];
      std.foldl(function(prev2, allowedName) prev2 { [allowedName]: meta }, meta.allowedNames, prev)
    ),
    std.objectFields(fields),
    {}
  );

  // utility functions

  // subset of names that are set on the object
  local fieldsSet = function(object, names) (
    std.foldl(function(prev, name) if std.objectHas(object, name) then prev + [name] else prev, names, [])
  );

  // checks that no more than one of the names is set for the object and at least one is set if the required flag is set
  local checkOneOf = function(input, ctx, group, names, required=false) (
    local setNames = fieldsSet(input, names);
    if std.length(setNames) > 1 then (
      error '%s (group: %s) - fields %s cannot be set at the same time' % [ctx, group, std.toString(setNames)]
    ) else (
      if required && std.length(setNames) == 0 then (
        if std.length(names) > 1 then
          error '%s (group: %s) - at least one field of %s must be set' % [ctx, group, std.toString(names)]
        else
          error '%s - field "%s" must be set ' % [ctx, names[0]]
      )
      else input
    )
  );

  // check function to ensure that only known field names are set in the object.
  local checkValidFields = function(userInput, ctx) (
    local badFields = std.foldl(
      function(prev, name) if std.objectHas(allFields, name) then prev else prev + [name],
      std.objectFields(userInput),
      []
    );
    if std.length(badFields) > 0
    then
      error '%s: invalid field(s) %s found' % [ctx, std.toString(badFields)]
    else
      userInput
  );

  // apply a check function that accepts the field name over all declared fields
  local applyChecksOverFields = function(input, check) std.foldl(function(prev, name) check(prev, name), std.objectFields(fields), input);

  // check function to check that the same field is not used twice via JSON aliases
  local checkAliases = function(userInput, ctx) (
    local checker = function(input, name) (
      local meta = fields[name];
      if std.length(meta.allowedNames) == 1 then input else checkOneOf(input, ctx, 'alias', meta.allowedNames)
    );
    applyChecksOverFields(userInput, checker)
  );

  // check function to check required fields.
  local checkRequiredFields = function(userInput, ctx) (
    local checker = function(input, name) (
      local meta = fields[name];
      if !meta.required then input else checkOneOf(input, ctx, 'alias', meta.allowedNames, true)
    );
    applyChecksOverFields(userInput, checker)
  );

  // checks a single field for type and constraints correctness if it exists in the object.
  // Either canonical or aliased names can be specified.
  local checkField = function(input, name, ctx) (
    local meta = allFields[name];
    local fn = containerValidateMap[meta.containerType];
    if !std.objectHas(input, name)
    then input
    else (
      local innerCtx = '%s.%s' % [ctx, name];
      local val0 = fn(meta.type, input[name], innerCtx);
      local val1 = constraintsCheck(meta, val0, innerCtx);
      input { [name]: val1 }
    )
  );

  // check function to check field values against their type, for all fields that are set on the object.
  local checkFields = function(userInput, ctx) (
    std.foldl(function(prev, name) checkField(prev, name, ctx), std.objectFields(userInput), userInput)
  );

  // expanded a list of canonical field names to include both canonical and JSON field names in the output
  local expandFieldNames(flds) = std.flatMap(function(name) fields[name].allowedNames, flds);

  // check function to check that only one of the fields in the one of groups is set
  local checkOneOfs = function(userInput, ctx) (
    local oneOfCheck = function(input, oneOf) checkOneOf(input, ctx, oneOf.group, expandFieldNames(oneOf.fields));
    std.foldl(function(prev, oneOf) oneOfCheck(prev, oneOf), oneOfs, userInput)
  );

  // check function to check that required one ofs have been set
  local checkRequiredOneOfs = function(userInput, ctx) (
    local oneOfCheck = function(input, oneOf) if oneOf.required then checkOneOf(input, ctx, oneOf.group, expandFieldNames(oneOf.fields), true) else input;
    std.foldl(function(prev, oneOf) oneOfCheck(prev, oneOf), oneOfs, userInput)
  );

  // compose an array of checks to make it look like one check.
  local compositeChecks = function(checks) (
    function(userInput, ctx) (
      std.foldl(function(prev, check) check(prev, ctx), checks, userInput)
    )
  );

  local canonicalKeyMap = std.foldl(function(prev, key) prev { [key]: allFields[key].allowedNames[0] }, std.objectFields(allFields), {});
  local jsonKeyMap = std.foldl(function(prev, key) prev { [key]: allFields[key].allowedNames[std.length(allFields[key].allowedNames) - 1] }, std.objectFields(allFields), {});

  {
    validateAll: function(input0, ctx='') (
      local context = if ctx == '' then type else ctx;
      local input = if std.type(input0) == 'object' then input0 else error '%s: want object, found %s' % [context, std.type(input0)];
      local checker = compositeChecks([
        checkValidFields,
        checkAliases,
        checkRequiredFields,
        checkFields,
        checkOneOfs,
        checkRequiredOneOfs,
      ]);
      checker(input, context)
    ),
    validatePartial: function(input0, ctx='') (
      local context = if ctx == '' then type else ctx;
      local input = if std.type(input0) == 'object' then input0 else error '%s: want object, found %s' % [context, std.type(input0)];
      local checker = compositeChecks([
        checkValidFields,
        checkAliases,
        checkFields,
        checkOneOfs,
      ]);
      checker(input, context)
    ),
    validateField: function(input0, name, ctx='') (
      local input = if std.type(input0) == 'object' then input0 else error '%s: want object, found %s' % [ctx, std.type(input0)];
      local checker = compositeChecks([
        checkAliases,
        function(input, ctx) checkField(input, name, ctx),
        checkOneOfs,
      ]);
      checker(input, ctx)
    ),
    normalizeAll: function(input, kind='') (
      local keyMap = if kind == 'json' then jsonKeyMap else canonicalKeyMap;
      std.foldl(function(prev, key) (
        if !std.objectHas(allFields, key)
        then prev { [key]: input[key] }
        else (
          local meta = allFields[key];
          local normalizer = containerNormalizeMap[meta.containerType];
          local nKey = keyMap[key];
          prev { [nKey]: normalizer(meta.type, input[key], kind) }
        )
      ), std.objectFields(input), {})
    ),
  }
);

generator
