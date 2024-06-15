; Function calls

(call_expression
  function: (identifier) @function.builtin
  (.match? @function.builtin "^(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)$"))

(call_expression
  function: (identifier) @function)

(call_expression
  function: (selector_expression
    field: (field_identifier) @function.method))

; Function definitions

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (field_identifier) @function.method)

; Identifiers

(type_identifier) @type
(field_identifier) @property
(identifier) @variable

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

; Literals

[
  (interpreted_string_literal)
  (raw_string_literal)
  (rune_literal)
] @string

(escape_sequence) @escape


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


(component_declaration
  name: (component_identifier) @function)

([
  (tag_start)
  (tag_end)
  (self_closing_tag)
  (style_element)
] @tag
  (#set! priority 90))

(attribute
  name: (attribute_name) @attribute)

(attribute
  value: (quoted_attribute_value) @string)

;[
;  (element_text)
;  (style_element_text)
;] @string.special

(css_identifier) @function

(css_property
  name: (css_property_name) @property)

(css_property
  value: (css_property_value) @string)

[
  (expression)
  (dynamic_class_attribute_value)
] @function.method

(component_import
  name: (component_identifier) @function)

(component_render) @function.call

(element_comment) @comment @spell

"@" @operator

("=" @punctuation.delimiter
  (#set! priority 101))

([
  "<"
  ">"
] @punctuation.bracket
  (#set! priority 101))

[
  "templ"
  "css"
  "script"
] @keyword