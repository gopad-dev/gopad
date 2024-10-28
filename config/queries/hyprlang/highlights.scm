(comment) @comment @spell

[
  "source"
  "exec"
  "exec-once"
  ] @keyword

(keyword
  (name) @keyword)

(assignment
  (name) @property)

(section
  (name) @section)

(section
  device: (device_name) @type)

(variable) @variable

"$" @punctuation.special

(boolean) @constant.builtin.boolean

(string) @string

(mod) @constant

[
  "rgb"
  "rgba"
  ] @function.builtin

[
  (number)
  (legacy_hex)
  (angle)
  (hex)
  ] @constant.numeric

"deg" @type

"," @punctuation.delimiter

[
  "("
  ")"
  "{"
  "}"
  ] @punctuation.bracket

[
  "="
  "-"
  "+"
  ] @operator
