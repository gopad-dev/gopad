; Scopes

[
  (source_file)
  (function_declaration)
  (type_declaration)
  (block)
  ] @local.scope

; Definitions

(const_declaration
  (const_spec
    (identifier) @local.definition))

(var_declaration
  (var_spec
    (identifier) @local.definition))

(function_declaration
  (identifier) @local.definition)

(parameter_declaration
  (identifier) @local.definition)

(type_declaration
  (type_spec
    (type_identifier) @local.definition))

; References

(identifier) @local.reference

(field_identifier) @local.reference

(call_expression
  function: (identifier) @local.reference)
