; Keywords
(scheme) @module

; Methods
(method) @function.method

; Headers
(header
  name: (name) @constant)

; Variables
(variable_declaration
  name: (identifier) @variable)

; Fields
(pair
  name: (identifier) @variable.member)

; URL / Host
(host) @string.special.url

(path
  (identifier) @string.special.url)

; Parameters
(query_param
  (key) @variable.parameter)

; Operators
[
  "="
  "?"
  "&"
  "@"
  "<"
  ] @operator

; Literals
(target_url) @string.special.url

(http_version) @constant

; Response
(status_code) @constant.numeric
(status_text) @string

(string) @string

(number) @constant.numeric

(boolean) @constant.builtin.boolean

; Punctuation
[
  "{{"
  "}}"
  ] @punctuation.bracket

":" @punctuation.delimiter

; external JSON body
(external_body
  file_path: (path) @string.special.path)

; Comments
(comment) @comment @spell
