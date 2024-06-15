; Identifiers

(type_identifier) @type
(field_identifier) @property
(identifier) @variable

; Function calls

(call_expression
  function: (identifier) @function.builtin
  (#match? @function.builtin "^(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover|min|max)$"))

(call_expression
  function: (identifier) @function)

(call_expression
  function: (selector_expression
              field: (field_identifier) @function.method))

; Types

(type_parameter_list
  (type_parameter_declaration
    name: (identifier) @type.parameter))

((type_identifier) @type.builtin
  (#match? @type.builtin "^(any|bool|byte|comparable|complex128|complex64|error|float32|float64|int|int16|int32|int64|int8|rune|string|uint|uint16|uint32|uint64|uint8|uintptr)$"))

; Function definitions

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (field_identifier) @function.method)

; Labels

(labeled_statement
  (label_name) @label)

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
  "case"
  "chan"
  "const"
  "continue"
  "default"
  "defer"
  "else"
  "fallthrough"
  "for"
  "func"
  "go"
  "goto"
  "if"
  "import"
  "interface"
  "map"
  "package"
  "range"
  "return"
  "select"
  "struct"
  "switch"
  "type"
  "var"
  ] @keyword

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
 (rune_literal)
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
