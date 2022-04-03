local valMap = import 'validators.libsonnet';
local wellKnown = import 'well-known.libsonnet';
local typeMap = valMap + wellKnown;  // wellKnown will override keys in valMap for well-known types

local dispatch = function(to='validator', trace=true) (
  local unknown = function(typeName) (
    function(input, ctx) (
      if trace then
        std.trace('WARN: %s: no %s found for type %s' % [ctx, to, typeName], input)
      else
        input
    )
  );

  function(typeName, input, ctx='') (
    local context = if ctx == '' then typeName else ctx;
    local fn = if std.objectHas(typeMap, typeName) && std.objectHasAll(typeMap[typeName], to) then typeMap[typeName][to] else unknown(typeName);
    fn(input, context)
  )
);

dispatch
