; Identifiers

(type_identifier) @type
(field_identifier) @variable
(identifier) @variable

; mark exported fields as properties
((field_identifier) @property
  (#match? @property "^[A-Z]"))

(const_declaration
  (const_spec
    name: (identifier) @constant))

; mark exported var as properties
(source_file
  (var_declaration
    (var_spec
      name: (identifier) @property
      (#match? @property "^[A-Z]"))))

; Function calls

(call_expression
  function: (identifier) @function)

(call_expression
  function: (selector_expression
              field: (field_identifier) @function.method))

(call_expression
  function: (identifier) @function.builtin
  (#any-of? @function.builtin "append" "cap" "close" "complex" "copy" "delete" "imag" "len" "make" "new" "panic" "print" "println" "real" "recover" "min" "max"))

(call_expression
  function: (selector_expression
              field: (field_identifier) @function.method))

; Type Conversions

(call_expression
  function: (identifier) @type.builtin
  (#any-of? @type.builtin "any" "bool" "byte" "comparable" "complex128" "complex64" "error" "float32" "float64" "int" "int16" "int32" "int64" "int8" "rune" "string" "uint" "uint16" "uint32" "uint64" "uint8" "uintptr"))

; Function definitions

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (field_identifier) @function.method)

(function_declaration
  (parameter_list
    (parameter_declaration
      name: (identifier) @variable.parameter)))

(method_declaration
  (parameter_list
    (parameter_declaration
      name: (identifier) @variable.parameter)))

(type_declaration
  (type_spec
    (interface_type
      (method_elem
        name: (field_identifier) @function.method))))

(type_declaration
  (type_spec
    (interface_type
      (method_elem
        (parameter_list
          (parameter_declaration
            name: (identifier) @variable.parameter))))))

; Types

[
  "chan"
  "map"
  ] @type.builtin

(type_parameter_list
  (type_parameter_declaration
    name: (identifier) @type.parameter))

(func_literal
  parameters: (parameter_list
    (parameter_declaration
      name: (identifier) @variable.parameter)))

((type_identifier) @type.builtin
  (#any-of? @type.builtin "any" "bool" "byte" "comparable" "complex128" "complex64" "error" "float32" "float64" "int" "int16" "int32" "int64" "int8" "rune" "string" "uint" "uint16" "uint32" "uint64" "uint8" "uintptr"))

(composite_literal
  (literal_value
    (keyed_element
      .
      (literal_element
        (identifier) @variable))))

; mark exported fields as properties
(composite_literal
  (literal_value
    (keyed_element
      .
      (literal_element
        (identifier) @property
        (#match? @property "^[A-Z]")))))

; Labels

(label_name) @label

; Operators

[
  "--"
  "-"
  "-="
  ":="
  "!"
  "!="
  "..."
  "*"
  "*"
  "*="
  "/"
  "/="
  "&"
  "&&"
  "&="
  "%"
  "%="
  "^"
  "^="
  "+"
  "++"
  "+="
  "<-"
  "<"
  "<<"
  "<<="
  "<="
  "="
  "=="
  ">"
  ">="
  ">>"
  ">>="
  "|"
  "|="
  "||"
  "~"
  ] @operator

; Keywords

[
  "break"
  "continue"
  "default"
  "defer"
  "fallthrough"
  "go"
  "goto"
  "range"
  "select"
  ] @keyword

"func" @keyword.function

"return" @keyword.return

[
  "import"
  "package"
  ] @keyword.control.import

[
  "else"
  "case"
  "switch"
  "if"
  ] @keyword.conditional

"for" @keyword.repeat

[
  "var"
  "const"
  "type"
  "struct"
  "interface"
  ] @keyword.storage.type

; Delimiters

[
  ":"
  "."
  ","
  ";"
  ] @punctuation.delimiter

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
  ] @punctuation.bracket

; Literals

[
  (interpreted_string_literal)
  (raw_string_literal)
  ] @string

(rune_literal) @constant.character

(escape_sequence) @constant.character.escape

(int_literal) @constant.numeric.integer

[
  (float_literal)
  (imaginary_literal)
  ] @constant.numeric.float

[
  (true)
  (false)
  ] @constant.builtin.boolean

[
  (nil)
  (iota)
  ] @constant.builtin

(comment) @comment

((comment) @keyword.directive
  (#match? @keyword.directive "^//go:")
  (#offset! @keyword.directive 0 2 0 0))
