; Scopes

[
  (source_file)
  (function_declaration)
  (func_literal)
  (method_declaration)
  (type_declaration)
  (block)
  (if_statement)
  (expression_switch_statement)
  (for_statement)
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
  name: (identifier) @local.definition)

(type_declaration
  (type_spec
    (type_identifier) @local.definition))

; References

;(identifier) @local.reference

(type_identifier) @local.reference

;(field_identifier) @local.reference

(parameter_declaration
  name: (identifier) @local.reference)

(parameter_declaration
  type: (type_identifier) @local.reference)

(composite_literal
  body: (literal_value
          (keyed_element
            (literal_element)
            (literal_element
              (identifier) @local.reference))))


(call_expression
  function: (identifier) @local.reference)

(call_expression
  arguments: (argument_list
               (identifier) @local.reference))
(selector_expression
  operand: (identifier) @local.reference)

(expression_list
  (identifier) @local.reference)

(inc_statement
  (identifier) @local.reference)

(dec_statement
  (identifier) @local.reference)

(binary_expression
  left: (identifier) @local.reference)
