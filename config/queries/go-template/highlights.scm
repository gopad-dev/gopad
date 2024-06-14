; Identifiers

[
  (field)
  (field_identifier)
  ] @property

(variable) @variable

; Function calls

(function_call
  function: (identifier) @function)

(method_call
  method: (selector_expression
            field: (field_identifier) @function.method))

; Operators

[
  "|"
  ":="
  ] @operator

; Builtin functions

((identifier) @function.builtin
  (#match? @function.builtin "^(and|call|html|index|slice|js|len|not|or|print|printf|println|urlquery|eq|ne|lt|ge|gt|ge)$"))

; Delimiters

[
  "."
  ","
  ] @punctuation.delimiter

[
  "{{"
  "}}"
  "{{-"
  "-}}"
  "-}}"
  ")"
  "("
  ] @punctuation.bracket

; Keywords

[
  "else"
  "if"
  "range"
  "with"
  "end"
  "template"
  "define"
  "block"
  ] @keyword

; Literals

[
  (interpreted_string_literal)
  (raw_string_literal)
  (rune_literal)
  ] @string

(escape_sequence) @string.special

(int_literal) @constant.numeric.integer

[
  (float_literal)
  (imaginary_literal)
  ] @constant.numeric.float

[
  (true)
  (false)
  ] @constant.builtin.boolean


(nil) @constant.builtin

(comment) @comment @spell
